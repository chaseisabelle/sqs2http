# sqs2http
forward messages from sqs to http endpoints

---
## docker images

get the latest from [docker](https://hub.docker.com/repository/docker/chaseisabelle/sqs2http/)

---
## usage

```
./sqs2http --help
Usage of ./sqs2http:
  -endpoint string
    	the aws endpoint
  -id string
    	aws account id (leave blank for no-auth)
  -key string
    	aws account key (leave blank for no-auth)
  -method string
    	the request method timeout send the message with (default "GET")
  -queue string
    	the queue name
  -region string
    	aws region (i.e. us-east-1)
  -requeue value
    	the http status code timeout requeue a message for
  -retries int
    	the workers number of retries (default -1)
  -secret string
    	aws account secret (leave blank for no-auth)
  -send-to string
    	the url to forward the messages to
  -timeout int
    	the message visibility timeout in seconds (default 30)
  -url string
    	the sqs queue url
  -verbose
    	verbose output
  -wait int
    	wait time in seconds
  -workers int
    	the number of concurrent workers (default 1)
```