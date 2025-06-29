package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
			HandshakeTimeout: 10 * time.Second,
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

	// Read initial response
	if _, err := s.readResponse(conn); err != nil {
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

	s.logger.Infof("✅ Connected to ISE service successfully")
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
		"text":     base64.StdEncoding.EncodeToString([]byte(request.Text)),
		"ttp_skip": true, // Skip TTS
		"vad_eos":  5000, // Voice activity detection
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

// sendAudioAndGetResults sends audio data in chunks and collects results
func (s *ISEService) sendAudioAndGetResults(conn *websocket.Conn, audioData []byte) (*model.ISEResponse, error) {
	chunkSize := 1280 // ~40ms of 16kHz 16-bit mono audio (optimal for ISE)
	chunks := s.splitAudioData(audioData, chunkSize)
	totalChunks := len(chunks)

	var finalResult *model.ISEResponse

	for i, chunk := range chunks {
		isLast := i == totalChunks-1

		// Determine audio chunk status
		var aus int
		if i == 0 {
			aus = ISE_AUS_FIRST_CHUNK
		} else if isLast {
			aus = ISE_AUS_LAST_CHUNK
		} else {
			aus = ISE_AUS_CONTINUE_CHUNK
		}

		// Send audio chunk
		if err := s.sendAudioChunk(conn, chunk, aus, isLast); err != nil {
			return nil, fmt.Errorf("failed to send audio chunk %d: %v", i, err)
		}

		// Read response
		response, err := s.readResponse(conn)
		if err != nil {
			return nil, fmt.Errorf("failed to read response for chunk %d: %v", i, err)
		}

		// Parse evaluation result if available
		if response != nil {
			if result := s.parseEvaluationResult(response); result != nil {
				s.logger.Debugf("📊 ISE partial result for chunk %d: score %.2f", i, result.OverallScore)
				finalResult = result
				if isLast {
					finalResult.IsFinal = true
				}
			}
		}
	}

	if finalResult == nil {
		finalResult = &model.ISEResponse{
			OverallScore: 0.0,
			IsFinal:      true,
		}
	}

	return finalResult, nil
}

// sendAudioChunk sends a single audio chunk
func (s *ISEService) sendAudioChunk(conn *websocket.Conn, chunk []byte, aus int, isLast bool) error {
	status := ISE_STATUS_CONTINUE
	if isLast {
		status = ISE_STATUS_LAST_FRAME
	}

	// Encode audio data to base64
	audioBase64 := base64.StdEncoding.EncodeToString(chunk)

	data := map[string]interface{}{
		"status":   status,
		"format":   "audio/L16;rate=16000",
		"audio":    audioBase64,
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
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON message: %v", err)
	}

	s.logger.Debugf("📤 Sending ISE message: %s", string(jsonData))
	return conn.WriteMessage(websocket.TextMessage, jsonData)
}

// readResponse reads and returns response from WebSocket
func (s *ISEService) readResponse(conn *websocket.Conn) (map[string]interface{}, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
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
func (s *ISEService) parseEvaluationResult(response map[string]interface{}) *model.ISEResponse {
	// Check if response contains evaluation data
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Check response status
	status, ok := data["status"].(float64)
	if !ok || status != 2 { // Status 2 indicates final result
		return nil
	}

	// Parse the evaluation result
	resultStr, ok := data["data"].(string)
	if !ok {
		return nil
	}

	// Decode base64 result
	resultBytes, err := base64.StdEncoding.DecodeString(resultStr)
	if err != nil {
		s.logger.Errorf("Failed to decode ISE result: %v", err)
		return nil
	}

	// Parse JSON result
	var evalResult map[string]interface{}
	if err := json.Unmarshal(resultBytes, &evalResult); err != nil {
		s.logger.Errorf("Failed to unmarshal ISE evaluation result: %v", err)
		return nil
	}

	s.logger.Debugf("🔍 Raw ISE evaluation result: %s", string(resultBytes))

	// Extract scores
	result := &model.ISEResponse{
		IsFinal: true,
	}

	// Parse overall scores
	if readResults, ok := evalResult["read_results"].(map[string]interface{}); ok {
		if totalScore, ok := readResults["total_score"].(float64); ok {
			result.OverallScore = totalScore
		}

		// Parse accuracy, fluency, completeness scores
		if accuracyScore, ok := readResults["accuracy_score"].(float64); ok {
			result.AccuracyScore = accuracyScore
		}
		if fluencyScore, ok := readResults["fluency_score"].(float64); ok {
			result.FluencyScore = fluencyScore
		}
		if completenessScore, ok := readResults["completeness_score"].(float64); ok {
			result.CompletenessScore = completenessScore
		}

		// Parse word-level scores
		result.WordScores = s.parseWordScores(readResults)

		// Parse phone-level scores
		result.PhoneScores = s.parsePhoneScores(readResults)

		// Parse sentence-level scores
		result.SentenceScores = s.parseSentenceScores(readResults)
	}

	return result
}

// parseWordScores parses word-level evaluation scores
func (s *ISEService) parseWordScores(readResults map[string]interface{}) []model.WordScore {
	var wordScores []model.WordScore

	if words, ok := readResults["words"].([]interface{}); ok {
		for _, wordInterface := range words {
			if word, ok := wordInterface.(map[string]interface{}); ok {
				wordScore := model.WordScore{}

				if content, ok := word["content"].(string); ok {
					wordScore.Word = content
				}
				if score, ok := word["total_score"].(float64); ok {
					wordScore.Score = score
				}
				if startTime, ok := word["start_time"].(float64); ok {
					wordScore.StartTime = int64(startTime)
				}
				if endTime, ok := word["end_time"].(float64); ok {
					wordScore.EndTime = int64(endTime)
				}
				if dpMessage, ok := word["dp_message"].(float64); ok {
					wordScore.IsCorrect = dpMessage > 0
				}
				if confidence, ok := word["confidence"].(float64); ok {
					wordScore.Confidence = confidence
				}

				wordScores = append(wordScores, wordScore)
			}
		}
	}

	return wordScores
}

// parsePhoneScores parses phoneme-level evaluation scores
func (s *ISEService) parsePhoneScores(readResults map[string]interface{}) []model.PhoneScore {
	var phoneScores []model.PhoneScore

	if words, ok := readResults["words"].([]interface{}); ok {
		for _, wordInterface := range words {
			if word, ok := wordInterface.(map[string]interface{}); ok {
				if phones, ok := word["phones"].([]interface{}); ok {
					for _, phoneInterface := range phones {
						if phone, ok := phoneInterface.(map[string]interface{}); ok {
							phoneScore := model.PhoneScore{}

							if content, ok := phone["content"].(string); ok {
								phoneScore.Phone = content
							}
							if score, ok := phone["dp_message"].(float64); ok {
								phoneScore.Score = score
								phoneScore.IsCorrect = score > 50 // Threshold for correctness
							}
							if startTime, ok := phone["start_time"].(float64); ok {
								phoneScore.StartTime = int64(startTime)
							}
							if endTime, ok := phone["end_time"].(float64); ok {
								phoneScore.EndTime = int64(endTime)
							}

							phoneScores = append(phoneScores, phoneScore)
						}
					}
				}
			}
		}
	}

	return phoneScores
}

// parseSentenceScores parses sentence-level evaluation scores
func (s *ISEService) parseSentenceScores(readResults map[string]interface{}) []model.SentenceScore {
	var sentenceScores []model.SentenceScore

	if sentences, ok := readResults["sentences"].([]interface{}); ok {
		for _, sentenceInterface := range sentences {
			if sentence, ok := sentenceInterface.(map[string]interface{}); ok {
				sentenceScore := model.SentenceScore{}

				if content, ok := sentence["content"].(string); ok {
					sentenceScore.Sentence = content
				}
				if score, ok := sentence["total_score"].(float64); ok {
					sentenceScore.Score = score
				}
				if accuracyScore, ok := sentence["accuracy_score"].(float64); ok {
					sentenceScore.AccuracyScore = accuracyScore
				}
				if fluencyScore, ok := sentence["fluency_score"].(float64); ok {
					sentenceScore.FluencyScore = fluencyScore
				}
				if totalWords, ok := sentence["word_count"].(float64); ok {
					sentenceScore.TotalWords = int(totalWords)
				}
				if correctWords, ok := sentence["correct_words"].(float64); ok {
					sentenceScore.CorrectWords = int(correctWords)
				}

				sentenceScores = append(sentenceScores, sentenceScore)
			}
		}
	}

	return sentenceScores
}
