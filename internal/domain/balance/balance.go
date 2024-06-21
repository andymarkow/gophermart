package balance

import (
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/shopspring/decimal"
)

type Balance struct {
	userLogin string
	current   decimal.Decimal
	withdrawn decimal.Decimal
}

func NewBalance(userLogin string, current, withdrawn decimal.Decimal) (*Balance, error) {
	if err := users.ValidateLogin(userLogin); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &Balance{
		userLogin: userLogin,
		current:   current,
		withdrawn: withdrawn,
	}, nil
}

func (b *Balance) UserLogin() string {
	return b.userLogin
}

func (b *Balance) Current() decimal.Decimal {
	return b.current
}

func (b *Balance) Withdrawn() decimal.Decimal {
	return b.withdrawn
}

func (b *Balance) SetCurrent(current decimal.Decimal) {
	b.current = current
}

func (b *Balance) SetWithdrawn(withdrawn decimal.Decimal) {
	b.withdrawn = withdrawn
}

func (b *Balance) AddCurrent(amount decimal.Decimal) {
	b.current = b.current.Add(amount)
}

func (b *Balance) SubCurrent(amount decimal.Decimal) {
	b.current = b.current.Sub(amount)
}

func (b *Balance) AddWithdrawn(amount decimal.Decimal) {
	b.withdrawn = b.withdrawn.Add(amount)
}
