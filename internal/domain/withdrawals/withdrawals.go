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
	orderNumber string
	amount      decimal.Decimal
	processedAt time.Time
}

func NewWithdrawal(userLogin, orderNumber string, amount decimal.Decimal) (*Withdrawal, error) {
	if err := users.ValidateLogin(userLogin); err != nil {
		return nil, err
	}

	if err := orders.ValidateOrderNumber(orderNumber); err != nil {
		return nil, err
	}

	return &Withdrawal{
		userLogin:   userLogin,
		orderNumber: orderNumber,
		amount:      amount,
		processedAt: time.Now(),
	}, nil
}

func (w *Withdrawal) UserLogin() string {
	return w.userLogin
}

func (w *Withdrawal) OrderNumber() string {
	return w.orderNumber
}

func (w *Withdrawal) Amount() decimal.Decimal {
	return w.amount
}

func (w *Withdrawal) ProcessedAt() time.Time {
	return w.processedAt
}
