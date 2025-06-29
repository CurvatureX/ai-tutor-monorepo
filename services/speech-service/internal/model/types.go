package model

import (
	"time"

	"github.com/gorilla/websocket"
)

// AudioSpec defines audio format specifications
type AudioSpec struct {
	Format     string `json:"format"`     // "pcm" or "wav"
	SampleRate int    `json:"sampleRate"` // 16000 Hz
	Channels   int    `json:"channels"`   // 1 (mono)
	BitDepth   int    `json:"bitDepth"`   // 16 bit
	Encoding   string `json:"encoding"`   // "pcm_s16le"
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
	MessageTypeAudio   = "audio"
	MessageTypeText    = "text"
	MessageTypeControl = "control"
	MessageTypeError   = "error"
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

// ISERequest to speech evaluation service
type ISERequest struct {
	AudioData []byte `json:"audio_data"`
	Text      string `json:"text"`     // Reference text for evaluation
	Language  string `json:"language"` // "zh_cn" or "en_us"
	Category  string `json:"category"` // "read_syllable", "read_word", "read_sentence", etc.
}

// ISEResponse from speech evaluation service
type ISEResponse struct {
	OverallScore      float64         `json:"overall_score"`      // 总分 0-100
	AccuracyScore     float64         `json:"accuracy_score"`     // 准确度分数
	FluencyScore      float64         `json:"fluency_score"`      // 流利度分数
	CompletenessScore float64         `json:"completeness_score"` // 完整度分数
	WordScores        []WordScore     `json:"word_scores"`        // 单词级别评分
	PhoneScores       []PhoneScore    `json:"phone_scores"`       // 音素级别评分
	SentenceScores    []SentenceScore `json:"sentence_scores"`    // 句子级别评分
	IsFinal           bool            `json:"is_final"`
}

// WordScore represents word-level scoring
type WordScore struct {
	Word       string  `json:"word"`
	Score      float64 `json:"score"`
	StartTime  int64   `json:"start_time"` // milliseconds
	EndTime    int64   `json:"end_time"`   // milliseconds
	IsCorrect  bool    `json:"is_correct"`
	Confidence float64 `json:"confidence"`
}

// PhoneScore represents phoneme-level scoring
type PhoneScore struct {
	Phone     string  `json:"phone"`
	Score     float64 `json:"score"`
	StartTime int64   `json:"start_time"`
	EndTime   int64   `json:"end_time"`
	IsCorrect bool    `json:"is_correct"`
}

// SentenceScore represents sentence-level scoring
type SentenceScore struct {
	Sentence      string  `json:"sentence"`
	Score         float64 `json:"score"`
	AccuracyScore float64 `json:"accuracy_score"`
	FluencyScore  float64 `json:"fluency_score"`
	TotalWords    int     `json:"total_words"`
	CorrectWords  int     `json:"correct_words"`
}
