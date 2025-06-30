<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { draggable } from '@neodrag/svelte';
  import { toast } from 'svelte-sonner';
  import { Play, Film, GripVertical, X } from '@lucide/svelte';
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import { 
    orderedHighlights, 
    highlightsLoading, 
    loadProjectHighlights, 
    updateHighlightOrder, 
    clearHighlights 
  } from '$lib/stores/projectHighlights.js';

  let { projectId, onHighlightClick = () => {} } = $props();
  
  // Local state for drag and drop
  let draggedItem = $state(null);
  let dragOverItem = $state(null);
  let error = $state('');
  
  // Video player dialog state
  let videoDialogOpen = $state(false);
  let currentHighlight = $state(null);
  let videoURL = $state('');
  let videoElement = $state(null);
  let videoLoading = $state(false);

  // Initialize on mount and watch for project changes
  onMount(() => {
    if (projectId) {
      loadProjectHighlights(projectId);
    }
  });

  // Watch for project ID changes
  $effect(() => {
    if (projectId) {
      loadProjectHighlights(projectId);
    }
  });

  // Cleanup on unmount
  onDestroy(() => {
    clearHighlights();
  });

  // These functions are now handled by the centralized store

  // Format timestamp for display
  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  // Format time range
  function formatTimeRange(start, end) {
    return `${formatTimestamp(start)} - ${formatTimestamp(end)}`;
  }

  // Handle drag start
  function handleDragStart(event, index) {
    draggedItem = index;
    event.dataTransfer.effectAllowed = 'move';
  }

  // Handle drag over
  function handleDragOver(event, index) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
    dragOverItem = index;
  }

  // Handle drag leave
  function handleDragLeave() {
    dragOverItem = null;
  }

  // Handle drop
  async function handleDrop(event, dropIndex) {
    event.preventDefault();
    
    if (draggedItem === null || draggedItem === dropIndex) {
      dragOverItem = null;
      return;
    }

    // Reorder the array using current store value
    const currentHighlights = $orderedHighlights;
    const newOrder = [...currentHighlights];
    const [draggedHighlight] = newOrder.splice(draggedItem, 1);
    newOrder.splice(dropIndex, 0, draggedHighlight);
    
    // Update via store (this will handle database save and state management)
    await updateHighlightOrder(newOrder);
    
    draggedItem = null;
    dragOverItem = null;
  }

  // Handle drag end
  function handleDragEnd() {
    draggedItem = null;
    dragOverItem = null;
  }

  // Handle highlight click
  async function handleHighlightClick(highlight) {
    currentHighlight = highlight;
    videoLoading = true;
    videoDialogOpen = true;
    
    try {
      // Get video URL for playback
      const url = await GetVideoURL(highlight.filePath);
      videoURL = url;
    } catch (err) {
      console.error('Failed to get video URL:', err);
      toast.error('Failed to load video', {
        description: 'Could not load the video file for playback'
      });
      videoURL = '';
    } finally {
      videoLoading = false;
    }
    
    // Also call the original callback
    onHighlightClick({
      videoClipId: highlight.videoClipId,
      filePath: highlight.filePath,
      start: highlight.start,
      end: highlight.end
    });
  }

  // Handle video loaded event
  function handleVideoLoaded() {
    if (videoElement && currentHighlight) {
      // Seek to the start of the highlight
      videoElement.currentTime = currentHighlight.start;
    }
  }

  // Handle video time update to stay within highlight bounds
  function handleVideoTimeUpdate() {
    if (videoElement && currentHighlight) {
      const currentTime = videoElement.currentTime;
      
      // If we've gone past the end of the highlight, pause and reset
      if (currentTime > currentHighlight.end) {
        videoElement.pause();
        videoElement.currentTime = currentHighlight.start;
      }
    }
  }

  // Close video dialog
  function closeVideoDialog() {
    if (videoElement) {
      videoElement.pause();
    }
    videoDialogOpen = false;
    currentHighlight = null;
    videoURL = '';
  }


  // Expose refresh method
  export function refresh() {
    loadHighlights();
  }
</script>

