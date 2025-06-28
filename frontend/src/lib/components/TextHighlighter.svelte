<script>
  import { onMount } from 'svelte';
  import { Button } from "$lib/components/ui/button";
  
  let { text = '', words = [], initialHighlights = [], onHighlightsChange } = $props();
  
  // === CORE STATE ===
  let highlights = $state([]);
  let usedColors = $state(new Set());
  
  // === SELECTION STATE ===
  let isSelecting = $state(false);
  let selectionStart = $state(null);
  let selectionEnd = $state(null);
  
  // === DRAG STATE ===
  let isDragging = $state(false);
  let dragTarget = $state(null);
  
  // === UI STATE ===
  let showDeleteButton = $state(false);
  let deleteButtonHighlight = $state(null);
  let deleteButtonPosition = $state({ x: 0, y: 0 });
  
  // === CACHED DATA ===
  let wordsWithIds = $state([]);
  let wordToHighlightMap = $state(new Map());
  let initialized = $state(false);
  
  // === PURE FUNCTIONS ===
  
  function generateUniqueColor() {
    const baseColors = ['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292'];
    
    for (const color of baseColors) {
      if (!usedColors.has(color)) {
        return color;
      }
    }
    
    const hue = Math.floor(Math.random() * 360);
    const saturation = 45 + Math.random() * 30;
    const lightness = 65 + Math.random() * 20;
    return `hsl(${hue}, ${saturation}%, ${lightness}%)`;
  }
  
  function createHighlight(startTime, endTime, startWordId, endWordId, wordIds = []) {
    return {
      id: `highlight_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      start: startTime,
      end: endTime,
      startWordId,
      endWordId,
      wordIds,
      color: generateUniqueColor()
    };
  }
  
  function isWordInSelection(wordIndex) {
    if (!isSelecting || selectionStart === null || selectionEnd === null) return false;
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);
    return wordIndex >= start && wordIndex <= end;
  }
  
  function findHighlightForWord(wordIndex) {
    if (!wordsWithIds || !wordsWithIds[wordIndex]) return null;
    const word = wordsWithIds[wordIndex];
    const highlightId = wordToHighlightMap.get(word.id);
    return highlightId ? highlights.find(h => h.id === highlightId) : null;
  }
  
  function checkOverlap(startTime, endTime, excludeId = null) {
    return highlights.some(h => 
      h.id !== excludeId && 
      startTime < h.end && endTime > h.start
    );
  }
  
  function calculateTimestamps(startIndex, endIndex) {
    if (!wordsWithIds || wordsWithIds.length === 0) return { start: 0, end: 0 };
    
    const startWord = wordsWithIds[Math.max(0, Math.min(startIndex, wordsWithIds.length - 1))];
    const endWord = wordsWithIds[Math.max(0, Math.min(endIndex, wordsWithIds.length - 1))];
    
    return {
      start: startWord.start || 0,
      end: endWord.end || 0
    };
  }
  
  function getWordsInTimeRange(startTime, endTime) {
    if (!wordsWithIds) return [];
    return wordsWithIds.filter(word => 
      word.start < endTime && word.end > startTime
    );
  }
  
  // === WORD PROCESSING ===
  
  function processWords() {
    if (words && words.length > 0) {
      return words.map((word, index) => ({
        ...word,
        id: `word_${index}_${word.word}_${word.start}`
      }));
    } else if (text) {
      const wordMatches = text.match(/\S+/g) || [];
      return wordMatches.map((word, index) => ({
        id: `text_word_${index}`,
        word: word,
        start: index * 0.5,
        end: index * 0.5 + 0.4
      }));
    }
    return [];
  }
  
  function rebuildWordMap() {
    const newMap = new Map();
    highlights.forEach(highlight => {
      if (highlight.wordIds) {
        highlight.wordIds.forEach(wordId => {
          newMap.set(wordId, highlight.id);
        });
      }
    });
    wordToHighlightMap = newMap;
  }
  
  function initializeData() {
    if (initialized) return;
    
    wordsWithIds = processWords();
    
    if (initialHighlights && initialHighlights.length > 0 && wordsWithIds.length > 0) {
      highlights = initialHighlights.map(h => {
        const wordsInRange = getWordsInTimeRange(h.start, h.end);
        const wordIds = wordsInRange.map(word => word.id);
        
        return {
          id: h.id,
          start: h.start,
          end: h.end,
          color: h.color,
          startWordId: wordsInRange[0]?.id,
          endWordId: wordsInRange[wordsInRange.length - 1]?.id,
          wordIds
        };
      });
      
      highlights.forEach(h => usedColors.add(h.color));
    }
    
    rebuildWordMap();
    initialized = true;
  }
  
  // === ACTIONS ===
  
  function addHighlight(startIndex, endIndex) {
    const timestamps = calculateTimestamps(startIndex, endIndex);
    const startWord = wordsWithIds[startIndex];
    const endWord = wordsWithIds[endIndex];
    
    const wordsInRange = getWordsInTimeRange(timestamps.start, timestamps.end);
    const wordIds = wordsInRange.map(word => word.id);
    
    const newHighlight = createHighlight(
      timestamps.start, 
      timestamps.end, 
      startWord?.id, 
      endWord?.id, 
      wordIds
    );
    
    highlights = [...highlights, newHighlight];
    usedColors.add(newHighlight.color);
    rebuildWordMap();
    emitChanges();
    return newHighlight;
  }
  
  function removeHighlight(highlightId) {
    const highlight = highlights.find(h => h.id === highlightId);
    if (highlight) {
      usedColors.delete(highlight.color);
      usedColors = new Set(usedColors);
    }
    highlights = highlights.filter(h => h.id !== highlightId);
    rebuildWordMap();
    emitChanges();
  }
  
  function updateHighlight(highlightId, newStartTime, newEndTime) {
    const wordsInRange = getWordsInTimeRange(newStartTime, newEndTime);
    const wordIds = wordsInRange.map(word => word.id);
    const startWordId = wordsInRange[0]?.id;
    const endWordId = wordsInRange[wordsInRange.length - 1]?.id;
    
    highlights = highlights.map(h => 
      h.id === highlightId 
        ? { 
            ...h, 
            start: newStartTime, 
            end: newEndTime,
            startWordId,
            endWordId,
            wordIds 
          }
        : h
    );
    rebuildWordMap();
  }
  
  function emitChanges() {
    if (onHighlightsChange) {
      onHighlightsChange(highlights);
    }
  }
  
  // === EVENT HANDLERS ===
  
  function handleWordMouseDown(wordIndex, event) {
    const existingHighlight = findHighlightForWord(wordIndex);
    
    if (existingHighlight) {
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    
    isSelecting = true;
    selectionStart = wordIndex;
    selectionEnd = wordIndex;
    showDeleteButton = false;
  }
  
  function handleWordMouseEnter(wordIndex) {
    if (isSelecting) {
      selectionEnd = wordIndex;
    }
    
    if (isDragging && dragTarget && wordsWithIds && wordsWithIds[wordIndex]) {
      const currentHighlight = highlights.find(h => h.id === dragTarget.highlightId);
      if (!currentHighlight) return;
      
      const word = wordsWithIds[wordIndex];
      const newStartTime = dragTarget.type === 'start' ? word.start : currentHighlight.start;
      const newEndTime = dragTarget.type === 'end' ? word.end : currentHighlight.end;
      
      if (newStartTime >= newEndTime) return;
      
      if (!checkOverlap(newStartTime, newEndTime, dragTarget.highlightId)) {
        updateHighlight(dragTarget.highlightId, newStartTime, newEndTime);
      }
    }
  }
  
  function handleWordClick(wordIndex, event) {
    const highlight = findHighlightForWord(wordIndex);
    
    if (highlight) {
      const rect = event.target.getBoundingClientRect();
      deleteButtonHighlight = highlight;
      deleteButtonPosition = { 
        x: rect.left + (rect.width / 2) - 40,
        y: rect.top - 45
      };
      showDeleteButton = true;
      event.stopPropagation();
    }
  }
  
  function handleMouseUp() {
    if (isSelecting && selectionStart !== null && selectionEnd !== null) {
      const startIndex = Math.min(selectionStart, selectionEnd);
      const endIndex = Math.max(selectionStart, selectionEnd);
      
      if (startIndex !== endIndex) {
        const timestamps = calculateTimestamps(startIndex, endIndex);
        
        if (!checkOverlap(timestamps.start, timestamps.end)) {
          addHighlight(startIndex, endIndex);
        }
      }
    }
    
    if (isDragging) {
      emitChanges();
    }
    
    isSelecting = false;
    selectionStart = null;
    selectionEnd = null;
    isDragging = false;
    dragTarget = null;
  }
  
  function handleDeleteHighlight(highlightId) {
    removeHighlight(highlightId);
    showDeleteButton = false;
    deleteButtonHighlight = null;
  }
  
  // === MOUNT ===
  
  onMount(() => {
    initializeData();
    
    document.addEventListener('mouseup', handleMouseUp);
    document.addEventListener('click', (e) => {
      if (!e.target.closest('.delete-popup')) {
        showDeleteButton = false;
      }
    });
    
    return () => {
      document.removeEventListener('mouseup', handleMouseUp);
    };
  });
</script>

<div class="highlighter" class:dragging={isDragging}>
  {#each wordsWithIds as word, wordIndex}
    {@const highlightId = wordToHighlightMap.get(word.id)}
    {@const highlight = highlightId ? highlights.find(h => h.id === highlightId) : null}
    {@const inSelection = isWordInSelection(wordIndex)}
    
    {#if highlight}
      <!-- Highlighted word -->
      <span 
        class="word highlighted"
        style:background-color={highlight.color}
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
      >
        {word.word}
      </span>
    {:else if inSelection}
      <!-- Selection preview -->
      <span class="selection-word">
        {word.word}
      </span>
    {:else}
      <!-- Regular word -->
      <span
        class="word"
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
      >
        {word.word}
      </span>
    {/if}
    
    <!-- Space between words -->
    {#if wordIndex < wordsWithIds.length - 1}{' '}{/if}
  {/each}
</div>

<!-- Delete button popup -->
{#if showDeleteButton && deleteButtonHighlight}
  <div 
    class="delete-popup"
    style:left="{deleteButtonPosition.x}px"
    style:top="{deleteButtonPosition.y - 50}px"
  >
    <Button
      variant="destructive"
      size="sm"
      onclick={() => handleDeleteHighlight(deleteButtonHighlight.id)}
      class="flex items-center gap-2"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14zM10 11v6M14 11v6"/>
      </svg>
      Delete
    </Button>
  </div>
{/if}

<style>
  .highlighter {
    line-height: 1.6;
    user-select: none;
  }
  
  .word {
    cursor: pointer;
    display: inline;
    padding: 3px 6px;
    border-radius: 4px;
  }
  
  .word.highlighted {
    cursor: pointer;
  }
  
  .selection-word {
    display: inline;
    padding: 3px 6px;
    border-radius: 4px;
    background-color: rgba(156, 163, 175, 0.3);
  }
  
  .delete-popup {
    position: fixed;
    z-index: 1000;
    background: white;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    padding: 4px;
    pointer-events: auto;
  }
</style>