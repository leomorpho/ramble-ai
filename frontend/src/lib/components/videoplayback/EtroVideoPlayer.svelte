<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import { Play, Pause, SkipForward, SkipBack, Square } from '@lucide/svelte';
  import { Button } from "$lib/components/ui/button";
  import * as etro from 'etro';

  let { highlights = [] } = $props();

  // Core state
  let canvasElement = $state(null);
  let movie = $state(null);
  let isPlaying = $state(false);
  let isPaused = $state(false);
  let currentTime = $state(0);
  let totalDuration = $state(0);
  let currentHighlightIndex = $state(0);

  // Video URLs and loading
  let videoURLs = $state(new Map());
  let loadingProgress = $state(0);
  let allVideosLoaded = $state(false);

  // Player initialization
  let isInitialized = $state(false);
  let initializationError = $state(null);

  // Animation frame for progress updates
  let animationFrame = null;

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
    
    for (const highlight of videoFiles) {
      try {
        console.log('Loading URL for:', highlight.filePath);
        
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

  // Get video dimensions from a test video element
  async function getVideoDimensions(videoURL) {
    return new Promise((resolve, reject) => {
      const video = document.createElement('video');
      video.onloadedmetadata = () => {
        resolve({
          width: video.videoWidth,
          height: video.videoHeight
        });
      };
      video.onerror = () => reject(new Error('Failed to load video for dimension detection'));
      video.src = videoURL;
    });
  }

  // Calculate aspect ratio preserving dimensions
  function calculateScaledDimensions(videoWidth, videoHeight, canvasWidth, canvasHeight) {
    const videoAspect = videoWidth / videoHeight;
    const canvasAspect = canvasWidth / canvasHeight;
    
    let scaledWidth, scaledHeight, x, y;
    
    if (videoAspect > canvasAspect) {
      // Video is wider than canvas - fit by width
      scaledWidth = canvasWidth;
      scaledHeight = canvasWidth / videoAspect;
      x = 0;
      y = (canvasHeight - scaledHeight) / 2;
    } else {
      // Video is taller than canvas - fit by height
      scaledHeight = canvasHeight;
      scaledWidth = canvasHeight * videoAspect;
      x = (canvasWidth - scaledWidth) / 2;
      y = 0;
    }
    
    return { width: scaledWidth, height: scaledHeight, x, y };
  }

  // Create Etro movie with video layers
  async function createEtroMovie() {
    if (!canvasElement || !allVideosLoaded || highlights.length === 0) {
      console.error('Cannot create Etro movie: missing requirements');
      return false;
    }

    try {
      console.log('Creating Etro movie with', highlights.length, 'video layers');
      
      // Set canvas dimensions
      const canvasWidth = 1280;
      const canvasHeight = 720;
      canvasElement.width = canvasWidth;
      canvasElement.height = canvasHeight;
      
      // Get video dimensions from the first video
      const firstVideoURL = videoURLs.get(highlights[0].filePath);
      if (!firstVideoURL) {
        throw new Error('No video URL for first highlight');
      }
      
      console.log('Getting video dimensions from first video...');
      const videoDimensions = await getVideoDimensions(firstVideoURL);
      console.log('Video dimensions:', videoDimensions);
      
      // Create movie first (Etro determines dimensions from canvas)
      movie = new etro.Movie({ 
        canvas: canvasElement
      });
      
      // Now calculate scaled dimensions using movie dimensions
      const scaledDims = calculateScaledDimensions(
        videoDimensions.width, 
        videoDimensions.height, 
        movie.width || canvasWidth,  // Use movie width or fallback to canvas width
        movie.height || canvasHeight // Use movie height or fallback to canvas height
      );
      console.log('Scaled dimensions:', scaledDims);
      console.log('Movie dimensions after creation:', movie.width, 'x', movie.height);
      
      let currentStartTime = 0;
      
      // Create video layers for each highlight
      for (let i = 0; i < highlights.length; i++) {
        const highlight = highlights[i];
        const videoURL = videoURLs.get(highlight.filePath);
        
        if (!videoURL) {
          console.warn(`Skipping highlight ${i}: no video URL for ${highlight.filePath}`);
          continue;
        }
        
        const segmentDuration = highlight.end - highlight.start;
        
        console.log(`Creating layer ${i}: ${highlight.videoClipName} (${segmentDuration}s)`);
        console.log(`Layer ${i} settings:`, {
          layerSize: { width: movie.width || canvasWidth, height: movie.height || canvasHeight },
          destPosition: { x: scaledDims.x, y: scaledDims.y },
          destSize: { width: scaledDims.width, height: scaledDims.height }
        });
        
        // Create video layer with proper destination sizing
        const videoLayer = new etro.layer.Video({
          startTime: currentStartTime,
          duration: segmentDuration,
          source: videoURL,
          sourceStartTime: highlight.start,
          // Layer position and size (covers full canvas)
          x: 0,
          y: 0,
          width: movie.width || canvasWidth,
          height: movie.height || canvasHeight,
          // Video rendering within the layer
          destX: scaledDims.x,
          destY: scaledDims.y,
          destWidth: scaledDims.width,
          destHeight: scaledDims.height
        });
        
        movie.layers.push(videoLayer);
        currentStartTime += segmentDuration;
      }
      
      totalDuration = currentStartTime;
      console.log(`Etro movie created with total duration: ${totalDuration}s`);
      console.log('Movie details - width:', movie.width, 'height:', movie.height, 'layers:', movie.layers.length);
      console.log('Movie paused state:', movie.paused, 'ready state:', movie.ready);
      
      isInitialized = true;
      return true;
      
    } catch (err) {
      console.error('Failed to create Etro movie:', err);
      initializationError = err.message;
      return false;
    }
  }

  // Update time and highlight index from Etro movie
  function updateTimeAndHighlight() {
    if (!movie) return;
    
    currentTime = movie.currentTime;
    
    // Sync our state with Etro's actual state
    isPaused = movie.paused;
    isPlaying = !movie.paused;
    
    // Determine current highlight based on timeline
    let highlightIndex = 0;
    let accumulatedTime = 0;
    
    for (let i = 0; i < highlights.length; i++) {
      const segmentDuration = highlights[i].end - highlights[i].start;
      if (currentTime < accumulatedTime + segmentDuration) {
        highlightIndex = i;
        break;
      }
      accumulatedTime += segmentDuration;
      highlightIndex = i + 1; // In case we're past all segments
    }
    
    currentHighlightIndex = Math.min(highlightIndex, highlights.length - 1);
    
    // Check if playback has ended
    if (movie.ended) {
      isPlaying = false;
      isPaused = false;
      stopProgressTracking();
      console.log('Playback ended');
    }
  }

  // Playback controls
  async function startPlayback() {
    console.log('startPlayback called');
    console.log('State check - allVideosLoaded:', allVideosLoaded, 'isInitialized:', isInitialized, 'movie:', !!movie);
    
    if (!allVideosLoaded) {
      toast.error('Videos still loading');
      console.log('Playback blocked: videos still loading');
      return;
    }
    
    if (!isInitialized || !movie) {
      toast.error('Video player not ready');
      console.log('Playback blocked: player not ready - isInitialized:', isInitialized, 'movie:', !!movie);
      return;
    }
    
    try {
      console.log('Movie state before play - paused:', movie.paused, 'currentTime:', movie.currentTime, 'layers:', movie.layers.length);
      console.log('Movie ready state:', movie.ready);
      
      // Set playing state immediately to show playhead
      isPlaying = true;
      isPaused = false;
      startProgressTracking();
      
      // Wait for movie to be ready if it's not
      if (!movie.ready) {
        console.log('Movie not ready, waiting...');
        // Simple polling to wait for ready state
        let attempts = 0;
        while (!movie.ready && attempts < 50) { // Wait up to 5 seconds
          await new Promise(resolve => setTimeout(resolve, 100));
          attempts++;
          console.log('Waiting for movie ready... attempt', attempts);
        }
        
        if (!movie.ready) {
          throw new Error('Movie failed to become ready after 5 seconds');
        }
        console.log('Movie is now ready!');
      }
      
      // Only play if not already playing
      if (movie.paused) {
        console.log('Calling movie.play()...');
        await movie.play();
        console.log('movie.play() completed');
      } else {
        console.log('Movie was already playing');
      }
      
      console.log('Playback started successfully');
    } catch (err) {
      console.error('Error starting playback:', err);
      isPlaying = false;
      isPaused = false;
      stopProgressTracking();
      toast.error('Failed to start playback: ' + err.message);
    }
  }

  async function togglePlayback() {
    if (!movie || !isInitialized) return;
    
    try {
      if (movie.paused) {
        // Movie is paused, resume playback
        await movie.play();
        isPaused = false;
        isPlaying = true;
        startProgressTracking();
      } else {
        // Movie is playing, pause it
        movie.pause();
        isPaused = true;
        stopProgressTracking();
      }
    } catch (err) {
      console.error('Error toggling playback:', err);
    }
  }

  function stopPlayback() {
    if (movie) {
      movie.pause();
      movie.currentTime = 0;
    }
    
    isPlaying = false;
    isPaused = false;
    currentTime = 0;
    currentHighlightIndex = 0;
    
    stopProgressTracking();
  }

  // Jump to a specific highlight
  async function jumpToHighlight(highlightIndex) {
    if (!movie || highlightIndex < 0 || highlightIndex >= highlights.length) return;
    
    // Calculate time at start of target highlight
    let targetTime = 0;
    for (let i = 0; i < highlightIndex; i++) {
      targetTime += highlights[i].end - highlights[i].start;
    }
    
    console.log(`Jumping to highlight ${highlightIndex} at time ${targetTime}s`);
    movie.currentTime = targetTime;
    
    // Continue playing if we were already playing
    if (isPlaying && movie.paused) {
      try {
        await movie.play();
      } catch (err) {
        if (!err.message.includes('Already playing')) {
          console.error('Error resuming playback:', err);
        }
      }
    }
  }

  // Progress tracking
  function startProgressTracking() {
    stopProgressTracking();
    
    function updateProgress() {
      if (!movie || !isPlaying) return;
      
      updateTimeAndHighlight();
      
      if (isPlaying && !isPaused && !movie.ended) {
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
    if (!movie || !isInitialized) return;
    
    movie.currentTime = Math.max(0, Math.min(targetTime, totalDuration));
    
    // Resume playing if we were playing before seeking
    if (isPlaying && movie.paused) {
      try {
        await movie.play();
      } catch (err) {
        // Ignore "already playing" errors
        if (!err.message.includes('Already playing')) {
          console.error('Error resuming playback after seek:', err);
        }
      }
    }
  }

  // Progress percentage for timeline
  function getProgressPercentage() {
    return totalDuration > 0 ? (currentTime / totalDuration) * 100 : 0;
  }

  // Watch for when videos are loaded to initialize
  $effect(() => {
    if (allVideosLoaded && highlights.length > 0 && !isInitialized && canvasElement) {
      console.log('Effect: Creating Etro movie with', highlights.length, 'highlights');
      createEtroMovie();
    }
  });

  // Watch for highlights changes and reinitialize if needed
  $effect(() => {
    if (highlights.length > 0) {
      console.log('Effect: Highlights changed, checking initialization state');
      console.log('Current state - allVideosLoaded:', allVideosLoaded, 'isInitialized:', isInitialized, 'videoURLs size:', videoURLs.size);
      
      if (!allVideosLoaded) {
        console.log('Effect: Starting video URL loading for', highlights.length, 'highlights');
        loadVideoURLs();
      }
    }
  });

  // Initialize component
  onMount(async () => {
    console.log('EtroVideoPlayer mounted with highlights:', highlights);
    
    // Wait for canvas element to be ready
    const waitForCanvas = () => {
      return new Promise((resolve) => {
        const checkCanvas = () => {
          if (canvasElement) {
            console.log('Canvas element is ready');
            resolve();
          } else {
            setTimeout(checkCanvas, 50);
          }
        };
        checkCanvas();
      });
    };
    
    await waitForCanvas();
    
    // The reactive effects will handle initialization when highlights are available
    if (highlights.length > 0) {
      console.log('First highlight:', highlights[0]);
      console.log('Highlight file paths:', highlights.map(h => h.filePath));
    }
  });

  // Cleanup
  onDestroy(() => {
    stopProgressTracking();
    
    if (movie) {
      movie.pause();
    }
  });
</script>

{#if highlights.length > 0}
  <div class="video-player p-6 bg-card border rounded-lg">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold">Etro Video Player</h3>
      <div class="text-sm text-muted-foreground">
        {highlights.length} highlights • {formatTime(totalDuration)} total
        <br>
        <span class="text-xs">
          URLs: {videoURLs.size}/{highlights.length} • 
          Ready: {allVideosLoaded ? 'Yes' : 'No'} • 
          Init: {isInitialized ? 'Yes' : 'No'}
        </span>
      </div>
    </div>
    
    <!-- Canvas Element for Etro rendering -->
    <div class="relative w-full aspect-video bg-black overflow-hidden mb-4">
      <canvas
        bind:this={canvasElement}
        class="w-full h-full bg-black"
        style="object-fit: contain; max-width: 100%; max-height: 100%;"
      ></canvas>
      
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
            <p>Initializing Etro video player...</p>
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
              {formatTime(currentTime)} / {formatTime(totalDuration)}
            </div>
            <div class="text-xs text-muted-foreground">
              {Math.round(getProgressPercentage())}%
            </div>
          </div>
        </div>
      </div>
    {/if}
    
    <!-- Simplified Clip Timeline -->
    <div class="timeline-container mb-4">
      <div class="space-y-2">
        <!-- Clip segments -->
        <div class="flex gap-1 w-full">
          {#each highlights as highlight, index}
            {@const segmentDuration = highlight.end - highlight.start}
            {@const segmentWidth = totalDuration > 0 ? (segmentDuration / totalDuration) * 100 : 0}
            {@const isActive = index === currentHighlightIndex}
            
            <button
              class="relative h-8 rounded transition-all duration-200 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-primary/50 {isActive ? 'ring-2 ring-primary' : ''}"
              style="width: {segmentWidth}%; background-color: {highlight.color}; min-width: 20px;"
              title="{highlight.videoClipName}: {formatTime(highlight.start)} - {formatTime(highlight.end)}"
              onclick={() => jumpToHighlight(index)}
            >
              <!-- Progress indicator for active segment -->
              {#if isActive && (isPlaying || !isPaused)}
                {@const segmentStartTime = highlights.slice(0, index).reduce((sum, h) => sum + (h.end - h.start), 0)}
                {@const segmentProgress = Math.max(0, Math.min(1, (currentTime - segmentStartTime) / segmentDuration))}
                <div 
                  class="absolute left-0 top-0 h-full bg-white/30 rounded transition-all duration-100"
                  style="width: {segmentProgress * 100}%;"
                ></div>
              {/if}
              
              <!-- Segment label -->
              <div class="absolute inset-0 flex items-center justify-center text-xs font-medium text-white drop-shadow">
                {index + 1}
              </div>
            </button>
          {/each}
        </div>
        
        <!-- Time display -->
        <div class="flex justify-between text-xs text-muted-foreground">
          <span>{formatTime(currentTime)}</span>
          <span>Clip {currentHighlightIndex + 1} of {highlights.length}</span>
          <span>{formatTime(totalDuration)}</span>
        </div>
      </div>
    </div>
    
    <!-- Simplified Controls -->
    <div class="playback-controls flex items-center justify-center gap-3">
      <Button 
        onclick={isPlaying ? togglePlayback : startPlayback} 
        disabled={!allVideosLoaded || !isInitialized}
        class="flex items-center gap-2"
      >
        {#if !isPlaying}
          <Play class="w-4 h-4" />
          Play All Clips
        {:else if isPaused}
          <Play class="w-4 h-4" />
          Resume
        {:else}
          <Pause class="w-4 h-4" />
          Pause
        {/if}
      </Button>
      
      <!-- Debug info -->
      <div class="text-xs text-muted-foreground ml-4">
        Movie: {movie ? (movie.paused ? 'Paused' : 'Playing') : 'Not ready'} | 
        Time: {currentTime.toFixed(1)}s
      </div>
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