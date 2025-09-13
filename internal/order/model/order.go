package model

import (
	"time"
)

type OrderType string

const (
	OrderTypeDineIn   OrderType = "dine_in"
	OrderTypeTakeout  OrderType = "takeout"
	OrderTypeDelivery OrderType = "delivery"
)

type OrderStatus string

const (
	StatusReceived  OrderStatus = "received"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              int         `json:"id"`
	Number          string      `json:"number"`
	CustomerName    string      `json:"customer_name"`
	Type            OrderType   `json:"type"`
	TableNumber     *int        `json:"table_number,omitempty"`
	DeliveryAddress *string     `json:"delivery_address,omitempty"`
	TotalAmount     float64     `json:"total_amount"`
	Priority        int         `json:"priority"`
	Status          OrderStatus `json:"status"`
	ProcessedBy     *string     `json:"processed_by,omitempty"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderStatusLog struct {
	ID        int         `json:"id"`
	OrderID   int         `json:"order_id"`
	Status    OrderStatus `json:"status"`
	ChangedBy string      `json:"changed_by"`
	ChangedAt time.Time   `json:"changed_at"`
	Notes     *string     `json:"notes,omitempty"`
}

type CreateOrderRequest struct {
	CustomerName    string             `json:"customer_name"`
	OrderType       OrderType          `json:"order_type"`
	TableNumber     *int               `json:"table_number,omitempty"`
	DeliveryAddress *string            `json:"delivery_address,omitempty"`
	Items           []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}
