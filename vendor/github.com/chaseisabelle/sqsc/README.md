# sqs client

### wtf is it?
a very very very simple sqs client package in go

---

### usage

#### configs
```go
type Config struct {
	ID       string //<< aws account id
	Secret   string //<< aws account secret
	Token    string //<< aws auth token
	Region   string //<< aws region
	URL      string //<< queue url
	Endpoint string //<< aws endpoint
	Retries  int    //<< max retries
	Timeout  int    //<< visibility timeout (seconds)
	Wait     int    //<< wait time (seconds)
}
```

#### new client
```go
cli, err := sqsc.New(&sqsc.Config{
    ...
})
```

#### produce a message
```go
id, err := cli.Produce("my cool message", del)
```
- del - the delay for the message (just use 0)
- id - the message id
- err - any error

#### consume a message
```go
bod, rh, err := cli.Consume()
```
- bod - the message body
- rh - the receipt handle (use for deleting message)
- err - any error

note: if `bod == "" && rh == "" && err == nil` then the queue is empty, or no messages are visible

#### delete a message
```go
res, err = cli.Delete(rh)
```
- rh - the receipt handle (from `cli.Consume()`)
- res - the delete response (empty if successful)
- err - any error

---

### example
```go
package main

import (
	"fmt"
	"github.com/chaseisabelle/sqsc"
	"os"
)

func main() {
	bod := os.Args[1]

	cli, err := sqsc.New(&sqsc.Config{
		Region:   "us-east-1",
		URL:      "http://localhost:4100/queue/job",
		Endpoint: "http://127.0.0.1:4100",
	})

	res, err := cli.Produce(bod, 0)

	fmt.Printf("produce response: %+v\n", res)
	fmt.Printf("produce error:    %+v\n", err)

	res, rh, err := cli.Consume()

	fmt.Printf("consume response:       %+v\n", res)
	fmt.Printf("consume receipt handle: %+v\n", rh)
	fmt.Printf("consume error:          %+v\n", err)

	res, err = cli.Delete(rh)

	fmt.Printf("delete response: %+v\n", res)
	fmt.Printf("delete error:    %+v\n", err)
}
```

---

## notes and junk
- this is a super simple client, intentionally
- if you need to get more fancy, fork it or build your own
- please feel free to contribute, but do not complicate
- please report any bugs or feature requests in the "issues" tab

merci et bonne journee
