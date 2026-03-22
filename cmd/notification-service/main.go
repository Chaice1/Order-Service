package main

import (
	"context"
	"os"

	notificationadapter "github.com/Chaice1/Order-Service/internal/notification/adapter"
	notificationservice "github.com/Chaice1/Order-Service/internal/notification/service"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func main() {
	godotenv.Load()

	TgBot := notificationservice.NewTelegramBot(os.Getenv("CHAT_ID"), os.Getenv("BOT_TOKEN"))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_ADDR")},
		Topic:   "created_order",
		GroupID: "notification_service1",
	})

	consumer := notificationadapter.NewConsumer(reader, TgBot)

	consumer.ConsumeMessage(context.Background())

}
