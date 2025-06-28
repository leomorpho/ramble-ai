<script>
  let { 
    highlights = [],
    virtualTime = 0,
    totalVirtualDuration = 0,
    currentHighlightIndex = 0,
    isPlaying = false,
    segmentStartTimes = [],
    onSeek = () => {}
  } = $props();

  // Drag state
  let isDragging = $state(false);
  let dragStartX = $state(0);
  let timelineElement = $state(null);
  let dragTargetTime = $state(0);
  let dragTargetSegment = $state(0);
  let dragTargetSegmentTime = $state(0);

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  // Calculate progress percentage
  function getProgressPercentage() {
    if (isDragging) {
      return totalVirtualDuration > 0 ? (dragTargetTime / totalVirtualDuration) * 100 : 0;
    }
    return totalVirtualDuration > 0 ? (virtualTime / totalVirtualDuration) * 100 : 0;
  }

  // Handle timeline click for seeking
  function handleTimelineClick(event) {
    if (isDragging) return; // Don't handle clicks while dragging
    
    const timeline = event.currentTarget;
    const rect = timeline.getBoundingClientRect();
    const clickX = event.clientX - rect.left;
    const clickPercentage = clickX / rect.width;
    const targetTime = clickPercentage * totalVirtualDuration;
    
    seekToPosition(targetTime);
  }

  // Handle mouse down to start dragging
  function handleMouseDown(event) {
    isDragging = true;
    dragStartX = event.clientX;
    timelineElement = event.currentTarget;
    
    // Add global mouse event listeners
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    
    event.preventDefault();
  }

  // Handle mouse move during drag
  function handleMouseMove(event) {
    if (!isDragging || !timelineElement) return;
    
    const rect = timelineElement.getBoundingClientRect();
    const mouseX = event.clientX - rect.left;
    const clickPercentage = Math.max(0, Math.min(1, mouseX / rect.width));
    const targetTime = clickPercentage * totalVirtualDuration;
    
    // Update drag targets without seeking (just for visual feedback)
    updateDragTargets(targetTime);
  }

  // Handle mouse up to end dragging
  function handleMouseUp() {
    if (isDragging) {
      // Only seek when drag ends
      seekToPosition(dragTargetTime);
    }
    
    isDragging = false;
    timelineElement = null;
    
    // Remove global mouse event listeners
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  }

  // Update drag target values for visual feedback
  function updateDragTargets(targetTime) {
    dragTargetTime = targetTime;
    
    // Calculate which segment this corresponds to
    let targetSegmentIndex = 0;
    let targetSegmentTime = targetTime;
    
    for (let i = 0; i < segmentStartTimes.length; i++) {
      if (i === segmentStartTimes.length - 1 || targetTime < segmentStartTimes[i + 1]) {
        targetSegmentIndex = i;
        targetSegmentTime = targetTime - segmentStartTimes[i];
        break;
      }
    }
    
    dragTargetSegment = targetSegmentIndex;
    dragTargetSegmentTime = targetSegmentTime;
  }

  // Seek to a specific time position
  function seekToPosition(targetTime) {
    // Find which segment this time falls into
    let targetSegmentIndex = 0;
    let targetSegmentTime = targetTime;
    
    for (let i = 0; i < segmentStartTimes.length; i++) {
      if (i === segmentStartTimes.length - 1 || targetTime < segmentStartTimes[i + 1]) {
        targetSegmentIndex = i;
        targetSegmentTime = targetTime - segmentStartTimes[i];
        break;
      }
    }
    
    // Call the seek handler with segment index and local time
    onSeek(targetSegmentIndex, targetSegmentTime);
  }

  // Handle keyboard events for accessibility
  function handleTimelineKeydown(event) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleTimelineClick(event);
    }
  }
</script>

<div class="timeline-container">
  <!-- Custom Progress Bar -->
  <div class="mb-4">
    <div 
      class="w-full bg-secondary rounded-full h-3 overflow-hidden cursor-pointer hover:h-4 transition-all duration-200 select-none"
      class:cursor-grabbing={isDragging}
      onclick={handleTimelineClick}
      onmousedown={handleMouseDown}
      onkeydown={handleTimelineKeydown}
      role="slider"
      tabindex="0"
      aria-label="Video timeline"
      aria-valuemin="0"
      aria-valuemax={totalVirtualDuration}
      aria-valuenow={virtualTime}
    >
      <!-- Progress segments -->
      <div class="relative h-full flex w-full">
        {#each highlights as highlight, index}
          {@const segmentDuration = highlight.end - highlight.start}
          {@const segmentWidth = (segmentDuration / totalVirtualDuration) * 100}
          
          <!-- Segment background -->
          <div 
            class="relative border-r border-background/50 last:border-r-0 hover:brightness-110 transition-all duration-100"
            style="width: {segmentWidth}%; background-color: {highlight.color}40;"
            title="{highlight.videoClipName}: {formatTime(highlight.start)} - {formatTime(highlight.end)}"
          >
            <!-- Progress within this segment -->
            {#if index === currentHighlightIndex && isPlaying}
              {@const segmentProgress = Math.max(0, Math.min(1, (virtualTime - segmentStartTimes[index]) / segmentDuration))}
              <div 
                class="h-full transition-all duration-100"
                style="width: {segmentProgress * 100}%; background-color: {highlight.color};"
              ></div>
            {:else if index < currentHighlightIndex}
              <!-- Completed segment -->
              <div 
                class="h-full"
                style="width: 100%; background-color: {highlight.color};"
              ></div>
            {/if}
            
            <!-- Segment divider -->
            {#if index < highlights.length - 1}
              <div class="absolute right-0 top-0 w-px h-full bg-background/30"></div>
            {/if}
          </div>
        {/each}
        
        <!-- Current position indicator -->
        <div 
          class="absolute top-0 w-1 h-full bg-white rounded-full shadow-md pointer-events-none transition-all duration-100"
          style="left: {getProgressPercentage()}%; transform: translateX(-50%);"
        ></div>
      </div>
    </div>
    
    <!-- Time labels -->
    <div class="flex justify-between text-xs text-muted-foreground mt-1">
      <span>{formatTime(isDragging ? dragTargetTime : virtualTime)}</span>
      <span>{Math.round(getProgressPercentage())}%</span>
      <span>{formatTime(totalVirtualDuration)}</span>
    </div>
    
    <!-- Segment labels (for long timelines) -->
    {#if highlights.length <= 10}
      <div class="flex mt-1 text-xs text-muted-foreground/70">
        {#each highlights as highlight, index}
          {@const segmentWidth = ((highlight.end - highlight.start) / totalVirtualDuration) * 100}
          <div 
            class="truncate px-1"
            style="width: {segmentWidth}%;"
            title="{highlight.videoClipName}"
          >
            {#if segmentWidth > 8}
              {highlight.videoClipName.slice(0, 12)}{highlight.videoClipName.length > 12 ? '...' : ''}
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .timeline-container {
    user-select: none;
  }
  
  /* Ensure proper cursor for clickable areas */
  [role="slider"] {
    cursor: pointer;
  }
  
  [role="slider"]:focus {
    outline: 2px solid hsl(var(--primary));
    outline-offset: 2px;
  }
</style>