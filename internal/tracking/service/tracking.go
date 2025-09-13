package service

import (
	"context"
	"fmt"
	"time"

	"restaurant-system/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TrackingService struct {
	db *pgxpool.Pool
}

func NewTrackingService(db *pgxpool.Pool) *TrackingService {
	return &TrackingService{db: db}
}

func (s *TrackingService) GetOrderStatus(ctx context.Context, orderNumber string) (map[string]interface{}, error) {
	query := `
		SELECT number, status, updated_at, processed_by,
			   CASE 
				   WHEN status = 'cooking' THEN updated_at + INTERVAL '10 minutes'
				   ELSE NULL 
			   END as estimated_completion
		FROM orders 
		WHERE number = $1
	`

	var orderNumberDB string
	var status string
	var updatedAt time.Time
	var processedBy *string
	var estimatedCompletion *time.Time

	err := s.db.QueryRow(ctx, query, orderNumber).Scan(
		&orderNumberDB,
		&status,
		&updatedAt,
		&processedBy,
		&estimatedCompletion,
	)
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "db_query_failed", "failed to query order status", "",
			map[string]interface{}{"order_number": orderNumber}, err)
		return nil, fmt.Errorf("order not found")
	}

	result := make(map[string]interface{})
	result["order_number"] = orderNumberDB
	result["current_status"] = status
	result["updated_at"] = updatedAt.Format(time.RFC3339)
	if processedBy != nil {
		result["processed_by"] = *processedBy
	}
	if estimatedCompletion != nil {
		result["estimated_completion"] = estimatedCompletion.Format(time.RFC3339)
	}

	return result, nil
}

func (s *TrackingService) GetOrderHistory(ctx context.Context, orderNumber string) ([]map[string]interface{}, error) {
	query := `
		SELECT osl.status, osl.changed_at, osl.changed_by
		FROM order_status_log osl
		JOIN orders o ON osl.order_id = o.id
		WHERE o.number = $1
		ORDER BY osl.changed_at ASC
	`

	rows, err := s.db.Query(ctx, query, orderNumber)
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "db_query_failed", "failed to query order history", "",
			map[string]interface{}{"order_number": orderNumber}, err)
		return nil, fmt.Errorf("failed to get order history")
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var status, changedBy string
		var changedAt time.Time

		if err := rows.Scan(&status, &changedAt, &changedBy); err != nil {
			return nil, err
		}

		history = append(history, map[string]interface{}{
			"status":     status,
			"timestamp":  changedAt.Format(time.RFC3339),
			"changed_by": changedBy,
		})
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return history, nil
}

func (s *TrackingService) GetWorkersStatus(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT name, status, orders_processed, last_seen,
			   CASE 
				   WHEN NOW() - last_seen > INTERVAL '60 seconds' THEN 'offline'
				   ELSE status
			   END as current_status
		FROM workers
		ORDER BY name
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "db_query_failed", "failed to query workers status", "", nil, err)
		return nil, fmt.Errorf("failed to get workers status")
	}
	defer rows.Close()

	var workers []map[string]interface{}
	for rows.Next() {
		var name, status string
		var ordersProcessed int
		var lastSeen time.Time
		var currentStatus string

		if err := rows.Scan(&name, &status, &ordersProcessed, &lastSeen, &currentStatus); err != nil {
			return nil, err
		}

		workers = append(workers, map[string]interface{}{
			"worker_name":      name,
			"status":           currentStatus,
			"orders_processed": ordersProcessed,
			"last_seen":        lastSeen.Format(time.RFC3339),
		})
	}

	return workers, nil
}
