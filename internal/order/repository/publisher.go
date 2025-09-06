package repository

import (
	"context"
	"encoding/json"
	"restaurant-system/internal/order/model"
	"restaurant-system/internal/rabbitmq"
)

type Publisher struct {
	rmq        *rabbitmq.RabbitMQ
	exchange   string
	routingKey string
}

func NewPublisher(r *rabbitmq.RabbitMQ) (*Publisher, error) {
	exchange := "orders_exchange"
	routingKey := "orders.created"

	err := r.Channel().ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		rmq:        r,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *Publisher) PublishOrder(ctx context.Context, order *model.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return p.rmq.Publish(ctx, p.exchange, p.routingKey, body)
}
