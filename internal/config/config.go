package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL   string
	Port          string
	TMDBAPIKey    string
	LogLevel      string
	AllowedOrigin string
	SupabaseJWKSURL string
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
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          getOrDefault("PORT", "8080"),
		TMDBAPIKey:    os.Getenv("TMDB_API_KEY"),
		LogLevel:      getOrDefault("LOG_LEVEL", "info"),
		AllowedOrigin: os.Getenv("ALLOWED_ORIGIN"),
		SupabaseJWKSURL: supabaseURL + "/auth/v1/.well-known/jwks.json",
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
