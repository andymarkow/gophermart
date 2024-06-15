package users

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserLoginEmpty  = errors.New("user login is empty")
	ErrUserPasswdEmpty = errors.New("user password is empty")
)

type User struct {
	Login        string
	PasswordHash string
}

type UserBalance struct {
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}

func NewUser(login, password string) (*User, error) {
	if err := ValidateLogin(login); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	passwordHash, err := getPasswordHash(password)
	if err != nil {
		return nil, fmt.Errorf("getPasswordHash: %w", err)
	}

	return &User{
		Login:        login,
		PasswordHash: passwordHash,
	}, nil
}

func getPasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	return string(hash), nil
}

func ValidateLogin(login string) error {
	if login == "" {
		return ErrUserLoginEmpty
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrUserPasswdEmpty
	}

	return nil
}
