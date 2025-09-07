package rmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"restaurant-system/internal/rabbitmq"
)

type OrderConsumer struct {
	ch         *amqp091.Channel
	queue      string
	exchange   string
	routingKey string
	workerName string
}

func NewOrderConsumer(r *rabbitmq.RabbitMQ, workerName string) (*OrderConsumer, error) {
	exchange := "orders_exchange"
	routingKey := "orders.created"

	ch := r.Channel()

	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}
	args := amqp091.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "orders.dead",
	}
	queueName := fmt.Sprintf("orders_queue_%s", workerName)
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		return nil, err
	}

	if err := ch.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		return nil, err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}

	return &OrderConsumer{
		ch:         ch,
		queue:      q.Name,
		exchange:   exchange,
		routingKey: routingKey,
		workerName: workerName,
	}, nil
}

func (c *OrderConsumer) Consume(ctx context.Context) (*OrderMessage, error) {
	// TODO: implement consumer
	return nil, nil
}

func (c *OrderConsumer) Ack(deliveryTag uint64) error {
	return nil
}
func (c *OrderConsumer) Nack(deliveryTag uint64, requeue bool) error {
	return nil
}
