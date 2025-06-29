package model

import (
	"time"
	"github.com/gorilla/websocket"
)

// AudioSpec defines audio format specifications
type AudioSpec struct {
	Format     string `json:"format"`      // "pcm" or "wav"
	SampleRate int    `json:"sampleRate"`  // 16000 Hz
	Channels   int    `json:"channels"`    // 1 (mono)
	BitDepth   int    `json:"bitDepth"`    // 16 bit
	Encoding   string `json:"encoding"`    // "pcm_s16le"
}

// VoiceSession represents a voice practice session
type VoiceSession struct {
	ID          string    `json:"id"`
	AudioBuffer []byte    `json:"-"`
	IsRecording bool      `json:"isRecording"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	Connections map[string]*websocket.Conn `json:"-"`
	Sessions    map[string]*VoiceSession   `json:"-"`
}

// Message types for WebSocket communication
const (
	MessageTypeAudio    = "audio"
	MessageTypeText     = "text"
	MessageTypeControl  = "control"
	MessageTypeError    = "error"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Session string      `json:"session"`
}

// ControlMessage for session control
type ControlMessage struct {
	Action string `json:"action"` // "start_recording", "stop_recording", "end_session"
}

// ASRResponse from speech recognition service
type ASRResponse struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	IsFinal    bool    `json:"is_final"`
}

// LLMRequest to language model service
type LLMRequest struct {
	Message string `json:"message"`
	Context string `json:"context"`
}

// LLMResponse from language model service
type LLMResponse struct {
	Reply   string `json:"reply"`
	Context string `json:"context"`
}

// TTSRequest to text-to-speech service
type TTSRequest struct {
	Text     string `json:"text"`
	Voice    string `json:"voice"`
	Language string `json:"language"`
}

// TTSResponse from text-to-speech service
type TTSResponse struct {
	AudioData []byte `json:"audio_data"`
	Format    string `json:"format"`
}