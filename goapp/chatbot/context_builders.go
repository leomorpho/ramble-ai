package chatbot

import (
	"fmt"
	"strings"
)

// HighlightOrderingContextBuilder builds context for highlight ordering operations
type HighlightOrderingContextBuilder struct{}

// BuildContext builds context about current highlights for the LLM
func (h *HighlightOrderingContextBuilder) BuildContext(projectID int, service *ChatbotService) (string, error) {
	// Get current order
	currentOrder, err := service.highlightService.GetProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return "", err
	}
	
	// Get highlight summaries
	projectHighlights, err := service.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	// Build optimized context string for faster processing
	var contextBuilder strings.Builder
	
	// Create highlight ID to text map for efficient lookup
	highlightMap := make(map[string]string)
	allIDs := []string{}
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			// Keep full text for better context, but still reasonable length
			text := h.Text
			if len(text) > 150 {
				text = text[:150] + "..."
			}
			highlightMap[h.ID] = text
			allIDs = append(allIDs, h.ID)
		}
	}
	
	// Provide all highlight references with ID to text mapping
	contextBuilder.WriteString("Available highlights for reordering:\n")
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			contextBuilder.WriteString(fmt.Sprintf("- %s: \"%s\"\n", h.ID, highlightMap[h.ID]))
		}
	}
	
	contextBuilder.WriteString(fmt.Sprintf("\nTotal: %d highlights - ALL highlight IDs must be included in your new_order array.\n", len(allIDs)))
	
	// Ask if user wants to see current order or start fresh
	contextBuilder.WriteString("\nCurrent highlight order (for reference - ask user if they want to use this as starting point):\n")
	for i, item := range currentOrder {
		switch v := item.(type) {
		case string:
			contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, v))
		case map[string]interface{}:
			if title, ok := v["title"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION] %s\n", i+1, title))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION]\n", i+1))
			}
		}
	}
	
	return contextBuilder.String(), nil
}

// GetContextDescription returns a description of what context this builder provides
func (h *HighlightOrderingContextBuilder) GetContextDescription() string {
	return "Highlight content mapping and current order reference for reordering operations"
}

// GenericContextBuilder provides basic project context
type GenericContextBuilder struct{}

// BuildContext builds basic project context
func (g *GenericContextBuilder) BuildContext(projectID int, service *ChatbotService) (string, error) {
	// Get project highlights for basic context
	projectHighlights, err := service.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Project ID: %d\n", projectID))
	contextBuilder.WriteString(fmt.Sprintf("Total highlights: %d\n\n", len(projectHighlights)))
	
	// Add basic highlight information
	for _, ph := range projectHighlights {
		if len(ph.Highlights) == 0 {
			continue
		}
		
		contextBuilder.WriteString(fmt.Sprintf("Video: %s (%d highlights)\n", ph.VideoClipName, len(ph.Highlights)))
		for i, h := range ph.Highlights {
			text := h.Text
			if len(text) > 60 {
				text = text[:60] + "..."
			}
			contextBuilder.WriteString(fmt.Sprintf("  %d. \"%s\"\n", i+1, text))
		}
		contextBuilder.WriteString("\n")
	}
	
	return contextBuilder.String(), nil
}

// GetContextDescription returns a description of what context this builder provides
func (g *GenericContextBuilder) GetContextDescription() string {
	return "Basic project information and highlight overview"
}

// ContentAnalysisContextBuilder provides detailed content for analysis
type ContentAnalysisContextBuilder struct{}

// BuildContext builds detailed content context for analysis
func (c *ContentAnalysisContextBuilder) BuildContext(projectID int, service *ChatbotService) (string, error) {
	// Get project highlights with full content
	projectHighlights, err := service.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Content Analysis Context for Project %d\n\n", projectID))
	
	totalHighlights := 0
	totalTextLength := 0
	
	// Build detailed content overview
	for _, ph := range projectHighlights {
		if len(ph.Highlights) == 0 {
			continue
		}
		
		contextBuilder.WriteString(fmt.Sprintf("=== Video: %s ===\n", ph.VideoClipName))
		contextBuilder.WriteString(fmt.Sprintf("Duration: %.1f seconds\n", ph.Duration))
		contextBuilder.WriteString(fmt.Sprintf("Highlights: %d\n\n", len(ph.Highlights)))
		
		for i, h := range ph.Highlights {
			contextBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n\n", i+1, h.ID, h.Text))
			totalTextLength += len(h.Text)
			totalHighlights++
		}
		contextBuilder.WriteString("\n")
	}
	
	// Add summary statistics
	contextBuilder.WriteString("=== Content Summary ===\n")
	contextBuilder.WriteString(fmt.Sprintf("Total highlights: %d\n", totalHighlights))
	contextBuilder.WriteString(fmt.Sprintf("Total text content: %d characters\n", totalTextLength))
	if totalHighlights > 0 {
		avgLength := totalTextLength / totalHighlights
		contextBuilder.WriteString(fmt.Sprintf("Average highlight length: %d characters\n", avgLength))
	}
	
	return contextBuilder.String(), nil
}

// GetContextDescription returns a description of what context this builder provides
func (c *ContentAnalysisContextBuilder) GetContextDescription() string {
	return "Detailed content analysis with full highlight text and statistics"
}

// ExportOptimizationContextBuilder provides context for export optimization
type ExportOptimizationContextBuilder struct{}

// BuildContext builds context for export optimization
func (e *ExportOptimizationContextBuilder) BuildContext(projectID int, service *ChatbotService) (string, error) {
	// Get project highlights
	projectHighlights, err := service.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	// Get current order for duration calculations
	currentOrder, err := service.highlightService.GetProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return "", err
	}
	
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Export Optimization Context for Project %d\n\n", projectID))
	
	// Calculate total durations and highlight distribution
	totalDuration := 0.0
	videoCount := 0
	highlightCount := 0
	
	for _, ph := range projectHighlights {
		if len(ph.Highlights) > 0 {
			totalDuration += ph.Duration
			videoCount++
			highlightCount += len(ph.Highlights)
		}
	}
	
	contextBuilder.WriteString("=== Project Overview ===\n")
	contextBuilder.WriteString(fmt.Sprintf("Total videos: %d\n", videoCount))
	contextBuilder.WriteString(fmt.Sprintf("Total highlights: %d\n", highlightCount))
	contextBuilder.WriteString(fmt.Sprintf("Combined video duration: %.1f seconds\n", totalDuration))
	contextBuilder.WriteString(fmt.Sprintf("Current highlight order: %d items\n\n", len(currentOrder)))
	
	// Add video breakdown for export planning
	contextBuilder.WriteString("=== Video Breakdown ===\n")
	for _, ph := range projectHighlights {
		if len(ph.Highlights) == 0 {
			continue
		}
		
		contextBuilder.WriteString(fmt.Sprintf("Video: %s\n", ph.VideoClipName))
		contextBuilder.WriteString(fmt.Sprintf("  Duration: %.1f seconds\n", ph.Duration))
		contextBuilder.WriteString(fmt.Sprintf("  Highlights: %d\n", len(ph.Highlights)))
		contextBuilder.WriteString(fmt.Sprintf("  File: %s\n\n", ph.FilePath))
	}
	
	return contextBuilder.String(), nil
}

// GetContextDescription returns a description of what context this builder provides
func (e *ExportOptimizationContextBuilder) GetContextDescription() string {
	return "Project structure and timing information for export optimization"
}