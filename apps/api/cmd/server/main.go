package main

import (
	"context"
	"fmt"
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
	appCfg := config.LoadAppConfig()
	if err := config.ValidateProductionConfig(appCfg, dbCfg); err != nil {
		logger.Error("invalid production configuration", "error", err)
		os.Exit(1)
	}
	port := config.GetEnv("PORT", "8080")

	// Database
	database := db.NewPostgresDB(dbCfg)
	defer database.Close()

	// Migrations
	if err := db.RunMigrations(ctx, database, "./migrations"); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// Redis
	var redisClient *redis.Client
	if appCfg.RedisURL != "" {
		opt, err := redis.ParseURL(appCfg.RedisURL)
		if err != nil {
			logger.Error("failed to parse REDIS_URL", "error", err)
			os.Exit(1)
		}
		redisClient = redis.NewClient(opt)
	} else {
		redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
	}

	redisCtx, redisCancel := context.WithTimeout(ctx, 5*time.Second)
	defer redisCancel()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		if appCfg.Env == "production" {
			logger.Error("redis unavailable in production", "error", err)
			os.Exit(1)
		}
		logger.Warn("redis unavailable, token blacklisting disabled", "error", err)
		_ = redisClient.Close()
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
	handlers.InitAuthHandlers(authService, redisClient)
	handlers.InitAdminHandlers(authService)

	// Router
	readinessCheck := func(ctx context.Context) error {
		if err := database.Ping(ctx); err != nil {
			return fmt.Errorf("postgres: %w", err)
		}
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("redis: %w", err)
			}
		}
		return nil
	}
	router := routes.SetupRouter(authService, readinessCheck)

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
		"env", appCfg.Env,
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
