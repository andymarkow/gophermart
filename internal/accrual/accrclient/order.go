package accrclient

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusRegistered OrderStatus = "REGISTERED"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderModel struct {
	Number  string          `json:"number"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func parseOrderStatus(status string) (OrderStatus, error) {
	switch status {
	case "REGISTERED":
		return OrderStatusRegistered, nil
	case "INVALID":
		return OrderStatusInvalid, nil
	case "PROCESSING":
		return OrderStatusProcessing, nil
	case "PROCESSED":
		return OrderStatusProcessed, nil
	default:
		return "", fmt.Errorf("unknown order status: %s", status)
	}
}

type Order struct {
	number  string
	status  OrderStatus
	accrual decimal.Decimal
}

func newOrder(number string, status string, accrual decimal.Decimal) (*Order, error) {
	orderStatus, err := parseOrderStatus(status)
	if err != nil {
		return nil, err
	}

	return &Order{
		number:  number,
		status:  orderStatus,
		accrual: accrual,
	}, nil
}

func (o *Order) Number() string {
	return o.number
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) Accrual() decimal.Decimal {
	return o.accrual
}
