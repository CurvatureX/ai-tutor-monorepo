package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ai-tutor-monorepo/gateway/internal/config"
	"github.com/ai-tutor-monorepo/gateway/internal/handler"
	"github.com/ai-tutor-monorepo/gateway/internal/manager"
	speechv1 "github.com/ai-tutor-monorepo/gateway/pkg/proto/speech"
)

func main() {
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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}
	logger.Infof("Loaded configuration: %+v", cfg)

	// Connect to speech service
	conn, err := grpc.Dial(cfg.SpeechService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("Failed to connect to speech service: %v", err)
	}
	defer conn.Close()

	speechClient := speechv1.NewSpeechServiceClient(conn)

	// Initialize WebSocket manager
	wsManager := manager.NewWebSocketManager(logger)

	// Initialize handlers
	wsHandler := handler.NewEnhancedWebSocketHandler(wsManager, speechClient, logger)
	healthHandler := handler.NewHealthHandler(speechClient, logger)

	// Setup Gin router
	router := gin.Default()

	// Enable CORS
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

	// Routes
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Serve static files
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Start server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting Gateway server on %s", cfg.Server.Address)
		logger.Infof("WebSocket endpoint: ws://%s/ws", cfg.Server.Address)
		logger.Infof("Health check: http://%s/health", cfg.Server.Address)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start session cleanup routine
	go wsManager.StartCleanupRoutine(5*time.Minute, 30*time.Minute)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Gateway server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	wsManager.Shutdown()
	logger.Info("Gateway server exited")
}
