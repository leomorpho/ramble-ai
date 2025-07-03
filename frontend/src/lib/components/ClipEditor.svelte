<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL, UpdateVideoClipHighlights } from '$lib/wailsjs/go/app/App';
  import { toast } from 'svelte-sonner';
  import { Edit3, Save, X, RotateCcw, RotateCw } from '@lucide/svelte';
  import VideoTimelineEditor from '$lib/components/VideoTimelineEditor.svelte';
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";

  let { 
    open = $bindable(false), 
    highlight = null, 
    projectId = null,
    onSave = () => {} 
  } = $props();


  // Local state
  let videoURL = $state('');
  let videoElement = $state(null);
  let videoLoading = $state(false);
  
  // Editable values
  let editedStart = $state(0);
  let editedEnd = $state(0);
  let originalStart = $state(0);
  let originalEnd = $state(0);
  
  // Video player state
  let currentTime = $state(0);
  let duration = $state(0);
  let isPlaying = $state(false);

  // Timeline zoom state
  let zoomLevel = $state(1); // 1 = full timeline, higher = more zoomed
  let zoomCenter = $state(0); // Center point of zoom (in seconds)
  let isDraggingMarker = $state(false);
  let dragMarkerType = $state(''); // 'start' or 'end'

  // Only reset when a new highlight is loaded (track the highlight ID to prevent constant resets)
  let lastHighlightId = $state(null);
  
  $effect(() => {
    if (highlight && open && highlight.id !== lastHighlightId) {
      lastHighlightId = highlight.id;
      loadVideo();
      resetValues();
      setDefaultZoom();
    }
  });


  // Format time for display (MM:SS)
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  // Format time for input (SS.SSS)
  function formatTimeForInput(seconds) {
    return seconds.toFixed(3);
  }

  // Parse time from input
  function parseTimeFromInput(timeString) {
    const parsed = parseFloat(timeString);
    return isNaN(parsed) ? 0 : Math.max(0, parsed);
  }

  // Reset values to original
  function resetValues() {
    if (highlight) {
      originalStart = highlight.start;
      originalEnd = highlight.end;
      editedStart = highlight.start;
      editedEnd = highlight.end;
    }
  }

  // Load video
  async function loadVideo() {
    if (!highlight) return;
    
    videoLoading = true;
    try {
      const url = await GetVideoURL(highlight.filePath);
      videoURL = url;
    } catch (err) {
      console.error('Failed to get video URL:', err);
      toast.error('Failed to load video', {
        description: 'Could not load the video file for editing'
      });
      videoURL = '';
    } finally {
      videoLoading = false;
    }
  }


  // Handle video loaded
  function handleVideoLoaded() {
    if (videoElement) {
      duration = videoElement.duration;
      // Seek to the start of the highlight
      videoElement.currentTime = editedStart;
    }
  }

  // Handle when video can play (has enough data)
  function handleVideoCanPlay() {
    if (videoElement && editedStart !== undefined) {
      // Set to the highlight start time to show the first frame
      videoElement.currentTime = editedStart;
    }
  }

  // Handle time update
  function handleTimeUpdate() {
    if (videoElement) {
      currentTime = videoElement.currentTime;
    }
  }

  // Seek to specific time
  function seekTo(time) {
    if (videoElement) {
      videoElement.currentTime = Math.max(0, Math.min(time, duration));
    }
  }

  // Set start time to current playback time
  function setStartToCurrent() {
    editedStart = Math.min(currentTime, editedEnd - 0.1); // Ensure start is before end
    scheduleAutoSave();
  }

  // Set end time to current playback time
  function setEndToCurrent() {
    editedEnd = Math.max(currentTime, editedStart + 0.1); // Ensure end is after start
    scheduleAutoSave();
  }

  // Play/pause video
  function togglePlayPause() {
    if (videoElement) {
      if (videoElement.paused) {
        videoElement.play();
        isPlaying = true;
      } else {
        videoElement.pause();
        isPlaying = false;
      }
    }
  }


  // Check if values have changed
  function hasChanges() {
    return Math.abs(editedStart - originalStart) > 0.001 || 
           Math.abs(editedEnd - originalEnd) > 0.001;
  }

  // Calculate visible timeline range based on zoom
  function getVisibleRange() {
    if (zoomLevel === 1) {
      return { start: 0, end: duration };
    }
    
    const visibleDuration = duration / zoomLevel;
    const halfVisible = visibleDuration / 2;
    
    let rangeStart = zoomCenter - halfVisible;
    let rangeEnd = zoomCenter + halfVisible;
    
    // Clamp to video bounds
    if (rangeStart < 0) {
      rangeStart = 0;
      rangeEnd = visibleDuration;
    }
    if (rangeEnd > duration) {
      rangeEnd = duration;
      rangeStart = duration - visibleDuration;
    }
    
    return { start: Math.max(0, rangeStart), end: Math.min(duration, rangeEnd) };
  }

  // Convert timeline position to time based on zoom
  function timelinePositionToTime(percentage) {
    const { start, end } = getVisibleRange();
    return start + (percentage * (end - start));
  }

  // Convert time to timeline position based on zoom
  function timeToTimelinePosition(time) {
    const { start, end } = getVisibleRange();
    if (end === start) return 0;
    return (time - start) / (end - start);
  }

  // Handle timeline click for seeking
  function handleTimelineClick(event) {
    if (!videoElement || duration === 0 || isDraggingMarker) return;
    
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const clickPercentage = x / rect.width;
    const targetTime = timelinePositionToTime(clickPercentage);
    
    seekTo(targetTime);
  }

  // Zoom functions
  function zoomIn() {
    zoomLevel = Math.min(zoomLevel * 2, 20); // Max 20x zoom
    // Center zoom on current highlight midpoint
    zoomCenter = (editedStart + editedEnd) / 2;
  }

  function zoomOut() {
    zoomLevel = Math.max(zoomLevel / 2, 1); // Min 1x (full timeline)
    if (zoomLevel === 1) {
      zoomCenter = duration / 2;
    }
  }

  function resetZoom() {
    zoomLevel = 1;
    zoomCenter = duration / 2;
  }

  // Handle marker dragging with simpler approach
  let timelineRef = $state(null);
  
  function handleMarkerMouseDown(event, markerType) {
    event.stopPropagation();
    isDraggingMarker = true;
    dragMarkerType = markerType;
  }
  
  function handleTimelineMouseMove(event) {
    if (!isDraggingMarker || !timelineRef) return;
    
    const rect = timelineRef.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const percentage = Math.max(0, Math.min(1, x / rect.width));
    const newTime = timelinePositionToTime(percentage);
    
    if (dragMarkerType === 'start') {
      editedStart = Math.max(0, Math.min(newTime, editedEnd - 0.1));
    } else if (dragMarkerType === 'end') {
      editedEnd = Math.min(duration, Math.max(newTime, editedStart + 0.1));
    }
  }
  
  function handleTimelineMouseUp() {
    if (isDraggingMarker) {
      // Trigger auto-save when user finishes dragging
      scheduleAutoSave();
    }
    isDraggingMarker = false;
    dragMarkerType = '';
  }

  // Input handlers with auto-save
  function handleStartTimeChange(e) {
    const newValue = parseTimeFromInput(e.target.value);
    editedStart = Math.max(0, Math.min(newValue, duration));
    scheduleAutoSave();
  }

  function handleEndTimeChange(e) {
    const newValue = parseTimeFromInput(e.target.value);
    editedEnd = Math.max(0, Math.min(newValue, duration));
    scheduleAutoSave();
  }

  // Set default zoom to focus on highlight segment
  function setDefaultZoom() {
    if (!highlight || !duration) return;
    
    const highlightDuration = editedEnd - editedStart;
    
    // Calculate zoom level to make highlight take up about 80% of timeline width
    // This leaves some padding on both sides for context
    const targetZoom = Math.min(duration / (highlightDuration / 0.8), 20);
    
    zoomLevel = Math.max(1, targetZoom);
    zoomCenter = (editedStart + editedEnd) / 2;
  }

  // Auto-save state
  let autoSaveTimeout = $state(null);
  let isAutoSaving = $state(false);

  // Close dialog
  function closeDialog() {
    if (videoElement) {
      videoElement.pause();
    }
    open = false;
  }

  // Auto-save changes with debouncing
  function scheduleAutoSave() {
    // Clear existing timeout
    if (autoSaveTimeout) {
      clearTimeout(autoSaveTimeout);
    }
    
    // Schedule auto-save after 1 second of no changes
    autoSaveTimeout = setTimeout(() => {
      if (hasChanges()) {
        autoSaveChanges();
      }
    }, 1000);
  }

  // Auto-save the changes
  async function autoSaveChanges() {
    if (!highlight || isAutoSaving) return;

    // Validate times before saving
    if (editedStart >= editedEnd) {
      return; // Don't auto-save invalid ranges
    }

    if (editedStart < 0 || editedEnd > duration) {
      return; // Don't auto-save out-of-bounds times
    }

    isAutoSaving = true;
    try {
      const updatedHighlight = {
        id: highlight.id,
        start: editedStart,
        end: editedEnd,
        color: highlight.color
      };

      const videoClipId = highlight.videoClipId;
      await UpdateVideoClipHighlights(videoClipId, [updatedHighlight]);

      // Update original values to reflect saved state
      originalStart = editedStart;
      originalEnd = editedEnd;

      // Call the onSave callback with updated highlight
      onSave({
        ...highlight,
        start: editedStart,
        end: editedEnd
      });

    } catch (err) {
      console.error('Auto-save failed:', err);
      // Don't show error toast for auto-save failures to avoid being intrusive
    } finally {
      isAutoSaving = false;
    }
  }

  // Cleanup timeout on component destroy
  onDestroy(() => {
    if (autoSaveTimeout) {
      clearTimeout(autoSaveTimeout);
    }
  });
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[1200px] max-h-[95vh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle class="flex items-center gap-2">
        <Edit3 class="w-5 h-5" />
        Edit Highlight Times
      </DialogTitle>
      <DialogDescription>
        {#if highlight}
          Adjust the start and end times for "{highlight.videoClipName}"
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    {#if highlight}
      <div class="space-y-6">
        <!-- Video Player -->
        <div class="bg-background border rounded-lg overflow-hidden">
          {#if videoLoading}
            <div class="p-8 text-center text-muted-foreground">
              <div class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-spin">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </div>
              <p class="text-lg font-medium">Loading video...</p>
            </div>
          {:else if videoURL}
            <div class="relative">
              <video 
                bind:this={videoElement}
                class="w-full h-auto max-h-96" 
                preload="auto"
                src={videoURL}
                onloadeddata={handleVideoLoaded}
                oncanplay={handleVideoCanPlay}
                onloadedmetadata={() => {
                  if (videoElement) {
                    duration = videoElement.duration;
                  }
                }}
                ontimeupdate={handleTimeUpdate}
                onplay={() => isPlaying = true}
                onpause={() => isPlaying = false}
                onclick={togglePlayPause}
              >
                <track kind="captions" src="" label="No captions available" />
              </video>
              
              <!-- Custom Play/Pause Overlay -->
              <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                <button
                  class="pointer-events-auto bg-black/60 hover:bg-black/80 text-white rounded-full p-4 transition-all duration-200 opacity-0 hover:opacity-100 group-hover:opacity-100"
                  onclick={togglePlayPause}
                  title={isPlaying ? 'Pause' : 'Play'}
                >
                  {#if isPlaying}
                    <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z"/>
                    </svg>
                  {:else}
                    <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z"/>
                    </svg>
                  {/if}
                </button>
              </div>
              
              <!-- Video click area for play/pause -->
              <div 
                class="absolute inset-0 cursor-pointer group"
                onclick={togglePlayPause}
                title={isPlaying ? 'Pause' : 'Play'}
              ></div>
            </div>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <p class="text-lg font-medium">Video not available</p>
            </div>
          {/if}
        </div>

        <!-- Timeline Visualization -->
        {#if videoURL && !videoLoading && duration > 0}
          <VideoTimelineEditor 
            {highlight}
            {projectId}
            {currentTime}
            {duration}
            {isPlaying}
            bind:editedStart
            bind:editedEnd
            onSeek={seekTo}
            onTogglePlay={togglePlayPause}
          />
        {/if}


        <!-- Current Time Info -->
        {#if videoURL && !videoLoading}
          <div class="grid grid-cols-3 gap-4 p-4 bg-secondary/30 rounded-lg text-sm">
            <div class="text-center">
              <div class="font-medium">Current Time</div>
              <div class="text-lg font-mono">{formatTime(currentTime)}</div>
              <div class="text-xs text-muted-foreground">{formatTimeForInput(currentTime)}s</div>
            </div>
            <div class="text-center">
              <div class="font-medium">Duration</div>
              <div class="text-lg font-mono">{formatTime(duration)}</div>
              <div class="text-xs text-muted-foreground">{formatTimeForInput(duration)}s</div>
            </div>
            <div class="text-center">
              <div class="font-medium">Highlight Range</div>
              <div class="text-lg font-mono">{formatTime(editedEnd - editedStart)}</div>
              <div class="text-xs text-muted-foreground">{formatTimeForInput(editedEnd - editedStart)}s</div>
            </div>
          </div>
        {/if}

        <!-- Time Controls -->
        <div class="grid grid-cols-2 gap-6">
          <!-- Start Time -->
          <div class="space-y-3">
            <Label for="start-time" class="text-base font-medium">Start Time</Label>
            <div class="space-y-2">
              <Input
                id="start-time"
                type="number"
                step="0.001"
                min="0"
                max={duration}
                value={formatTimeForInput(editedStart)}
                onchange={handleStartTimeChange}
                class="font-mono"
              />
              <div class="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onclick={() => seekTo(editedStart)}
                  class="flex-1"
                >
                  Seek to Start
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onclick={setStartToCurrent}
                  class="flex-1"
                >
                  Set to Current
                </Button>
              </div>
              <div class="text-xs text-muted-foreground">
                Display: {formatTime(editedStart)}
              </div>
            </div>
          </div>

          <!-- End Time -->
          <div class="space-y-3">
            <Label for="end-time" class="text-base font-medium">End Time</Label>
            <div class="space-y-2">
              <Input
                id="end-time"
                type="number"
                step="0.001"
                min="0"
                max={duration}
                value={formatTimeForInput(editedEnd)}
                onchange={handleEndTimeChange}
                class="font-mono"
              />
              <div class="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onclick={() => seekTo(editedEnd)}
                  class="flex-1"
                >
                  Seek to End
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onclick={setEndToCurrent}
                  class="flex-1"
                >
                  Set to Current
                </Button>
              </div>
              <div class="text-xs text-muted-foreground">
                Display: {formatTime(editedEnd)}
              </div>
            </div>
          </div>
        </div>

        <!-- Highlight Preview -->
        <div class="p-4 rounded-lg border" style="background-color: {highlight.color}20; border-left: 4px solid {highlight.color};">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="font-medium">{highlight.videoClipName}</h3>
              <p class="text-sm text-muted-foreground">
                Original: {formatTime(originalStart)} - {formatTime(originalEnd)}
              </p>
              <p class="text-sm font-medium">
                Current: {formatTime(editedStart)} - {formatTime(editedEnd)}
              </p>
            </div>
            <div class="text-sm">
              {#if isAutoSaving}
                <div class="text-blue-600 dark:text-blue-400 flex items-center gap-1">
                  <div class="w-3 h-3 animate-spin rounded-full border border-current border-t-transparent"></div>
                  Saving...
                </div>
              {:else if hasChanges()}
                <div class="text-amber-600 dark:text-amber-400">
                  ● Auto-save pending
                </div>
              {:else}
                <div class="text-green-600 dark:text-green-400">
                  ✓ Saved
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {/if}
    
    <!-- Actions -->
    <div class="flex justify-between gap-2 mt-6">
      <Button
        variant="outline"
        onclick={resetValues}
        disabled={!hasChanges()}
        class="flex items-center gap-2 invisible"
      >
        <RotateCcw class="w-4 h-4" />
        Reset to Original
      </Button>
      
      <Button variant="outline" onclick={closeDialog}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>