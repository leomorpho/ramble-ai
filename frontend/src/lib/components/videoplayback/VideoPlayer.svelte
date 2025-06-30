<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import { Play, Pause, SkipForward, SkipBack, Square } from '@lucide/svelte';
  import { Button } from "$lib/components/ui/button";

  let { highlights = [] } = $props();

  // Core state
  let videoElement = $state(null);
  let preloadElement = $state(null);
  let isPlaying = $state(false);
  let isPaused = $state(false);
  let currentHighlightIndex = $state(0);
  let virtualTime = $state(0);
  let totalVirtualDuration = $state(0);
  let segmentStartTimes = $state([]);
  
  // Preloading state
  let nextSegmentPreloaded = $state(false);
  let preloadingSegmentIndex = $state(-1);
  let activeElementIsMain = $state(true); // true = videoElement is active, false = preloadElement is active

  // Video URLs and loading
  let videoURLs = $state(new Map());
  let loadingProgress = $state(0);
  let allVideosLoaded = $state(false);

  // Player initialization
  let isInitialized = $state(false);
  let initializationError = $state(null);

  // Animation frame for progress updates
  let animationFrame = null;
  
  // Reset preload state (used for edge cases)
  function resetPreloadState() {
    nextSegmentPreloaded = false;
    preloadingSegmentIndex = -1;
    console.log('Preload state reset');
  }
  
  // Check if we can preload (prevent duplicate preloading)
  function canPreloadNextSegment() {
    return !nextSegmentPreloaded && 
           currentHighlightIndex + 1 < highlights.length &&
           allVideosLoaded;
  }

  // Calculate virtual timeline
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

  // Load video URLs from backend
  async function loadVideoURLs() {
    if (highlights.length === 0) {
      console.warn('No highlights provided to load video URLs');
      return;
    }
    
    console.log('Starting to load video URLs for', highlights.length, 'highlights');
    loadingProgress = 0;
    videoURLs.clear();
    
    const uniqueVideos = new Map();
    for (const highlight of highlights) {
      if (!uniqueVideos.has(highlight.filePath)) {
        uniqueVideos.set(highlight.filePath, highlight);
      }
    }
    
    const videoFiles = Array.from(uniqueVideos.values());
    console.log('Loading URLs for', videoFiles.length, 'unique video files:', videoFiles.map(h => h.filePath));
    
    let loadedCount = 0;
    
    // Load URLs one by one to better track progress
    for (const highlight of videoFiles) {
      try {
        console.log('Loading URL for:', highlight.filePath);
        
        // Add timeout to prevent hanging
        const videoURL = await Promise.race([
          GetVideoURL(highlight.filePath),
          new Promise((_, reject) => 
            setTimeout(() => reject(new Error('GetVideoURL timeout after 10 seconds')), 10000)
          )
        ]);
        
        console.log('Got URL for', highlight.filePath, ':', videoURL ? 'SUCCESS' : 'EMPTY');
        
        if (videoURL) {
          videoURLs.set(highlight.filePath, videoURL);
          loadedCount++;
          loadingProgress = (loadedCount / videoFiles.length) * 100;
          console.log(`Progress: ${loadedCount}/${videoFiles.length} (${Math.round(loadingProgress)}%)`);
        } else {
          throw new Error('Empty video URL returned');
        }
      } catch (err) {
        console.error('Error loading video URL for:', highlight.filePath, err);
        toast.error('Failed to load video', {
          description: `Could not load ${highlight.videoClipName}: ${err.message}`
        });
        // Continue with other videos instead of stopping
      }
    }
    
    console.log('Finished loading video URLs. Loaded:', loadedCount, 'out of', videoFiles.length);
    
    if (loadedCount === videoFiles.length) {
      allVideosLoaded = true;
      console.log('All video URLs loaded successfully');
      toast.success('All video URLs loaded!');
    } else if (loadedCount > 0) {
      allVideosLoaded = true; // Allow partial loading
      console.log('Partial video URLs loaded:', loadedCount, '/', videoFiles.length);
      toast.warning(`Loaded ${loadedCount} out of ${videoFiles.length} videos`);
    } else {
      console.error('No video URLs could be loaded');
      toast.error('Failed to load any video URLs');
    }
  }

  // Simple initialization - just mark as ready once URLs are loaded
  function initializeSimplePlayer() {
    if (allVideosLoaded && highlights.length > 0) {
      isInitialized = true;
      console.log('Simple video player initialized');
      return true;
    }
    return false;
  }

  // Initialize both video elements
  function initializeVideoElements() {
    if (!videoElement || !preloadElement) {
      console.error('Video elements not ready');
      return false;
    }
    
    // Set up both elements with proper attributes
    videoElement.preload = 'metadata';
    videoElement.muted = false;
    
    preloadElement.preload = 'metadata';
    preloadElement.muted = false;
    
    console.log('Video elements initialized');
    return true;
  }
  
  // Load first video into active video element  
  async function loadFirstVideo() {
    if (!allVideosLoaded || highlights.length === 0) return false;
    
    try {
      const firstHighlight = highlights[0];
      const videoURL = videoURLs.get(firstHighlight.filePath);
      
      if (!videoURL) {
        throw new Error('No video URL for first highlight');
      }
      
      console.log('Loading first video:', firstHighlight.videoClipName);
      
      const activeEl = getActiveElement();
      
      // Ensure element is ready
      if (!activeEl) {
        throw new Error('Active video element not available');
      }
      
      activeEl.src = videoURL;
      activeEl.addEventListener('timeupdate', handleTimeUpdate);
      
      // Wait for the video to load
      await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => reject(new Error('First video load timeout')), 10000);
        
        const onLoadedData = () => {
          clearTimeout(timeout);
          activeEl.removeEventListener('loadeddata', onLoadedData);
          activeEl.removeEventListener('error', onError);
          console.log('First video loaded successfully');
          // Seek to the start of the first highlight
          activeEl.currentTime = firstHighlight.start;
          resolve();
        };
        
        const onError = () => {
          clearTimeout(timeout);
          activeEl.removeEventListener('loadeddata', onLoadedData);
          activeEl.removeEventListener('error', onError);
          reject(new Error('First video load error'));
        };
        
        if (activeEl.readyState >= 2) {
          clearTimeout(timeout);
          console.log('First video already loaded');
          activeEl.currentTime = firstHighlight.start;
          resolve();
        } else {
          activeEl.addEventListener('loadeddata', onLoadedData);
          activeEl.addEventListener('error', onError);
        }
      });
      
      return true;
      
    } catch (err) {
      console.error('Failed to load first video:', err);
      initializationError = err.message;
      return false;
    }
  }

  // Get the currently active video element
  function getActiveElement() {
    return activeElementIsMain ? videoElement : preloadElement;
  }
  
  // Get the preload video element (the one not currently active)
  function getPreloadElement() {
    return activeElementIsMain ? preloadElement : videoElement;
  }
  
  // Preload next segment 3 seconds before switching
  async function preloadNextSegment() {
    const nextIndex = currentHighlightIndex + 1;
    if (!canPreloadNextSegment()) {
      console.log('Cannot preload: conditions not met');
      return;
    }
    
    console.log(`Preloading next segment ${nextIndex} (3s before switch)`);
    
    const nextHighlight = highlights[nextIndex];
    const nextVideoURL = videoURLs.get(nextHighlight.filePath);
    
    if (!nextVideoURL) {
      console.error(`No video URL for next highlight ${nextIndex}`);
      return;
    }
    
    try {
      const preloadEl = getPreloadElement();
      
      // Load next video into preload element
      if (preloadEl.src !== nextVideoURL) {
        console.log(`Preloading video: ${nextHighlight.videoClipName}`);
        preloadEl.src = nextVideoURL;
        
        // Wait for preload element to load
        await new Promise((resolve, reject) => {
          const timeout = setTimeout(() => reject(new Error('Preload timeout')), 8000);
          
          const onLoadedData = () => {
            clearTimeout(timeout);
            preloadEl.removeEventListener('loadeddata', onLoadedData);
            preloadEl.removeEventListener('error', onError);
            resolve();
          };
          
          const onError = () => {
            clearTimeout(timeout);
            preloadEl.removeEventListener('loadeddata', onLoadedData);
            preloadEl.removeEventListener('error', onError);
            reject(new Error('Preload error'));
          };
          
          if (preloadEl.readyState >= 2) {
            clearTimeout(timeout);
            resolve();
          } else {
            preloadEl.addEventListener('loadeddata', onLoadedData);
            preloadEl.addEventListener('error', onError);
          }
        });
      }
      
      // Pre-seek to the start of next highlight
      preloadEl.currentTime = nextHighlight.start;
      nextSegmentPreloaded = true;
      preloadingSegmentIndex = nextIndex;
      
      console.log(`Preloaded segment ${nextIndex}: ${nextHighlight.videoClipName} at ${nextHighlight.start}s`);
      
    } catch (err) {
      console.error(`Failed to preload segment ${nextIndex}:`, err);
    }
  }
  
  // Seamlessly switch to preloaded segment
  async function switchToPreloadedSegment() {
    if (!nextSegmentPreloaded || preloadingSegmentIndex !== currentHighlightIndex + 1) {
      console.warn('No preloaded segment available, falling back to regular switch');
      return await switchToHighlight(currentHighlightIndex + 1);
    }
    
    console.log(`Switching to preloaded segment ${preloadingSegmentIndex}`);
    
    // Store playback state before swapping
    const wasPlaying = isPlaying && !isPaused;
    const oldActive = getActiveElement();
    
    // Seamlessly swap active elements
    activeElementIsMain = !activeElementIsMain;
    currentHighlightIndex = preloadingSegmentIndex;
    
    // Reset preload state
    resetPreloadState();
    
    // Update event listeners to new active element
    const newActive = getActiveElement(); // Now the active element
    const newPreload = getPreloadElement(); // Now the preload element (old active)
    
    // Remove listeners from old active element (now preload)
    newPreload.removeEventListener('timeupdate', handleTimeUpdate);
    
    // Add listeners to new active element
    newActive.addEventListener('timeupdate', handleTimeUpdate);
    
    // Continue playback on new active element if we were playing
    if (wasPlaying) {
      console.log('Continuing playback on new active element');
      const playPromise = newActive.play();
      if (playPromise !== undefined) {
        playPromise.catch(error => {
          console.error('Error continuing playback:', error);
        });
      }
    }
    
    console.log(`Seamlessly switched to highlight ${currentHighlightIndex}`);
    return true;
  }
  
  // Switch to a specific highlight (fallback for seeking/manual switching)
  async function switchToHighlight(highlightIndex) {
    if (highlightIndex < 0 || highlightIndex >= highlights.length) return false;
    
    const highlight = highlights[highlightIndex];
    const videoURL = videoURLs.get(highlight.filePath);
    
    if (!videoURL) {
      console.error(`No video URL for highlight ${highlightIndex}`);
      return false;
    }
    
    try {
      const activeEl = getActiveElement();
      
      // If it's a different video file, load it
      if (activeEl.src !== videoURL) {
        console.log(`Loading video: ${highlight.videoClipName}`);
        activeEl.src = videoURL;
        
        // Wait for video to load
        await new Promise((resolve, reject) => {
          const timeout = setTimeout(() => reject(new Error('Video load timeout')), 10000);
          
          const onLoadedData = () => {
            clearTimeout(timeout);
            activeEl.removeEventListener('loadeddata', onLoadedData);
            activeEl.removeEventListener('error', onError);
            resolve();
          };
          
          const onError = () => {
            clearTimeout(timeout);
            activeEl.removeEventListener('loadeddata', onLoadedData);
            activeEl.removeEventListener('error', onError);
            reject(new Error('Video load error'));
          };
          
          if (activeEl.readyState >= 2) {
            clearTimeout(timeout);
            resolve();
          } else {
            activeEl.addEventListener('loadeddata', onLoadedData);
            activeEl.addEventListener('error', onError);
          }
        });
      }
      
      // Seek to the start of this highlight
      activeEl.currentTime = highlight.start;
      currentHighlightIndex = highlightIndex;
      
      // Reset preload state when manually switching
      resetPreloadState();
      
      console.log(`Switched to highlight ${highlightIndex}: ${highlight.videoClipName} at ${highlight.start}s`);
      return true;
      
    } catch (err) {
      console.error(`Failed to switch to highlight ${highlightIndex}:`, err);
      return false;
    }
  }


  // Handle time updates with predictive preloading
  function handleTimeUpdate(event) {
    const currentTime = event.target.currentTime;
    
    // Calculate virtual time based on current highlight
    const highlight = highlights[currentHighlightIndex];
    if (highlight) {
      const segmentStartTime = segmentStartTimes[currentHighlightIndex] || 0;
      const highlightProgress = Math.max(0, currentTime - highlight.start);
      virtualTime = segmentStartTime + highlightProgress;
      
      // Calculate time until current highlight ends
      const timeUntilEnd = highlight.end - currentTime;
      
      // Predictive preloading: start preloading next segment 3 seconds before switch
      if (timeUntilEnd <= 3 && timeUntilEnd > 0 && canPreloadNextSegment()) {
        console.log(`3 seconds until switch (${timeUntilEnd.toFixed(1)}s remaining), preloading next segment...`);
        preloadNextSegment();
      }
      
      // Check if we've reached the end of current highlight
      if (currentTime >= highlight.end - 0.1) {
        console.log(`Highlight ${currentHighlightIndex} ended, switching to ${nextSegmentPreloaded ? 'preloaded' : 'next'} segment`);
        if (nextSegmentPreloaded) {
          // Use preloaded segment for seamless switch
          switchToPreloadedSegment();
        } else {
          // Fallback to regular switch
          playNextHighlight();
        }
      }
    }
  }

  // Playback controls
  async function startPlayback() {
    if (!allVideosLoaded) {
      toast.error('Videos still loading');
      return;
    }
    
    if (highlights.length === 0) return;
    
    if (!isInitialized) {
      toast.error('Video player not ready');
      return;
    }
    
    try {
      // Switch to first highlight and start playing
      const success = await switchToHighlight(0);
      if (success) {
        isPlaying = true;
        isPaused = false;
        
        const activeEl = getActiveElement();
        const playPromise = activeEl.play();
        if (playPromise !== undefined) {
          playPromise.catch(error => {
            console.error('Error playing video:', error);
            toast.error('Failed to start playback');
          });
        }
        
        startProgressTracking();
      } else {
        toast.error('Failed to load first video');
      }
    } catch (err) {
      console.error('Error starting playback:', err);
      toast.error('Failed to start playback');
    }
  }

  function togglePlayback() {
    const activeEl = getActiveElement();
    if (!activeEl || !isInitialized) return;
    
    if (isPaused) {
      // Resume
      activeEl.play();
      isPaused = false;
      startProgressTracking();
    } else {
      // Pause
      activeEl.pause();
      isPaused = true;
      stopProgressTracking();
    }
  }

  function stopPlayback() {
    const activeEl = getActiveElement();
    if (activeEl) {
      activeEl.pause();
    }
    
    isPlaying = false;
    isPaused = false;
    currentHighlightIndex = 0;
    virtualTime = 0;
    
    // Reset preload state
    resetPreloadState();
    
    stopProgressTracking();
  }

  async function playPreviousHighlight() {
    if (currentHighlightIndex > 0) {
      // Reset preload state when going backwards
      resetPreloadState();
      
      const success = await switchToHighlight(currentHighlightIndex - 1);
      if (success && isPlaying && !isPaused) {
        const activeEl = getActiveElement();
        activeEl.play();
      }
    }
  }

  async function playNextHighlight() {
    if (currentHighlightIndex < highlights.length - 1) {
      let success = false;
      
      if (nextSegmentPreloaded && preloadingSegmentIndex === currentHighlightIndex + 1) {
        // Use preloaded segment for seamless switching
        success = await switchToPreloadedSegment();
      } else {
        // Fallback to regular switching
        success = await switchToHighlight(currentHighlightIndex + 1);
      }
      
      if (success && isPlaying && !isPaused) {
        const activeEl = getActiveElement();
        activeEl.play();
      }
    } else {
      // End of sequence
      stopPlayback();
      toast.success('Sequence completed!');
    }
  }

  // Progress tracking
  function startProgressTracking() {
    stopProgressTracking();
    
    function updateProgress() {
      const activeEl = getActiveElement();
      if (!isPlaying || !activeEl) return;
      
      // Virtual time is handled by time update events
      
      if (isPlaying && !isPaused) {
        animationFrame = requestAnimationFrame(updateProgress);
      }
    }
    
    animationFrame = requestAnimationFrame(updateProgress);
  }

  function stopProgressTracking() {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame);
      animationFrame = null;
    }
  }

  // Timeline seeking
  async function handleTimelineSeek(targetTime) {
    const activeEl = getActiveElement();
    if (!activeEl || !isInitialized) return;
    
    // Reset preload state when seeking
    resetPreloadState();
    
    // Find which highlight this time corresponds to
    let targetHighlightIndex = 0;
    let segmentTime = targetTime;
    
    for (let i = 0; i < segmentStartTimes.length; i++) {
      if (i === segmentStartTimes.length - 1 || targetTime < segmentStartTimes[i + 1]) {
        targetHighlightIndex = i;
        segmentTime = targetTime - segmentStartTimes[i];
        break;
      }
    }
    
    // Switch to target highlight
    const success = await switchToHighlight(targetHighlightIndex);
    if (success) {
      const highlight = highlights[targetHighlightIndex];
      const seekTime = highlight.start + segmentTime;
      const newActiveEl = getActiveElement();
      newActiveEl.currentTime = seekTime;
      
      console.log(`Seeked to time ${targetTime} (highlight ${targetHighlightIndex} at ${seekTime}s)`);
    }
  }

  // Progress percentage for timeline
  function getProgressPercentage() {
    return totalVirtualDuration > 0 ? (virtualTime / totalVirtualDuration) * 100 : 0;
  }

  // Watch for when videos are loaded to initialize
  $effect(() => {
    if (allVideosLoaded && highlights.length > 0 && !isInitialized && videoElement && preloadElement) {
      // Initialize video elements first
      const elementsReady = initializeVideoElements();
      if (elementsReady) {
        const success = initializeSimplePlayer();
        if (success) {
          loadFirstVideo();
        }
      }
    }
  });

  // Initialize component
  onMount(async () => {
    console.log('VideoPlayer mounted with highlights:', highlights);
    console.log('Highlights length:', highlights.length);
    
    // Wait for video elements to be ready
    const waitForElements = () => {
      return new Promise((resolve) => {
        const checkElements = () => {
          if (videoElement && preloadElement) {
            console.log('Video elements are ready');
            resolve();
          } else {
            setTimeout(checkElements, 50);
          }
        };
        checkElements();
      });
    };
    
    await waitForElements();
    
    // Simple reactive function to handle highlights when they arrive
    async function initializeWhenReady() {
      if (highlights.length > 0 && !allVideosLoaded) {
        console.log('First highlight:', highlights[0]);
        console.log('Highlight file paths:', highlights.map(h => h.filePath));
        
        calculateVirtualTimeline();
        await loadVideoURLs();
      }
    }
    
    // Call immediately if highlights are already available
    await initializeWhenReady();
    
    // Also watch highlights with a simple interval check (fallback)
    const checkInterval = setInterval(async () => {
      if (highlights.length > 0 && !allVideosLoaded) {
        console.log('VideoPlayer: Highlights detected, initializing...');
        clearInterval(checkInterval);
        await initializeWhenReady();
      }
    }, 100);
    
    // Clean up interval after 10 seconds to avoid infinite checking
    setTimeout(() => clearInterval(checkInterval), 10000);
  });

  // Cleanup
  onDestroy(() => {
    stopProgressTracking();
    
    // Clean up both video elements
    if (videoElement) {
      videoElement.removeEventListener('timeupdate', handleTimeUpdate);
      videoElement.pause();
    }
    if (preloadElement) {
      preloadElement.removeEventListener('timeupdate', handleTimeUpdate);
      preloadElement.pause();
    }
  });
