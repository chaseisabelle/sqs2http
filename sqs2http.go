package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/chaseisabelle/backoff/expbo"
	"github.com/chaseisabelle/flagz"
	"github.com/chaseisabelle/sqsc"
	"github.com/chaseisabelle/stop"
	"github.com/g3n/engine/util/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"time"
)

var (
	httpHisto = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "sqs2http_http_latency_seconds",
			Help: "http request latency",
		},
		[]string{"status", "context"},
	)
	sqsHisto = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "sqs2http_sqs_latency_seconds",
			Help: "sqs request latency",
		},
		[]string{"status", "context"},
	)
)

// main process
func main() {
	// init the configs for everything
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
	metrics := flag.Bool("metrics", false, "enable/disable metrics")
	host := flag.String("host", "127.0.0.1:8088", "the http host for the metrics endpoint")

	var flags flagz.Flagz

	flag.Var(&flags, "requeue", "the http status code timeout requeue a message for")

	flag.Parse()

	if *verbose {
		logger.SetLevel(logger.DEBUG)
	} else {
		logger.SetLevel(logger.INFO)
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

	// init the sqs client to connect to the queue
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

	// init metrics server if enabled
	if *metrics {
		http.Handle("/metrics", promhttp.Handler())

		go func() {
			err = http.ListenAndServe(*host, nil)

			if err != nil {
				die("failed to start server", err)
			}
		}()
	}

	// init http client to connect to the web server
	cli := http.Client{}

	// listen for kill/term/stop signals from user/os
	stop.Listen()

	// listen for workers to gracefully exit
	ear := make(chan struct{})

	// create the workers
	for i := 0; i < *workers; i++ {
		// spawn a goroutine for each worker
		go func() {
			empties := uint64(0)                                       //<< keeps track of number of subsequent empty replies from queue
			bo, err := expbo.New(uint64(1000), uint64(*boMax)*1000, 2) //<< exponential backoff

			if err != nil {
				die("failed to init backoff", err)
			}

			// create forever loop to constantly listen/ping for messages from the queue
			for {
				// check if user/os has stopped the program and exit gracefully
				if stop.Stopped() {
					debug("graceful exit", "worker")

					break
				}

				// attempt to consume a message from the queue
				start := time.Now()
				bod, rh, err := sqs.Consume()
				dur := time.Since(start).Seconds()

				if err != nil {
					fail("consumer failure", err)

					if *metrics {
						sqsHisto.WithLabelValues("failure", "consume").Observe(dur)
					}

					continue
				}

				// if theres no body and no receipt handle then the queue is empty
				if bod == "" && rh == "" {
					if *metrics {
						sqsHisto.WithLabelValues("empty", "consume").Observe(dur)
					}

					empties++ //<< track number of empty replies

					// back off if necessary
					if *boAfter != 0 && empties >= uint64(*boAfter) {
						debug("sleeping", bo)

						bo.Backoff()
					}

					// skip to next iteration
					continue
				}

				if *metrics {
					sqsHisto.WithLabelValues("success", "consume").Observe(dur)
				}

				// reset number of empty replies
				empties = 0

				// reset the exponential backoff
				bo.Reset()

				debug("consumed message", bod)

				// create new http request
				req, err := http.NewRequest(*method, *to, bytes.NewBuffer([]byte(bod)))

				if err != nil {
					fail("http request failure", err)

					if *metrics {
						httpHisto.WithLabelValues("failure", "request").Observe(dur)
					}

					continue
				}

				// execute http request
				res, err := cli.Do(req)

				if err != nil {
					fail("http query failure", err)

					if *metrics {
						httpHisto.WithLabelValues("failure", "query").Observe(dur)
					}

					continue
				}

				// check the http response
				if res == nil {
					fail("http response failure", errors.New("received nil response"))

					if *metrics {
						httpHisto.WithLabelValues("failure", "response").Observe(dur)
					}

					continue
				}

				err = res.Body.Close()

				if err != nil {
					fail("failed to close http response body", err)
				}

				sc := res.StatusCode

				debug("received http status code", sc)

				// if the http code is in the requeue codes the requeue the message
				if has(sc, statuses) {
					info("requeue due to http response code", sc)

					if *metrics {
						httpHisto.WithLabelValues(strconv.Itoa(sc), "requeue").Observe(dur)
					}

					continue
				}

				if *metrics {
					httpHisto.WithLabelValues(strconv.Itoa(sc), "query").Observe(dur)
				}

				// elsewise delete the message from the queue
				start = time.Now()
				_, err = sqs.Delete(rh)
				dur = time.Since(start).Seconds()

				if err != nil {
					fail("sqs delete failure", err)

					if *metrics {
						sqsHisto.WithLabelValues("failure", "delete").Observe(dur)
					}

					continue
				}

				if *metrics {
					sqsHisto.WithLabelValues("success", "delete").Observe(dur)
				}
			}

			// if we have broken out of the forever loop then we need to exit gracefully
			ear <- struct{}{}
		}()
	}

	// listen for workers to exit gracefully
	for i := 0; i < *workers; i++ {
		<-ear
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
