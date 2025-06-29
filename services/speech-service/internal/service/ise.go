package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ai-tutor-monorepo/services/speech-service/internal/config"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/model"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ISE Protocol constants
const (
	ISE_CMD_START_BUSINESS = "ssb" // Start business
	ISE_CMD_AUDIO_WRITE    = "auw" // Audio write
	ISE_STATUS_FIRST_FRAME = 0     // First frame
	ISE_STATUS_CONTINUE    = 1     // Continue frame
	ISE_STATUS_LAST_FRAME  = 2     // Last frame
	ISE_AUS_FIRST_CHUNK    = 1     // First audio chunk
	ISE_AUS_CONTINUE_CHUNK = 2     // Continue audio chunk
	ISE_AUS_LAST_CHUNK     = 4     // Last audio chunk
)

// XML structures for parsing iFlytek ISE response - updated to match actual API response
type ISEXMLResponse struct {
	XMLName      xml.Name     `xml:"xml_result"`
	ReadSentence ReadSentence `xml:"read_sentence"`
}

type ReadSentence struct {
	XMLName  xml.Name `xml:"read_sentence"`
	Language string   `xml:"lan,attr"`
	Type     string   `xml:"type,attr"`
	Version  string   `xml:"version,attr"`
	RecPaper RecPaper `xml:"rec_paper"`
}

type RecPaper struct {
	XMLName     xml.Name    `xml:"rec_paper"`
	ReadChapter ReadChapter `xml:"read_chapter"`
}

type ReadChapter struct {
	XMLName        xml.Name   `xml:"read_chapter"`
	AccuracyScore  float64    `xml:"accuracy_score,attr"`
	FluencyScore   float64    `xml:"fluency_score,attr"`
	IntegrityScore float64    `xml:"integrity_score,attr"`
	TotalScore     float64    `xml:"total_score,attr"`
	StandardScore  float64    `xml:"standard_score,attr"`
	Content        string     `xml:"content,attr"`
	WordCount      int        `xml:"word_count,attr"`
	Sentences      []Sentence `xml:"sentence"`
}

type Word struct {
	XMLName     xml.Name `xml:"word"`
	Index       int      `xml:"index,attr"`
	GlobalIndex int      `xml:"global_index,attr"`
	Content     string   `xml:"content,attr"`
	TotalScore  float64  `xml:"total_score,attr"`
	BegPos      int      `xml:"beg_pos,attr"`
	EndPos      int      `xml:"end_pos,attr"`
	DpMessage   int      `xml:"dp_message,attr"`
	Property    int      `xml:"property,attr"`
	Pitch       string   `xml:"pitch,attr"`
	PitchBeg    int      `xml:"pitch_beg,attr"`
	PitchEnd    int      `xml:"pitch_end,attr"`
	Syllables   []Syll   `xml:"syll"`
}

type Syll struct {
	XMLName    xml.Name `xml:"syll"`
	BegPos     int      `xml:"beg_pos,attr"`
	EndPos     int      `xml:"end_pos,attr"`
	Content    string   `xml:"content,attr"`
	SerrMsg    int      `xml:"serr_msg,attr"`
	SyllAccent int      `xml:"syll_accent,attr"`
	SyllScore  float64  `xml:"syll_score,attr"`
	Phones     []Phone  `xml:"phone"`
}

type Phone struct {
	XMLName   xml.Name `xml:"phone"`
	Content   string   `xml:"content,attr"`
	BegPos    int      `xml:"beg_pos,attr"`
	EndPos    int      `xml:"end_pos,attr"`
	DpMessage int      `xml:"dp_message,attr"`
	Gwpp      float64  `xml:"gwpp,attr"`
}

type Sentence struct {
	XMLName       xml.Name `xml:"sentence"`
	Index         int      `xml:"index,attr"`
	Content       string   `xml:"content,attr"`
	TotalScore    float64  `xml:"total_score,attr"`
	AccuracyScore float64  `xml:"accuracy_score,attr"`
	FluencyScore  float64  `xml:"fluency_score,attr"`
	StandardScore float64  `xml:"standard_score,attr"`
	BegPos        int      `xml:"beg_pos,attr"`
	EndPos        int      `xml:"end_pos,attr"`
	WordCount     int      `xml:"word_count,attr"`
	Words         []Word   `xml:"word"`
}

