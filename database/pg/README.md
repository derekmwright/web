# go-pg

A reusable, production-ready PostgreSQL module for Go applications using **pgxpool** and **Goose** migrations.

This module provides:
- Connection pooling with sensible defaults
- Health checks
- Graceful shutdown
- Embedded database migrations (via Goose)
- Functional options for configuration
- Custom logger support

## Features

- **pgxpool** for high-performance connection pooling
- Automatic health checks (`Ping`)
- Embedded migrations with **Goose** (no external tools needed)
- Configurable pool settings (max/min conns, lifetimes)
- Custom `slog.Logger` injection
- Clean, consistent API with functional options

## Installation

```bash
go get github.com/derekmwright/web/pg
```

## Quick Start

```go
package main

import (
    "log/slog"
    "os"

    "github.com/derekmwright/web/pg"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        logger.Fatal("DATABASE_URL environment variable required")
    }

    db, err := pg.New(
        pg.WithDSN(dsn),
        pg.WithMaxConns(20),
        pg.WithLogger(logger),
    )
    if err != nil {
        logger.Fatal("failed to connect to database", "error", err)
    }
    defer db.Close()

    // Run embedded migrations on startup
    if err := pg.Migrate(db); err != nil {
        logger.Fatal("migration failed", "error", err)
    }

    // Use db.Pool in your handlers/repositories
    // e.g., row := db.Pool.QueryRow(context.Background(), "SELECT ...")
}
```

## Options

| Option                        | Description                              | Default               |
|-------------------------------|------------------------------------------|-----------------------|
| `WithDSN(dsn)`                | PostgreSQL connection string (required)  | -                     |
| `WithMaxConns(n)`             | Maximum number of connections            | 25                    |
| `WithMinConns(n)`             | Minimum number of connections            | 5                     |
| `WithMaxConnLifetime(d)`      | Maximum lifetime of a connection         | 1 hour                |
| `WithHealthCheckPeriod(d)`    | How often to check connection health     | 30 seconds            |
| `WithLogger(l)`               | Custom slog logger                       | `slog.Default()`      |

## Migrations

Migrations are embedded using `//go:embed` and run automatically via Goose.

Create migration files in `migrations/` with timestamp prefixes:

```
migrations/
├── 202512190001_create_users.up.sql
├── 202512190001_create_users.down.sql
└── 202512190002_add_index.up.sql
```

Example:

```sql
-- 202512190001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

Goose will apply only new migrations on startup.

## Health Checks

Use with your server module:

```go
srv.Router.Get("/healthz/db", func(w http.ResponseWriter, r *http.Request) {
    if err := db.Health(); err != nil {
        http.Error(w, "database unhealthy", http.StatusServiceUnavailable)
        return
    }
    w.Write([]byte("database ok"))
})
```