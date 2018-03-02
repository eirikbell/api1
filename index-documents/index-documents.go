package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	elastic "github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

var (
	sem = make(chan struct{}, 100)
)

func main() {
	application := newApp()

	if application != nil {
		application.processMessages()
	}

	log.Infoln("Terminating application")
}

type app struct {
	consumer            *cluster.Consumer
	elasticsearchClient *elastic.Client
}

func newApp() *app {
	brokers := []string{"localhost:9092"}
	consumer, err := newKafkaConsumer(brokers)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	log.Infoln("Connected to kafka on ", brokers)

	elasticsearchURL := "http://localhost:9200"
	elasticsearchClient, err := elastic.NewClient(elastic.SetURL(elasticsearchURL), elastic.SetSniff(false), elastic.SetHealthcheck(false))
	if err != nil {
		log.Errorln(err)
		return nil
	}

	template := map[string]interface{}{
		"index_patterns": []string{"pageviews2"},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"pv": map[string]interface{}{
				"properties": map[string]interface{}{
					"timeSpentMs": map[string]interface{}{
						"type": "long",
					},
				},
			},
		},
	}

	elasticsearchClient.IndexPutTemplate("pageviews2-mapping").BodyJson(template).Do(context.Background())
	log.Infoln("Connected to elasticsearch on ", elasticsearchURL)

	return &app{consumer: consumer, elasticsearchClient: elasticsearchClient}
}

func (application *app) processMessages() {
	for {
		select {
		case err := <-application.consumer.Errors():
			log.Errorln(err)
			return
		case ntf := <-application.consumer.Notifications():
			log.Infoln(ntf)
		case msg := <-application.consumer.Messages():
			sem <- struct{}{}
			go application.indexDocument(msg.Value)
		case <-time.After(1 * time.Second):
			log.Infoln("END OF TOPIC")
			return
		}
	}
}

func (application *app) indexDocument(msgValue []byte) {
	defer func() { <-sem }()

	pageview := make(map[string]string)
	err := json.Unmarshal(msgValue, &pageview)
	if err != nil {
		log.Errorln("Error parsing pageview ", err)
		return
	}

	_, err = application.elasticsearchClient.Index().Index("pageviews2").Type("pv").Id(pageview["id"]).BodyJson(pageview).Do(context.Background())
	if err != nil {
		log.Errorln("Error indexing pageview ", err)
		return
	}

	log.Infoln("successfully indexed ", pageview["id"])
}

func newKafkaConsumer(brokers []string) (*cluster.Consumer, error) {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	return cluster.NewConsumer(brokers, "pageviews2-cg90", []string{"pageviews2-2"}, config)
}
