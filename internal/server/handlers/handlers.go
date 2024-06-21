package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/andymarkow/gophermart/internal/auth"
	"github.com/andymarkow/gophermart/internal/domain/orders"
	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/andymarkow/gophermart/internal/domain/withdrawals"
	"github.com/andymarkow/gophermart/internal/errmsg"
	"github.com/andymarkow/gophermart/internal/server/models"
	"github.com/andymarkow/gophermart/internal/storage"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	storage storage.Storage
	log     *slog.Logger
	auth    *auth.JWTAuth
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(store storage.Storage, opts ...Option) *Handlers {
	handlers := &Handlers{
		storage: store,
		log:     slog.New(&slog.JSONHandler{}),
		auth:    auth.NewJWTAuth([]byte("")),
	}

	// Apply options
	for _, opt := range opts {
		opt(handlers)
	}

	return handlers
}

// Option is a functional option for Handlers.
type Option func(h *Handlers)

// WithLogger is a option for Handlers that sets logger.
func WithLogger(logger *slog.Logger) Option {
	return func(h *Handlers) {
		h.log = logger
	}
}

func WithAuth(auth *auth.JWTAuth) Option {
	return func(h *Handlers) {
		h.auth = auth
	}
}

type JSONResponse struct {
	Message any `json:"message,omitempty"`
	Error   any `json:"error,omitempty"`
}

