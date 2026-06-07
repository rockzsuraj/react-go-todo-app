package config

import (
	"strings"
	"testing"
)

func validProductionConfig() (AppConfig, DBConfig) {
	return AppConfig{
		Env:                "production",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		GoogleRedirectURL:  "https://example.com/api/auth/callback/google",
		JWTSecret:          "a-production-jwt-secret-with-32-chars",
		FrontendURL:        "https://example.com",
	}, DBConfig{DatabaseURL: "postgresql://example"}
}

func TestValidateProductionConfigAcceptsCompleteConfig(t *testing.T) {
	appCfg, dbCfg := validProductionConfig()

	if err := ValidateProductionConfig(appCfg, dbCfg); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateProductionConfigRequiresJWTSecret(t *testing.T) {
	appCfg, dbCfg := validProductionConfig()
	appCfg.JWTSecret = ""

	err := ValidateProductionConfig(appCfg, dbCfg)
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected JWT_SECRET error, got %v", err)
	}
}

func TestValidateProductionConfigRejectsWeakJWTSecret(t *testing.T) {
	appCfg, dbCfg := validProductionConfig()
	appCfg.JWTSecret = "too-short"

	err := ValidateProductionConfig(appCfg, dbCfg)
	if err == nil || !strings.Contains(err.Error(), "at least 32") {
		t.Fatalf("expected weak JWT_SECRET error, got %v", err)
	}
}

func TestValidateProductionConfigRequiresHTTPSURLs(t *testing.T) {
	appCfg, dbCfg := validProductionConfig()
	appCfg.FrontendURL = "http://example.com"

	err := ValidateProductionConfig(appCfg, dbCfg)
	if err == nil || !strings.Contains(err.Error(), "FRONTEND_URL") {
		t.Fatalf("expected FRONTEND_URL error, got %v", err)
	}
}

func TestValidateProductionConfigAllowsDevelopmentDefaults(t *testing.T) {
	if err := ValidateProductionConfig(AppConfig{Env: "development"}, DBConfig{}); err != nil {
		t.Fatalf("expected development config to pass, got %v", err)
	}
}
