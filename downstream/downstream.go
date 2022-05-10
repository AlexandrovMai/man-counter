package downstream

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"man-counter/storage"
	"time"
)

type notifier struct {
	s          *storage.Storage
	w          *kafka.Writer
	ctx        context.Context
	cancelFunc context.CancelFunc
	notifying  bool
	sleepTime  int
}

func (n *notifier) StartNotifyAsync(errorChan chan error) {
	go func() {
		for n.notifying {
			activeUsers, err := n.s.GetActiveUsersCount()
			if err != nil {
				log.Println("Notifier: GetActiveUsersCount Error:", err.Error())
				errorChan <- err
				return
			}
			// do not send messages if no receivers online...
			if activeUsers < 1 {
				continue
			}

			visitors, err := n.s.GetVisitorsCount()
			if err != nil {
				log.Println("Notifier: GetVisitorsCount Error:", err.Error())
				errorChan <- err
				return
			}
			if visitors < 0 {
				log.Println("VISITORS < 0... WTF WTF WTF...")
				_ = n.s.SetZeroVisitors()
				visitors = 0
			}

			downstreamPackage := make([]Message, 2)
			downstreamPackage[0].Name = VisitorsMesName
			downstreamPackage[0].Value = visitors

			downstreamPackage[1].Name = ActiveUsersMesName
			downstreamPackage[1].Value = activeUsers

			payload, err := json.Marshal(downstreamPackage)
			if err != nil {
				log.Println("Notifier: json.Marshal Error:", err.Error())
				errorChan <- err
				return
			}
			if err = n.w.WriteMessages(n.ctx, kafka.Message{
				Key:   []byte("downstream"),
				Value: payload,
			}); err != nil {
				log.Println("Notifier: WriteMessages Error:", err.Error())
				errorChan <- err
				return
			}

			time.Sleep(time.Duration(n.sleepTime) * time.Second)
		}
	}()
}

func (n *notifier) StopNotify() {
	n.notifying = false
	n.cancelFunc()
}

type Notifier interface {
	StartNotifyAsync(errorChan chan error)
	StopNotify()
}

func New(ctx context.Context, s *storage.Storage, kafkaUrl, queueName string, sleepTime int) Notifier {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(kafkaUrl),
		Topic:                  queueName,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	ctx, cancelFun := context.WithCancel(ctx)

	return &notifier{
		s:          s,
		w:          w,
		ctx:        ctx,
		cancelFunc: cancelFun,
		notifying:  true,
		sleepTime:  sleepTime,
	}
}
