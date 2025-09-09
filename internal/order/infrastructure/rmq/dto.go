package rmq

import "restaurant-system/internal/order/model"

type OrderMessage struct {
	OrderNumber     string         `json:"order_number"`
	CustomerName    string         `json:"customer_name"`
	OrderType       string         `json:"order_type"`
	TableNumber     *int           `json:"table_number,omitempty"`
	DeliveryAddress *string        `json:"delivery_address,omitempty"`
	Items           []OrderItemMsg `json:"items"`
	TotalAmount     float64        `json:"total_amount"`
	Priority        int            `json:"priority"`
}

type OrderItemMsg struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func NewOrderMessage(order *model.Order) OrderMessage {
	msg := OrderMessage{
		OrderNumber:     order.Number,
		CustomerName:    order.CustomerName,
		OrderType:       string(order.Type),
		TableNumber:     order.TableNumber,
		DeliveryAddress: order.DeliveryAddress,
		TotalAmount:     order.TotalAmount,
		Priority:        order.Priority,
	}

	for _, item := range order.Items {
		msg.Items = append(msg.Items, OrderItemMsg{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	return msg
}
