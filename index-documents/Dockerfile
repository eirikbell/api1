FROM golang:1.8

RUN mkdir /app /data
COPY index-documents/bin/index-documents /app/index-documents

VOLUME ["/data"]

WORKDIR /app
ENTRYPOINT ["/app/index-documents"]