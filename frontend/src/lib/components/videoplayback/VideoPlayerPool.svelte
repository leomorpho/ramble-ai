<script>
  import { onMount, onDestroy } from 'svelte';

  let {
    highlights = [],
    videoURLs = new Map(),
    onPlayerReady = () => {},
    onTimeUpdate = () => {},
    poolSize = 3
  } = $props();

  // Video player pool
  let videoPool = $state([]);
  let activePlayerIndex = $state(0);
  let nextPlayerIndex = $state(1);
  let bufferPlayerIndex = $state(2);
  
  // Track time update handlers for cleanup
  let timeUpdateHandlers = new Map();
  
  // Pool state
  let poolInitialized = $state(false);

  // Create video player pool
  function createVideoPool() {
    videoPool = Array.from({ length: poolSize }, (_, index) => ({
      element: null,
      index,
      state: 'idle', // idle, loading, ready, playing, error
      segmentIndex: -1,
      videoURL: '',
      startTime: 0,
      ready: false
    }));
    
    poolInitialized = true;
  }

  // Video element references for binding
  let videoElements = $state([]);

  // Initialize video elements
  function initializeVideoElements() {
    videoElements = Array.from({ length: poolSize }, () => null);
    
    // Update pool with element references
    videoPool.forEach((player, index) => {
      player.elementIndex = index;
    });
  }

  // Get next available player for loading
  function getNextAvailablePlayer() {
    // Try designated next player first
    if (videoPool[nextPlayerIndex] && (videoPool[nextPlayerIndex].state === 'idle' || videoPool[nextPlayerIndex].state === 'error')) {
      return nextPlayerIndex;
    }
    
    // Try buffer player
    if (videoPool[bufferPlayerIndex] && (videoPool[bufferPlayerIndex].state === 'idle' || videoPool[bufferPlayerIndex].state === 'error')) {
      return bufferPlayerIndex;
    }
    
    // Find any idle player (excluding active)
    for (let i = 0; i < videoPool.length; i++) {
      if (i !== activePlayerIndex && videoPool[i] && (videoPool[i].state === 'idle' || videoPool[i].state === 'error')) {
        return i;
      }
    }
    
    return -1;
  }

  // Load video into specific player
  async function loadVideoIntoPlayer(playerIndex, segmentIndex) {
    if (playerIndex < 0 || playerIndex >= videoPool.length) return false;
    if (segmentIndex < 0 || segmentIndex >= highlights.length) return false;
    
    const player = videoPool[playerIndex];
    const highlight = highlights[segmentIndex];
    const videoURL = videoURLs.get(highlight.filePath);
    const videoElement = videoElements[playerIndex];
    
    if (!videoURL || !videoElement) return false;
    
    try {
      player.state = 'loading';
      player.segmentIndex = segmentIndex;
      player.videoURL = videoURL;
      player.startTime = highlight.start;
      
      // Set source and wait for it to load
      videoElement.src = videoURL;
      
      await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('Load timeout'));
        }, 5000);
        
        const onCanPlay = () => {
          clearTimeout(timeout);
          videoElement.removeEventListener('canplay', onCanPlay);
          videoElement.removeEventListener('error', onError);
          resolve();
        };
        
        const onError = () => {
          clearTimeout(timeout);
          videoElement.removeEventListener('canplay', onCanPlay);
          videoElement.removeEventListener('error', onError);
          reject(new Error('Video load error'));
        };
        
        videoElement.addEventListener('canplay', onCanPlay);
        videoElement.addEventListener('error', onError);
        
        // Force load
        videoElement.load();
      });
      
      // Set start time and end constraint
      videoElement.currentTime = highlight.start;
      player.state = 'ready';
      player.ready = true;
      player.endTime = highlight.end;
      
      // Add time update listener with segment boundary checking
      const timeUpdateHandler = (event) => {
        if (player.index === activePlayerIndex) {
          const currentTime = event.target.currentTime;
          
          // Check if we've exceeded the segment end time
          if (currentTime >= highlight.end) {
            console.log(`Segment ${segmentIndex} ended at ${currentTime}, should be ${highlight.end}`);
            event.target.pause();
            onTimeUpdate(event); // Trigger the end-of-segment logic
            return;
          }
          
          onTimeUpdate(event);
        }
      };
      
      // Remove existing listener if any
      const existingHandler = timeUpdateHandlers.get(playerIndex);
      if (existingHandler) {
        videoElement.removeEventListener('timeupdate', existingHandler);
      }
      
      // Store and add new listener
      timeUpdateHandlers.set(playerIndex, timeUpdateHandler);
      videoElement.addEventListener('timeupdate', timeUpdateHandler);
      
      console.log(`Player ${playerIndex} loaded segment ${segmentIndex} (${highlight.videoClipName}) from ${highlight.start}s to ${highlight.end}s`);
      return true;
      
    } catch (err) {
      console.error(`Failed to load segment ${segmentIndex} into player ${playerIndex}:`, err);
      player.state = 'error';
      player.ready = false;
      return false;
    }
  }

  // Switch to next player instantly
  function switchToPlayer(targetPlayerIndex) {
    if (targetPlayerIndex < 0 || targetPlayerIndex >= videoPool.length) return false;
    
    const currentPlayer = videoPool[activePlayerIndex];
    const nextPlayer = videoPool[targetPlayerIndex];
    const currentElement = videoElements[activePlayerIndex];
    const nextElement = videoElements[targetPlayerIndex];
    
    if (!nextPlayer.ready || nextPlayer.state !== 'ready' || !nextElement) {
      console.warn(`Player ${targetPlayerIndex} not ready for switching`);
      return false;
    }
    
    // Hide current player
    if (currentElement) {
      currentElement.style.display = 'none';
      currentElement.pause();
      currentPlayer.state = 'idle';
    }
    
    // Show and play next player
    nextElement.style.display = 'block';
    nextElement.currentTime = nextPlayer.startTime;
    
    // Ensure we start from the correct time
    const playPromise = nextElement.play();
    if (playPromise !== undefined) {
      playPromise.catch(error => {
        console.error('Error playing video:', error);
      });
    }
    
    nextPlayer.state = 'playing';
    
    // Update indices
    activePlayerIndex = targetPlayerIndex;
    rotatePlayerIndices();
    
    console.log(`Switched to player ${targetPlayerIndex} for segment ${nextPlayer.segmentIndex}`);
    return true;
  }

  // Rotate player indices for next cycle
  function rotatePlayerIndices() {
    const oldNext = nextPlayerIndex;
    const oldBuffer = bufferPlayerIndex;
    const oldActive = activePlayerIndex;
    
    // Find next available indices (skip active player)
    const availableIndices = videoPool.map((_, i) => i).filter(i => i !== activePlayerIndex);
    
    nextPlayerIndex = availableIndices[0] || (oldNext + 1) % poolSize;
    bufferPlayerIndex = availableIndices[1] || (oldBuffer + 1) % poolSize;
    
    // Ensure no conflicts
    if (nextPlayerIndex === activePlayerIndex) nextPlayerIndex = (nextPlayerIndex + 1) % poolSize;
    if (bufferPlayerIndex === activePlayerIndex || bufferPlayerIndex === nextPlayerIndex) {
      bufferPlayerIndex = (bufferPlayerIndex + 1) % poolSize;
    }
  }

  // Cleanup player (free memory)
  function cleanupPlayer(playerIndex) {
    if (playerIndex < 0 || playerIndex >= videoPool.length) return;
    
    const player = videoPool[playerIndex];
    const videoElement = videoElements[playerIndex];
    
    if (videoElement) {
      // Remove all event listeners
      const handler = timeUpdateHandlers.get(playerIndex);
      if (handler) {
        videoElement.removeEventListener('timeupdate', handler);
        timeUpdateHandlers.delete(playerIndex);
      }
      
      if (player.index !== activePlayerIndex) {
        videoElement.pause();
        videoElement.src = '';
        videoElement.load();
      }
      
      player.state = 'idle';
      player.segmentIndex = -1;
      player.ready = false;
      player.videoURL = '';
      player.endTime = 0;
    }
  }

  // Get current active player
  function getActivePlayer() {
    return videoPool[activePlayerIndex];
  }
  
  // Get current time from active player
  function getCurrentTime() {
    const activePlayer = videoPool[activePlayerIndex];
    const videoElement = videoElements[activePlayer?.index];
    return videoElement ? videoElement.currentTime : 0;
  }

  // Get player by index
  function getPlayer(index) {
    return videoPool[index];
  }

  // Check if player is ready for segment
  function isPlayerReady(playerIndex, segmentIndex) {
    const player = videoPool[playerIndex];
    return player && player.ready && player.segmentIndex === segmentIndex && player.state === 'ready';
  }

  // Expose methods to parent
  onMount(() => {
    createVideoPool();
    initializeVideoElements();
    
    // Notify parent that pool is ready
    onPlayerReady({
      loadVideoIntoPlayer,
      switchToPlayer,
      getActivePlayer,
      getPlayer,
      isPlayerReady,
      cleanupPlayer,
      getNextAvailablePlayer,
      getCurrentTime
    });
  });

  // Cleanup on destroy
  onDestroy(() => {
    videoElements.forEach((element, index) => {
      if (element) {
        element.pause();
        element.src = '';
      }
    });
  });
</script>

<!-- Video pool container -->
<div class="relative w-full aspect-video bg-black overflow-hidden">
  {#if poolInitialized}
    {#each videoElements as videoElement, index (index)}
      <video
        bind:this={videoElements[index]}
        class="absolute inset-0 w-full h-full bg-black"
        style="display: none;"
        preload="auto"
        muted={false}
        controls={false}
      >
        <track kind="captions" />
      </video>
    {/each}
  {/if}
  
  <!-- Loading indicator -->
  {#if !poolInitialized}
    <div class="absolute inset-0 flex items-center justify-center bg-black text-white">
      <div class="text-center">
        <div class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"></div>
        <p>Initializing video pool...</p>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Ensure video elements are properly positioned */
  :global(.video-pool video) {
    object-fit: contain;
  }
</style>