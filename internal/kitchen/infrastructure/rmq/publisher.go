package rmq

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-system/pkg/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type StatusPublisher struct {
	rabbitmq *rabbitmq.RabbitMQ
}

func NewStatusPublisher(rabbitmq *rabbitmq.RabbitMQ) (*StatusPublisher, error) {
	ch := rabbitmq.Channel()

	err := ch.ExchangeDeclare(
		"notifications_fanout",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &StatusPublisher{rabbitmq: rabbitmq}, nil
}

func (p *StatusPublisher) PublishStatusUpdate(ctx context.Context, update *StatusUpdateMessage) error {
	body, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	err = p.rabbitmq.Channel().PublishWithContext(ctx,
		"notifications_fanout",
		"",
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish status update: %w", err)
	}

	return nil
}
