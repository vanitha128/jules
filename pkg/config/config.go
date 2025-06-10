package config

import (
	"os"
	"strconv"
	// For a more robust solution, consider libraries like godotenv for .env file loading
	// and Viper for more advanced configuration management.
	// For this exercise, os.Getenv will suffice.
	// "github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	ServerPort    string // e.g., ":8080"
}

// LoadConfig loads configuration from environment variables or uses default values.
func LoadConfig() (*Config, error) {
	// Example: Load .env file if present (useful for local development)
	// godotenv.Load() // This will ignore error if .env file doesn't exist.

	cfg := &Config{
		PostgresDSN:   getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=go_moon_db port=5432 sslmode=disable TimeZone=UTC"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""), // Default to no password
		RedisDB:       getEnvAsInt("REDIS_DB", 0),   // Default to DB 0
		JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-key-for-jwt-!@#$%^&*()_+"), // CHANGE THIS IN PRODUCTION!
		ServerPort:    getEnv("SERVER_PORT", ":8080"),
	}

	// Basic validation (e.g., ensure JWTSecret is not the default in a "production" env)
	// if getEnv("APP_ENV", "development") == "production" && cfg.JWTSecret == "your-super-secret-key-for-jwt-!@#$%^&*()_+" {
	// 	return nil, errors.New("default JWT_SECRET must be changed in production")
	// }

	return cfg, nil
}

// getEnv is a helper function to read an environment variable or return a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt is a helper function to read an environment variable as an integer or return a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
