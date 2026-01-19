package db

import (
	"context"
	"fmt"
	"log"

	"react-todos/apps/api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(cfg config.DBConfig) *pgxpool.Pool {
	var dsn string

	// ✅ Preferred path: DATABASE_URL (prod / Render / cloud)
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
		log.Println("🔐 Using DATABASE_URL for DB connection")
	} else {
		// 🔁 Fallback: individual DB_* vars (local dev)
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
			cfg.SSLMode,
		)
		log.Println("⚙️ Using DB_* variables for DB connection")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("❌ Failed to create DB pool:", err)
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatal("❌ DB ping failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return db
}
