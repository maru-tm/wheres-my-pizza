package model

import (
	"errors"
	"fmt"
	"regexp"
)

var ValidationError = errors.New("validation error")
var validName = regexp.MustCompile(`^[a-zA-Z\s'-]+$`)

func (o *Order) Validate() error {
	if len(o.CustomerName) < 1 || len(o.CustomerName) > 100 {
		return errors.New("customer_name must be 1-100 characters")
	}
	if !validName.MatchString(o.CustomerName) {
		return errors.New("customer_name contains invalid characters")
	}

	switch o.Type {
	case OrderTypeDineIn:
		if o.TableNumber == nil || *o.TableNumber < 1 || *o.TableNumber > 100 {
			return fmt.Errorf("table_number must be set and between 1 and 100 for dine_in orders")
		}
		if o.DeliveryAddress != nil {
			return fmt.Errorf("delivery_address must not be present for dine_in orders")
		}
	case OrderTypeDelivery:
		if o.DeliveryAddress == nil || len(*o.DeliveryAddress) < 10 {
			return fmt.Errorf("delivery_address must be set and at least 10 characters for delivery orders")
		}
		if o.TableNumber != nil {
			return fmt.Errorf("table_number must not be present for delivery orders")
		}
	case OrderTypeTakeout:
		if o.TableNumber != nil {
			return fmt.Errorf("table_number must not be present for takeout orders")
		}
		if o.DeliveryAddress != nil {
			return fmt.Errorf("delivery_address must not be present for takeout orders")
		}
	default:
		return fmt.Errorf("unsupported order type: %s", o.Type)
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
