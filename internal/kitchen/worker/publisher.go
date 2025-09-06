package worker

import (
	"context"
	"encoding/json"
	"time"

	"restaurant-system/internal/rabbitmq"
)

type Publisher struct {
	rmq      *rabbitmq.RabbitMQ
	exchange string
}

func NewPublisher(r *rabbitmq.RabbitMQ) (*Publisher, error) {
	exchange := "notifications_fanout"

	err := r.Channel().ExchangeDeclare(
		exchange, // name
		"fanout", // type
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
		rmq:      r,
		exchange: exchange,
	}, nil
}

func (p *Publisher) PublishStatusUpdate(ctx context.Context, orderID int, oldStatus, newStatus, worker string) error {
	event := map[string]interface{}{
		"order_id":   orderID,
		"old_status": oldStatus,
		"new_status": newStatus,
		"changed_by": worker,
		"timestamp":  time.Now().UTC(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.rmq.Publish(ctx, p.exchange, "", body)
}
