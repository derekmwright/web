package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	Addr        string
	TLSKeyFile  string
	TLSCertFile string
}
type Server struct {
	srv    *http.Server
	Log    *slog.Logger
	Router chi.Router

	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

func New(opts ...Option) *Server {
	s := &Server{
		Log:             slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		Router:          chi.NewRouter(),
		shutdownTimeout: 10 * time.Second,
		readTimeout:     5 * time.Second,
		writeTimeout:    10 * time.Second,
		idleTimeout:     30 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.Router.Use(
		MiddlewareRecovery(s.Log),
		MiddlewareRequestID(),
		MiddlewareLogging(s.Log),
	)

	s.Router.Get("/healthz", HealthHandler)
	s.Router.Get("/readyz", ReadinessHandler)

	s.srv = &http.Server{
		Addr:         getAddr(),
		Handler:      s.Router,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	return s
}

func getAddr() string {
	addr := os.Getenv("APP_SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return addr
}

func getTLSKey() string  { return os.Getenv("APP_SERVER_TLS_KEY_FILE") }
func getTLSCert() string { return os.Getenv("APP_SERVER_TLS_CERT_FILE") }

func isTLSEnabled() bool {
	return getTLSKey() != "" && getTLSCert() != ""
}

func (s *Server) Start() error {
	s.Log.Info("starting server", "addr", s.srv.Addr)

	go func() {
		var err error
		if isTLSEnabled() {
			err = s.srv.ListenAndServeTLS(getTLSCert(), getTLSKey())
		} else {
			err = s.srv.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Log.Error("server failed to start", "error", err)
		}
	}()

	return s.waitForShutdown()
}

func (s *Server) LoggerWithContext(ctx context.Context) *slog.Logger {
	return LoggerFromContext(ctx)
}

func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.Log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		s.Log.Error("server failed to shutdown", "error", err)
		return err
	}

	s.Log.Info("server shut down successfully")
	return nil
}
