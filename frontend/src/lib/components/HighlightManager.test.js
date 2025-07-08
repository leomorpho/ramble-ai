/**
 * Unit tests for HighlightManager
 * Run with: cd frontend && npm test
 */

import { describe, test, expect, beforeEach } from 'vitest';
import { HighlightManager } from './HighlightManager.js';

// Mock words data for testing
const mockWords = [
  { word: 'Hello', start: 0.0, end: 0.5 },
  { word: 'world', start: 0.6, end: 1.0 },
  { word: 'this', start: 1.1, end: 1.4 },
  { word: 'is', start: 1.5, end: 1.7 },
  { word: 'a', start: 1.8, end: 1.9 },
  { word: 'test', start: 2.0, end: 2.4 },
];

describe('HighlightManager', () => {
  let manager;

  beforeEach(() => {
    manager = new HighlightManager(mockWords);
  });

  describe('Basic functionality', () => {
    test('should initialize with empty highlights', () => {
      expect(manager.highlights).toEqual([]);
      expect(manager.usedColors.size).toBe(0);
    });

    test('should find word index by timestamp', () => {
      expect(manager.findWordIndexByTime(0.25)).toBe(0); // 'Hello'
      expect(manager.findWordIndexByTime(0.8)).toBe(1);  // 'world'
      expect(manager.findWordIndexByTime(1.6)).toBe(3);  // 'is'
    });

    test('should find closest word when timestamp is not exact', () => {
      expect(manager.findWordIndexByTime(0.55)).toBe(1); // closer to 'world'
      expect(manager.findWordIndexByTime(1.05)).toBe(2); // closer to 'this'
    });
  });

  describe('Highlight creation', () => {
    test('should create highlight from selection', () => {
      const result = manager.createHighlightFromSelection(1, 3);
      
      expect(result.indexHighlights).toHaveLength(1);
      expect(result.newHighlight.start).toBe(1);
      expect(result.newHighlight.end).toBe(3);
      expect(result.newHighlight.id).toBeDefined();
      expect(result.newHighlight.color).toBeDefined();
    });

    test('should normalize start/end order', () => {
      const result = manager.createHighlightFromSelection(3, 1);
      
      expect(result.newHighlight.start).toBe(1);
      expect(result.newHighlight.end).toBe(3);
    });

    test('should throw error for same start and end', () => {
      expect(() => {
        manager.createHighlightFromSelection(1, 1);
      }).toThrow('Cannot create highlight with same start and end index');
    });

    test('should throw error for overlapping highlights', () => {
      manager.createHighlightFromSelection(1, 3);
      
      expect(() => {
        manager.createHighlightFromSelection(2, 4);
      }).toThrow('Cannot create overlapping highlight');
    });

    test('should allow non-overlapping highlights', () => {
      manager.createHighlightFromSelection(1, 2);
      const result = manager.createHighlightFromSelection(4, 5);
      
      expect(manager.highlights).toHaveLength(2);
    });
  });

  describe('Highlight updates', () => {
    test('should update highlight bounds', () => {
      const createResult = manager.createHighlightFromSelection(1, 3);
      const highlightId = createResult.newHighlight.id;
      
      const updateResult = manager.updateHighlightBounds(highlightId, 1, 4);
      
      const updatedHighlight = manager.highlights.find(h => h.id === highlightId);
      expect(updatedHighlight.start).toBe(1);
      expect(updatedHighlight.end).toBe(4);
    });

    test('should throw error when updating non-existent highlight', () => {
      expect(() => {
        manager.updateHighlightBounds('non-existent-id', 1, 3);
      }).toThrow('Highlight with id non-existent-id not found');
    });

    test('should prevent overlapping updates', () => {
      const result1 = manager.createHighlightFromSelection(1, 2);
      const result2 = manager.createHighlightFromSelection(4, 5);
      
      expect(() => {
        manager.updateHighlightBounds(result1.newHighlight.id, 1, 4);
      }).toThrow('Updated highlight would overlap with existing highlights');
    });
  });

  describe('Highlight deletion', () => {
    test('should delete highlight', () => {
      const result = manager.createHighlightFromSelection(1, 3);
      const highlightId = result.newHighlight.id;
      
      manager.deleteHighlight(highlightId);
      
      expect(manager.highlights).toHaveLength(0);
      expect(manager.usedColors.has(result.newHighlight.color)).toBe(false);
    });

    test('should handle deletion of non-existent highlight gracefully', () => {
      expect(() => {
        manager.deleteHighlight('non-existent-id');
      }).not.toThrow();
    });
  });

  describe('Drag operations', () => {
    test('should calculate drag selection for first word expansion', () => {
      const originalHighlight = { start: 2, end: 4 };
      
      const result = manager.calculateDragSelection(2, 0, originalHighlight, true, false);
      
      expect(result.start).toBe(0);
      expect(result.end).toBe(4);
      expect(result.mode).toBe('expand');
    });

    test('should calculate drag selection for first word contraction', () => {
      const originalHighlight = { start: 2, end: 4 };
      
      const result = manager.calculateDragSelection(2, 3, originalHighlight, true, false);
      
      expect(result.start).toBe(3);
      expect(result.end).toBe(4);
      expect(result.mode).toBe('contract');
    });

    test('should calculate drag selection for last word expansion', () => {
      const originalHighlight = { start: 1, end: 3 };
      
      const result = manager.calculateDragSelection(3, 5, originalHighlight, false, true);
      
      expect(result.start).toBe(1);
      expect(result.end).toBe(5);
      expect(result.mode).toBe('expand');
    });

    test('should calculate drag selection for last word contraction', () => {
      const originalHighlight = { start: 1, end: 4 };
      
      const result = manager.calculateDragSelection(4, 2, originalHighlight, false, true);
      
      expect(result.start).toBe(1);
      expect(result.end).toBe(2);
      expect(result.mode).toBe('contract');
    });
  });

  describe('Timestamp conversion', () => {
    test('should convert timestamp highlights to indices', () => {
      const timestampHighlights = [
        { id: 'test-1', start: 0.25, end: 1.0, color: 'red' }
      ];
      
      const indexHighlights = manager.convertTimestampsToIndices(timestampHighlights);
      
      expect(indexHighlights[0].start).toBe(0); // 'Hello'
      expect(indexHighlights[0].end).toBe(1);   // 'world'
    });

    test('should convert index highlights to timestamps', () => {
      const indexHighlights = [
        { id: 'test-1', start: 0, end: 1, color: 'red' }
      ];
      
      const timestampHighlights = manager.convertIndicesToTimestamps(indexHighlights);
      
      expect(timestampHighlights[0].start).toBe(0.0);
      expect(timestampHighlights[0].end).toBe(1.0);
    });
  });

  describe('Utility functions', () => {
    test('should find highlight for word', () => {
      const result = manager.createHighlightFromSelection(1, 3);
      
      const foundHighlight = manager.findHighlightForWord(2);
      expect(foundHighlight.id).toBe(result.newHighlight.id);
      
      const notFound = manager.findHighlightForWord(5);
      expect(notFound).toBeUndefined();
    });

    test('should check if word is in selection', () => {
      expect(manager.isWordInSelection(2, 1, 4)).toBe(true);
      expect(manager.isWordInSelection(0, 1, 4)).toBe(false);
      expect(manager.isWordInSelection(5, 1, 4)).toBe(false);
    });

    test('should provide debug info', () => {
      manager.createHighlightFromSelection(1, 3);
      const debug = manager.getDebugInfo();
      
      expect(debug.wordCount).toBe(6);
      expect(debug.highlightCount).toBe(1);
      expect(debug.highlights).toHaveLength(1);
      expect(debug.usedColors).toHaveLength(1);
    });
  });
});

// Simple test runner for browser console (if no test framework)
if (typeof window !== 'undefined') {
  console.log('Running HighlightManager tests...');
  
  // Basic smoke test
  const manager = new HighlightManager(mockWords);
  
  try {
    // Test creation
    const result = manager.createHighlightFromSelection(1, 3);
    console.assert(result.newHighlight.start === 1, 'Highlight start should be 1');
    console.assert(result.newHighlight.end === 3, 'Highlight end should be 3');
    
    // Test update
    manager.updateHighlightBounds(result.newHighlight.id, 1, 4);
    const updated = manager.findHighlightForWord(4);
    console.assert(updated !== undefined, 'Word 4 should be highlighted after update');
    
    // Test deletion
    manager.deleteHighlight(result.newHighlight.id);
    console.assert(manager.highlights.length === 0, 'Should have no highlights after deletion');
    
    console.log('✅ Basic HighlightManager tests passed!');
  } catch (error) {
    console.error('❌ HighlightManager test failed:', error);
  }
}