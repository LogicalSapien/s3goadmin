FROM "golang:1.15"

MAINTAINER "Elmo Yeldo <contact@logicalsapien.com>"

WORKDIR /go/src

COPY . /go/src

RUN cd /go/src

RUN go get -u github.com/aws/aws-sdk-go && go get github.com/google/uuid && go get github.com/boltdb/bolt/

RUN go build -o main

EXPOSE 8080

ENTRYPOINT "./main"
