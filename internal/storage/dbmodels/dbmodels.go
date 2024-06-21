package dbmodels

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	Login        string
	PasswordHash string
}

type UserBalance struct {
	Login     string
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}

type UserWithdrawal struct {
	OrderID     string
	Login       string
	Amount      decimal.Decimal
	ProcessedAt time.Time
}

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	Accrual    decimal.Decimal
	UploadedAt time.Time
}
