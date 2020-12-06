package event

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
)

// SQSClient is a abstraction of a concrete SQS struct.
type SQSClient interface {
	NewRequest(operation *request.Operation, params interface{}, data interface{}) *request.Request
	GetQueueUrlRequest(input *sqs.GetQueueUrlInput) (req *request.Request, output *sqs.GetQueueUrlOutput)
	GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessageRequest(input *sqs.ReceiveMessageInput) (req *request.Request, output *sqs.ReceiveMessageOutput)
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageRequest(input *sqs.DeleteMessageInput) (req *request.Request, output *sqs.DeleteMessageOutput)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
}

// SQSReceiverAdapter is a concrete struct of SQS Receiver Adapter.
type SQSReceiverAdapter struct {
	logger            *logrus.Logger
	signal            chan os.Signal
	queueName         *string
	waitTimeSecond    *int64
	visibilityTimeout *int64
	client            SQSClient
	messageHandler    MessageHandler
}

// NewSQSReceiverAdapter is a constructor.
func NewSQSReceiverAdapter(signal chan os.Signal, logger *logrus.Logger, client SQSClient, queueName string, WaitTimeSecond, VisibilityTimeout int64, messageHandler MessageHandler) Receiver {
	return &SQSReceiverAdapter{
		logger:            logger,
		signal:            signal,
		queueName:         &queueName,
		waitTimeSecond:    &WaitTimeSecond,
		visibilityTimeout: &VisibilityTimeout,
		client:            client,
		messageHandler:    messageHandler,
	}
}

// Receive will poll the message from the broker.
func (r SQSReceiverAdapter) Receive() {
	queueURL, err := r.getQueueURL()
	if err != nil {
		return
	}

	run := true
	for run {
		select {
		case <-r.signal:
			run = false
		default:
			if err := r.poll(queueURL); err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist || awsErr.Code() == sqs.ErrCodeQueueDeletedRecently {
						run = false
					}
				}
			}
		}
	}
}

func (r SQSReceiverAdapter) poll(queueURL *string) error {
	msgResult, err := r.client.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   r.visibilityTimeout,
		WaitTimeSeconds:     r.waitTimeSecond,
	})

	if err != nil {
		r.logger.Error(err)
		return err
	}

	if len(msgResult.Messages) < 1 {
		return nil
	}

	for _, message := range msgResult.Messages {
		err := r.messageHandler.Handle(message)
		if err != nil {
			continue
		}
		r.client.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
	}
	return nil
}

func (r SQSReceiverAdapter) getQueueURL() (*string, error) {
	queueURLResult, err := r.client.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: r.queueName,
	})

	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	return queueURLResult.QueueUrl, nil
}
