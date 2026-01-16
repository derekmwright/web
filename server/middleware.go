package server

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type requestLoggerKey struct{}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(requestLoggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

func MiddlewareRecovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					log.Error("panic recovered",
						"panic", rec,
						"stack", string(stack),
						"request_id", middleware.GetReqID(r.Context()),
						"path", r.URL.Path,
						"method", r.Method,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
func MiddlewareRequestID() func(http.Handler) http.Handler {
	return middleware.RequestID
}

func MiddlewareLogging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := middleware.GetReqID(r.Context())
			if requestID != "" {
				requestID = uuid.New().String()
			}

			reqLog := log.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			reqLog = reqLog.With("scheme", scheme)

			ctx := context.WithValue(r.Context(), requestLoggerKey{}, requestID)
			r = r.WithContext(ctx)

			defer func() {
				reqLog.Info("request completed",
					"duration_sec", time.Since(start).Seconds(),
					"duration_ms", time.Since(start).Milliseconds(),
					"status", ww.Status(),
					"bytes_written", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
