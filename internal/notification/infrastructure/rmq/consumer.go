package rmq

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-system/pkg/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type NotificationConsumer struct {
	channel *amqp091.Channel
	queue   amqp091.Queue
}

func NewNotificationConsumer(rabbitmq *rabbitmq.RabbitMQ) (*NotificationConsumer, error) {
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

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		queue.Name,
		"",
		"notifications_fanout",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &NotificationConsumer{
		channel: ch,
		queue:   queue,
	}, nil
}

func (c *NotificationConsumer) Consume(ctx context.Context) (<-chan *StatusUpdateMessage, error) {
	msgs, err := c.channel.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	messages := make(chan *StatusUpdateMessage, 100)

	go func() {
		defer close(messages)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var statusMsg StatusUpdateMessage
				if err := json.Unmarshal(msg.Body, &statusMsg); err != nil {
					fmt.Printf("Failed to unmarshal message: %v\n", err)
					err := msg.Nack(false, false)
					if err != nil {
						return
					}
					continue
				}

				statusMsg.DeliveryTag = msg.DeliveryTag
				messages <- &statusMsg
			}
		}
	}()

	return messages, nil
}

func (c *NotificationConsumer) Ack(deliveryTag uint64) error {
	return c.channel.Ack(deliveryTag, false)
}

func (c *NotificationConsumer) Nack(deliveryTag uint64, requeue bool) error {
	return c.channel.Nack(deliveryTag, false, requeue)
}

func (c *NotificationConsumer) Close() error {
	return c.channel.Close()
}
