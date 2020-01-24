package sqsc

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// the client
type SQSC struct {
	sqs    *sqs.SQS
	config Config
}

// the client configs
type Config struct {
	ID       string //<< aws account id
	Key    string //<< aws auth key
	Secret   string //<< aws account secret
	Region   string //<< aws region
	Queue    string //<< queue name
	URL      string //<< queue url
	Endpoint string //<< aws endpoint
	Retries  int    //<< max retries
	Timeout  int    //<< visibility timeout (seconds)
	Wait     int    //<< wait time (seconds)
}

// creates a new client instance
func New(cfg *Config) (*SQSC, error) {
	// default is no-auth
	crd := credentials.AnonymousCredentials

	// check if we do need to auth
	if cfg.Key != "" && cfg.Secret != "" {
		crd = credentials.NewStaticCredentials(cfg.Key, cfg.Secret, "")
	}

	// build the aws configs
	acf := aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: crd,
		MaxRetries:  aws.Int(cfg.Retries),
		Endpoint:    &cfg.Endpoint,
	}

	// boot the session
	ses, err := session.NewSession(&acf)

	// build the aws sqs client
	cli := sqs.New(ses, &acf)

	// get the queue url
	if cfg.URL == "" {
		url, err := cli.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName:              aws.String(cfg.Queue),
			QueueOwnerAWSAccountId: aws.String(cfg.ID),
		})

		if err != nil {
			return nil, err
		}

		if url == nil {
			return nil, errors.New("failed to get queue url")
		}

		cfg.URL = *url.QueueUrl
	}

	// build the struct
	return &SQSC{
		sqs:    cli,
		config: *cfg,
	}, err
}

// produce a new message on the queue
//
// bod - the message body
// del - the delay in seconds (usually just use 0)
//
// returns
// - the message id
// - error
func (c *SQSC) Produce(bod string, del int) (string, error) {
	// send message
	inp := sqs.SendMessageInput{
		MessageBody:  aws.String(bod),
		QueueUrl:     aws.String(c.config.URL),
		DelaySeconds: aws.Int64(int64(del)),
	}

	// send it
	res, err := c.sqs.SendMessage(&inp)

	// default message id
	id := ""

	// we get a response?
	if res != nil {
		// get id pointer
		ptr := res.MessageId

		// can we dereference it?
		if ptr != nil {
			// dereference it
			id = *res.MessageId
		}
	}

	// return the message id
	return id, err
}

// consume a single message from the queue
//
// returns
// - the message body
// - the receipt handle (use for deleting messages)
// - any error
func (c *SQSC) Consume() (string, string, error) {
	// receive message
	res, err := c.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:          aws.String(c.config.URL),
		VisibilityTimeout: aws.Int64(int64(c.config.Timeout)),
		WaitTimeSeconds:   aws.Int64(int64(c.config.Wait)),
	})

	// default message body
	bod := ""
	rh := ""

	// get message body if we can
	if res != nil && len(res.Messages) != 0 {
		// get the body pointer first
		msg := res.Messages[0]
		ptr := msg.Body

		// make sure we can dereference it
		if ptr != nil {
			// dereference it
			bod = *ptr
		}

		// get the receipt handle pointer
		ptr = msg.ReceiptHandle

		// make sure we can dereference it
		if ptr != nil {
			// dereference it
			rh = *ptr
		}
	}

	// we done fam
	return bod, rh, err
}

// delete a message from the queue
//
// rh - the receipt handle (from sqsc.Consume())
//
// returns
// - the response (will be empty if success)
// - any error
func (c *SQSC) Delete(rh string) (string, error) {
	// delete that pesky message
	res, err := c.sqs.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.config.URL),
		ReceiptHandle: &rh,
	}) // no response returned when success

	// default body
	bod := ""

	// did we get a response
	if res != nil {
		// convert to string
		bod = res.String()
	}

	// we done fam
	return bod, err
}
