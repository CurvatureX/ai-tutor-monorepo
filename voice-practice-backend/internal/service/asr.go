package service

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"voice-practice-backend/internal/config"
	"voice-practice-backend/internal/model"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ASR Protocol constants (based on the Python demo)
const (
	PROTOCOL_VERSION      = 0b0001
	DEFAULT_HEADER_SIZE   = 0b0001
	FULL_CLIENT_REQUEST   = 0b0001
	AUDIO_ONLY_REQUEST    = 0b0010
	FULL_SERVER_RESPONSE  = 0b1001
	SERVER_ACK            = 0b1011
	SERVER_ERROR_RESPONSE = 0b1111
	NO_SEQUENCE           = 0b0000
	POS_SEQUENCE          = 0b0001
	NEG_SEQUENCE          = 0b0010
	NEG_WITH_SEQUENCE     = 0b0011
	NO_SERIALIZATION      = 0b0000
	JSON_SERIALIZATION    = 0b0001
	NO_COMPRESSION        = 0b0000
	GZIP_COMPRESSION      = 0b0001
)

// ASRService handles speech recognition
type ASRService struct {
	config *config.ASRConfig
	logger *logrus.Logger
	wsURL  string
	dialer *websocket.Dialer
}

// NewASRService creates a new ASR service
func NewASRService(cfg *config.ASRConfig, logger *logrus.Logger) *ASRService {
	wsURL := cfg.BaseURL
	if wsURL == "" {
		wsURL = "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel"
	}

	return &ASRService{
		config: cfg,
		logger: logger,
		wsURL:  wsURL,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// ProcessAudio processes audio data and returns ASR result
func (s *ASRService) ProcessAudio(audioData []byte) (*model.ASRResponse, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}

	s.logger.Infof("🔊 ASR Processing audio data: %d bytes", len(audioData))

	// Check if it's a valid WAV file
	if len(audioData) >= 12 {
		if string(audioData[0:4]) == "RIFF" && string(audioData[8:12]) == "WAVE" {
			s.logger.Infof("✅ Valid WAV file detected (RIFF+WAVE headers)")
		} else {
			s.logger.Warnf("⚠️ Audio data doesn't appear to be a valid WAV file")
			s.logger.Debugf("First 12 bytes: %v", audioData[0:12])
		}
	}

	// Create WebSocket connection
	headers := http.Header{}
	headers.Set("X-Api-Resource-Id", "volc.bigasr.sauc.duration")
	headers.Set("X-Api-Access-Key", s.config.AccessKey)
	headers.Set("X-Api-App-Key", s.config.AppKey)
	headers.Set("X-Api-Request-Id", generateRequestID())

	s.logger.Infof("🌐 Connecting to ASR service: %s", s.wsURL)
	s.logger.Debugf("🔑 ASR Headers: Access-Key=%s, App-Key=%s",
		s.config.AccessKey[:8]+"...", s.config.AppKey)

	conn, resp, err := s.dialer.Dial(s.wsURL, headers)
	if err != nil {
		s.logger.Errorf("❌ Failed to connect to ASR service: %v", err)
		if resp != nil {
			s.logger.Errorf("❌ Response status: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to connect to ASR service: %v", err)
	}
	defer conn.Close()

	s.logger.Infof("✅ Connected to ASR service successfully")

	// Send initial request
	if err := s.sendInitialRequest(conn); err != nil {
		return nil, fmt.Errorf("failed to send initial request: %v", err)
	}

	// Read initial response
	if _, err := s.readResponse(conn); err != nil {
		return nil, fmt.Errorf("failed to read initial response: %v", err)
	}

	// Send audio data in chunks
	result, err := s.sendAudioChunks(conn, audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio chunks: %v", err)
	}

	return result, nil
}

// sendInitialRequest sends the initial configuration request
func (s *ASRService) sendInitialRequest(conn *websocket.Conn) error {
	// Construct request parameters
	req := map[string]interface{}{
		"user": map[string]interface{}{
			"uid": "voice-practice-user",
		},
		"audio": map[string]interface{}{
			"format":      "wav",
			"sample_rate": 16000,
			"bits":        16,
			"channel":     1,
			"codec":       "raw",
		},
		"request": map[string]interface{}{
			"model_name":  "bigmodel",
			"enable_punc": true,
		},
	}

	// Serialize to JSON and compress
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	compressedData := s.compressData(jsonData)

	// Generate header and message
	header := s.generateHeader(FULL_CLIENT_REQUEST, POS_SEQUENCE, JSON_SERIALIZATION, GZIP_COMPRESSION)
	sequenceBytes := s.generateSequence(1)
	payloadSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeBytes, uint32(len(compressedData)))

	// Combine all parts
	message := append(header, sequenceBytes...)
	message = append(message, payloadSizeBytes...)
	message = append(message, compressedData...)

	return conn.WriteMessage(websocket.BinaryMessage, message)
}

// sendAudioChunks sends audio data in chunks and processes responses
func (s *ASRService) sendAudioChunks(conn *websocket.Conn, audioData []byte) (*model.ASRResponse, error) {
	chunkSize := 3200 // ~100ms of 16kHz 16-bit mono audio
	seq := 2
	var finalResult *model.ASRResponse

	chunks := s.splitAudioData(audioData, chunkSize)
	totalChunks := len(chunks)

	for i, chunk := range chunks {
		isLast := i == totalChunks-1

		if isLast {
			seq = -seq // Negative sequence for last chunk
		}

		// Compress audio chunk
		compressedChunk := s.compressData(chunk)

		// Generate audio-only request
		messageType := AUDIO_ONLY_REQUEST
		flags := POS_SEQUENCE
		if isLast {
			flags = NEG_WITH_SEQUENCE
		}

		header := s.generateHeader(byte(messageType), byte(flags), NO_SERIALIZATION, GZIP_COMPRESSION)
		sequenceBytes := s.generateSequence(int32(seq))
		payloadSizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(payloadSizeBytes, uint32(len(compressedChunk)))

		// Combine message parts
		message := append(header, sequenceBytes...)
		message = append(message, payloadSizeBytes...)
		message = append(message, compressedChunk...)

		// Send chunk
		if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			return nil, fmt.Errorf("failed to send audio chunk %d: %v", i, err)
		}

		// Read response
		response, err := s.readResponse(conn)
		if err != nil {
			return nil, fmt.Errorf("failed to read response for chunk %d: %v", i, err)
		}

		// Parse ASR result if available
		if response != nil && response.PayloadMsg != nil {
			s.logger.Debugf("📝 ASR Response for chunk %d: %v", i, response.PayloadMsg)
			if asrResult := s.parseASRResult(response.PayloadMsg); asrResult != nil {
				s.logger.Infof("🎯 ASR Result: '%s' (confidence: %.2f)", asrResult.Text, asrResult.Confidence)
				finalResult = asrResult
				if isLast {
					finalResult.IsFinal = true
				}
			} else {
				s.logger.Warnf("⚠️ Failed to parse ASR result from payload")
			}
		} else {
			s.logger.Debugf("📭 No payload in ASR response for chunk %d", i)
		}

		if !isLast {
			seq++
		}
	}

	if finalResult == nil {
		finalResult = &model.ASRResponse{
			Text:       "",
			Confidence: 0.0,
			IsFinal:    true,
		}
	}

	return finalResult, nil
}

