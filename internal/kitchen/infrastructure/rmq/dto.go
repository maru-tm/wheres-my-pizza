package rmq

import (
	"time"
)

type OrderMessage struct {
	OrderNumber     string  `json:"order_number"`
	CustomerName    string  `json:"customer_name"`
	OrderType       string  `json:"order_type"`
	TableNumber     *int    `json:"table_number,omitempty"`
	DeliveryAddress *string `json:"delivery_address,omitempty"`
	Items           []struct {
		Name     string  `json:"name"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
	} `json:"items"`
	TotalAmount float64 `json:"total_amount"`
	Priority    int     `json:"priority"`
	DeliveryTag uint64  `json:"-"`
}

type StatusUpdateMessage struct {
	OrderNumber         string    `json:"order_number"`
	OldStatus           string    `json:"old_status"`
	NewStatus           string    `json:"new_status"`
	ChangedBy           string    `json:"changed_by"`
	Timestamp           time.Time `json:"timestamp"`
	EstimatedCompletion time.Time `json:"estimated_completion"`
	DeliveryTag         uint64    `json:"-"`
}
