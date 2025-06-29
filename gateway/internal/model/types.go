package model

import (
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Session   string      `json:"session"`
	Timestamp int64       `json:"timestamp"`
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeAudio   MessageType = "audio"
	MessageTypeControl MessageType = "control"
	MessageTypeError   MessageType = "error"
	MessageTypeStatus  MessageType = "status"
)

// WebSocketSession represents a WebSocket session
type WebSocketSession struct {
	ID           string
	Connection   *websocket.Conn
	IsRecording  bool
	StartTime    time.Time
	LastActivity time.Time
	Metadata     map[string]interface{}
}

// ControlMessage represents a control message
type ControlMessage struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// AudioMessage represents an audio message
type AudioMessage struct {
	Format   string `json:"format"`
	Data     []byte `json:"data"`
	Metadata struct {
		Duration  int64 `json:"duration_ms"`
		ChunkIndex int32 `json:"chunk_index"`
		IsFinal   bool  `json:"is_final"`
	} `json:"metadata"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StatusMessage represents a status message
type StatusMessage struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// ASRResult represents speech recognition result
type ASRResult struct {
	Text       string  `json:"text"`
	Confidence float32 `json:"confidence"`
	IsFinal    bool    `json:"is_final"`
	StartTime  int64   `json:"start_time_ms"`
	EndTime    int64   `json:"end_time_ms"`
}

// LLMResult represents language model result
type LLMResult struct {
	Text       string                 `json:"text"`
	Context    string                 `json:"context"`
	Confidence float32                `json:"confidence"`
	Type       string                 `json:"type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TTSResult represents text-to-speech result
type TTSResult struct {
	AudioData  []byte `json:"audio_data"`
	Format     string `json:"format"`
	Duration   int64  `json:"duration_ms"`
	IsFinal    bool   `json:"is_final"`
	ChunkIndex int32  `json:"chunk_index"`
}