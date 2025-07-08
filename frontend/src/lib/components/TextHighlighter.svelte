<script>
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import TimeGap from "$lib/components/ui/TimeGap.svelte";
  import { DeleteSuggestedHighlight } from "$lib/wailsjs/go/main/App";
  import { HighlightManager } from "./HighlightManager.js";
  // All highlight operations are now handled through HighlightManager

  let {
    text = "",
    words = [],
    highlights = [],
    suggestedHighlights = [],
    onHighlightsChange,
    videoId,
  } = $props();

  // Debug logging for suggested highlights
  $effect(() => {
    console.log("🎨 TextHighlighter: suggestedHighlights changed:", {
      count: suggestedHighlights.length,
      highlights: suggestedHighlights,
    });
    
    // Log word timing information for debugging
    if (words && words.length > 0 && suggestedHighlights.length > 0) {
      console.log("📊 Word timing info (first 10 words):", 
        words.slice(0, 10).map((w, i) => ({
          index: i,
          word: w.word,
          start: w.start,
          end: w.end,
          duration: (w.end - w.start).toFixed(3)
        }))
      );
      
      console.log("📊 Suggested highlight ranges (word indices):", 
        suggestedHighlights.map(s => ({
          id: s.id,
          startIndex: s.start,
          endIndex: s.end,
          wordCount: s.end - s.start + 1,
          text: s.text?.substring(0, 50) + "..."
        }))
      );
      
      // For each suggestion, show which words should be highlighted
      console.log("🎯 Words per suggestion:");
      suggestedHighlights.forEach((s, sIndex) => {
        // Get words in the index range
        const matchingWords = [];
        for (let i = s.start; i <= s.end && i < words.length; i++) {
          matchingWords.push({ ...words[i], index: i });
        }
        
        console.log(`  Suggestion ${sIndex} (indices ${s.start}-${s.end}):`, {
          wordCount: matchingWords.length,
          wordIndices: matchingWords.map(w => w.index),
          firstWord: matchingWords[0]?.word,
          lastWord: matchingWords[matchingWords.length - 1]?.word,
          text: matchingWords.map(w => w.word).join(' ').substring(0, 50) + "..."
        });
      });
    }
  });

  // === CORE STATE ===
  let highlightManager = $state(null);
  let indexHighlights = $state([]);
  
  // Pause detection settings
  const SHOW_ALL_PAUSES = false; // show even normal pauses with subtle indicators
  
  // Initialize highlight manager when words are available
  $effect(() => {
    if (words && words.length > 0 && !highlightManager) {
      highlightManager = new HighlightManager(words);
    }
  });
  
  // Update highlights when they change from props (but not from our own changes)
  $effect(() => {
    if (highlightManager && highlights) {
      // Only update if the highlights are different from what we have
      const currentTimestampHighlights = highlightManager.getTimestampHighlights();
      const highlightsChanged = JSON.stringify(highlights) !== JSON.stringify(currentTimestampHighlights);
      
      if (highlightsChanged) {
        indexHighlights = highlightManager.setHighlights(highlights);
      }
    }
  });
  
  // Debug pause detection when words change
  $effect(() => {
    if (words && words.length > 1) {
      let normalPauses = 0;
      let longPauses = 0;
      let veryLongPauses = 0;
      
      for (let i = 0; i < words.length - 1; i++) {
        const pauseDuration = getPauseDuration(i);
        
        if (pauseDuration >= 1.5) {
          veryLongPauses++;
        } else if (pauseDuration >= 0.8) {
          longPauses++;
        } else {
          normalPauses++;
        }
      }
      
      console.log("⏸️ Pause analysis:", {
        totalWords: words.length,
        normalPauses,
        longPauses,
        veryLongPauses,
        thresholds: {
          long: 0.8,
          veryLong: 1.5
        }
      });
    }
  });

  // === SELECTION STATE ===
  let isSelecting = $state(false);
  let selectionStart = $state(null);
  let selectionEnd = $state(null);
  let selectionAnchor = $state(null); // The fixed point where selection started

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
        word: word,
      }));
    } else {
      displayWords = [];
    }
  }

  // Single initialization function
  function initialize() {
    if (initialized) return;
    initializeDisplayWords();
    initialized = true;
  }
  
  // Calculate pause duration between two consecutive words
  function getPauseDuration(wordIndex) {
    if (!words || words.length === 0 || wordIndex >= words.length - 1) {
      return 0;
    }
    
    const currentWord = words[wordIndex];
    const nextWord = words[wordIndex + 1];
    
    if (!currentWord || !nextWord) {
      return 0;
    }
    
    // Pause is the gap between current word end and next word start
    return nextWord.start - currentWord.end;
  }
  

  function emitChanges(result) {
    if (onHighlightsChange) {
      // Update our local state
      indexHighlights = result.indexHighlights;
      // Notify parent of the change
      onHighlightsChange(result.timestampHighlights);
    }
  }

  function isWordInDragPreview(wordIndex) {
    if (
      !isDragging ||
      !dragTarget ||
      selectionStart === null ||
      selectionEnd === null
    ) {
      return false;
    }

    const { originalHighlight } = dragTarget;

    // Check if word is in the new selection range
    const inNewSelection = wordIndex >= selectionStart && wordIndex <= selectionEnd;
    
    // Check if word is in the original highlight
    const inOriginalHighlight =
      wordIndex >= originalHighlight.start &&
      wordIndex <= originalHighlight.end;

    if (dragMode === "expand") {
      // Show preview for words that would be added (in new selection but not in original)
      return inNewSelection && !inOriginalHighlight;
    } else if (dragMode === "contract") {
      // Show preview for words that would be removed (in original but not in new selection)
      return inOriginalHighlight && !inNewSelection;
    }

    return false;
  }

  // Find suggested highlight for a word index
  function findSuggestedHighlightForWord(wordIndex, suggestions) {
    if (!words || words.length === 0 || wordIndex >= words.length) return null;

    // Suggested highlights use word indices, not timestamps!
    // Check if the current word index falls within any suggestion's index range
    const found = suggestions.find((s) => {
      // Word is highlighted if its index is within the suggestion's start/end indices
      const isInRange = wordIndex >= s.start && wordIndex <= s.end;
      
      // Log detailed info for debugging
      if (wordIndex < 5 && suggestions.length > 0) {
        console.log(`🔍 Word ${wordIndex} index check:`, {
          wordIndex,
          word: words[wordIndex].word,
          suggestionStart: s.start,
          suggestionEnd: s.end,
          isInRange,
        });
      }
      
      return isInRange;
    });

    return found;
  }

  // Handle accepting a suggested highlight
  async function handleAcceptSuggestion(suggestion, event) {
    event.preventDefault();
    event.stopPropagation();

    if (!highlightManager) {
      console.error("Cannot accept suggestion: HighlightManager not initialized");
      return;
    }

    // Suggested highlights already use word indices, not timestamps
    const startIndex = suggestion.start;
    const endIndex = suggestion.end;

    try {
      // Use HighlightManager to create the highlight
      const result = highlightManager.createHighlightFromSelection(startIndex, endIndex);
      emitChanges(result);
      
      // Delete the suggestion from the database since it's now accepted
      if (videoId && suggestion.id) {
        try {
          await DeleteSuggestedHighlight(videoId, suggestion.id);
          console.log("✅ Deleted accepted suggestion from DB:", suggestion.id);
        } catch (error) {
          console.error("Failed to delete accepted suggestion:", error);
        }
      }
    } catch (error) {
      console.error("Failed to accept suggestion:", error);
    }
  }

  // Handle rejecting a suggested highlight
  async function handleRejectSuggestion(suggestion, event) {
    event.preventDefault();
    event.stopPropagation();
    
    if (!videoId || !suggestion.id) {
      console.error("Cannot reject suggestion: missing videoId or suggestion ID");
      return;
    }
    
    try {
      // Delete the specific suggestion from the database
      await DeleteSuggestedHighlight(videoId, suggestion.id);
      
      // Small delay to ensure database operation completes
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Trigger a change event to force parent to reload
      // This ensures the parent's suggestedHighlights array stays in sync
      if (onHighlightsChange) {
        onHighlightsChange(highlights); // Keep current highlights unchanged
      }
    } catch (error) {
      console.error("Failed to reject suggestion:", error);
    }
  }

  // === EVENT HANDLERS ===

  function handleWordMouseDown(wordIndex, event) {
    // Don't interfere with double-click events
    if (event.detail === 2) {
      return;
    }

    const existingHighlight = highlightManager ? highlightManager.findHighlightForWord(wordIndex) : null;

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
        originalHighlight: { ...existingHighlight },
      };

      selectionStart = wordIndex;
      selectionEnd = wordIndex;
      selectionAnchor = wordIndex;
      showDeleteButton = false;

      event.preventDefault();
      event.stopPropagation();
      return;
    }

    // Start new selection
    isSelecting = true;
    selectionStart = wordIndex;
    selectionEnd = wordIndex;
    selectionAnchor = wordIndex;
    showDeleteButton = false;
  }

  function handleWordMouseEnter(wordIndex) {
    if (isSelecting && selectionAnchor !== null) {
      // Update selection dynamically based on current position
      // The selection should be from anchor to current position
      selectionStart = Math.min(selectionAnchor, wordIndex);
      selectionEnd = Math.max(selectionAnchor, wordIndex);
    }

    if (isDragging && dragTarget && selectionAnchor !== null) {
      // Use HighlightManager to calculate drag selection
      const { originalHighlight, isFirstWord, isLastWord } = dragTarget;
      
      if (highlightManager) {
        try {
          const dragResult = highlightManager.calculateDragSelection(
            selectionAnchor, 
            wordIndex, 
            originalHighlight, 
            isFirstWord, 
            isLastWord
          );
          
          selectionStart = dragResult.start;
          selectionEnd = dragResult.end;
          dragMode = dragResult.mode;
        } catch (error) {
          console.warn('Drag calculation error:', error);
          dragMode = null;
        }
      } else {
        dragMode = null;
      }
    }
  }

  function handleWordClick(wordIndex, event) {
    const highlight = highlightManager ? highlightManager.findHighlightForWord(wordIndex) : null;

    if (highlight) {
      const rect = event.target.getBoundingClientRect();
      deleteButtonHighlight = highlight;
      deleteButtonPosition = {
        x: rect.left + rect.width / 2 - 40,
        y: rect.top - 45,
      };
      showDeleteButton = true;
      event.stopPropagation();
    }
  }

  function handleWordDoubleClick(wordIndex, event) {
    // Check if this word is part of a suggested highlight
    const suggestedHighlight = findSuggestedHighlightForWord(
      wordIndex,
      suggestedHighlights
    );
    if (suggestedHighlight) {
      // Accept the suggestion instead of creating a new highlight
      handleAcceptSuggestion(suggestedHighlight, event);
      return;
    }

    // Don't create highlight if word is already highlighted
    if (highlightManager && highlightManager.findHighlightForWord(wordIndex)) {
      return;
    }

    // Use HighlightManager to create single-word highlight
    if (highlightManager) {
      try {
        const result = highlightManager.createHighlightFromSelection(wordIndex, wordIndex, null, true);
        emitChanges(result);
      } catch (error) {
        console.error("Failed to create single-word highlight:", error);
      }
    }

    event.preventDefault();
    event.stopPropagation();
  }

  // Handle keyboard events for accessibility
  function handleWordKeydown(wordIndex, event) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      handleWordClick(wordIndex, event);
    }
  }

  function handleMouseUp() {
    if (
      isDragging &&
      dragTarget &&
      selectionStart !== null &&
      selectionEnd !== null
    ) {
      // Use HighlightManager to update highlight
      if (highlightManager) {
        try {
          console.log("🎯 Applying drag selection:", {
            highlightId: dragTarget.highlightId,
            originalStart: dragTarget.originalHighlight.start,
            originalEnd: dragTarget.originalHighlight.end,
            newStart: selectionStart,
            newEnd: selectionEnd,
            dragMode,
            isFirstWord: dragTarget.isFirstWord,
            isLastWord: dragTarget.isLastWord
          });

          const result = highlightManager.updateHighlightBounds(
            dragTarget.highlightId,
            selectionStart,
            selectionEnd
          );
          emitChanges(result);
        } catch (error) {
          console.error("Failed to update highlight:", error);
        }
      }
    } else if (
      isSelecting &&
      selectionStart !== null &&
      selectionEnd !== null &&
      selectionAnchor !== null
    ) {
      // Use HighlightManager to create new highlight
      const startIndex = selectionStart;
      const endIndex = selectionEnd;

      if (startIndex !== endIndex && highlightManager) {
        try {
          const result = highlightManager.createHighlightFromSelection(startIndex, endIndex);
          emitChanges(result);
        } catch (error) {
          console.error("Failed to create highlight:", error);
        }
      }
    }

    // Reset all states
    isSelecting = false;
    selectionStart = null;
    selectionEnd = null;
    selectionAnchor = null;
    isDragging = false;
    dragTarget = null;
    dragMode = null;
  }

  function handleDeleteHighlight(highlightId) {
    if (highlightManager) {
      try {
        const result = highlightManager.deleteHighlight(highlightId);
        showDeleteButton = false;
        deleteButtonHighlight = null;
        emitChanges(result);
      } catch (error) {
        console.error("Failed to delete highlight:", error);
      }
    }
  }

  // === MOUNT ===

  onMount(() => {
    initialize();

    document.addEventListener("mouseup", handleMouseUp);
    document.addEventListener("click", (e) => {
      if (!e.target.closest(".delete-popup")) {
        showDeleteButton = false;
      }
    });

    return () => {
      document.removeEventListener("mouseup", handleMouseUp);
    };
  });