// ISEService handles intelligent speech evaluation
type ISEService struct {
	config *config.ISEConfig
	logger *logrus.Logger
	wsURL  string
	dialer *websocket.Dialer
}

// NewISEService creates a new ISE service
func NewISEService(cfg *config.ISEConfig, logger *logrus.Logger) *ISEService {
	wsURL := cfg.BaseURL
	if wsURL == "" {
		wsURL = "wss://ise-api.xfyun.cn/v2/open-ise"
	}

	return &ISEService{
		config: cfg,
		logger: logger,
		wsURL:  wsURL,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 30 * time.Second,
			ReadBufferSize:   4096,
			WriteBufferSize:  4096,
		},
	}
}

// EvaluateSpeech evaluates speech quality and pronunciation
func (s *ISEService) EvaluateSpeech(request *model.ISERequest) (*model.ISEResponse, error) {
	if len(request.AudioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}

	if strings.TrimSpace(request.Text) == "" {
		return nil, fmt.Errorf("empty reference text")
	}

	s.logger.Infof("🎯 ISE Processing: %d bytes audio, text: '%s', language: %s",
		len(request.AudioData), request.Text, request.Language)

	// Create authenticated WebSocket connection
	conn, err := s.createAuthenticatedConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated connection: %v", err)
	}
	defer conn.Close()

	// Send business parameters
	if err := s.sendBusinessParameters(conn, request); err != nil {
		return nil, fmt.Errorf("failed to send business parameters: %v", err)
	}

	// Read initial response with standard timeout
	initialTimeout := 30 * time.Second
	if _, err := s.readResponseWithTimeout(conn, initialTimeout); err != nil {
		return nil, fmt.Errorf("failed to read initial response: %v", err)
	}

	// Send audio data and get evaluation results
	result, err := s.sendAudioAndGetResults(conn, request.AudioData)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio evaluation: %v", err)
	}

	s.logger.Infof("✅ ISE Evaluation complete: overall score %.2f", result.OverallScore)
	return result, nil
}

// createAuthenticatedConnection creates WebSocket connection with authentication
func (s *ISEService) createAuthenticatedConnection() (*websocket.Conn, error) {
	// Generate authentication URL
	authURL, err := s.generateAuthURL()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth URL: %v", err)
	}

	s.logger.Debugf("🔐 Connecting to ISE with auth: %s", authURL)

	// Create WebSocket connection
	conn, resp, err := s.dialer.Dial(authURL, nil)
	if err != nil {
		s.logger.Errorf("❌ Failed to connect to ISE service: %v", err)
		if resp != nil {
			s.logger.Errorf("❌ Response status: %s", resp.Status)
		}
		return nil, err
	}

	// Set connection timeouts to prevent server timeout issues
	// ISE servers may timeout if processing takes too long
	writeTimeout := 30 * time.Second
	readTimeout := 60 * time.Second // Longer read timeout for ISE processing

	// Set write deadline for sending messages
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set write deadline: %v", err)
	}

	// Set read deadline for receiving responses
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set read deadline: %v", err)
	}

	s.logger.Infof("✅ Connected to ISE service successfully (read: %v, write: %v)", readTimeout, writeTimeout)
	return conn, nil
}

