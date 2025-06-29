package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"voice-practice-backend/internal/config"
	"voice-practice-backend/internal/handler"
	"voice-practice-backend/internal/service"
	wsManager "voice-practice-backend/pkg/websocket"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create debug directory for audio files
	if err := os.MkdirAll("debug", 0755); err != nil {
		fmt.Printf("Warning: Failed to create debug directory: %v\n", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})

	// Initialize services
	audioService := service.NewAudioService(&cfg.Audio, logger)
	asrService := service.NewASRService(&cfg.ASR, logger)
	llmService := service.NewLLMService(&cfg.LLM, logger)
	ttsService := service.NewTTSService(&cfg.TTS, logger)

	// Initialize WebSocket manager
	manager := wsManager.NewManager(logger)

	// Initialize handlers
	wsHandler := handler.NewWebSocketHandler(manager, audioService, asrService, llmService, ttsService, logger)
	healthHandler := handler.NewHealthHandler(manager)

	// Setup Gin router
	router := gin.Default()

	// Enable CORS for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoints
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Serve static files (frontend)
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Start cleanup routine for inactive sessions
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			manager.CleanupInactiveSessions(30 * time.Minute)
		}
	}()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.Infof("Starting voice practice server on %s", addr)
	logger.Infof("WebSocket endpoint: ws://%s/ws", addr)
	logger.Infof("Health check: http://%s/health", addr)

	if err := router.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
