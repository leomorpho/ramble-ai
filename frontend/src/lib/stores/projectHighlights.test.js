import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import {
  rawHighlights,
  highlightOrder,
  orderedHighlights,
  insertNewLine,
  removeNewLine,
  updateHighlightOrder,
  currentProjectId
} from './projectHighlights.js';
import { 
  UpdateProjectHighlightOrderWithTitles, 
  GetProjectHighlightOrder,
  UpdateProjectHighlightOrderWithTitlesWithTitles,
  SaveSectionTitle,
  GetProjectHighlightOrderWithTitles
} from '$lib/wailsjs/go/main/App';

// Mock all the Wails functions that are used in the store
vi.mock('$lib/wailsjs/go/main/App', () => ({
  GetProjectHighlights: vi.fn(),
  GetProjectHighlightOrder: vi.fn(),
  UpdateProjectHighlightOrderWithTitles: vi.fn(),
  UpdateProjectHighlightOrderWithTitlesWithTitles: vi.fn(),
  DeleteHighlight: vi.fn(),
  UpdateVideoClipHighlights: vi.fn(),
  UndoOrderChange: vi.fn(),
  RedoOrderChange: vi.fn(),
  GetOrderHistoryStatus: vi.fn(),
  UndoHighlightsChange: vi.fn(),
  RedoHighlightsChange: vi.fn(),
  GetHighlightsHistoryStatus: vi.fn(),
  SaveSectionTitle: vi.fn(),
  GetProjectHighlightOrderWithTitles: vi.fn()
}));

