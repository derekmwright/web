# web

[![Go Reference](https://pkg.go.dev/badge/github.com/derekmwright/web.svg)](https://pkg.go.dev/github.com/derekmwright/web)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Version](https://img.shields.io/github/v/release/derekmwright/web?sort=semver)

**A small, opinionated collection of Go packages to make building production web services faster and less repetitive.**

This library bundles battle-tested patterns and helpers I use across multiple projects.  
Everything is deliberately opinionated so you can move fast without endless configuration.

## Packages

| Package              | Description                                      |
|----------------------|--------------------------------------------------|
| `server`             | HTTP server setup, middleware, graceful shutdown |
| `auth/auth0`         | Auth0 JWT validation & user context              |
| `database/pg`        | PostgreSQL connection pool, common queries & tx  |
| `worker`             | Simple background worker with graceful shutdown  |
| `nats`               | NATS client utilities & common patterns          |

## Installation

```bash
go get github.com/derekmwright/web@latest
```

Or pin a specific version:

```bash
go get github.com/derekmwright/web@v0.1.0
```

## Quick Start Example

A minimal web service skeleton using several packages together:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/derekmwright/web/server"
    "github.com/derekmwright/web/auth/auth0"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Initialize Auth0 validator (expects AUTH0_ISSUER & AUTH0_AUDIENCE env vars)
    auth, err := auth0.NewValidator()
    if err != nil {
        log.Fatalf("failed to init auth0: %v", err)
    }

    mux := http.NewServeMux()

    // Public route
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })

    // Protected route example
    mux.Handle("GET /api/me", auth.Protect(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := auth0.ClaimsFrom(r.Context())
            w.Write([]byte("Hello, " + claims.Subject))
        }),
    ))

    srv := &http.Server{
        Addr:    ":8080",
        Handler: server.WithDefaultMiddleware(mux),
    }

    go func() {
        log.Printf("starting server on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server failed: %v", err)
        }
    }()

    <-ctx.Done()
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer shutdownCancel()

    log.Println("shutting down server...")
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("graceful shutdown error: %v", err)
    }
}
```

## Contributing

Contributions are very welcome!  
Especially interested in:

- Better error handling patterns
- Testing helpers
- Real-world usage feedback

Feel free to open issues or PRs.

## License

[MIT License](LICENSE) © 2026 Derek Wright