package tracking

import (
	"context"
	"restaurant-system/internal/db"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pg *db.DB, rmq *rabbitmq.RabbitMQ, port int, rid string) {}
