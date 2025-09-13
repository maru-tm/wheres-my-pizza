package model

import "time"

type Worker struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	LastSeen        time.Time `json:"last_seen"`
	OrdersProcessed int       `json:"orders_processed"`
	OrderTypes      []string  `json:"order_types,omitempty"`
}
