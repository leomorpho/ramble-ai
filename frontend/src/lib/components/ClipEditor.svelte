<script>
  import { onMount } from 'svelte';
  import { GetVideoURL, UpdateVideoClipHighlights } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import { Edit3, Save, X, RotateCcw, Play, Pause, SkipBack, SkipForward } from '@lucide/svelte';
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
    onSave = () => {} 
  } = $props();

  // Local state
  let videoURL = $state('');
  let videoElement = $state(null);
  let videoLoading = $state(false);
  let saving = $state(false);
  
  // Editable values
  let editedStart = $state(0);
  let editedEnd = $state(0);
  let originalStart = $state(0);
  let originalEnd = $state(0);
  
  // Video player state
  let currentTime = $state(0);
  let duration = $state(0);
  let isPlaying = $state(false);

  // Watch for highlight changes
  $effect(() => {
    if (highlight && open) {
      loadVideo();
      resetValues();
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
  }

  // Set end time to current playback time
  function setEndToCurrent() {
    editedEnd = Math.max(currentTime, editedStart + 0.1); // Ensure end is after start
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

  // Save changes
  async function saveChanges() {
    if (!highlight) return;

    // Validate times
    if (editedStart >= editedEnd) {
      toast.error('Invalid time range', {
        description: 'Start time must be before end time'
      });
      return;
    }

    if (editedStart < 0 || editedEnd > duration) {
      toast.error('Invalid time range', {
        description: 'Times must be within video duration'
      });
      return;
    }

    saving = true;
    try {
      // Create the updated highlight object
      const updatedHighlight = {
        id: highlight.id,
        start: editedStart,
        end: editedEnd,
        color: highlight.color
      };

      // Get the video clip ID from the highlight
      const videoClipId = highlight.videoClipId;

      // For now, we'll update just this highlight
      // In a production system, you might want to fetch all highlights for this clip
      // and update the specific one, but this simpler approach should work
      await UpdateVideoClipHighlights(videoClipId, [updatedHighlight]);

      toast.success('Highlight updated', {
        description: 'Start and end times have been saved'
      });

      // Call the onSave callback with updated highlight
      onSave({
        ...highlight,
        start: editedStart,
        end: editedEnd
      });

      // Close the dialog
      open = false;
    } catch (err) {
      console.error('Failed to save highlight:', err);
      toast.error('Failed to save changes', {
        description: 'Could not update the highlight times'
      });
    } finally {
      saving = false;
    }
  }

  // Check if values have changed
  function hasChanges() {
    return Math.abs(editedStart - originalStart) > 0.001 || 
           Math.abs(editedEnd - originalEnd) > 0.001;
  }

  // Handle timeline click for seeking
  function handleTimelineClick(event) {
    if (!videoElement || duration === 0) return;
    
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const clickPercentage = x / rect.width;
    const targetTime = clickPercentage * duration;
    
    seekTo(targetTime);
  }

  // Close dialog
  function closeDialog() {
    if (videoElement) {
      videoElement.pause();
    }
    open = false;
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[1000px] max-h-[90vh] overflow-y-auto">
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
          <div class="space-y-3">
            <div class="flex items-center justify-between text-sm text-muted-foreground">
              <span>Video Timeline</span>
              <span>Total: {formatTime(duration)}</span>
            </div>
            
            <!-- Timeline Bar -->
            <div 
              class="relative w-full h-12 bg-secondary rounded-lg overflow-hidden cursor-pointer hover:bg-secondary/80 transition-colors"
              onclick={handleTimelineClick}
              title="Click to seek to position"
            >
              <!-- Full video background -->
              <div class="absolute inset-0 bg-secondary"></div>
              
              <!-- Highlight segment -->
              {#if duration > 0}
                {@const highlightStart = (editedStart / duration) * 100}
                {@const highlightWidth = ((editedEnd - editedStart) / duration) * 100}
                <div 
                  class="absolute top-0 h-full rounded transition-all duration-200"
                  style="left: {highlightStart}%; width: {highlightWidth}%; background-color: {highlight.color};"
                  title="Highlight segment: {formatTime(editedStart)} - {formatTime(editedEnd)}"
                ></div>
              {/if}
              
              <!-- Current playhead -->
              {#if duration > 0}
                {@const playheadPosition = (currentTime / duration) * 100}
                <div 
                  class="absolute top-0 w-0.5 h-full bg-white shadow-lg z-10 transition-all duration-75"
                  style="left: {playheadPosition}%;"
                ></div>
              {/if}
              
              <!-- Start/End markers -->
              {#if duration > 0}
                {@const startPosition = (editedStart / duration) * 100}
                {@const endPosition = (editedEnd / duration) * 100}
                
                <!-- Start marker -->
                <div 
                  class="absolute top-0 w-1 h-full bg-green-500 z-20 transition-all duration-200"
                  style="left: {startPosition}%;"
                  title="Start: {formatTime(editedStart)}"
                ></div>
                
                <!-- End marker -->
                <div 
                  class="absolute top-0 w-1 h-full bg-red-500 z-20 transition-all duration-200"
                  style="left: {endPosition}%;"
                  title="End: {formatTime(editedEnd)}"
                ></div>
              {/if}
              
              <!-- Time labels -->
              <div class="absolute inset-0 flex items-center justify-between px-2 text-xs text-white/80 font-mono pointer-events-none">
                <span>0:00</span>
                <span class="bg-black/50 px-1 rounded">
                  {formatTime(currentTime)}
                </span>
                <span>{formatTime(duration)}</span>
              </div>
            </div>
            
            <!-- Video Controls -->
            <div class="flex items-center justify-center gap-4">
              <Button
                variant="outline"
                size="sm"
                onclick={() => seekTo(Math.max(0, currentTime - 10))}
                title="Skip back 10 seconds"
              >
                <SkipBack class="w-4 h-4" />
                -10s
              </Button>
              
              <Button
                variant="default"
                size="sm"
                onclick={togglePlayPause}
                class="px-6"
              >
                {#if isPlaying}
                  <Pause class="w-4 h-4 mr-2" />
                  Pause
                {:else}
                  <Play class="w-4 h-4 mr-2" />
                  Play
                {/if}
              </Button>
              
              <Button
                variant="outline"
                size="sm"
                onclick={() => seekTo(Math.min(duration, currentTime + 10))}
                title="Skip forward 10 seconds"
              >
                <SkipForward class="w-4 h-4" />
                +10s
              </Button>
            </div>

            <!-- Timeline Legend -->
            <div class="flex items-center justify-center gap-6 text-xs text-muted-foreground">
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded" style="background-color: {highlight.color};"></div>
                <span>Highlight Segment</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-1 h-3 bg-green-500"></div>
                <span>Start ({formatTime(editedStart)})</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-1 h-3 bg-red-500"></div>
                <span>End ({formatTime(editedEnd)})</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-0.5 h-3 bg-white"></div>
                <span>Playhead</span>
              </div>
            </div>
          </div>
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
                oninput={(e) => editedStart = parseTimeFromInput(e.target.value)}
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
                oninput={(e) => editedEnd = parseTimeFromInput(e.target.value)}
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
                Updated: {formatTime(editedStart)} - {formatTime(editedEnd)}
              </p>
            </div>
            {#if hasChanges()}
              <div class="text-sm text-amber-600 dark:text-amber-400">
                ● Changes pending
              </div>
            {/if}
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
        class="flex items-center gap-2"
      >
        <RotateCcw class="w-4 h-4" />
        Reset
      </Button>
      
      <div class="flex gap-2">
        <Button variant="outline" onclick={closeDialog}>
          Cancel
        </Button>
        <Button
          onclick={saveChanges}
          disabled={saving || !hasChanges()}
          class="flex items-center gap-2"
        >
          {#if saving}
            <div class="w-4 h-4 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
          {:else}
            <Save class="w-4 h-4" />
          {/if}
          Save Changes
        </Button>
      </div>
    </div>
  </DialogContent>
</Dialog>