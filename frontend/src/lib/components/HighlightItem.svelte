<script>
  import HighlightMenu from "./HighlightMenu.svelte";

  let {
    highlight,
    index,
    isSelected = false,
    isDragging = false,
    isBeingDragged = false,
    showDropIndicatorBefore = false,
    onSelect = () => {},
    onDragStart = () => {},
    onDragEnd = () => {},
    onDragOver = () => {},
    onDrop = () => {},
    onEdit = () => {},
    onDelete = () => {},
    popoverOpen = false,
    onPopoverOpenChange = () => {},
  } = $props();
</script>

<!-- Drop indicator before this highlight -->
{#if showDropIndicatorBefore}
  <span class="drop-indicator">|</span>
{/if}

<!-- Highlight as inline text span with embedded eye icon -->
<span
  class="highlight-span
         {isSelected ? 'highlight-selected' : ''}
         {isBeingDragged ? 'highlight-dragging' : ''}"
  style="background-color: {highlight.color}40;"
  draggable="true"
  ondragstart={(e) => onDragStart(e, highlight, index)}
  ondragend={onDragEnd}
  onclick={(e) => onSelect(e, highlight)}
  ondragover={(e) => onDragOver(e, index)}
  ondrop={(e) => onDrop(e, index)}
  role="button"
  tabindex="0"
  >{highlight.text ||
    highlight.videoClipName}<!--
--><!-- Eye icon inside highlight --><!--
--><span
    class="inline-flex items-center ml-1"
  >
    <HighlightMenu
      {highlight}
      {onEdit}
      {onDelete}
      {popoverOpen}
      {onPopoverOpenChange}
      iconSize="w-3 h-3"
      triggerSize="w-5 h-5"
    />
  </span></span
>

<style>
  /* Natural text flow highlight spans */
  .highlight-span {
    display: inline;
    padding: 2px 4px;
    border-radius: 3px;
    cursor: move;
    user-select: none;
    transition: all 0.2s ease;
    font-weight: 500;
    position: relative;
    color: hsl(var(--foreground));
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
