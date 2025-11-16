package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	RateLimit RateLimitConfig
	CORS     CORSConfig
	AppEnv   string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                string
	AccessTokenExpiration time.Duration
	RefreshTokenExpiration time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerDuration int
	Duration            time.Duration
	CleanupInterval     time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (for development)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "localhost"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "15s")),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "15s")),
			IdleTimeout:  parseDuration(getEnv("SERVER_IDLE_TIMEOUT", "60s")),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "gojwt_db"),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", ""),
			AccessTokenExpiration:  parseDuration(getEnv("JWT_ACCESS_EXPIRATION", "15m")),
			RefreshTokenExpiration: parseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h")), // 7 days
		},
		RateLimit: RateLimitConfig{
			RequestsPerDuration: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Duration:            parseDuration(getEnv("RATE_LIMIT_DURATION", "1m")),
			CleanupInterval:     parseDuration(getEnv("RATE_LIMIT_CLEANUP_INTERVAL", "1m")),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
		AppEnv: getEnv("APP_ENV", "development"),
	}

	// Validate required fields
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets environment variable as integer with fallback
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// parseDuration parses duration string with fallback
func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return time.Minute
	}
	return duration
}

// GetDSN returns MySQL DSN string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
	)
}
