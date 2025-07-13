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
	contextBuilder.WriteString(fmt.Sprintf("REORDER REQUEST: The user wants you to reorder %d highlights for better flow.\n\n", len(projectHighlights)))
	
	// Create highlight ID to text map for efficient lookup
	highlightMap := make(map[string]string)
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			// Truncate long text to reduce context size and improve LLM speed
			text := h.Text
			if len(text) > 60 {
				text = text[:60] + "..."
			}
			highlightMap[h.ID] = text
		}
	}
	
	contextBuilder.WriteString("Current highlight order (optimized for reordering):\n")
	for i, item := range currentOrder {
		switch v := item.(type) {
		case string:
			if text, exists := highlightMap[v]; exists {
				contextBuilder.WriteString(fmt.Sprintf("%d. %s: \"%s\"\n", i+1, v, text))
			}
		case map[string]interface{}:
			if title, ok := v["title"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION] %s\n", i+1, title))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION]\n", i+1))
			}
		}
	}
	
	// Add compact highlight reference (only first 50 for performance)
	contextBuilder.WriteString("\n\nHighlight reference (first 50 for optimal processing):\n")
	count := 0
	for _, ph := range projectHighlights {
		if count >= 50 {
			contextBuilder.WriteString(fmt.Sprintf("... and %d more highlights\n", len(projectHighlights)-count))
			break
		}
		
		for _, h := range ph.Highlights {
			if count >= 50 {
				break
			}
			contextBuilder.WriteString(fmt.Sprintf("- %s: \"%s\"\n", h.ID, highlightMap[h.ID]))
			count++
		}
	}
	
	// Add complete list of ALL highlight IDs that MUST be included in reorder
	contextBuilder.WriteString("\n\nCOMPLETE LIST OF ALL HIGHLIGHT IDs (YOU MUST INCLUDE ALL OF THESE IN YOUR REORDER):\n")
	var allIDs []string
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			allIDs = append(allIDs, h.ID)
		}
	}
	for i, id := range allIDs {
		contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, id))
	}
	contextBuilder.WriteString(fmt.Sprintf("\nTOTAL: %d highlight IDs - ALL must be in your new_order array\n", len(allIDs)))
	
	return contextBuilder.String(), nil
}

// GetContextDescription returns a description of what context this builder provides
func (h *HighlightOrderingContextBuilder) GetContextDescription() string {
	return "Current highlight order and detailed content for reordering operations"
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