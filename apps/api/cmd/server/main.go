package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"react-todos/apps/api/internal/config"
	"react-todos/apps/api/internal/db"
	"react-todos/apps/api/internal/handlers"
	"react-todos/apps/api/internal/repository"
	"react-todos/apps/api/internal/routes"
	"react-todos/apps/api/internal/services"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
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

	// Load config
	dbCfg := config.LoadDBConfig()
	port := config.GetEnv("PORT", "8080")

	// Database
	database := db.NewPostgresDB(dbCfg)
	defer database.Close()

	// Redis
	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("redis unavailable, token blacklisting disabled", "error", err)
		redisClient = nil
	}
	defer func() {
		if redisClient != nil {
			redisClient.Close()
		}
	}()

	// Repositories
	todoRepo := repository.NewTodoRepository(database)
	userRepo := repository.NewUserRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	var blacklistRepo services.TokenBlacklistRepository
	if redisClient != nil {
		blacklistRepo = repository.NewRedisBlacklistRepository(redisClient)
	} else {
		blacklistRepo = &repository.NoopBlacklistRepository{}
	}

	// Services
	todoService := services.NewTodoService(todoRepo)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, blacklistRepo)

	// Handlers (INIT ONCE)
	handlers.InitHandlers(todoService)
	handlers.InitAuthHandlers(authService)
	handlers.InitAdminHandlers(authService)

	// Router
	router := routes.SetupRouter(authService)

	// HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server starting",
		"port", port,
		"env", config.GetEnv("ENV", "development"),
	)

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
