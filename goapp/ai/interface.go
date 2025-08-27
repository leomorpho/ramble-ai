package ai

// AIService interface defines methods that both local and remote AI services implement
type AIService interface {
	// ProcessText handles text-based AI processing
	ProcessText(request *TextProcessingRequest) (*TextProcessingResult, error)
	
	// ProcessAudio handles audio processing (transcription)
	ProcessAudio(audioFile string) (*AudioProcessingResult, error)
}

// Ensure LocalAIService implements AIService
var _ AIService = (*LocalAIService)(nil)

// Ensure RemoteAIService implements AIService  
var _ AIService = (*RemoteAIService)(nil)