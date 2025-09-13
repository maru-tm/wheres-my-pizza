package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ValidationError = errors.New("validation error")

func (o *Order) Validate() error {
	var validationErrors []string

	if o.CustomerName == "" {
		validationErrors = append(validationErrors, "customer_name is required")
	} else if utf8.RuneCountInString(o.CustomerName) > 100 {
		validationErrors = append(validationErrors, "customer_name must be 100 characters or less")
	} else if !isValidName(o.CustomerName) {
		validationErrors = append(validationErrors, "customer_name contains invalid characters")
	}

	switch o.Type {
	case OrderTypeDineIn, OrderTypeTakeout, OrderTypeDelivery:
	default:
		validationErrors = append(validationErrors, "order_type must be one of: dine_in, takeout, delivery")
	}

	switch o.Type {
	case OrderTypeDineIn:
		if o.TableNumber == nil {
			validationErrors = append(validationErrors, "table_number is required for dine_in orders")
		} else if *o.TableNumber < 1 || *o.TableNumber > 100 {
			validationErrors = append(validationErrors, "table_number must be between 1 and 100")
		}
		if o.DeliveryAddress != nil {
			validationErrors = append(validationErrors, "delivery_address must not be present for dine_in orders")
		}

	case OrderTypeDelivery:
		if o.DeliveryAddress == nil {
			validationErrors = append(validationErrors, "delivery_address is required for delivery orders")
		} else if utf8.RuneCountInString(*o.DeliveryAddress) < 10 {
			validationErrors = append(validationErrors, "delivery_address must be at least 10 characters")
		}
		if o.TableNumber != nil {
			validationErrors = append(validationErrors, "table_number must not be present for delivery orders")
		}

	case OrderTypeTakeout:
		if o.TableNumber != nil {
			validationErrors = append(validationErrors, "table_number must not be present for takeout orders")
		}
		if o.DeliveryAddress != nil {
			validationErrors = append(validationErrors, "delivery_address must not be present for takeout orders")
		}
	}

	if len(o.Items) == 0 {
		validationErrors = append(validationErrors, "items must contain at least 1 item")
	} else if len(o.Items) > 20 {
		validationErrors = append(validationErrors, "items cannot contain more than 20 items")
	}

	for i, item := range o.Items {
		if item.Name == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("items[%d].name is required", i))
		} else if utf8.RuneCountInString(item.Name) > 50 {
			validationErrors = append(validationErrors, fmt.Sprintf("items[%d].name must be 50 characters or less", i))
		}

		if item.Quantity < 1 || item.Quantity > 10 {
			validationErrors = append(validationErrors, fmt.Sprintf("items[%d].quantity must be between 1 and 10", i))
		}

		if item.Price < 0.01 || item.Price > 999.99 {
			validationErrors = append(validationErrors, fmt.Sprintf("items[%d].price must be between 0.01 and 999.99", i))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("%w: %s", ValidationError, strings.Join(validationErrors, "; "))
	}

	return nil
}

func isValidName(name string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, name)
	return match
}
