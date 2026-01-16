# go-server

A minimal, production-ready, reusable HTTP server foundation for Go web services using **Chi** router and **slog** structured logging.

This package provides everything you need to spin up a secure, observable, and gracefully shutdown-capable web server with almost zero boilerplate.

## Features

- Automatic TLS support (via env vars)
- Graceful shutdown on `SIGINT` and `SIGTERM`
- Structured JSON logging with `log/slog`
- Request ID generation and propagation
- Panic recovery with stack traces
- Structured access logging (method, path, duration, status, bytes, request_id)
- Built-in `/healthz` and `/readyz` endpoints
- Configurable timeouts (read, write, idle, shutdown)
- Functional options for clean configuration
- Full request-scoped contextual logging via `slog.Logger.With()`

## Installation

```bash
go get github.com/derekmwright/web/server
```

## Quick Start

```go
package main

import (
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/derekmwright/web/server"
)

func main() {
    r := chi.NewRouter()

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from your new service!"))
    })

    // Optional: custom logger, timeouts, etc.
    srv := server.New(
        server.WithRouter(r),
        // server.WithLogger(customLogger),
        // server.WithShutdownTimeout(20 * time.Second),
    )

    if err := srv.Start(); err != nil {
        srv.Log.Error("server stopped with error", "error", err)
        os.Exit(1)
    }
}
```

That's it — your service is now running with full production features.

## Configuration (Environment Variables)

| Variable                     | Description                          | Default      |
|------------------------------|--------------------------------------|--------------|
| `APP_SERVER_ADDR`            | Listen address (host:port)           | `:8080`      |
| `APP_SERVER_TLS_KEY_FILE`    | Path to TLS private key              | (none)       |
| `APP_SERVER_TLS_CERT_FILE`   | Path to TLS certificate              | (none)       |

TLS is automatically enabled if both key and cert files exist and are readable.

## Logging

- Uses `log/slog` with JSON output by default
- All logs include timestamps and levels
- Access logs include duration, status, bytes, and request_id
- Panics are recovered and logged with full stack trace
- Request-scoped logging: use `server.LoggerFromContext(r.Context())` in handlers for logs that automatically include `request_id`, `method`, `path`, etc.

### Example Handler with Request-Scoped Logging

```go
func protectedHandler(w http.ResponseWriter, r *http.Request) {
    log := server.LoggerFromContext(r.Context())

    log.Info("handling protected request", "action", "load_dashboard")

    // All logs here automatically include request_id, method, path, etc.
    log.Debug("fetching user data", "user_id", 123)

    w.Write([]byte("Protected content"))
}
```

## Middleware Stack (Applied by Default)

1. Panic recovery (with structured error logging)
2. Request ID generation (`X-Request-ID` header)
3. Structured request logging

You can override the router completely with `WithRouter()` — middleware will still apply unless you replace the router after `New()`.

## Options

```go
server.WithLogger(logger *slog.Logger)
server.WithRouter(router chi.Router)
server.WithReadTimeout(duration time.Duration)
server.WithWriteTimeout(duration time.Duration)
server.WithIdleTimeout(duration time.Duration)
server.WithShutdownTimeout(duration time.Duration)
```

## Health Endpoints

- `GET /healthz` → returns "ok" (200)
- `GET /readyz` → returns "ready" (200)