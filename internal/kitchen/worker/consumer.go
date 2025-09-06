package worker

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"restaurant-system/internal/rabbitmq"
)

type Consumer struct {
	ch         *amqp091.Channel
	queue      string
	exchange   string
	routingKey string
}

func NewConsumer(r *rabbitmq.RabbitMQ, workerName string) (*Consumer, error) {
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

	queueName := fmt.Sprintf("orders_queue_%s", workerName)
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
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

	return &Consumer{
		ch:         ch,
		queue:      q.Name,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}
