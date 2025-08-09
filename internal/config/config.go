package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	BucketName  string
	TokenSecret string
	LogLevel    string
	DocsPath    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		BucketName:  getEnv("BUCKET_NAME", ""),
		TokenSecret: getEnv("TOKEN_SECRET", "default-secret-change-in-production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DocsPath:    getEnv("DOCS_PATH", "/docs"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}