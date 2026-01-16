package auth0

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/derekmwright/web/auth/auth0/authenticator"
)

func init() {
	gob.Register(SessionUser{})
}

type SessionManager interface {
	Get(ctx context.Context, key string) any
	Put(ctx context.Context, key string, value any)
}

type config struct {
	Logger         *slog.Logger
	Sessions       SessionManager
	postLoginHooks []PostLoginHook
}

type deps struct {
	auth           *authenticator.Authenticator
	logoutBase     *url.URL
	log            *slog.Logger
	sessions       SessionManager
	postLoginHooks []PostLoginHook
}

func New(opts ...Option) (func(chi.Router), Middleware, error) {
	cfg := config{
		Logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Logger == nil {
		return nil, nil, ErrNilLogger
	}
	if cfg.Sessions == nil {
		return nil, nil, ErrNilSessions
	}

	auth, err := authenticator.New()
	if err != nil {
		return nil, nil, err
	}

	logoutURL, err := url.Parse(auth.LogoutURL)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse logout URL: %w", err)
	}

	d := &deps{
		log:            cfg.Logger,
		logoutBase:     logoutURL,
		sessions:       cfg.Sessions,
		auth:           auth,
		postLoginHooks: cfg.postLoginHooks,
	}

	mw := func(next http.Handler) http.Handler {
		return authenticatedMiddleware(d, next)
	}

	return func(r chi.Router) {
		Register(r, d)
	}, mw, nil
}
