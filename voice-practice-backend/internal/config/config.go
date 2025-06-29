package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig
	ASR    ASRConfig
	LLM    LLMConfig
	TTS    TTSConfig
	Audio  AudioConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// ASRConfig holds ASR service configuration
type ASRConfig struct {
	AccessKey string
	AppKey    string
	BaseURL   string
}

// LLMConfig holds LLM service configuration
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// TTSConfig holds TTS service configuration
type TTSConfig struct {
	AppID    string
	Token    string
	Cluster  string
	BaseURL  string
	Voice    string
	Language string
}

// AudioConfig holds audio processing configuration
type AudioConfig struct {
	ChunkSize  int
	BufferSize int
	SampleRate int
	Channels   int
	BitDepth   int
}

// Load loads configuration from .env file and environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log if it's not a "file not found" error
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
		ASR: ASRConfig{
			AccessKey: getEnv("ASR_ACCESS_KEY", ""),
			AppKey:    getEnv("ASR_APP_KEY", ""),
			BaseURL:   getEnv("ASR_BASE_URL", ""),
		},
		LLM: LLMConfig{
			APIKey:  getEnv("LLM_API_KEY", ""),
			BaseURL: getEnv("LLM_BASE_URL", ""),
			Model:   getEnv("LLM_MODEL", "doubao-pro-4k"),
		},
		TTS: TTSConfig{
			AppID:    getEnv("TTS_APP_ID", ""),
			Token:    getEnv("TTS_TOKEN", ""),
			Cluster:  getEnv("TTS_CLUSTER", "volcano_tts"),
			BaseURL:  getEnv("TTS_BASE_URL", ""),
			Voice:    getEnv("TTS_VOICE", "en_us_001"),
			Language: getEnv("TTS_LANGUAGE", "en"),
		},
		Audio: AudioConfig{
			ChunkSize:  getEnvInt("AUDIO_CHUNK_SIZE", 4096),
			BufferSize: getEnvInt("AUDIO_BUFFER_SIZE", 16384),
			SampleRate: getEnvInt("AUDIO_SAMPLE_RATE", 16000),
			Channels:   getEnvInt("AUDIO_CHANNELS", 1),
			BitDepth:   getEnvInt("AUDIO_BIT_DEPTH", 16),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
