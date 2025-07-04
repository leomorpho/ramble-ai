<script>
  import { onMount, onDestroy } from "svelte";
  import { toast } from "svelte-sonner";
  import { Play, Pause } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import {
    updateHighlightOrder,
    deleteHighlight,
    editHighlight,
  } from "$lib/stores/projectHighlights.js";
  import { browser } from "$app/environment";
  import ClipEditor from "$lib/components/ClipEditor.svelte";
  import TimelineSegment from "./TimelineSegment.svelte";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$lib/components/ui/dialog";
  import { Film } from "@lucide/svelte";

  // Import utility functions
  import {
    formatTime,
    calculateSeekTime,
    isDragHandleClick,
  } from "./timelineUtils.js";

  let {
    highlights = [],
    projectId = null,
    enableEyeButton = true,
    onReorder = null,
    enableReordering = true,
    debounceDelay=5000,
  } = $props();

  // Core state
  let videoElement = $state(null);
  let isPlaying = $state(false);
  let currentTime = $state(0);
  let totalDuration = $state(0);
  let currentHighlightIndex = $state(0);
  let currentVideoSource = $state("");

  // Drag and drop state (use highlights prop directly from store)
  let isDragging = $state(false);
  let dragStartIndex = $state(-1);
  let dragOverIndex = $state(-1);

  // Loading state
  let isLoading = $state(false);

  // Player initialization
  let isInitialized = $state(false);

  // Seeking state
  let isSeeking = $state(false);
  let isAutoTransitioning = $state(false);

  // Preloading state
  let preloadedHighlight = $state(null);
  let preloadedVideoElement = $state(null);
  let isPreloading = $state(false);

  // Track highlight order to detect external changes
  let lastKnownOrder = $state("");
  let isInternalReorder = $state(false);

  // Popover state management for eye icon menus
  let popoverStates = $state(new Map());

  // Clip editor state
  let clipEditorOpen = $state(false);
  let editingHighlight = $state(null);

  // Delete confirmation state
  let deleteDialogOpen = $state(false);
  let highlightToDelete = $state(null);
  let deleting = $state(false);

  // Calculate total duration from highlights
  let calculatedTotalDuration = $derived(() => {
    const duration = highlights.reduce((sum, h) => sum + (h.end - h.start), 0);
    return duration;
  });
  
  // Active segment display threshold
  const ACTIVE_SEGMENT_THRESHOLD = 0.2; // Show if any segment is less than 20% of total duration

  // Calculate if we should show the active segment based on highlight durations
  let shouldShowActiveSegment = $derived(() => {
    // Don't show for 0 or 1 highlights
    if (highlights.length <= 1) return false;

    const totalDurationCalc = calculatedTotalDuration;
    if (totalDurationCalc === 0) return false;

    // Check if any highlight is less than the threshold percentage of total duration
    return highlights.some((h) => {
      const segmentDuration = h.end - h.start;
      const percentage = segmentDuration / totalDurationCalc;
      return percentage < ACTIVE_SEGMENT_THRESHOLD;
    });
  });
  
  // Update total duration when highlights change
  $effect(() => {
    totalDuration = calculatedTotalDuration;
  });

  // Load and play a specific highlight
  async function loadHighlight(highlight) {
    if (!highlight) return false;
    
    try {
      isLoading = true;
      const videoURL = encodeURI(highlight.filePath);
      
      // Set the video source with fragment URL for the specific segment
      const fragmentURL = `${videoURL}#t=${highlight.start},${highlight.end}`;
      
      currentVideoSource = fragmentURL;
      
      // If video element is available, also set src directly
      if (videoElement) {
        videoElement.src = fragmentURL;
        videoElement.load();
      }
      
      return true;
    } catch (err) {
      console.error("Failed to load highlight:", err);
      toast.error("Failed to load video", {
        description: `Could not load ${highlight.videoClipName}`
      });
      return false;
    } finally {
      isLoading = false;
    }
  }

  // Preload the next highlight
  async function preloadNextHighlight(nextHighlight) {
    if (!nextHighlight || isPreloading) return false;
    
    try {
      isPreloading = true;
      console.log("Preloading next highlight:", nextHighlight.videoClipName);
      
      const videoURL = encodeURI(nextHighlight.filePath);
      const fragmentURL = `${videoURL}#t=${nextHighlight.start},${nextHighlight.end}`;
      
      // Create a hidden video element for preloading
      const preloadVideo = document.createElement('video');
      preloadVideo.src = fragmentURL;
      preloadVideo.preload = 'metadata';
      preloadVideo.style.display = 'none';
      document.body.appendChild(preloadVideo);
      
      // Store the preloaded data
      preloadedHighlight = nextHighlight;
      preloadedVideoElement = preloadVideo;
      
      return true;
    } catch (err) {
      console.error("Failed to preload highlight:", err);
      return false;
    } finally {
      isPreloading = false;
    }
  }

  // Use preloaded highlight if available
  async function usePreloadedHighlight() {
    if (!preloadedHighlight || !preloadedVideoElement) return false;
    
    try {
      console.log("Using preloaded highlight:", preloadedHighlight.videoClipName);
      
      // Transfer the preloaded source to the main video element
      const fragmentURL = preloadedVideoElement.src;
      currentVideoSource = fragmentURL;
      
      if (videoElement) {
        videoElement.src = fragmentURL;
        videoElement.load();
      }
      
      // Clean up preloaded element
      cleanupPreloadedElement();
      
      return true;
    } catch (err) {
      console.error("Failed to use preloaded highlight:", err);
      cleanupPreloadedElement();
      return false;
    }
  }

  // Clean up preloaded elements
  function cleanupPreloadedElement() {
    if (preloadedVideoElement) {
      if (preloadedVideoElement.parentNode) {
        preloadedVideoElement.parentNode.removeChild(preloadedVideoElement);
      }
      preloadedVideoElement = null;
    }
    preloadedHighlight = null;
    isPreloading = false;
  }

  // Update current highlight index based on current time
  function updateCurrentHighlightIndex() {
    let accumulatedTime = 0;
    for (let i = 0; i < highlights.length; i++) {
      const segmentDuration = highlights[i].end - highlights[i].start;
      if (currentTime >= accumulatedTime && currentTime < accumulatedTime + segmentDuration) {
        currentHighlightIndex = i;
        break;
      }
      accumulatedTime += segmentDuration;
    }
  }

  // Playback controls
  async function playPauseWrapper() {
    if (!videoElement || !isInitialized) {
      toast.error("Video player not ready");
      return;
    }
    
    try {
      if (videoElement.paused || videoElement.ended) {
        await videoElement.play();
        isPlaying = true;
      } else {
        videoElement.pause();
        isPlaying = false;
      }
    } catch (err) {
      console.error("Error toggling playback:", err);
      toast.error("Failed to toggle playback");
    }
  }

  // Jump to a specific highlight
  async function jumpToHighlightWrapper(highlightIndex) {
    if (highlightIndex < 0 || highlightIndex >= highlights.length) return;
    
    // Clean up any preloaded element since we're jumping manually
    cleanupPreloadedElement();
    
    const targetHighlight = highlights[highlightIndex];
    const wasPlaying = isPlaying;
    
    // Pause current playback
    if (videoElement && !videoElement.paused) {
      videoElement.pause();
      isPlaying = false;
    }
    
    // Load the new highlight
    const success = await loadHighlight(targetHighlight);
    if (success) {
      currentHighlightIndex = highlightIndex;
      
      // Calculate the start time in the concatenated timeline
      let accumulatedTime = 0;
      for (let i = 0; i < highlightIndex; i++) {
        accumulatedTime += highlights[i].end - highlights[i].start;
      }
      currentTime = accumulatedTime;
      
      // Resume playing if it was playing before
      if (wasPlaying && videoElement) {
        try {
          await videoElement.play();
          isPlaying = true;
        } catch (err) {
          console.error("Failed to resume playback:", err);
        }
      }
    }
  }

  // Handle video time updates
  function handleTimeUpdate() {
    if (!videoElement) return;
    
    // Update the current time in the context of the concatenated timeline
    let accumulatedTime = 0;
    for (let i = 0; i < currentHighlightIndex; i++) {
      accumulatedTime += highlights[i].end - highlights[i].start;
    }
    
    const currentHighlight = highlights[currentHighlightIndex];
    if (currentHighlight) {
      // Calculate time within the current highlight segment
      const videoCurrentTime = videoElement.currentTime;
      const timeWithinSegment = Math.max(0, videoCurrentTime - currentHighlight.start);
      currentTime = accumulatedTime + timeWithinSegment;
      
      // Check if we should preload the next highlight (3 seconds before end)
      const timeUntilEnd = currentHighlight.end - videoCurrentTime;
      const nextIndex = currentHighlightIndex + 1;
      if (timeUntilEnd <= 3 && nextIndex < highlights.length && !preloadedHighlight && !isPreloading) {
        const nextHighlight = highlights[nextIndex];
        if (nextHighlight) {
          preloadNextHighlight(nextHighlight);
        }
      }
      
      // Check if we've reached the end of the current highlight
      if (videoCurrentTime >= currentHighlight.end) {
        console.log(`Reached end of highlight ${currentHighlightIndex + 1} at time ${videoCurrentTime}/${currentHighlight.end}`);
        handleHighlightEnd();
      }
    }
    
    updateCurrentHighlightIndex();
  }

  // Handle when a highlight reaches its end
  async function handleHighlightEnd() {
    const nextIndex = currentHighlightIndex + 1;
    if (nextIndex < highlights.length) {
      console.log(`Auto-advancing from highlight ${currentHighlightIndex + 1} to ${nextIndex + 1}`);
      const wasPlaying = isPlaying;
      
      // Set transition flag to prevent seeking indicator
      isAutoTransitioning = true;
      
      try {
        // Try to use preloaded highlight first, fallback to normal loading
        const nextHighlight = highlights[nextIndex];
        let success = false;
        
        if (preloadedHighlight && preloadedHighlight.id === nextHighlight.id) {
          console.log("Using preloaded highlight for seamless transition");
          success = await usePreloadedHighlight();
        }
        
        if (!success) {
          console.log("Fallback to normal highlight loading");
          await jumpToHighlightWrapper(nextIndex);
        } else {
          // Update the current highlight index manually since we used preloaded
          currentHighlightIndex = nextIndex;
          
          // Calculate the start time in the concatenated timeline
          let accumulatedTime = 0;
          for (let i = 0; i < nextIndex; i++) {
            accumulatedTime += highlights[i].end - highlights[i].start;
          }
          currentTime = accumulatedTime;
        }
        
        // Resume playing if it was playing before
        if (wasPlaying && videoElement && videoElement.paused) {
          try {
            await videoElement.play();
            isPlaying = true;
          } catch (err) {
            console.error("Failed to auto-play next highlight:", err);
          }
        }
      } finally {
        // Clear transition flag after a small delay to ensure all events are processed
        setTimeout(() => {
          isAutoTransitioning = false;
        }, 100);
      }
    } else {
      console.log("Reached end of all highlights");
      isPlaying = false;
      if (videoElement && !videoElement.paused) {
        videoElement.pause();
      }
    }
  }

  // Timeline seeking
  async function handleTimelineSeekWrapper(targetTime) {
    
    // Find which highlight contains the target time
    let accumulatedTime = 0;
    let targetHighlightIndex = -1;
    let targetHighlight = null;
    let timeBeforeTarget = 0;
    
    for (let i = 0; i < highlights.length; i++) {
      const segmentDuration = highlights[i].end - highlights[i].start;
      if (targetTime >= accumulatedTime && targetTime < accumulatedTime + segmentDuration) {
        targetHighlightIndex = i;
        targetHighlight = highlights[i];
        timeBeforeTarget = accumulatedTime;
        break;
      }
      accumulatedTime += segmentDuration;
    }
    
    if (!targetHighlight) {
      return;
    }
    
    const wasPlaying = isPlaying;
    isSeeking = true;
    
    try {
      // If we need to switch to a different highlight
      if (targetHighlightIndex !== currentHighlightIndex) {
        await jumpToHighlightWrapper(targetHighlightIndex);
      }
      
      // Calculate the exact time within the video file
      if (videoElement && targetHighlight) {
        const timeWithinTimeline = targetTime - timeBeforeTarget;
        const videoSeekTime = targetHighlight.start + timeWithinTimeline;
        
        videoElement.currentTime = videoSeekTime;
        currentTime = targetTime;
        
        // Update the display immediately
        updateCurrentHighlightIndex();
      }
      
      // Resume playing if it was playing before
      if (wasPlaying && videoElement && videoElement.paused) {
        try {
          await videoElement.play();
          isPlaying = true;
        } catch (err) {
          console.error("Failed to resume playback after seek:", err);
        }
      }
    } finally {
      isSeeking = false;
    }
  }

  // Handle timeline segment clicks for seeking
  function handleSegmentClick(event, segmentIndex) {
    // Check if the click was on the drag handle
    if (isDragHandleClick(event)) {
      // This is a drag handle click, don't seek
      return;
    }

    // Calculate the target time using utility function
    const targetTime = calculateSeekTime(event, segmentIndex, highlights);

    // Seek to the calculated time
    handleTimelineSeekWrapper(targetTime);
  }

  // Progress percentage
  function getProgressPercentage() {
    return totalDuration > 0 ? (currentTime / totalDuration) * 100 : 0;
  }

  // Helper functions for popover state management
  function openPopover(highlightId) {
    const newStates = new Map(popoverStates);
    newStates.set(highlightId, true);
    popoverStates = newStates;
  }

  function closePopover(highlightId) {
    const newStates = new Map(popoverStates);
    newStates.set(highlightId, false);
    popoverStates = newStates;
  }

  function isPopoverOpen(highlightId) {
    return popoverStates.get(highlightId) || false;
  }

  // Handle edit highlight
  function handleEditHighlight(event, highlight) {
    if (event) {
      event.stopPropagation();
    }
    closePopover(highlight.id);
    editingHighlight = highlight;
    clipEditorOpen = true;
  }

  // Handle highlight save from editor
  async function handleHighlightSave(updatedHighlight) {
    // Use the store's editHighlight function to ensure both components react
    const updates = {
      id: updatedHighlight.id,
      start: updatedHighlight.start,
      end: updatedHighlight.end,
      color: updatedHighlight.color,
    };

    await editHighlight(
      updatedHighlight.id,
      updatedHighlight.videoClipId,
      updates
    );
  }

  // Handle delete confirmation
  function handleDeleteConfirm(event, highlight) {
    if (event) {
      event.stopPropagation();
    }
    closePopover(highlight.id);
    highlightToDelete = highlight;
    deleteDialogOpen = true;
  }

  // Handle delete highlight
  async function handleDeleteHighlight() {
    if (!highlightToDelete) return;

    deleting = true;

    try {
      const success = await deleteHighlight(
        highlightToDelete.id,
        highlightToDelete.videoClipId
      );

      if (success) {
        deleteDialogOpen = false;
        highlightToDelete = null;
      }
    } catch (error) {
      console.error("Error deleting highlight:", error);
    } finally {
      deleting = false;
    }
  }

  // Cancel delete
  function cancelDelete() {
    deleteDialogOpen = false;
    highlightToDelete = null;
  }

  // Drag and drop functions
  function handleDragStart(event, index) {
    // Check if reordering is enabled
    if (!enableReordering) {
      event.preventDefault();
      return false;
    }

    // Check if the drag started from the drag handle
    if (!isDragHandleClick(event)) {
      // Prevent drag if not started from the handle
      event.preventDefault();
      return false;
    }

    isDragging = true;
    dragStartIndex = index;
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", index.toString());
  }

  function handleDragEnd() {
    isDragging = false;
    dragStartIndex = -1;
    dragOverIndex = -1;
  }

  function handleDragOver(event, targetIndex) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";

    // Calculate insertion point based on mouse position within the target
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const isLeftHalf = x < rect.width / 2;

    // Determine where the item would be inserted
    let insertionIndex;
    if (isLeftHalf) {
      insertionIndex = targetIndex;
    } else {
      insertionIndex = targetIndex + 1;
    }

    // Adjust for the dragged item being removed first
    if (dragStartIndex !== -1 && dragStartIndex < insertionIndex) {
      insertionIndex--;
    }

    dragOverIndex = insertionIndex;
  }

  async function handleDrop(event, targetIndex) {
    event.preventDefault();

    if (dragStartIndex === -1 || dragStartIndex === targetIndex) {
      handleDragEnd();
      return;
    }

    const newHighlights = [...highlights];
    const draggedItem = newHighlights[dragStartIndex];

    // Remove dragged item
    newHighlights.splice(dragStartIndex, 1);

    // Insert at new position
    const insertIndex =
      dragStartIndex < targetIndex ? targetIndex - 1 : targetIndex;
    newHighlights.splice(insertIndex, 0, draggedItem);

    // Mark as internal reorder to prevent external change detection
    isInternalReorder = true;

    let success = false;

    if (onReorder) {
      // Use custom reorder handler (for preview mode)
      try {
        await onReorder(newHighlights);
        success = true;
      } catch (error) {
        console.error("Error in custom reorder handler:", error);
        toast.error("Failed to reorder highlights");
      }
    } else {
      // Update via centralized store (this handles database save and state updates)
      success = await updateHighlightOrder(newHighlights);
    }

    if (success) {
      // Update our known order
      lastKnownOrder = newHighlights.map((h) => h.id).join(",");
      // Update the video player with new order
      await handleReorderComplete(newHighlights);
    }

    // Reset the internal reorder flag
    isInternalReorder = false;

    handleDragEnd();
  }

  // Handle reordering by updating the current highlight
  async function handleReorderComplete(newHighlights) {
    console.log("Handling reorder with new highlight order");
    
    // Clean up any preloaded element since order changed
    cleanupPreloadedElement();
    
    // Reset to first highlight with new order
    if (newHighlights.length > 0) {
      const firstHighlight = newHighlights[0];
      const success = await loadHighlight(firstHighlight);
      if (success) {
        currentHighlightIndex = 0;
        currentTime = 0;
      }
    }
  }

  // Initialize video when highlights are available
  $effect(() => {
    if (browser && highlights.length > 0 && !isInitialized) {
      console.log("Initializing video player with", highlights.length, "highlights");
      console.log("First highlight:", JSON.stringify(highlights[0], null, 2));
      
      // Load the first highlight
      const firstHighlight = highlights[0];
      if (firstHighlight) {
        console.log("Loading first highlight:", firstHighlight.videoClipName);
        loadHighlight(firstHighlight).then(success => {
          console.log("First highlight load result:", success);
          if (success) {
            isInitialized = true;
            currentHighlightIndex = 0;
            currentTime = 0;
            console.log("Video player initialized successfully");
          } else {
            console.error("Failed to initialize video player");
          }
        });
      }
    }
  });

  // Watch for highlight order changes
  $effect(() => {
    if (browser && highlights.length > 0 && isInitialized) {
      const currentOrder = highlights.map((h) => h.id).join(",");
      
      // If order changed, reset to first highlight
      if (lastKnownOrder && lastKnownOrder !== currentOrder && !isInternalReorder) {
        console.log("Highlight order changed, resetting to first highlight");
        const firstHighlight = highlights[0];
        if (firstHighlight) {
          loadHighlight(firstHighlight).then(() => {
            currentHighlightIndex = 0;
            currentTime = 0;
          });
        }
      }
      
      lastKnownOrder = currentOrder;
    }
  });



  // Sync playing state with video element and handle auto-progression
  $effect(() => {
    if (browser && videoElement) {
      const handlePlay = () => { isPlaying = true; };
      const handlePause = () => { isPlaying = false; };
      const handleEnded = async () => { 
        isPlaying = false;
        
        // Auto-advance to next highlight if available
        const nextIndex = currentHighlightIndex + 1;
        if (nextIndex < highlights.length) {
          console.log(`Auto-advancing from highlight ${currentHighlightIndex + 1} to ${nextIndex + 1}`);
          await jumpToHighlightWrapper(nextIndex);
          
          // Auto-play the next highlight
          if (videoElement && videoElement.paused) {
            try {
              await videoElement.play();
              isPlaying = true;
            } catch (err) {
              console.error("Failed to auto-play next highlight:", err);
            }
          }
        } else {
          console.log("Reached end of all highlights");
        }
      };
      
      videoElement.addEventListener('play', handlePlay);
      videoElement.addEventListener('pause', handlePause);
      videoElement.addEventListener('ended', handleEnded);
      
      return () => {
        videoElement.removeEventListener('play', handlePlay);
        videoElement.removeEventListener('pause', handlePause);
        videoElement.removeEventListener('ended', handleEnded);
      };
    }
  });



  // Initialize component
  onMount(() => {
    console.log("EtroVideoPlayer mounted");
    console.log("Highlights on mount:", highlights.length);
    console.log("videoElement on mount:", videoElement);
  });

  // Cleanup
  onDestroy(() => {
    if (videoElement && !videoElement.paused) {
      videoElement.pause();
    }
    // Clean up any preloaded elements
    cleanupPreloadedElement();
  });
