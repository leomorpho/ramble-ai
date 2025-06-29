<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import VideoPlayerPool from './VideoPlayerPool.svelte';
  import PlaybackTimeline from './PlaybackTimeline.svelte';
  import PlaybackControls from './PlaybackControls.svelte';
  import CurrentHighlightInfo from './CurrentHighlightInfo.svelte';

  let { highlights = [] } = $props();
  
  // Player state
  let isPlaying = $state(false);
  let isPaused = $state(false);
  let currentHighlightIndex = $state(0);
  let videoURLs = $state(new Map()); // Map of filePath -> videoURL
  let loadingProgress = $state(0);
  let allVideosLoaded = $state(false);
  
  // Video pool management
  let videoPoolAPI = $state(null);
  let preloadingNextSegment = $state(false);
  
  // Progress tracking
  let virtualTime = $state(0);
  let totalVirtualDuration = $state(0);
  let segmentStartTimes = $state([]);
  
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

  // Load video URLs only (no complex preloading)
  async function loadVideoURLs() {
    if (highlights.length === 0) return;
    
    loadingProgress = 0;
    videoURLs.clear();
    
    const uniqueVideos = new Map();
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
        videoURLs.set(highlight.filePath, videoURL);
        loadedCount++;
        loadingProgress = (loadedCount / videoFiles.length) * 100;
      } catch (err) {
        console.error('Error loading video URL for:', highlight.filePath, err);
        toast.error('Failed to load video', {
          description: `Could not load ${highlight.videoClipName}: ${err.message}`
        });
      }
    }
    
    allVideosLoaded = loadedCount === videoFiles.length;
    
    if (allVideosLoaded) {
      toast.success('All videos ready!');
    }
  }

  // Handle video pool ready
  function handleVideoPoolReady(api) {
    videoPoolAPI = api;
    console.log('Video pool ready');
    
    // Load first video if everything is ready
    if (allVideosLoaded && highlights.length > 0) {
      loadFirstVideo();
    }
  }

  // Watch for when everything is ready
  $effect(() => {
    if (videoPoolAPI && allVideosLoaded && highlights.length > 0) {
      // Load first video into pool when ready
      videoPoolAPI.loadVideoIntoPlayer(0, 0);
    }
  });

  // Start playing the sequence
  async function startPlayback() {
    if (!allVideosLoaded) {
      toast.error('Videos still loading');
      return;
    }
    
    if (highlights.length === 0) return;
    
    if (!videoPoolAPI) {
      toast.error('Video player not ready');
      return;
    }
    
    currentHighlightIndex = 0;
    virtualTime = 0;
    isPlaying = true;
    isPaused = false;
    
    // Switch to first player and start playing
    const success = videoPoolAPI.switchToPlayer(0);
    if (success) {
      startProgressTracking();
      preloadNextSegments();
    } else {
      toast.error('Failed to start playback');
      stopPlayback();
    }
  }

  // Preload next segments for seamless playback
  async function preloadNextSegments() {
    if (!videoPoolAPI || preloadingNextSegment) return;
    
    preloadingNextSegment = true;
    
    try {
      console.log(`Preloading segments after ${currentHighlightIndex}`);
      
      // Load next segment into next available player
      if (currentHighlightIndex + 1 < highlights.length) {
        // Check if next segment is already loaded
        let nextAlreadyLoaded = false;
        for (let i = 0; i < 3; i++) {
          if (videoPoolAPI.isPlayerReady(i, currentHighlightIndex + 1)) {
            nextAlreadyLoaded = true;
            break;
          }
        }
        
        if (!nextAlreadyLoaded) {
          const nextPlayerIndex = videoPoolAPI.getNextAvailablePlayer();
          if (nextPlayerIndex >= 0) {
            console.log(`Preloading segment ${currentHighlightIndex + 1} into player ${nextPlayerIndex}`);
            await videoPoolAPI.loadVideoIntoPlayer(nextPlayerIndex, currentHighlightIndex + 1);
          }
        }
      }
      
      // Load segment after next if available
      if (currentHighlightIndex + 2 < highlights.length) {
        let bufferAlreadyLoaded = false;
        for (let i = 0; i < 3; i++) {
          if (videoPoolAPI.isPlayerReady(i, currentHighlightIndex + 2)) {
            bufferAlreadyLoaded = true;
            break;
          }
        }
        
        if (!bufferAlreadyLoaded) {
          const bufferPlayerIndex = videoPoolAPI.getNextAvailablePlayer();
          if (bufferPlayerIndex >= 0) {
            console.log(`Preloading segment ${currentHighlightIndex + 2} into player ${bufferPlayerIndex}`);
            await videoPoolAPI.loadVideoIntoPlayer(bufferPlayerIndex, currentHighlightIndex + 2);
          }
        }
      }
    } catch (err) {
      console.error('Error preloading segments:', err);
    } finally {
      preloadingNextSegment = false;
    }
  }

  // Handle video time updates
  function handleVideoTimeUpdate(event) {
    if (!isPlaying || !videoPoolAPI) return;
    
    const highlight = highlights[currentHighlightIndex];
    if (!highlight) return;
    
    const activePlayer = videoPoolAPI.getActivePlayer();
    if (!activePlayer) return;
    
    const currentTime = event.target.currentTime;
    
    console.log(`Time update: ${currentTime.toFixed(2)}s, segment ends at ${highlight.end}s`);
    
    // Check if we've reached the end of the current highlight
    if (currentTime >= highlight.end - 0.1) { // Small buffer for timing precision
      console.log('Segment ended, switching to next highlight');
      playNextHighlight();
      return;
    }
    
    // Preload next segments when approaching end (3 seconds before for better timing)
    const timeToEnd = highlight.end - currentTime;
    if (timeToEnd <= 3 && !preloadingNextSegment) {
      console.log('Preloading next segments');
      preloadNextSegments();
    }
  }

  // Seek to a specific time in the virtual timeline
  async function seekToTime(targetSegmentIndex, segmentTime) {
    if (targetSegmentIndex < 0 || targetSegmentIndex >= highlights.length) return;
    
    // Update current segment
    currentHighlightIndex = targetSegmentIndex;
    const highlight = highlights[currentHighlightIndex];
    const seekTime = Math.max(highlight.start, Math.min(highlight.end, highlight.start + segmentTime));
    
    // Update virtual time
    virtualTime = segmentStartTimes[currentHighlightIndex] + (seekTime - highlight.start);
    
    // If we were playing, load the segment and seek
    if (isPlaying && videoPoolAPI) {
      // Load the target segment into active player
      const success = await videoPoolAPI.loadVideoIntoPlayer(0, targetSegmentIndex);
      if (success) {
        const activePlayer = videoPoolAPI.getActivePlayer();
        if (activePlayer) {
          videoPoolAPI.switchToPlayer(activePlayer.index);
          // The seek time is already set during loading in the pool
        }
      }
    }
  }

  // Play next highlight
  async function playNextHighlight() {
    const previousIndex = currentHighlightIndex;
    currentHighlightIndex++;
    
    if (currentHighlightIndex >= highlights.length) {
      stopPlayback();
      toast.success('Sequence completed!');
      return;
    }
    
    console.log(`Switching from highlight ${previousIndex} to ${currentHighlightIndex}`);
    
    // Switch to next available player with preloaded content
    if (videoPoolAPI) {
      // First, check if any player already has this segment loaded
      for (let i = 0; i < 3; i++) {
        if (videoPoolAPI.isPlayerReady(i, currentHighlightIndex)) {
          console.log(`Found preloaded segment ${currentHighlightIndex} in player ${i}`);
          const success = videoPoolAPI.switchToPlayer(i);
          if (success) {
            preloadNextSegments(); // Continue preloading
            return;
          }
        }
      }
      
      // No preloaded segment found, load into next available player
      const nextPlayerIndex = videoPoolAPI.getNextAvailablePlayer();
      if (nextPlayerIndex >= 0) {
        console.log(`Loading segment ${currentHighlightIndex} into player ${nextPlayerIndex}`);
        const success = await videoPoolAPI.loadVideoIntoPlayer(nextPlayerIndex, currentHighlightIndex);
        if (success) {
          videoPoolAPI.switchToPlayer(nextPlayerIndex);
          preloadNextSegments();
        } else {
          console.error(`Failed to load segment ${currentHighlightIndex}`);
        }
      } else {
        console.error('No available player found');
      }
    }
  }

  // Play previous highlight
  async function playPreviousHighlight() {
    currentHighlightIndex = Math.max(0, currentHighlightIndex - 1);
    virtualTime = segmentStartTimes[currentHighlightIndex] || 0;
    
    if (videoPoolAPI) {
      // Load previous segment
      const success = await videoPoolAPI.loadVideoIntoPlayer(0, currentHighlightIndex);
      if (success) {
        videoPoolAPI.switchToPlayer(0);
        preloadNextSegments();
      }
    }
  }

  // Toggle play/pause
  function togglePlayback() {
    if (!videoPoolAPI) return;
    
    const activePlayer = videoPoolAPI.getActivePlayer();
    if (!activePlayer) return;
    
    if (isPaused) {
      // Resume playback
      videoPoolAPI.switchToPlayer(activePlayer.index);
      isPaused = false;
      startProgressTracking();
    } else {
      // Pause playback
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
    
    if (videoPoolAPI) {
      const activePlayer = videoPoolAPI.getActivePlayer();
      if (activePlayer) {
        // Cleanup active player
        videoPoolAPI.cleanupPlayer(activePlayer.index);
      }
    }
    
    stopProgressTracking();
  }

  // Start tracking progress for smooth updates
  function startProgressTracking() {
    stopProgressTracking();
    
    function updateProgress() {
      if (!isPlaying || !videoPoolAPI) return;
      
      const highlight = highlights[currentHighlightIndex];
      if (!highlight) return;
      
      const segmentStartTime = segmentStartTimes[currentHighlightIndex] || 0;
      const currentVideoTime = videoPoolAPI.getCurrentTime();
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

  // Cleanup
  onDestroy(() => {
    stopProgressTracking();
    if (videoPoolAPI) {
      // Cleanup all players in the pool
      for (let i = 0; i < 3; i++) {
        videoPoolAPI.cleanupPlayer(i);
      }
    }
  });

  // Initialize with proper change detection
  let initialized = false;
  let lastHighlightsLength = 0;
  
  onMount(() => {
    initializePlayer();
  });
  
  // Watch for highlights changes with proper change detection
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
      loadVideoURLs();
    }
  }
  
  function reinitializePlayer() {
    if (isPlaying) {
      stopPlayback();
    }
    
    videoURLs.clear();
    allVideosLoaded = false;
    loadingProgress = 0;
    lastHighlightsLength = highlights.length; // Update tracking variable
    
    calculateVirtualTimeline();
    
    if (highlights.length > 0) {
      loadVideoURLs();
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
    
    <!-- Video Player Pool -->
    <VideoPlayerPool 
      {highlights}
      {videoURLs}
      onPlayerReady={handleVideoPoolReady}
      onTimeUpdate={handleVideoTimeUpdate}
      poolSize={3}
    />
    
    <!-- Current Highlight Info -->
    <CurrentHighlightInfo 
      {isPlaying}
      {currentHighlightIndex}
      {highlights}
    />
    
    <!-- Playback Timeline -->
    <PlaybackTimeline 
      {highlights}
      {virtualTime}
      {totalVirtualDuration}
      {currentHighlightIndex}
      {isPlaying}
      {segmentStartTimes}
      onSeek={seekToTime}
    />
    
    <!-- Playback Controls -->
    <PlaybackControls 
      {isPlaying}
      {isPaused}
      {allVideosLoaded}
      {currentHighlightIndex}
      {highlights}
      onStartPlayback={startPlayback}
      onTogglePlayback={togglePlayback}
      onPreviousHighlight={playPreviousHighlight}
      onNextHighlight={playNextHighlight}
      onStopPlayback={stopPlayback}
    />
  </div>
{/if}