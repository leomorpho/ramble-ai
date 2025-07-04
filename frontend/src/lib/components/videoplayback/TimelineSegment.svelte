<script>
  import HighlightMenu from '$lib/components/HighlightMenu.svelte';

  let {
    highlight,
    index,
    isActive,
    segmentWidth,
    currentTime,
    totalDuration,
    highlights,
    enableReordering = false,
    enableEyeButton = true,
    isDragging = false,
    dragStartIndex = null,
    isFirst = false,
    isLast = false,
    isPopoverOpen,
    openPopover,
    closePopover,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDrop,
    onSegmentClick,
    onEditHighlight,
    onDeleteConfirm
  } = $props();

  let segmentDuration = $derived(highlight.end - highlight.start);

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }
  
  // Get conditional rounding classes
  let roundingClasses = $derived(
    isFirst && isLast ? 'rounded' : // Single segment gets full rounding
    isFirst ? 'rounded-l' : // First segment gets left rounding
    isLast ? 'rounded-r' : // Last segment gets right rounding
    '' // Middle segments get no rounding
  );
</script>

<button
  class="group relative h-8 {roundingClasses} transition-all duration-200 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-primary/50 {isActive
    ? 'border-2 border-primary'
    : ''} {isDragging && dragStartIndex === index
    ? 'opacity-50 scale-95'
    : ''} cursor-pointer"
  style="width: {segmentWidth}%; background-color: {highlight.color}; min-width: 20px;"
  title="{highlight.videoClipName}: {formatTime(highlight.start)} - {formatTime(
    highlight.end
  )}{enableReordering
    ? ' (click to seek, drag handle to reorder)'
    : ' (click to seek)'}"
  draggable={enableReordering}
  ondragstart={(e) =>
    enableReordering ? onDragStart(e, index) : e.preventDefault()}
  ondragend={enableReordering ? onDragEnd : undefined}
  ondragover={(e) => (enableReordering ? onDragOver(e, index) : undefined)}
  ondrop={(e) => (enableReordering ? onDrop(e, index) : undefined)}
  onclick={(e) => onSegmentClick(e, index)}
>
  <!-- Progress indicator for active segment -->
  {#if isActive}
    {@const segmentStartTime = highlights
      .slice(0, index)
      .reduce((sum, h) => sum + (h.end - h.start), 0)}
    {@const segmentProgress = Math.max(
      0,
      Math.min(1, (currentTime - segmentStartTime) / segmentDuration)
    )}
    <div
      class="absolute left-0 top-0 h-full bg-white/30 {roundingClasses} transition-all duration-100"
      style="width: {segmentProgress * 100}%;"
    ></div>
  {/if}

  <!-- Segment label and eye icon -->
  <div
    class="absolute inset-0 flex items-center justify-center text-xs font-medium text-white drop-shadow pointer-events-none"
  >
    <!-- Number label -->
    <span>{index + 1}</span>

    <!-- Eye icon (only show on hover and if enabled) -->
    {#if enableEyeButton}
      <span
        class="ml-1 opacity-0 group-hover:opacity-100 hidden group-hover:block transition-opacity pointer-events-auto"
      >
        <HighlightMenu
          {highlight}
          onEdit={onEditHighlight}
          onDelete={onDeleteConfirm}
          popoverOpen={isPopoverOpen(highlight.id)}
          onPopoverOpenChange={(open) => {
            if (open) {
              openPopover(highlight.id);
            } else {
              closePopover(highlight.id);
            }
          }}
          iconSize="w-4 h-4"
          triggerSize="w-6 h-6"
        />
      </span>
    {/if}
  </div>

  <!-- Drag handle -->
  {#if enableReordering}
    <div
      class="absolute top-0 right-0 w-4 h-4 bg-black/80 {isLast ? 'rounded-bl rounded-tr' : 'rounded-bl'} opacity-0 group-hover:opacity-100 transition-opacity cursor-move flex items-center justify-center"
      title="Drag to reorder"
    >
      <div class="w-1.5 h-1.5 bg-white rounded-full"></div>
    </div>
  {/if}
</button>