package auth0

import (
	"context"
	"log/slog"
	"net/http"
)

type Option func(deps *config)

type PostLoginHook func(ctx context.Context, user *SessionUser, r *http.Request) error

func WithLogger(l *slog.Logger) Option {
	return func(cfg *config) {
		cfg.Logger = l
	}
}

func WithSessions(s SessionManager) Option {
	return func(cfg *config) {
		cfg.Sessions = s
	}
}

func WithPostLoginHooks(hooks ...PostLoginHook) Option {
	return func(c *config) {
		c.postLoginHooks = append(c.postLoginHooks, hooks...)
	}
}
