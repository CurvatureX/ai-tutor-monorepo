package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"voice-practice-backend/internal/model"
	"voice-practice-backend/internal/service"
	wsManager "voice-practice-backend/pkg/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// WebSocketHandler handles WebSocket connections for voice practice
type WebSocketHandler struct {
	manager      *wsManager.Manager
	audioService *service.AudioService
	asrService   *service.ASRService
	llmService   *service.LLMService
	ttsService   *service.TTSService
	logger       *logrus.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	manager *wsManager.Manager,
	audioService *service.AudioService,
	asrService *service.ASRService,
	llmService *service.LLMService,
	ttsService *service.TTSService,
	logger *logrus.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		manager:      manager,
		audioService: audioService,
		asrService:   asrService,
		llmService:   llmService,
		ttsService:   ttsService,
		logger:       logger,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	h.logger.Infof("🌐 WebSocket connection request for session: %s", sessionID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Errorf("❌ Failed to upgrade WebSocket connection: %v", err)
		return
	}

	h.manager.AddConnection(sessionID, conn)
	defer h.manager.RemoveConnection(sessionID)

	h.logger.Infof("✅ WebSocket connection established for session: %s", sessionID)

	// Send welcome message
	welcomeMsg := &model.WebSocketMessage{
		Type:    model.MessageTypeText,
		Data:    "Welcome to AI English Practice! Start speaking to begin your practice session.",
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, welcomeMsg)

	// Handle incoming messages
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			h.logger.Errorf("❌ WebSocket read error for session %s: %v", sessionID, err)
			break
		}

		h.logger.Debugf("📨 Received message for session %s: type=%d, size=%d", sessionID, messageType, len(data))

		switch messageType {
		case websocket.TextMessage:
			h.logger.Debugf("📝 Processing text message for session %s", sessionID)
			h.handleTextMessage(sessionID, data)
		case websocket.BinaryMessage:
			h.logger.Infof("🎵 Processing binary message for session %s (%d bytes)", sessionID, len(data))
			h.handleBinaryMessage(sessionID, data)
		default:
			h.logger.Warnf("⚠️ Unknown message type %d for session %s", messageType, sessionID)
		}
	}
}

// handleTextMessage processes text messages (control commands)
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

// handleBinaryMessage processes binary messages (complete audio data)
func (h *WebSocketHandler) handleBinaryMessage(sessionID string, data []byte) {
	h.logger.Infof("🔥 handleBinaryMessage called for session %s with %d bytes", sessionID, len(data))
	
	session, exists := h.manager.GetSession(sessionID)
	if !exists {
		h.logger.Errorf("❌ Session not found: %s", sessionID)
		return
	}

	h.logger.Infof("✅ Session found, IsRecording: %v", session.IsRecording)

	if !session.IsRecording {
		h.logger.Warnf("⚠️ Received audio data but session %s is not recording", sessionID)
		// 仍然处理，可能是时序问题
	}

	h.logger.Infof("🎵 Processing complete audio file for session %s: %d bytes", sessionID, len(data))

	// 直接处理完整的音频文件
	h.processCompleteAudio(sessionID, data)
}

// handleControlMessage processes control commands
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

	session, exists := h.manager.GetSession(sessionID)
	if !exists {
		h.sendErrorMessage(sessionID, "Session not found")
		return
	}

	switch action {
	case "start_recording":
		session.IsRecording = true
		session.AudioBuffer = make([]byte, 0)
		h.manager.UpdateSession(sessionID, session)
		h.logger.Infof("Started recording for session: %s", sessionID)

	case "stop_recording":
		session.IsRecording = false
		h.manager.UpdateSession(sessionID, session)
		h.logger.Infof("Stopped recording for session: %s", sessionID)
		// 注意：现在音频处理在handleBinaryMessage中完成

	case "end_session":
		h.logger.Infof("Ending session: %s", sessionID)
		h.manager.RemoveConnection(sessionID)

	default:
		h.sendErrorMessage(sessionID, "Unknown control action")
	}
}

// handleUserTextMessage processes text input from user
func (h *WebSocketHandler) handleUserTextMessage(sessionID string, message *model.WebSocketMessage) {
	userText, ok := message.Data.(string)
	if !ok {
		h.sendErrorMessage(sessionID, "Invalid text message data")
		return
	}

	h.logger.Infof("Received text from user in session %s: %s", sessionID, userText)

	// Process with LLM and generate response
	go h.processTextWithLLM(sessionID, userText)
}


