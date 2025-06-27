<script>
  import { onMount } from 'svelte';
  import { Button } from "$lib/components/ui/button";
  
  let { text = '', words = [], initialHighlights = [], onHighlightsChange } = $props();
  
  // State for highlights
  let highlights = $state([]);
  
  // Initialize highlights from props
  $effect(() => {
    if (initialHighlights && initialHighlights.length > 0) {
      // Convert timestamp-based highlights to word-index-based highlights
      const convertedHighlights = initialHighlights.map(h => {
        const startIndex = findWordIndexByTimestamp(h.start);
        const endIndex = findWordIndexByTimestamp(h.end);
        return {
          id: h.id,
          start: startIndex,
          end: endIndex,
          color: h.color,
          timestampStart: h.start,
          timestampEnd: h.end
        };
      }).filter(h => h.start !== -1 && h.end !== -1);
      
      highlights = convertedHighlights;
      // Update colorIndex to continue from where we left off
      if (convertedHighlights.length > 0) {
        colorIndex = Math.max(...convertedHighlights.map(h => colors.indexOf(h.color))) + 1;
      }
    }
  });
  let isSelecting = $state(false);
  let selectionStart = $state(null);
  let selectionEnd = $state(null);
  let highlightId = $state(0);
  let showDeleteButton = $state(false);
  let deleteButtonHighlight = $state(null);
  let deleteButtonPosition = $state({ x: 0, y: 0 });
  
  // State for resizing handles
  let isDragging = $state(false);
  let dragHighlight = $state(null);
  let dragType = $state(null); // 'start' or 'end'
  let originalHighlight = $state(null);
  let hoveredHighlight = $state(null);
  
  // Color palette
  const colors = ['#ffeb3b', '#81c784', '#64b5f6', '#ff8a65', '#f06292'];
  let colorIndex = $state(0);
  
  // If no words provided, create simple word array from text
  let displayWords = $state([]);
  let groupedElements = $state([]);
  
  $effect(() => {
    if (words && words.length > 0) {
      displayWords = words;
    } else if (text) {
      // Simple word splitting
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
      const highlight = getWordHighlight(i);
      
      if (highlight) {
        // Start of a highlight group
        const group = {
          type: 'highlight',
          highlight: highlight,
          words: [],
          startIndex: i
        };
        
        // Collect all consecutive words with the same highlight
        while (i < displayWords.length && getWordHighlight(i)?.id === highlight.id) {
          group.words.push({ word: displayWords[i], index: i });
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
  
  function getWordHighlight(index) {
    return highlights.find(h => index >= h.start && index <= h.end);
  }
  
  function findWordIndexByTimestamp(timestamp) {
    if (!words || words.length === 0) return -1;
    
    // Find the word whose timestamp range contains the given timestamp
    for (let i = 0; i < words.length; i++) {
      const word = words[i];
      if (word.start <= timestamp && timestamp <= word.end) {
        return i;
      }
    }
    
    // If no exact match, find the closest word
    let closestIndex = -1;
    let minDistance = Infinity;
    
    for (let i = 0; i < words.length; i++) {
      const word = words[i];
      const distance = Math.min(
        Math.abs(word.start - timestamp),
        Math.abs(word.end - timestamp)
      );
      
      if (distance < minDistance) {
        minDistance = distance;
        closestIndex = i;
      }
    }
    
    return closestIndex;
  }
  
  function calculateTimestamps(startIndex, endIndex) {
    if (!words || words.length === 0) {
      return { start: 0, end: 0 };
    }
    
    const startWord = words[Math.max(0, Math.min(startIndex, words.length - 1))];
    const endWord = words[Math.max(0, Math.min(endIndex, words.length - 1))];
    
    return {
      start: startWord.start || 0,
      end: endWord.end || 0
    };
  }
  
  function emitHighlightsChange() {
    if (onHighlightsChange) {
      // Convert highlights back to timestamp-based format
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
  
  function isInSelection(index) {
    if (!isSelecting || selectionStart === null || selectionEnd === null) return false;
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);
    return index >= start && index <= end;
  }
  
  function handleMouseDown(index, event) {
    const highlight = getWordHighlight(index);
    
    if (highlight) {
      // Don't start selection on highlighted words
      event.preventDefault();
    } else {
      // Start new selection
      isSelecting = true;
      selectionStart = index;
      selectionEnd = index;
      showDeleteButton = false;
    }
  }
  
  function handleClick(index, event) {
    const highlight = getWordHighlight(index);
    
    if (highlight) {
      // Get the bounding rect of the clicked element
      const rect = event.target.getBoundingClientRect();
      
      // Position delete button just above the clicked word
      deleteButtonHighlight = highlight;
      deleteButtonPosition = { 
        x: rect.left + (rect.width / 2) - 40, // Center horizontally, offset for button width
        y: rect.top - 45 // Position above the text
      };
      showDeleteButton = true;
      event.stopPropagation();
    }
  }
  
  function handleMouseEnter(index) {
    if (isSelecting) {
      selectionEnd = index;
    }
  }
  
  function handleMouseUp() {
    if (isSelecting && selectionStart !== null && selectionEnd !== null) {
      const start = Math.min(selectionStart, selectionEnd);
      const end = Math.max(selectionStart, selectionEnd);
      
      // Only create highlight if more than just a click (start != end)
      if (start !== end) {
        // Check for overlap
        const hasOverlap = highlights.some(h => 
          (start <= h.end && end >= h.start)
        );
        
        if (!hasOverlap) {
          const newHighlight = {
            id: String(highlightId++),
            start,
            end,
            color: colors[colorIndex % colors.length]
          };
          
          highlights = [...highlights, newHighlight];
          colorIndex++;
          
          // Emit changes
          emitHighlightsChange();
        }
      }
    }
    
    isSelecting = false;
    selectionStart = null;
    selectionEnd = null;
  }
  
  function deleteHighlight(highlightId) {
    highlights = highlights.filter(h => h.id !== highlightId);
    showDeleteButton = false;
    deleteButtonHighlight = null;
    
    // Emit changes
    emitHighlightsChange();
  }
  
  function startDrag(highlight, type, event) {
    isDragging = true;
    dragHighlight = highlight;
    dragType = type;
    originalHighlight = { ...highlight };
    showDeleteButton = false;
    event.stopPropagation();
    event.preventDefault();
  }
  
  function handleDragOver(index) {
    if (!isDragging || !dragHighlight) return;
    
    const newStart = dragType === 'start' ? index : dragHighlight.start;
    const newEnd = dragType === 'end' ? index : dragHighlight.end;
    
    // Validate the new range
    if (newStart > newEnd) return;
    
    // Check for overlaps with other highlights
    const hasOverlap = highlights.some(h => 
      h.id !== dragHighlight.id && 
      (newStart <= h.end && newEnd >= h.start)
    );
    
    if (!hasOverlap) {
      // Update the highlight
      highlights = highlights.map(h => 
        h.id === dragHighlight.id 
          ? { ...h, start: newStart, end: newEnd }
          : h
      );
      dragHighlight = { ...dragHighlight, start: newStart, end: newEnd };
      
      // Emit changes
      emitHighlightsChange();
    }
  }
  
  function stopDrag() {
    isDragging = false;
    dragHighlight = null;
    dragType = null;
    originalHighlight = null;
  }
  
  // Global mouse up handler
  onMount(() => {
    const handleGlobalMouseUp = () => {
      if (isSelecting) {
        handleMouseUp();
      }
      if (isDragging) {
        stopDrag();
      }
    };
    
    const handleGlobalClick = (e) => {
      // Hide delete button if clicking outside
      if (!e.target.closest('.delete-popup')) {
        showDeleteButton = false;
      }
    };
    
    document.addEventListener('mouseup', handleGlobalMouseUp);
    document.addEventListener('click', handleGlobalClick);
    
    return () => {
      document.removeEventListener('mouseup', handleGlobalMouseUp);
      document.removeEventListener('click', handleGlobalClick);
    };
  });
</script>

<div class="highlighter">
  {#each groupedElements as group, groupIndex}
    {#if group.type === 'highlight'}
      <!-- Highlight group - all words together -->
      <span 
        class="highlight-group"
        style:background-color={group.highlight.color}
        onmousedown={(e) => handleMouseDown(group.startIndex, e)}
        onclick={(e) => handleClick(group.startIndex, e)}
      >
        <!-- Start handle -->
        <span
          class="drag-handle drag-handle-start"
          onmousedown={(e) => startDrag(group.highlight, 'start', e)}
          title="Drag to resize highlight"
        ></span>
        
        {#each group.words as { word, index }, wordIndex}
          {@const inSelection = isInSelection(index)}
          <span
            class="word highlighted"
            class:selecting={inSelection}
            onmouseenter={() => {
              handleMouseEnter(index);
              handleDragOver(index);
            }}
          >
            {word.word}
          </span>
          {#if wordIndex < group.words.length - 1}{' '}{/if}
        {/each}
        
        <!-- End handle -->
        <span
          class="drag-handle drag-handle-end"
          onmousedown={(e) => startDrag(group.highlight, 'end', e)}
          title="Drag to resize highlight"
        ></span>
      </span>
    {:else}
      <!-- Regular word -->
      {@const inSelection = isInSelection(group.index)}
      <span
        class="word"
        class:selecting={inSelection}
        onmousedown={(e) => handleMouseDown(group.index, e)}
        onmouseenter={() => {
          handleMouseEnter(group.index);
          handleDragOver(group.index);
        }}
        onclick={(e) => handleClick(group.index, e)}
        style:background-color={inSelection ? 'rgba(100, 181, 246, 0.3)' : ''}
      >
        {group.word.word}
      </span>
    {/if}
    
    <!-- Add space between groups -->
    {#if groupIndex < groupedElements.length - 1}{' '}{/if}
  {/each}
</div>

<!-- Simple delete button popup -->
{#if showDeleteButton && deleteButtonHighlight}
  <div 
    class="delete-popup"
    style:left="{deleteButtonPosition.x}px"
    style:top="{deleteButtonPosition.y - 50}px"
  >
    <Button
      variant="destructive"
      size="sm"
      onclick={() => deleteHighlight(deleteButtonHighlight.id)}
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
    position: relative;
  }
  
  .highlight-group {
    display: inline;
    position: relative;
    padding: 3px 6px;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .highlight-group:hover .drag-handle {
    opacity: 1;
  }
  
  .word.highlighted {
    display: inline;
    padding: 0;
  }
  
  .word.selecting {
    padding: 3px 6px;
    border-radius: 4px;
    background-color: rgba(100, 181, 246, 0.3);
    transition: background-color 0.2s ease, padding 0.2s ease, transform 0.2s ease;
  }
  
  
  .drag-handle {
    position: absolute;
    width: 12px;
    height: 100%;
    top: 0;
    cursor: ew-resize;
    opacity: 0;
    transition: opacity 0.3s ease, transform 0.2s ease;
    background-color: rgba(0, 0, 0, 0.4);
    border-radius: 2px;
    transform: scale(1);
  }
  
  .drag-handle:hover {
    background-color: rgba(0, 0, 0, 0.6);
    transform: scale(1.1);
  }
  
  .drag-handle-start {
    left: -6px;
  }
  
  .drag-handle-end {
    right: -6px;
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