package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/db"
	"react-todos/apps/api/internal/handlers"
	"react-todos/apps/api/internal/repository"
	"react-todos/apps/api/internal/routes"
	"react-todos/apps/api/internal/services"
)

func main() {
	// Structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Graceful shutdown context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	cfg := config.LoadDBConfig()
	port := getEnv("PORT", "8080")

	// Initialize database with context
	database := db.NewPostgresDB(cfg)
	defer database.Close()

	// Initialize repository, service and handlers
	todoRepo := repository.NewTodoRepository(database)
	todoService := services.NewTodoService(todoRepo)
	handlers.InitHandlers(todoService)

	// Setup routes with middleware
	r := routes.SetupRouter()

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server starting", "port", port, "env", getEnv("ENV", "development"))

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
