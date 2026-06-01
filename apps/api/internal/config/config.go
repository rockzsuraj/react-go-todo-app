package config

import (
	"os"
)

type DBConfig struct {
	// Preferred (industry standard)
	DatabaseURL string

	// Fallback (legacy / dev)
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func LoadDBConfig() DBConfig {
	return DBConfig{
		// ✅ Primary (Render / prod / modern)
		DatabaseURL: os.Getenv("DATABASE_URL"),

		// 🔁 Fallback (local / legacy)
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", "postgres"),
		Name:     GetEnv("DB_NAME", "todos"),
		SSLMode:  GetEnv("DB_SSL_MODE", "disable"),
	}
}

// AppConfig holds application-level configuration such as OAuth and JWT settings.
type AppConfig struct {
	Env                string
	Port               string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
	FrontendURL        string
}

// LoadAppConfig loads non-DB related configuration from environment variables.
func LoadAppConfig() AppConfig {
	return AppConfig{
		Env:                GetEnv("ENV", "development"),
		Port:               GetEnv("PORT", "8080"),
		GoogleClientID:     GetEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: GetEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  GetEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/callback/google"),
		JWTSecret:          GetEnv("JWT_SECRET", "dev-jwt-secret"),
		FrontendURL:        GetEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}
