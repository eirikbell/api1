package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

func main() {
	application := newApp()

	if application != nil {
		application.parseFile("/data/data.csv")
		// application.parseFile("./../data.csv")
	}

	log.Infoln("Terminating application")
}

type app struct {
	producer sarama.SyncProducer
}

func newApp() *app {
	brokers := []string{"localhost:9092"}

	producer, err := newKafkaSyncProducer(brokers)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	return &app{producer: producer}
}

func (application *app) parseFile(filename string) {
	log.Infoln("Parse file", filename)

	csvReader, err := fileToReader(filename)
	if err != nil {
		log.Errorln(err)
		return
	}

	sendChannel := make(chan error)
	linecount := 0
	firstLine := true
	var headerRow []string

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorln(err)
			return
		}

		if firstLine {
			headerRow = line
			firstLine = false
		} else {
			linecount++
			id := fmt.Sprintf("%s-%d", filename, linecount)

			go application.sendLineToKafka(sendChannel, headerRow, line, filename, id)
		}
	}

	for i := 0; i < linecount; i++ {
		select {
		case err := <-sendChannel:
			if err != nil {
				log.Errorln(err)
				return
			}
		case <-time.After(1 * time.Second):
			log.Errorln("Only", i, "completed")
			return
		}
	}

	log.Infoln("Completed processing file", filename)
}

func (application *app) sendLineToKafka(sendChannel chan error, headerRow, line []string, filename, id string) {
	row := make(map[string]string)
	for i := 0; i < len(headerRow); i++ {
		row[headerRow[i]] = line[i]
	}

	row["id"] = id
	row["filename"] = filename
	log.Infoln(row)

	bytes, err := json.Marshal(row)
	if err != nil {
		msg := fmt.Errorf("Failed to marshal json with error: %s", err)
		sendChannel <- msg
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: "pageviews2-2",
		Value: sarama.StringEncoder(bytes),
	}

	_, _, err = application.producer.SendMessage(msg)
	sendChannel <- err
}

func fileToReader(filename string) (*csv.Reader, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		msg := fmt.Sprintf("File %s do not exist", filename)
		return nil, errors.New(msg)
	}

	f, err := os.Open(filename)
	if err != nil {
		msg := fmt.Sprintf("Unable to open %s with error: %s", filename, err)
		return nil, errors.New(msg)
	}

	reader := bufio.NewReader(f)
	return csv.NewReader(reader), nil
}

func newKafkaSyncProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer(brokers, config)
}
