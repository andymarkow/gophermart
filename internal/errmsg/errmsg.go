package errmsg

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	Code    int
	Message error
}

func NewHTTPError(code int, message error) HTTPError {
	return HTTPError{Code: code, Message: message}
}

func (e *HTTPError) Error() string {
	return e.Message.Error()
}

var (
	ErrRequestPayloadEmpty = NewHTTPError(
		http.StatusBadRequest,
		errors.New("request payload is empty"),
	)

	ErrRequestPayloadInvalid = NewHTTPError(
		http.StatusBadRequest,
		errors.New("request payload is invalid"),
	)
)

var (
	ErrUserAlreadyExists = NewHTTPError(
		http.StatusConflict,
		errors.New("user already exists"),
	)

	ErrUserNotFound = NewHTTPError(
		http.StatusNotFound,
		errors.New("user not found"),
	)

	ErrUserCredentialsInvalid = NewHTTPError(
		http.StatusUnauthorized,
		errors.New("user credentials invalid"),
	)

	ErrUserBalanceNotEnough = NewHTTPError(
		http.StatusPaymentRequired,
		errors.New("user balance not enough funds"),
	)

	ErrBalanceWithdrawalsNotFound = NewHTTPError(
		http.StatusNoContent,
		errors.New("balance withdrawals not found"),
	)
)

var (
	ErrOrderNumberFormatInvalid = NewHTTPError(
		http.StatusUnprocessableEntity,
		errors.New("order number format is invalid"),
	)

	ErrOrderCreatedByAnotherUser = NewHTTPError(
		http.StatusConflict,
		errors.New("order created by another user"),
	)
)
