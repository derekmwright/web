package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Option func(*Server)

func WithLogger(logger *slog.Logger) Option {
	return func(server *Server) { server.Log = logger }
}

func WithRouter(router chi.Router) Option {
	return func(server *Server) { server.Router = router }
}

func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) { s.shutdownTimeout = d }
}

func WithReadTimeout(d time.Duration) Option {
	return func(server *Server) { server.readTimeout = d }
}

func WithWriteTimeout(d time.Duration) Option {
	return func(server *Server) { server.writeTimeout = d }
}

func WithIdleTimeout(d time.Duration) Option {
	return func(server *Server) { server.idleTimeout = d }
}

func WithMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(server *Server) { server.Router.Use(mw) }
}
