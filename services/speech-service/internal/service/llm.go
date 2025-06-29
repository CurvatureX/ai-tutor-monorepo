package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/config"
	"github.com/ai-tutor-monorepo/services/speech-service/internal/model"
)

// LLMService handles language model interactions
type LLMService struct {
	config *config.LLMConfig
	client *resty.Client
	logger *logrus.Logger
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.LLMConfig, logger *logrus.Logger) *LLMService {
	client := resty.New()
	
	// Set base URL if provided
	if cfg.BaseURL != "" {
		client.SetBaseURL(cfg.BaseURL)
	}

	return &LLMService{
		config: cfg,
		client: client,
		logger: logger,
	}
}

// GenerateResponse generates a conversational response for English practice
func (s *LLMService) GenerateResponse(userText, context string) (*model.LLMResponse, error) {
	if strings.TrimSpace(userText) == "" {
		return nil, fmt.Errorf("empty user input")
	}

	s.logger.Debugf("Generating LLM response for: %s", userText)

	// Construct the prompt for English conversation practice
	prompt := s.buildConversationPrompt(userText, context)

	// Prepare the request
	requestBody := map[string]interface{}{
		"model": s.config.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": s.getSystemPrompt(),
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   150,
		"temperature":  0.7,
		"stream":      false,
	}

	// Make the API request
	resp, err := s.client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post("/api/v3/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("failed to call LLM API: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	// Parse response
	var apiResponse LLMAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	responseText := apiResponse.Choices[0].Message.Content
	
	// Clean up the response
	responseText = strings.TrimSpace(responseText)
	
	s.logger.Debugf("LLM response generated: %s", responseText)

	return &model.LLMResponse{
		Reply:   responseText,
		Context: s.updateContext(context, userText, responseText),
	}, nil
}

// getSystemPrompt returns the system prompt for English conversation practice
func (s *LLMService) getSystemPrompt() string {
	return `You are an AI English conversation partner designed to help users practice English speaking. 

Your role:
- Act as a friendly, patient, and encouraging conversation partner
- Help users practice natural English conversation
- Provide gentle corrections when users make mistakes
- Ask follow-up questions to keep the conversation flowing
- Use simple to intermediate English appropriate for language learners
- Be supportive and positive in your responses

Guidelines:
- Keep responses concise (1-3 sentences)
- Speak naturally and conversationally
- If the user makes a grammar or vocabulary mistake, gently suggest the correct form
- Ask questions to encourage the user to speak more
- Cover various topics like daily life, hobbies, travel, food, etc.
- Adapt your language level to match the user's proficiency

Remember: Your goal is to help the user practice speaking English in a comfortable, non-judgmental environment.`
}

// buildConversationPrompt builds the conversation prompt
func (s *LLMService) buildConversationPrompt(userText, context string) string {
	prompt := fmt.Sprintf("User said: \"%s\"\n\n", userText)
	
	if context != "" {
		prompt += fmt.Sprintf("Previous conversation context: %s\n\n", context)
	}
	
	prompt += "Please respond naturally as an English conversation partner. Keep your response concise and engaging."
	
	return prompt
}

// updateContext updates the conversation context
func (s *LLMService) updateContext(currentContext, userText, aiResponse string) string {
	// Keep the last few exchanges for context
	newEntry := fmt.Sprintf("User: %s | AI: %s", userText, aiResponse)
	
	if currentContext == "" {
		return newEntry
	}
	
	// Keep only the last 2-3 exchanges to avoid context becoming too long
	contextParts := strings.Split(currentContext, " | AI: ")
	if len(contextParts) > 4 { // More than 2 complete exchanges
		// Keep only the last 2 exchanges
		contextParts = contextParts[len(contextParts)-4:]
		currentContext = strings.Join(contextParts, " | AI: ")
	}
	
	return currentContext + " | " + newEntry
}

// GenerateCorrection generates a correction for user's English
func (s *LLMService) GenerateCorrection(userText string) (*model.LLMResponse, error) {
	if strings.TrimSpace(userText) == "" {
		return nil, fmt.Errorf("empty user input")
	}

	prompt := fmt.Sprintf(`Please analyze this English text for grammar, vocabulary, and pronunciation issues: "%s"

If there are mistakes, provide:
1. The corrected version
2. A brief explanation of what was wrong
3. An encouraging comment

If the English is already correct, just provide positive feedback and maybe suggest an alternative way to express the same idea.

Keep your response brief and encouraging.`, userText)

	requestBody := map[string]interface{}{
		"model": s.config.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are an English language tutor providing gentle corrections and feedback to language learners.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   100,
		"temperature":  0.3,
		"stream":      false,
	}

	resp, err := s.client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post("/api/v3/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("failed to call LLM API for correction: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var apiResponse LLMAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse correction response: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no correction response from LLM")
	}

	responseText := strings.TrimSpace(apiResponse.Choices[0].Message.Content)

	return &model.LLMResponse{
		Reply:   responseText,
		Context: "",
	}, nil
}

// GenerateConversationStarter generates a conversation starter
func (s *LLMService) GenerateConversationStarter() (*model.LLMResponse, error) {
	prompt := `Generate a friendly conversation starter for an English language learner. The starter should:
- Be simple and accessible for intermediate English learners
- Be engaging and encourage response
- Cover everyday topics like hobbies, daily life, food, travel, etc.
- Be just one or two sentences

Examples of good starters:
- "What did you have for breakfast today? I love trying different morning foods!"
- "Do you have any fun plans for the weekend?"
- "What's your favorite season and why?"

Please generate one new conversation starter.`

	requestBody := map[string]interface{}{
		"model": s.config.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are an English conversation partner helping language learners practice.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   50,
		"temperature":  0.8,
		"stream":      false,
	}

	resp, err := s.client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post("/api/v3/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("failed to call LLM API for starter: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var apiResponse LLMAPIResponse
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse starter response: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no starter response from LLM")
	}

	responseText := strings.TrimSpace(apiResponse.Choices[0].Message.Content)

	return &model.LLMResponse{
		Reply:   responseText,
		Context: "",
	}, nil
}

// LLMAPIResponse represents the response from the LLM API
type LLMAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}