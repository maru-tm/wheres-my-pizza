package worker

//
//import (
//	"context"
//	"fmt"
//	"github.com/jackc/pgx/v5"
//	"restaurant-system/internal/logger"
//	"restaurant-system/internal/rabbitmq"
//	"time"
//)
//
//type KitchenWorker struct {
//	Name       string
//	OrderTypes []string
//	Pg         *pgx.Conn
//	Rmq        *rabbitmq.RabbitMQ
//	Prefetch   int
//	Heartbeat  time.Duration
//	quit       chan struct{}
//}
//
//func NewKitchenWorker(name string, types []string, pg *pgx.Conn, rmq *rabbitmq.RabbitMQ, prefetch int, heartbeat time.Duration) *KitchenWorker {
//	return &KitchenWorker{
//		Name:       name,
//		OrderTypes: types,
//		Pg:         pg,
//		Rmq:        rmq,
//		Prefetch:   prefetch,
//		Heartbeat:  heartbeat,
//		quit:       make(chan struct{}),
//	}
//}
//
//// Register worker in DB
//func (w *KitchenWorker) Register(ctx context.Context) error {
//	var status string
//	err := w.Pg.QueryRow(ctx, "SELECT status FROM workers WHERE name=$1", w.Name).Scan(&status)
//	if err == nil && status == "online" {
//		return fmt.Errorf("worker %s already online", w.Name)
//	}
//
//	if err := w.Pg.QueryRow(ctx, `
//	INSERT INTO workers(name, type, status, last_seen)
//	VALUES($1, $2, 'online', NOW())
//	ON CONFLICT(name) DO UPDATE SET status='online', last_seen=NOW() RETURNING id
//	`, w.Name, "chef").Scan(new(int)); err != nil {
//		return err
//	}
//
//	logger.Log(logger.INFO, "kitchen-worker", "worker_registered", "worker successfully registered", w.Name, nil, nil)
//	return nil
//}
//
//// Start consuming messages
//func (w *KitchenWorker) Start(ctx context.Context) {
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-w.quit:
//			return
//		default:
//			// TODO: consume message from RabbitMQ, process order
//		}
//	}
//}
//
//// Graceful shutdown
//func (w *KitchenWorker) Shutdown() {
//	close(w.quit)
//	// TODO: update status offline, nack unack messages
//}
