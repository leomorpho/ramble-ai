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
  let dragTarget = $state(null); // { highlightId, type: 'start'|'end', originalBounds }
  
  // === UI STATE ===
  let showDeleteButton = $state(false);
  let deleteButtonHighlight = $state(null);
  let deleteButtonPosition = $state({ x: 0, y: 0 });
  let showHandles = $state(null);
  
  // === PURE FUNCTIONS (TESTABLE) ===
  
  function generateUniqueColor() {
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
  
  function createHighlight(start, end) {
    return {
      id: `highlight_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      start,
      end,
      color: generateUniqueColor()
    };
  }
  
  function isWordInHighlight(wordIndex, highlight) {
    return wordIndex >= highlight.start && wordIndex <= highlight.end;
  }
  
  function isWordInSelection(wordIndex) {
    if (!isSelecting || selectionStart === null || selectionEnd === null) return false;
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);
    return wordIndex >= start && wordIndex <= end;
  }
  
  function findHighlightForWord(wordIndex) {
    return highlights.find(h => isWordInHighlight(wordIndex, h));
  }
  
  function checkOverlap(start, end, excludeId = null) {
    return highlights.some(h => 
      h.id !== excludeId && 
      start <= h.end && end >= h.start
    );
  }
  
  function calculateTimestamps(startIndex, endIndex) {
    if (!words || words.length === 0) return { start: 0, end: 0 };
    
    const startWord = words[Math.max(0, Math.min(startIndex, words.length - 1))];
    const endWord = words[Math.max(0, Math.min(endIndex, words.length - 1))];
    
    return {
      start: startWord.start || 0,
      end: endWord.end || 0
    };
  }
  
  function findWordByTimestamp(timestamp) {
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
  
  // === ACTIONS ===
  
  function addHighlight(start, end) {
    const newHighlight = createHighlight(start, end);
    highlights = [...highlights, newHighlight];
    usedColors.add(newHighlight.color);
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
    emitChanges();
  }
  
  function updateHighlight(highlightId, newStart, newEnd) {
    highlights = highlights.map(h => 
      h.id === highlightId 
        ? { ...h, start: newStart, end: newEnd }
        : h
    );
  }
  
  function emitChanges() {
    if (onHighlightsChange) {
      const timestampHighlights = highlights.map(h => {
        const timestamps = calculateTimestamps(h.start, h.end);
        return {
          id: h.id,
          start: timestamps.start,
          end: timestamps.end,
          color: h.color
        };
      });
      onHighlightsChange(timestampHighlights);
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
    
    // Start selection
    isSelecting = true;
    selectionStart = wordIndex;
    selectionEnd = wordIndex;
    showDeleteButton = false;
    showHandles = null;
  }
  
  function handleWordMouseEnter(wordIndex) {
    if (isSelecting) {
      selectionEnd = wordIndex;
    }
    
    if (isDragging && dragTarget) {
      const currentHighlight = highlights.find(h => h.id === dragTarget.highlightId);
      if (!currentHighlight) return;
      
      const newStart = dragTarget.type === 'start' ? wordIndex : currentHighlight.start;
      const newEnd = dragTarget.type === 'end' ? wordIndex : currentHighlight.end;
      
      if (newStart <= newEnd && !checkOverlap(newStart, newEnd, dragTarget.highlightId)) {
        updateHighlight(dragTarget.highlightId, newStart, newEnd);
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
      const start = Math.min(selectionStart, selectionEnd);
      const end = Math.max(selectionStart, selectionEnd);
      
      if (start !== end && !checkOverlap(start, end)) {
        addHighlight(start, end);
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
  
  function handleHandleDrag(highlightId, type, event) {
    const highlight = highlights.find(h => h.id === highlightId);
    if (!highlight) return;
    
    isDragging = true;
    dragTarget = {
      highlightId,
      type,
      originalBounds: { start: highlight.start, end: highlight.end }
    };
    
    showDeleteButton = false;
    event.stopPropagation();
    event.preventDefault();
  }
  
  function handleDeleteHighlight(highlightId) {
    removeHighlight(highlightId);
    showDeleteButton = false;
    deleteButtonHighlight = null;
  }
  
  // === INITIALIZATION ===
  
  $effect(() => {
    if (initialHighlights && initialHighlights.length > 0) {
      const convertedHighlights = initialHighlights
        .map(h => ({
          id: h.id,
          start: findWordByTimestamp(h.start),
          end: findWordByTimestamp(h.end),
          color: h.color
        }))
        .filter(h => h.start !== -1 && h.end !== -1);
      
      highlights = convertedHighlights;
      convertedHighlights.forEach(h => usedColors.add(h.color));
    }
  });
  
  // === WORD PROCESSING ===
  
  let displayWords = $state([]);
  let groupedElements = $state([]);
  
  $effect(() => {
    if (words && words.length > 0) {
      displayWords = words;
    } else if (text) {
      const wordMatches = text.match(/\S+/g) || [];
      displayWords = wordMatches.map((word, index) => ({
        id: index,
        word: word
      }));
    }
  });
  
  // Group words and highlights for better rendering
  $effect(() => {
    const groups = [];
    let i = 0;
    
    while (i < displayWords.length) {
      const highlight = findHighlightForWord(i);
      
      if (highlight) {
        // Start of a highlight group
        const group = {
          type: 'highlight',
          highlight: highlight,
          words: [],
          startIndex: i
        };
        
        // Collect all consecutive words in this highlight
        while (i < displayWords.length && findHighlightForWord(i)?.id === highlight.id) {
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
    
    groupedElements = groups;
  });
  
  // === GLOBAL HANDLERS ===
  
  onMount(() => {
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

<div class="highlighter">
  {#each groupedElements as group, groupIndex}
    {#if group.type === 'highlight'}
      <!-- Highlight group with edge handles -->
      {@const showThisHandle = showHandles === group.highlight.id}
      
      <!-- Start handle at left edge -->
      <span
        class="drag-handle drag-handle-start"
        class:visible={showThisHandle}
        onmousedown={(e) => handleHandleDrag(group.highlight.id, 'start', e)}
        onmouseenter={() => showHandles = group.highlight.id}
        onmouseleave={() => showHandles = null}
        title="Drag to resize highlight"
      ></span>
      
      <!-- Highlight group -->
      <span 
        class="highlight-group"
        style:background-color={group.highlight.color}
        onmousedown={(e) => handleWordMouseDown(group.startIndex, e)}
        onclick={(e) => handleWordClick(group.startIndex, e)}
        onmouseenter={() => showHandles = group.highlight.id}
        onmouseleave={() => showHandles = null}
      >
        {#each group.words as { word, index }, wordIndex}
          {@const inSelection = isWordInSelection(index)}
          <span
            class="word highlighted"
            class:selecting={inSelection}
            onmouseenter={() => handleWordMouseEnter(index)}
          >
            {word.word}
          </span>
          {#if wordIndex < group.words.length - 1}{' '}{/if}
        {/each}
      </span>
      
      <!-- End handle at right edge -->
      <span
        class="drag-handle drag-handle-end"
        class:visible={showThisHandle}
        onmousedown={(e) => handleHandleDrag(group.highlight.id, 'end', e)}
        onmouseenter={() => showHandles = group.highlight.id}
        onmouseleave={() => showHandles = null}
        title="Drag to resize highlight"
      ></span>
      
    {:else if isWordInSelection(group.index)}
      <!-- Selection preview -->
      <span class="selection-word">
        {group.word.word}
      </span>
      
    {:else}
      <!-- Regular word -->
      <span
        class="word"
        onmousedown={(e) => handleWordMouseDown(group.index, e)}
        onmouseenter={() => handleWordMouseEnter(group.index)}
        onclick={(e) => handleWordClick(group.index, e)}
      >
        {group.word.word}
      </span>
    {/if}
    
    <!-- Space between groups -->
    {#if groupIndex < groupedElements.length - 1}{' '}{/if}
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
  }
  
  .highlight-word {
    display: inline;
    padding: 3px 6px;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .highlight-group {
    display: inline;
    padding: 3px 6px;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .selection-word {
    display: inline;
    padding: 3px 6px;
    border-radius: 4px;
    background-color: rgba(156, 163, 175, 0.3);
  }
  
  .drag-handle {
    display: inline-block;
    width: 4px;
    height: 1.2em;
    cursor: ew-resize;
    opacity: 0;
    transition: opacity 0.3s ease, background-color 0.2s ease;
    background-color: rgba(0, 0, 0, 0.6);
    border-radius: 2px;
    vertical-align: baseline;
  }
  
  .drag-handle:hover,
  .drag-handle.visible {
    opacity: 1;
  }
  
  .drag-handle:hover {
    background-color: rgba(0, 0, 0, 0.8);
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