package handler

type CreateOrderRequest struct {
	CustomerName string         `json:"customer_name"`
	OrderType    string         `json:"order_type"`
	Items        []ItemsRequest `json:"items"`
}

type ItemsRequest struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type CreateOrderResponse struct {
	OrderNumber string  `json:"order_number"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
}
