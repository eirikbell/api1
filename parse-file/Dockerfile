FROM golang:1.8

RUN mkdir /app /data

COPY parse-file/bin/parse-file /app/parse-file
# RUN ["chmod", "+x", "/app/parse-file"]

VOLUME ["/data"]

WORKDIR /app
CMD ["/app/parse-file"]