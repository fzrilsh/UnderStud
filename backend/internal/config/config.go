package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration. Every field has a default so that
// zero env vars are required — `docker-compose up` must always be enough.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
	CORSOrigins []string
	MaxDBConns  int32
	MinDBConns  int32
	LogLevel    string
}

// Load reads configuration from environment variables with sane defaults.
func Load() (Config, error) {
	cfg := Config{
		Port: getEnv("PORT", "8080"),
		// 127.0.0.1 + port 5433: hindari Postgres native di host yang biasanya
		// memakai 5432 (dan bisa nabrak resolusi localhost→IPv6 ::1).
		// Port 5433 = port host default container db (lihat project/docker-compose.yml).
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5433/hackathon?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me-in-prod"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000"), ","),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	ttl, err := time.ParseDuration(getEnv("AUTH_TOKEN_TTL", "87600h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid AUTH_TOKEN_TTL: %w", err)
	}
	cfg.TokenTTL = ttl

	if cfg.MaxDBConns, err = parseInt32Env("DB_MAX_CONNS", 10); err != nil {
		return Config{}, err
	}
	if cfg.MinDBConns, err = parseInt32Env("DB_MIN_CONNS", 2); err != nil {
		return Config{}, err
	}

	for i := range cfg.CORSOrigins {
		cfg.CORSOrigins[i] = strings.TrimSpace(cfg.CORSOrigins[i])
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt32Env(key string, fallback int32) (int32, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return int32(n), nil
}
