# Auth0

A clean, secure, and reusable Auth0 authentication module for Go web applications using **Chi** router.

This module provides everything you need to add Auth0-based login/logout to your Chi-based application:

- `/login` — Redirects user to Auth0 Universal Login
- `/callback` — Handles Auth0 redirect, verifies ID token, stores user profile & access token in session
- `/logout` — Clears session and redirects to Auth0 logout (full single sign-out)
- Authentication middleware — Protects routes, redirects unauthenticated users to login
- `CurrentUser(r *http.Request)` helper — Retrieve authenticated user claims in handlers

## Features

- Secure OAuth2/OIDC flow with state validation and CSRF protection
- Clean functional options pattern for dependency injection
- Easy-to-use middleware for protected routes

## Installation

```bash
go get github.com/derekmwright/web/auth/auth0
```

## Usage

```go
package main

import (
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/derekmwright/web/auth/auth0"
    // your session manager, e.g. gorilla/sessions, scollett/chi-sessions, etc.
)

func main() {
    r := chi.NewRouter()

    // Your session manager (must implement auth0.SessionManager interface)
    sessionManager := NewYourSessionManager() // e.g. cookie store

    // Your logger (zap, zerolog, etc. — must implement auth0.Logger)
    logger := NewYourLogger()

    // Initialize the Auth0 module
    registerRoutes, requireAuth, err := auth0.New(
        auth0.WithLogger(logger),
        auth0.WithSessions(sessionManager),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Mount the auth routes (usually under root or /auth)
    registerRoutes(r)

    // Public routes
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Home page — public"))
    })

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(requireAuth) // ← enforces authentication

        r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
            user := auth0.CurrentUser(r)
            // user is map[string]any with Auth0 claims (sub, name, email, etc.)
            w.Write([]byte("Welcome to the dashboard!"))
        })

        r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
            profile := auth0.CurrentUser(r).(map[string]any)
            // Render profile...
        })
    })

    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", r)
}
```

## Routes Added

When you call `registerRoutes(r)`, the following routes are registered:

| Route       | Method | Purpose |
|-------------|--------|---------|
| `/login`    | GET    | Initiates login: generates state, stores in session, redirects to Auth0 |
| `/callback` | GET    | Auth0 redirect URI: validates state, exchanges code, verifies ID token, stores user & access token in session, redirects to `/` |
| `/logout`   | GET    | Clears session and redirects to Auth0 `/v2/logout` with proper `returnTo` and `client_id` (full SSO logout) |

You can mount these under a subrouter if preferred:

```go
authRouter := chi.NewRouter()
registerRoutes(authRouter)
r.Mount("/auth", authRouter) // → /auth/login, /auth/callback, etc.
```

## Required Environment Variables

The module reads Auth0 configuration from environment variables:

```env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
AUTH0_REDIRECT_URI=http://localhost:8080/callback
```

Make sure `AUTH0_REDIRECT_URI` is listed in your Auth0 Application → **Allowed Callback URLs**.

Also add your post-logout URL (e.g. `http://localhost:8080/`) to **Allowed Logout URLs** in the Auth0 dashboard.

## Dependencies Injected

You must provide:

- `Logger` — with `Debug`, `Info`, `Error` methods (easy to adapt zap, zerolog, log/slog, etc.)
- `SessionManager` — with `Get(ctx, key)` and `Put(ctx, key, value)` (compatible with gorilla/sessions, etc.)

## Session Storage

The module stores:

- `"user"` → `map[string]any` with decoded ID token claims (sub, name, email, picture, etc.)
- `"access_token"` → raw access token string (useful for calling APIs)

You can extend this as needed in your own handlers.

## Testing

The module is designed for easy testing — all dependencies are interfaces. See the `_test.go` files for examples using mocks.

## License

MIT

