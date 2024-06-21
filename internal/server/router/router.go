package router

import (
	"log/slog"

	"github.com/andymarkow/gophermart/internal/auth"
	"github.com/andymarkow/gophermart/internal/server/handlers"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Options struct {
	log    *slog.Logger
	secret []byte
}

func NewRouter(store storage.Storage, opts ...Option) chi.Router {
	r := chi.NewRouter()

	rOpts := Options{
		log:    slog.New(&slog.JSONHandler{}),
		secret: []byte(""),
	}

	for _, opt := range opts {
		opt(&rOpts)
	}

	tokenAuth := jwtauth.New("HS256", rOpts.secret, nil)

	r.Use(
		middleware.Recoverer,
		middleware.StripSlashes,
		middleware.Logger,
	)

	h := handlers.NewHandlers(store,
		handlers.WithLogger(rOpts.log),
		handlers.WithAuth(auth.NewJWTAuth(rOpts.secret)),
	)

	r.Get("/ping", h.Ping)

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.UserRegister)
		r.Post("/api/user/login", h.UserLogin)
	})

	r.Group(func(r chi.Router) {
		r.Use(
			jwtauth.Verifier(tokenAuth),
			jwtauth.Authenticator(tokenAuth),
		)

		r.Get("/api/user/orders", h.GetUserOrders)
		r.Post("/api/user/orders", h.CreateUserOrder)
		r.Get("/api/user/balance", h.GetUserBalance)
		r.Get("/api/user/withdrawals", h.GetUserWithdrawals)
		r.Post("/api/user/balance/withdraw", h.WithdrawUserBalance)
	})

	return r
}

type Option func(r *Options)

func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.log = logger
	}
}

func WithSecret(secret []byte) Option {
	return func(o *Options) {
		o.secret = secret
	}
}
