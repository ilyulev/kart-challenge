package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port     string
	LogLevel string
	APIKey   string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		APIKey:   getEnv("API_KEY", "apitest"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
