<script>
  import { onMount } from 'svelte';
  import { Button } from "$lib/components/ui/button";
  import { 
    isWordInSelection, 
    findHighlightForWord, 
    checkOverlap,
    addHighlight,
    removeHighlight,
    updateHighlight
  } from './TextHighlighter.utils.js';
  
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
  let dragTarget = $state(null); // { highlightId, wordIndex, isFirstWord, isLastWord, originalHighlight }
  let dragMode = $state(null); // 'expand' | 'contract'
  
  // === UI STATE ===
  let showDeleteButton = $state(false);
  let deleteButtonHighlight = $state(null);
  let deleteButtonPosition = $state({ x: 0, y: 0 });
  
  // === DISPLAY WORDS ===
  let displayWords = $state([]);
  let initialized = $state(false);
  
  // Initialize display words - NO EFFECTS
  function initializeDisplayWords() {
    if (words && words.length > 0) {
      displayWords = words;
    } else if (text) {
      const wordMatches = text.match(/\S+/g) || [];
      displayWords = wordMatches.map((word, index) => ({
        id: index,
        word: word
      }));
    } else {
      displayWords = [];
    }
  }
  
  // Initialize highlights - NO EFFECTS
  function initializeHighlights() {
    if (initialHighlights && initialHighlights.length > 0 && displayWords.length > 0) {
      highlights = initialHighlights.map(h => ({
        ...h,
        // Convert timestamps back to word indices for simple approach
        start: words && words.length > 0 ? findWordIndexByTime(h.start) : 0,
        end: words && words.length > 0 ? findWordIndexByTime(h.end) : 0
      }));
      highlights.forEach(h => usedColors.add(h.color));
    }
  }
  
  // Single initialization function
  function initialize() {
    if (initialized) return;
    initializeDisplayWords();
    initializeHighlights();
    initialized = true;
  }
  
  function findWordIndexByTime(timestamp) {
    if (!words || words.length === 0) return 0;
    
    for (let i = 0; i < words.length; i++) {
      const word = words[i];
      if (word.start <= timestamp && timestamp <= word.end) {
        return i;
      }
    }
    
    // Find closest
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
  
  function emitChanges() {
    if (onHighlightsChange) {
      // Convert indices back to timestamps for storage
      const timestampHighlights = highlights.map(h => ({
        ...h,
        start: words && words.length > 0 ? words[h.start]?.start || 0 : 0,
        end: words && words.length > 0 ? words[h.end]?.end || 0 : 0
      }));
      onHighlightsChange(timestampHighlights);
    }
  }
  
  function isWordInDragPreview(wordIndex) {
    if (!isDragging || !dragTarget || !dragMode || selectionStart === null || selectionEnd === null) {
      return false;
    }
    
    const { originalHighlight } = dragTarget;
    const currentDragPosition = selectionEnd;
    
    if (dragMode === 'expand') {
      // Show preview for words that would be added to the highlight
      const newStart = Math.min(originalHighlight.start, currentDragPosition);
      const newEnd = Math.max(originalHighlight.end, currentDragPosition);
      
      // Only show preview for words outside the original highlight
      const inExpandedArea = wordIndex >= newStart && wordIndex <= newEnd;
      const inOriginalHighlight = wordIndex >= originalHighlight.start && wordIndex <= originalHighlight.end;
      
      return inExpandedArea && !inOriginalHighlight;
    } else if (dragMode === 'contract') {
      // Show preview for words that would be removed from the highlight
      const dragPosition = selectionEnd;
      
      if (dragTarget.isFirstWord) {
        // Dragging first word inward - show words that will be removed from start
        return wordIndex >= originalHighlight.start && wordIndex < Math.min(dragPosition, originalHighlight.end);
      } else if (dragTarget.isLastWord) {
        // Dragging last word inward - show words that will be removed from end
        return wordIndex > Math.max(dragPosition, originalHighlight.start) && wordIndex <= originalHighlight.end;
      }
    }
    
    return false;
  }
  
  // === EVENT HANDLERS ===
  
  function handleWordMouseDown(wordIndex, event) {
    // Don't interfere with double-click events
    if (event.detail === 2) {
      return;
    }

    const existingHighlight = findHighlightForWord(wordIndex, highlights);
    
    if (existingHighlight) {
      // Start drag operation on existing highlight
      const isFirstWord = wordIndex === existingHighlight.start;
      const isLastWord = wordIndex === existingHighlight.end;
      
      isDragging = true;
      dragTarget = {
        highlightId: existingHighlight.id,
        wordIndex,
        isFirstWord,
        isLastWord,
        originalHighlight: { ...existingHighlight }
      };
      
      selectionStart = wordIndex;
      selectionEnd = wordIndex;
      showDeleteButton = false;
      
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    
    // Start new selection
    isSelecting = true;
    selectionStart = wordIndex;
    selectionEnd = wordIndex;
    showDeleteButton = false;
  }
  
  function handleWordMouseEnter(wordIndex) {
    if (isSelecting) {
      selectionEnd = wordIndex;
    }
    
    if (isDragging && dragTarget) {
      selectionEnd = wordIndex;
      
      // Determine drag mode based on whether we're inside or outside the original highlight
      const { originalHighlight, isFirstWord, isLastWord } = dragTarget;
      const insideOriginalHighlight = wordIndex >= originalHighlight.start && wordIndex <= originalHighlight.end;
      
      if (insideOriginalHighlight && (isFirstWord || isLastWord)) {
        // Contraction: dragging first/last word over existing highlight
        dragMode = 'contract';
      } else if (!insideOriginalHighlight) {
        // Expansion: dragging outside the original highlight
        dragMode = 'expand';
      } else {
        // Dragging middle word within highlight - no operation
        dragMode = null;
      }
    }
  }
  
  function handleWordClick(wordIndex, event) {
    const highlight = findHighlightForWord(wordIndex, highlights);
    
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

  function handleWordDoubleClick(wordIndex, event) {
    // Don't create highlight if word is already highlighted
    if (findHighlightForWord(wordIndex, highlights)) {
      return;
    }

    // Check if single word would overlap with existing highlights
    if (!checkOverlap(wordIndex, wordIndex, highlights)) {
      const result = addHighlight(highlights, wordIndex, wordIndex, usedColors);
      highlights = result.highlights;
      usedColors.add(result.newHighlight.color);
      emitChanges();
    }

    event.preventDefault();
    event.stopPropagation();
  }

  // Handle keyboard events for accessibility
  function handleWordKeydown(wordIndex, event) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleWordClick(wordIndex, event);
    }
  }
  
  
  function handleMouseUp() {
    if (isDragging && dragTarget && selectionStart !== null && selectionEnd !== null && dragMode) {
      // Apply word-based expansion/contraction
      const currentHighlight = highlights.find(h => h.id === dragTarget.highlightId);
      if (currentHighlight) {
        const { originalHighlight } = dragTarget;
        let newStartIndex = originalHighlight.start;
        let newEndIndex = originalHighlight.end;
        
        if (dragMode === 'expand') {
          // Expansion: include all words from original highlight to drag position
          const dragPosition = selectionEnd;
          newStartIndex = Math.min(originalHighlight.start, dragPosition);
          newEndIndex = Math.max(originalHighlight.end, dragPosition);
        } else if (dragMode === 'contract') {
          // Contraction: remove words based on drag direction
          const dragPosition = selectionEnd;
          
          if (dragTarget.isFirstWord) {
            // Dragging first word inward - move start position
            newStartIndex = Math.min(dragPosition, originalHighlight.end);
          } else if (dragTarget.isLastWord) {
            // Dragging last word inward - move end position
            newEndIndex = Math.max(dragPosition, originalHighlight.start);
          }
        }
        
        // Ensure valid bounds
        if (newStartIndex <= newEndIndex) {
          // Check if new bounds would overlap with other highlights
          if (!checkOverlap(newStartIndex, newEndIndex, highlights, dragTarget.highlightId)) {
            const result = updateHighlight(highlights, dragTarget.highlightId, newStartIndex, newEndIndex);
            highlights = result;
            emitChanges();
          }
        }
      }
    } else if (isSelecting && selectionStart !== null && selectionEnd !== null) {
      // Create new highlight
      const startIndex = Math.min(selectionStart, selectionEnd);
      const endIndex = Math.max(selectionStart, selectionEnd);
      
      if (startIndex !== endIndex) {
        if (!checkOverlap(startIndex, endIndex, highlights)) {
          const result = addHighlight(highlights, startIndex, endIndex, usedColors);
          highlights = result.highlights;
          usedColors.add(result.newHighlight.color);
          emitChanges();
        }
      }
    }
    
    // Reset all states
    isSelecting = false;
    selectionStart = null;
    selectionEnd = null;
    isDragging = false;
    dragTarget = null;
    dragMode = null;
  }
  
  function handleDeleteHighlight(highlightId) {
    const highlight = highlights.find(h => h.id === highlightId);
    if (highlight) {
      usedColors.delete(highlight.color);
    }
    highlights = removeHighlight(highlights, highlightId);
    showDeleteButton = false;
    deleteButtonHighlight = null;
    emitChanges();
  }
  
  // === MOUNT ===
  
  onMount(() => {
    initialize();
    
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

<div class="leading-relaxed select-none" class:dragging={isDragging}>
  {#each displayWords as word, wordIndex}
    {@const highlight = findHighlightForWord(wordIndex, highlights)}
    {@const inSelection = isWordInSelection(wordIndex, selectionStart, selectionEnd, isSelecting)}
    {@const inDragPreview = isDragging && dragTarget && isWordInDragPreview(wordIndex)}
    
    {#if inDragPreview}
      <!-- Drag expansion/contraction preview -->
      <span class="inline px-1.5 py-0.5 rounded bg-blue-300/40">
        {word.word}
      </span>
    {:else if highlight}
      <!-- Highlighted word -->
      <span 
        class="inline cursor-pointer px-1.5 py-0.5 rounded"
        class:opacity-80={isDragging && dragTarget && dragTarget.highlightId === highlight.id}
        style:background-color={highlight.color}
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
        ondblclick={(e) => handleWordDoubleClick(wordIndex, e)}
        onkeydown={(e) => handleWordKeydown(wordIndex, e)}
        role="button"
        tabindex="0"
        aria-label="Highlighted text: {word.word}"
      >
        {word.word}
      </span>
    {:else if inSelection}
      <!-- Selection preview -->
      <span class="inline px-1.5 py-0.5 rounded bg-gray-400/30">
        {word.word}
      </span>
    {:else}
      <!-- Regular word -->
      <span
        class="inline cursor-pointer px-1.5 py-0.5 rounded"
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
        ondblclick={(e) => handleWordDoubleClick(wordIndex, e)}
        onkeydown={(e) => handleWordKeydown(wordIndex, e)}
        role="button"
        tabindex="0"
        aria-label="Text: {word.word}"
      >
        {word.word}
      </span>
    {/if}
    
    <!-- Space between words -->
    {#if wordIndex < displayWords.length - 1}{' '}{/if}
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