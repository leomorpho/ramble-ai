package ai

import (
	"context"

	"ramble-ai/ent"
)

// LocalAIService wraps CoreAIService with pre-loaded API keys
type LocalAIService struct {
	coreService   *CoreAIService
	openaiKey     string
	openrouterKey string
}

// NewLocalAIService creates a new local AI service with pre-loaded API keys
func NewLocalAIService(client *ent.Client, ctx context.Context, openaiKey, openrouterKey string) *LocalAIService {
	return &LocalAIService{
		coreService:   NewCoreAIService(client, ctx),
		openaiKey:     openaiKey,
		openrouterKey: openrouterKey,
	}
}

// ProcessText implements AIService interface
func (s *LocalAIService) ProcessText(request *TextProcessingRequest) (*OpenRouterResponse, error) {
	// Use OpenRouter API key for text processing - returns raw response for local parsing
	return s.coreService.ProcessText(request, s.openrouterKey)
}

// ProcessAudio implements AIService interface
func (s *LocalAIService) ProcessAudio(audioFile string) (*AudioProcessingResult, error) {
	// Use OpenAI API key for audio processing (transcription)
	return s.coreService.ProcessAudio(audioFile, s.openaiKey)
}