package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration sourced from the environment.
type Config struct {
	Port            string
	DatabaseURL     string
	TickRate        int
	MatchDurationMs int
}

// Load reads configuration from environment variables, applying defaults.
func Load() Config {
	return Config{
		Port:            getenv("APP_PORT", "8080"),
		DatabaseURL:     getenv("DATABASE_URL", ""),
		TickRate:        getenvInt("TICK_RATE", 20),
		MatchDurationMs: getenvInt("MATCH_DURATION_MS", 120000),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
