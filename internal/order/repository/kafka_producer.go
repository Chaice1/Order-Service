package orderrepo

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type OrderProducer struct {
	w *kafka.Writer
}

func NewOrderProducer(broker string) *OrderProducer {
	return &OrderProducer{w: &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        "created_order",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}}
}

func (op *OrderProducer) PublishOrderCreated(ctx context.Context, order_id, user_id string) error {
	message := map[string]string{
		"order_id": order_id,
		"user_id":  user_id,
		"event":    "ORDER_CREATED",
	}

	payload, _ := json.Marshal(message)

	return op.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order_id),
		Value: payload,
	})

}

func (op *OrderProducer) Close() error {
	return op.w.Close()
}
