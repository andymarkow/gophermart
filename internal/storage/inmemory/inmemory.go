package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/andymarkow/gophermart/internal/domain/withdrawals"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/shopspring/decimal"
)

var _ storage.Storage = (*Storage)(nil)

type UserStore struct {
	users map[string]*users.User
	mu    sync.Mutex
}

type UserBalanceStore struct {
	balances map[string]*users.UserBalance
	mu       sync.Mutex
}

type UserWithdrawalStore struct {
	withdrawals map[string][]*withdrawals.Withdrawal
	mu          sync.Mutex
}

type OrderStore struct {
	orders map[string]*orders.Order
	mu     sync.Mutex
}

type Storage struct {
	UserStore           UserStore
	UserBalanceStore    UserBalanceStore
	UserWithdrawalStore UserWithdrawalStore
	OrderStore          OrderStore
}

func NewStorage() *Storage {
	return &Storage{
		UserStore: UserStore{
			users: make(map[string]*users.User),
		},
		UserBalanceStore: UserBalanceStore{
			balances: make(map[string]*users.UserBalance),
		},
		UserWithdrawalStore: UserWithdrawalStore{
			withdrawals: make(map[string][]*withdrawals.Withdrawal),
		},
		OrderStore: OrderStore{
			orders: make(map[string]*orders.Order),
		},
	}
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Ping(_ context.Context) error {
	return nil
}

func (s *Storage) CreateUser(_ context.Context, usr *users.User) error {
	s.UserStore.mu.Lock()
	defer s.UserStore.mu.Unlock()

	s.UserBalanceStore.mu.Lock()
	defer s.UserBalanceStore.mu.Unlock()

	if _, ok := s.UserStore.users[usr.Login]; ok {
		return storage.ErrUserAlreadyExists
	}

	s.UserStore.users[usr.Login] = usr

	if _, ok := s.UserBalanceStore.balances[usr.Login]; ok {
		return storage.ErrUserBalanceAlreadyExists
	}

	s.UserBalanceStore.balances[usr.Login] = new(users.UserBalance)

	return nil
}

func (s *Storage) GetUser(_ context.Context, login string) (*users.User, error) {
	s.UserStore.mu.Lock()
	defer s.UserStore.mu.Unlock()

	user, ok := s.UserStore.users[login]
	if !ok {
		return nil, storage.ErrUserNotFound
	}

	return user, nil
}

func (s *Storage) GetUserBalance(_ context.Context, login string) (*users.UserBalance, error) {
	s.UserBalanceStore.mu.Lock()
	defer s.UserBalanceStore.mu.Unlock()

	balance, ok := s.UserBalanceStore.balances[login]
	if !ok {
		return new(users.UserBalance), nil
	}

	return balance, nil
}

func (s *Storage) DepositUserBalance(_ context.Context, login string, amount decimal.Decimal) error {
	s.UserBalanceStore.mu.Lock()
	defer s.UserBalanceStore.mu.Unlock()

	balance, ok := s.UserBalanceStore.balances[login]
	if !ok {
		return storage.ErrUserNotFound
	}

	balance.Current = balance.Current.Add(amount)

	return nil
}

func (s *Storage) WithdrawUserBalance(_ context.Context, withdrawal *withdrawals.Withdrawal) error {
	s.UserBalanceStore.mu.Lock()
	defer s.UserBalanceStore.mu.Unlock()

	s.UserWithdrawalStore.mu.Lock()
	defer s.UserWithdrawalStore.mu.Unlock()

	userLogin := withdrawal.UserLogin()

	balance, ok := s.UserBalanceStore.balances[userLogin]
	if !ok {
		return storage.ErrUserBalanceNotFound
	}

	if balance.Current.LessThanOrEqual(withdrawal.Amount()) {
		return storage.ErrUserBalanceNotEnough
	}

	balance.Current = balance.Current.Sub(withdrawal.Amount())
	balance.Withdrawn = balance.Withdrawn.Add(withdrawal.Amount())

	s.UserWithdrawalStore.withdrawals[userLogin] = append(s.UserWithdrawalStore.withdrawals[userLogin], withdrawal)

	return nil
}

func (s *Storage) GetWithdrawalsByUserLogin(_ context.Context, login string) ([]*withdrawals.Withdrawal, error) {
	s.UserWithdrawalStore.mu.Lock()
	defer s.UserWithdrawalStore.mu.Unlock()

	withdrawals, ok := s.UserWithdrawalStore.withdrawals[login]
	if !ok {
		return nil, storage.ErrBalanceWithdrawalsNotFound
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt().Before(withdrawals[j].ProcessedAt())
	})

	return withdrawals, nil
}

func (s *Storage) CreateOrder(_ context.Context, ord *orders.Order) error {
	s.OrderStore.mu.Lock()
	defer s.OrderStore.mu.Unlock()

	if _, ok := s.OrderStore.orders[ord.Number]; ok {
		return storage.ErrOrderAlreadyExists
	}

	s.OrderStore.orders[ord.Number] = ord

	return nil
}

func (s *Storage) GetOrder(_ context.Context, number string) (*orders.Order, error) {
	s.OrderStore.mu.Lock()
	defer s.OrderStore.mu.Unlock()

	ord, ok := s.OrderStore.orders[number]
	if !ok {
		return nil, storage.ErrOrderNotFound
	}

	return ord, nil
}

func (s *Storage) GetOrdersByUserLogin(_ context.Context, login string) ([]*orders.Order, error) {
	s.OrderStore.mu.Lock()
	defer s.OrderStore.mu.Unlock()

	var orders []*orders.Order
	for _, ord := range s.OrderStore.orders {
		if ord.UserLogin == login {
			orders = append(orders, ord)
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	return orders, nil
}
