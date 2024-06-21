package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/andymarkow/gophermart/internal/server/router"
	"github.com/andymarkow/gophermart/internal/storage"
)

type Server struct {
	srv *http.Server
	log *slog.Logger
}

type config struct {
	serverAddr   string
	jwtSecretKey []byte
	logger       *slog.Logger
}

func NewServer(store storage.Storage, opts ...Option) (*Server, error) {
	cfg := &config{
		jwtSecretKey: []byte("secretkey"),
		logger:       slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	r := router.NewRouter(
		store,
		router.WithLogger(cfg.logger),
		router.WithSecret(cfg.jwtSecretKey),
	)

	srv := &http.Server{
		Addr:              cfg.serverAddr,
		Handler:           r,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return &Server{
		srv: srv,
		log: cfg.logger,
	}, nil
}

type Option func(c *config)

func WithServerAddr(addr string) Option {
	return func(c *config) {
		c.serverAddr = addr
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func WithJWTSecretKey(secretKey []byte) Option {
	return func(c *config) {
		c.jwtSecretKey = secretKey
	}
}

func (s *Server) Start() error {
	s.log.Info(fmt.Sprintf("Starting server on %s", s.srv.Addr))

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}

	return nil
}
