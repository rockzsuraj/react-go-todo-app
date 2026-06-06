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
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	// Repositories
	todoRepo := repository.NewTodoRepository(database)
	userRepo := repository.NewUserRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	oauthStateRepo := repository.NewOAuthStateRepository(database)
	blacklistRepo := repository.NewPostgresBlacklistRepository(database)

	// Services
	todoService := services.NewTodoService(todoRepo)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, blacklistRepo)

	// Handlers
	handlers.InitHandlers(todoService)
	handlers.InitAuthHandlers(authService, oauthStateRepo)
	handlers.InitAdminHandlers(authService)

	// Router
	readinessCheck := func(ctx context.Context) error {
		if err := database.Ping(ctx); err != nil {
			return fmt.Errorf("postgres: %w", err)
		}
		return nil
	}
	router := routes.SetupRouter(authService, readinessCheck)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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
