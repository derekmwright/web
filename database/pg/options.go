package pg

import (
	"log/slog"
	"time"
)

type Option func(*config)

type config struct {
	dsn               string
	maxConns          int32
	minConns          int32
	maxConnLifetime   time.Duration
	healthCheckPeriod time.Duration
	log               *slog.Logger
}

func WithDSN(dsn string) Option {
	return func(c *config) { c.dsn = dsn }
}

func WithMaxConns(n int32) Option {
	return func(c *config) { c.maxConns = n }
}

func WithMinConns(n int32) Option {
	return func(c *config) { c.minConns = n }
}

func WithMaxConnLifetime(d time.Duration) Option {
	return func(c *config) { c.maxConnLifetime = d }
}

func WithHealthCheckPeriod(d time.Duration) Option {
	return func(c *config) { c.healthCheckPeriod = d }
}

func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.log = l }
}
