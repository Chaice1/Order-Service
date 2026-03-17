package notificationadapter

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type NotificationService interface {
	SendMessage(context.Context, string, string) error
}
type Consumer struct {
	r  *kafka.Reader
	tb NotificationService
}

func NewConsumer(reader *kafka.Reader, telegrambot NotificationService) *Consumer {
	return &Consumer{
		r:  reader,
		tb: telegrambot,
	}
}

func (k *Consumer) ConsumeMessage(ctx context.Context) {

	type msg struct {
		OrderId string `json:"order_id"`
		UserId  string `json:"user_id"`
	}

	jobs := make(chan ([]byte), 100)

	for i := 0; i < 10; i++ {
		go func() {
			for job := range jobs {
				req := msg{}
				err := json.Unmarshal(job, &req)
				if err != nil {
					log.Println("Parsing error")
				}
				err = k.tb.SendMessage(ctx, req.OrderId, req.UserId)
				if err != nil {
					log.Println("Notification Error")
				}
			}
		}()
	}

	for {
		message, err := k.r.ReadMessage(ctx)

		if err != nil {
			log.Printf("There are not messages")
			continue
		}

		jobs <- message.Value
	}
}
