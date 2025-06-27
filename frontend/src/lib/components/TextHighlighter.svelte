<script>
  let { text = '', words = [] } = $props();
  
  // State for highlights
  let highlights = $state([]);
  let isSelecting = $state(false);
  let selectionStart = $state(null);
  let selectionEnd = $state(null);
  let highlightId = $state(0);
  
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
  
  function startSelection(index) {
    if (!getWordHighlight(index)) {
      isSelecting = true;
      selectionStart = index;
      selectionEnd = index;
    }
  }
  
  function updateSelection(index) {
    if (isSelecting) {
      selectionEnd = index;
    }
  }
  
  function finishSelection() {
    if (isSelecting && selectionStart !== null && selectionEnd !== null) {
      const start = Math.min(selectionStart, selectionEnd);
      const end = Math.max(selectionStart, selectionEnd);
      
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
    
    isSelecting = false;
    selectionStart = null;
    selectionEnd = null;
  }
</script>

<div class="highlighter">
  {#each displayWords as word, index}
    {@const highlight = getWordHighlight(index)}
    {@const inSelection = isInSelection(index)}
    
    <span
      class="word"
      class:highlighted={!!highlight}
      class:selecting={inSelection}
      style:background-color={highlight?.color || ''}
      onmousedown={() => startSelection(index)}
      onmouseenter={() => updateSelection(index)}
      onmouseup={finishSelection}
    >
      {word.word}
    </span>
    {' '}
  {/each}
</div>

<style>
  .highlighter {
    line-height: 1.6;
    user-select: none;
  }
  
  .word {
    cursor: pointer;
    padding: 2px;
    border-radius: 2px;
  }
  
  .word.highlighted {
    padding: 2px 4px;
  }
  
  .word.selecting {
    background-color: rgba(100, 181, 246, 0.3) !important;
  }
</style>