</script>

{#if highlights.length > 0}
  <div class="video-player p-6 bg-card border rounded-lg">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="text-sm text-muted-foreground">
        {highlights.length} highlights • {formatTime(totalDuration)} total
        <br />
      </div>
    </div>

    <!-- HTML5 Video Element -->
    <div
      class="relative w-full aspect-video bg-black overflow-hidden mb-4 rounded"
    >
      {#if currentVideoSource}
        <video
          bind:this={videoElement}
          class="w-full h-full bg-black"
          style="object-fit: contain; max-width: 100%; max-height: 100%;"
          src={currentVideoSource}
          preload="metadata"
          ontimeupdate={handleTimeUpdate}
          onloadeddata={() => { isInitialized = true; }}
          onwaiting={() => { 
            if (!isAutoTransitioning) {
              isSeeking = true; 
            }
          }}
          oncanplay={() => { 
            if (!isAutoTransitioning) {
              isSeeking = false; 
            }
          }}
        >
          <track kind="captions" />
        </video>
      {:else}
        <div class="w-full h-full bg-black flex items-center justify-center text-white">
          <div class="text-center">
            <p>No video selected</p>
          </div>
        </div>
      {/if}

      <!-- Loading indicator -->
      {#if isLoading}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black/50 text-white"
        >
          <div class="text-center">
            <div
              class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"
            ></div>
            <p>Loading video...</p>
          </div>
        </div>
      {/if}

      <!-- Seeking indicator -->
      {#if isSeeking}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black/50 text-white"
        >
          <div class="text-center">
            <div
              class="animate-spin w-6 h-6 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"
            ></div>
            <p class="text-sm">Seeking...</p>
          </div>
        </div>
      {/if}
    </div>

    <!-- Current Highlight Info -->
    {#if highlights[currentHighlightIndex]}
      <div class="bg-secondary/30 p-3 rounded-md mb-4">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="font-medium text-sm">
              {highlights[currentHighlightIndex].videoClipName}
            </h4>
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

    <!-- Draggable Clip Timeline -->
    <div class="timeline-container mb-4">
      <div class="space-y-2">
        {#if enableReordering}
          <div class="text-xs text-muted-foreground mb-2">
            💡 Click segments to seek, drag handle (⚫) to reorder
          </div>
        {:else}
          <div class="text-xs text-muted-foreground mb-2">
            💡 Click segments to seek
          </div>
        {/if}

        <!-- Clip segments with drag and drop -->
        <div class="flex w-full">
          {#each highlights as highlight, index}
            {@const segmentDuration = highlight.end - highlight.start}
            {@const calculatedTotalDuration = highlights.reduce(
              (sum, h) => sum + (h.end - h.start),
              0
            )}
            {@const segmentWidth =
              calculatedTotalDuration > 0
                ? (segmentDuration / calculatedTotalDuration) * 100
                : 100 / highlights.length}
            {@const isActive = index === currentHighlightIndex}

            <!-- Drop indicator before this segment -->
            {#if enableReordering && isDragging && dragOverIndex === index}
              {@render dropIndicator()}
            {/if}

            <TimelineSegment
              {highlight}
              {index}
              {isActive}
              {segmentWidth}
              {currentTime}
              {totalDuration}
              {highlights}
              {enableReordering}
              enableEyeButton={enableEyeButton && !shouldShowActiveSegment}
              {isDragging}
              {dragStartIndex}
              {isPopoverOpen}
              {openPopover}
              {closePopover}
              isFirst={index === 0}
              isLast={index === highlights.length - 1}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              onDragOver={handleDragOver}
              onDrop={handleDrop}
              onSegmentClick={handleSegmentClick}
              onEditHighlight={handleEditHighlight}
              onDeleteConfirm={handleDeleteConfirm}
            />

            <!-- Drop indicator after the last segment -->
            {#if enableReordering && index === highlights.length - 1 && isDragging && dragOverIndex === highlights.length}
              {@render dropIndicator()}
            {/if}
          {/each}
        </div>

        <!-- Active segment in full width -->
        {#if shouldShowActiveSegment && highlights[currentHighlightIndex]}
          {@const activeHighlight = highlights[currentHighlightIndex]}
          {@const segmentStartTime = highlights
            .slice(0, currentHighlightIndex)
            .reduce((sum, h) => sum + (h.end - h.start), 0)}
          {@const segmentDuration = activeHighlight.end - activeHighlight.start}

          <div class="mt-1">
            <div class="w-full">
              <TimelineSegment
                highlight={activeHighlight}
                index={currentHighlightIndex}
                isActive={true}
                isFirst={true}
                isLast={true}
                segmentWidth={100}
                {currentTime}
                {totalDuration}
                {highlights}
                enableReordering={false}
                enableEyeButton={true}
                isDragging={false}
                dragStartIndex={null}
                {isPopoverOpen}
                {openPopover}
                {closePopover}
                onDragStart={() => {}}
                onDragEnd={() => {}}
                onDragOver={() => {}}
                onDrop={() => {}}
                onSegmentClick={(e) => {
                  // Calculate click position to seek within current segment
                  const rect = e.currentTarget.getBoundingClientRect();
                  const x = e.clientX - rect.left;
                  const clickPercentage = x / rect.width;
                  const clickTargetTime =
                    segmentStartTime + clickPercentage * segmentDuration;
                  handleTimelineSeekWrapper(clickTargetTime);
                }}
                onEditHighlight={handleEditHighlight}
                onDeleteConfirm={handleDeleteConfirm}
              />
            </div>
          </div>
        {/if}

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
      {#key isPlaying}
        <Button
          onclick={playPauseWrapper}
          disabled={!isInitialized || isLoading}
          class="flex items-center gap-2"
        >
          {#if isPlaying}
            <Pause class="w-4 h-4" />
            Pause
          {:else}
            <Play class="w-4 h-4" />
            Play
          {/if}
        </Button>
      {/key}
    </div>

    <!-- Loading Progress (removed since we load on-demand now) -->
  </div>
{:else}
  <div class="video-player p-6 bg-card border rounded-lg">
    <div class="text-center text-muted-foreground">
      <p>No video highlights available</p>
    </div>
  </div>
{/if}

{#snippet dropIndicator()}
  <div class="w-0.5 h-8 bg-black dark:bg-white rounded flex-shrink-0"></div>
{/snippet}

<!-- Clip Editor -->
<ClipEditor
  bind:open={clipEditorOpen}
  highlight={editingHighlight}
  {projectId}
  onSave={handleHighlightSave}
/>

<!-- Delete Confirmation Dialog -->
<Dialog bind:open={deleteDialogOpen}>
  <DialogContent class="sm:max-w-[425px]">
    <DialogHeader>
      <DialogTitle>Delete Highlight</DialogTitle>
      <DialogDescription>
        Are you sure you want to delete this highlight? This action cannot be
        undone.
      </DialogDescription>
    </DialogHeader>

    {#if highlightToDelete}
      <div class="space-y-3">
        <div
          class="flex items-center gap-3 p-3 rounded-lg border"
          style="background-color: {highlightToDelete.color}20; border-left: 4px solid {highlightToDelete.color};"
        >
          <Film
            class="w-6 h-6 flex-shrink-0"
            style="color: {highlightToDelete.color}"
          />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">
              {highlightToDelete.videoClipName}
            </h3>
            <p class="text-sm text-muted-foreground">
              {formatTime(highlightToDelete.start)} - {formatTime(
                highlightToDelete.end
              )}
            </p>
            {#if highlightToDelete.text}
              <p class="text-sm mt-1 italic line-clamp-2">
                "{highlightToDelete.text}"
              </p>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={cancelDelete} disabled={deleting}>
        Cancel
      </Button>
      <Button
        variant="destructive"
        onclick={handleDeleteHighlight}
        disabled={deleting}
      >
        {#if deleting}
          Deleting...
        {:else}
          Delete Highlight
        {/if}
      </Button>
    </div>
  </DialogContent>
</Dialog>
