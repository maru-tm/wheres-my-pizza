package rmq

import "restaurant-system/internal/order/model"

type OrderMessage struct {
	OrderNumber     string            `json:"order_number"`
	CustomerName    string            `json:"customer_name"`
	OrderType       string            `json:"order_type"`
	TableNumber     *int              `json:"table_number,omitempty"`
	DeliveryAddress *string           `json:"delivery_address,omitempty"`
	Items           []model.OrderItem `json:"items"`
	TotalAmount     float64           `json:"total_amount"`
	Priority        int               `json:"priority"`
	DeliveryTag     uint64            `json:"-"`
}
