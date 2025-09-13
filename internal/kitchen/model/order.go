package model

import "time"

type Order struct {
	ID              int        `json:"id"`
	Number          string     `json:"number"`
	CustomerName    string     `json:"customer_name"`
	Type            string     `json:"type"`
	TableNumber     *int       `json:"table_number,omitempty"`
	DeliveryAddress *string    `json:"delivery_address,omitempty"`
	TotalAmount     float64    `json:"total_amount"`
	Priority        int        `json:"priority"`
	Status          string     `json:"status"`
	ProcessedBy     *string    `json:"processed_by,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}
