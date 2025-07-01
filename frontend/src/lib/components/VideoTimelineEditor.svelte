<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoClipsByProject } from '$lib/wailsjs/go/main/App';
  import { toast } from 'svelte-sonner';
  import { ZoomIn, ZoomOut, RotateCw, Play, Pause, SkipBack, SkipForward } from '@lucide/svelte';
  import { Button } from "$lib/components/ui/button";

  let { 
    highlight = null,
    projectId = null,
    videoElement = null,
    currentTime = 0,
    duration = 0,
    isPlaying = false,
    editedStart = $bindable(0),
    editedEnd = $bindable(0),
    onSeek = () => {},
    onTogglePlay = () => {}
  } = $props();

  // Timeline state
  let zoomLevel = $state(4); // Higher zoom for word-level detail
  let zoomCenter = $state(0); // Center point of zoom (in seconds)
  let timelineContainer = $state(null);
  let isDraggingMarker = $state(false);
  let dragMarkerType = $state(''); // 'start' or 'end'
  
  // Transcription words
  let transcriptionWords = $state([]);
  let loadingWords = $state(false);
  
  // Virtual scrolling state
  let scrollTop = $state(0);
  let containerHeight = $state(200);
  let wordHeight = $state(24); // Height of each word row in pixels
  let visibleBuffer = $state(5); // Extra items to render above/below visible area
  let wordsContainer = $state(null);
  let autoScrolling = $state(false); // Prevent manual scroll conflicts

  // Calculate zoom and timeline bounds
  let visibleStartTime = $derived(() => {
    const halfWindow = duration / (2 * zoomLevel);
    return Math.max(0, zoomCenter - halfWindow);
  });
  
  let visibleEndTime = $derived(() => {
    const halfWindow = duration / (2 * zoomLevel);
    return Math.min(duration, zoomCenter + halfWindow);
  });
  
  let visibleDuration = $derived(() => visibleEndTime - visibleStartTime);

  // Virtual scrolling calculations
  let totalHeight = $derived(() => transcriptionWords.length * wordHeight);
  let scrollableHeight = $derived(() => Math.max(0, totalHeight - containerHeight));
  let startIndex = $derived(() => {
    if (transcriptionWords.length === 0) return 0;
    return Math.max(0, Math.floor(scrollTop / wordHeight) - visibleBuffer);
  });
  let endIndex = $derived(() => {
    if (transcriptionWords.length === 0) return 0;
    const visibleCount = Math.ceil(containerHeight / wordHeight);
    return Math.min(transcriptionWords.length, startIndex + visibleCount + visibleBuffer * 2);
  });
  let visibleWords = $derived(() => {
    if (transcriptionWords.length === 0) return [];
    return transcriptionWords.slice(startIndex, endIndex);
  });
  let offsetY = $derived(() => startIndex * wordHeight);
  
  // Find current word index
  let currentWordIndex = $state(-1);
  
  // Simple function to find current word
  function findCurrentWordIndex(time, words) {
    if (words.length === 0) return -1;
    
    // First try to find exact match
    let exactIndex = words.findIndex(word => 
      time >= word.start && time <= word.end
    );
    
    if (exactIndex >= 0) return exactIndex;
    
    // If no exact match, find the closest word
    let closestIndex = 0;
    let closestDistance = Math.abs(words[0].start - time);
    
    for (let i = 1; i < words.length; i++) {
      const distance = Math.abs(words[i].start - time);
      if (distance < closestDistance) {
        closestDistance = distance;
        closestIndex = i;
      }
    }
    
    // Only return closest if we're within 2 seconds of a word
    return closestDistance <= 2 ? closestIndex : -1;
  }
  
  // Update current word index
  $effect(() => {
    if (transcriptionWords.length === 0) {
      currentWordIndex = -1;
      return;
    }
    
    currentWordIndex = findCurrentWordIndex(currentTime, transcriptionWords);
  });

  // Load transcription words when highlight changes
  $effect(() => {
    if (highlight) {
      loadTranscriptionWords();
      setDefaultZoom();
    }
  });

  async function loadTranscriptionWords() {
    if (!highlight) return;
    
    if (!projectId) {
      console.warn('No projectId provided to VideoTimelineEditor');
      transcriptionWords = [];
      return;
    }
    
    loadingWords = true;
    try {
      // We need to get the video clip data to access transcription words
      // The highlight has videoClipId which we can use
      console.log('Loading transcription words for project:', projectId, 'videoClipId:', highlight.videoClipId);
      const clips = await GetVideoClipsByProject(projectId);
      console.log('Got clips:', clips.length, clips);
      const clip = clips.find(c => c.id === highlight.videoClipId);
      console.log('Found clip:', clip);
      
      if (clip && clip.transcriptionWords && clip.transcriptionWords.length > 0) {
        transcriptionWords = clip.transcriptionWords;
        console.log('Loaded transcription words:', transcriptionWords.length, transcriptionWords.slice(0, 5));
      } else {
        transcriptionWords = [];
        console.warn('No transcription words found for video clip', { 
          clip: !!clip, 
          transcriptionWords: clip?.transcriptionWords?.length,
          clipData: clip 
        });
      }
    } catch (error) {
      console.error('Failed to load transcription words:', error);
      transcriptionWords = [];
    } finally {
      loadingWords = false;
    }
  }

  function setDefaultZoom() {
    if (!highlight || !duration) return;
    
    const highlightDuration = editedEnd - editedStart;
    // Set zoom to show highlight plus some context
    zoomLevel = Math.min(Math.max(4, duration / (highlightDuration * 3)), 20);
    zoomCenter = (editedStart + editedEnd) / 2;
  }

  // Convert time to timeline position (0-1)
  function timeToPosition(time) {
    if (visibleDuration === 0) return 0;
    return (time - visibleStartTime) / visibleDuration;
  }

  // Convert timeline position to time
  function positionToTime(position) {
    return visibleStartTime + (position * visibleDuration);
  }

  // Handle timeline click
  function handleTimelineClick(event) {
    if (!timelineContainer || isDraggingMarker || duration === 0) return;
    
    const rect = timelineContainer.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const position = x / rect.width;
    const targetTime = position * duration;
    
    onSeek(targetTime);
  }

  // Marker dragging
  function handleMarkerMouseDown(event, markerType) {
    event.stopPropagation();
    isDraggingMarker = true;
    dragMarkerType = markerType;
    
    // Add global mouse events
    document.addEventListener('mousemove', handleGlobalMouseMove);
    document.addEventListener('mouseup', handleGlobalMouseUp);
  }

  function handleGlobalMouseMove(event) {
    if (!isDraggingMarker || !timelineContainer || duration === 0) return;
    
    const rect = timelineContainer.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const position = Math.max(0, Math.min(1, x / rect.width));
    const newTime = position * duration;
    
    if (dragMarkerType === 'start') {
      editedStart = Math.max(0, Math.min(newTime, editedEnd - 0.1));
    } else if (dragMarkerType === 'end') {
      editedEnd = Math.min(duration, Math.max(newTime, editedStart + 0.1));
    }
  }

  function handleGlobalMouseUp() {
    isDraggingMarker = false;
    dragMarkerType = '';
    
    // Remove global mouse events
    document.removeEventListener('mousemove', handleGlobalMouseMove);
    document.removeEventListener('mouseup', handleGlobalMouseUp);
  }

  // Zoom controls
  function zoomIn() {
    zoomLevel = Math.min(zoomLevel * 1.5, 20);
  }

  function zoomOut() {
    zoomLevel = Math.max(zoomLevel / 1.5, 1);
  }

  function resetZoom() {
    zoomLevel = 1;
    zoomCenter = duration / 2;
  }

  // Scroll to specific word index (center it in the view)
  function scrollToWordIndex(wordIndex) {
    if (wordIndex < 0 || !wordsContainer || transcriptionWords.length === 0) return;
    
    autoScrolling = true;
    // Center the word in the container
    const targetScrollTop = wordIndex * wordHeight - containerHeight / 2 + wordHeight / 2;
    const maxScrollTop = (transcriptionWords.length * wordHeight) - containerHeight;
    const newScrollTop = Math.max(0, Math.min(targetScrollTop, maxScrollTop));
    
    wordsContainer.scrollTo({
      top: newScrollTop,
      behavior: 'smooth'
    });
    
    // Reset auto-scrolling flag after scroll completes
    setTimeout(() => {
      autoScrolling = false;
    }, 500);
  }

  // Auto-scroll to current word
  $effect(() => {
    if (currentWordIndex >= 0 && wordsContainer && !autoScrolling) {
      scrollToWordIndex(currentWordIndex);
    }
  });

  // Handle scroll
  function handleScroll(event) {
    if (!autoScrolling) {
      scrollTop = event.target.scrollTop;
    }
  }

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    const ms = Math.floor((seconds % 1) * 1000);
    return `${mins}:${secs.toString().padStart(2, '0')}.${ms.toString().padStart(3, '0')}`;
  }

  // Check if word is within current highlight
  function isWordInHighlight(word) {
    return word.start < editedEnd && word.end > editedStart;
  }

  // Check if word is currently being spoken
  function isWordActive(word) {
    return currentTime >= word.start && currentTime <= word.end;
  }

  // Handle keyboard events
  function handleKeyDown(event) {
    if (event.code === 'Space') {
      event.preventDefault(); // Prevent page scrolling
      onTogglePlay();
    }
  }

  // Add keyboard event listener when component mounts
  onMount(() => {
    document.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    document.removeEventListener('mousemove', handleGlobalMouseMove);
    document.removeEventListener('mouseup', handleGlobalMouseUp);
    document.removeEventListener('keydown', handleKeyDown);
  });
