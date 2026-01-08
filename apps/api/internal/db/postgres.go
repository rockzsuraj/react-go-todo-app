package db

import (
	"context"
	"fmt"
	"log"
	"react-todos/apps/api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(cfg config.DBConfig) *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return db
}
