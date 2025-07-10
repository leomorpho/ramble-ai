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
  style="background-color: {highlight.color};"
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
    transition: all 0.2s ease;
    font-weight: 500;
    position: relative;
    color: hsl(var(--foreground));
  }

  /* Non-draggable state */
  .highlight-non-draggable {
    cursor: default;
  }

  .highlight-span:hover {
    filter: brightness(1.1);
    transform: translateY(-0.5px);
  }

  .highlight-span:active {
    transform: translateY(0);
  }

  /* Selection state for highlights */
  .highlight-selected {
    box-shadow: 0 0 0 2px currentColor;
    transform: translateY(-1px);
  }

  /* Dragging state */
  .highlight-dragging {
    opacity: 0.5;
    transform: scale(0.95);
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

  /* Smooth transitions */
  .highlight-span {
    transition: all 0.15s ease;
  }

  /* Improved visual feedback */
  .highlight-span:focus {
    outline: 2px solid hsl(var(--ring));
    outline-offset: 1px;
  }
</style>
