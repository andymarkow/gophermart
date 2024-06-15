package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andymarkow/gophermart/internal/config"
	"github.com/andymarkow/gophermart/internal/logger"
	"github.com/andymarkow/gophermart/internal/server/router"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/andymarkow/gophermart/internal/storage/inmemory"
)

type Server struct {
	srv *http.Server
	log *slog.Logger
}

func NewServer() (*Server, error) {
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

	memstore := inmemory.NewStorage()

	store := storage.NewStorage(memstore)

	r := router.NewRouter(store,
		router.WithLogger(logg),
		router.WithSecret([]byte(cfg.JWTSecretKey)),
	)

	srv := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           r,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return &Server{
		srv: srv,
		log: logg,
	}, nil
}

func (s *Server) Close() {}

func (s *Server) Start() error {
	defer s.Close()

	errChan := make(chan error, 1)

	go func() {
		s.log.Info(fmt.Sprintf("Starting server on %s", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server.ListenAndServe: %w", err)
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
			s.log.Info("Gracefully shutting down server...")

			return nil
		}
	}
}
