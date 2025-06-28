<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import { Play, Pause, SkipForward, SkipBack, Square, Film } from '@lucide/svelte';
  import { Button } from "$lib/components/ui/button";

  let { highlights = [] } = $props();
  
  // Player state
  let isPlaying = $state(false);
  let isPaused = $state(false); // Separate reactive state for play/pause
  let currentHighlightIndex = $state(0);
  let videoElements = $state([]);
  let currentVideoElement = $state(null);
  let displayVideoElement = $state(null); // Visible video element for UI
  let loadedVideos = $state(new Set());
  let loadingProgress = $state(0);
  let allVideosLoaded = $state(false);
  
  // Progress tracking
  let virtualTime = $state(0); // Time position in the virtual stitched video
  let totalVirtualDuration = $state(0);
  let segmentStartTimes = $state([]); // Start time of each segment in virtual timeline
  
  // Animation frame for smooth progress updates
  let animationFrame = null;

  // Calculate total duration and segment start times
  function calculateVirtualTimeline() {
    let runningTime = 0;
    segmentStartTimes = [];
    
    for (let i = 0; i < highlights.length; i++) {
      segmentStartTimes[i] = runningTime;
      const segmentDuration = highlights[i].end - highlights[i].start;
      runningTime += segmentDuration;
    }
    
    totalVirtualDuration = runningTime;
  }

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  // Preload all videos
  async function preloadVideos() {
    if (highlights.length === 0) return;
    
    loadingProgress = 0;
    loadedVideos.clear();
    videoElements = [];
    
    const uniqueVideos = new Map();
    
    // Group highlights by video file to avoid loading same video multiple times
    for (const highlight of highlights) {
      if (!uniqueVideos.has(highlight.filePath)) {
        uniqueVideos.set(highlight.filePath, highlight);
      }
    }
    
    const videoFiles = Array.from(uniqueVideos.values());
    let loadedCount = 0;
    
    for (const highlight of videoFiles) {
      try {
        const videoURL = await GetVideoURL(highlight.filePath);
        
        // Create video element
        const video = document.createElement('video');
        video.preload = 'auto';
        video.style.display = 'none';
        document.body.appendChild(video);
        
        // Wait for video to load
        await new Promise((resolve, reject) => {
          video.onloadeddata = () => {
            loadedVideos.add(highlight.filePath);
            loadedCount++;
            loadingProgress = (loadedCount / videoFiles.length) * 100;
            resolve();
          };
          
          video.onerror = () => {
            console.error(`Failed to load video: ${highlight.filePath}`);
            reject(new Error(`Failed to load ${highlight.videoClipName}`));
          };
          
          video.src = videoURL;
        });
        
        videoElements.push({
          filePath: highlight.filePath,
          element: video
        });
        
      } catch (err) {
        console.error('Error preloading video:', err);
        toast.error('Failed to load video', {
          description: `Could not load ${highlight.videoClipName}`
        });
      }
    }
    
    allVideosLoaded = loadedCount === videoFiles.length;
    
    if (allVideosLoaded) {
      toast.success('All videos loaded!', {
        description: 'Ready for seamless playback'
      });
    }
  }

  // Get video element for a specific file path
  function getVideoElement(filePath) {
    const videoData = videoElements.find(v => v.filePath === filePath);
    return videoData ? videoData.element : null;
  }

  // Start playing the sequence
  async function startPlayback() {
    if (!allVideosLoaded) {
      toast.error('Videos still loading', {
        description: 'Please wait for all videos to finish loading'
      });
      return;
    }
    
    if (highlights.length === 0) return;
    
    currentHighlightIndex = 0;
    virtualTime = 0;
    isPlaying = true;
    isPaused = false;
    
    await playCurrentHighlight();
    startProgressTracking();
  }

  // Play the current highlight
  async function playCurrentHighlight() {
    if (currentHighlightIndex >= highlights.length) {
      stopPlayback();
      return;
    }
    
    const highlight = highlights[currentHighlightIndex];
    const video = getVideoElement(highlight.filePath);
    
    if (!video) {
      console.error('Video element not found for:', highlight.filePath);
      await playNextHighlight();
      return;
    }
    
    // Set current video element
    currentVideoElement = video;
    
    // Update visible video element source and sync
    if (displayVideoElement) {
      // Pause and reset display video first to avoid conflicts
      displayVideoElement.pause();
      displayVideoElement.ontimeupdate = null;
      
      // Only update src if it's different to avoid reloading
      if (displayVideoElement.src !== video.src) {
        displayVideoElement.src = video.src;
        // Wait for the video to be ready with timeout
        await new Promise((resolve, reject) => {
          const timeout = setTimeout(() => {
            reject(new Error('Video load timeout'));
          }, 3000);
          
          const cleanup = () => {
            clearTimeout(timeout);
            displayVideoElement.onloadeddata = null;
            displayVideoElement.oncanplay = null;
            displayVideoElement.onerror = null;
          };
          
          displayVideoElement.onloadeddata = () => {
            cleanup();
            resolve();
          };
          displayVideoElement.oncanplay = () => {
            cleanup();
            resolve();
          };
          displayVideoElement.onerror = () => {
            cleanup();
            reject(new Error('Video load error'));
          };
        }).catch(err => {
          console.warn('Video loading issue:', err);
        });
      }
      
      // Set start times
      displayVideoElement.currentTime = highlight.start;
      video.currentTime = highlight.start;
      
      // Set up time update handler on both elements
      video.ontimeupdate = handleVideoTimeUpdate;
      displayVideoElement.ontimeupdate = handleDisplayVideoUpdate;
      
      // Sync play state
      try {
        await video.play();
        await displayVideoElement.play();
        isPaused = false;
      } catch (err) {
        console.error('Error playing video:', err);
        if (err.name !== 'AbortError') {
          await playNextHighlight();
        }
      }
    } else {
      // Fallback to hidden video only
      video.currentTime = highlight.start;
      video.ontimeupdate = handleVideoTimeUpdate;
      
      try {
        await video.play();
        isPaused = false;
      } catch (err) {
        console.error('Error playing video:', err);
        if (err.name !== 'AbortError') {
          await playNextHighlight();
        }
      }
    }
  }

  // Handle video time updates
  function handleVideoTimeUpdate() {
    if (!currentVideoElement || !isPlaying) return;
    
    const highlight = highlights[currentHighlightIndex];
    if (!highlight) return;
    
    const currentTime = currentVideoElement.currentTime;
    
    // Sync display video if it exists
    if (displayVideoElement && Math.abs(displayVideoElement.currentTime - currentTime) > 0.1) {
      displayVideoElement.currentTime = currentTime;
    }
    
    // Check if we've reached the end of the current highlight
    if (currentTime >= highlight.end) {
      playNextHighlight();
    }
  }

  // Handle display video updates (keep in sync with hidden video)
  function handleDisplayVideoUpdate() {
    if (!displayVideoElement || !currentVideoElement || !isPlaying) return;
    
    const displayTime = displayVideoElement.currentTime;
    
    // Sync hidden video if it exists
    if (Math.abs(currentVideoElement.currentTime - displayTime) > 0.1) {
      currentVideoElement.currentTime = displayTime;
    }
  }

  // Play next highlight
  async function playNextHighlight() {
    if (currentVideoElement) {
      currentVideoElement.pause();
      currentVideoElement.ontimeupdate = null;
    }
    
    if (displayVideoElement) {
      displayVideoElement.pause();
      displayVideoElement.ontimeupdate = null;
    }
    
    currentHighlightIndex++;
    
    if (currentHighlightIndex >= highlights.length) {
      stopPlayback();
      toast.success('Sequence completed!');
      return;
    }
    
    await playCurrentHighlight();
  }

  // Play previous highlight
  async function playPreviousHighlight() {
    if (currentVideoElement) {
      currentVideoElement.pause();
      currentVideoElement.ontimeupdate = null;
    }
    
    if (displayVideoElement) {
      displayVideoElement.pause();
      displayVideoElement.ontimeupdate = null;
    }
    
    currentHighlightIndex = Math.max(0, currentHighlightIndex - 1);
    
    // Update virtual time to start of current segment
    virtualTime = segmentStartTimes[currentHighlightIndex] || 0;
    
    await playCurrentHighlight();
  }

  // Toggle play/pause
  function togglePlayback() {
    if (!currentVideoElement) return;
    
    if (isPaused || currentVideoElement.paused) {
      currentVideoElement.play();
      if (displayVideoElement) {
        displayVideoElement.play();
      }
      isPaused = false;
      startProgressTracking();
    } else {
      currentVideoElement.pause();
      if (displayVideoElement) {
        displayVideoElement.pause();
      }
      isPaused = true;
      stopProgressTracking();
    }
  }

  // Stop playback
  function stopPlayback() {
    isPlaying = false;
    isPaused = false;
    currentHighlightIndex = 0;
    virtualTime = 0;
    
    if (currentVideoElement) {
      currentVideoElement.pause();
      currentVideoElement.ontimeupdate = null;
    }
    
    if (displayVideoElement) {
      displayVideoElement.pause();
      displayVideoElement.ontimeupdate = null;
    }
    
    stopProgressTracking();
  }

  // Start tracking progress for smooth updates
  function startProgressTracking() {
    stopProgressTracking(); // Clear any existing animation frame
    
    function updateProgress() {
      if (!isPlaying || !currentVideoElement) return;
      
      const highlight = highlights[currentHighlightIndex];
      if (!highlight) return;
      
      const segmentStartTime = segmentStartTimes[currentHighlightIndex] || 0;
      const currentVideoTime = currentVideoElement.currentTime;
      const highlightProgress = Math.max(0, currentVideoTime - highlight.start);
      
      virtualTime = segmentStartTime + highlightProgress;
      
      if (isPlaying) {
        animationFrame = requestAnimationFrame(updateProgress);
      }
    }
    
    animationFrame = requestAnimationFrame(updateProgress);
  }

  // Stop progress tracking
  function stopProgressTracking() {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame);
      animationFrame = null;
    }
  }

  // Calculate progress percentage
  function getProgressPercentage() {
    return totalVirtualDuration > 0 ? (virtualTime / totalVirtualDuration) * 100 : 0;
  }

  // Cleanup
  onDestroy(() => {
    stopProgressTracking();
    
    // Clean up video elements
    videoElements.forEach(({ element }) => {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
    });
    
    // Clean up display video element
    if (displayVideoElement) {
      displayVideoElement.pause();
      displayVideoElement.ontimeupdate = null;
      displayVideoElement.src = '';
    }
  });

  // Initialize and handle highlights changes
  let initialized = false;
  let lastHighlightsLength = 0;
  
  onMount(() => {
    initializePlayer();
  });
  
  // Watch for highlights changes more carefully
  $effect(() => {
    if (initialized && highlights.length !== lastHighlightsLength) {
      reinitializePlayer();
    }
  });
  
  function initializePlayer() {
    calculateVirtualTimeline();
    lastHighlightsLength = highlights.length;
    initialized = true;
    
    if (highlights.length > 0) {
      preloadVideos();
    }
  }
  
  function reinitializePlayer() {
    // Reset state when highlights change
    if (isPlaying) {
      stopPlayback();
    }
    
    // Clear previous videos
    videoElements.forEach(({ element }) => {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
    });
    
    // Reset state without triggering effects
    videoElements = [];
    loadedVideos.clear();
    allVideosLoaded = false;
    loadingProgress = 0;
    lastHighlightsLength = highlights.length;
    
    // Recalculate and preload
    calculateVirtualTimeline();
    
    if (highlights.length > 0) {
      preloadVideos();
    }
  }
