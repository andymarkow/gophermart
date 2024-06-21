package accrprocessor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/andymarkow/gophermart/internal/accrual/accrclient"
	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/storage"
)

type AccrualProcessor struct {
	log        *slog.Logger
	storage    storage.Storage
	accrclient *accrclient.AccrualClient
}

type Config struct {
	logger *slog.Logger
}

type Option func(a *Config)

func WithLogger(logger *slog.Logger) Option {
	return func(a *Config) {
		a.logger = logger
	}
}

func New(store storage.Storage, accrclient *accrclient.AccrualClient, opts ...Option) *AccrualProcessor {
	cfg := &Config{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &AccrualProcessor{
		log:        cfg.logger.With(slog.String("module", "accrual_processor")),
		storage:    store,
		accrclient: accrclient,
	}
}

func (a *AccrualProcessor) Process(ctx context.Context) error {
	a.log.Info("Start orders processing")

	orderStatuses := []orders.OrderStatus{
		orders.OrderStatusNew,
		orders.OrderStatusProcessing,
	}

	// Get unprocessed orders
	ords, err := a.storage.GetOrdersByStatus(ctx, orderStatuses...)
	if err != nil {
		return fmt.Errorf("storage.GetOrdersByStatus: %w", err)
	}

	if len(ords) == 0 {
		a.log.Info("No orders to process, stopping processing")

		return nil
	}

	ordCh := orderGenerator(ctx, ords)

	a.orderProcessor(ctx, ordCh)

	return nil
}

func orderGenerator(ctx context.Context, ords []*orders.Order) chan orders.Order {
	ordersCh := make(chan orders.Order)

	go func() {
		defer close(ordersCh)

		for _, ord := range ords {
			select {
			case <-ctx.Done():
				return
			case ordersCh <- *ord:
			}
		}
	}()

	return ordersCh
}

func (a *AccrualProcessor) orderProcessor(ctx context.Context, ordCh chan orders.Order) {
	poolSize := 1

	wg := &sync.WaitGroup{}

	// Spawn workers
	for w := 1; w <= poolSize; w++ {
		wg.Add(1)
		go a.orderProcessorWorker(ctx, wg, ordCh)
	}

	// Wait for workers
	wg.Wait()
}

func (a *AccrualProcessor) orderProcessorWorker(ctx context.Context, wg *sync.WaitGroup, ordersCh chan orders.Order) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			a.log.Info("Context done, stopping processing")

			return

		case order, ok := <-ordersCh:
			if !ok {
				a.log.Info("Orders channel closed, stopping processing")

				return
			}

			a.log.Info("Processing order", slog.String("order_number", order.ID()))

			// Get order data from accrual system
			ord, err := a.accrclient.GetOrder(ctx, order.ID())
			if err != nil {
				a.log.Error("accrclient.GetOrder()", slog.Any("error", err))

				continue
			}

			// If order is not processed yet, skip it
			if ord.Status() == accrclient.OrderStatusRegistered || ord.Status() == accrclient.OrderStatusProcessing {
				a.log.Info("Order is not processed yet in accrual system", slog.String("order_number", order.ID()))

				continue
			}

			order.SetStatus(orders.OrderStatus(ord.Status()))
			order.SetAccrual(ord.Accrual())

			if err := a.storage.ProcessOrderAccrual(ctx, &order); err != nil {
				a.log.Error("storage.ProcessOrderAccrual()", slog.Any("error", err))

				continue
			}

			a.log.Info("Order processed",
				slog.String("order_number", order.ID()),
				slog.String("order_status", string(order.Status())),
				slog.String("order_user", order.UserLogin()),
				slog.String("order_accrual", order.Accrual().String()),
			)
		}
	}
}
