package accrclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/andymarkow/gophermart/internal/httpclient"
	"github.com/go-resty/resty/v2"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrSomethingWentWrong = errors.New("something went wrong")
)

type AccrualClient struct {
	log    *slog.Logger
	client *resty.Client
}

func New(opts ...Option) *AccrualClient {
	accrClient := &AccrualClient{
		log:    slog.New(&slog.JSONHandler{}),
		client: httpclient.New(),
	}

	for _, opt := range opts {
		opt(accrClient)
	}

	return accrClient
}

type Option func(a *AccrualClient)

func WithLogger(logger *slog.Logger) Option {
	return func(a *AccrualClient) {
		a.log = logger
	}
}

func WithClient(client *resty.Client) Option {
	return func(a *AccrualClient) {
		a.client = client
	}
}

func (a *AccrualClient) GetOrder(ctx context.Context, orderNumber string) (*Order, error) {
	orderData := new(OrderModel)

	resp, err := a.client.R().
		SetContext(ctx).
		SetResult(orderData).
		SetPathParams(map[string]string{
			"orderNumber": orderNumber,
		}).
		Get("/api/orders/{orderNumber}")
	if err != nil {
		return nil, fmt.Errorf("client.R: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil, ErrOrderNotFound
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	case http.StatusInternalServerError:
		return nil, ErrSomethingWentWrong
	}

	order, err := newOrder(orderData.Number, orderData.Status, orderData.Accrual)
	if err != nil {
		return nil, fmt.Errorf("newOrder: %w", err)
	}

	return order, nil
}
