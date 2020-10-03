FROM golang:latest AS sqs2http

RUN go get github.com/chaseisabelle/sqs2http

WORKDIR /go/src/github.com/chaseisabelle/sqs2http

RUN go get -v && go build -o /sqs2http && rm -rf /go

CMD ["/sqs2http"]