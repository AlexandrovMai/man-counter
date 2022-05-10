package consumer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
)

type MessageReceiver interface {
	OnMessage(key, value []byte) error
}

type Consumer interface {
	StartConsumeAsync(receiver MessageReceiver, errorChan chan error)
	StopConsume()
}

type consumer struct {
	receive  bool
	ctx      context.Context
	cancelFn context.CancelFunc
	reader   *kafka.Reader
}

func (c *consumer) StartConsumeAsync(receiver MessageReceiver, errorChan chan error) {
	go func() {
		for c.receive {
			m, err := c.reader.FetchMessage(c.ctx)
			if err != nil {
				log.Println("Cannot fetch message: " + err.Error())
				errorChan <- err
				return
			}

			if err := receiver.OnMessage(m.Key, m.Value); err != nil {
				log.Println("failed to process messages: " + err.Error())
				errorChan <- err
				return
			}

			if err := c.reader.CommitMessages(c.ctx, m); err != nil {
				log.Println("failed to commit messages: " + err.Error())
				errorChan <- err
				return
			}
		}

	}()
}

func (c *consumer) StopConsume() {
	c.receive = false
	c.cancelFn()
}

func New(ctx context.Context, kafkaUrl, groupId, queueName string) Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaUrl},
		GroupID:  groupId,
		Topic:    queueName,
		MaxBytes: 10e6, // 10MB
	})
	ctx, cancelFunc := context.WithCancel(ctx)
	return &consumer{
		receive:  true,
		ctx:      ctx,
		cancelFn: cancelFunc,
		reader:   r,
	}
}
