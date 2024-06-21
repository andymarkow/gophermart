//nolint:wrapcheck
package withdrawals

import (
	"time"

	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	userLogin   string
	orderID     string
	amount      decimal.Decimal
	processedAt time.Time
}

func NewWithdrawal(userLogin, orderID string, amount decimal.Decimal, processedAt time.Time) (*Withdrawal, error) {
	if err := users.ValidateLogin(userLogin); err != nil {
		return nil, err
	}

	if err := orders.ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	return &Withdrawal{
		userLogin:   userLogin,
		orderID:     orderID,
		amount:      amount,
		processedAt: time.Now(),
	}, nil
}

func CreateWithdrawal(userLogin, orderID string, amount decimal.Decimal) (*Withdrawal, error) {
	return NewWithdrawal(userLogin, orderID, amount, time.Now())
}

func (w *Withdrawal) UserLogin() string {
	return w.userLogin
}

func (w *Withdrawal) OrderID() string {
	return w.orderID
}

func (w *Withdrawal) Amount() decimal.Decimal {
	return w.amount
}

func (w *Withdrawal) ProcessedAt() time.Time {
	return w.processedAt
}
