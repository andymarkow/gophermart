package models

import (
	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/shopspring/decimal"
)

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserBalanceResponse struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type BalanceWithdrawalRequest struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
}

type BalanceWithdrawalResponse struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
	ProcessedAt string          `json:"processed_at"`
}

type OrderResponse struct {
	Number     string             `json:"number"`
	Status     orders.OrderStatus `json:"status"`
	Accrual    int                `json:"accrual,omitempty"`
	UploadedAt string             `json:"uploaded_at"`
}