// processCompleteAudio processes complete WebM audio file
func (h *WebSocketHandler) processCompleteAudio(sessionID string, webmData []byte) {
	if len(webmData) == 0 {
		h.logger.Warnf("Received empty audio data for session %s", sessionID)
		return
	}

	h.logger.Infof("🎬 Processing complete audio file for session %s (%d bytes)", sessionID, len(webmData))

	// 🐛 DEBUG: Save original WebM data BEFORE conversion
	webmFileName := fmt.Sprintf("debug/webm_COMPLETE_%s_%d.webm", sessionID, time.Now().UnixNano())
	if err := os.WriteFile(webmFileName, webmData, 0644); err != nil {
		h.logger.Warnf("Failed to save debug WebM file %s: %v", webmFileName, err)
	} else {
		h.logger.Infof("🎵 Saved debug WebM file: %s (%d bytes)", webmFileName, len(webmData))
	}

	// Validate WebM file format
	if len(webmData) >= 4 {
		// WebM files should start with specific magic bytes
		magic := webmData[:4]
		h.logger.Debugf("🔍 WebM file magic bytes: %v", magic)
		// WebM container uses EBML format, should start with 0x1A, 0x45, 0xDF, 0xA3
	}

	// Convert WebM audio to WAV format for ASR API
	convertedAudio, err := h.audioService.OptimizeAudioForASR(webmData)
	if err != nil {
		h.logger.Errorf("Failed to optimize audio for ASR in session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Audio processing failed")
		return
	}

	// 🐛 DEBUG: Save converted WAV file for debugging
	debugFileName := fmt.Sprintf("debug/audio_COMPLETE_%s_%d.wav", sessionID, time.Now().UnixNano())
	if err := os.WriteFile(debugFileName, convertedAudio, 0644); err != nil {
		h.logger.Warnf("Failed to save debug audio file %s: %v", debugFileName, err)
	} else {
		h.logger.Infof("🎵 Saved debug audio file: %s (%d bytes)", debugFileName, len(convertedAudio))
	}

	// Send to ASR service
	go h.processAudioWithASR(sessionID, convertedAudio)
}

// processAudioWithASR sends audio to ASR service and processes result
func (h *WebSocketHandler) processAudioWithASR(sessionID string, audioData []byte) {
	response, err := h.asrService.ProcessAudio(audioData)
	if err != nil {
		h.logger.Errorf("ASR processing failed for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Speech recognition failed")
		return
	}

	if response.Text == "" {
		h.logger.Debugf("Empty ASR result for session %s", sessionID)
		return
	}

	h.logger.Infof("ASR result for session %s: %s (confidence: %.2f)",
		sessionID, response.Text, response.Confidence)

	// Send ASR result to client
	asrMessage := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type":       "asr_result",
			"text":       response.Text,
			"confidence": response.Confidence,
			"is_final":   response.IsFinal,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, asrMessage)

	// Process with LLM if final result
	if response.IsFinal && response.Text != "" {
		go h.processTextWithLLM(sessionID, response.Text)
	}
}

// processTextWithLLM sends text to LLM and generates response
func (h *WebSocketHandler) processTextWithLLM(sessionID string, text string) {
	response, err := h.llmService.GenerateResponse(text, "")
	if err != nil {
		h.logger.Errorf("LLM processing failed for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Language model processing failed")
		return
	}

	h.logger.Infof("LLM response for session %s: %s", sessionID, response.Reply)

	// Send LLM response to client
	llmMessage := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type": "llm_response",
			"text": response.Reply,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, llmMessage)

	// Generate TTS audio
	go h.processTextWithTTS(sessionID, response.Reply)
}

// processTextWithTTS converts text to speech and sends audio
func (h *WebSocketHandler) processTextWithTTS(sessionID string, text string) {
	response, err := h.ttsService.SynthesizeSpeech(text)
	if err != nil {
		h.logger.Errorf("TTS processing failed for session %s: %v", sessionID, err)
		h.sendErrorMessage(sessionID, "Text-to-speech failed")
		return
	}

	h.logger.Infof("Generated TTS audio for session %s (%d bytes)", sessionID, len(response.AudioData))

	// Send TTS audio as binary message
	h.manager.SendBinaryMessage(sessionID, response.AudioData)

	// Also send notification that audio is ready
	ttsMessage := &model.WebSocketMessage{
		Type: model.MessageTypeText,
		Data: map[string]interface{}{
			"type":   "tts_ready",
			"format": response.Format,
		},
		Session: sessionID,
	}
	h.manager.SendMessage(sessionID, ttsMessage)
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
