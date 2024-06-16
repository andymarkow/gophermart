package accrual

import (
	"context"
	"log/slog"
	"time"

	"github.com/andymarkow/gophermart/internal/accrual/accrclient"
	"github.com/andymarkow/gophermart/internal/accrual/accrprocessor"
	"github.com/andymarkow/gophermart/internal/httpclient"
	"github.com/andymarkow/gophermart/internal/storage"
)

type Accrual struct {
	log          *slog.Logger
	pollInterval time.Duration
	processor    *accrprocessor.AccrualProcessor
}

type Config struct {
	logger       *slog.Logger
	pollInterval time.Duration
	accrualURI   string
}

func NewAccrual(store storage.Storage, opts ...Option) *Accrual {
	cfg := &Config{
		logger:       slog.New(&slog.JSONHandler{}),
		pollInterval: 10 * time.Second,
		accrualURI:   "http://localhost:8080",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := httpclient.New()
	httpClient.SetBaseURL(cfg.accrualURI)

	accrClient := accrclient.New(
		accrclient.WithLogger(cfg.logger),
		accrclient.WithClient(httpClient),
	)

	accrProcessor := accrprocessor.New(
		store,
		accrClient,
		accrprocessor.WithLogger(cfg.logger),
	)

	return &Accrual{
		log:          cfg.logger.With(slog.String("module", "accrual")),
		pollInterval: cfg.pollInterval,
		processor:    accrProcessor,
	}
}

type Option func(a *Config)

func WithLogger(logger *slog.Logger) Option {
	return func(a *Config) {
		a.logger = logger
	}
}

func WithPollInterval(interval time.Duration) Option {
	return func(a *Config) {
		a.pollInterval = interval
	}
}

func WithAccrualURI(uri string) Option {
	return func(a *Config) {
		a.accrualURI = uri
	}
}

func (a *Accrual) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	a.log.Info("Start accrual daemon")

	for {
		select {
		case <-ctx.Done():
			a.log.Info("Context done, stopping accrual daemon")

			return nil

		case <-ticker.C:
			if err := a.processor.Process(ctx); err != nil {
				a.log.Error("processor.Process", slog.Any("error", err))
			}
		}
	}
}
