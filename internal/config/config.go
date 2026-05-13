package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL    string
	Port           string
	JWTSecret      string
	JWTExpiry      time.Duration
	RefreshSecret  string
	RefreshExpiry  time.Duration
}

func Load() *Config {
	jwtHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	refreshDays, _ := strconv.Atoi(getEnv("REFRESH_EXPIRY_DAYS", "7"))

	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/ecommerce?sslmode=disable"),
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiry:     time.Duration(jwtHours) * time.Hour,
		RefreshSecret: getEnv("REFRESH_SECRET", "change-this-refresh-secret-in-production"),
		RefreshExpiry: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
