package handler

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ai-tutor-monorepo/services/speech-service/internal/service"
	speechv1 "github.com/ai-tutor-monorepo/services/speech-service/pkg/proto/speech"
)

// SpeechHandler implements the gRPC SpeechService
type SpeechHandler struct {
	speechv1.UnimplementedSpeechServiceServer
	
	audioService *service.AudioService
	asrService   *service.ASRService
	llmService   *service.LLMService
	ttsService   *service.TTSService
	logger       *logrus.Logger
	
	// Track active sessions
	sessions map[string]*VoiceSession
	mu       sync.RWMutex
}

// VoiceSession represents an active voice conversation session
type VoiceSession struct {
	ID            string
	IsRecording   bool
	StartTime     time.Time
	LastActivity  time.Time
	AudioBuffer   []byte
	Context       string
	Stream        speechv1.SpeechService_ProcessVoiceConversationServer
}

// NewSpeechHandler creates a new speech handler
func NewSpeechHandler(
	audioService *service.AudioService,
	asrService *service.ASRService,
	llmService *service.LLMService,
	ttsService *service.TTSService,
	logger *logrus.Logger,
) *SpeechHandler {
	return &SpeechHandler{
		audioService: audioService,
		asrService:   asrService,
		llmService:   llmService,
		ttsService:   ttsService,
		logger:       logger,
		sessions:     make(map[string]*VoiceSession),
	}
}

// ProcessVoiceConversation handles the bidirectional streaming gRPC call
func (h *SpeechHandler) ProcessVoiceConversation(stream speechv1.SpeechService_ProcessVoiceConversationServer) error {
	var sessionID string
	var session *VoiceSession

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			h.logger.Infof("Client closed stream for session: %s", sessionID)
			break
		}
		if err != nil {
			h.logger.Errorf("Error receiving from stream: %v", err)
			return status.Error(codes.Internal, "stream receive error")
		}

		// Initialize session if not exists
		if sessionID == "" {
			sessionID = request.SessionId
			session = h.getOrCreateSession(sessionID, stream)
			h.logger.Infof("Processing voice conversation for session: %s", sessionID)
		}

		// Update session activity
		session.LastActivity = time.Now()

		// Process request based on type
		switch req := request.RequestType.(type) {
		case *speechv1.VoiceRequest_AudioData:
			h.handleAudioData(session, req.AudioData)
		case *speechv1.VoiceRequest_Control:
			h.handleControlMessage(session, req.Control)
		default:
			h.logger.Warnf("Unknown request type for session %s", sessionID)
			h.sendError(session, speechv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "unknown request type")
		}
	}

	// Clean up session
	if sessionID != "" {
		h.removeSession(sessionID)
	}

	return nil
}

// HealthCheck implements health check
func (h *SpeechHandler) HealthCheck(ctx context.Context, req *speechv1.HealthCheckRequest) (*speechv1.HealthCheckResponse, error) {
	return &speechv1.HealthCheckResponse{
		Status: speechv1.HealthStatus_HEALTH_STATUS_SERVING,
		Details: map[string]string{
			"service":     "speech-service",
			"version":     "1.0.0",
			"active_sessions": fmt.Sprintf("%d", len(h.sessions)),
		},
	}, nil
}

// getOrCreateSession gets existing session or creates new one
func (h *SpeechHandler) getOrCreateSession(sessionID string, stream speechv1.SpeechService_ProcessVoiceConversationServer) *VoiceSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	if session, exists := h.sessions[sessionID]; exists {
		session.Stream = stream // Update stream reference
		return session
	}

	session := &VoiceSession{
		ID:           sessionID,
		IsRecording:  false,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		AudioBuffer:  make([]byte, 0),
		Context:      "",
		Stream:       stream,
	}

	h.sessions[sessionID] = session
	return session
}

// removeSession removes a session
func (h *SpeechHandler) removeSession(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sessions, sessionID)
	h.logger.Infof("Removed session: %s", sessionID)
}

// handleAudioData processes incoming audio data
func (h *SpeechHandler) handleAudioData(session *VoiceSession, audioData *speechv1.AudioData) {
	h.logger.Infof("Processing audio data for session %s (%d bytes)", session.ID, len(audioData.Data))

	if !session.IsRecording {
		h.logger.Warnf("Received audio data but session %s is not recording", session.ID)
		// Still process in case of timing issues
	}

	// Send processing status
	h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_PROCESSING, "Processing audio data")

	// Process complete audio file
	go h.processCompleteAudio(session, audioData.Data)
}

