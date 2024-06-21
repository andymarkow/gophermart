package storage

import (
	"context"
	"errors"

	"github.com/andymarkow/gophermart/internal/domain/balance"
	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/andymarkow/gophermart/internal/domain/withdrawals"
)

var (
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrUserNotFound               = errors.New("user not found")
	ErrOrderAlreadyExists         = errors.New("order already exists")
	ErrOrderNotFound              = errors.New("order not found")
	ErrUserBalanceAlreadyExists   = errors.New("user balance already exists")
	ErrUserBalanceNotFound        = errors.New("user balance not found")
	ErrUserBalanceNotEnough       = errors.New("user balance not enough")
	ErrBalanceWithdrawalsNotFound = errors.New("balance withdrawals not found")
)

type UserStorage interface {
	GetUser(ctx context.Context, login string) (*users.User, error)
	CreateUser(ctx context.Context, usr *users.User) error
}

type UserBalanceStorage interface {
	GetUserBalance(ctx context.Context, login string) (*balance.Balance, error)
	WithdrawUserBalance(ctx context.Context, withdrawal *withdrawals.Withdrawal) error
	GetWithdrawalsByUserLogin(ctx context.Context, login string) ([]*withdrawals.Withdrawal, error)
}

type OrderStorage interface {
	GetOrder(ctx context.Context, number string) (*orders.Order, error)
	GetOrdersByLogin(ctx context.Context, login string) ([]*orders.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses ...orders.OrderStatus) ([]*orders.Order, error)
	CreateOrder(ctx context.Context, order *orders.Order) error
	ProcessOrderAccrual(ctx context.Context, order *orders.Order) error
}

type Storage interface {
	UserStorage
	UserBalanceStorage
	OrderStorage
	Close() error
	Ping(ctx context.Context) error
}

func NewStorage(store Storage) Storage {
	return store
}
