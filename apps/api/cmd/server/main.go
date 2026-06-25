package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"react-todos/apps/api/internal/delivery/http/handlers"
	appMiddleware "react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/delivery/http/router"
	"react-todos/apps/api/internal/infrastructure/config"
	"react-todos/apps/api/internal/infrastructure/db"
	"react-todos/apps/api/internal/infrastructure/events"
	"react-todos/apps/api/internal/infrastructure/repository"
	authUsecase "react-todos/apps/api/internal/usecase/auth"
	todoUsecase "react-todos/apps/api/internal/usecase/todo"

	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load config once — handler structs receive it as a value, never re-read env.
	dbCfg := config.LoadDBConfig()
	appCfg := config.LoadAppConfig()
	if err := config.ValidateProductionConfig(appCfg, dbCfg); err != nil {
		logger.Error("invalid production configuration", "error", err)
		os.Exit(1)
	}
	port := config.GetEnv("PORT", "8080")

	database := db.NewPostgresDB(dbCfg)
	defer database.Close()

	if err := db.RunMigrations(ctx, database, "./migrations"); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// Initialize Redis with connection pool
	var redisClient *redis.Client
	if appCfg.RedisAddr != "" {
		var err error
		redisClient, err = db.NewRedisClient(appCfg.RedisAddr)
		if err != nil {
			logger.Warn("running without Redis cache, falling back to local/in-memory", "error", err)
		} else {
			defer redisClient.Close()
		}
	}

	// Repositories
	todoRepo := repository.NewTodoRepository(database)
	userRepo := repository.NewUserRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	oauthStateRepo := repository.NewOAuthStateRepository(database)
	blacklistRepo := repository.NewCachedBlacklistRepository(database, redisClient)

	// Broadcaster/PubSub for real-time events
	broadcaster := events.NewBroadcaster()

	// Services
	todoService := todoUsecase.NewTodoService(todoRepo, broadcaster)
	authService := authUsecase.NewAuthService(userRepo, refreshTokenRepo, blacklistRepo)

	// Handlers — constructed with all dependencies at startup.
	// No Init* functions, no package-level globals, no per-request config reads.
	authHandler := handlers.NewAuthHandler(authService, appCfg, oauthStateRepo)
	todoHandler := handlers.NewTodoHandler(todoService)
	sseHandler := handlers.NewSSEHandler(broadcaster)
	adminHandler := handlers.NewAdminHandler(authService)

	// Rate Limiter
	rateLimiter := appMiddleware.NewRateLimiter(redisClient)

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
	router := router.SetupRouter(appCfg, authService, authHandler, todoHandler, sseHandler, adminHandler, rateLimiter, readinessCheck)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Bypassed for SSE connections
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("server starting", "port", port, "env", appCfg.Env)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

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
