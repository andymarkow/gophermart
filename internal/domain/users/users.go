package users

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserLoginEmpty  = errors.New("user login is empty")
	ErrUserPasswdEmpty = errors.New("user password is empty")
)

type User struct {
	login        string
	passwordHash string
}

func NewUser(login, passwordHash string) (*User, error) {
	if err := ValidateLogin(login); err != nil {
		return nil, err
	}

	return &User{
		login:        login,
		passwordHash: passwordHash,
	}, nil
}

func CreateUser(login, password string) (*User, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	passwordHash, err := getPasswordHash(password)
	if err != nil {
		return nil, fmt.Errorf("getPasswordHash: %w", err)
	}

	return NewUser(login, passwordHash)
}

func (u *User) Login() string {
	return u.login
}

func (u *User) PasswordHash() string {
	return u.passwordHash
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