</script>

{#if highlights.length > 0}
  <div class="video-player p-6 bg-card border rounded-lg">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold">Video Player</h3>
      <div class="text-sm text-muted-foreground">
        {highlights.length} highlights • {formatTime(totalVirtualDuration)} total
        <br>
        <span class="text-xs">
          URLs: {videoURLs.size}/{highlights.length} • 
          Ready: {allVideosLoaded ? 'Yes' : 'No'} • 
          Init: {isInitialized ? 'Yes' : 'No'}
          {#if nextSegmentPreloaded}
            • Next preloaded ✓
          {/if}
        </span>
      </div>
    </div>
    
    <!-- Video Elements (main + preload) -->
    <div class="relative w-full aspect-video bg-black overflow-hidden mb-4">
      <!-- Main video element -->
      <video
        bind:this={videoElement}
        class="w-full h-full bg-black absolute top-0 left-0"
        class:hidden={!activeElementIsMain}
        controls={false}
        muted={false}
        preload="metadata"
      >
        <track kind="captions" />
      </video>
      
      <!-- Preload video element (hidden until swapped) -->
      <video
        bind:this={preloadElement}
        class="w-full h-full bg-black absolute top-0 left-0"
        class:hidden={activeElementIsMain}
        controls={false}
        muted={false}
        preload="metadata"
      >
        <track kind="captions" />
      </video>
      
      <!-- Loading indicator -->
      {#if !allVideosLoaded}
        <div class="absolute inset-0 flex items-center justify-center bg-black text-white">
          <div class="text-center">
            <div class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"></div>
            <p>Loading video URLs... {Math.round(loadingProgress)}%</p>
          </div>
        </div>
      {:else if !isInitialized}
        <div class="absolute inset-0 flex items-center justify-center bg-black text-white">
          <div class="text-center">
            <div class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"></div>
            <p>Initializing video player...</p>
            {#if initializationError}
              <p class="text-red-400 text-sm mt-2">Error: {initializationError}</p>
            {/if}
          </div>
        </div>
      {/if}
    </div>
    
    <!-- Current Highlight Info -->
    {#if highlights[currentHighlightIndex]}
      <div class="bg-secondary/30 p-3 rounded-md mb-4">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="font-medium text-sm">{highlights[currentHighlightIndex].videoClipName}</h4>
            <p class="text-xs text-muted-foreground mt-1">
              Segment {currentHighlightIndex + 1} of {highlights.length}
            </p>
          </div>
          <div class="text-right">
            <div class="text-sm font-mono">
              {formatTime(virtualTime)} / {formatTime(totalVirtualDuration)}
            </div>
            <div class="text-xs text-muted-foreground">
              {Math.round(getProgressPercentage())}%
            </div>
          </div>
        </div>
      </div>
    {/if}
    
    <!-- Timeline -->
    <div class="timeline-container mb-4">
      <div class="w-full bg-secondary rounded-full h-3 overflow-hidden cursor-pointer hover:h-4 transition-all duration-200">
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
        <span>{formatTime(virtualTime)}</span>
        <span>{Math.round(getProgressPercentage())}%</span>
        <span>{formatTime(totalVirtualDuration)}</span>
      </div>
    </div>
    
    <!-- Controls -->
    <div class="playback-controls flex items-center justify-center gap-2">
      {#if !isPlaying}
        <Button 
          onclick={startPlayback} 
          disabled={!allVideosLoaded || !isInitialized}
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
    
    <!-- Loading Progress -->
    {#if !allVideosLoaded}
      <div class="mt-4 p-4 bg-secondary/20 rounded-md">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm">Loading videos...</span>
          <span class="text-sm">{Math.round(loadingProgress)}%</span>
        </div>
        <div class="w-full bg-secondary rounded-full h-2">
          <div 
            class="bg-primary h-2 rounded-full transition-all duration-300"
            style="width: {loadingProgress}%"
          ></div>
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="video-player p-6 bg-card border rounded-lg">
    <div class="text-center text-muted-foreground">
      <p>No video highlights available</p>
    </div>
  </div>
{/if}

<style>
  video {
    object-fit: contain;
  }
</style>