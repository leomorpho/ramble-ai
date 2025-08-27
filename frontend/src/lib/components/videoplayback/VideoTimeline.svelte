<script>
  import { formatTime } from "./timelineUtils.js";
  import TimelineSegment from "./TimelineSegment.svelte";

  let {
    videoHighlights = [],
    currentHighlightIndex = 0,
    currentTime = 0,
    totalDuration = 0,
    enableEyeButton = true,
    shouldEnableReordering = false,
    shouldShowSegmentNumbers = true,
    isDragging = false,
    dragStartIndex = -1,
    dragOverIndex = -1,
    isPopoverOpen,
    openPopover,
    closePopover,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDrop,
    onSegmentClick,
    onEditHighlight,
    onDeleteConfirm,
    onTimelineSeek,
    onPlayAfterSeek = null,
    DISABLE_REORDERING_THRESHOLD = 30
  } = $props();

  // Constants
  const ACTIVE_SEGMENT_THRESHOLD = 0.2; // Show if any segment is less than 20% of total duration
  const HIDE_SEGMENT_NUMBERS_THRESHOLD = 20; // Hide segment numbers when more than 20 segments

  // Calculate total duration from highlights
  let calculatedTotalDuration = $derived(() => {
    const duration = videoHighlights.reduce((sum, h) => sum + (h.end - h.start), 0);
    return duration;
  });

  // Calculate if we should show the active segment based on highlight durations
  let shouldShowActiveSegment = $derived(() => {
    // Don't show for 0 or 1 highlights
    if (videoHighlights.length <= 1) return false;

    const totalDurationCalc = calculatedTotalDuration();
    if (totalDurationCalc === 0) return false;

    // Check if any highlight is less than the threshold percentage of total duration
    return videoHighlights.some((h) => {
      const segmentDuration = h.end - h.start;
      const percentage = segmentDuration / totalDurationCalc;
      return percentage < ACTIVE_SEGMENT_THRESHOLD;
    });
  });

  // Calculate if we should show segment numbers
  let shouldShowSegmentNumbersComputed = $derived(() => {
    return videoHighlights.length <= HIDE_SEGMENT_NUMBERS_THRESHOLD;
  });

  // Calculate if reordering should be enabled
  let enableReordering = $derived(shouldEnableReordering);

  // Enhanced segment click handler that seeks and then plays
  async function handleSegmentClickAndPlay(event, index) {
    // First, handle the normal segment click (seeking)
    await onSegmentClick(event, index);
    
    // Then immediately start playing if callback is provided
    if (onPlayAfterSeek) {
      onPlayAfterSeek();
    }
  }

  // Enhanced timeline seek handler that seeks and then plays
  async function handleTimelineSeekAndPlay(targetTime) {
    // First, seek to the target time
    await onTimelineSeek(targetTime);
    
    // Then immediately start playing if callback is provided
    if (onPlayAfterSeek) {
      onPlayAfterSeek();
    }
  }
</script>

<!-- Draggable Clip Timeline -->
<div class="timeline-container mb-4 max-w-full overflow-hidden">
  <div class="space-y-2 max-w-full">
    {#if enableReordering}
      <div class="text-xs text-muted-foreground mb-2">
        Click segments to seek, drag handle (•) to reorder
      </div>
    {:else}
      <div class="text-xs text-muted-foreground mb-2">
        Click segments to seek{videoHighlights.length > DISABLE_REORDERING_THRESHOLD
          ? ` (reordering disabled for ${videoHighlights.length} segments)`
          : ""}
      </div>
    {/if}

    <!-- Clip segments with drag and drop -->
    <!-- Timeline always maintains proportional width to match video player -->
    <div class="flex pt-2 min-h-[2rem] w-full">
      {#each videoHighlights as highlight, index}
        {@const segmentDuration = highlight.end - highlight.start}
        {@const totalDurationCalc = calculatedTotalDuration()}
        {@const proportionalWidth = totalDurationCalc > 0
            ? (segmentDuration / totalDurationCalc) * 100
            : 100 / videoHighlights.length}
        {@const segmentWidth = proportionalWidth}
        {@const isActive = index === currentHighlightIndex}

        <!-- Drop indicator before this segment -->
        {#if enableReordering && isDragging && dragOverIndex === index}
          {@render dropIndicator()}
        {/if}

        <TimelineSegment
          {highlight}
          {index}
          {isActive}
          {segmentWidth}
          {currentTime}
          {totalDuration}
          highlights={videoHighlights}
          enableReordering={enableReordering}
          enableEyeButton={enableEyeButton && !shouldShowActiveSegment()}
          showSegmentNumber={shouldShowSegmentNumbersComputed()}
          {isDragging}
          {dragStartIndex}
          {isPopoverOpen}
          {openPopover}
          {closePopover}
          isFirst={index === 0}
          isLast={index === videoHighlights.length - 1}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
          onDragOver={onDragOver}
          onDrop={onDrop}
          onSegmentClick={handleSegmentClickAndPlay}
          onEditHighlight={onEditHighlight}
          onDeleteConfirm={onDeleteConfirm}
        />

        <!-- Drop indicator after the last segment -->
        {#if enableReordering && index === videoHighlights.length - 1 && isDragging && dragOverIndex === videoHighlights.length}
          {@render dropIndicator()}
        {/if}
      {/each}
    </div>

    <!-- Active segment in full width -->
    {#if shouldShowActiveSegment() && videoHighlights[currentHighlightIndex]}
      {@const activeHighlight = videoHighlights[currentHighlightIndex]}
      {@const segmentStartTime = videoHighlights
        .slice(0, currentHighlightIndex)
        .reduce((sum, h) => sum + (h.end - h.start), 0)}
      {@const segmentDuration = activeHighlight.end - activeHighlight.start}

      <div class="mt-1">
        <div class="w-full">
          <TimelineSegment
            highlight={activeHighlight}
            index={currentHighlightIndex}
            isActive={true}
            isFirst={true}
            isLast={true}
            segmentWidth={100}
            {currentTime}
            highlights={videoHighlights}
            enableReordering={false}
            enableEyeButton={true}
            showSegmentNumber={true}
            isDragging={false}
            dragStartIndex={null}
            {isPopoverOpen}
            {openPopover}
            {closePopover}
            onDragStart={() => {}}
            onDragEnd={() => {}}
            onDragOver={() => {}}
            onDrop={() => {}}
            onSegmentClick={(e) => {
              // Calculate click position to seek within current segment
              const rect = e.currentTarget.getBoundingClientRect();
              const x = e.clientX - rect.left;
              const clickPercentage = x / rect.width;
              const clickTargetTime =
                segmentStartTime + clickPercentage * segmentDuration;
              handleTimelineSeekAndPlay(clickTargetTime);
            }}
            onEditHighlight={onEditHighlight}
            onDeleteConfirm={onDeleteConfirm}
          />
        </div>
      </div>
    {/if}

    <!-- Time display -->
    <div class="flex justify-between text-xs text-muted-foreground">
      <span>{formatTime(currentTime)}</span>
      <span>Clip {currentHighlightIndex + 1} of {videoHighlights.length}</span>
      <span>{formatTime(totalDuration)}</span>
    </div>
  </div>
</div>

{#snippet dropIndicator()}
  <div class="w-0.5 h-8 bg-black dark:bg-white rounded flex-shrink-0"></div>
{/snippet}