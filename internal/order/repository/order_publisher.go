package repository

import (
	"context"
	"encoding/json"
	"restaurant-system/internal/order/model"
	"restaurant-system/internal/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type OrderPublisher struct {
	ch         *amqp091.Channel
	exchange   string
	routingKey string
}

func NewOrderPublisher(r *rabbitmq.RabbitMQ) (*OrderPublisher, error) {
	exchange := "orders_exchange"
	routingKey := "orders.created"

	ch := r.Channel()

	err := ch.ExchangeDeclare(
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

	return &OrderPublisher{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *OrderPublisher) PublishOrder(ctx context.Context, order *model.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx,
		p.exchange,   // exchange
		p.routingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
