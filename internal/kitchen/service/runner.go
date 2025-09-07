package service

import (
	"context"
	"restaurant-system/internal/kitchen/infrastructure/rmq"
)

type OrderConsumer interface {
	Consume(ctx context.Context) (*rmq.OrderMessage, error)
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
}

type StatusPublisher interface {
	PublishUpdatedStatus(ctx context.Context, msg rmq.StatusUpdateMessage) error
}
type KitchenWorker struct {
	service         *KitchenService
	orderConsumer   OrderConsumer
	statusPublisher StatusPublisher
	workerName      string
}

func NewKitchenWorker(service *KitchenService, orderConsumer OrderConsumer, statusPublisher StatusPublisher, workerName string) *KitchenWorker {
	return &KitchenWorker{
		service:         service,
		orderConsumer:   orderConsumer,
		statusPublisher: statusPublisher,
		workerName:      workerName,
	}
}
