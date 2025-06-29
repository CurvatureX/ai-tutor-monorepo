package service

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ai-tutor-monorepo/services/speech-service/internal/config"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/model"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// TTS Protocol constants (based on the Go demo)
const (
	TTS_DEFAULT_HEADER = 0x11100000 // version=1, header_size=1, msg_type=1, flags=0, serialization=1, compression=1, reserved=0
)

// TTSService handles text-to-speech conversion
type TTSService struct {
	config *config.TTSConfig
	logger *logrus.Logger
	wsURL  string
	dialer *websocket.Dialer
}

// NewTTSService creates a new TTS service
func NewTTSService(cfg *config.TTSConfig, logger *logrus.Logger) *TTSService {
	wsURL := cfg.BaseURL
	if wsURL == "" {
		wsURL = "wss://openspeech.bytedance.com/api/v1/tts/ws_binary"
	}

	return &TTSService{
		config: cfg,
		logger: logger,
		wsURL:  wsURL,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// SynthesizeSpeech converts text to speech
func (s *TTSService) SynthesizeSpeech(text string) (*model.TTSResponse, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("empty text input")
	}

	s.logger.Debugf("Synthesizing speech for text: %s", text)

	// Create WebSocket connection
	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer;%s", s.config.Token))

	u, _ := url.Parse(s.wsURL)
	conn, _, err := s.dialer.Dial(u.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TTS service: %v", err)
	}
	defer conn.Close()

	// Use streaming synthesis for better user experience
	audioData, err := s.streamSynthesize(conn, text)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %v", err)
	}

	return &model.TTSResponse{
		AudioData: audioData,
		Format:    "mp3",
	}, nil
}

// streamSynthesize performs streaming text-to-speech synthesis
func (s *TTSService) streamSynthesize(conn *websocket.Conn, text string) ([]byte, error) {
	// Setup input parameters
	input := s.setupInput(text, s.config.Voice, "submit") // "submit" for streaming

	// Compress the JSON input
	compressedInput := s.compressData(input)

	// Create the request message
	message := s.createTTSMessage(compressedInput)

	// Send the request
	if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		return nil, fmt.Errorf("failed to send TTS request: %v", err)
	}

	// Collect audio data from streaming responses
	var audioData []byte
	for {
		_, responseData, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to read TTS response: %v", err)
		}

		response, err := s.parseResponse(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TTS response: %v", err)
		}

		// Append audio data
		audioData = append(audioData, response.Audio...)

		// Check if this is the last chunk
		if response.IsLast {
			break
		}
	}

	s.logger.Debugf("TTS synthesis complete: %d bytes of audio data", len(audioData))
	return audioData, nil
}

// setupInput creates the JSON request for TTS
func (s *TTSService) setupInput(text, voiceType, operation string) []byte {
	reqID := generateRequestID()

	params := map[string]interface{}{
		"app": map[string]interface{}{
			"appid":   s.config.AppID,   // AppID from config
			"token":   s.config.Token,   // Fixed value as in demo
			"cluster": s.config.Cluster, // Cluster from config
		},
		"user": map[string]interface{}{
			"uid": "voice-practice-user",
		},
		"audio": map[string]interface{}{
			"voice_type":   voiceType,
			"encoding":     "mp3",
			"speed_ratio":  1.0,
			"volume_ratio": 1.0,
			"pitch_ratio":  1.0,
		},
		"request": map[string]interface{}{
			"reqid":     reqID,
			"text":      text,
			"text_type": "plain",
			"operation": operation,
		},
	}

	jsonData, _ := json.Marshal(params)
	s.logger.Info("setupInput: %v", string(jsonData))
	return jsonData
}

// createTTSMessage creates the binary message for TTS WebSocket
func (s *TTSService) createTTSMessage(compressedInput []byte) []byte {
	// Default header for TTS (based on the Go demo)
	defaultHeader := []byte{0x11, 0x10, 0x11, 0x00}

	// Payload size
	payloadSize := len(compressedInput)
	payloadSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeBytes, uint32(payloadSize))

	// Combine header + payload size + payload
	message := make([]byte, len(defaultHeader))
	copy(message, defaultHeader)
	message = append(message, payloadSizeBytes...)
	message = append(message, compressedInput...)

	return message
}

