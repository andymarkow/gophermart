package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andymarkow/gophermart/internal/accrual"
	"github.com/andymarkow/gophermart/internal/config"
	"github.com/andymarkow/gophermart/internal/logger"
	"github.com/andymarkow/gophermart/internal/server"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/andymarkow/gophermart/internal/storage/inmemory"
	"github.com/andymarkow/gophermart/internal/storage/pgstorage"
)

type Application struct {
	log     *slog.Logger
	storage storage.Storage
	server  *server.Server
	accrual *accrual.Accrual
}

func New() (*Application, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("config.NewConfig: %w", err)
	}

	logLevel, err := logger.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("logger.ParseLogLevel: %w", err)
	}

	logg := logger.NewLogger(
		logger.WithLevel(logLevel),
		logger.WithFormat(logger.LogFormatJSON),
		logger.WithAddSource(false),
	)

	var repo storage.Storage

	repo = inmemory.NewStorage()

	if cfg.DatabaseURI != "" {
		repo, err = pgstorage.NewStorage(cfg.DatabaseURI)
		if err != nil {
			return nil, fmt.Errorf("pgstorage.NewStorage: %w", err)
		}

		if pgrepo, ok := repo.(*pgstorage.Storage); ok {
			logg.Info("Running migrations for Postgres database...")

			if err := pgrepo.Bootstrap(context.Background()); err != nil {
				return nil, fmt.Errorf("pgstorage.Bootstrap: %w", err)
			}
		}
	}

	store := storage.NewStorage(repo)

	srv, err := server.NewServer(
		store,
		server.WithServerAddr(cfg.ServerAddr),
		server.WithJWTSecretKey([]byte(cfg.JWTSecretKey)),
		server.WithLogger(logg),
	)
	if err != nil {
		return nil, fmt.Errorf("server.NewServer: %w", err)
	}

	logg.Info("Accrual system address set to: " + cfg.AccrualURI)

	accr := accrual.NewAccrual(
		store,
		accrual.WithLogger(logg),
		accrual.WithAccrualURI(cfg.AccrualURI),
		accrual.WithPollInterval(cfg.AccrualPollInterval),
	)

	return &Application{
		log:     logg,
		storage: store,
		server:  srv,
		accrual: accr,
	}, nil
}

func (a *Application) Run() error {
	defer func() {
		if err := a.storage.Close(); err != nil {
			a.log.Error("storage.Close()", slog.Any("error", err))
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		if err := a.server.Start(); err != nil {
			errChan <- fmt.Errorf("server.Start: %w", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()

		time.Sleep(time.Second * 1)
	}()

	go func() {
		if err := a.accrual.Run(ctx); err != nil {
			errChan <- fmt.Errorf("accrual.Run: %w", err)
		}
	}()

	// Graceful shutdown handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			return err

		case <-quit:
			a.log.Info("Gracefully shutting down application...")

			return nil
		}
	}
}
