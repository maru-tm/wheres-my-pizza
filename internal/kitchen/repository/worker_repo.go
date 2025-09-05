package repository

//import (
//	"context"
//	"errors"
//	"github.com/jackc/pgx/v5/pgxpool"
//	"restaurant-system/internal/kitchen/model"
//)
//
//var (
//	ErrWorkerAlreadyOnline = errors.New("worker already online")
//)
//
//type WorkerRepository struct {
//	db *pgxpool.Pool
//}
//
//func NewWorkerRepository(db *pgxpool.Pool) *WorkerRepository {
//	return &WorkerRepository{db: db}
//}
//
//func (r *WorkerRepository) RegisterWorker(ctx context.Context, w *model.Worker) error {
//	var id int
//	var status string
//	err := r.db.QueryRow(ctx,
//		`SELECT id, status FROM workers WHERE name=$1`, w.Name).Scan(&id, &status)
//	if err == nil {
//		if status == "online" {
//			return ErrWorkerAlreadyOnline
//		}
//		_, err := r.db.Exec(ctx,
//			`UPDATE workers SET type=$1, status='online', last_seen=NOW() WHERE id=$2`,
//			w.Type, id)
//		return err
//	}
//
//	_, err = r.db.Exec(ctx,
//		`INSERT INTO workers (name, type, status, last_seen, orders_processed) VALUES ($1, $2, 'online', NOW(), 0)`,
//		w.Name, w.Type)
//	return err
//}
//
//func (r *WorkerRepository) UpdateHeartbeat(ctx context.Context, name string) error {
//	_, err := r.db.Exec(ctx,
//		`UPDATE workers SET last_seen=NOW(), status='online' WHERE name=$1`, name)
//	return err
//}
//
//func (r *WorkerRepository) SetOffline(ctx context.Context, name string) error {
//	_, err := r.db.Exec(ctx,
//		`UPDATE workers SET status='offline' WHERE name=$1`, name)
//	return err
//}
//
//func (r *WorkerRepository) IncrementOrdersProcessed(ctx context.Context, name string) error {
//	_, err := r.db.Exec(ctx,
//		`UPDATE workers SET orders_processed = orders_processed + 1 WHERE name=$1`, name)
//	return err
//}
//
//func (r *WorkerRepository) GetWorkerByName(ctx context.Context, name string) (*model.Worker, error) {
//	w := &model.Worker{}
//	err := r.db.QueryRow(ctx,
//		`SELECT id, created_at, name, type, status, last_seen, orders_processed
//         FROM workers WHERE name=$1`, name).
//		Scan(&w.ID, &w.CreatedAt, &w.Name, &w.Type, &w.Status, &w.LastSeen, &w.OrdersProcessed)
//	if err != nil {
//		return nil, err
//	}
//	return w, nil
//}
