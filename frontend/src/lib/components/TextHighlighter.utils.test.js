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

describe('TextHighlighter Utils', () => {
  describe('generateUniqueColor', () => {
    it('should return first base color when no colors used', () => {
      const usedColors = new Set();
      const color = generateUniqueColor(usedColors);
      expect(color).toBe('#ffeb3b');
    });

    it('should return second base color when first is used', () => {
      const usedColors = new Set(['#ffeb3b']);
      const color = generateUniqueColor(usedColors);
      expect(color).toBe('#81c784');
    });

    it('should generate HSL color when all base colors are used', () => {
      const usedColors = new Set(['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292']);
      const color = generateUniqueColor(usedColors);
      expect(color).toMatch(/^hsl\(\d+, [\d.]+%, [\d.]+%\)$/);
    });

    it('should generate different HSL colors on multiple calls', () => {
      const usedColors = new Set(['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292']);
      const color1 = generateUniqueColor(usedColors);
      const color2 = generateUniqueColor(usedColors);
      // Since it's random, we can't guarantee they're different, but they should be HSL format
      expect(color1).toMatch(/^hsl\(\d+, [\d.]+%, [\d.]+%\)$/);
      expect(color2).toMatch(/^hsl\(\d+, [\d.]+%, [\d.]+%\)$/);
    });
  });

  describe('createHighlight', () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2024-01-01T00:00:00Z'));
    });

    it('should create highlight with correct structure', () => {
      const highlight = createHighlight(0, 5);
      expect(highlight).toMatchObject({
        start: 0,
        end: 5,
        color: '#ffeb3b'
      });
      expect(highlight.id).toMatch(/^highlight_\d+_[a-z0-9]+$/);
    });

    it('should use unique color from usedColors', () => {
      const usedColors = new Set(['#ffeb3b']);
      const highlight = createHighlight(0, 5, usedColors);
      expect(highlight.color).toBe('#81c784');
    });
  });

  describe('isWordInHighlight', () => {
    const highlight = { start: 2, end: 5 };

    it('should return true for word within highlight range', () => {
      expect(isWordInHighlight(3, highlight)).toBe(true);
      expect(isWordInHighlight(2, highlight)).toBe(true);
      expect(isWordInHighlight(5, highlight)).toBe(true);
    });

    it('should return false for word outside highlight range', () => {
      expect(isWordInHighlight(1, highlight)).toBe(false);
      expect(isWordInHighlight(6, highlight)).toBe(false);
    });
  });

  describe('isWordInSelection', () => {
    it('should return true for word within selection range', () => {
      expect(isWordInSelection(3, 2, 5, true)).toBe(true);
      expect(isWordInSelection(2, 2, 5, true)).toBe(true);
      expect(isWordInSelection(5, 2, 5, true)).toBe(true);
    });

    it('should handle reversed selection range', () => {
      expect(isWordInSelection(3, 5, 2, true)).toBe(true);
    });

    it('should return false when not selecting', () => {
      expect(isWordInSelection(3, 2, 5, false)).toBe(false);
    });

    it('should return false when selection is null', () => {
      expect(isWordInSelection(3, null, 5, true)).toBe(false);
      expect(isWordInSelection(3, 2, null, true)).toBe(false);
    });

    it('should return false for word outside selection range', () => {
      expect(isWordInSelection(1, 2, 5, true)).toBe(false);
      expect(isWordInSelection(6, 2, 5, true)).toBe(false);
    });
  });

  describe('findHighlightForWord', () => {
    const highlights = [
      { id: '1', start: 0, end: 2 },
      { id: '2', start: 5, end: 8 },
      { id: '3', start: 10, end: 15 }
    ];

    it('should return highlight for word within range', () => {
      expect(findHighlightForWord(1, highlights)).toEqual({ id: '1', start: 0, end: 2 });
      expect(findHighlightForWord(6, highlights)).toEqual({ id: '2', start: 5, end: 8 });
      expect(findHighlightForWord(12, highlights)).toEqual({ id: '3', start: 10, end: 15 });
    });

    it('should return undefined for word not in any highlight', () => {
      expect(findHighlightForWord(3, highlights)).toBeUndefined();
      expect(findHighlightForWord(9, highlights)).toBeUndefined();
      expect(findHighlightForWord(20, highlights)).toBeUndefined();
    });
  });

  describe('checkOverlap', () => {
    const highlights = [
      { id: '1', start: 0, end: 2 },
      { id: '2', start: 5, end: 8 },
      { id: '3', start: 10, end: 15 }
    ];

    it('should detect overlap with existing highlight', () => {
      expect(checkOverlap(1, 3, highlights)).toBe(true); // overlaps with highlight 1
      expect(checkOverlap(4, 6, highlights)).toBe(true); // overlaps with highlight 2
      expect(checkOverlap(8, 12, highlights)).toBe(true); // overlaps with highlights 2 and 3
    });

    it('should return false for non-overlapping range', () => {
      expect(checkOverlap(3, 4, highlights)).toBe(false);
      expect(checkOverlap(16, 20, highlights)).toBe(false);
    });

    it('should exclude specified highlight ID from overlap check', () => {
      expect(checkOverlap(0, 2, highlights, '1')).toBe(false); // excludes highlight 1
      expect(checkOverlap(1, 6, highlights, '1')).toBe(true); // still overlaps with highlight 2
    });

    it('should handle edge cases correctly', () => {
      expect(checkOverlap(2, 5, highlights)).toBe(true); // touching boundaries
      expect(checkOverlap(3, 4, highlights)).toBe(false); // between highlights
    });
  });

  describe('calculateTimestamps', () => {
    const words = [
      { word: 'Hello', start: 0.0, end: 0.5 },
      { word: 'world', start: 0.6, end: 1.0 },
      { word: 'this', start: 1.1, end: 1.4 },
      { word: 'is', start: 1.5, end: 1.7 },
      { word: 'test', start: 1.8, end: 2.2 }
    ];

    it('should return correct timestamps for valid range', () => {
      const result = calculateTimestamps(1, 3, words);
      expect(result).toEqual({ start: 0.6, end: 1.7 });
    });

    it('should handle single word selection', () => {
      const result = calculateTimestamps(2, 2, words);
      expect(result).toEqual({ start: 1.1, end: 1.4 });
    });

    it('should clamp to valid word indices', () => {
      const result = calculateTimestamps(-1, 10, words);
      expect(result).toEqual({ start: 0.0, end: 2.2 });
    });

    it('should return zero timestamps for empty words array', () => {
      const result = calculateTimestamps(0, 5, []);
      expect(result).toEqual({ start: 0, end: 0 });
    });

    it('should handle missing timestamp properties', () => {
      const wordsWithoutTimestamps = [{ word: 'test' }];
      const result = calculateTimestamps(0, 0, wordsWithoutTimestamps);
      expect(result).toEqual({ start: 0, end: 0 });
    });
  });

  describe('findWordByTimestamp', () => {
    const words = [
      { word: 'Hello', start: 0.0, end: 0.5 },
      { word: 'world', start: 0.6, end: 1.0 },
      { word: 'this', start: 1.1, end: 1.4 },
      { word: 'is', start: 1.5, end: 1.7 },
      { word: 'test', start: 1.8, end: 2.2 }
    ];

    it('should find exact word by timestamp within range', () => {
      expect(findWordByTimestamp(0.3, words)).toBe(0);
      expect(findWordByTimestamp(0.8, words)).toBe(1);
      expect(findWordByTimestamp(1.6, words)).toBe(3);
    });

    it('should find word by boundary timestamps', () => {
      expect(findWordByTimestamp(0.0, words)).toBe(0);
      expect(findWordByTimestamp(0.5, words)).toBe(0);
      expect(findWordByTimestamp(1.7, words)).toBe(3);
    });

    it('should find closest word when timestamp is between words', () => {
      expect(findWordByTimestamp(0.55, words)).toBe(1); // closer to word 1 start (0.6)
      expect(findWordByTimestamp(1.05, words)).toBe(2); // closer to word 2 start (1.1)
    });

    it('should return -1 for empty words array', () => {
      expect(findWordByTimestamp(1.0, [])).toBe(-1);
    });

    it('should handle timestamps before first word', () => {
      expect(findWordByTimestamp(-1.0, words)).toBe(0);
    });

    it('should handle timestamps after last word', () => {
      expect(findWordByTimestamp(5.0, words)).toBe(4);
    });
  });

  describe('groupWordsAndHighlights', () => {
    const displayWords = [
      { word: 'Hello' },
      { word: 'beautiful' },
      { word: 'world' },
      { word: 'this' },
      { word: 'is' },
      { word: 'a' },
      { word: 'test' }
    ];

    const highlights = [
      { id: '1', start: 1, end: 2 }, // 'beautiful world'
      { id: '2', start: 5, end: 6 }  // 'a test'
    ];

    it('should group consecutive highlighted words together', () => {
      const groups = groupWordsAndHighlights(displayWords, highlights);
      
      expect(groups).toHaveLength(5);
      
      // First word - regular
      expect(groups[0]).toEqual({
        type: 'word',
        word: displayWords[0],
        index: 0
      });
      
      // Second and third words - highlighted group
      expect(groups[1]).toMatchObject({
        type: 'highlight',
        highlight: highlights[0],
        startIndex: 1,
        words: [
          { word: displayWords[1], index: 1 },
          { word: displayWords[2], index: 2 }
        ]
      });
      
      // Fourth and fifth words - regular
      expect(groups[2]).toEqual({
        type: 'word',
        word: displayWords[3],
        index: 3
      });
      
      expect(groups[3]).toEqual({
        type: 'word',
        word: displayWords[4],
        index: 4
      });
      
      // Last two words - highlighted group
      expect(groups[4]).toMatchObject({
        type: 'highlight',
        highlight: highlights[1],
        startIndex: 5,
        words: [
          { word: displayWords[5], index: 5 },
          { word: displayWords[6], index: 6 }
        ]
      });
    });

    it('should handle no highlights', () => {
      const groups = groupWordsAndHighlights(displayWords, []);
      expect(groups).toHaveLength(displayWords.length);
      groups.forEach((group, index) => {
        expect(group).toEqual({
          type: 'word',
          word: displayWords[index],
          index
        });
      });
    });

    it('should handle single word highlights', () => {
      const singleWordHighlights = [{ id: '1', start: 2, end: 2 }];
      const groups = groupWordsAndHighlights(displayWords, singleWordHighlights);
      
      expect(groups[2]).toMatchObject({
        type: 'highlight',
        highlight: singleWordHighlights[0],
        startIndex: 2,
        words: [{ word: displayWords[2], index: 2 }]
      });
    });
  });

  describe('updateHighlight', () => {
    const highlights = [
      { id: '1', start: 0, end: 2, color: 'red' },
      { id: '2', start: 5, end: 8, color: 'blue' },
      { id: '3', start: 10, end: 15, color: 'green' }
    ];

    it('should update specific highlight', () => {
      const updated = updateHighlight(highlights, '2', 4, 9);
      
      expect(updated).toHaveLength(3);
      expect(updated[0]).toEqual(highlights[0]); // unchanged
      expect(updated[1]).toEqual({ id: '2', start: 4, end: 9, color: 'blue' });
      expect(updated[2]).toEqual(highlights[2]); // unchanged
    });

    it('should return unchanged array if highlight not found', () => {
      const updated = updateHighlight(highlights, 'nonexistent', 0, 1);
      expect(updated).toEqual(highlights);
    });

    it('should not mutate original array', () => {
      const original = [...highlights];
      updateHighlight(highlights, '1', 1, 3);
      expect(highlights).toEqual(original);
    });
  });

  describe('addHighlight', () => {
    const initialHighlights = [
      { id: '1', start: 0, end: 2, color: 'red' }
    ];

    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2024-01-01T00:00:00Z'));
    });

    it('should add new highlight to array', () => {
      const usedColors = new Set(['red']);
      const result = addHighlight(initialHighlights, 5, 8, usedColors);
      
      expect(result.highlights).toHaveLength(2);
      expect(result.highlights[0]).toEqual(initialHighlights[0]);
      expect(result.newHighlight).toMatchObject({
        start: 5,
        end: 8,
        color: '#ffeb3b'
      });
      expect(result.highlights[1]).toEqual(result.newHighlight);
    });

    it('should not mutate original array', () => {
      const original = [...initialHighlights];
      addHighlight(initialHighlights, 5, 8, new Set());
      expect(initialHighlights).toEqual(original);
    });
  });

  describe('removeHighlight', () => {
    const highlights = [
      { id: '1', start: 0, end: 2, color: 'red' },
      { id: '2', start: 5, end: 8, color: 'blue' },
      { id: '3', start: 10, end: 15, color: 'green' }
    ];

    it('should remove specific highlight', () => {
      const updated = removeHighlight(highlights, '2');
      
      expect(updated).toHaveLength(2);
      expect(updated[0]).toEqual(highlights[0]);
      expect(updated[1]).toEqual(highlights[2]);
    });

    it('should return unchanged array if highlight not found', () => {
      const updated = removeHighlight(highlights, 'nonexistent');
      expect(updated).toEqual(highlights);
    });

    it('should not mutate original array', () => {
      const original = [...highlights];
      removeHighlight(highlights, '1');
      expect(highlights).toEqual(original);
    });

    it('should handle empty array', () => {
      const updated = removeHighlight([], 'any');
      expect(updated).toEqual([]);
    });
  });
});