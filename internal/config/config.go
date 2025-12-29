package config

import (
	"os"
	"strconv"
)

// Config holds configuration parameters
type Config struct {
	Port          int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	APIKeys       string // Comma separated keys
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	port := 8081
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	redisDB := 0
	if db := os.Getenv("REDIS_DB"); db != "" {
		if parsed, err := strconv.Atoi(db); err == nil {
			redisDB = parsed
		}
	}

	return &Config{
		Port:          port,
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		APIKeys:       getEnv("CHAOS_API_KEYS", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