// handleControlMessage processes control messages
func (h *SpeechHandler) handleControlMessage(session *VoiceSession, control *speechv1.ControlMessage) {
	h.logger.Infof("Processing control message for session %s: %s", session.ID, control.Action.String())

	switch control.Action {
	case speechv1.ControlAction_CONTROL_ACTION_START_RECORDING:
		session.IsRecording = true
		session.AudioBuffer = make([]byte, 0)
		h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_STARTED, "Recording started")

	case speechv1.ControlAction_CONTROL_ACTION_STOP_RECORDING:
		session.IsRecording = false
		h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_COMPLETED, "Recording stopped")

	case speechv1.ControlAction_CONTROL_ACTION_END_SESSION:
		h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_COMPLETED, "Session ended")
		h.removeSession(session.ID)

	case speechv1.ControlAction_CONTROL_ACTION_PAUSE_SESSION:
		session.IsRecording = false
		h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_PROCESSING, "Session paused")

	case speechv1.ControlAction_CONTROL_ACTION_RESUME_SESSION:
		session.IsRecording = true
		h.sendStatus(session, speechv1.ProcessingStatus_PROCESSING_STATUS_PROCESSING, "Session resumed")

	default:
		h.sendError(session, speechv1.ErrorCode_ERROR_CODE_INVALID_REQUEST, "unknown control action")
	}
}

// processCompleteAudio processes complete audio data through the pipeline
func (h *SpeechHandler) processCompleteAudio(session *VoiceSession, audioData []byte) {
	if len(audioData) == 0 {
		h.logger.Warnf("Received empty audio data for session %s", session.ID)
		return
	}

	// Convert audio format for ASR
	convertedAudio, err := h.audioService.OptimizeAudioForASR(audioData)
	if err != nil {
		h.logger.Errorf("Failed to optimize audio for ASR in session %s: %v", session.ID, err)
		h.sendError(session, speechv1.ErrorCode_ERROR_CODE_AUDIO_PROCESSING_FAILED, "audio processing failed")
		return
	}

	// Process with ASR
	h.processAudioWithASR(session, convertedAudio)
}

// processAudioWithASR sends audio to ASR service and processes result
func (h *SpeechHandler) processAudioWithASR(session *VoiceSession, audioData []byte) {
	response, err := h.asrService.ProcessAudio(audioData)
	if err != nil {
		h.logger.Errorf("ASR processing failed for session %s: %v", session.ID, err)
		h.sendError(session, speechv1.ErrorCode_ERROR_CODE_ASR_FAILED, "speech recognition failed")
		return
	}

	if response.Text == "" {
		h.logger.Debugf("Empty ASR result for session %s", session.ID)
		return
	}

	h.logger.Infof("ASR result for session %s: %s (confidence: %.2f)", session.ID, response.Text, response.Confidence)

	// Send ASR result
	asrResult := &speechv1.ASRResult{
		Text:       response.Text,
		Confidence: float32(response.Confidence),
		IsFinal:    response.IsFinal,
		StartTimeMs: 0,
		EndTimeMs:   int64(len(audioData) * 1000 / 16000), // Rough estimate
	}

	h.sendASRResult(session, asrResult)

	// Process with LLM if final result
	if response.IsFinal && response.Text != "" {
		go h.processTextWithLLM(session, response.Text)
	}
}

// processTextWithLLM sends text to LLM and generates response
func (h *SpeechHandler) processTextWithLLM(session *VoiceSession, text string) {
	response, err := h.llmService.GenerateResponse(text, session.Context)
	if err != nil {
		h.logger.Errorf("LLM processing failed for session %s: %v", session.ID, err)
		h.sendError(session, speechv1.ErrorCode_ERROR_CODE_LLM_FAILED, "language model processing failed")
		return
	}

	h.logger.Infof("LLM response for session %s: %s", session.ID, response.Reply)

	// Update session context
	session.Context = response.Reply

	// Send LLM result
	llmResult := &speechv1.LLMResult{
		Text:       response.Reply,
		Context:    session.Context,
		Confidence: 0.95, // Default confidence
		ResultType: speechv1.LLMResultType_LLM_RESULT_TYPE_RESPONSE,
	}

	h.sendLLMResult(session, llmResult)

	// Generate TTS audio
	go h.processTextWithTTS(session, response.Reply)
}