// generateAuthURL generates authenticated WebSocket URL
func (s *ISEService) generateAuthURL() (string, error) {
	// Parse base URL
	u, err := url.Parse(s.wsURL)
	if err != nil {
		return "", err
	}

	// Generate timestamp (RFC1123 format)
	timestamp := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// Generate signature
	authorization, err := s.generateAuthorization(u.Host, timestamp)
	if err != nil {
		return "", err
	}

	// Build query parameters
	params := url.Values{}
	params.Set("authorization", authorization)
	params.Set("host", u.Host)
	params.Set("date", timestamp)

	// Construct final URL
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// generateAuthorization generates HMAC-SHA256 authorization header
func (s *ISEService) generateAuthorization(host, date string) (string, error) {
	// Build request line
	requestLine := "GET /v2/open-ise HTTP/1.1"

	// Build signature string
	signatureString := fmt.Sprintf("host: %s\ndate: %s\n%s", host, date, requestLine)

	// Generate HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(s.config.APISecret))
	mac.Write([]byte(signatureString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Build authorization string
	authString := fmt.Sprintf(
		`api_key="%s",algorithm="hmac-sha256",headers="host date request-line",signature="%s"`,
		s.config.APIKey, signature,
	)

	// Base64 encode authorization string
	return base64.StdEncoding.EncodeToString([]byte(authString)), nil
}

// sendBusinessParameters sends business parameters to start evaluation
func (s *ISEService) sendBusinessParameters(conn *websocket.Conn, request *model.ISERequest) error {
	// Determine category based on text content
	category := s.determineCategory(request)

	// Build business parameters
	business := map[string]interface{}{
		"category": category,
		"sub":      "ise",
		"ent":      s.getEntityType(request.Language),
		"cmd":      ISE_CMD_START_BUSINESS,
		"auf":      "audio/L16;rate=16000", // Audio format
		"aue":      "raw",                  // Audio encoding
		"text":     request.Text,
		"ttp_skip": true, // Skip TTS
	}

	data := map[string]interface{}{
		"status": ISE_STATUS_FIRST_FRAME,
	}

	common := map[string]string{
		"app_id": s.config.AppID,
	}

	message := map[string]interface{}{
		"common":   common,
		"business": business,
		"data":     data,
	}
	businessJson, _ := json.Marshal(business)
	s.logger.Infof("✅ ISE req business :%s", businessJson)

	return s.sendJSONMessage(conn, message)
}

// determineCategory determines the evaluation category based on text content
func (s *ISEService) determineCategory(request *model.ISERequest) string {
	if request.Category != "" {
		return request.Category
	}

	text := strings.TrimSpace(request.Text)
	words := strings.Fields(text)

	switch {
	case len(words) == 1:
		// Single word
		if len(text) <= 3 {
			return "read_syllable" // Short syllable
		}
		return "read_word"
	case len(words) <= 5:
		return "read_sentence" // Short sentence
	default:
		return "read_chapter" // Long passage
	}
}

// getEntityType returns entity type based on language
func (s *ISEService) getEntityType(language string) string {
	switch language {
	case "zh_cn":
		return "cn_vip" // Chinese evaluation
	case "en_us", "en":
		return "en_vip" // English evaluation
	default:
		return "en_vip" // Default to English
	}
}

// sendAudioAndGetResults sends audio data and returns evaluation results
func (s *ISEService) sendAudioAndGetResults(conn *websocket.Conn, audioData []byte) (*model.ISEResponse, error) {
	chunkSize := 1280 // ~40ms of 16kHz 16-bit mono audio (optimal for ISE)
	chunks := s.splitAudioData(audioData, chunkSize)
	totalChunks := len(chunks)

	// Filter out silent chunks (first few chunks are often silent)
	validChunks := s.filterSilentChunks(chunks)
	if len(validChunks) == 0 {
		return nil, fmt.Errorf("no valid audio data found (all chunks are silent)")
	}

	s.logger.Debugf("🔊 Filtered audio chunks: %d -> %d (removed %d silent chunks)",
		totalChunks, len(validChunks), totalChunks-len(validChunks))

	// Combine all valid chunks into one continuous audio stream
	var combinedAudio []byte
	for _, chunk := range validChunks {
		combinedAudio = append(combinedAudio, chunk...)
	}

	s.logger.Debugf("📦 Combined valid audio: %d bytes from %d chunks", len(combinedAudio), len(validChunks))

	// ISE API limit: entire JSON message must be <= 26000 bytes
	// JSON includes: {"common":{},"business":{},"data":{"data":"base64..."}}
	// JSON overhead ≈ 1500 bytes (including field names, quotes, etc)
	// Base64 encoding increases size by ~33%: raw_size * 4/3
	// Available space: 26000 - 1500 = 24500 bytes
	// Therefore: raw_data_size <= 24500 * 3/4 ≈ 18375 bytes
	// Using 8000 bytes to reduce chunks and speed up processing for iFlytek server timeout
	maxISEChunkSize := 8000 // Reduced: minimize blocks to prevent iFlytek server 5-10s timeout

	// If audio is slightly over limit, try aggressive silence filtering to fit in one chunk
	if len(combinedAudio) > maxISEChunkSize && len(combinedAudio) <= int(float64(maxISEChunkSize)*1.5) {
		s.logger.Debugf("🔄 Audio slightly oversized (%d bytes), trying aggressive silence filtering", len(combinedAudio))
		aggressiveChunks := s.splitAudioData(combinedAudio, 1280)
		aggressiveFiltered := s.filterSilentChunksAggressive(aggressiveChunks)

		var recompressedAudio []byte
		for _, chunk := range aggressiveFiltered {
			recompressedAudio = append(recompressedAudio, chunk...)
		}

		if len(recompressedAudio) <= maxISEChunkSize {
			s.logger.Debugf("✅ Aggressive filtering successful: %d -> %d bytes, using single chunk", len(combinedAudio), len(recompressedAudio))
			combinedAudio = recompressedAudio
		}
	}

	if len(combinedAudio) <= maxISEChunkSize {
		// Send all audio as one chunk
		return s.sendSingleAudioChunk(conn, combinedAudio)
	} else {
		// Split into multiple chunks if too large
		return s.sendMultipleAudioChunks(conn, combinedAudio, maxISEChunkSize)
	}
}

// sendSingleAudioChunk sends all audio as one chunk
func (s *ISEService) sendSingleAudioChunk(conn *websocket.Conn, audioData []byte) (*model.ISEResponse, error) {
	s.logger.Debugf("📤 Sending single audio chunk: %d bytes", len(audioData))

	// Send the audio chunk - for single chunk, use ISE_AUS_LAST_CHUNK (4) not ISE_AUS_FIRST_CHUNK (1)
	// According to iFlytek API: single chunk must have aus=4 to indicate it's the final chunk
	if err := s.sendAudioChunk(conn, audioData, ISE_AUS_LAST_CHUNK, true); err != nil {
		return nil, fmt.Errorf("failed to send audio chunk: %v", err)
	}

	s.logger.Debugf("✅ Single audio chunk sent, listening for responses...")

	// Listen for responses until we get the final result (status=2)
	// Even for single chunk, server might send multiple responses
	standardTimeout := 60 * time.Second
	maxResponses := 5 // Reasonable limit for single chunk

	for responseCount := 0; responseCount < maxResponses; responseCount++ {
		response, err := s.readResponseWithTimeout(conn, standardTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to read response %d: %v", responseCount+1, err)
		}

		s.logger.Debugf("📥 Received single chunk response %d", responseCount+1)

		// Check if this is the final evaluation result (status=2)
		if last, result := s.parseEvaluationResult(response); last && result != nil {
			result.IsFinal = true
			s.logger.Infof("✅ ISE single chunk result received: score %.2f", result.OverallScore)
			return result, nil
		}
	}

	s.logger.Warnf("⚠️ No final evaluation result received after %d responses for single chunk", maxResponses)
	return &model.ISEResponse{OverallScore: 0.0, IsFinal: true}, nil
}

// sendMultipleAudioChunks splits large audio into ISE-compatible chunks
func (s *ISEService) sendMultipleAudioChunks(conn *websocket.Conn, audioData []byte, maxChunkSize int) (*model.ISEResponse, error) {
	// Split audio into chunks that respect ISE size limits
	// Ensure chunks are aligned to 16-bit sample boundaries (2 bytes per sample)
	var chunks [][]byte
	for i := 0; i < len(audioData); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(audioData) {
			end = len(audioData)
		}

		// Ensure chunk ends on sample boundary (even byte count for 16-bit audio)
		if (end-i)%2 == 1 && end < len(audioData) {
			end-- // Adjust to maintain sample alignment
		}

		chunks = append(chunks, audioData[i:end])
	}

	// Filter out chunks that are mostly silent (especially the last chunk)
	var filteredChunks [][]byte
	silenceThreshold := int16(500)

	for i, chunk := range chunks {
		// For the last chunk, be more strict about silence filtering
		isLastChunk := i == len(chunks)-1
		if isLastChunk && len(chunks) > 1 && s.isChunkSilent(chunk, silenceThreshold) {
			s.logger.Debugf("🔇 Skipping silent last chunk (%d bytes) to avoid ISE errors", len(chunk))
			continue
		}
		filteredChunks = append(filteredChunks, chunk)
	}

	if len(filteredChunks) == 0 {
		return nil, fmt.Errorf("no valid audio chunks after filtering")
	}

	s.logger.Debugf("📤 Sending %d audio chunks with ISE size limits (filtered from %d)", len(filteredChunks), len(chunks))

	// For multiple chunks, extend timeout proportionally to avoid server timeout
	// ISE server may timeout if processing multiple chunks takes too long
	// iFlytek server appears to have 5-10s timeout, so minimize our delays
	baseTimeout := 30 * time.Second                                                   // Reduced from 60s to match server limits
	extendedTimeout := baseTimeout + time.Duration(len(filteredChunks)*5)*time.Second // Reduced from 15s to 5s per chunk
	s.logger.Debugf("⏰ Setting extended timeout for %d chunks: %v (base: %v + %v per chunk)",
		len(filteredChunks), extendedTimeout, baseTimeout, time.Duration(5)*time.Second)

	s.logger.Debugf("📤 ISE filteredChunks %d", len(filteredChunks))

	// First phase: Send all audio chunks without waiting for individual responses
	for i, chunk := range filteredChunks {
		isFirst := i == 0
		isLast := i == len(filteredChunks)-1

		// Determine audio chunk status
		var aus int
		if isFirst {
			aus = ISE_AUS_FIRST_CHUNK
		} else if isLast {
			aus = ISE_AUS_LAST_CHUNK
		} else {
			aus = ISE_AUS_CONTINUE_CHUNK
		}

		s.logger.Debugf("📤 Sending chunk %d/%d: %d bytes (aus=%d)", i+1, len(filteredChunks), len(chunk), aus)

		// Send audio chunk without waiting for response
		if err := s.sendAudioChunk(conn, chunk, aus, isLast); err != nil {
			return nil, fmt.Errorf("failed to send chunk %d: %v", i+1, err)
		}
	}

	s.logger.Debugf("✅ All %d audio chunks sent, now listening for responses...", len(filteredChunks))

	// Second phase: Listen for responses until we get the final result (status=2)
	expectedResponses := len(filteredChunks)
	receivedResponses := 0

	for receivedResponses < expectedResponses+10 { // Safety limit
		response, err := s.readResponseWithTimeout(conn, extendedTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to read response %d: %v", receivedResponses+1, err)
		}

		receivedResponses++
		s.logger.Debugf("📥 Received response %d/%d", receivedResponses, expectedResponses)
		// Check if this is the final evaluation result (status=2)
		if last, result := s.parseEvaluationResult(response); last && result != nil {
			s.logger.Infof("✅ ISE final evaluation result received: score %.2f", result.OverallScore)
			return result, nil
		} else if last {
			s.logger.Warnf("⚠️ Received %d & last responses but no final result, stopping", receivedResponses)
			break
		}
	}

	// If we reach here, no final result was received despite sending all chunks
	s.logger.Warnf("⚠️ No final evaluation result received after %d responses", receivedResponses)
	return &model.ISEResponse{OverallScore: 0.0, IsFinal: true}, nil
}

// filterSilentChunks removes chunks that are mostly silent
func (s *ISEService) filterSilentChunks(chunks [][]byte) [][]byte {
	var validChunks [][]byte
	silenceThreshold := int16(500) // Increased threshold for more strict silence detection

	for i, chunk := range chunks {
		if s.isChunkSilent(chunk, silenceThreshold) {
			s.logger.Debugf("🔇 Skipping silent chunk %d (%d bytes)", i, len(chunk))
			continue
		}
		validChunks = append(validChunks, chunk)
	}

	return validChunks
}

// filterSilentChunksAggressive removes chunks with aggressive silence filtering
func (s *ISEService) filterSilentChunksAggressive(chunks [][]byte) [][]byte {
	var validChunks [][]byte
	silenceThreshold := int16(800) // Much higher threshold for aggressive filtering

	for i, chunk := range chunks {
		if s.isChunkSilent(chunk, silenceThreshold) {
			s.logger.Debugf("🔇 Aggressively skipping silent chunk %d (%d bytes)", i, len(chunk))
			continue
		}
		validChunks = append(validChunks, chunk)
	}

	return validChunks
}

// isChunkSilent checks if an audio chunk is mostly silent
func (s *ISEService) isChunkSilent(chunk []byte, threshold int16) bool {
	if len(chunk) < 2 {
		return true
	}

	// Count samples above threshold
	samples := len(chunk) / 2 // 16-bit samples
	loudSamples := 0

	for i := 0; i < len(chunk)-1; i += 2 {
		// Read 16-bit little-endian sample correctly
		sample := int16(chunk[i]) | (int16(chunk[i+1]) << 8)
		if sample < 0 {
			sample = -sample // Get absolute value
		}

		if sample > threshold {
			loudSamples++
		}
	}

	// If less than 10% of samples are above threshold, consider it silent
	silentRatio := float64(loudSamples) / float64(samples)
	isSilent := silentRatio < 0.10

	// Debug log for first few chunks
	if len(chunk) == 1280 { // Only log standard chunks
		s.logger.Debugf("🔍 Chunk analysis: %d samples, %d loud samples (%.1f%%), threshold=%d, silent=%v",
			samples, loudSamples, silentRatio*100, threshold, isSilent)
	}

	return isSilent
}

// sendAudioChunk sends a single audio chunk
func (s *ISEService) sendAudioChunk(conn *websocket.Conn, chunk []byte, aus int, isLast bool) error {
	// According to iFlytek API: first and continue frames use status=1, last frame uses status=2
	status := ISE_STATUS_CONTINUE // status=1 for first and continue frames
	if isLast {
		status = ISE_STATUS_LAST_FRAME // status=2 for last frame
	}

	// Encode audio data to base64
	audioBase64 := base64.StdEncoding.EncodeToString(chunk)

	data := map[string]interface{}{
		"status":   status,
		"format":   "audio/L16;rate=16000",
		"data":     audioBase64,
		"encoding": "raw",
	}

	common := map[string]string{
		"app_id": s.config.AppID,
	}

	business := map[string]interface{}{
		"cmd": ISE_CMD_AUDIO_WRITE,
		"aus": aus,
	}

	message := map[string]interface{}{
		"common":   common,
		"business": business,
		"data":     data,
	}

	return s.sendJSONMessage(conn, message)
}

// splitAudioData splits audio data into chunks
func (s *ISEService) splitAudioData(data []byte, chunkSize int) [][]byte {
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

// sendJSONMessage sends a JSON message over WebSocket
func (s *ISEService) sendJSONMessage(conn *websocket.Conn, message map[string]interface{}) error {
	// Set write deadline before each write operation
	writeTimeout := 30 * time.Second
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set write deadline: %v", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(writeTimeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set read deadline: %v", err)
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON message: %v", err)
	}

	s.logger.Debugf("📤 Sending ISE message: %s", string(jsonData))

	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		// Log specific timeout errors for debugging
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			s.logger.Errorf("⏰ ISE write timeout after %v: %v", writeTimeout, err)
		}
		return err
	}

	return nil
}