describe('ProjectHighlights Store - Newline Functionality', () => {
  // Sample highlight data for testing
  const mockHighlights = [
    {
      id: 'highlight-1',
      videoClipId: 1,
      videoClipName: 'Video 1',
      filePath: '/path/to/video1.mp4',
      start: 0,
      end: 10,
      color: 'var(--highlight-1)',
      text: 'First highlight'
    },
    {
      id: 'highlight-2',
      videoClipId: 1,
      videoClipName: 'Video 1',
      filePath: '/path/to/video1.mp4',
      start: 15,
      end: 25,
      color: 'var(--highlight-2)',
      text: 'Second highlight'
    },
    {
      id: 'highlight-3',
      videoClipId: 2,
      videoClipName: 'Video 2',
      filePath: '/path/to/video2.mp4',
      start: 5,
      end: 15,
      color: 'var(--highlight-3)',
      text: 'Third highlight'
    }
  ];

  beforeEach(() => {
    // Reset stores to initial state
    rawHighlights.set([]);
    highlightOrder.set([]);
    currentProjectId.set('test-project-1');
    
    // Clear all mocks
    vi.clearAllMocks();
  });

  afterEach(() => {
    // Clean up after each test
    rawHighlights.set([]);
    highlightOrder.set([]);
    currentProjectId.set(null);
  });

  describe('orderedHighlights derived store', () => {
    it('should return highlights in default order when no custom order exists', () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set([]);
      
      const ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(3);
      // Should be sorted by videoClipId then by start time
      expect(ordered[0].id).toBe('highlight-1');
      expect(ordered[1].id).toBe('highlight-2');
      expect(ordered[2].id).toBe('highlight-3');
    });

    it('should apply custom order with highlights only', () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-3', 'highlight-1', 'highlight-2']);
      
      const ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(3);
      expect(ordered[0].id).toBe('highlight-3');
      expect(ordered[1].id).toBe('highlight-1');
      expect(ordered[2].id).toBe('highlight-2');
    });

    it('should include newlines in the correct positions', () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
      
      const ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(5);
      expect(ordered[0].id).toBe('highlight-1');
      expect(ordered[1].type).toBe('newline');
      expect(ordered[2].id).toBe('highlight-2');
      expect(ordered[3].type).toBe('newline');
      expect(ordered[4].id).toBe('highlight-3');
    });

    it('should handle mixed order with newlines at different positions', () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['N', 'highlight-2', 'highlight-1', 'N', 'N', 'highlight-3']);
      
      const ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(6);
      expect(ordered[0].type).toBe('newline');
      expect(ordered[1].id).toBe('highlight-2');
      expect(ordered[2].id).toBe('highlight-1');
      expect(ordered[3].type).toBe('newline');
      expect(ordered[4].type).toBe('newline');
      expect(ordered[5].id).toBe('highlight-3');
    });
  });

  describe('insertNewLine', () => {
    beforeEach(() => {
      // Mock the API call to succeed
      UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    });

    it('should insert newline at the beginning of empty timeline', async () => {
      rawHighlights.set([]);
      highlightOrder.set([]);
      
      const result = await insertNewLine(0);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['N']);
    });

    it('should insert newline at the beginning of timeline with highlights', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      const result = await insertNewLine(0);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should insert newline between highlights correctly', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // Insert newline at position 2 (between highlight-2 and highlight-3 in visual timeline)
      const result = await insertNewLine(2);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'N', 'highlight-3']);
    });

    it('should insert newline at the end of timeline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      const result = await insertNewLine(3);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3', 'N']);
    });

    it('should handle insertion when newlines already exist', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
      
      // The visual timeline is: [highlight-1, newline, highlight-2, highlight-3]
      // Inserting at position 3 (between highlight-2 and highlight-3)
      const result = await insertNewLine(3);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
    });

    it('should convert visual positions correctly with multiple existing newlines', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['N', 'highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
      
      // Visual timeline: [newline, highlight-1, newline, highlight-2, newline, highlight-3]
      // Inserting at position 2 (between highlight-1 and existing newline)
      const result = await insertNewLine(2);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
    });

    it('should handle edge case: insert after last highlight when timeline ends with newline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'N']);
      
      // Visual timeline: [highlight-1, highlight-2, newline]
      // Inserting at position 2 (between highlight-2 and existing newline)
      const result = await insertNewLine(2);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'N']);
    });

    it('should handle API failure gracefully', async () => {
      UpdateProjectHighlightOrderWithTitles.mockRejectedValue(new Error('API Error'));
      GetProjectHighlightOrder.mockResolvedValue(['highlight-1', 'highlight-2', 'highlight-3']);
      
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      const result = await insertNewLine(1);
      
      expect(result).toBe(false);
      // Order should be reverted on failure
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
    });
  });

  describe('removeNewLine', () => {
    beforeEach(() => {
      // Mock the API call to succeed
      UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    });

    it('should remove newline from the beginning of timeline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['N', 'highlight-1', 'highlight-2', 'highlight-3']);
      
      const result = await removeNewLine(0);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should remove newline from the middle of timeline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
      
      // Visual timeline: [highlight-1, newline, highlight-2, highlight-3]
      // Removing newline at position 1
      const result = await removeNewLine(1);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should remove newline from the end of timeline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3', 'N']);
      
      // Visual timeline: [highlight-1, highlight-2, highlight-3, newline]
      // Removing newline at position 3
      const result = await removeNewLine(3);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should handle multiple newlines correctly', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'N', 'highlight-2', 'N', 'highlight-3']);
      
      // Visual timeline: [highlight-1, newline, newline, highlight-2, newline, highlight-3]
      // Removing second newline (position 2)
      const result = await removeNewLine(2);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
    });

    it('should handle consecutive newlines at different positions', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['N', 'N', 'highlight-1', 'highlight-2', 'N', 'N', 'highlight-3']);
      
      // Visual timeline: [newline, newline, highlight-1, highlight-2, newline, newline, highlight-3]
      // Removing first newline (position 0)
      const result = await removeNewLine(0);
      
      expect(result).toBe(true);
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'highlight-2', 'N', 'highlight-3']);
    });

    it('should return false when trying to remove non-existent newline', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // No newlines in timeline, trying to remove at position 1
      const result = await removeNewLine(1);
      
      expect(result).toBe(false);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should return false when trying to remove from invalid position', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'highlight-2']);
      
      // Trying to remove from position that doesn't exist
      const result = await removeNewLine(5);
      
      expect(result).toBe(false);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2']);
    });

    it('should handle API failure gracefully', async () => {
      UpdateProjectHighlightOrderWithTitles.mockRejectedValue(new Error('API Error'));
      GetProjectHighlightOrder.mockResolvedValue(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
      
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
      
      const result = await removeNewLine(1);
      
      expect(result).toBe(false);
      // Order should be reverted on failure
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
    });
  });

  describe('Integration tests - insertNewLine and removeNewLine', () => {
    beforeEach(() => {
      // Mock the API call to succeed
      UpdateProjectHighlightOrderWithTitles.mockResolvedValue();
    });

    it('should maintain correct positions through multiple insert/remove operations', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // Insert newline at position 1 (between highlight-1 and highlight-2)
      await insertNewLine(1);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2', 'highlight-3']);
      
      // Insert another newline at position 3 (between highlight-2 and highlight-3)
      await insertNewLine(3);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'N', 'highlight-2', 'N', 'highlight-3']);
      
      // Remove the first newline (position 1)
      await removeNewLine(1);
      expect(get(highlightOrder)).toEqual(['highlight-1', 'highlight-2', 'N', 'highlight-3']);
      
      // Insert newline at the beginning (position 0)
      await insertNewLine(0);
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'highlight-2', 'N', 'highlight-3']);
      
      // Remove the last newline (position 3)
      await removeNewLine(3);
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'highlight-2', 'highlight-3']);
    });

    it('should handle complex timeline scenarios', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set([]);
      
      // Build a complex timeline step by step
      // Start with highlights: [highlight-1, highlight-2, highlight-3]
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // Add newlines to create: [newline, highlight-1, newline, highlight-2, newline, highlight-3, newline]
      await insertNewLine(0); // Insert at beginning
      await insertNewLine(2); // Insert after highlight-1
      await insertNewLine(4); // Insert after highlight-2  
      await insertNewLine(6); // Insert at end
      
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'N', 'highlight-2', 'N', 'highlight-3', 'N']);
      
      // Remove newlines from end to beginning (so positions don't shift)
      await removeNewLine(6); // Remove last newline
      await removeNewLine(4); // Remove newline after highlight-2
      
      expect(get(highlightOrder)).toEqual(['N', 'highlight-1', 'N', 'highlight-2', 'highlight-3']);
    });

    it('should verify orderedHighlights reflects changes correctly', async () => {
      rawHighlights.set(mockHighlights);
      highlightOrder.set(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // Initial state
      let ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(3);
      expect(ordered.map(h => h.id || h.type)).toEqual(['highlight-1', 'highlight-2', 'highlight-3']);
      
      // Insert newline
      await insertNewLine(1);
      ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(4);
      expect(ordered.map(h => h.type || h.id)).toEqual(['highlight-1', 'newline', 'highlight-2', 'highlight-3']);
      
      // Insert another newline at position 0 (this creates two non-consecutive newlines)
      await insertNewLine(0);
      ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(5); // Length increases because newlines are not consecutive
      expect(ordered.map(h => h.type || h.id)).toEqual(['newline', 'highlight-1', 'newline', 'highlight-2', 'highlight-3']);
      
      // Remove a newline (should remove the one at position 0)
      await removeNewLine(0);
      ordered = get(orderedHighlights);
      expect(ordered).toHaveLength(4);
      expect(ordered.map(h => h.type || h.id)).toEqual(['highlight-1', 'newline', 'highlight-2', 'highlight-3']);
    });
  });
});