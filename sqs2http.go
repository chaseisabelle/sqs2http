package main

import (
	"bytes"
	"errors"
	"flag"
	"github.com/chaseisabelle/flagz"
	"github.com/chaseisabelle/sqsc"
	"github.com/g3n/engine/util/logger"
	"net/http"
)

func main() {
	id := flag.String("id", "", "aws account id (leave blank for no-auth)")
	secret := flag.String("secret", "", "aws account secret (leave blank for no-auth)")
	token := flag.String("token", "", "aws account token (leave blank for no-auth)")
	region := flag.String("region", "", "aws region (i.e. us-east-1)")
	url := flag.String("url", "", "the sqs queue url")
	endpoint := flag.String("endpoint", "", "the aws endpoint")
	retries := flag.Int("retries", -1, "the workers number of retries")
	timeout := flag.Int("timeout", 30, "the message visibility timeout in seconds")
	wait := flag.Int("wait", 0, "wait time in seconds")
	sendTimeout := flag.String("send-timeout", "", "the url timeout send the message timeout")
	method := flag.String("method", "GET", "the request method timeout send the message with")
	workers := flag.Int("workers", 1, "the number of concurrent workers")
	verbose := flag.Bool("verbose", false, "verbose output")

	var flags flagz.Flagz

	flag.Var(&flags, "requeue", "the http status code timeout requeue a message for")

	flag.Parse()

	if *verbose {
		logger.SetLevel(logger.DEBUG)
	}

	statuses, err := flags.Intz()

	if err != nil {
		fail("failed to load requeue status configs", err)

		panic(err)
	}

	if *workers < 1 {
		err := errors.New("need at least 1 worker")

		fail("failed to init workers", err)

		panic(err)
	}

	sqs, err := sqsc.New(&sqsc.Config{
		ID:       *id,
		Secret:   *secret,
		Token:    *token,
		Region:   *region,
		URL:      *url,
		Endpoint: *endpoint,
		Retries:  *retries,
		Timeout:  *timeout,
		Wait:     *wait,
	})

	if err != nil {
		fail("failed to init sqs client", err)

		panic(err)
	}

	cli := http.Client{}

	for *workers > 0 {
		go func() {
			for {
				bod, rh, err := sqs.Consume()

				if err != nil {
					fail("consumer failure", err)

					continue
				}

				if bod == "" && rh == "" {
					continue
				}

				debug("consumed message", bod)

				req, err := http.NewRequest(*method, *sendTimeout, bytes.NewBuffer([]byte(bod)))

				if err != nil {
					fail("http request failure", err)

					continue
				}

				res, err := cli.Do(req)

				if err != nil {
					fail("http query failure", err)

					continue
				}

				if res == nil {
					fail("http response failure", errors.New("received nil response"))

					continue
				}

				var bdy []byte

				_, err = res.Body.Read(bdy)

				if err != nil {
					fail("http response read failure", err)
				}

				debug("http response read", string(bdy))

				err = res.Body.Close()

				if err != nil {
					fail("http response close failure", err)
				}

				sc := res.StatusCode

				debug("received http status code", sc)

				if has(sc, statuses) {
					continue
				}

				rsp, err := sqs.Delete(rh)

				if err != nil {
					fail("sqs delete failure", err)
				}

				debug("sqs delete response", rsp)
			}
		}()

		*workers--
	}

	chn := make(chan bool)

	<-chn
}

func fail(msg string, etc interface{}) {
	logger.Error("[ERR] " + msg + ": %+v", etc)
}

func debug(msg string, etc interface{}) {
	logger.Debug("[DBG] " + msg + ": %+v", etc)
}

func has(num int, arr []int) bool {
	for _, val := range arr {
		if num == val {
			return true
		}
	}

	return false
}
