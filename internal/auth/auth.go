package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTAuth struct {
	secret   []byte
	issuer   string
	tokenTTL time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
}

func NewJWTAuth(secret []byte, opts ...Option) *JWTAuth {
	a := &JWTAuth{
		secret:   secret,
		tokenTTL: 24 * time.Hour,
		issuer:   "gophermart",
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

type Option func(a *JWTAuth)

func WithIssuer(issuer string) Option {
	return func(a *JWTAuth) {
		a.issuer = issuer
	}
}

func WithTokenTTL(ttl time.Duration) Option {
	return func(a *JWTAuth) {
		a.tokenTTL = ttl
	}
}

func (a *JWTAuth) CreateJWTString(sub string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
		},
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return tokenString, nil
}