</script>

{#if highlights.length > 0}
  <div class="sequential-player p-6 bg-card border rounded-lg">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold">Sequential Player</h3>
      <div class="text-sm text-muted-foreground">
        {highlights.length} highlights • {formatTime(totalVirtualDuration)} total
      </div>
    </div>
    
    <!-- Loading Progress -->
    {#if !allVideosLoaded}
      <div class="mb-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium">Loading videos...</span>
          <span class="text-sm text-muted-foreground">{Math.round(loadingProgress)}%</span>
        </div>
        <div class="w-full bg-secondary rounded-full h-2">
          <div 
            class="bg-primary h-full rounded-full transition-all duration-300"
            style="width: {loadingProgress}%"
          ></div>
        </div>
      </div>
    {/if}
    
    <!-- Player Container -->
    <div class="bg-background border rounded-lg overflow-hidden mb-4">
      {#if !allVideosLoaded}
        <div class="p-12 text-center text-muted-foreground">
          <div class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-pulse">
            <Film />
          </div>
          <p class="text-lg font-medium">Preparing videos...</p>
          <p class="text-sm">Loading {highlights.length} video segments for seamless playback</p>
        </div>
      {:else if !isPlaying}
        <div class="p-12 text-center text-muted-foreground">
          <Film class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
          <p class="text-lg font-medium">Ready to play sequence</p>
          <p class="text-sm">All videos loaded and ready for seamless playback</p>
        </div>
      {:else}
        <!-- Visible video player -->
        <video 
          bind:this={displayVideoElement}
          class="w-full aspect-video bg-black"
          controls={false}
          muted={false}
        >
          <track kind="captions" />
        </video>
      {/if}
    </div>
    
    <!-- Current Highlight Info -->
    {#if isPlaying && highlights[currentHighlightIndex]}
      {@const currentHighlight = highlights[currentHighlightIndex]}
      <div class="flex items-center gap-3 p-3 mb-4 rounded-lg" style="background-color: {currentHighlight.color}20; border-left: 4px solid {currentHighlight.color};">
        <div class="flex-shrink-0">
          <span class="inline-flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-medium">
            {currentHighlightIndex + 1}
          </span>
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="font-medium truncate">{currentHighlight.videoClipName}</h4>
          <p class="text-sm text-muted-foreground">
            {formatTime(currentHighlight.start)} - {formatTime(currentHighlight.end)}
            {#if currentHighlight.text}
              • "{currentHighlight.text}"
            {/if}
          </p>
        </div>
        <div class="text-sm text-muted-foreground">
          {currentHighlightIndex + 1} / {highlights.length}
        </div>
      </div>
    {/if}
    
    <!-- Custom Progress Bar -->
    <div class="mb-4">
      <div class="w-full bg-secondary rounded-full h-3 overflow-hidden">
        <!-- Progress segments -->
        <div class="relative h-full flex">
          {#each highlights as highlight, index}
            {@const segmentDuration = highlight.end - highlight.start}
            {@const segmentWidth = (segmentDuration / totalVirtualDuration) * 100}
            {@const segmentStart = (segmentStartTimes[index] / totalVirtualDuration) * 100}
            
            <!-- Segment background -->
            <div 
              class="relative border-r border-background/50 last:border-r-0"
              style="width: {segmentWidth}%; background-color: {highlight.color}40;"
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
            </div>
          {/each}
        </div>
      </div>
      
      <div class="flex justify-between text-xs text-muted-foreground mt-1">
        <span>{formatTime(virtualTime)}</span>
        <span>{Math.round(getProgressPercentage())}%</span>
        <span>{formatTime(totalVirtualDuration)}</span>
      </div>
    </div>
    
    <!-- Controls -->
    <div class="flex items-center justify-center gap-2">
      {#if !isPlaying}
        <Button 
          onclick={startPlayback} 
          disabled={!allVideosLoaded}
          class="flex items-center gap-2"
        >
          <Play class="w-4 h-4" />
          Play Sequence
        </Button>
      {:else}
        <Button 
          variant="outline" 
          onclick={playPreviousHighlight}
          disabled={currentHighlightIndex === 0}
          title="Previous highlight"
        >
          <SkipBack class="w-4 h-4" />
        </Button>
        
        <Button onclick={togglePlayback} class="flex items-center gap-2">
          {#if !isPaused}
            <Pause class="w-4 h-4" />
            Pause
          {:else}
            <Play class="w-4 h-4" />
            Play
          {/if}
        </Button>
        
        <Button 
          variant="outline" 
          onclick={playNextHighlight}
          disabled={currentHighlightIndex >= highlights.length - 1}
          title="Next highlight"
        >
          <SkipForward class="w-4 h-4" />
        </Button>
        
        <Button variant="outline" onclick={stopPlayback} title="Stop sequence">
          <Square class="w-4 h-4" />
        </Button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .sequential-player {
    /* Ensure the component has proper spacing */
  }
</style>