package rmq

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-system/pkg/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type OrderConsumer struct {
	channel *amqp091.Channel
	queue   amqp091.Queue
}

func NewOrderConsumer(rabbitmq *rabbitmq.RabbitMQ, prefetch int) (*OrderConsumer, error) {
	ch := rabbitmq.Channel()

	err := ch.Qos(
		prefetch,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	err = ch.ExchangeDeclare(
		"orders_topic",
		"topic",
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
		"kitchen_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		queue.Name,
		"kitchen.*.*",
		"orders_topic",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &OrderConsumer{
		channel: ch,
		queue:   queue,
	}, nil
}

func (c *OrderConsumer) ConsumeOrders(ctx context.Context) (<-chan *OrderMessage, error) {
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

	orderMsgs := make(chan *OrderMessage, 100)

	go func() {
		defer close(orderMsgs)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var orderMsg OrderMessage
				if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
					msg.Nack(false, false)
					continue
				}

				orderMsg.DeliveryTag = msg.DeliveryTag
				orderMsgs <- &orderMsg
			}
		}
	}()

	return orderMsgs, nil
}

func (c *OrderConsumer) AckMessage(deliveryTag uint64) error {
	return c.channel.Ack(deliveryTag, false)
}

func (c *OrderConsumer) NackMessage(deliveryTag uint64, requeue bool) error {
	return c.channel.Nack(deliveryTag, false, requeue)
}
