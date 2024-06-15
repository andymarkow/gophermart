package storage

import (
	"context"
	"errors"

	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/andymarkow/gophermart/internal/domain/withdrawals"
	"github.com/shopspring/decimal"
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

type Storage interface { //nolint:interfacebloat
	Close() error
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, usr *users.User) error
	GetUser(ctx context.Context, login string) (*users.User, error)
	GetUserBalance(ctx context.Context, login string) (*users.UserBalance, error)
	DepositUserBalance(ctx context.Context, login string, amount decimal.Decimal) error
	WithdrawUserBalance(ctx context.Context, withdrawal *withdrawals.Withdrawal) error
	GetWithdrawalsByUserLogin(ctx context.Context, login string) ([]*withdrawals.Withdrawal, error)
	CreateOrder(ctx context.Context, ord *orders.Order) error
	GetOrder(ctx context.Context, number string) (*orders.Order, error)
	GetOrdersByUserLogin(ctx context.Context, login string) ([]*orders.Order, error)
}

func NewStorage(store Storage) Storage {
	return store
}
