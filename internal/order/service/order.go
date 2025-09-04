package service

import (
	"restaurant-system/internal/order/repository"
	"restaurant-system/internal/rabbitmq"
)

type PGInterface interface {
}
type RabbitMQInterface interface{}
type OrderService struct {
	repo *repository.OrderRepository
	rmq  *rabbitmq.RabbitMQ
}

func NewOrderService(r *repository.OrderRepository, rmq *rabbitmq.RabbitMQ) *OrderService {
	return &OrderService{repo: r, rmq: rmq}
}
