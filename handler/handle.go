package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	MQ              *amqp.Channel
	WeatherURL      string
	QueuePublisher  amqp.Queue
	QueueSubscriber amqp.Queue
}

func NewHandler(MQ *amqp.Channel, weatherURL string, queuePublisher amqp.Queue, queueSubscriber amqp.Queue) *Handler {
	return &Handler{MQ: MQ, WeatherURL: weatherURL, QueuePublisher: queuePublisher, QueueSubscriber: queueSubscriber}
}

func (h *Handler) Handle(ctx context.Context) error {
	msgs, err := h.MQ.Consume(
		h.QueueSubscriber.Name, // queue
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to register a consumer", err)
		return err
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			err = h.subHandler(ctx, d)
			if err != nil {
				log.Fatalf("recieve msg fail: %s", err)
			}
			log.Printf("Success publishing!")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}

func (h *Handler) subHandler(ctx context.Context, d amqp.Delivery) error {
	req := new(WeatherAPI)
	if err := json.Unmarshal(d.Body, req); err != nil {
		log.Fatalf("fail unmarshaling delivery body: %s", err)
		return err
	}
	dns := fmt.Sprintf("%sforecast?latitude=%s&longitude=%s&hourly=%s&timezone=%s&forecast_days=%s",
		h.WeatherURL,
		req.Latitude,
		req.Longitude,
		req.Hourly,
		req.Timezone,
		req.Forecast_days,
	)
	r, err := http.Get(dns)
	if err != nil {
		log.Fatalf("bad request to weather api: %s", err)
		return err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	log.Printf("Body from api: \n %s", body)

	return h.Publish(ctx, body)
}

func (h *Handler) Publish(ctx context.Context, msg []byte) error {
	return h.MQ.PublishWithContext(ctx, "", h.QueuePublisher.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg,
	})
}

type WeatherAPI struct {
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	Hourly        string `json:"hourly"`
	Timezone      string `json:"timezone"`
	Forecast_days string `json:"forecast_Days"`
}
