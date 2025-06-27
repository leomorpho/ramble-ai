<script>
  import { onMount } from 'svelte';
  import { Button } from "$lib/components/ui/button";
  
  let { text = '', words = [] } = $props();
  
  // State for highlights
  let highlights = $state([]);
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
  
  function getWordHighlight(index) {
    return highlights.find(h => index >= h.start && index <= h.end);
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
    
    // Set hovered highlight for showing handles
    const highlight = getWordHighlight(index);
    if (highlight) {
      hoveredHighlight = highlight;
    }
  }
  
  function handleMouseLeave(index) {
    // Only clear hover if we're not moving to another word in the same highlight
    const highlight = getWordHighlight(index);
    if (highlight === hoveredHighlight) {
      // Add a small delay to prevent flickering when moving between words
      setTimeout(() => {
        // Check if we're still not hovering over this highlight
        hoveredHighlight = null;
      }, 50);
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
          highlights = [...highlights, {
            id: highlightId++,
            start,
            end,
            color: colors[colorIndex % colors.length]
          }];
          colorIndex++;
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
  {#each displayWords as word, index}
    {@const highlight = getWordHighlight(index)}
    {@const inSelection = isInSelection(index)}
    {@const isHighlightStart = highlight && (index === 0 || getWordHighlight(index - 1)?.id !== highlight.id)}
    {@const isHighlightEnd = highlight && (index === displayWords.length - 1 || getWordHighlight(index + 1)?.id !== highlight.id)}
    {@const isSelectionStart = inSelection && (index === 0 || !isInSelection(index - 1))}
    {@const isSelectionEnd = inSelection && (index === displayWords.length - 1 || !isInSelection(index + 1))}
    
    <span
      class="word"
      class:highlighted={!!highlight}
      class:selecting={inSelection}
      class:highlight-start={isHighlightStart}
      class:highlight-end={isHighlightEnd}
      class:highlight-middle={highlight && !isHighlightStart && !isHighlightEnd}
      class:selection-start={isSelectionStart}
      class:selection-end={isSelectionEnd}
      class:selection-middle={inSelection && !isSelectionStart && !isSelectionEnd}
      style:background-color={highlight?.color || (inSelection ? 'rgba(100, 181, 246, 0.3)' : '')}
      onmousedown={(e) => handleMouseDown(index, e)}
      onmouseenter={() => {
        handleMouseEnter(index);
        handleDragOver(index);
      }}
      onmouseleave={() => handleMouseLeave(index)}
      onclick={(e) => handleClick(index, e)}
    >
      {word.word}
      
      <!-- Start handle for dragging -->
      {#if isHighlightStart && (hoveredHighlight?.id === highlight.id || isDragging)}
        <span
          class="drag-handle drag-handle-start"
          onmousedown={(e) => startDrag(highlight, 'start', e)}
          title="Drag to resize highlight"
        ></span>
      {/if}
      
      <!-- End handle for dragging -->
      {#if isHighlightEnd && (hoveredHighlight?.id === highlight.id || isDragging)}
        <span
          class="drag-handle drag-handle-end"
          onmousedown={(e) => startDrag(highlight, 'end', e)}
          title="Drag to resize highlight"
        ></span>
      {/if}
    </span>
    
    <!-- Always add regular space -->
    {#if index < displayWords.length - 1}{' '}{/if}
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
  
  .word.highlighted, .word.selecting {
    padding: 3px 0;
    border-radius: 0;
    position: relative;
    transition: background-color 0.2s ease, padding 0.2s ease;
  }
  
  /* Extend background to cover the space after each highlighted word */
  .word.highlighted:not(.highlight-end), .word.selecting:not(.selection-end) {
    padding-right: 1ch; /* Extend padding to cover the space */
    margin-right: -1ch; /* Pull back to not affect layout */
    transition: padding 0.2s ease, margin 0.2s ease;
  }
  
  .word.highlight-start, .word.selection-start {
    border-radius: 4px 0 0 4px;
    padding-left: 6px;
    transition: border-radius 0.2s ease, padding 0.2s ease;
  }
  
  .word.highlight-end, .word.selection-end {
    border-radius: 0 4px 4px 0;
    padding-right: 6px;
    margin-right: 0; /* Reset margin for end words */
    transition: border-radius 0.2s ease, padding 0.2s ease, margin 0.2s ease;
  }
  
  .word.highlight-start.highlight-end, .word.selection-start.selection-end {
    border-radius: 4px;
    padding: 3px 6px;
    margin-right: 0;
    transition: border-radius 0.2s ease, padding 0.2s ease, margin 0.2s ease;
  }
  
  .highlighted-space {
    cursor: pointer;
    padding: 3px 0;
    display: inline;
  }
  
  .drag-handle {
    position: absolute;
    width: 8px;
    height: 100%;
    top: 0;
    cursor: ew-resize;
    opacity: 1;
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
    left: -4px;
  }
  
  .drag-handle-end {
    right: -4px;
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