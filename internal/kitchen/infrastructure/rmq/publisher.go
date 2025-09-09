package rmq

import (
	"context"
	"encoding/json"
	"time"

	"restaurant-system/internal/rabbitmq"
)

type StatusPublisher struct {
	rmq      *rabbitmq.RabbitMQ
	exchange string
}

func NewStatusPublisher(r *rabbitmq.RabbitMQ) (*StatusPublisher, error) {
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

	return &StatusPublisher{
		rmq:      r,
		exchange: exchange,
	}, nil
}

func (p *StatusPublisher) PublishUpdatedStatus(ctx context.Context, msg StatusUpdateMessage) error {
	event := map[string]interface{}{
		"order_number":         msg.OrderNumber,
		"old_status":           msg.OldStatus,
		"new_status":           msg.NewStatus,
		"changed_by":           msg.ChangedBy,
		"timestamp":            time.Now().UTC(),
		"estimated_completion": msg.EstimatedCompletion,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.rmq.Publish(ctx, p.exchange, "", body)
}