func handleJSONResponse(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handleError(w http.ResponseWriter, err errmsg.HTTPError) {
	resp := &JSONResponse{
		Error: err.Error(),
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(err.Code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.storage.Ping(r.Context()); err != nil {
		h.log.Error("storage.Ping", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	handleJSONResponse(w, http.StatusOK, &JSONResponse{Message: "ok"})
}

func (h *Handlers) UserRegister(w http.ResponseWriter, r *http.Request) {
	var userPayload models.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
			handleError(w, errmsg.ErrRequestPayloadEmpty)

			return
		}

		h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusBadRequest, err))

		return
	}

	defer r.Body.Close()

	user, err := users.CreateUser(userPayload.Login, userPayload.Password)
	if err != nil {
		h.log.Error("users.NewUser()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	if err := h.storage.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			h.log.Error("storage.CreateUser()", slog.Any("error", err))
			handleError(w, errmsg.ErrUserAlreadyExists)

			return
		}

		h.log.Error("storage.CreateUser()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	token, err := h.auth.CreateJWTString(user.Login())
	if err != nil {
		h.log.Error("jwtauth.CreateJWTString()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) //nolint:errcheck
}

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	var userPayload models.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
			handleError(w, errmsg.ErrRequestPayloadEmpty)

			return
		}

		h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
		handleError(w, errmsg.ErrRequestPayloadInvalid)

		return
	}

	defer r.Body.Close()

	user, err := h.storage.GetUser(r.Context(), userPayload.Login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.log.Error("storage.GetUser()", slog.Any("error", err))
			handleError(w, errmsg.ErrUserNotFound)

			return
		}

		h.log.Error("storage.GetUser()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(userPayload.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.log.Error("bcrypt.CompareHashAndPassword()", slog.Any("error", err))
			handleError(w, errmsg.ErrUserCredentialsInvalid)

			return
		}

		h.log.Error("bcrypt.CompareHashAndPassword()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	token, err := h.auth.CreateJWTString(user.Login())
	if err != nil {
		h.log.Error("jwtauth.CreateJWTString()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) //nolint:errcheck
}

func (h *Handlers) CreateUserOrder(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.log.Error("jwtauth.FromContext()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Set user login from JWT sub claim field
	userLogin := token.Subject()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("io.ReadAll()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}
	defer r.Body.Close()

	// Set order number from request body
	orderNumber := string(body)

	orderReq, err := orders.CreateOrder(orderNumber, userLogin)
	if err != nil {
		if errors.Is(err, orders.ErrOrderIDFormatInvalid) {
			h.log.Error("orders.CreateOrder()", slog.Any("error", err))
			handleError(w, errmsg.ErrOrderNumberFormatInvalid)

			return
		}

		h.log.Error("orders.CreateOrder()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusBadRequest, err))

		return
	}

	// Check if order already exists
	orderCheck, err := h.storage.GetOrder(r.Context(), orderNumber)
	if err != nil {
		// If order does not exist, create it
		if errors.Is(err, storage.ErrOrderNotFound) {
			if err := h.storage.CreateOrder(r.Context(), orderReq); err != nil {
				h.log.Error("storage.CreateOrder()", slog.Any("error", err))
				handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

				return
			}

			handleJSONResponse(w, http.StatusAccepted, &JSONResponse{Message: "ok"})

			return
		}

		h.log.Error("storage.GetOrder()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Order has already created by another user
	if orderReq.UserLogin() != orderCheck.UserLogin() {
		handleError(w, errmsg.ErrOrderCreatedByAnotherUser)

		return
	}

	// Order has already created by the same user
	handleJSONResponse(w, http.StatusOK, &JSONResponse{Message: "ok"})
}

func (h *Handlers) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.log.Error("jwtauth.FromContext()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Set user login from JWT sub claim field
	userLogin := token.Subject()

	userOrders, err := h.storage.GetOrdersByLogin(r.Context(), userLogin)
	if err != nil {
		h.log.Error("storage.GetOrdersByLogin()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	if len(userOrders) == 0 {
		handleJSONResponse(w, http.StatusNoContent, &JSONResponse{Message: "ok"})

		return
	}

	orderResp := make([]models.OrderResponse, 0, len(userOrders))
	for _, ord := range userOrders {
		orderResp = append(orderResp, models.OrderResponse{
			Number:     ord.ID(),
			Status:     ord.Status(),
			Accrual:    ord.Accrual().InexactFloat64(),
			UploadedAt: ord.UploadedAt().Format(time.RFC3339),
		})
	}

	handleJSONResponse(w, http.StatusOK, orderResp)
}

func (h *Handlers) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.log.Error("jwtauth.FromContext()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Set user login from JWT sub claim field
	userLogin := token.Subject()

	userBalance, err := h.storage.GetUserBalance(r.Context(), userLogin)
	if err != nil {
		h.log.Error("storage.GetUserBalance()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	resp := models.UserBalanceResponse{
		Current:   userBalance.Current().InexactFloat64(),
		Withdrawn: userBalance.Withdrawn().InexactFloat64(),
	}

	handleJSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) WithdrawUserBalance(w http.ResponseWriter, r *http.Request) {
	var withdrawalRequest models.BalanceWithdrawalRequest

	if err := json.NewDecoder(r.Body).Decode(&withdrawalRequest); err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
			handleError(w, errmsg.ErrRequestPayloadEmpty)

			return
		}

		h.log.Error("json.NewDecoder().Decode()", slog.Any("error", err))
		handleError(w, errmsg.ErrRequestPayloadInvalid)

		return
	}

	defer r.Body.Close()

	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.log.Error("jwtauth.FromContext()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Set user login from JWT sub claim field
	userLogin := token.Subject()

	withdrawal, err := withdrawals.CreateWithdrawal(
		userLogin, withdrawalRequest.OrderNumber, withdrawalRequest.Amount,
	)
	if err != nil {
		if errors.Is(err, orders.ErrOrderIDFormatInvalid) {
			h.log.Error("withdrawals.CreateWithdrawal()", slog.Any("error", err))
			handleError(w, errmsg.ErrOrderNumberFormatInvalid)

			return
		}

		h.log.Error("withdrawals.CreateWithdrawal()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	if err := h.storage.WithdrawUserBalance(r.Context(), withdrawal); err != nil {
		if errors.Is(err, storage.ErrUserBalanceNotEnough) {
			h.log.Error("storage.WithdrawUserBalance()", slog.Any("error", err))
			handleError(w, errmsg.ErrUserBalanceNotEnough)

			return
		}

		h.log.Error("storage.WithdrawUserBalance()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	handleJSONResponse(w, http.StatusOK, &JSONResponse{Message: "ok"})
}

func (h *Handlers) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		h.log.Error("jwtauth.FromContext()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// Set user login from JWT sub claim field
	userLogin := token.Subject()

	withdrawals, err := h.storage.GetWithdrawalsByUserLogin(r.Context(), userLogin)
	if err != nil {
		h.log.Error("storage.GetWithdrawalsByUserLogin()", slog.Any("error", err))
		handleError(w, errmsg.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	if len(withdrawals) == 0 {
		handleJSONResponse(w, http.StatusNoContent, []models.BalanceWithdrawalResponse{})

		return
	}

	withdrawalsResp := make([]models.BalanceWithdrawalResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		withdrawalsResp[i] = models.BalanceWithdrawalResponse{
			OrderNumber: withdrawal.OrderID(),
			Amount:      withdrawal.Amount().InexactFloat64(),
			ProcessedAt: withdrawal.ProcessedAt().Format(time.RFC3339),
		}
	}

	handleJSONResponse(w, http.StatusOK, withdrawalsResp)
}
