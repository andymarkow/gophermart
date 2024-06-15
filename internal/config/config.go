package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddr   string `env:"RUN_ADDRESS"`
	LogLevel     string `env:"LOG_LEVEL"`
	DatabaseURI  string `env:"DATABASE_URI"`
	AccrualURI   string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecretKey string `env:"JWT_SECRET_KEY"`
}

func NewConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.ServerAddr, "a", "0.0.0.0:8080", "server listening address [env:RUN_ADDRESS]")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log output level [env:LOG_LEVEL]")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection string [env:DATABASE_URI]")
	flag.StringVar(&cfg.AccrualURI, "r", "", "accrual system URI [env:ACCRUAL_SYSTEM_ADDRESS]")
	flag.StringVar(&cfg.JWTSecretKey, "s", "secretkey", "JWT secret to sign tokens [env:JWT_SECRET_KEY]")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("env.Parse: %w", err)
	}

	return cfg, nil
}
