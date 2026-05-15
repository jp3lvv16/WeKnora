// Package config provides configuration management for WeKnora.
// It handles loading, validation, and access to application configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration values.
type Config struct {
	// Server settings
	ServerHost string
	ServerPort int
	DebugMode  bool

	// Database settings
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Redis settings
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT settings
	JWTSecret          string
	JWTExpirationHours time.Duration

	// Application settings
	AppName    string
	AppVersion string
	LogLevel   string
}

// Load reads configuration from environment variables and returns a Config instance.
// It returns an error if any required configuration values are missing or invalid.
func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),
		DebugMode:  getEnvAsBool("DEBUG_MODE", false),

		// Database defaults
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBName:     getEnv("DB_NAME", "weknora"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis defaults
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// JWT defaults
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpirationHours: time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,

		// Application defaults
		AppName:    getEnv("APP_NAME", "WeKnora"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks that all required configuration values are present and valid.
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", c.ServerPort)
	}
	return nil
}

// DSN returns the PostgreSQL data source name string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// ServerAddr returns the full server address string.
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultVal int) int {
	if val, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvAsBool retrieves an environment variable as a boolean or returns a default value.
func getEnvAsBool(key string, defaultVal bool) bool {
	if val, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
