# sqs2http
forward messages from sqs to http endpoints

---
## usage

```
./sqs2http --help
Usage of ./sqs2http:
  -endpoint string
    	the aws endpoint
  -id string
    	aws account id (leave blank for no-auth)
  -method string
    	the request method to send the message with (default "GET")
  -region string
    	aws region (i.e. us-east-1)
  -retries int
    	the max number of retries (default -1)
  -secret string
    	aws account secret (leave blank for no-auth)
  -send-to string
    	the url to send the message to
  -status int
    	the successful status code (default 200)
  -timeout int
    	the message visibility timeout in seconds (default 30)
  -token string
    	aws account token (leave blank for no-auth)
  -url string
    	the sqs queue url
  -verbose
    	verbose output
  -wait int
    	wait time in seconds
  -workers int
    	the number of concurrent workers (default 1)
```