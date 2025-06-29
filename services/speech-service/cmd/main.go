package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ai-tutor-monorepo/services/speech-service/internal/config"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/handler"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/service"
	speechv1 "github.com/ai-tutor-monorepo/services/speech-service/pkg/proto/speech"
)

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg := config.Load()

	// Create debug directory for audio files
	if err := os.MkdirAll("debug", 0755); err != nil {
		logger.Warnf("Failed to create debug directory: %v", err)
	}

	// Initialize services
	audioService := service.NewAudioService(&cfg.Audio, logger)
	asrService := service.NewASRService(&cfg.ASR, logger)
	llmService := service.NewLLMService(&cfg.LLM, logger)
	ttsService := service.NewTTSService(&cfg.TTS, logger)

	// Initialize gRPC handler
	speechHandler := handler.NewSpeechHandler(audioService, asrService, llmService, ttsService, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	speechv1.RegisterSpeechServiceServer(grpcServer, speechHandler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	logger.Infof("Starting Speech Service gRPC server on %s", addr)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Speech Service...")
	grpcServer.GracefulStop()
	logger.Info("Speech Service stopped")
}