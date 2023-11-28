package main

import (
	"context"
	"log/slog"
	"os"

	"kagorbunov/PubSub_service/handler"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx := context.Background()
	connection, err := amqp.Dial(os.Getenv("AMQP_SERVER_URL"))
	if err != nil {
		slog.ErrorContext(ctx, "bad connect to broker: %s", err)
		return
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		slog.ErrorContext(ctx, "bad open channel of broker: %s", err)
		return
	}
	defer channel.Close()
	queuePublish, err := channel.QueueDeclare(
		"mq-publish",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.ErrorContext(ctx, "error with create queue: %s", err)
		return
	}
	queueSubscribe, err := channel.QueueDeclare(
		"mq-subscribe",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.ErrorContext(ctx, "error with create queue: %s", err)
		return
	}
	h := handler.NewHandler(channel, os.Getenv("WEATHER_URL"), queuePublish, queueSubscribe)

	if err = h.Handle(ctx); err != nil {
		slog.ErrorContext(ctx, "service failed to start: %v", err)
		return
	}
}
