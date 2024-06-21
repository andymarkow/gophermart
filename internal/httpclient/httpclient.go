package httpclient

import (
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

type Config struct {
	baseURL            string
	retryCount         int
	retryWaitTime      time.Duration
	retryMaxWaitTime   time.Duration
	retryAfterInterval int
}

type Option func(c *Config)

func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.baseURL = baseURL
	}
}

func WithRetryCount(count int) Option {
	return func(c *Config) {
		c.retryCount = count
	}
}

func WithRetryWaitTime(waitTime time.Duration) Option {
	return func(c *Config) {
		c.retryWaitTime = waitTime
	}
}

func WithRetryMaxWaitTime(maxWaitTime time.Duration) Option {
	return func(c *Config) {
		c.retryMaxWaitTime = maxWaitTime
	}
}

func WithRetryAfterInterval(retryAfterInterval int) Option {
	return func(c *Config) {
		c.retryAfterInterval = retryAfterInterval
	}
}

func New(opts ...Option) *resty.Client {
	cfg := &Config{
		baseURL:            "",
		retryCount:         3,
		retryWaitTime:      1 * time.Second,
		retryMaxWaitTime:   10 * time.Second,
		retryAfterInterval: 2,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	client := resty.New().
		// SetLogger(cfg.logger).
		SetBaseURL(cfg.baseURL).
		SetRetryCount(cfg.retryCount).       // Number of retry attempts
		SetRetryWaitTime(cfg.retryWaitTime). // Initial wait time between retries
		SetRetryMaxWaitTime(cfg.retryMaxWaitTime).
		SetRetryAfter(retryAfterWithInterval(cfg.retryAfterInterval)).
		AddRetryCondition(func(_ *resty.Response, err error) bool {
			// Retry for retryable errors
			return isRetryableError(err)
		})

	return client
}

// retryAfterWithInterval returns duration intervals between retries.
func retryAfterWithInterval(retryWaitInterval int) resty.RetryAfterFunc {
	return func(_ *resty.Client, resp *resty.Response) (time.Duration, error) {
		return time.Duration((resp.Request.Attempt*retryWaitInterval - 1)) * time.Second, nil
	}
}

// isRetryableError checks if the error is a retryable error.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		// Connection refused error
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			// Connection timeout error
			return true
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// DNS error
		return true
	}

	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		// Address error
		return true
	}

	// Operational error
	var opErr *net.OpError

	return errors.As(err, &opErr)
}