// readResponse reads and returns response from WebSocket
func (s *ISEService) readResponse(conn *websocket.Conn) (map[string]interface{}, error) {
	// Set read deadline before each read operation
	readTimeout := 60 * time.Second // ISE processing can take time, especially for multiple chunks
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set read deadline: %v", err)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		// Log specific timeout errors for debugging
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			s.logger.Errorf("⏰ ISE read timeout after %v: %v", readTimeout, err)
		}
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	s.logger.Debugf("📥 Received ISE response: %s", string(message))
	return response, nil
}

// readResponseWithTimeout reads response with custom timeout
func (s *ISEService) readResponseWithTimeout(conn *websocket.Conn, timeout time.Duration) (map[string]interface{}, error) {
	// Set read deadline with custom timeout
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		s.logger.Warnf("⚠️ Failed to set read deadline: %v", err)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		// Log specific timeout errors for debugging
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			s.logger.Errorf("⏰ ISE read timeout after %v: %v", timeout, err)
		}
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	s.logger.Debugf("📥 Received ISE response: %s", string(message))
	return response, nil
}

// parseEvaluationResult parses the evaluation result from response
func (s *ISEService) parseEvaluationResult(response map[string]interface{}) (bool, *model.ISEResponse) {
	// Check if response contains evaluation data
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	// Check response status - only status=2 indicates final result according to iFlytek docs
	status, ok := data["status"].(float64)
	if !ok || status != 2 { // Status 2 indicates final result
		s.logger.Debugf("📋 ISE intermediate response (status=%v), waiting for final result", status)
		return false, nil
	}

	// Parse the evaluation result - iFlytek returns base64-encoded XML
	resultStr, ok := data["data"].(string)
	if !ok {
		s.logger.Errorf("❌ ISE response missing data field")
		return true, nil
	}

	// Decode base64 result
	resultBytes, err := base64.StdEncoding.DecodeString(resultStr)
	if err != nil {
		s.logger.Errorf("❌ Failed to decode ISE base64 result: %v", err)
		return true, nil
	}

	s.logger.Debugf("🔍 Raw ISE XML result: %s", string(resultBytes))

	// Parse XML result (not JSON!)
	var xmlResult ISEXMLResponse
	if err := xml.Unmarshal(resultBytes, &xmlResult); err != nil {
		s.logger.Errorf("❌ Failed to unmarshal ISE XML result: %v", err)
		s.logger.Debugf("Raw XML content: %s", string(resultBytes))
		return true, nil
	}

	// Convert XML structure to our model
	chapter := xmlResult.ReadSentence.RecPaper.ReadChapter
	result := &model.ISEResponse{
		IsFinal:           true,
		OverallScore:      chapter.TotalScore,
		AccuracyScore:     chapter.AccuracyScore,
		FluencyScore:      chapter.FluencyScore,
		CompletenessScore: chapter.IntegrityScore, // IntegrityScore maps to CompletenessScore
		WordScores:        s.convertXMLWordScores(chapter.Sentences),
		PhoneScores:       s.convertXMLPhoneScores(chapter.Sentences),
		SentenceScores:    s.convertXMLSentenceScores(chapter.Sentences),
	}

	s.logger.Infof("✅ ISE XML parsing successful: overall score %.2f", result.OverallScore)
	return true, result
}