// compressData compresses data using gzip
func (s *TTSService) compressData(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// decompressData decompresses gzip data
func (s *TTSService) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// TTSResponse represents the parsed TTS response
type TTSResponse struct {
	Audio  []byte
	IsLast bool
}

// parseResponse parses the TTS WebSocket response
func (s *TTSService) parseResponse(res []byte) (*TTSResponse, error) {
	if len(res) < 4 {
		return nil, fmt.Errorf("response too short")
	}

	_ = res[0] >> 4 // protocolVersion (unused)
	headerSize := res[0] & 0x0f
	messageType := res[1] >> 4
	messageTypeSpecificFlags := res[1] & 0x0f
	serializationMethod := res[2] >> 4
	messageCompression := res[2] & 0x0f

	headerExtensionsEnd := headerSize * 4
	if len(res) < int(headerExtensionsEnd) {
		return nil, fmt.Errorf("invalid header size")
	}

	payload := res[headerExtensionsEnd:]
	response := &TTSResponse{
		Audio:  []byte{},
		IsLast: false,
	}

	s.logger.Debugf("TTS Response - Type: %x, Flags: %x, Serialization: %x, Compression: %x",
		messageType, messageTypeSpecificFlags, serializationMethod, messageCompression)

	switch messageType {
	case 0xb: // audio-only server response
		if messageTypeSpecificFlags == 0 {
			// No sequence number (ACK)
			s.logger.Debug("Received TTS ACK")
		} else {
			// Has sequence number and audio data
			if len(payload) < 8 {
				return nil, fmt.Errorf("payload too short for audio response")
			}

			sequenceNumber := int32(binary.BigEndian.Uint32(payload[0:4]))
			payloadSize := int32(binary.BigEndian.Uint32(payload[4:8]))
			audioData := payload[8:]

			response.Audio = audioData

			s.logger.Debugf("TTS Audio chunk - Sequence: %d, Size: %d", sequenceNumber, payloadSize)

			// Check if this is the last chunk (negative sequence number)
			if sequenceNumber < 0 {
				response.IsLast = true
			}
		}

	case 0xf: // error message
		if len(payload) < 8 {
			return nil, fmt.Errorf("payload too short for error response")
		}

		code := int32(binary.BigEndian.Uint32(payload[0:4]))
		errorMsg := payload[8:]

		if messageCompression == 1 {
			decompressed, err := s.decompressData(errorMsg)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress error message: %v", err)
			}
			errorMsg = decompressed
		}

		return nil, fmt.Errorf("TTS error %d: %s", code, string(errorMsg))

	case 0xc: // frontend server response
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for frontend response")
		}

		_ = int32(binary.BigEndian.Uint32(payload[0:4])) // msgSize
		frontendMsg := payload[4:]

		if messageCompression == 1 {
			decompressed, err := s.decompressData(frontendMsg)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress frontend message: %v", err)
			}
			frontendMsg = decompressed
		}

		s.logger.Debugf("TTS Frontend message: %s", string(frontendMsg))

	default:
		return nil, fmt.Errorf("unknown TTS message type: %d", messageType)
	}

	return response, nil
}

// SynthesizeSpeechNonStreaming performs non-streaming synthesis (for comparison)
func (s *TTSService) SynthesizeSpeechNonStreaming(text string) (*model.TTSResponse, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("empty text input")
	}

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer;%s", s.config.Token))

	u, _ := url.Parse(s.wsURL)
	conn, _, err := s.dialer.Dial(u.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TTS service: %v", err)
	}
	defer conn.Close()

	// Setup input for non-streaming (query operation)
	input := s.setupInput(text, s.config.Voice, "query")
	compressedInput := s.compressData(input)
	message := s.createTTSMessage(compressedInput)

	// Send request
	if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		return nil, fmt.Errorf("failed to send TTS request: %v", err)
	}

	// Read single response
	_, responseData, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read TTS response: %v", err)
	}

	response, err := s.parseResponse(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TTS response: %v", err)
	}

	return &model.TTSResponse{
		AudioData: response.Audio,
		Format:    "mp3",
	}, nil
}
