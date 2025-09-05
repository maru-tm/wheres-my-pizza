package kitchen

import (
	"context"
	"restaurant-system/internal/db"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pg *db.DB, rmq *rabbitmq.RabbitMQ, name, types string, prefetch, heartbeat int, rid string) {
}
