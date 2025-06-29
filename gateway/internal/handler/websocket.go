package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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

// SessionStream holds gRPC stream for a session
type SessionStream struct {
	Stream     speechv1.SpeechService_ProcessVoiceConversationClient
	Context    context.Context
	CancelFunc context.CancelFunc
	Mutex      sync.Mutex
}

// EnhancedWebSocketHandler handles WebSocket connections and bridges to gRPC with stream management
type EnhancedWebSocketHandler struct {
	manager      *manager.WebSocketManager
	speechClient speechv1.SpeechServiceClient
	logger       *logrus.Logger
	streams      map[string]*SessionStream
	streamsMutex sync.RWMutex
}

// NewEnhancedWebSocketHandler creates a new enhanced WebSocket handler
func NewEnhancedWebSocketHandler(
	manager *manager.WebSocketManager,
	speechClient speechv1.SpeechServiceClient,
	logger *logrus.Logger,
) *EnhancedWebSocketHandler {
	return &EnhancedWebSocketHandler{
		manager:      manager,
		speechClient: speechClient,
		logger:       logger,
		streams:      make(map[string]*SessionStream),
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *EnhancedWebSocketHandler) HandleWebSocket(c *gin.Context) {
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
	defer func() {
		h.manager.RemoveConnection(sessionID)
		h.closeGRPCStream(sessionID)
	}()

	h.logger.Infof("WebSocket connection established for session: %s", sessionID)

	// Send welcome message
	welcomeMsg := &model.WebSocketMessage{
		Type:    model.MessageTypeText,
		Data:    "Welcome to AI English Practice! Start speaking to begin your practice session.",
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, welcomeMsg)

	// Initialize gRPC stream for this session
	if err := h.initGRPCStream(sessionID); err != nil {
		h.logger.Errorf("Failed to initialize gRPC stream for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Failed to initialize voice processing")
		return
	}

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

// initGRPCStream initializes a gRPC stream for a session
func (h *EnhancedWebSocketHandler) initGRPCStream(sessionID string) error {
	h.streamsMutex.Lock()
	defer h.streamsMutex.Unlock()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create gRPC stream
	stream, err := h.speechClient.ProcessVoiceConversation(ctx)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create gRPC stream: %w", err)
	}

	// Store stream
	h.streams[sessionID] = &SessionStream{
		Stream:     stream,
		Context:    ctx,
		CancelFunc: cancel,
	}

	// Start goroutine to handle responses from gRPC service
	go h.handleGRPCResponses(sessionID, stream)

	h.logger.Infof("Initialized gRPC stream for session: %s", sessionID)
	return nil
}

// closeGRPCStream closes the gRPC stream for a session
func (h *EnhancedWebSocketHandler) closeGRPCStream(sessionID string) {
	h.streamsMutex.Lock()
	defer h.streamsMutex.Unlock()

	if sessionStream, exists := h.streams[sessionID]; exists {
		sessionStream.CancelFunc()
		if err := sessionStream.Stream.CloseSend(); err != nil {
			h.logger.Errorf("Failed to close gRPC stream for session %s: %v", sessionID, err)
		}
		delete(h.streams, sessionID)
		h.logger.Infof("Closed gRPC stream for session: %s", sessionID)
	}
}

// getGRPCStream safely gets the gRPC stream for a session
func (h *EnhancedWebSocketHandler) getGRPCStream(sessionID string) (*SessionStream, bool) {
	h.streamsMutex.RLock()
	defer h.streamsMutex.RUnlock()
	stream, exists := h.streams[sessionID]
	return stream, exists
}

// handleGRPCResponses handles responses from the gRPC service
func (h *EnhancedWebSocketHandler) handleGRPCResponses(sessionID string, stream speechv1.SpeechService_ProcessVoiceConversationClient) {
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
func (h *EnhancedWebSocketHandler) processGRPCResponse(sessionID string, response *speechv1.VoiceResponse) {
	switch result := response.ResponseType.(type) {
	case *speechv1.VoiceResponse_AsrResult:
		h.handleASRResult(sessionID, result.AsrResult)
	case *speechv1.VoiceResponse_LlmResult:
		h.handleLLMResult(sessionID, result.LlmResult)
	case *speechv1.VoiceResponse_TtsResult:
		h.handleTTSResult(sessionID, result.TtsResult)
	case *speechv1.VoiceResponse_Error:
		h.handleGRPCError(sessionID, result.Error)
	case *speechv1.VoiceResponse_StatusResult:
		h.handleStatusResult(sessionID, result.StatusResult)
	default:
		h.logger.Warnf("Unknown gRPC response type for session %s", sessionID)
	}
}

// handleASRResult handles ASR results from gRPC service
func (h *EnhancedWebSocketHandler) handleASRResult(sessionID string, result *speechv1.ASRResult) {
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
	h.logger.Infof("ASR result for session %s: %s (confidence: %.2f)", sessionID, result.Text, result.Confidence)
}

// handleLLMResult handles LLM results from gRPC service
func (h *EnhancedWebSocketHandler) handleLLMResult(sessionID string, result *speechv1.LLMResult) {
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
	h.logger.Infof("LLM response for session %s: %s", sessionID, result.Text)
}

// handleTTSResult handles TTS results from gRPC service
func (h *EnhancedWebSocketHandler) handleTTSResult(sessionID string, result *speechv1.TTSResult) {
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
	h.logger.Infof("TTS audio ready for session %s (%d bytes)", sessionID, len(result.AudioData))
}

// handleGRPCError handles errors from gRPC service
func (h *EnhancedWebSocketHandler) handleGRPCError(sessionID string, error *speechv1.ErrorResult) {
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
	h.logger.Errorf("gRPC error for session %s: %s", sessionID, error.Message)
}

// handleStatusResult handles status updates from gRPC service
func (h *EnhancedWebSocketHandler) handleStatusResult(sessionID string, result *speechv1.StatusResult) {
	message := &model.WebSocketMessage{
		Type: model.MessageTypeStatus,
		Data: map[string]interface{}{
			"status":  result.ProcessingStatus.String(),
			"message": result.Message,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
	h.logger.Debugf("Status update for session %s: %s", sessionID, result.Message)
}

// handleTextMessage processes text messages from WebSocket
func (h *EnhancedWebSocketHandler) handleTextMessage(sessionID string, data []byte) {
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
func (h *EnhancedWebSocketHandler) handleBinaryMessage(sessionID string, data []byte) {
	h.logger.Infof("Processing binary message for session %s (%d bytes)", sessionID, len(data))

	// Forward audio data to gRPC service
	h.forwardAudioToGRPC(sessionID, data)
}

// handleControlMessage processes control messages
func (h *EnhancedWebSocketHandler) handleControlMessage(sessionID string, message *model.WebSocketMessage) {
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
func (h *EnhancedWebSocketHandler) handleUserTextMessage(sessionID string, message *model.WebSocketMessage) {
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
func (h *EnhancedWebSocketHandler) forwardAudioToGRPC(sessionID string, audioData []byte) {
	sessionStream, exists := h.getGRPCStream(sessionID)
	if !exists {
		h.logger.Errorf("No gRPC stream found for session %s", sessionID)
		h.sendErrorMessage(sessionID, "Voice processing not available")
		return
	}

	// Lock the stream for sending
	sessionStream.Mutex.Lock()
	defer sessionStream.Mutex.Unlock()

	// Create gRPC request with audio data
	request := &speechv1.VoiceRequest{
		SessionId: sessionID,
		Timestamp: time.Now().UnixMilli(),
		RequestType: &speechv1.VoiceRequest_AudioData{
			AudioData: &speechv1.AudioData{
				Data: audioData,
				Format: &speechv1.AudioFormat{
					Codec:      "webm",
					SampleRate: 48000, // WebM default
					Channels:   1,
					BitDepth:   16,
				},
				Metadata: &speechv1.AudioMetadata{
					IsFinal: true,
				},
			},
		},
	}

	// Send to gRPC stream
	if err := sessionStream.Stream.Send(request); err != nil {
		h.logger.Errorf("Failed to send audio data to gRPC for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Failed to process audio")
		return
	}

	h.logger.Debugf("Successfully forwarded %d bytes of audio data for session %s", len(audioData), sessionID)
}

// forwardControlToGRPC forwards control messages to the gRPC service
func (h *EnhancedWebSocketHandler) forwardControlToGRPC(sessionID string, action string, params map[string]interface{}) {
	sessionStream, exists := h.getGRPCStream(sessionID)
	if !exists {
		h.logger.Errorf("No gRPC stream found for session %s", sessionID)
		h.sendErrorMessage(sessionID, "Voice processing not available")
		return
	}

	// Lock the stream for sending
	sessionStream.Mutex.Lock()
	defer sessionStream.Mutex.Unlock()

	// Convert action to gRPC control action
	var controlAction speechv1.ControlAction
	switch action {
	case "start_recording":
		controlAction = speechv1.ControlAction_CONTROL_ACTION_START_RECORDING
	case "stop_recording":
		controlAction = speechv1.ControlAction_CONTROL_ACTION_STOP_RECORDING
	case "end_session":
		controlAction = speechv1.ControlAction_CONTROL_ACTION_END_SESSION
	case "pause_session":
		controlAction = speechv1.ControlAction_CONTROL_ACTION_PAUSE_SESSION
	case "resume_session":
		controlAction = speechv1.ControlAction_CONTROL_ACTION_RESUME_SESSION
	default:
		h.logger.Warnf("Unknown control action: %s", action)
		controlAction = speechv1.ControlAction_CONTROL_ACTION_UNSPECIFIED
	}

	// Convert params to string map
	stringParams := make(map[string]string)
	for key, value := range params {
		if str, ok := value.(string); ok {
			stringParams[key] = str
		} else {
			stringParams[key] = fmt.Sprintf("%v", value)
		}
	}

	// Create gRPC request with control data
	request := &speechv1.VoiceRequest{
		SessionId: sessionID,
		Timestamp: time.Now().UnixMilli(),
		RequestType: &speechv1.VoiceRequest_Control{
			Control: &speechv1.ControlMessage{
				Action: controlAction,
				Params: stringParams,
			},
		},
	}

	// Send to gRPC stream
	if err := sessionStream.Stream.Send(request); err != nil {
		h.logger.Errorf("Failed to send control message to gRPC for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Failed to process control message")
		return
	}

	h.logger.Debugf("Successfully forwarded control action '%s' for session %s", action, sessionID)
}

// sendErrorMessage sends an error message to the client
func (h *EnhancedWebSocketHandler) sendErrorMessage(sessionID string, errorMsg string) {
	message := &model.WebSocketMessage{
		Type:    model.MessageTypeError,
		Data:    errorMsg,
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, message)
}