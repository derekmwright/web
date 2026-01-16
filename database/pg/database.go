package pg

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
	log  *slog.Logger
}

func New(opts ...Option) (*Database, error) {
	cfg := &config{
		maxConns:          25,
		minConns:          5,
		maxConnLifetime:   time.Hour,
		healthCheckPeriod: 30 * time.Second,
		log:               slog.Default(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.dsn == "" {
		return nil, ErrDSNRequired
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.maxConns
	poolCfg.MinConns = cfg.minConns
	poolCfg.MaxConnLifetime = cfg.maxConnLifetime
	poolCfg.HealthCheckPeriod = cfg.healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	cfg.log.Info("connected to database")

	return &Database{Pool: pool, log: cfg.log}, nil
}

func (d *Database) Close() {
	d.log.Info("closing database connection pool")
	d.Pool.Close()
}

func (d *Database) Health() error {
	return d.Pool.Ping(context.Background())
}
