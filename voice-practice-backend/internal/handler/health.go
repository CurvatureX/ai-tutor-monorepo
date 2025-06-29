package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	wsManager "voice-practice-backend/pkg/websocket"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	manager *wsManager.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(manager *wsManager.Manager) *HealthHandler {
	return &HealthHandler{
		manager: manager,
	}
}

// HealthCheck returns the health status of the service
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":            "ok",
		"timestamp":         time.Now().Unix(),
		"active_sessions":   h.manager.GetActiveSessionCount(),
		"service":          "voice-practice-backend",
		"version":          "1.0.0",
	})
}

// ReadinessCheck returns the readiness status of the service
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ready",
		"timestamp":   time.Now().Unix(),
		"service":     "voice-practice-backend",
	})
}