</script>

<div class="leading-relaxed select-none" class:dragging={isDragging}>
  {#each displayWords as word, wordIndex}
    {@const highlight = highlightManager ? highlightManager.findHighlightForWord(wordIndex) : null}
    {@const suggestedHighlight = findSuggestedHighlightForWord(
      wordIndex,
      suggestedHighlights
    )}
    {@const inSelection = isSelecting && highlightManager ? highlightManager.isWordInSelection(
      wordIndex,
      selectionStart,
      selectionEnd
    ) : false}
    {@const inDragPreview =
      isDragging && dragTarget && isWordInDragPreview(wordIndex)}

    {#if isDragging && dragTarget && highlight && dragTarget.highlightId === highlight.id}
      <!-- Word is part of the highlight being dragged -->
      {@const inNewSelection = wordIndex >= selectionStart && wordIndex <= selectionEnd}
      {#if inNewSelection}
        <!-- Word will remain in the highlight -->
        <span
          class="inline cursor-pointer px-1.5 py-0.5 rounded"
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
      {:else}
        <!-- Word will be removed (contraction preview) -->
        <span
          class="inline cursor-pointer px-1.5 py-0.5 rounded bg-red-300/40 line-through"
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
      {/if}
    {:else if isDragging && dragTarget && wordIndex >= selectionStart && wordIndex <= selectionEnd && !highlight}
      <!-- Drag expansion preview (word will be added) -->
      <span 
        class="inline cursor-pointer px-1.5 py-0.5 rounded bg-blue-300/40"
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
        ondblclick={(e) => handleWordDoubleClick(wordIndex, e)}
        onkeydown={(e) => handleWordKeydown(wordIndex, e)}
        role="button"
        tabindex="0"
        aria-label="Preview text: {word.word}"
      >
        {word.word}
      </span>
    {:else if highlight}
      <!-- Regular highlighted word (not being dragged) -->
      <span
        class="inline cursor-pointer px-1.5 py-0.5 rounded"
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
    {:else if suggestedHighlight}
      <!-- Suggested highlight -->
      <span
        class="inline relative group cursor-pointer px-1.5 py-0.5 rounded border-2 border-dashed transition-all duration-200 hover:opacity-100"
        style:background-color={`${suggestedHighlight.color}40`}
        style:border-color={suggestedHighlight.color}
        class:opacity-70={true}
        onmousedown={(e) => handleWordMouseDown(wordIndex, e)}
        onmouseenter={() => handleWordMouseEnter(wordIndex)}
        onclick={(e) => handleWordClick(wordIndex, e)}
        ondblclick={(e) => handleWordDoubleClick(wordIndex, e)}
        onkeydown={(e) => handleWordKeydown(wordIndex, e)}
        role="button"
        tabindex="0"
        aria-label="Suggested highlight: {word.word}"
      >
        {word.word}

        <!-- Accept/Reject icons at the end of the suggested highlight -->
        {#if words && words.length > 0 && wordIndex == suggestedHighlight.end}
          <span
            class="inline-flex items-center gap-0.5 ml-1 opacity-80 group-hover:opacity-100 transition-opacity"
          >
            <button
              class="w-4 h-4 rounded-full bg-green-500 hover:bg-green-600 text-white flex items-center justify-center transition-all hover:scale-110"
              onclick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                e.stopImmediatePropagation();
                handleAcceptSuggestion(suggestedHighlight, e);
              }}
              onmousedown={(e) => {
                e.preventDefault();
                e.stopPropagation();
                e.stopImmediatePropagation();
              }}
              title="Accept suggestion"
            >
              <svg
                class="w-2.5 h-2.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="3"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </button>
            <button
              class="w-4 h-4 rounded-full bg-red-500 hover:bg-red-600 text-white flex items-center justify-center transition-all hover:scale-110"
              onclick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                e.stopImmediatePropagation();
                handleRejectSuggestion(suggestedHighlight, e);
              }}
              onmousedown={(e) => {
                e.preventDefault();
                e.stopPropagation();
                e.stopImmediatePropagation();
              }}
              title="Reject suggestion"
            >
              <svg
                class="w-2.5 h-2.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="3"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </span>
        {/if}
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

    <!-- Pause indicator between words -->
    {#if wordIndex < displayWords.length - 1}
      {#if words && words.length > 0}
        {@const pauseDuration = getPauseDuration(wordIndex)}
        <TimeGap duration={pauseDuration} showNormal={SHOW_ALL_PAUSES} size="sm" />
      {:else}
        <!-- Fallback space when no timing data -->
        {" "}
      {/if}
    {/if}
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
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14zM10 11v6M14 11v6"
        />
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
