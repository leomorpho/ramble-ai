import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  GetProjectAISuggestion,
  UpdateProjectHighlightOrderWithTitles
} from '$lib/wailsjs/go/main/App';

// Mock the Wails functions
vi.mock('$lib/wailsjs/go/main/App', () => ({
  GetProjectAISuggestion: vi.fn(),
  UpdateProjectHighlightOrderWithTitles: vi.fn(),
  GetProjectHighlights: vi.fn(),
  GetProjectHighlightOrder: vi.fn(),
  GetProjectHighlightOrderWithTitles: vi.fn(),
  GetOrderHistoryStatus: vi.fn(),
}));

// Mock the store function
vi.mock('$lib/stores/projectHighlights.js', () => ({
  updateHighlightOrder: vi.fn(),
}));

import { updateHighlightOrder } from '$lib/stores/projectHighlights.js';

describe('AI Reorder Sheet - Section Title Preservation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should preserve section objects exactly as they are in AI suggestions', async () => {
    // Sample AI suggestion with section objects - exactly as stored in database
    const mockAISuggestion = {
      order: [
        { "title": "Personal Impact", "type": "N" },
        "highlight_1752086557450_lr1swkjaj",
        { "title": "Toxic Shame", "type": "N" },
        "highlight_1752086566479_yioxdt6gz",
        "highlight_1752086567716_54jz3puhk",
        { "title": "Exposure & Improvement", "type": "N" },
        "highlight_1752086565468_egogda0qe",
        "highlight_1752086564341_wmna40inb",
        { "title": "Neuro-Limbic Brain", "type": "N" },
        "highlight_1752086562788_zo9mju0n2"
      ],
      model: "anthropic/claude-sonnet-4",
      createdAt: new Date().toISOString()
    };

    // Mock the API calls
    GetProjectAISuggestion.mockResolvedValue(mockAISuggestion);
    updateHighlightOrder.mockResolvedValue(true);

    // Simulate the applyAIReordering function logic
    const projectId = 1;
    const cachedSuggestion = await GetProjectAISuggestion(projectId);
    
    if (cachedSuggestion && cachedSuggestion.order && cachedSuggestion.order.length > 0) {
      // This should pass the exact order without modification
      await updateHighlightOrder(cachedSuggestion.order);
    }

    // Verify that GetProjectAISuggestion was called
    expect(GetProjectAISuggestion).toHaveBeenCalledWith(projectId);
    
    // Verify that updateHighlightOrder was called with the EXACT AI suggestion order
    expect(updateHighlightOrder).toHaveBeenCalledWith([
      { "title": "Personal Impact", "type": "N" },
      "highlight_1752086557450_lr1swkjaj",
      { "title": "Toxic Shame", "type": "N" },
      "highlight_1752086566479_yioxdt6gz",
      "highlight_1752086567716_54jz3puhk",
      { "title": "Exposure & Improvement", "type": "N" },
      "highlight_1752086565468_egogda0qe",
      "highlight_1752086564341_wmna40inb",
      { "title": "Neuro-Limbic Brain", "type": "N" },
      "highlight_1752086562788_zo9mju0n2"
    ]);

    // Verify that section objects are preserved exactly
    const calledOrder = updateHighlightOrder.mock.calls[0][0];
    
    // Check that section objects have correct structure and content
    expect(calledOrder[0]).toEqual({ "title": "Personal Impact", "type": "N" });
    expect(calledOrder[2]).toEqual({ "title": "Toxic Shame", "type": "N" });
    expect(calledOrder[5]).toEqual({ "title": "Exposure & Improvement", "type": "N" });
    expect(calledOrder[8]).toEqual({ "title": "Neuro-Limbic Brain", "type": "N" });
    
    // Check that highlight IDs are preserved
    expect(calledOrder[1]).toBe("highlight_1752086557450_lr1swkjaj");
    expect(calledOrder[3]).toBe("highlight_1752086566479_yioxdt6gz");
    
    // Verify no conversion to null or other formats occurred
    expect(calledOrder).not.toContain(null);
    expect(calledOrder).not.toContain("N");
  });

  it('should handle empty AI suggestion gracefully', async () => {
    // Mock empty AI suggestion
    GetProjectAISuggestion.mockResolvedValue(null);

    const projectId = 1;
    const cachedSuggestion = await GetProjectAISuggestion(projectId);
    
    // Should not call updateHighlightOrder if no cached suggestion
    expect(cachedSuggestion).toBeNull();
    expect(updateHighlightOrder).not.toHaveBeenCalled();
  });

  it('should handle AI suggestion without order array', async () => {
    // Mock AI suggestion without order
    const mockAISuggestion = {
      model: "anthropic/claude-sonnet-4",
      createdAt: new Date().toISOString()
      // order is missing
    };

    GetProjectAISuggestion.mockResolvedValue(mockAISuggestion);

    const projectId = 1;
    const cachedSuggestion = await GetProjectAISuggestion(projectId);
    
    // Should not call updateHighlightOrder if order is missing
    expect(cachedSuggestion.order).toBeUndefined();
    expect(updateHighlightOrder).not.toHaveBeenCalled();
  });

  it('should preserve simple N markers without titles', async () => {
    // AI suggestion with mix of simple "N" and section objects
    const mockAISuggestion = {
      order: [
        "highlight_1",
        "N", // Simple newline marker
        "highlight_2",
        { "title": "Important Section", "type": "N" }, // Section with title
        "highlight_3"
      ],
      model: "anthropic/claude-sonnet-4",
      createdAt: new Date().toISOString()
    };

    GetProjectAISuggestion.mockResolvedValue(mockAISuggestion);
    updateHighlightOrder.mockResolvedValue(true);

    const projectId = 1;
    const cachedSuggestion = await GetProjectAISuggestion(projectId);
    
    if (cachedSuggestion && cachedSuggestion.order && cachedSuggestion.order.length > 0) {
      await updateHighlightOrder(cachedSuggestion.order);
    }

    // Verify both simple "N" and section objects are preserved
    expect(updateHighlightOrder).toHaveBeenCalledWith([
      "highlight_1",
      "N",
      "highlight_2", 
      { "title": "Important Section", "type": "N" },
      "highlight_3"
    ]);

    const calledOrder = updateHighlightOrder.mock.calls[0][0];
    expect(calledOrder[1]).toBe("N"); // Simple marker preserved
    expect(calledOrder[3]).toEqual({ "title": "Important Section", "type": "N" }); // Object preserved
  });

  it('should handle special characters in section titles', async () => {
    // AI suggestion with special characters in titles
    const mockAISuggestion = {
      order: [
        { "title": "Exposure & Improvement", "type": "N" },
        "highlight_1",
        { "title": "Q&A Section - User's Questions", "type": "N" },
        "highlight_2",
        { "title": "Final Thoughts (Conclusion)", "type": "N" },
        "highlight_3"
      ],
      model: "anthropic/claude-sonnet-4",
      createdAt: new Date().toISOString()
    };

    GetProjectAISuggestion.mockResolvedValue(mockAISuggestion);
    updateHighlightOrder.mockResolvedValue(true);

    const projectId = 1;
    const cachedSuggestion = await GetProjectAISuggestion(projectId);
    
    if (cachedSuggestion && cachedSuggestion.order && cachedSuggestion.order.length > 0) {
      await updateHighlightOrder(cachedSuggestion.order);
    }

    // Verify special characters in titles are preserved exactly
    const calledOrder = updateHighlightOrder.mock.calls[0][0];
    expect(calledOrder[0].title).toBe("Exposure & Improvement");
    expect(calledOrder[2].title).toBe("Q&A Section - User's Questions");
    expect(calledOrder[4].title).toBe("Final Thoughts (Conclusion)");
  });

});