// splitAudioData splits audio data into chunks
func (s *ASRService) splitAudioData(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// generateHeader generates protocol header
func (s *ASRService) generateHeader(messageType, flags, serialization, compression byte) []byte {
	header := make([]byte, 4)
	header[0] = (PROTOCOL_VERSION << 4) | DEFAULT_HEADER_SIZE
	header[1] = (messageType << 4) | flags
	header[2] = (serialization << 4) | compression
	header[3] = 0x00 // reserved
	return header
}

// generateSequence generates sequence number bytes
func (s *ASRService) generateSequence(seq int32) []byte {
	seqBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBytes, uint32(seq))
	return seqBytes
}

// compressData compresses data using gzip
func (s *ASRService) compressData(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// decompressData decompresses gzip data
func (s *ASRService) decompressData(data []byte) ([]byte, error) {
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

// ResponseData represents the parsed WebSocket response
type ResponseData struct {
	IsLastPackage   bool
	PayloadSequence *int32
	PayloadMsg      interface{}
	PayloadSize     int32
	Code            *uint32
}

// readResponse reads and parses WebSocket response
func (s *ASRService) readResponse(conn *websocket.Conn) (*ResponseData, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return s.parseResponse(message)
}

// parseResponse parses the binary response according to the protocol
func (s *ASRService) parseResponse(res []byte) (*ResponseData, error) {
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
	result := &ResponseData{
		IsLastPackage: false,
	}

	// Check if this is the last package
	if messageTypeSpecificFlags&0x02 != 0 {
		result.IsLastPackage = true
	}

	// Handle sequence number
	if messageTypeSpecificFlags&0x01 != 0 {
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for sequence")
		}
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		result.PayloadSequence = &seq
		payload = payload[4:]
	}

	var payloadMsg []byte
	var payloadSize int32

	switch messageType {
	case FULL_SERVER_RESPONSE:
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for full response")
		}
		payloadSize = int32(binary.BigEndian.Uint32(payload[:4]))
		payloadMsg = payload[4:]

	case SERVER_ACK:
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for ack")
		}
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		result.PayloadSequence = &seq
		if len(payload) >= 8 {
			payloadSize = int32(binary.BigEndian.Uint32(payload[4:8]))
			payloadMsg = payload[8:]
		}

	case SERVER_ERROR_RESPONSE:
		if len(payload) < 8 {
			return nil, fmt.Errorf("payload too short for error response")
		}
		code := binary.BigEndian.Uint32(payload[:4])
		result.Code = &code
		payloadSize = int32(binary.BigEndian.Uint32(payload[4:8]))
		payloadMsg = payload[8:]
	}

	if payloadMsg != nil {
		result.PayloadSize = payloadSize

		// Decompress if needed
		if messageCompression == GZIP_COMPRESSION {
			decompressed, err := s.decompressData(payloadMsg)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress payload: %v", err)
			}
			payloadMsg = decompressed
		}

		// Deserialize if needed
		if serializationMethod == JSON_SERIALIZATION {
			var jsonData interface{}
			if err := json.Unmarshal(payloadMsg, &jsonData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
			}
			result.PayloadMsg = jsonData
		} else {
			result.PayloadMsg = string(payloadMsg)
		}
	}

	return result, nil
}

