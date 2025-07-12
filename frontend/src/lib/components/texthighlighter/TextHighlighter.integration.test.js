import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  generateUniqueColor,
  createHighlight,
  isWordInHighlight,
  isWordInSelection,
  findHighlightForWord,
  checkOverlap,
  calculateTimestamps,
  findWordByTimestamp,
  groupWordsAndHighlights,
  updateHighlight,
  addHighlight,
  removeHighlight
} from './TextHighlighter.utils.js';

describe('TextHighlighter Integration Tests', () => {
  const sampleWords = [
    { word: 'Hello', start: 0.0, end: 0.5 },
    { word: 'beautiful', start: 0.6, end: 1.2 },
    { word: 'world', start: 1.3, end: 1.8 },
    { word: 'this', start: 2.0, end: 2.3 },
    { word: 'is', start: 2.4, end: 2.6 },
    { word: 'a', start: 2.7, end: 2.8 },
    { word: 'test', start: 2.9, end: 3.3 }
  ];

  describe('Complete workflow integration', () => {
    it('should handle complete highlight creation workflow', () => {
      const usedColors = new Set();
      const highlights = [];
      
      // 1. Create first highlight
      const result1 = addHighlight(highlights, 1, 2, usedColors);
      expect(result1.highlights).toHaveLength(1);
      expect(result1.newHighlight.start).toBe(1);
      expect(result1.newHighlight.end).toBe(2);
      expect(result1.newHighlight.color).toBe('var(--highlight-1)');
      
      // 2. Update used colors
      usedColors.add(result1.newHighlight.color);
      
      // 3. Create second highlight with different color
      const result2 = addHighlight(result1.highlights, 4, 5, usedColors);
      expect(result2.highlights).toHaveLength(2);
      expect(result2.newHighlight.color).toBe('var(--highlight-2)'); // Next color
      
      // 4. Check grouping works correctly
      const groups = groupWordsAndHighlights(sampleWords, result2.highlights);
      expect(groups).toHaveLength(5); // word 0, highlight 1-2, word 3, highlight 4-5, word 6
      
      // 5. Verify no overlap detection
      expect(checkOverlap(1, 3, result2.highlights)).toBe(true); // should overlap with first highlight
      expect(checkOverlap(0, 0, result2.highlights)).toBe(false); // should not overlap
    });

    it('should handle highlight modification workflow', () => {
      let highlights = [
        { id: 'h1', start: 1, end: 3, color: 'var(--highlight-1)' },
        { id: 'h2', start: 5, end: 6, color: 'var(--highlight-2)' }
      ];
      
      // 1. Update highlight position
      highlights = updateHighlight(highlights, 'h1', 0, 2);
      expect(highlights[0]).toMatchObject({ id: 'h1', start: 0, end: 2 });
      
      // 2. Check for overlaps after update
      expect(checkOverlap(1, 4, highlights, 'h1')).toBe(false); // excluding h1
      expect(checkOverlap(1, 4, highlights)).toBe(true); // including h1
      
      // 3. Remove highlight
      highlights = removeHighlight(highlights, 'h2');
      expect(highlights).toHaveLength(1);
      expect(highlights.find(h => h.id === 'h2')).toBeUndefined();
      
      // 4. Verify grouping after removal
      const groups = groupWordsAndHighlights(sampleWords, highlights);
      const highlightGroups = groups.filter(g => g.type === 'highlight');
      expect(highlightGroups).toHaveLength(1);
    });

    it('should handle timestamp conversion workflow', () => {
      // 1. Convert word indices to timestamps
      const timestamps = calculateTimestamps(1, 3, sampleWords);
      expect(timestamps).toEqual({ start: 0.6, end: 2.3 });
      
      // 2. Convert timestamps back to word indices
      const startIndex = findWordByTimestamp(timestamps.start, sampleWords);
      const endIndex = findWordByTimestamp(timestamps.end, sampleWords);
      expect(startIndex).toBe(1);
      expect(endIndex).toBe(3);
      
      // 3. Verify round-trip conversion
      const roundTripTimestamps = calculateTimestamps(startIndex, endIndex, sampleWords);
      expect(roundTripTimestamps).toEqual(timestamps);
    });

    it('should handle edge cases in selection workflow', () => {
      // 1. Single word selection
      const isSelecting = true;
      expect(isWordInSelection(2, 2, 2, isSelecting)).toBe(true);
      expect(isWordInSelection(1, 2, 2, isSelecting)).toBe(false);
      expect(isWordInSelection(3, 2, 2, isSelecting)).toBe(false);
      
      // 2. Reversed selection
      expect(isWordInSelection(2, 3, 1, isSelecting)).toBe(true);
      expect(isWordInSelection(0, 3, 1, isSelecting)).toBe(false);
      expect(isWordInSelection(4, 3, 1, isSelecting)).toBe(false);
      
      // 3. Invalid selection states
      expect(isWordInSelection(2, null, 3, isSelecting)).toBe(false);
      expect(isWordInSelection(2, 1, null, isSelecting)).toBe(false);
      expect(isWordInSelection(2, 1, 3, false)).toBe(false);
    });

    it('should handle complex grouping scenarios', () => {
      const highlights = [
        { id: 'h1', start: 0, end: 1 },   // "Hello beautiful"
        { id: 'h2', start: 2, end: 2 },   // "world" (single word)
        { id: 'h3', start: 4, end: 6 }    // "is a test"
      ];
      
      const groups = groupWordsAndHighlights(sampleWords, highlights);
      
      // Should have 4 groups: highlight(0-1), word(3), highlight(2), highlight(4-6)
      expect(groups).toHaveLength(4);
      
      // First group: highlight with 2 words
      expect(groups[0]).toMatchObject({
        type: 'highlight',
        highlight: highlights[0],
        startIndex: 0,
        words: [
          { word: sampleWords[0], index: 0 },
          { word: sampleWords[1], index: 1 }
        ]
      });
      
      // Second group: single word highlight
      expect(groups[1]).toMatchObject({
        type: 'highlight',
        highlight: highlights[1],
        startIndex: 2,
        words: [{ word: sampleWords[2], index: 2 }]
      });
      
      // Third group: regular word
      expect(groups[2]).toMatchObject({
        type: 'word',
        word: sampleWords[3],
        index: 3
      });
      
      // Fourth group: highlight with 3 words
      expect(groups[3]).toMatchObject({
        type: 'highlight',
        highlight: highlights[2],
        startIndex: 4,
        words: [
          { word: sampleWords[4], index: 4 },
          { word: sampleWords[5], index: 5 },
          { word: sampleWords[6], index: 6 }
        ]
      });
    });

    it('should handle color management workflow', () => {
      let usedColors = new Set();
      
      // 1. Use all base colors
      const baseColors = ['var(--highlight-1)', 'var(--highlight-2)', 'var(--highlight-3)', 'var(--highlight-4)', 'var(--highlight-5)'];
      baseColors.forEach(color => {
        const generatedColor = generateUniqueColor(usedColors);
        expect(generatedColor).toBe(color);
        usedColors.add(generatedColor);
      });
      
      // 2. Generate extended colors when base colors exhausted
      const extendedColor1 = generateUniqueColor(usedColors);
      usedColors.add(extendedColor1);
      const extendedColor2 = generateUniqueColor(usedColors);
      
      expect(extendedColor1).toBe('var(--highlight-6)');
      expect(extendedColor2).toBe('var(--highlight-7)');
      
      // 3. Color recycling simulation
      usedColors.delete('var(--highlight-1)');
      const recycledColor = generateUniqueColor(usedColors);
      expect(recycledColor).toBe('var(--highlight-1)');
    });
  });

  describe('Error handling and boundary conditions', () => {
    it('should handle empty inputs gracefully', () => {
      expect(groupWordsAndHighlights([], [])).toEqual([]);
      expect(calculateTimestamps(0, 0, [])).toEqual({ start: 0, end: 0 });
      expect(findWordByTimestamp(1.0, [])).toBe(-1);
      expect(removeHighlight([], 'any')).toEqual([]);
    });

    it('should handle invalid indices', () => {
      const highlights = [{ id: 'h1', start: 1, end: 3, color: 'red' }];
      
      // Test with out-of-bounds indices
      const timestamps = calculateTimestamps(-5, 100, sampleWords);
      expect(timestamps.start).toBe(0.0); // clamped to first word
      expect(timestamps.end).toBe(3.3);   // clamped to last word
      
      // Test with negative word index
      expect(isWordInHighlight(-1, highlights[0])).toBe(false);
      expect(isWordInHighlight(100, highlights[0])).toBe(false);
    });

    it('should handle malformed word data', () => {
      const malformedWords = [
        { word: 'test1' }, // missing timestamps
        { word: 'test2', start: 1.0 }, // missing end
        { word: 'test3', end: 2.0 }, // missing start
        { word: 'test4', start: 2.0, end: 2.5 } // complete
      ];
      
      const timestamps = calculateTimestamps(0, 3, malformedWords);
      expect(timestamps.start).toBe(0); // fallback for missing start
      expect(timestamps.end).toBe(2.5); // uses valid end
    });

    it('should handle overlapping highlights correctly', () => {
      const highlights = [
        { id: 'h1', start: 0, end: 3, color: 'red' },
        { id: 'h2', start: 2, end: 5, color: 'blue' },
        { id: 'h3', start: 6, end: 8, color: 'green' }
      ];
      
      // Test various overlap scenarios
      expect(checkOverlap(1, 4, highlights)).toBe(true);  // overlaps h1 and h2
      expect(checkOverlap(5, 7, highlights)).toBe(true);  // overlaps h2 and h3
      expect(checkOverlap(9, 10, highlights)).toBe(false); // no overlap
      expect(checkOverlap(2, 3, highlights, 'h1')).toBe(true); // still overlaps h2
      expect(checkOverlap(2, 3, highlights, 'h2')).toBe(true); // still overlaps h1
    });
  });

  describe('Highlight expansion functionality', () => {
    it('should expand highlight to the right correctly', () => {
      // Setup: highlight on words 1-2, then expand to include word 3
      const originalHighlight = { id: 'h1', start: 1, end: 2, color: '#ffeb3b' };
      const dragTarget = { highlightId: 'h1', originalStart: 1, originalEnd: 2 };
      
      // Test expansion preview: dragging from word 2 to word 3
      const selectionStart = 2; // last word of original highlight
      const selectionEnd = 3;   // new word to include
      
      // Only word 3 should show expansion preview (not already highlighted words)
      expect(isWordInExpansionPreview(0, dragTarget, selectionStart, selectionEnd)).toBe(false); // before range
      expect(isWordInExpansionPreview(1, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(2, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(3, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion
      expect(isWordInExpansionPreview(4, dragTarget, selectionStart, selectionEnd)).toBe(false); // after range
    });

    it('should expand highlight to the left correctly', () => {
      // Setup: highlight on words 2-3, then expand to include word 1
      const originalHighlight = { id: 'h1', start: 2, end: 3, color: '#ffeb3b' };
      const dragTarget = { highlightId: 'h1', originalStart: 2, originalEnd: 3 };
      
      // Test expansion preview: dragging from word 2 to word 1
      const selectionStart = 2; // first word of original highlight
      const selectionEnd = 1;   // new word to include (to the left)
      
      // Only word 1 should show expansion preview
      expect(isWordInExpansionPreview(0, dragTarget, selectionStart, selectionEnd)).toBe(false); // before range
      expect(isWordInExpansionPreview(1, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion
      expect(isWordInExpansionPreview(2, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(3, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(4, dragTarget, selectionStart, selectionEnd)).toBe(false); // after range
    });

    it('should expand highlight in both directions correctly', () => {
      // Setup: highlight on word 2, then expand to include words 1 and 3
      const originalHighlight = { id: 'h1', start: 2, end: 2, color: '#ffeb3b' };
      const dragTarget = { highlightId: 'h1', originalStart: 2, originalEnd: 2 };
      
      // Test expansion preview: dragging from word 2 to span words 1-3
      const selectionStart = 1; // expand left
      const selectionEnd = 3;   // expand right
      
      // Words 1 and 3 should show expansion preview, word 2 should not (already highlighted)
      expect(isWordInExpansionPreview(0, dragTarget, selectionStart, selectionEnd)).toBe(false); // before range
      expect(isWordInExpansionPreview(1, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion left
      expect(isWordInExpansionPreview(2, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(3, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion right
      expect(isWordInExpansionPreview(4, dragTarget, selectionStart, selectionEnd)).toBe(false); // after range
    });

    it('should not show expansion preview when not in drag expansion mode', () => {
      const dragTarget = { highlightId: 'h1', originalStart: 1, originalEnd: 2 };
      const selectionStart = 1;
      const selectionEnd = 3;
      
      // Without dragExpansion flag, should return false
      expect(isWordInExpansionPreview(3, null, selectionStart, selectionEnd)).toBe(false);
      expect(isWordInExpansionPreview(3, dragTarget, null, selectionEnd)).toBe(false);
      expect(isWordInExpansionPreview(3, dragTarget, selectionStart, null)).toBe(false);
    });

    it('should handle expansion with reversed selection correctly', () => {
      // Setup: highlight on words 2-3, selection from 4 back to 1
      const dragTarget = { highlightId: 'h1', originalStart: 2, originalEnd: 3 };
      const selectionStart = 4; // drag starts from right
      const selectionEnd = 1;   // drag ends on left
      
      // Words 1 and 4 should show expansion preview
      expect(isWordInExpansionPreview(1, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion
      expect(isWordInExpansionPreview(2, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(3, dragTarget, selectionStart, selectionEnd)).toBe(false); // in original
      expect(isWordInExpansionPreview(4, dragTarget, selectionStart, selectionEnd)).toBe(true);  // new expansion
    });

    it('should prevent expansion when it would cause overlap', () => {
      const highlights = [
        { id: 'h1', start: 2, end: 3, color: '#ffeb3b' },
        { id: 'h2', start: 5, end: 6, color: '#81c784' }
      ];
      
      // Try to expand h1 to overlap with h2
      const newStartTime = 2.0; // original start of h1
      const newEndTime = 5.5;   // would overlap with h2
      
      expect(checkOverlap(newStartTime, newEndTime, highlights, 'h1')).toBe(true);
      
      // But expansion that doesn't overlap should be allowed
      const safeEndTime = 4.9;
      expect(checkOverlap(newStartTime, safeEndTime, highlights, 'h1')).toBe(false);
    });
  });

  // Helper function for expansion preview tests
  function isWordInExpansionPreview(wordIndex, dragTarget, selectionStart, selectionEnd) {
    if (!dragTarget || selectionStart === null || selectionEnd === null) return false;
    
    // Get the original highlight bounds
    const originalStart = Math.min(dragTarget.originalStart, dragTarget.originalEnd);
    const originalEnd = Math.max(dragTarget.originalStart, dragTarget.originalEnd);
    
    // Get the current selection bounds
    const selStart = Math.min(selectionStart, selectionEnd);
    const selEnd = Math.max(selectionStart, selectionEnd);
    
    // Check if word is in the original highlight
    const inOriginal = wordIndex >= originalStart && wordIndex <= originalEnd;
    
    // Check if word is in the current selection
    const inSelection = wordIndex >= selStart && wordIndex <= selEnd;
    
    // Show expansion preview only for words in selection that are NOT in original highlight
    return inSelection && !inOriginal;
  }

  describe('Performance characteristics', () => {
    it('should handle large datasets efficiently', () => {
      // Create large dataset
      const largeWords = Array.from({ length: 1000 }, (_, i) => ({
        word: `word${i}`,
        start: i * 0.5,
        end: i * 0.5 + 0.4
      }));
      
      const largeHighlights = Array.from({ length: 100 }, (_, i) => ({
        id: `h${i}`,
        start: i * 10,
        end: i * 10 + 5,
        color: `color${i}`
      }));
      
      // Test grouping performance
      const start = performance.now();
      const groups = groupWordsAndHighlights(largeWords, largeHighlights);
      const duration = performance.now() - start;
      
      expect(duration).toBeLessThan(100); // Should complete in < 100ms
      expect(groups.length).toBeGreaterThan(0);
    });

    it('should handle rapid timestamp lookups', () => {
      const timestamps = Array.from({ length: 100 }, (_, i) => i * 0.1);
      
      const start = performance.now();
      timestamps.forEach(timestamp => {
        findWordByTimestamp(timestamp, sampleWords);
      });
      const duration = performance.now() - start;
      
      expect(duration).toBeLessThan(50); // Should complete in < 50ms
    });
  });
});