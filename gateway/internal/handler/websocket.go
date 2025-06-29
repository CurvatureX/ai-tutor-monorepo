package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/ai-tutor-monorepo/gateway/internal/manager"
	"github.com/ai-tutor-monorepo/gateway/internal/model"
	speechv1 "github.com/ai-tutor-monorepo/gateway/pkg/proto/speech"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketHandler handles WebSocket connections and bridges to gRPC
type WebSocketHandler struct {
	manager      *manager.WebSocketManager
	speechClient speechv1.SpeechServiceClient
	logger       *logrus.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	manager *manager.WebSocketManager,
	speechClient speechv1.SpeechServiceClient,
	logger *logrus.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		manager:      manager,
		speechClient: speechClient,
		logger:       logger,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	h.logger.Infof("WebSocket connection request for session: %s", sessionID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	h.manager.AddConnection(sessionID, conn)
	defer h.manager.RemoveConnection(sessionID)

	h.logger.Infof("WebSocket connection established for session: %s", sessionID)

	// Send welcome message
	welcomeMsg := &model.WebSocketMessage{
		Type:    model.MessageTypeText,
		Data:    "Welcome to AI English Practice! Start speaking to begin your practice session.",
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, welcomeMsg)

	// Start gRPC stream for this session
	go h.handleGRPCStream(sessionID)

	// Handle incoming WebSocket messages
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			h.logger.Errorf("WebSocket read error for session %s: %v", sessionID, err)
			break
		}

		h.logger.Debugf("Received message for session %s: type=%d, size=%d", sessionID, messageType, len(data))

		switch messageType {
		case websocket.TextMessage:
			h.handleTextMessage(sessionID, data)
		case websocket.BinaryMessage:
			h.handleBinaryMessage(sessionID, data)
		default:
			h.logger.Warnf("Unknown message type %d for session %s", messageType, sessionID)
		}
	}
}

// handleGRPCStream establishes and manages the gRPC stream for a session
func (h *WebSocketHandler) handleGRPCStream(sessionID string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := h.speechClient.ProcessVoiceConversation(ctx)
	if err != nil {
		h.logger.Errorf("Failed to create gRPC stream for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Failed to initialize voice processing")
		return
	}

	// Start goroutine to handle responses from gRPC service
	go h.handleGRPCResponses(sessionID, stream)

	// Keep the stream alive and handle session lifecycle
	for {
		session, exists := h.manager.GetSession(sessionID)
		if !exists {
			h.logger.Infof("Session %s ended, closing gRPC stream", sessionID)
			break
		}

		// Send periodic heartbeat if needed
		time.Sleep(time.Second)
		_ = session // Keep session reference
	}

	if err := stream.CloseSend(); err != nil {
		h.logger.Errorf("Failed to close gRPC stream for session %s: %v", sessionID, err)
	}
}

// handleGRPCResponses handles responses from the gRPC service
func (h *WebSocketHandler) handleGRPCResponses(sessionID string, stream speechv1.SpeechService_ProcessVoiceConversationClient) {
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			h.logger.Infof("gRPC stream ended for session %s", sessionID)
			break
		}
		if err != nil {
			h.logger.Errorf("gRPC stream error for session %s: %v", sessionID, err)
			h.sendErrorMessage(sessionID, "Voice processing error")
			break
		}

		h.processGRPCResponse(sessionID, response)
	}
}

// processGRPCResponse processes a response from the gRPC service
func (h *WebSocketHandler) processGRPCResponse(sessionID string, response *speechv1.VoiceResponse) {
	switch result := response.ResponseType.(type) {
	case *speechv1.VoiceResponse_AsrResult:
		h.handleASRResult(sessionID, result.AsrResult)
	case *speechv1.VoiceResponse_LlmResult:
		h.handleLLMResult(sessionID, result.LlmResult)
	case *speechv1.VoiceResponse_TtsResult:
		h.handleTTSResult(sessionID, result.TtsResult)
	case *speechv1.VoiceResponse_IseResult:
		h.handleISEResult(sessionID, result.IseResult)
	case *speechv1.VoiceResponse_Error:
		h.handleGRPCError(sessionID, result.Error)
	case *speechv1.VoiceResponse_StatusResult:
		h.handleStatusResult(sessionID, result.StatusResult)
	default:
		h.logger.Warnf("Unknown gRPC response type for session %s", sessionID)
	}
}

