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
  import { loadVideoURLs } from "./videoUtils.js";
  import {
    formatTime,
    updateTimeAndHighlight,
    getProgressPercentage,
    calculateSeekTime,
    isDragHandleClick,
  } from "./timelineUtils.js";
  import { createEtroMovieWithOrder } from "./etroUtils.js";
  import { createProgressTracker } from "./progressUtils.js";
  import {
    playPause as playPauseUtil,
    jumpToHighlight as jumpToHighlightUtil,
    handleTimelineSeek as handleTimelineSeekUtil,
  } from "./playbackUtils.js";

  let {
    highlights = [],
    projectId = null,
    enableEyeButton = true,
    onReorder = null,
    enableReordering = true,
  } = $props();

  // Core state
  let canvasElement = $state(null);
  let movie = $state(null);
  let isPlaying = $state(false);
  let currentTime = $state(0);
  let totalDuration = $state(0);
  let currentHighlightIndex = $state(0);

  // Drag and drop state (use highlights prop directly from store)
  let isDragging = $state(false);
  let dragStartIndex = $state(-1);
  let dragOverIndex = $state(-1);

  // Video URLs and loading
  let videoURLs = $state(new Map());
  let loadingProgress = $state(0);
  let allVideosLoaded = $state(false);

  // Player initialization
  let isInitialized = $state(false);
  let initializationError = $state(null);

  // Buffering state for seeking
  let isBuffering = $state(false);

  // Progress tracker instance
  const progressTracker = createProgressTracker();

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

  // Active segment display threshold
  const ACTIVE_SEGMENT_THRESHOLD = 0.2; // Show if any segment is less than 20% of total duration

  // Calculate if we should show the active segment based on highlight durations
  let shouldShowActiveSegment = $derived(() => {
    // Don't show for 0 or 1 highlights
    if (highlights.length <= 1) return false;

    const totalDurationCalc = highlights.reduce(
      (sum, h) => sum + (h.end - h.start),
      0
    );
    if (totalDurationCalc === 0) return false;

    // Check if any highlight is less than the threshold percentage of total duration
    return highlights.some((h) => {
      const segmentDuration = h.end - h.start;
      const percentage = segmentDuration / totalDurationCalc;
      console.log(
        "segmentDuration",
        segmentDuration,
        "totalDurationCalc",
        totalDurationCalc,
        "percentage",
        percentage
      );
      return percentage < ACTIVE_SEGMENT_THRESHOLD;
    });
  });

  // Load video URLs wrapper
  async function loadVideoURLsWrapper() {
    await loadVideoURLs(
      highlights,
      videoURLs,
      (progress) => {
        loadingProgress = progress;
      },
      (loaded) => {
        allVideosLoaded = loaded;
      }
    );
  }

  // Update time and highlight wrapper
  function updateTimeAndHighlightWrapper() {
    const result = updateTimeAndHighlight(
      movie,
      highlights,
      (time) => {
        currentTime = time;
      },
      (playing) => {
        isPlaying = playing;
      },
      (index) => {
        currentHighlightIndex = index;
      }
    );

    if (result?.ended) {
      progressTracker.stopProgressTracking();
    }
  }

  // Playback controls wrapper
  async function playPauseWrapper() {
    await playPauseUtil(
      movie,
      isInitialized,
      (playing) => {
        isPlaying = playing;
      },
      startProgressTrackingWrapper,
      progressTracker.stopProgressTracking
    );
  }

  // Jump to a specific highlight wrapper
  async function jumpToHighlightWrapper(highlightIndex) {
    await jumpToHighlightUtil(
      movie,
      highlightIndex,
      highlights,
      updateTimeAndHighlightWrapper,
      isPlaying,
      startProgressTrackingWrapper,
      (buffering) => { isBuffering = buffering; }
    );
  }

  // Progress tracking wrapper
  function startProgressTrackingWrapper() {
    progressTracker.startProgressTracking(
      movie,
      highlights,
      {
        setCurrentTime: (time) => {
          currentTime = time;
        },
        setIsPlaying: (playing) => {
          isPlaying = playing;
        },
        setCurrentHighlightIndex: (index) => {
          currentHighlightIndex = index;
        },
      },
      () => {
        isPlaying = false;
      }
    );
  }

  // Timeline seeking wrapper
  async function handleTimelineSeekWrapper(targetTime) {
    await handleTimelineSeekUtil(
      movie,
      targetTime,
      totalDuration,
      isInitialized,
      updateTimeAndHighlightWrapper,
      isPlaying,
      startProgressTrackingWrapper,
      (buffering) => { isBuffering = buffering; }
    );
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

  // Progress percentage wrapper
  function getProgressPercentageWrapper() {
    return getProgressPercentage(currentTime, totalDuration);
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
      // Reinitialize the video player with new order (only if save was successful)
      await reinitializeWithNewOrderWrapper(newHighlights);
    }

    // Reset the internal reorder flag
    isInternalReorder = false;

    handleDragEnd();
  }

  // Reinitialize video player with new segment order wrapper
  async function reinitializeWithNewOrderWrapper(newHighlights = highlights) {
    console.log(
      "Reinitializing video player with new order:",
      newHighlights.map((h) => h.id)
    );

    // Pause and clean up existing movie
    if (movie) {
      movie.pause();
      progressTracker.stopProgressTracking();
      movie = null;
    }

    // Reset state
    isInitialized = false;
    currentTime = 0;
    currentHighlightIndex = 0;
    initializationError = null;

    // Recreate the movie with the new order
    if (allVideosLoaded && newHighlights.length > 0 && canvasElement) {
      const success = await createEtroMovieWithOrderWrapper(newHighlights);
      if (success) {
        console.log("Video player successfully reinitialized with new order");
      } else {
        console.error("Failed to reinitialize video player with new order");
      }
    }
  }

  // Create Etro movie wrapper
  async function createEtroMovieWithOrderWrapper(highlightOrder) {
    const result = await createEtroMovieWithOrder(
      highlightOrder,
      canvasElement,
      videoURLs,
      allVideosLoaded
    );

    if (result.success) {
      movie = result.movie;
      totalDuration = result.totalDuration;
      isInitialized = true;
      return true;
    } else {
      initializationError = result.error || "Unknown error";
      return false;
    }
  }

  // Watch for when videos are loaded to initialize
  $effect(() => {
    if (
      browser &&
      allVideosLoaded &&
      highlights.length > 0 &&
      !isInitialized &&
      canvasElement
    ) {
      console.log(
        "Effect: Creating Etro movie with",
        highlights.length,
        "highlights"
      );
      createEtroMovieWithOrderWrapper(highlights);
    }
  });

  // Watch for highlights changes and reinitialize if needed
  $effect(() => {
    if (browser && highlights.length > 0) {
      console.log("Effect: Highlights changed, checking initialization state");
      console.log(
        "Current state - allVideosLoaded:",
        allVideosLoaded,
        "isInitialized:",
        isInitialized,
        "videoURLs size:",
        videoURLs.size
      );

      // Check if we need to load video URLs for new highlights
      const needsVideoURLs = highlights.some((h) => !videoURLs.has(h.filePath));

      if (!allVideosLoaded || needsVideoURLs) {
        console.log(
          "Effect: Starting video URL loading for",
          highlights.length,
          "highlights",
          needsVideoURLs ? "(missing URLs detected)" : ""
        );
        // Reset the loaded state to force reload
        allVideosLoaded = false;
        loadVideoURLsWrapper();
      }
    }
  });

  // Watch for external highlight order changes (from timeline component)
  $effect(() => {
    if (
      browser &&
      highlights.length > 0 &&
      isInitialized &&
      allVideosLoaded &&
      !isInternalReorder
    ) {
      const currentOrder = highlights.map((h) => h.id).join(",");

      // If we have a previous order and it's different, refresh the video
      if (lastKnownOrder && lastKnownOrder !== currentOrder) {
        console.log(
          "External highlight order change detected, refreshing video"
        );
        console.log("Previous order:", lastKnownOrder);
        console.log("New order:", currentOrder);

        // Update our known order and refresh video
        lastKnownOrder = currentOrder;
        reinitializeWithNewOrderWrapper(highlights);
      } else if (!lastKnownOrder) {
        // First time initialization - just record the order
        lastKnownOrder = currentOrder;
      }
    }
  });

  // Force update of isPlaying state when movie state changes
  $effect(() => {
    if (browser && movie) {
      const interval = setInterval(() => {
        if (movie) {
          const shouldBePlaying = !movie.paused && !movie.ended;
          if (isPlaying !== shouldBePlaying) {
            isPlaying = shouldBePlaying;
          }
        }
      }, 100);

      return () => clearInterval(interval);
    }
  });

  // Initialize component
  onMount(async () => {
    console.log("EtroVideoPlayer mounted with highlights:", highlights);

    // Wait for canvas element to be ready
    const waitForCanvas = () => {
      return new Promise((resolve) => {
        const checkCanvas = () => {
          if (canvasElement) {
            console.log("Canvas element is ready");
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
      console.log("First highlight:", highlights[0]);
      console.log(
        "Highlight file paths:",
        highlights.map((h) => h.filePath)
      );
    }
  });

  // Cleanup
  onDestroy(() => {
    progressTracker.stopProgressTracking();

    if (movie) {
      movie.pause();
    }
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

    <!-- Canvas Element for Etro rendering -->
    <div
      class="relative w-full aspect-video bg-black overflow-hidden mb-4 rounded"
    >
      <canvas
        bind:this={canvasElement}
        class="w-full h-full bg-black"
        style="object-fit: contain; max-width: 100%; max-height: 100%;"
      ></canvas>

      <!-- Loading indicator -->
      {#if !allVideosLoaded}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black text-white"
        >
          <div class="text-center">
            <div
              class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"
            ></div>
            <p>Loading video URLs... {Math.round(loadingProgress)}%</p>
          </div>
        </div>
      {:else if !isInitialized}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black text-white"
        >
          <div class="text-center">
            <div
              class="animate-spin w-8 h-8 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"
            ></div>
            <p>Initializing Etro video player...</p>
            {#if initializationError}
              <p class="text-red-400 text-sm mt-2">
                Error: {initializationError}
              </p>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Buffering indicator -->
      {#if isBuffering}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black/50 text-white"
        >
          <div class="text-center">
            <div
              class="animate-spin w-6 h-6 border-2 border-white border-t-transparent rounded-full mx-auto mb-2"
            ></div>
            <p class="text-sm">Buffering...</p>
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
              {Math.round(getProgressPercentageWrapper())}%
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
        <div class="flex gap-0.5 w-full">
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
        {#if shouldShowActiveSegment() && highlights[currentHighlightIndex]}
          {@const activeHighlight = highlights[currentHighlightIndex]}
          {@const segmentStartTime = highlights
            .slice(0, currentHighlightIndex)
            .reduce((sum, h) => sum + (h.end - h.start), 0)}
          {@const segmentDuration = activeHighlight.end - activeHighlight.start}
          {@const segmentProgress = Math.max(
            0,
            Math.min(1, (currentTime - segmentStartTime) / segmentDuration)
          )}

          <div class="mt-1">
            <div class="w-full">
              <TimelineSegment
                highlight={activeHighlight}
                index={currentHighlightIndex}
                isActive={true}
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
                  const targetTime =
                    segmentStartTime + clickPercentage * segmentDuration;
                  handleTimelineSeekWrapper(targetTime);
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
          disabled={!allVideosLoaded || !isInitialized}
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