<div class="highlights-timeline space-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-semibold">Highlight Timeline</h2>
    <div class="text-sm text-muted-foreground">
      {$orderedHighlights.length} {$orderedHighlights.length === 1 ? 'highlight' : 'highlights'}
    </div>
  </div>

  {#if $highlightsLoading}
    <div class="text-center py-8 text-muted-foreground">
      <p>Loading highlights...</p>
    </div>
  {:else if error}
    <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4">
      <p class="font-medium">Error</p>
      <p class="text-sm">{error}</p>
    </div>
  {:else if $orderedHighlights.length === 0}
    <div class="text-center py-8 text-muted-foreground">
      <p class="text-lg">No highlights yet</p>
      <p class="text-sm">Create highlights in your video transcriptions to see them here</p>
    </div>
  {:else}
    <div class="space-y-2">
      {#each $orderedHighlights as highlight, index}
        <div
          class="highlight-card group relative flex items-start gap-3 p-4 rounded-lg border transition-all duration-200 cursor-move
                 {dragOverItem === index ? 'border-primary shadow-lg' : 'border-border hover:border-primary/50 hover:shadow-md'}
                 {draggedItem === index ? 'opacity-50' : ''}"
          style="background-color: {highlight.color}20; border-left: 4px solid {highlight.color};"
          draggable="true"
          ondragstart={(e) => handleDragStart(e, index)}
          ondragover={(e) => handleDragOver(e, index)}
          ondragleave={handleDragLeave}
          ondrop={(e) => handleDrop(e, index)}
          ondragend={handleDragEnd}
          role="button"
          tabindex="0"
          onclick={() => handleHighlightClick(highlight)}
          onkeydown={(e) => e.key === 'Enter' && handleHighlightClick(highlight)}
        >
          <!-- Drag handle -->
          <div class="flex-shrink-0 opacity-50 group-hover:opacity-100 transition-opacity">
            <GripVertical class="w-5 h-5 text-muted-foreground" />
          </div>

          <!-- Video info -->
          <div class="flex-shrink-0">
            <div class="w-10 h-10 rounded bg-secondary flex items-center justify-center">
              <Film class="w-5 h-5 text-muted-foreground" />
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <h3 class="font-medium text-sm truncate" title={highlight.videoClipName}>
                  {highlight.videoClipName}
                </h3>
                <p class="text-xs text-muted-foreground mt-1">
                  {formatTimeRange(highlight.start, highlight.end)}
                </p>
              </div>
              <button
                class="flex-shrink-0 p-1 rounded hover:bg-secondary/50 transition-colors"
                onclick={(e) => {
                  e.stopPropagation();
                  handleHighlightClick(highlight);
                }}
                title="Play this highlight"
              >
                <Play class="w-4 h-4" />
              </button>
            </div>
            
            {#if highlight.text}
              <p class="text-sm mt-2 line-clamp-2" title={highlight.text}>
                {highlight.text}
              </p>
            {/if}
          </div>
        </div>

        <!-- Drop indicator -->
        {#if dragOverItem === index && draggedItem !== null && draggedItem !== index}
          <div class="h-1 bg-primary rounded-full animate-pulse"></div>
        {/if}
      {/each}
    </div>
  {/if}
  
  <!-- Etro Video Player -->
  <EtroVideoPlayer highlights={$orderedHighlights} {projectId} />
</div>

<!-- Video Player Dialog -->
<Dialog bind:open={videoDialogOpen}>
  <DialogContent class="sm:max-w-[900px] max-h-[90vh]">
    <DialogHeader>
      <DialogTitle>Highlight Playback</DialogTitle>
      <DialogDescription>
        {#if currentHighlight}
          Playing highlight from {currentHighlight.videoClipName} ({formatTimeRange(currentHighlight.start, currentHighlight.end)})
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    {#if currentHighlight}
      <div class="space-y-4">
        <!-- Highlight info -->
        <div class="flex items-center gap-3 p-3 rounded-lg" style="background-color: {currentHighlight.color}20; border-left: 4px solid {currentHighlight.color};">
          <Film class="w-6 h-6 flex-shrink-0" style="color: {currentHighlight.color}" />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">{currentHighlight.videoClipName}</h3>
            <p class="text-sm text-muted-foreground">
              {formatTimeRange(currentHighlight.start, currentHighlight.end)}
            </p>
            {#if currentHighlight.text}
              <p class="text-sm mt-1 italic">"{currentHighlight.text}"</p>
            {/if}
          </div>
        </div>

        <!-- Video player -->
        <div class="bg-background border rounded-lg overflow-hidden">
          {#if videoLoading}
            <div class="p-8 text-center text-muted-foreground">
              <div class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-spin">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </div>
              <p class="text-lg font-medium">Loading video...</p>
              <p class="text-sm">Preparing video for playback</p>
            </div>
          {:else if videoURL}
            <video 
              bind:this={videoElement}
              class="w-full h-auto max-h-96" 
              controls 
              preload="metadata"
              src={videoURL}
              onloadeddata={handleVideoLoaded}
              ontimeupdate={handleVideoTimeUpdate}
            >
              <track kind="captions" src="" label="No captions available" />
              <p class="p-4 text-center text-muted-foreground">
                Your browser doesn't support video playback or the video format is not supported.
              </p>
            </video>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.864-.833-2.634 0L4.18 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <p class="text-lg font-medium">Video not available</p>
              <p class="text-sm">The video file could not be loaded</p>
            </div>
          {/if}
        </div>

        <!-- Video controls info -->
        {#if videoURL && !videoLoading}
          <div class="p-3 bg-secondary/30 rounded-lg">
            <div class="flex items-center gap-4 text-sm">
              <div class="flex items-center gap-2">
                <Play class="w-4 h-4" />
                <span>Video will auto-loop within highlight bounds</span>
              </div>
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full" style="background-color: {currentHighlight.color}"></span>
                <span>Highlight: {formatTimeRange(currentHighlight.start, currentHighlight.end)}</span>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
    
    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={closeVideoDialog}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>

<style>
  .highlight-card {
    user-select: none;
  }
  
  .highlight-card:hover {
    transform: translateY(-1px);
  }
  
  .highlight-card:active {
    transform: translateY(0);
  }
</style>