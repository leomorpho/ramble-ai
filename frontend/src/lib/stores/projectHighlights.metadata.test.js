import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the Wails functions
vi.mock('$lib/wailsjs/go/main/App', () => ({
  UpdateProjectHighlightOrderWithTitles: vi.fn(),
  GetProjectHighlights: vi.fn(),
  GetProjectHighlightOrder: vi.fn(),
  GetProjectHighlightOrderWithTitles: vi.fn(),
  GetOrderHistoryStatus: vi.fn(),
}));

// Mock toast to prevent errors
vi.mock('svelte-sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { UpdateProjectHighlightOrderWithTitles, GetOrderHistoryStatus } from '$lib/wailsjs/go/main/App';
import { updateHighlightOrder, currentProjectId } from './projectHighlights.js';

describe('Project Highlights Store - Metadata Preservation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Set up a mock project ID
    currentProjectId.set(1);
  });

  it('should preserve database format section objects exactly in updateHighlightOrder', async () => {
    // Mock the backend functions to succeed
    UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    GetOrderHistoryStatus.mockResolvedValue({ canUndo: false, canRedo: false });

    // Test data with database format section objects (exactly as they come from AI suggestions)
    const orderWithSections = [
      { "title": "Personal Impact", "type": "N" },
      "highlight_1752086557450_lr1swkjaj",
      { "title": "Toxic Shame", "type": "N" },
      "highlight_1752086566479_yioxdt6gz",
      "highlight_1752086567716_54jz3puhk"
    ];

    // Call the store function directly
    const result = await updateHighlightOrder(orderWithSections);

    // Verify it succeeded
    expect(result).toBe(true);

    // Verify the backend was called with the EXACT data (no conversion)
    expect(UpdateProjectHighlightOrderWithTitles).toHaveBeenCalledWith(
      1, // project ID
      orderWithSections // Should be exactly the same, no conversion
    );

    // Verify section objects were preserved exactly
    const calledData = UpdateProjectHighlightOrderWithTitles.mock.calls[0][1];
    expect(calledData[0]).toEqual({ "title": "Personal Impact", "type": "N" });
    expect(calledData[2]).toEqual({ "title": "Toxic Shame", "type": "N" });
    expect(calledData[1]).toBe("highlight_1752086557450_lr1swkjaj");
    expect(calledData[3]).toBe("highlight_1752086566479_yioxdt6gz");
    
    // Verify no null values are present (this was the bug)
    expect(calledData).not.toContain(null);
    expect(calledData).not.toContain(undefined);
  });

  it('should handle mixed simple N markers and section objects', async () => {
    UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    GetOrderHistoryStatus.mockResolvedValue({ canUndo: false, canRedo: false });

    // Test data with mix of simple "N" and section objects
    const orderWithMixed = [
      "highlight_1",
      "N", // Simple newline marker
      "highlight_2",
      { "title": "Important Section", "type": "N" }, // Section with title
      "highlight_3"
    ];

    const result = await updateHighlightOrder(orderWithMixed);

    expect(result).toBe(true);

    const calledData = UpdateProjectHighlightOrderWithTitles.mock.calls[0][1];
    expect(calledData[1]).toBe("N"); // Simple marker preserved
    expect(calledData[3]).toEqual({ "title": "Important Section", "type": "N" }); // Object preserved
    expect(calledData).not.toContain(null);
  });

  it('should handle future metadata types without breaking', async () => {
    UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    GetOrderHistoryStatus.mockResolvedValue({ canUndo: false, canRedo: false });

    // Test data with hypothetical future metadata types
    const orderWithFutureMetadata = [
      { "title": "Chapter 1", "type": "N", "icon": "🎬", "color": "#ff0000" },
      "highlight_1",
      { "type": "BOOKMARK", "id": "bookmark_1", "label": "Important Moment" },
      "highlight_2",
      { "type": "TRANSITION", "effect": "fade", "duration": 1000 },
      "highlight_3",
      { "title": "Conclusion", "type": "N", "metadata": { "priority": "high", "tags": ["ending"] }}
    ];

    const result = await updateHighlightOrder(orderWithFutureMetadata);

    expect(result).toBe(true);

    // Verify the backend was called with the EXACT data (no conversion or loss)
    const calledData = UpdateProjectHighlightOrderWithTitles.mock.calls[0][1];
    
    // All metadata objects should be preserved exactly as-is
    expect(calledData[0]).toEqual({ "title": "Chapter 1", "type": "N", "icon": "🎬", "color": "#ff0000" });
    expect(calledData[2]).toEqual({ "type": "BOOKMARK", "id": "bookmark_1", "label": "Important Moment" });
    expect(calledData[4]).toEqual({ "type": "TRANSITION", "effect": "fade", "duration": 1000 });
    expect(calledData[6]).toEqual({ "title": "Conclusion", "type": "N", "metadata": { "priority": "high", "tags": ["ending"] }});
    
    // Highlight IDs should be preserved
    expect(calledData[1]).toBe("highlight_1");
    expect(calledData[3]).toBe("highlight_2");
    expect(calledData[5]).toBe("highlight_3");
    
    // No data should be lost or converted to null
    expect(calledData).not.toContain(null);
    expect(calledData).not.toContain(undefined);
    expect(calledData.length).toBe(orderWithFutureMetadata.length);
  });

  it('should handle display format newlines (type: newline) correctly', async () => {
    UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    GetOrderHistoryStatus.mockResolvedValue({ canUndo: false, canRedo: false });

    // Test data with display format newlines (as they appear in UI)
    const orderWithDisplayFormat = [
      "highlight_1",
      { id: "newline_123", type: "newline", title: "My Section" }, // Display format
      "highlight_2"
    ];

    const result = await updateHighlightOrder(orderWithDisplayFormat);

    expect(result).toBe(true);

    const calledData = UpdateProjectHighlightOrderWithTitles.mock.calls[0][1];
    expect(calledData[0]).toBe("highlight_1");
    expect(calledData[1]).toEqual({ "type": "N", "title": "My Section" }); // Converted to database format
    expect(calledData[2]).toBe("highlight_2");
  });

  it('should handle objects with special characters in titles', async () => {
    UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    GetOrderHistoryStatus.mockResolvedValue({ canUndo: false, canRedo: false });

    // Test with special characters that might break JSON parsing
    const orderWithSpecialChars = [
      { "title": "Exposure & Improvement", "type": "N" },
      "highlight_1",
      { "title": "Q&A Section - User's \"Questions\"", "type": "N" },
      "highlight_2",
      { "title": "Final Thoughts (100% Complete)", "type": "N" },
      "highlight_3"
    ];

    const result = await updateHighlightOrder(orderWithSpecialChars);

    expect(result).toBe(true);

    const calledData = UpdateProjectHighlightOrderWithTitles.mock.calls[0][1];
    expect(calledData[0].title).toBe("Exposure & Improvement");
    expect(calledData[2].title).toBe('Q&A Section - User\'s "Questions"');
    expect(calledData[4].title).toBe("Final Thoughts (100% Complete)");
  });
});