// parseASRResult extracts ASR result from payload
func (s *ASRService) parseASRResult(payload interface{}) *model.ASRResponse {
	if payload == nil {
		s.logger.Debugf("🚫 parseASRResult: payload is nil")
		return nil
	}

	b, _ := json.Marshal(payload)
	s.logger.Debugf("🔍 Parsing ASR result: %s", string(b))

	// Try to parse as JSON object
	if jsonMap, ok := payload.(map[string]interface{}); ok {
		s.logger.Debugf("📊 ASR JSON payload keys: %v", getMapKeys(jsonMap))

		result := &model.ASRResponse{
			Text:       "",
			Confidence: 0.0,
			IsFinal:    false,
		}

		// Extract result from nested structure
		if resultData, exists := jsonMap["result"]; exists {
			if resultMap, ok := resultData.(map[string]interface{}); ok {
				// Extract main text
				if text, exists := resultMap["text"]; exists {
					if textStr, ok := text.(string); ok {
						result.Text = textStr
					}
				}

				// Extract utterances for more detailed information
				if utterances, exists := resultMap["utterances"]; exists {
					if utteranceList, ok := utterances.([]interface{}); ok && len(utteranceList) > 0 {
						if utterance, ok := utteranceList[0].(map[string]interface{}); ok {
							// Check if definite (final result)
							if definite, exists := utterance["definite"]; exists {
								if definiteBool, ok := definite.(bool); ok {
									result.IsFinal = definiteBool
								}
							}
							
							// Use utterance text if available (more accurate)
							if utteranceText, exists := utterance["text"]; exists {
								if utteranceTextStr, ok := utteranceText.(string); ok {
									result.Text = utteranceTextStr
								}
							}
						}
					}
				}
			}
		}

		// Set confidence to 1.0 if we have definite results, otherwise 0.5
		if result.IsFinal {
			result.Confidence = 1.0
		} else {
			result.Confidence = 0.5
		}

		return result
	}

	return nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getMapKeys helper function for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
