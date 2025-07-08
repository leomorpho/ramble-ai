/**
 * Highlighting state manager - handles all highlight operations in a testable way
 * This module contains pure functions for managing highlight state
 */

import { 
  generateUniqueColor, 
  createHighlight, 
  isWordInHighlight, 
  findHighlightForWord, 
  checkOverlap, 
  updateHighlight, 
  addHighlight, 
  removeHighlight 
} from './TextHighlighter.utils.js';

export class HighlightManager {
  constructor(words = []) {
    this.words = words;
    this.highlights = [];
    this.usedColors = new Set();
  }

  /**
   * Convert timestamp-based highlights to word-index-based
   */
  convertTimestampsToIndices(timestampHighlights) {
    return timestampHighlights.map((h) => ({
      ...h,
      start: this.findWordIndexByTime(h.start),
      end: this.findWordIndexByTime(h.end),
    }));
  }

  /**
   * Convert word-index-based highlights back to timestamp-based
   */
  convertIndicesToTimestamps(indexHighlights) {
    return indexHighlights.map((h) => ({
      ...h,
      start: this.words[h.start]?.start || 0,
      end: this.words[h.end]?.end || 0,
    }));
  }

  /**
   * Find word index by timestamp
   */
  findWordIndexByTime(timestamp) {
    if (!this.words || this.words.length === 0) return 0;

    // Find exact match
    for (let i = 0; i < this.words.length; i++) {
      const word = this.words[i];
      if (word.start <= timestamp && timestamp <= word.end) {
        return i;
      }
    }

    // Find closest
    let closestIndex = 0;
    let minDistance = Math.abs(this.words[0].start - timestamp);

    for (let i = 1; i < this.words.length; i++) {
      const distance = Math.abs(this.words[i].start - timestamp);
      if (distance < minDistance) {
        minDistance = distance;
        closestIndex = i;
      }
    }

    return closestIndex;
  }

  /**
   * Set the current highlights (timestamp-based)
   */
  setHighlights(timestampHighlights) {
    this.highlights = this.convertTimestampsToIndices(timestampHighlights);
    this.usedColors.clear();
    this.highlights.forEach((h) => this.usedColors.add(h.color));
    return this.highlights;
  }

  /**
   * Get current highlights as timestamps
   */
  getTimestampHighlights() {
    return this.convertIndicesToTimestamps(this.highlights);
  }

  /**
   * Create a new highlight from selection
   */
  createHighlightFromSelection(startIndex, endIndex, color = null, allowSingleWord = false) {
    if (startIndex === endIndex && !allowSingleWord) {
      throw new Error('Cannot create highlight with same start and end index');
    }

    const normalizedStart = Math.min(startIndex, endIndex);
    const normalizedEnd = Math.max(startIndex, endIndex);

    if (checkOverlap(normalizedStart, normalizedEnd, this.highlights)) {
      throw new Error('Cannot create overlapping highlight');
    }

    const result = addHighlight(this.highlights, normalizedStart, normalizedEnd, this.usedColors, color);
    this.highlights = result.highlights;
    this.usedColors.add(result.newHighlight.color);

    return {
      indexHighlights: this.highlights,
      timestampHighlights: this.getTimestampHighlights(),
      newHighlight: result.newHighlight
    };
  }

  /**
   * Update an existing highlight
   */
  updateHighlightBounds(highlightId, newStartIndex, newEndIndex) {
    const normalizedStart = Math.min(newStartIndex, newEndIndex);
    const normalizedEnd = Math.max(newStartIndex, newEndIndex);

    // Find the highlight being updated
    const existingHighlight = this.highlights.find(h => h.id === highlightId);
    if (!existingHighlight) {
      throw new Error(`Highlight with id ${highlightId} not found`);
    }

    // Check for overlap with other highlights (excluding the one being updated)
    if (checkOverlap(normalizedStart, normalizedEnd, this.highlights, highlightId)) {
      throw new Error('Updated highlight would overlap with existing highlights');
    }

    this.highlights = updateHighlight(this.highlights, highlightId, normalizedStart, normalizedEnd);

    return {
      indexHighlights: this.highlights,
      timestampHighlights: this.getTimestampHighlights()
    };
  }

  /**
   * Delete a highlight
   */
  deleteHighlight(highlightId) {
    const highlight = this.highlights.find(h => h.id === highlightId);
    if (highlight) {
      this.usedColors.delete(highlight.color);
    }

    this.highlights = removeHighlight(this.highlights, highlightId);

    return {
      indexHighlights: this.highlights,
      timestampHighlights: this.getTimestampHighlights()
    };
  }

  /**
   * Calculate selection bounds for drag operations
   */
  calculateDragSelection(anchorWordIndex, currentWordIndex, originalHighlight, isFirstWord, isLastWord) {
    if (isFirstWord) {
      // Dragging from first word - selection is from current position to end of original highlight
      return {
        start: Math.min(currentWordIndex, originalHighlight.end),
        end: originalHighlight.end,
        mode: currentWordIndex < originalHighlight.start ? 'expand' : 
              currentWordIndex > originalHighlight.start && currentWordIndex <= originalHighlight.end ? 'contract' : 
              null
      };
    } else if (isLastWord) {
      // Dragging from last word - selection is from start of original highlight to current position
      return {
        start: originalHighlight.start,
        end: Math.max(currentWordIndex, originalHighlight.start),
        mode: currentWordIndex > originalHighlight.end ? 'expand' : 
              currentWordIndex < originalHighlight.end && currentWordIndex >= originalHighlight.start ? 'contract' : 
              null
      };
    }

    throw new Error('Invalid drag operation: must be first or last word');
  }

  /**
   * Find highlight for a specific word
   */
  findHighlightForWord(wordIndex) {
    return findHighlightForWord(wordIndex, this.highlights);
  }

  /**
   * Check if a word is in the current selection
   */
  isWordInSelection(wordIndex, selectionStart, selectionEnd) {
    if (selectionStart === null || selectionEnd === null) return false;
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);
    return wordIndex >= start && wordIndex <= end;
  }

  /**
   * Get debug info for troubleshooting
   */
  getDebugInfo() {
    return {
      wordCount: this.words.length,
      highlightCount: this.highlights.length,
      highlights: this.highlights.map(h => ({
        id: h.id,
        start: h.start,
        end: h.end,
        color: h.color,
        wordCount: h.end - h.start + 1
      })),
      usedColors: Array.from(this.usedColors)
    };
  }
}