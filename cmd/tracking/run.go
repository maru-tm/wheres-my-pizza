package tracking

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pg *pgxpool.Pool, rmq *rabbitmq.RabbitMQ, port int, rid string) {}
