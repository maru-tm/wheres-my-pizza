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

	// Declare fanout exchange for notifications
	err := ch.ExchangeDeclare(
		"notifications_fanout", // name
		"fanout",               // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue with auto-generated name (exclusive)
	queue, err := ch.QueueDeclare(
		"",    // name - empty means auto-generate
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		queue.Name,             // queue name
		"",                     // routing key (empty for fanout)
		"notifications_fanout", // exchange
		false,                  // no-wait
		nil,                    // arguments
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
		c.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
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
					// Log error but don't requeue malformed messages
					fmt.Printf("Failed to unmarshal message: %v\n", err)
					msg.Nack(false, false) // Discard message
					continue
				}

				// Add delivery tag for acknowledgement
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
