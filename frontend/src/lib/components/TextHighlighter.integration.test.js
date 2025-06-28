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
      expect(result1.newHighlight.color).toBe('#ffeb3b');
      
      // 2. Update used colors
      usedColors.add(result1.newHighlight.color);
      
      // 3. Create second highlight with different color
      const result2 = addHighlight(result1.highlights, 4, 5, usedColors);
      expect(result2.highlights).toHaveLength(2);
      expect(result2.newHighlight.color).toBe('#81c784'); // Next color
      
      // 4. Check grouping works correctly
      const groups = groupWordsAndHighlights(sampleWords, result2.highlights);
      expect(groups).toHaveLength(5); // word 0, highlight 1-2, word 3, highlight 4-5, word 6
      
      // 5. Verify no overlap detection
      expect(checkOverlap(1, 3, result2.highlights)).toBe(true); // should overlap with first highlight
      expect(checkOverlap(0, 0, result2.highlights)).toBe(false); // should not overlap
    });

    it('should handle highlight modification workflow', () => {
      let highlights = [
        { id: 'h1', start: 1, end: 3, color: '#ffeb3b' },
        { id: 'h2', start: 5, end: 6, color: '#81c784' }
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
      const baseColors = ['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292'];
      baseColors.forEach(color => {
        const generatedColor = generateUniqueColor(usedColors);
        expect(generatedColor).toBe(color);
        usedColors.add(generatedColor);
      });
      
      // 2. Generate HSL colors when base colors exhausted
      const hslColor1 = generateUniqueColor(usedColors);
      const hslColor2 = generateUniqueColor(usedColors);
      
      expect(hslColor1).toMatch(/^hsl\(\d+, [\d.]+%, [\d.]+%\)$/);
      expect(hslColor2).toMatch(/^hsl\(\d+, [\d.]+%, [\d.]+%\)$/);
      
      // 3. Color recycling simulation
      usedColors.delete('#ffeb3b');
      const recycledColor = generateUniqueColor(usedColors);
      expect(recycledColor).toBe('#ffeb3b');
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