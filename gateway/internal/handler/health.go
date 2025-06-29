package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	speechv1 "github.com/ai-tutor-monorepo/gateway/pkg/proto/speech"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	speechClient speechv1.SpeechServiceClient
	logger       *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(speechClient speechv1.SpeechServiceClient, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		speechClient: speechClient,
		logger:       logger,
	}
}

// HealthCheck performs a basic health check
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "gateway",
	})
}

// ReadinessCheck performs a readiness check including downstream services
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check speech service health
	speechHealth := h.checkSpeechServiceHealth(ctx)

	status := "ready"
	statusCode := http.StatusOK
	
	if !speechHealth {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().Unix(),
		"service":   "gateway",
		"dependencies": gin.H{
			"speech_service": speechHealth,
		},
	})
}

// checkSpeechServiceHealth checks if the speech service is healthy
func (h *HealthHandler) checkSpeechServiceHealth(ctx context.Context) bool {
	req := &speechv1.HealthCheckRequest{
		Service: "speech",
	}

	resp, err := h.speechClient.HealthCheck(ctx, req)
	if err != nil {
		h.logger.Errorf("Speech service health check failed: %v", err)
		return false
	}

	return resp.Status == speechv1.HealthStatus_HEALTH_STATUS_SERVING
}