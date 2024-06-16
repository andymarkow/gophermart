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
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type BalanceWithdrawalRequest struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
}

type BalanceWithdrawalResponse struct {
	OrderNumber string  `json:"order"`
	Amount      float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type OrderResponse struct {
	Number     string             `json:"number"`
	Status     orders.OrderStatus `json:"status"`
	Accrual    float64            `json:"accrual,omitempty"`
	UploadedAt string             `json:"uploaded_at"`
}