// handleASRResult handles ASR results from gRPC service
func (h *WebSocketHandler) handleASRResult(sessionID string, result *speechv1.ASRResult) {
	message := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type":       "asr_result",
			"text":       result.Text,
			"confidence": result.Confidence,
			"is_final":   result.IsFinal,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleLLMResult handles LLM results from gRPC service
func (h *WebSocketHandler) handleLLMResult(sessionID string, result *speechv1.LLMResult) {
	message := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type":       "llm_response",
			"text":       result.Text,
			"context":    result.Context,
			"confidence": result.Confidence,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleTTSResult handles TTS results from gRPC service
func (h *WebSocketHandler) handleTTSResult(sessionID string, result *speechv1.TTSResult) {
	// Send binary audio data
	h.manager.SendBinaryMessage(sessionID, result.AudioData)

	// Send notification
	message := &model.WebSocketMessage{
		Type: model.MessageTypeStatus,
		Data: map[string]interface{}{
			"type":        "tts_ready",
			"format":      result.Format.Codec,
			"duration_ms": result.DurationMs,
			"is_final":    result.IsFinal,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleISEResult handles ISE evaluation results from gRPC service
func (h *WebSocketHandler) handleISEResult(sessionID string, result *speechv1.ISEResult) {
	// Convert word scores
	wordScores := make([]map[string]interface{}, len(result.WordScores))
	for i, ws := range result.WordScores {
		wordScores[i] = map[string]interface{}{
			"word":       ws.Word,
			"score":      ws.Score,
			"start_time": ws.StartTime,
			"end_time":   ws.EndTime,
			"is_correct": ws.IsCorrect,
			"confidence": ws.Confidence,
		}
	}

	// Convert phone scores
	phoneScores := make([]map[string]interface{}, len(result.PhoneScores))
	for i, ps := range result.PhoneScores {
		phoneScores[i] = map[string]interface{}{
			"phone":      ps.Phone,
			"score":      ps.Score,
			"start_time": ps.StartTime,
			"end_time":   ps.EndTime,
			"is_correct": ps.IsCorrect,
		}
	}

	// Convert sentence scores
	sentenceScores := make([]map[string]interface{}, len(result.SentenceScores))
	for i, ss := range result.SentenceScores {
		sentenceScores[i] = map[string]interface{}{
			"sentence":       ss.Sentence,
			"score":          ss.Score,
			"accuracy_score": ss.AccuracyScore,
			"fluency_score":  ss.FluencyScore,
			"total_words":    ss.TotalWords,
			"correct_words":  ss.CorrectWords,
		}
	}

	message := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type":               "ise_result",
			"overall_score":      result.OverallScore,
			"accuracy_score":     result.AccuracyScore,
			"fluency_score":      result.FluencyScore,
			"completeness_score": result.CompletenessScore,
			"word_scores":        wordScores,
			"phone_scores":       phoneScores,
			"sentence_scores":    sentenceScores,
			"is_final":           result.IsFinal,
			"reference_text":     result.ReferenceText,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleGRPCError handles errors from gRPC service
func (h *WebSocketHandler) handleGRPCError(sessionID string, error *speechv1.ErrorResult) {
	message := &model.WebSocketMessage{
		Type: model.MessageTypeError,
		Data: map[string]interface{}{
			"code":      error.Code.String(),
			"message":   error.Message,
			"details":   error.Details,
			"retryable": error.Retryable,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleStatusResult handles status updates from gRPC service
func (h *WebSocketHandler) handleStatusResult(sessionID string, result *speechv1.StatusResult) {
	message := &model.WebSocketMessage{
		Type: model.MessageTypeStatus,
		Data: map[string]interface{}{
			"status":  result.ProcessingStatus.String(),
			"message": result.Message,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}

// handleTextMessage processes text messages from WebSocket
func (h *WebSocketHandler) handleTextMessage(sessionID string, data []byte) {
	var message model.WebSocketMessage
	if err := json.Unmarshal(data, &message); err != nil {
		h.logger.Errorf("Failed to unmarshal text message: %v", err)
		h.sendErrorMessage(sessionID, "Invalid message format")
		return
	}

	switch message.Type {
	case model.MessageTypeControl:
		h.handleControlMessage(sessionID, &message)
	case model.MessageTypeText:
		h.handleUserTextMessage(sessionID, &message)
	default:
		h.sendErrorMessage(sessionID, "Unknown message type")
	}
}

// handleBinaryMessage processes binary messages from WebSocket
func (h *WebSocketHandler) handleBinaryMessage(sessionID string, data []byte) {
	h.logger.Infof("Processing binary message for session %s (%d bytes)", sessionID, len(data))

	// Forward audio data to gRPC service
	h.forwardAudioToGRPC(sessionID, data)
}

// handleControlMessage processes control messages
func (h *WebSocketHandler) handleControlMessage(sessionID string, message *model.WebSocketMessage) {
	controlData, ok := message.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(sessionID, "Invalid control message data")
		return
	}

	action, ok := controlData["action"].(string)
	if !ok {
		h.sendErrorMessage(sessionID, "Missing action in control message")
		return
	}

	// Forward control message to gRPC service
	h.forwardControlToGRPC(sessionID, action, controlData)
}

// handleUserTextMessage processes text input from user
func (h *WebSocketHandler) handleUserTextMessage(sessionID string, message *model.WebSocketMessage) {
	userText, ok := message.Data.(string)
	if !ok {
		h.sendErrorMessage(sessionID, "Invalid text message data")
		return
	}

	h.logger.Infof("Received text from user in session %s: %s", sessionID, userText)

	// Forward text to gRPC service as control message
	h.forwardControlToGRPC(sessionID, "text_input", map[string]interface{}{
		"text": userText,
	})
}

// forwardAudioToGRPC forwards audio data to the gRPC service
func (h *WebSocketHandler) forwardAudioToGRPC(sessionID string, audioData []byte) {
	// This would require maintaining a map of session to gRPC stream
	// For now, we'll log and implement later
	h.logger.Debugf("Forwarding %d bytes of audio data for session %s", len(audioData), sessionID)
}

// forwardControlToGRPC forwards control messages to the gRPC service
func (h *WebSocketHandler) forwardControlToGRPC(sessionID string, action string, params map[string]interface{}) {
	// This would require maintaining a map of session to gRPC stream
	// For now, we'll log and implement later
	h.logger.Debugf("Forwarding control action '%s' for session %s", action, sessionID)
}

// sendErrorMessage sends an error message to the client
func (h *WebSocketHandler) sendErrorMessage(sessionID string, errorMsg string) {
	message := &model.WebSocketMessage{
		Type:    model.MessageTypeError,
		Data:    errorMsg,
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}
