<script>
  import HighlightMenu from "./HighlightMenu.svelte";
  import TimeGap from "./ui/TimeGap.svelte";

  let {
    highlight,
    index,
    isSelected = false,
    isDragging = false,
    isBeingDragged = false,
    showDropIndicatorBefore = false,
    enableDrag = true,
    enableEdit = true,
    onSelect = () => {},
    onDragStart = () => {},
    onDragEnd = () => {},
    onDragOver = () => {},
    onDrop = () => {},
    onEdit = () => {},
    onDelete = () => {},
    popoverOpen = false,
    onPopoverOpenChange = () => {},
    words = [], // Transcription words for this highlight
  } = $props();
  
  // Function to calculate pause duration between two consecutive words
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

  // Get the highlight color with fallback
  function getHighlightColor() {
    const color = highlight?.color;
    // Return the color if it's a valid string, otherwise use a default
    if (color && typeof color === 'string' && color.trim() !== '') {
      return color;
    }
    // Default fallback color
    return 'rgba(59, 130, 246, 0.3)'; // Blue with opacity
  }
</script>

<!-- Drop indicator before this highlight -->
{#if showDropIndicatorBefore}
  <span class="drop-indicator">|</span>
{/if}

<!-- Highlight as inline text span with embedded eye icon -->
<span
  class="highlight-span  
         {isSelected ? 'highlight-selected' : ''}
         {isBeingDragged ? 'highlight-dragging' : ''}
         {!enableDrag ? 'highlight-non-draggable' : ''}"
  style="background-color: {getHighlightColor()};"
  draggable={enableDrag}
  ondragstart={(e) => onDragStart(e, highlight, index)}
  ondragend={onDragEnd}
  onclick={(e) => onSelect(e, highlight)}
  ondragover={(e) => onDragOver(e, index)}
  ondrop={(e) => onDrop(e, index)}
  role="button"
  tabindex="0"
>
  {#if words && words.length > 0}
    <!-- Show individual words with pauses -->
    {#each words as word, wordIndex}
      <span class="inline">{word.word}</span>
      
      <!-- Show pause between words -->
      {#if wordIndex < words.length - 1}
        {@const pauseDuration = getPauseDuration(wordIndex)}
        {#if pauseDuration > 0}
          <TimeGap duration={pauseDuration} showNormal={false} size="xs" />
        {/if}
      {/if}
    {/each}
  {:else}
    <!-- Fallback to highlight text if no words available -->
    {highlight.text || highlight.videoClipName}
  {/if}
  
  <!-- Eye icon inside highlight -->
  {#if enableEdit}
    <span class="inline-flex items-center ml-1">
      <HighlightMenu
        {highlight}
        {onEdit}
        {onDelete}
        {popoverOpen}
        {onPopoverOpenChange}
        iconSize="w-3 h-3"
        triggerSize="w-5 h-5"
      />
    </span>
  {/if}
</span>

<style>
  /* Natural text flow highlight spans */
  .highlight-span {
    display: inline;
    padding: 2px 2px;
    margin: 0px 2px;
    border-radius: 3px;
    cursor: move;
    user-select: none;
    /* Only transition specific properties to avoid WebView rendering issues */
    transition: opacity 0.15s ease, box-shadow 0.15s ease;
    font-weight: 500;
    position: relative;
    color: hsl(var(--foreground));
  }

  /* Non-draggable state */
  .highlight-non-draggable {
    cursor: default;
  }

  .highlight-span:hover {
    /* Use box-shadow instead of filter/transform for better WebView compatibility */
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 0 0 1px rgba(255, 255, 255, 0.1) inset;
    opacity: 0.95;
  }

  .highlight-span:active {
    opacity: 0.9;
  }

  /* Selection state for highlights */
  .highlight-selected {
    box-shadow: 0 0 0 2px currentColor;
    opacity: 1;
  }

  /* Dragging state */
  .highlight-dragging {
    opacity: 0.5;
  }

  /* Drop indicator styling */
  .drop-indicator {
    display: inline;
    color: hsl(var(--primary));
    font-weight: bold;
    font-size: 1.2em;
    margin: 0 2px;
    animation: pulse 1s infinite;
    vertical-align: baseline;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }


  /* Improved visual feedback */
  .highlight-span:focus {
    outline: 2px solid hsl(var(--ring));
    outline-offset: 1px;
  }
</style>
