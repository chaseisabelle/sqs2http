package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/chaseisabelle/flagz"
	"github.com/chaseisabelle/sqsc"
	"github.com/chaseisabelle/stop"
	"github.com/g3n/engine/util/logger"
	"net/http"
	"time"
)

var listener chan struct{}

func init() {
	listener = make(chan struct{})
}

func main() {
	id := flag.String("id", "", "aws account id (leave blank for no-auth)")
	key := flag.String("key", "", "aws account key (leave blank for no-auth)")
	secret := flag.String("secret", "", "aws account secret (leave blank for no-auth)")
	region := flag.String("region", "", "aws region (i.e. us-east-1)")
	url := flag.String("url", "", "the sqs queue url")
	queue := flag.String("queue", "", "the queue name")
	endpoint := flag.String("endpoint", "", "the aws endpoint")
	retries := flag.Int("retries", -1, "the workers number of retries")
	timeout := flag.Int("timeout", 30, "the message visibility timeout in seconds")
	wait := flag.Int("wait", 0, "wait time in seconds")
	to := flag.String("send-to", "", "the url to forward the messages to")
	method := flag.String("method", "GET", "the request method timeout send the message with")
	workers := flag.Int("workers", 1, "the number of concurrent workers")
	verbose := flag.Bool("verbose", false, "verbose output")
	boAfter := flag.Int("backoff-after", 1, "backoff after this many empty responses (0 for no backoff)")
	boMax := flag.Int("backoff-max", 10, "max sleep time for the exponential backoff")

	var flags flagz.Flagz

	flag.Var(&flags, "requeue", "the http status code timeout requeue a message for")

	flag.Parse()

	if *verbose {
		logger.SetLevel(logger.DEBUG)
	}

	var err error

	if *region == "" {
		err = errors.New("no region provided")
	}

	if *to == "" {
		err = errors.New("no send-to url provided")
	}

	if *url == "" && *queue == "" {
		err = errors.New("no queue url or name provided")
	}

	if err != nil {
		die("config failure", err)
	}

	statuses, err := flags.Intz()

	if err != nil {
		die("failed to load requeue status configs", err)
	}

	if *workers < 1 {
		err := errors.New("need at least 1 worker")

		die("failed to init workers", err)
	}

	sqs, err := sqsc.New(&sqsc.Config{
		ID:       *id,
		Secret:   *secret,
		Key:      *key,
		Region:   *region,
		Queue:    *queue,
		URL:      *url,
		Endpoint: *endpoint,
		Retries:  *retries,
		Timeout:  *timeout,
		Wait:     *wait,
	})

	if err != nil {
		die("failed to init sqs client", err)
	}

	cli := http.Client{}
	tmp := *workers

	stop.Listen()

	for tmp > 0 {
		go func() {
			empties := uint64(0)
			bo := 1

			for {
				if stop.Stopped() {
					debug("graceful exit", "worker")

					break
				}

				bod, rh, err := sqs.Consume()

				if err != nil {
					fail("consumer failure", err)

					continue
				}

				if bod == "" && rh == "" {
					empties++

					if empties >= uint64(*boAfter) {
						dur := time.Duration(bo) * time.Second

						debug("sleeping", dur)

						time.Sleep(dur)

						bo += bo

						if bo > *boMax {
							bo = *boMax
						}
					}

					continue
				}

				empties = 0
				bo = 1

				debug("consumed message", bod)

				req, err := http.NewRequest(*method, *to, bytes.NewBuffer([]byte(bod)))

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

				sc := res.StatusCode

				debug("received http status code", sc)

				if has(sc, statuses) {
					info("requeue due to http response code", sc)

					continue
				}

				_, err = sqs.Delete(rh)

				if err != nil {
					fail("sqs delete failure", err)
				}
			}

			listener <- struct{}{}
		}()

		tmp--
	}

	tmp = 0

	for range listener {
		tmp++

		if tmp >= *workers {
			break
		}
	}

	info("graceful exit", "bye bye")
}

func die(msg string, etc interface{}) {
	panic(fmt.Sprintf("[DIE] "+msg+": %+v", etc))
}

func fail(msg string, etc interface{}) {
	logger.Error("[ERR] "+msg+": %+v", etc)
}

func debug(msg string, etc interface{}) {
	logger.Debug("[DBG] "+msg+": %+v", etc)
}

func info(msg string, etc interface{}) {
	logger.Info("[INF] "+msg+": %+v", etc)
}

func has(num int, arr []int) bool {
	for _, val := range arr {
		if num == val {
			return true
		}
	}

	return false
}
