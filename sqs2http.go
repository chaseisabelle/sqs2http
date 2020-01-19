package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/chaseisabelle/sqsc"
	"github.com/g3n/engine/util/logger"
	"net/http"
)

func main() {
	id := flag.String("id", "", "aws account id (leave blank for no-auth)")
	sec := flag.String("secret", "", "aws account secret (leave blank for no-auth)")
	tok := flag.String("token", "", "aws account token (leave blank for no-auth)")
	reg := flag.String("region", "", "aws region (i.e. us-east-1)")
	url := flag.String("url", "", "the sqs queue url")
	ep := flag.String("endpoint", "", "the aws endpoint")
	ret := flag.Int("retries", -1, "the max number of retries (defaults to -1)")
	to := flag.Int("timeout", 30, "the message visibility timeout in seconds (defaults to 30 seconds)")
	wts := flag.Int("wait", 0, "wait time in seconds (defaults to 0)")
	st := flag.String("send-to", "", "the url to send the message to")
	mth := flag.String("method", "GET", "the request method to send the message with")
	sta := flag.Int("status", 200, "the successful status code")
	max := flag.Int("workers", 1, "the number of concurrent workers (defaults to 1)")
	vrb := flag.Bool("verbose", false, "verbose output")

	flag.Parse()

	if *vrb {
		logger.SetLevel(logger.DEBUG)
	}

	sqs, err := sqsc.New(&sqsc.Config{
		ID:       *id,
		Secret:   *sec,
		Token:    *tok,
		Region:   *reg,
		URL:      *url,
		Endpoint: *ep,
		Retries:  *ret,
		Timeout:  *to,
		Wait:     *wts,
	})

	if err != nil {
		panic(err)
	}

	cli := http.Client{}

	for *max > 0 {
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

				req, err := http.NewRequest(*mth, *st, bytes.NewBuffer([]byte(bod)))

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

				bdy := make([]byte, 0)

				_, err = res.Body.Read(bdy)

				if err != nil {
					fail("http response read failure", err)
				}

				debug("http response read", string(bdy))

				err = res.Body.Close()

				if err != nil {
					fail("http response close failure", err)

					continue
				}

				if *sta != res.StatusCode {
					err = errors.New(fmt.Sprintf("expected %+v, received %+v", *sta, res.StatusCode))

					fail("http response status failure", err)

					continue
				}

				rsp, err := sqs.Delete(rh)

				if err != nil {
					fail("sqs delete failure", err)

					continue
				}

				debug("sqs delete response", rsp)
			}
		}()

		*max--
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
