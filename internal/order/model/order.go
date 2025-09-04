package model

import (
	"errors"
	"fmt"
	"regexp"
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
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
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
	Items           []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	OrderID   int       `json:"order_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
}

type OrderStatusLog struct {
	ID        int         `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	OrderID   int         `json:"order_id"`
	Status    OrderStatus `json:"status"`
	ChangedBy string      `json:"changed_by"`
	ChangedAt time.Time   `json:"changed_at"`
	Notes     *string     `json:"notes,omitempty"`
}

func (o *Order) Validate() error {
	if len(o.CustomerName) < 1 || len(o.CustomerName) > 100 {
		return errors.New("customer_name must be 1-100 characters")
	}
	validName := regexp.MustCompile(`^[a-zA-Z\s'-]+$`)
	if !validName.MatchString(o.CustomerName) {
		return errors.New("customer_name contains invalid characters")
	}

	switch o.Type {
	case OrderTypeDineIn, OrderTypeTakeout, OrderTypeDelivery:
	default:
		return fmt.Errorf("invalid order_type: %s", o.Type)
	}

	if len(o.Items) < 1 || len(o.Items) > 20 {
		return errors.New("items must contain between 1 and 20 items")
	}

	for _, item := range o.Items {
		if len(item.Name) < 1 || len(item.Name) > 50 {
			return fmt.Errorf("item name must be 1-50 characters: %s", item.Name)
		}
		if item.Quantity < 1 || item.Quantity > 10 {
			return fmt.Errorf("item quantity must be between 1 and 10: %d", item.Quantity)
		}
		if item.Price < 0.01 || item.Price > 999.99 {
			return fmt.Errorf("item price must be between 0.01 and 999.99: %.2f", item.Price)
		}
	}

	return nil
}
