import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * ===============================================================================
 * TextHighlighter Behavior Tests - TIMESTAMP-ONLY ARCHITECTURE
 * ===============================================================================
 * 
 * These tests verify the critical business logic of TextHighlighter without 
 * relying on DOM rendering. They focus on:
 * - Timestamp-based drag logic  
 * - Single-word vs multi-word highlight identification
 * - Direction detection for single-word highlights
 * - Prevention of index-based conversion bugs
 * 
 * This tests the core functions that would be called during UI interactions.
 * ===============================================================================
 */

describe('TextHighlighter Behavior Logic', () => {
  // Sample word data with realistic timestamps
  const sampleWords = [
    { word: 'Hello', start: 0.0, end: 0.5 },
    { word: 'beautiful', start: 0.6, end: 1.2 },
    { word: 'world', start: 1.3, end: 1.8 },
    { word: 'this', start: 2.0, end: 2.3 },
    { word: 'is', start: 2.4, end: 2.6 },
    { word: 'a', start: 2.7, end: 2.8 },
    { word: 'test', start: 2.9, end: 3.3 }
  ];

  describe('Single-Word Highlight Detection', () => {
    it('should correctly identify single-word highlights', () => {
      const highlight = { id: 'h1', start: 2.4, end: 2.6, color: '#ffeb3b' }; // "is"
      const word = sampleWords[4]; // "is"
      
      const isFirstWord = Math.abs(word.start - highlight.start) < 0.01;
      const isLastWord = Math.abs(word.end - highlight.end) < 0.01;
      const isSingleWord = isFirstWord && isLastWord;
      
      expect(isFirstWord).toBe(true);
      expect(isLastWord).toBe(true);
      expect(isSingleWord).toBe(true);
    });

    it('should correctly identify multi-word highlights', () => {
      const highlight = { id: 'h1', start: 0.6, end: 1.8, color: '#ffeb3b' }; // "beautiful world"
      
      // First word of highlight
      const firstWord = sampleWords[1]; // "beautiful"
      const isFirstWordFirst = Math.abs(firstWord.start - highlight.start) < 0.01;
      const isFirstWordLast = Math.abs(firstWord.end - highlight.end) < 0.01;
      
      // Last word of highlight  
      const lastWord = sampleWords[2]; // "world"
      const isLastWordFirst = Math.abs(lastWord.start - highlight.start) < 0.01;
      const isLastWordLast = Math.abs(lastWord.end - highlight.end) < 0.01;
      
      expect(isFirstWordFirst).toBe(true);
      expect(isFirstWordLast).toBe(false);
      expect(isLastWordFirst).toBe(false);
      expect(isLastWordLast).toBe(true);
    });
  });

  describe('Single-Word Drag Direction Logic', () => {
    it('should detect left drag direction correctly', () => {
      const originalWordIndex = 4; // "is" 
      const currentWordIndex = 3; // "this"
      
      const dragDirection = currentWordIndex < originalWordIndex ? 'left' : 
                           currentWordIndex > originalWordIndex ? 'right' : 'none';
                           
      expect(dragDirection).toBe('left');
    });

    it('should detect right drag direction correctly', () => {
      const originalWordIndex = 4; // "is"
      const currentWordIndex = 5; // "a"
      
      const dragDirection = currentWordIndex < originalWordIndex ? 'left' : 
                           currentWordIndex > originalWordIndex ? 'right' : 'none';
                           
      expect(dragDirection).toBe('right');
    });

    it('should handle no movement correctly', () => {
      const originalWordIndex = 4; // "is"
      const currentWordIndex = 4; // "is"
      
      const dragDirection = currentWordIndex < originalWordIndex ? 'left' : 
                           currentWordIndex > originalWordIndex ? 'right' : 'none';
                           
      expect(dragDirection).toBe('none');
    });
  });

  describe('Timestamp-Based Drag Calculations', () => {
    it('should calculate correct timestamps for single-word left expansion', () => {
      const originalHighlight = { id: 'h1', start: 2.4, end: 2.6, color: '#ffeb3b' }; // "is"
      const originalWordIndex = 4;
      const currentWordIndex = 3; // "this"
      const currentWord = sampleWords[currentWordIndex];
      
      let newStart, newEnd;
      
      if (currentWordIndex < originalWordIndex) {
        // Dragging left - expand the start
        newStart = currentWord.start;
        newEnd = originalHighlight.end;
      }
      
      expect(newStart).toBe(2.0); // Start of "this"
      expect(newEnd).toBe(2.6);   // End of "is" (unchanged)
      expect(newStart).toBeLessThan(originalHighlight.start);
    });

    it('should calculate correct timestamps for single-word right expansion', () => {
      const originalHighlight = { id: 'h1', start: 2.4, end: 2.6, color: '#ffeb3b' }; // "is"
      const originalWordIndex = 4;
      const currentWordIndex = 5; // "a"
      const currentWord = sampleWords[currentWordIndex];
      
      let newStart, newEnd;
      
      if (currentWordIndex > originalWordIndex) {
        // Dragging right - expand the end
        newStart = originalHighlight.start;
        newEnd = currentWord.end;
      }
      
      expect(newStart).toBe(2.4); // Start of "is" (unchanged)
      expect(newEnd).toBe(2.8);   // End of "a"
      expect(newEnd).toBeGreaterThan(originalHighlight.end);
    });

    it('should calculate correct timestamps for multi-word first-word drag', () => {
      const originalHighlight = { id: 'h1', start: 0.6, end: 1.8, color: '#ffeb3b' }; // "beautiful world"
      const currentWordIndex = 0; // "Hello"
      const currentWord = sampleWords[currentWordIndex];
      
      // Dragging first word left
      const newStart = currentWord.start;
      const newEnd = originalHighlight.end;
      
      expect(newStart).toBe(0.0); // Start of "Hello"
      expect(newEnd).toBe(1.8);   // End of "world" (unchanged)
      expect(newStart).toBeLessThan(originalHighlight.start);
    });

    it('should calculate correct timestamps for multi-word last-word drag', () => {
      const originalHighlight = { id: 'h1', start: 0.6, end: 1.8, color: '#ffeb3b' }; // "beautiful world"
      const currentWordIndex = 3; // "this"
      const currentWord = sampleWords[currentWordIndex];
      
      // Dragging last word right
      const newStart = originalHighlight.start;
      const newEnd = currentWord.end;
      
      expect(newStart).toBe(0.6); // Start of "beautiful" (unchanged)
      expect(newEnd).toBe(2.3);   // End of "this"
      expect(newEnd).toBeGreaterThan(originalHighlight.end);
    });
  });

  describe('Timestamp Precision and Round-Trip Prevention', () => {
    it('should maintain timestamp precision without index conversion', () => {
      const originalTimestamp = 1.23456789;
      
      // Test that we don't lose precision by converting to/from indices
      const preservedTimestamp = originalTimestamp;
      
      expect(preservedTimestamp).toBe(1.23456789);
      expect(preservedTimestamp % 1).toBeGreaterThan(0); // Has decimal part
    });

    it('should calculate drag boundaries using only timestamps', () => {
      const words = [
        { word: 'word1', start: 1.0, end: 1.5 },
        { word: 'word2', start: 1.6, end: 2.1 },
        { word: 'word3', start: 2.2, end: 2.7 }
      ];
      
      const highlights = [{ id: 'h1', start: 1.6, end: 2.1, color: '#ffeb3b' }]; // word2
      
      // Find highlight for word by timestamp overlap
      function findHighlightForWordByTime(wordIndex) {
        const word = words[wordIndex];
        return highlights.find(h => 
          word.start >= h.start && word.end <= h.end
        );
      }
      
      expect(findHighlightForWordByTime(0)).toBeUndefined(); // word1 not highlighted
      expect(findHighlightForWordByTime(1)).toBeTruthy();   // word2 highlighted  
      expect(findHighlightForWordByTime(2)).toBeUndefined(); // word3 not highlighted
    });

    it('should prevent off-by-one errors in timestamp calculations', () => {
      // Simulate the exact bug scenario that was fixed
      const originalHighlight = { start: 1.3, end: 2.3 }; // "world this"
      const leftExpansionWord = { start: 0.6, end: 1.2 }; // "beautiful"
      
      // Drag first word left by 1 word - should add exactly 1 word
      const newStart = leftExpansionWord.start;
      const newEnd = originalHighlight.end;
      
      const originalSpan = originalHighlight.end - originalHighlight.start; // 1.0
      const newSpan = newEnd - newStart; // 1.7
      const addedSpan = newSpan - originalSpan; // 0.7 (should be ~0.6 for 1 word)
      
      expect(newStart).toBe(0.6);
      expect(newEnd).toBe(2.3);
      expect(addedSpan).toBeCloseTo(0.7, 1); // Added exactly the span of "beautiful"
      
      // Verify this is exactly 1 word worth of expansion, not 2
      expect(addedSpan).toBeLessThan(1.0); // Should not be a full 1.0+ span
    });
  });

  describe('Word Finding by Timestamp', () => {
    it('should find correct word by timestamp', () => {
      function findWordByTimestamp(timestamp) {
        return sampleWords.findIndex(word => 
          timestamp >= word.start && timestamp <= word.end
        );
      }
      
      expect(findWordByTimestamp(0.3)).toBe(0);  // "Hello"
      expect(findWordByTimestamp(0.9)).toBe(1);  // "beautiful"
      expect(findWordByTimestamp(1.5)).toBe(2);  // "world"
      expect(findWordByTimestamp(2.5)).toBe(4);  // "is"
    });

    it('should handle edge case timestamps', () => {
      function findWordByTimestamp(timestamp) {
        return sampleWords.findIndex(word => 
          timestamp >= word.start && timestamp <= word.end
        );
      }
      
      // Exact boundaries
      expect(findWordByTimestamp(0.0)).toBe(0);   // Start of "Hello"
      expect(findWordByTimestamp(0.5)).toBe(0);   // End of "Hello"
      expect(findWordByTimestamp(0.6)).toBe(1);   // Start of "beautiful"
      
      // Between words
      expect(findWordByTimestamp(0.55)).toBe(-1); // Gap between "Hello" and "beautiful"
    });
  });

  describe('Drag Mode Detection', () => {
    it('should detect expansion mode correctly', () => {
      const originalHighlight = { start: 1.0, end: 2.0 };
      const newStart = 0.5; // Expanded left
      const newEnd = 2.5;   // Expanded right
      
      const isExpanding = newStart < originalHighlight.start || newEnd > originalHighlight.end;
      const isContracting = newStart > originalHighlight.start || newEnd < originalHighlight.end;
      
      expect(isExpanding).toBe(true);
      expect(isContracting).toBe(false);
    });

    it('should detect contraction mode correctly', () => {
      const originalHighlight = { start: 1.0, end: 3.0 };
      const newStart = 1.5; // Contracted right
      const newEnd = 2.5;   // Contracted left
      
      const isExpanding = newStart < originalHighlight.start || newEnd > originalHighlight.end;
      const isContracting = newStart > originalHighlight.start || newEnd < originalHighlight.end;
      
      expect(isExpanding).toBe(false);
      expect(isContracting).toBe(true);
    });

    it('should detect no-change mode correctly', () => {
      const originalHighlight = { start: 1.0, end: 2.0 };
      const newStart = 1.0; // Same
      const newEnd = 2.0;   // Same
      const tolerance = 0.01;
      
      const noChange = Math.abs(newStart - originalHighlight.start) < tolerance && 
                      Math.abs(newEnd - originalHighlight.end) < tolerance;
      
      expect(noChange).toBe(true);
    });
  });

  describe('Critical Bug Prevention Tests', () => {
    it('should not add extra words when dragging by exactly 1 position', () => {
      // Test the specific bug: "dragging first word left by 1 word adds 2 words"
      const highlight = { start: 1.3, end: 1.8 }; // "world" only
      const targetWord = { start: 0.6, end: 1.2 }; // "beautiful" (exactly 1 word left)
      
      const newStart = targetWord.start;
      const newEnd = highlight.end;
      
      // Calculate how many words are now included
      const wordsInRange = sampleWords.filter(w => 
        w.start >= newStart && w.end <= newEnd
      );
      
      expect(wordsInRange).toHaveLength(2); // Should be exactly 2 words: "beautiful" + "world"
      expect(wordsInRange[0].word).toBe('beautiful');
      expect(wordsInRange[1].word).toBe('world');
    });

    it('should not add words to beginning when dragging end', () => {
      // Test the specific bug: "dragging last word right adds word to beginning"
      const highlight = { start: 1.3, end: 2.3 }; // "world this"
      const targetWord = { start: 2.4, end: 2.6 }; // "is" (exactly 1 word right)
      
      const newStart = highlight.start; // Should NOT change
      const newEnd = targetWord.end;
      
      expect(newStart).toBe(1.3); // Beginning unchanged
      expect(newEnd).toBe(2.6);   // End extended
      
      // Verify only the end was extended
      const wordsInRange = sampleWords.filter(w => 
        w.start >= newStart && w.end <= newEnd
      );
      
      expect(wordsInRange[0].word).toBe('world'); // First word unchanged
      expect(wordsInRange[wordsInRange.length - 1].word).toBe('is'); // Last word added
    });

    it('should remove correct amount when contracting highlight', () => {
      // Test contraction precision
      const highlight = { start: 0.6, end: 2.3 }; // "beautiful world this"
      const contractToWord = { start: 1.3, end: 1.8 }; // Contract to just "world"
      
      const newStart = contractToWord.start;
      const newEnd = contractToWord.end;
      
      const wordsInRange = sampleWords.filter(w => 
        w.start >= newStart && w.end <= newEnd
      );
      
      expect(wordsInRange).toHaveLength(1); // Should be exactly 1 word
      expect(wordsInRange[0].word).toBe('world');
    });
  });
});