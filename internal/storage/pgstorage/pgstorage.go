package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/andymarkow/gophermart/internal/domain/balance"
	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/andymarkow/gophermart/internal/domain/withdrawals"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/andymarkow/gophermart/internal/storage/dbmodels"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	// Postgres driver.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var _ storage.Storage = (*Storage)(nil)

type Storage struct {
	db *sql.DB
}

type Config struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
}

type Option func(s *Config)

func WithMaxOpenConns(conns int) Option {
	return func(c *Config) {
		c.maxOpenConns = conns
	}
}

func WithMaxIdleConns(conns int) Option {
	return func(c *Config) {
		c.maxIdleConns = conns
	}
}

func WithConnMaxIdleTime(idleTime time.Duration) Option {
	return func(c *Config) {
		c.connMaxIdleTime = idleTime
	}
}

func WithConnMaxLifetime(lifetime time.Duration) Option {
	return func(c *Config) {
		c.connMaxLifetime = lifetime
	}
}

func NewStorage(connStr string, opts ...Option) (*Storage, error) {
	cfg := &Config{
		maxOpenConns:    10,
		maxIdleConns:    5,
		connMaxIdleTime: 180 * time.Second,
		connMaxLifetime: 3600 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(cfg.maxOpenConns)
	db.SetMaxIdleConns(cfg.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	db.SetConnMaxLifetime(cfg.connMaxLifetime)

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Bootstrap(ctx context.Context) error {
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		s.db,
		os.DirFS("internal/storage/pgstorage/migrations"),
	)
	if err != nil {
		return fmt.Errorf("goose.NewProvider: %w", err)
	}

	_, err = provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("provider.Up: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("db.Close: %w", err)
	}

	return nil
}

// isRetryableError checks if error is retryable.
func isRetryableError(err error) bool {
	// Connection refused error
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		// https://github.com/jackc/pgerrcode/blob/6e2875d9b438d43808cc033afe2d978db3b9c9e7/errcode.go#L393C6-L393C27
		return true
	}

	return false
}

// WithRetry retries operations in case of retryable errors.
func WithRetry(operation func() error) error {
	// Retry count
	retryCount := 3

	// Initial retry wait time
	var retryWaitTime time.Duration

	// Define the interval between retries
	retryWaitInterval := 2

	var err error

	for i := 0; i < retryCount; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if isRetryableError(err) {
			retryWaitTime = time.Duration((i*retryWaitInterval + 1)) * time.Second // 1s, 3s, 5s, etc.

			time.Sleep(retryWaitTime)
		} else {
			return fmt.Errorf("%w", err)
		}
	}

	return fmt.Errorf("retry attempts exceeded: %w", err)
}
func (s *Storage) Ping(ctx context.Context) error {
	err := WithRetry(func() error {
		if err := s.db.PingContext(ctx); err != nil {
			return fmt.Errorf("db.PingContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) CreateUser(ctx context.Context, usr *users.User) error {
	err := WithRetry(func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("db.Begin: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck

		createUsrQuery := `INSERT INTO users (login, password_hash) VALUES ($1, $2)`

		if _, err := tx.ExecContext(ctx, createUsrQuery, usr.Login(), usr.PasswordHash()); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return storage.ErrUserAlreadyExists
			}

			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		createUsrBalanceQuery := `INSERT INTO user_balance (login) VALUES ($1)`

		if _, err := tx.ExecContext(ctx, createUsrBalanceQuery, usr.Login()); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return storage.ErrUserBalanceAlreadyExists
			}

			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("tx.Commit: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (*users.User, error) {
	dbUser := new(dbmodels.User)

	err := WithRetry(func() error {
		query := `SELECT login, password_hash FROM users WHERE login = $1`

		row := s.db.QueryRowContext(ctx, query, login)

		if err := row.Scan(&dbUser.Login, &dbUser.PasswordHash); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrUserNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	user, err := users.NewUser(dbUser.Login, dbUser.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("users.NewUser: %w", err)
	}

	return user, nil
}

func (s *Storage) GetUserBalance(ctx context.Context, login string) (*balance.Balance, error) {
	dbBalance := new(dbmodels.UserBalance)

	err := WithRetry(func() error {
		query := `SELECT login, current, withdrawn FROM user_balance WHERE login = $1`

		row := s.db.QueryRowContext(ctx, query, login)
		if err := row.Scan(&dbBalance.Login, &dbBalance.Current, &dbBalance.Withdrawn); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrUserBalanceNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	balance, err := balance.NewBalance(dbBalance.Login, dbBalance.Current, dbBalance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("balance.NewBalance: %w", err)
	}

	return balance, nil
}

func (s *Storage) WithdrawUserBalance(ctx context.Context, withdrawal *withdrawals.Withdrawal) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	err = WithRetry(func() error {
		dbBalance := new(dbmodels.UserBalance)

		// Get current user balance.
		row := tx.QueryRowContext(ctx,
			`SELECT login, current, withdrawn FROM user_balance WHERE login = $1`, withdrawal.UserLogin())

		if err := row.Scan(&dbBalance.Login, &dbBalance.Current, &dbBalance.Withdrawn); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrUserBalanceNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		blnc, err := balance.NewBalance(dbBalance.Login, dbBalance.Current, dbBalance.Withdrawn)
		if err != nil {
			return fmt.Errorf("balance.NewBalance: %w", err)
		}

		if blnc.Current().LessThan(withdrawal.Amount()) {
			return storage.ErrUserBalanceNotEnough
		}

		blnc.SubCurrent(withdrawal.Amount())
		blnc.AddWithdrawn(withdrawal.Amount())

		// Update user balance.
		if _, err := tx.ExecContext(ctx,
			`UPDATE user_balance SET current = $1, withdrawn = $2 WHERE login = $3`,
			blnc.Current(), blnc.Withdrawn(), blnc.UserLogin(),
		); err != nil {
			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		// Insert withdrawal record.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_withdrawals (order_id, login, amount, processed_at) VALUES ($1, $2, $3, $4)`,
			withdrawal.OrderID(), withdrawal.UserLogin(), withdrawal.Amount(), withdrawal.ProcessedAt(),
		); err != nil {
			return fmt.Errorf("tx.ExecContext: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("tx.Commit: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetWithdrawalsByUserLogin(ctx context.Context, login string) ([]*withdrawals.Withdrawal, error) {
	dbWithdrawals := make([]*dbmodels.UserWithdrawal, 0)

	err := WithRetry(func() error {
		query := `SELECT order_id, login, amount, processed_at FROM user_withdrawals WHERE login = $1` +
			` ORDER BY processed_at DESC`

		rows, err := s.db.QueryContext(ctx, query, login)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}

		defer rows.Close()

		for rows.Next() {
			dbWithdrowal := new(dbmodels.UserWithdrawal)

			if err := rows.Scan(
				&dbWithdrowal.OrderID, &dbWithdrowal.Login, &dbWithdrowal.Amount, &dbWithdrowal.ProcessedAt,
			); err != nil {
				return fmt.Errorf("rows.Scan: %w", err)
			}

			dbWithdrawals = append(dbWithdrawals, dbWithdrowal)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bWithdrawals := make([]*withdrawals.Withdrawal, 0)

	for _, dbWithdrawal := range dbWithdrawals {
		withdrawal, err := withdrawals.NewWithdrawal(
			dbWithdrawal.Login,
			dbWithdrawal.OrderID,
			dbWithdrawal.Amount,
			dbWithdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("withdrawals.NewWithdrawal: %w", err)
		}

		bWithdrawals = append(bWithdrawals, withdrawal)
	}

	return bWithdrawals, nil
}

func (s *Storage) CreateOrder(ctx context.Context, order *orders.Order) error {
	err := WithRetry(func() error {
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO orders (id, user_login, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)`,
			order.ID(), order.UserLogin(), order.Status().String(), order.Accrual(), order.UploadedAt(),
		); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return storage.ErrOrderAlreadyExists
			}

			return fmt.Errorf("db.ExecContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetOrder(ctx context.Context, orderID string) (*orders.Order, error) {
	dbOrder := new(dbmodels.Order)

	err := WithRetry(func() error {
		query := `SELECT id, user_login, status, accrual, uploaded_at FROM orders WHERE id = $1`

		row := s.db.QueryRowContext(ctx, query, orderID)
		if err := row.Scan(
			&dbOrder.ID, &dbOrder.UserLogin, &dbOrder.Status, &dbOrder.Accrual, &dbOrder.UploadedAt,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrOrderNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	order, err := orders.NewOrder(
		dbOrder.ID,
		dbOrder.UserLogin,
		orders.OrderStatus(dbOrder.Status),
		dbOrder.Accrual,
		dbOrder.UploadedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("orders.NewOrder: %w", err)
	}

	return order, nil
}

func (s *Storage) GetOrdersByLogin(ctx context.Context, login string) ([]*orders.Order, error) {
	dbOrders := make([]*dbmodels.Order, 0)

	err := WithRetry(func() error {
		query := `SELECT id, user_login, status, accrual, uploaded_at FROM orders` +
			` WHERE user_login = $1 ORDER BY uploaded_at DESC`

		rows, err := s.db.QueryContext(ctx, query, login)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			dbOrder := new(dbmodels.Order)

			if err := rows.Scan(
				&dbOrder.ID,
				&dbOrder.UserLogin,
				&dbOrder.Status,
				&dbOrder.Accrual,
				&dbOrder.UploadedAt,
			); err != nil {
				return fmt.Errorf("rows.Scan: %w", err)
			}

			dbOrders = append(dbOrders, dbOrder)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bOrders := make([]*orders.Order, 0)

	for _, dbOrder := range dbOrders {
		order, err := orders.NewOrder(
			dbOrder.ID,
			dbOrder.UserLogin,
			orders.OrderStatus(dbOrder.Status),
			dbOrder.Accrual,
			dbOrder.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("orders.NewOrder: %w", err)
		}

		bOrders = append(bOrders, order)
	}

	return bOrders, nil
}

func (s *Storage) GetOrdersByStatus(ctx context.Context, statuses ...orders.OrderStatus) ([]*orders.Order, error) {
	dbOrders := make([]*dbmodels.Order, 0)

	err := WithRetry(func() error {
		query := `SELECT id, user_login, status, accrual, uploaded_at FROM orders`

		if len(statuses) > 0 {
			query += ` WHERE status = ANY($1)`
		}

		query += ` ORDER BY uploaded_at DESC`

		rows, err := s.db.QueryContext(ctx, query, pq.Array(statuses))
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			dbOrder := new(dbmodels.Order)

			if err := rows.Scan(
				&dbOrder.ID,
				&dbOrder.UserLogin,
				&dbOrder.Status,
				&dbOrder.Accrual,
				&dbOrder.UploadedAt,
			); err != nil {
				return fmt.Errorf("rows.Scan: %w", err)
			}

			dbOrders = append(dbOrders, dbOrder)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bOrders := make([]*orders.Order, 0)

	for _, dbOrder := range dbOrders {
		order, err := orders.NewOrder(
			dbOrder.ID,
			dbOrder.UserLogin,
			orders.OrderStatus(dbOrder.Status),
			dbOrder.Accrual,
			dbOrder.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("orders.NewOrder: %w", err)
		}

		bOrders = append(bOrders, order)
	}

	return bOrders, nil
}

func (s *Storage) ProcessOrderAccrual(ctx context.Context, order *orders.Order) error {
	err := WithRetry(func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("db.BeginTx: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck

		dbOrder := new(dbmodels.Order)

		query := `SELECT id, user_login, status, accrual, uploaded_at FROM orders WHERE id = $1`

		// Get order.
		row := tx.QueryRowContext(ctx, query, order.ID())

		if err := row.Scan(
			&dbOrder.ID, &dbOrder.UserLogin, &dbOrder.Status, &dbOrder.Accrual, &dbOrder.UploadedAt,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrOrderNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		// Set order properties.
		dbOrder.Status = order.Status().String()
		dbOrder.Accrual = order.Accrual()

		// Update order.
		if _, err := tx.ExecContext(ctx,
			`UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`,
			dbOrder.Status, dbOrder.Accrual, order.ID(),
		); err != nil {
			return fmt.Errorf("db.ExecContext: %w", err)
		}

		// Get user balance.
		dbUserBalance := new(dbmodels.UserBalance)

		query = `SELECT login, current, withdrawn FROM user_balance WHERE login = $1`

		row = tx.QueryRowContext(ctx, query, order.UserLogin())
		if err := row.Scan(
			&dbUserBalance.Login, &dbUserBalance.Current, &dbUserBalance.Withdrawn,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storage.ErrUserBalanceNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		// Set user balance properties.
		dbUserBalance.Current = dbUserBalance.Current.Add(order.Accrual())

		// Update user balance.
		if _, err := tx.ExecContext(ctx,
			`UPDATE user_balance SET current = $1 WHERE login = $2`,
			dbUserBalance.Current, order.UserLogin(),
		); err != nil {
			return fmt.Errorf("db.ExecContext: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("tx.Commit: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
