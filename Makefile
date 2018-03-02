current_dir = $(shell pwd)

build:
	echo $(current_dir)
	docker run --rm -e GOPATH=/usr -v "$(current_dir)/parse-file:/usr/src/parse-file" -w /usr/src/parse-file golang:1.9.4 go build -v -ldflags "-s" -a -installsuffix cgo -o bin/parse-file
	docker run --rm -e GOPATH=/usr -v "$(current_dir)/index-documents:/usr/src/index-documents" -w /usr/src/index-documents golang:1.9.4 go build -v -ldflags "-s" -a -installsuffix cgo -o bin/index-documents

package: build
	docker build -f parse-file/Dockerfile -t parse-file:initial .
	docker tag parse-file:initial parse-file:latest
	docker build -f index-documents/Dockerfile -t index-documents:initial .
	docker tag index-documents:initial index-documents:latest

run-parse: package
	docker run -v "$(current_dir)/data.csv:/data/data.csv" --net=host parse-file:latest

run-index: package
	docker run --net=host index-documents:latest