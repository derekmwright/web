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
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/derekmwright/web/auth/auth0"
	"github.com/derekmwright/web/database/pg"
	"github.com/derekmwright/web/server"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := pg.New(
		pg.WithLogger(logger),
		pg.WithDSN(os.Getenv("DATABASE_URL")),
	)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	sessions := scs.New()
	sessions.Store = postgresstore.New(stdlib.OpenDBFromPool(db.Pool))

	// Setup site-wide Auth0
	registerAuth, requireAuth, err := auth0.New(
		auth0.WithLogger(logger),
		auth0.WithSessions(sessions),
		auth0.WithPostLoginHooks(
			// Sync user data from Auth0 to your database
			users.SyncPostLogin(db.Pool, logger),
		),
	)
	if err != nil {
		logger.Error("failed to setup auth0", "err", err)
		os.Exit(1)
	}

	srv := server.New(
		server.WithLogger(logger),
		server.WithMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Datastar-Request") == "true" {
					token, _ := r.Cookie(sessions.Cookie.Name)
					ctx, _ := sessions.Load(r.Context(), token.Value)
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					sessions.LoadAndSave(next).ServeHTTP(w, r)
				}
			})
		}),
	)
	
	srv.Router.Route("/", func(r chi.Router) {
		r.Use(requireAuth)
		r.Use(users.LocalUserMiddleware(db.Pool, logger))

    // ...register your custom routes/handlers here
	})

	registerAuth(srv.Router)

	if err = srv.Start(); err != nil {
		logger.Error("failed to start server", "err", err)
		os.Exit(1)
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