</script>

<div class="video-timeline-editor flex flex-col h-full">
  <!-- Timeline Controls -->
  <div class="timeline-controls flex items-center justify-between p-3 border-b bg-background">
    <div class="flex items-center gap-2">
      <Button variant="outline" size="sm" onclick={onTogglePlay}>
        {#if isPlaying}
          <Pause class="w-4 h-4" />
        {:else}
          <Play class="w-4 h-4" />
        {/if}
      </Button>
      
      <div class="text-sm text-muted-foreground">
        {formatTime(currentTime)} / {formatTime(duration)}
      </div>
    </div>
    
    <div class="flex items-center gap-2">
      <Button variant="outline" size="sm" onclick={zoomOut}>
        <ZoomOut class="w-4 h-4" />
      </Button>
      <Button variant="outline" size="sm" onclick={zoomIn}>
        <ZoomIn class="w-4 h-4" />
      </Button>
      <Button variant="outline" size="sm" onclick={resetZoom}>
        <RotateCw class="w-4 h-4" />
      </Button>
    </div>
  </div>

  <!-- Simplified Timeline -->
  <div class="border rounded-lg">
    <div class="p-4">
      <h3 class="text-sm font-medium mb-2">
        Video Timeline with Transcription 
        <span class="text-xs text-muted-foreground">
          (Words: {transcriptionWords.length}, Current Time: {currentTime.toFixed(2)}s, Loading: {loadingWords})
        </span>
      </h3>
      
      <!-- Basic timeline bar -->
      <div 
        bind:this={timelineContainer}
        class="relative h-12 bg-secondary rounded-lg cursor-pointer overflow-hidden mb-4"
        onclick={handleTimelineClick}
      >
        <!-- Timeline background -->
        <div class="absolute inset-0 bg-secondary"></div>
        
        <!-- Current playhead -->
        {#if duration > 0}
          <div 
            class="absolute top-0 w-0.5 h-full bg-primary z-30"
            style="left: {(currentTime / duration) * 100}%"
          ></div>
        {/if}
        
        <!-- Highlight region -->
        {#if duration > 0}
          <div 
            class="absolute top-0 h-full bg-blue-500/30"
            style="left: {(editedStart / duration) * 100}%; width: {((editedEnd - editedStart) / duration) * 100}%"
          ></div>
        {/if}
        
        <!-- Start marker -->
        {#if duration > 0}
          <div 
            class="absolute top-0 w-2 h-full bg-green-500 z-20 cursor-ew-resize"
            style="left: calc({(editedStart / duration) * 100}% - 4px)"
            onmousedown={(e) => handleMarkerMouseDown(e, 'start')}
          ></div>
        {/if}
        
        <!-- End marker -->
        {#if duration > 0}
          <div 
            class="absolute top-0 w-2 h-full bg-red-500 z-20 cursor-ew-resize"
            style="left: calc({(editedEnd / duration) * 100}% - 4px)"
            onmousedown={(e) => handleMarkerMouseDown(e, 'end')}
          ></div>
        {/if}
      </div>
      
      <!-- Words display -->
      <div class="border rounded bg-background/50" style="height: {containerHeight}px;">
        {#if loadingWords}
          <div class="flex items-center justify-center h-full text-muted-foreground">
            <div class="animate-pulse">Loading transcription words...</div>
          </div>
        {:else if transcriptionWords.length === 0}
          <div class="flex items-center justify-center h-full text-muted-foreground">
            <div class="text-center">
              <p>No transcription words available</p>
              <p class="text-xs">Transcribe the video to see word-level timing</p>
            </div>
          </div>
        {:else}
          <div 
            bind:this={wordsContainer}
            class="h-full overflow-y-auto"
            onscroll={handleScroll}
          >
            <div class="p-2">
              {#each transcriptionWords as word, wordIndex}
                {@const isInHighlight = word.start < editedEnd && word.end > editedStart}
                {@const isCurrentWord = currentWordIndex >= 0 && wordIndex === currentWordIndex}
                
                <div 
                  class="flex items-center gap-2 p-1 rounded text-xs transition-colors
                         {isCurrentWord ? 'bg-orange-500/30 border-l-4 border-orange-500 font-bold shadow-sm' : isInHighlight ? 'bg-blue-500/10 border-l-2 border-blue-500' : ''}"
                  style="height: {wordHeight}px;"
                >
                  <span class="font-mono text-muted-foreground min-w-[60px]">
                    {formatTime(word.start)}
                  </span>
                  <span class="flex-1 {isCurrentWord ? 'text-orange-900 dark:text-orange-100' : ''}">{word.word}</span>
                  <button 
                    class="opacity-50 hover:opacity-100"
                    onclick={() => onSeek(word.start)}
                    title="Seek to word"
                  >
                    <Play class="w-3 h-3" />
                  </button>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .video-timeline-editor {
    height: 300px;
  }
  
</style>