// convertXMLWordScores parses word-level evaluation scores from sentences
func (s *ISEService) convertXMLWordScores(sentences []Sentence) []model.WordScore {
	var wordScores []model.WordScore

	for _, sentence := range sentences {
		for _, word := range sentence.Words {
			wordScore := model.WordScore{}

			wordScore.Word = word.Content
			wordScore.Score = word.TotalScore
			wordScore.StartTime = int64(word.BegPos)
			wordScore.EndTime = int64(word.EndPos)
			wordScore.IsCorrect = word.DpMessage == 0 // DpMessage=0 means correct
			wordScore.Confidence = word.TotalScore

			wordScores = append(wordScores, wordScore)
		}
	}

	return wordScores
}

// convertXMLPhoneScores parses phoneme-level evaluation scores from sentences
func (s *ISEService) convertXMLPhoneScores(sentences []Sentence) []model.PhoneScore {
	var phoneScores []model.PhoneScore

	for _, sentence := range sentences {
		for _, word := range sentence.Words {
			for _, syll := range word.Syllables {
				for _, phone := range syll.Phones {
					phoneScore := model.PhoneScore{}

					phoneScore.Phone = phone.Content
					phoneScore.Score = math.Abs(phone.Gwpp)     // Use absolute value of GWPP score
					phoneScore.IsCorrect = phone.DpMessage == 0 // DpMessage=0 means correct
					phoneScore.StartTime = int64(phone.BegPos)
					phoneScore.EndTime = int64(phone.EndPos)

					phoneScores = append(phoneScores, phoneScore)
				}
			}
		}
	}

	return phoneScores
}

// convertXMLSentenceScores parses sentence-level evaluation scores from XML
func (s *ISEService) convertXMLSentenceScores(sentences []Sentence) []model.SentenceScore {
	var sentenceScores []model.SentenceScore

	for _, sentence := range sentences {
		sentenceScore := model.SentenceScore{}

		sentenceScore.Sentence = sentence.Content
		sentenceScore.Score = sentence.TotalScore
		sentenceScore.AccuracyScore = sentence.AccuracyScore
		sentenceScore.FluencyScore = sentence.FluencyScore
		sentenceScore.TotalWords = sentence.WordCount
		sentenceScore.CorrectWords = s.countCorrectWords(sentence.Words)

		sentenceScores = append(sentenceScores, sentenceScore)
	}

	return sentenceScores
}

// countCorrectWords counts the number of correctly pronounced words
func (s *ISEService) countCorrectWords(words []Word) int {
	correctCount := 0
	for _, word := range words {
		if word.DpMessage == 0 { // DpMessage=0 means correct pronunciation
			correctCount++
		}
	}
	return correctCount
}
