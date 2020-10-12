FROM golang:latest AS sqs2http

WORKDIR /go/src/github.com/chaseisabelle/sqs2http

COPY ./go.sum ./
COPY ./go.mod ./
COPY ./*.go ./

RUN go get -v && go build -o /sqs2http && rm -rf /go

CMD ["/sqs2http"]
