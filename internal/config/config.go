package config

import (
	"log"
	"os"
)

// Config holds all application configuration
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	SessionKey         string
	AppBaseURL         string
	MongoDBURI         string
	ServerPort         string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionKey:         os.Getenv("SESSION_KEY"),
		AppBaseURL:         os.Getenv("APP_BASE_URL"),
		MongoDBURI:         os.Getenv("DB_DSN"),
		ServerPort:         getEnvOrDefault("SERVER_PORT", "8080"),
	}

	// Validate required configuration
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" ||
		cfg.SessionKey == "" || cfg.AppBaseURL == "" || cfg.MongoDBURI == "" {
		log.Fatal("Required environment variables not set: GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, SESSION_KEY, APP_BASE_URL, DB_DSN")
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
