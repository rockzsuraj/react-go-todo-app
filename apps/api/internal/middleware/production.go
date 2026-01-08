package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// StructuredLogger provides structured logging middleware
func StructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&structuredLogger{logger})
}

type structuredLogger struct {
	Logger *slog.Logger
}

func (l *structuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &structuredLoggerEntry{Logger: l.Logger}
	entry.Logger = entry.Logger.With(slog.Any("method", r.Method), slog.Any("path", r.URL.Path), slog.Any("remote_addr", r.RemoteAddr), slog.Any("user_agent", r.UserAgent()))
	return entry
}

type structuredLoggerEntry struct {
	Logger *slog.Logger
}

func (l *structuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger.Info("request completed",
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Duration("elapsed", elapsed),
	)
}

func (l *structuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger.Error("request panic",
		slog.Any("panic", v),
		slog.String("stack", string(stack)),
	)
}

// SecurityHeaders adds security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}