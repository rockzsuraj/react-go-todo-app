package config

import (
	"fmt"
	"net/url"
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
	MobileRedirectURI  string
	JWTSecret          string
	FrontendURL        string
}

// LoadAppConfig loads non-DB related configuration from environment variables.
func LoadAppConfig() AppConfig {
	env := GetEnv("ENV", GetEnv("NODE_ENV", "development"))
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && env != "production" {
		jwtSecret = "dev-jwt-secret"
	}

	return AppConfig{
		Env:                env,
		Port:               GetEnv("PORT", "8080"),
		GoogleClientID:     GetEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: GetEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  GetEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/callback/google"),
		MobileRedirectURI:  GetEnv("MOBILE_REDIRECT_URI", "todoapp://oauth/callback"),
		JWTSecret:          jwtSecret,
		FrontendURL:        GetEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func ValidateProductionConfig(cfg AppConfig, dbCfg DBConfig) error {
	if cfg.Env != "production" {
		return nil
	}

	required := map[string]string{
		"DATABASE_URL":         dbCfg.DatabaseURL,
		"GOOGLE_CLIENT_ID":     cfg.GoogleClientID,
		"GOOGLE_CLIENT_SECRET": cfg.GoogleClientSecret,
		"GOOGLE_REDIRECT_URL":  cfg.GoogleRedirectURL,
		"JWT_SECRET":           cfg.JWTSecret,
		"FRONTEND_URL":         cfg.FrontendURL,
	}
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required in production", name)
		}
	}
	if len(cfg.JWTSecret) < 32 || cfg.JWTSecret == "dev-jwt-secret" {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters and non-default in production")
	}

	for name, rawURL := range map[string]string{
		"GOOGLE_REDIRECT_URL": cfg.GoogleRedirectURL,
		"FRONTEND_URL":        cfg.FrontendURL,
	} {
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must be a valid HTTPS URL in production", name)
		}
	}

	return nil
}
