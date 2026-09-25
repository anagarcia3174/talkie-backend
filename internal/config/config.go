package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL       string
	Port              string
	TMDBAPIKey        string
	LogLevel          string
	AllowedOrigin     string
	SupabaseJWKSURL   string
	SupabaseURL       string
	SupabaseSecretKey string
	AppleTeamID       string
	AppleKeyID        string
	AppleClientID     string
	ApplePrivateKey   string
}

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func Load() (*Config, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	cfg := &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Port:              getOrDefault("PORT", "8080"),
		TMDBAPIKey:        os.Getenv("TMDB_API_KEY"),
		LogLevel:          getOrDefault("LOG_LEVEL", "info"),
		AllowedOrigin:     os.Getenv("ALLOWED_ORIGIN"),
		SupabaseJWKSURL:   supabaseURL + "/auth/v1/.well-known/jwks.json",
		SupabaseURL:       supabaseURL,
		SupabaseSecretKey: os.Getenv("SUPABASE_SECRET_KEY"),
		AppleTeamID:       os.Getenv("APPLE_TEAM_ID"),
		AppleKeyID:        os.Getenv("APPLE_KEY_ID"),
		AppleClientID:     os.Getenv("APPLE_CLIENT_ID"),
		ApplePrivateKey:   os.Getenv("APPLE_PRIVATE_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.TMDBAPIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is required")
	}

	if cfg.AllowedOrigin == "" {
		return nil, fmt.Errorf("ALLOWED_ORIGIN is required")
	}

	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}

	for name, value := range map[string]string{
		"SUPABASE_SECRET_KEY": cfg.SupabaseSecretKey,
		"APPLE_TEAM_ID":       cfg.AppleTeamID,
		"APPLE_KEY_ID":        cfg.AppleKeyID,
		"APPLE_CLIENT_ID":     cfg.AppleClientID,
		"APPLE_PRIVATE_KEY":   cfg.ApplePrivateKey,
	} {
		if value == "" {
			return nil, fmt.Errorf("%s is required", name)
		}
	}

	if !validLogLevels[cfg.LogLevel] {
		return nil, fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error (got %q)", cfg.LogLevel)
	}

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
