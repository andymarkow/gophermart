package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// LogFormat represents the format of the log output.
type LogFormat string

const (
	// LogFormatJSON represents the JSON log format.
	LogFormatJSON LogFormat = "json"

	// LogFormatText represents the text log format.
	LogFormatText LogFormat = "text"
)

type logger struct {
	level     *slog.LevelVar
	format    LogFormat
	addSource bool
}

// NewLogger creates a new logger.
func NewLogger(opts ...Option) *slog.Logger {
	logg := &logger{
		level:  &slog.LevelVar{}, // Default log level is INFO.
		format: LogFormatJSON,
	}

	// Apply options
	for _, opt := range opts {
		opt(logg)
	}

	slogOpts := &slog.HandlerOptions{
		AddSource: logg.addSource,
		Level:     logg.level,
	}

	var logHandler slog.Handler = slog.NewJSONHandler(os.Stdout, slogOpts)
	if logg.format == LogFormatText {
		logHandler = slog.NewTextHandler(os.Stdout, slogOpts)
	}

	return slog.New(logHandler)
}

type Option func(l *logger)

func WithLevel(level slog.Level) Option {
	return func(l *logger) {
		l.level.Set(level)
	}
}

func WithFormat(format LogFormat) Option {
	return func(l *logger) {
		l.format = format
	}
}

func WithAddSource(addSource bool) Option {
	return func(l *logger) {
		l.addSource = addSource
	}
}

func ParseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", level)
	}
}
