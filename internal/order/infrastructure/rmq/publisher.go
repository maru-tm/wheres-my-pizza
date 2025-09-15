package rmq

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-system/internal/order/model"
	"restaurant-system/pkg/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type OrderPublisher struct {
	rabbitmq *rabbitmq.RabbitMQ
}

func NewOrderPublisher(rabbitmq *rabbitmq.RabbitMQ) (*OrderPublisher, error) {
	ch := rabbitmq.Channel()

	err := ch.ExchangeDeclare(
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

	return &OrderPublisher{rabbitmq: rabbitmq}, nil
}

func (p *OrderPublisher) PublishCreatedOrder(ctx context.Context, order *model.Order) error {
	message := OrderMessage{
		OrderNumber:     order.Number,
		CustomerName:    order.CustomerName,
		OrderType:       string(order.Type),
		TableNumber:     order.TableNumber,
		DeliveryAddress: order.DeliveryAddress,
		Items:           order.Items,
		TotalAmount:     order.TotalAmount,
		Priority:        order.Priority,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order message: %w", err)
	}

	routingKey := fmt.Sprintf("kitchen.%s.%d", order.Type, order.Priority)

	err = p.rabbitmq.Channel().PublishWithContext(ctx,
		"orders_topic",
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Priority:     uint8(order.Priority),
		})
	if err != nil {
		return fmt.Errorf("failed to publish order: %w", err)
	}

	return nil
}
