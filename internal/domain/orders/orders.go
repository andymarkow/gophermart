//nolint:wrapcheck
package orders

import (
	"errors"
	"time"
	"unicode"

	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/shopspring/decimal"
)

var (
	ErrOrderIDEmpty         = errors.New("order ID is empty")
	ErrOrderIDFormatInvalid = errors.New("order ID format is invalid")
)

type OrderStatus string

func (s OrderStatus) String() string {
	return string(s)
}

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	id         string
	userLogin  string
	status     OrderStatus
	accrual    decimal.Decimal
	uploadedAt time.Time
}

func NewOrder(
	id, userLogin string, status OrderStatus, accrual decimal.Decimal, uploadedAt time.Time,
) (*Order, error) {
	if err := ValidateOrderID(id); err != nil {
		return nil, err
	}

	if err := validateUserLogin(userLogin); err != nil {
		return nil, err
	}

	return &Order{
		id:         id,
		userLogin:  userLogin,
		status:     status,
		accrual:    accrual,
		uploadedAt: uploadedAt,
	}, nil
}

func CreateOrder(id string, userLogin string) (*Order, error) {
	return NewOrder(id, userLogin, OrderStatusNew, decimal.Zero, time.Now())
}

func (o *Order) ID() string {
	return o.id
}

func (o *Order) UserLogin() string {
	return o.userLogin
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) Accrual() decimal.Decimal {
	return o.accrual
}

func (o *Order) UploadedAt() time.Time {
	return o.uploadedAt
}

func (o *Order) SetStatus(status OrderStatus) {
	o.status = status
}

func (o *Order) SetAccrual(accrual decimal.Decimal) {
	o.accrual = accrual
}

func ValidateOrderID(id string) error {
	if id == "" {
		return ErrOrderIDEmpty
	}

	if !validateByLuhn(id) {
		return ErrOrderIDFormatInvalid
	}

	return nil
}

func validateUserLogin(userLogin string) error {
	return users.ValidateLogin(userLogin)
}

// validateByLuhn checks id is valid or not based on Luhn algorithm.
func validateByLuhn(id string) bool {
	var sum int
	double := false

	for i := len(id) - 1; i >= 0; i-- {
		n := id[i]

		if !unicode.IsDigit(rune(n)) {
			return false // invalid character
		}

		digit := int(n - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit

		double = !double
	}

	return sum%10 == 0
}
