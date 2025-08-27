package ai

import "fmt"

// AIService interface defines methods that both local and remote AI services implement
type AIService interface {
	// ProcessText handles text-based AI processing - returns raw OpenRouter response for local parsing
	ProcessText(request *TextProcessingRequest) (*OpenRouterResponse, error)
	
	// ProcessAudio handles audio processing (transcription)
	ProcessAudio(audioFile string) (*AudioProcessingResult, error)
}

// ParseTextResponse parses a raw OpenRouter response into a TextProcessingResult
// This shared parsing logic ensures consistency between local and remote services
func ParseTextResponse(response *OpenRouterResponse, taskType string) (*TextProcessingResult, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	
	if response.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", response.Error.Message)
	}
	
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	
	content := response.Choices[0].Message.Content
	
	return &TextProcessingResult{
		Content:    content,
		TaskType:   taskType,
		Structured: nil, // Raw text response, no structured parsing yet
		TokensUsed: 0,   // OpenRouter doesn't provide token count in this format
	}, nil
}

// Ensure LocalAIService implements AIService
var _ AIService = (*LocalAIService)(nil)

// Ensure RemoteAIService implements AIService  
var _ AIService = (*RemoteAIService)(nil)