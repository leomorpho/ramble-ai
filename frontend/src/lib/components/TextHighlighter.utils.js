/**
 * Pure utility functions for TextHighlighter component
 * These functions are extracted for easier testing
 */

export function generateUniqueColor(usedColors = new Set()) {
  const baseColors = ['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292'];
  
  // Try base colors first
  for (const color of baseColors) {
    if (!usedColors.has(color)) {
      return color;
    }
  }
  
  // Generate random pastel color
  const hue = Math.floor(Math.random() * 360);
  const saturation = 45 + Math.random() * 30;
  const lightness = 65 + Math.random() * 20;
  return `hsl(${hue}, ${saturation}%, ${lightness}%)`;
}

export function createHighlight(start, end, usedColors = new Set(), color = null) {
  return {
    id: `highlight_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    start,
    end,
    color: color || generateUniqueColor(usedColors)
  };
}

export function isWordInHighlight(wordIndex, highlight) {
  return wordIndex >= highlight.start && wordIndex <= highlight.end;
}

export function isWordInSelection(wordIndex, selectionStart, selectionEnd, isSelecting) {
  if (!isSelecting || selectionStart === null || selectionEnd === null) return false;
  const start = Math.min(selectionStart, selectionEnd);
  const end = Math.max(selectionStart, selectionEnd);
  return wordIndex >= start && wordIndex <= end;
}

export function findHighlightForWord(wordIndex, highlights) {
  return highlights.find(h => isWordInHighlight(wordIndex, h));
}

export function checkOverlap(start, end, highlights, excludeId = null) {
  return highlights.some(h => 
    h.id !== excludeId && 
    start <= h.end && end >= h.start
  );
}

export function calculateTimestamps(startIndex, endIndex, words) {
  if (!words || words.length === 0) return { start: 0, end: 0 };
  
  const startWord = words[Math.max(0, Math.min(startIndex, words.length - 1))];
  const endWord = words[Math.max(0, Math.min(endIndex, words.length - 1))];
  
  return {
    start: startWord.start || 0,
    end: endWord.end || 0
  };
}

export function findWordByTimestamp(timestamp, words) {
  if (!words || words.length === 0) return -1;
  
  // Find exact match
  for (let i = 0; i < words.length; i++) {
    const word = words[i];
    if (word.start <= timestamp && timestamp <= word.end) {
      return i;
    }
  }
  
  // Find closest by start time
  let closestIndex = 0;
  let minDistance = Math.abs(words[0].start - timestamp);
  
  for (let i = 1; i < words.length; i++) {
    const distance = Math.abs(words[i].start - timestamp);
    if (distance < minDistance) {
      minDistance = distance;
      closestIndex = i;
    }
  }
  
  return closestIndex;
}

export function groupWordsAndHighlights(displayWords, highlights) {
  const groups = [];
  let i = 0;
  
  while (i < displayWords.length) {
    const highlight = findHighlightForWord(i, highlights);
    
    if (highlight) {
      // Start of a highlight group
      const group = {
        type: 'highlight',
        highlight: highlight,
        words: [],
        startIndex: i
      };
      
      // Collect all consecutive words in this highlight
      while (i < displayWords.length && findHighlightForWord(i, highlights)?.id === highlight.id) {
        group.words.push({
          word: displayWords[i],
          index: i
        });
        i++;
      }
      
      groups.push(group);
    } else {
      // Regular word
      groups.push({
        type: 'word',
        word: displayWords[i],
        index: i
      });
      i++;
    }
  }
  
  return groups;
}

export function updateHighlight(highlights, highlightId, newStart, newEnd) {
  return highlights.map(h => 
    h.id === highlightId 
      ? { ...h, start: newStart, end: newEnd }
      : h
  );
}

export function addHighlight(highlights, start, end, usedColors, color = null) {
  const newHighlight = createHighlight(start, end, usedColors, color);
  return {
    highlights: [...highlights, newHighlight],
    newHighlight
  };
}

export function removeHighlight(highlights, highlightId) {
  return highlights.filter(h => h.id !== highlightId);
}