// processTextWithTTS converts text to speech and sends audio
func (h *SpeechHandler) processTextWithTTS(session *VoiceSession, text string) {
	response, err := h.ttsService.SynthesizeSpeech(text)
	if err != nil {
		h.logger.Errorf("TTS processing failed for session %s: %v", session.ID, err)
		h.sendError(session, speechv1.ErrorCode_ERROR_CODE_TTS_FAILED, "text-to-speech failed")
		return
	}

	h.logger.Infof("Generated TTS audio for session %s (%d bytes)", session.ID, len(response.AudioData))

	// Send TTS result
	ttsResult := &speechv1.TTSResult{
		AudioData: response.AudioData,
		Format: &speechv1.AudioFormat{
			Codec:      response.Format,
			SampleRate: 22050, // Typical TTS sample rate
			Channels:   1,
			BitDepth:   16,
		},
		DurationMs: int64(len(response.AudioData) * 1000 / (22050 * 2)), // Rough estimate
		IsFinal:    true,
		ChunkIndex: 0,
	}

	h.sendTTSResult(session, ttsResult)
}

// Helper methods to send different response types

func (h *SpeechHandler) sendASRResult(session *VoiceSession, result *speechv1.ASRResult) {
	response := &speechv1.VoiceResponse{
		SessionId: session.ID,
		Timestamp: time.Now().UnixMilli(),
		Status: &speechv1.ResponseStatus{
			Success: true,
			Message: "ASR processing completed",
		},
		ResponseType: &speechv1.VoiceResponse_AsrResult{
			AsrResult: result,
		},
	}

	if err := session.Stream.Send(response); err != nil {
		h.logger.Errorf("Failed to send ASR result to session %s: %v", session.ID, err)
	}
}

func (h *SpeechHandler) sendLLMResult(session *VoiceSession, result *speechv1.LLMResult) {
	response := &speechv1.VoiceResponse{
		SessionId: session.ID,
		Timestamp: time.Now().UnixMilli(),
		Status: &speechv1.ResponseStatus{
			Success: true,
			Message: "LLM processing completed",
		},
		ResponseType: &speechv1.VoiceResponse_LlmResult{
			LlmResult: result,
		},
	}

	if err := session.Stream.Send(response); err != nil {
		h.logger.Errorf("Failed to send LLM result to session %s: %v", session.ID, err)
	}
}

func (h *SpeechHandler) sendTTSResult(session *VoiceSession, result *speechv1.TTSResult) {
	response := &speechv1.VoiceResponse{
		SessionId: session.ID,
		Timestamp: time.Now().UnixMilli(),
		Status: &speechv1.ResponseStatus{
			Success: true,
			Message: "TTS processing completed",
		},
		ResponseType: &speechv1.VoiceResponse_TtsResult{
			TtsResult: result,
		},
	}

	if err := session.Stream.Send(response); err != nil {
		h.logger.Errorf("Failed to send TTS result to session %s: %v", session.ID, err)
	}
}

func (h *SpeechHandler) sendError(session *VoiceSession, code speechv1.ErrorCode, message string) {
	response := &speechv1.VoiceResponse{
		SessionId: session.ID,
		Timestamp: time.Now().UnixMilli(),
		Status: &speechv1.ResponseStatus{
			Success: false,
			Message: message,
		},
		ResponseType: &speechv1.VoiceResponse_Error{
			Error: &speechv1.ErrorResult{
				Code:      code,
				Message:   message,
				Retryable: code != speechv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
			},
		},
	}

	if err := session.Stream.Send(response); err != nil {
		h.logger.Errorf("Failed to send error to session %s: %v", session.ID, err)
	}
}

func (h *SpeechHandler) sendStatus(session *VoiceSession, status speechv1.ProcessingStatus, message string) {
	response := &speechv1.VoiceResponse{
		SessionId: session.ID,
		Timestamp: time.Now().UnixMilli(),
		Status: &speechv1.ResponseStatus{
			Success: true,
			Message: message,
		},
		ResponseType: &speechv1.VoiceResponse_StatusResult{
			StatusResult: &speechv1.StatusResult{
				ProcessingStatus: status,
				Message:          message,
			},
		},
	}

	if err := session.Stream.Send(response); err != nil {
		h.logger.Errorf("Failed to send status to session %s: %v", session.ID, err)
	}
}