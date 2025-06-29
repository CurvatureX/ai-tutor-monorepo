package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the gateway configuration
type Config struct {
	Server       ServerConfig
	SpeechService SpeechServiceConfig
	Logger       LoggerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host    string
	Port    string
	Address string
}

// SpeechServiceConfig holds speech service configuration
type SpeechServiceConfig struct {
	Host    string
	Port    string
	Address string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("GATEWAY_HOST", "0.0.0.0"),
			Port: getEnv("GATEWAY_PORT", "8080"),
		},
		SpeechService: SpeechServiceConfig{
			Host: getEnv("SPEECH_SERVICE_HOST", "localhost"),
			Port: getEnv("SPEECH_SERVICE_PORT", "50051"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Construct addresses
	cfg.Server.Address = fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	cfg.SpeechService.Address = fmt.Sprintf("%s:%s", cfg.SpeechService.Host, cfg.SpeechService.Port)

	return cfg, nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}