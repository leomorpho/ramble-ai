<script>
  import { onMount, onDestroy } from "svelte";
  import { GetVideoURL } from "$lib/wailsjs/go/main/App";
  import { toast } from "svelte-sonner";
  import { Play, Pause, SkipForward, SkipBack, Square } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { updateHighlightOrder, deleteHighlight, editHighlight } from "$lib/stores/projectHighlights.js";
  import { browser } from "$app/environment";
  import HighlightMenu from "$lib/components/HighlightMenu.svelte";
  import ClipEditor from "$lib/components/ClipEditor.svelte";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle 
  } from "$lib/components/ui/dialog";
  import { Film, Trash2 } from '@lucide/svelte';

  let { highlights = [], projectId = null, enableEyeButton = true, onReorder = null } = $props();

  // Core state
  let canvasElement = $state(null);
  let movie = $state(null);
  let etro = $state(null);
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

  // Animation frame for progress updates
  let animationFrame = null;

  // Track highlight order to detect external changes
  let lastKnownOrder = $state('');
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

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }

  // Load etro library dynamically (client-side only)
  async function loadEtro() {
    if (!browser || etro) return etro;
    
    try {
      const etroModule = await import("etro");
      etro = etroModule;
      return etro;
    } catch (err) {
      console.error("Failed to load etro library:", err);
      toast.error("Failed to load video library");
      return null;
    }
  }

  // Load video URLs from backend
  async function loadVideoURLs() {
    if (highlights.length === 0) {
      console.warn("No highlights provided to load video URLs");
      return;
    }

    console.log(
      "Starting to load video URLs for",
      highlights.length,
      "highlights"
    );
    loadingProgress = 0;
    videoURLs.clear();

    const uniqueVideos = new Map();
    for (const highlight of highlights) {
      if (!uniqueVideos.has(highlight.filePath)) {
        uniqueVideos.set(highlight.filePath, highlight);
      }
    }

    const videoFiles = Array.from(uniqueVideos.values());
    console.log(
      "Loading URLs for",
      videoFiles.length,
      "unique video files:",
      videoFiles.map((h) => h.filePath)
    );

    let loadedCount = 0;

    for (const highlight of videoFiles) {
      try {
        console.log("Loading URL for:", highlight.filePath);

        const videoURL = await Promise.race([
          GetVideoURL(highlight.filePath),
          new Promise((_, reject) =>
            setTimeout(
              () => reject(new Error("GetVideoURL timeout after 10 seconds")),
              10000
            )
          ),
        ]);

        console.log(
          "Got URL for",
          highlight.filePath,
          ":",
          videoURL ? "SUCCESS" : "EMPTY"
        );

        if (videoURL) {
          videoURLs.set(highlight.filePath, videoURL);
          loadedCount++;
          loadingProgress = (loadedCount / videoFiles.length) * 100;
          console.log(
            `Progress: ${loadedCount}/${videoFiles.length} (${Math.round(loadingProgress)}%)`
          );
        } else {
          throw new Error("Empty video URL returned");
        }
      } catch (err) {
        console.error("Error loading video URL for:", highlight.filePath, err);
        toast.error("Failed to load video", {
          description: `Could not load ${highlight.videoClipName}: ${err.message}`,
        });
      }
    }

    console.log(
      "Finished loading video URLs. Loaded:",
      loadedCount,
      "out of",
      videoFiles.length
    );

    if (loadedCount === videoFiles.length) {
      allVideosLoaded = true;
      console.log("All video URLs loaded successfully");
      toast.success("All video URLs loaded!");
    } else if (loadedCount > 0) {
      allVideosLoaded = true; // Allow partial loading
      console.log(
        "Partial video URLs loaded:",
        loadedCount,
        "/",
        videoFiles.length
      );
      toast.warning(`Loaded ${loadedCount} out of ${videoFiles.length} videos`);
    } else {
      console.error("No video URLs could be loaded");
      toast.error("Failed to load any video URLs");
    }
  }

  // Get video dimensions from a test video element
  async function getVideoDimensions(videoURL) {
    return new Promise((resolve, reject) => {
      const video = document.createElement("video");
      video.onloadedmetadata = () => {
        resolve({
          width: video.videoWidth,
          height: video.videoHeight,
        });
      };
      video.onerror = () =>
        reject(new Error("Failed to load video for dimension detection"));
      video.src = videoURL;
    });
  }

  // Calculate aspect ratio preserving dimensions
  function calculateScaledDimensions(
    videoWidth,
    videoHeight,
    canvasWidth,
    canvasHeight
  ) {
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
      console.error("Cannot create Etro movie: missing requirements");
      return false;
    }

    // Load etro library if not already loaded
    const etroLib = await loadEtro();
    if (!etroLib) {
      console.error("Failed to load etro library");
      return false;
    }

    try {
      console.log(
        "Creating Etro movie with",
        highlights.length,
        "video layers"
      );

      // Set canvas dimensions
      const canvasWidth = 1280;
      const canvasHeight = 720;
      canvasElement.width = canvasWidth;
      canvasElement.height = canvasHeight;

      // Get video dimensions from the first video
      const firstVideoURL = videoURLs.get(highlights[0].filePath);
      if (!firstVideoURL) {
        throw new Error("No video URL for first highlight");
      }

      console.log("Getting video dimensions from first video...");
      const videoDimensions = await getVideoDimensions(firstVideoURL);
      console.log("Video dimensions:", videoDimensions);

      // Create movie first (Etro determines dimensions from canvas)
      movie = new etroLib.Movie({
        canvas: canvasElement,
      });

      // Now calculate scaled dimensions using movie dimensions
      const scaledDims = calculateScaledDimensions(
        videoDimensions.width,
        videoDimensions.height,
        movie.width || canvasWidth, // Use movie width or fallback to canvas width
        movie.height || canvasHeight // Use movie height or fallback to canvas height
      );
      console.log("Scaled dimensions:", scaledDims);
      console.log(
        "Movie dimensions after creation:",
        movie.width,
        "x",
        movie.height
      );

      let currentStartTime = 0;

      // Create video layers for each highlight
      for (let i = 0; i < highlights.length; i++) {
        const highlight = highlights[i];
        const videoURL = videoURLs.get(highlight.filePath);

        if (!videoURL) {
          console.warn(
            `Skipping highlight ${i}: no video URL for ${highlight.filePath}`
          );
          continue;
        }

        const segmentDuration = highlight.end - highlight.start;

        console.log(
          `Creating layer ${i}: ${highlight.videoClipName} (${segmentDuration}s)`
        );
        console.log(`Layer ${i} settings:`, {
          layerSize: {
            width: movie.width || canvasWidth,
            height: movie.height || canvasHeight,
          },
          destPosition: { x: scaledDims.x, y: scaledDims.y },
          destSize: { width: scaledDims.width, height: scaledDims.height },
        });

        // Create video layer with proper destination sizing
        const videoLayer = new etroLib.layer.Video({
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
          destHeight: scaledDims.height,
        });

        movie.layers.push(videoLayer);
        currentStartTime += segmentDuration;
      }

      totalDuration = currentStartTime;
      console.log(`Etro movie created with total duration: ${totalDuration}s`);
      console.log(
        "Movie details - width:",
        movie.width,
        "height:",
        movie.height,
        "layers:",
        movie.layers.length
      );
      console.log(
        "Movie paused state:",
        movie.paused,
        "ready state:",
        movie.ready
      );

      isInitialized = true;
      return true;
    } catch (err) {
      console.error("Failed to create Etro movie:", err);
      initializationError = err.message;
      return false;
    }
  }

  // Update time and highlight index from Etro movie
  function updateTimeAndHighlight() {
    if (!movie) return;

    currentTime = movie.currentTime;

    // Force reactivity by reassigning
    isPlaying = !movie.paused && !movie.ended;

    // Determine current highlight based on timeline using highlights from store
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
      stopProgressTracking();
      console.log("Playback ended");
    }
  }

  // Playback controls
  async function playPause() {
    if (!movie || !isInitialized) {
      toast.error("Video player not ready");
      return;
    }

    startProgressTracking();

    try {
      if (movie.paused || movie.ended) {
        // Start or resume playback
        if (movie.ended) {
          movie.currentTime = 0; // Reset if ended
        }

        console.log("Starting/resuming playback");
        await movie.play();
        isPlaying = true;
      } else {
        // Pause playback
        console.log("Pausing playback");
        movie.pause();
        isPlaying = false;
        stopProgressTracking();
      }
    } catch (err) {
      console.error("Error toggling playback:", err);
      toast.error("Failed to toggle playback");
      // Sync state with actual movie state
      isPlaying = !movie.paused && !movie.ended;
    }
  }

  // Jump to a specific highlight
  async function jumpToHighlight(highlightIndex) {
    if (!movie || highlightIndex < 0 || highlightIndex >= highlights.length)
      return;

    // Calculate time at start of target highlight using highlights from store
    let targetTime = 0;
    for (let i = 0; i < highlightIndex; i++) {
      targetTime += highlights[i].end - highlights[i].start;
    }

    console.log(
      `Jumping to highlight ${highlightIndex} at time ${targetTime}s`
    );
    movie.currentTime = targetTime;

    // Update time and highlight index immediately
    updateTimeAndHighlight();

    // Continue playing if we were already playing
    if (isPlaying && movie.paused) {
      try {
        await movie.play();
        // Ensure progress tracking continues
        startProgressTracking();
      } catch (err) {
        if (!err.message.includes("Already playing")) {
          console.error("Error resuming playback:", err);
        }
      }
    } else if (isPlaying && !movie.paused) {
      // Already playing, just ensure progress tracking is active
      startProgressTracking();
    }
  }

  // Progress tracking
  function startProgressTracking() {
    stopProgressTracking();
    console.log("Starting progress tracking");

    function updateProgress() {
      if (!movie) {
        console.log("No movie in updateProgress");
        return;
      }

      updateTimeAndHighlight();

      // Continue tracking if movie is actually playing (not paused)
      if (!movie.paused && !movie.ended) {
        animationFrame = requestAnimationFrame(updateProgress);
      } else {
        console.log(
          "Stopping progress tracking - paused:",
          movie.paused,
          "ended:",
          movie.ended
        );
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
    updateTimeAndHighlight();

    // Resume playing if we were playing before seeking
    if (isPlaying && movie.paused) {
      try {
        await movie.play();
        startProgressTracking();
      } catch (err) {
        // Ignore "already playing" errors
        if (!err.message.includes("Already playing")) {
          console.error("Error resuming playback after seek:", err);
        }
      }
    }
  }

  // Handle timeline segment clicks for seeking
  function handleSegmentClick(event, segmentIndex) {
    // Check if the click was on the drag handle (upper right corner)
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    // Define drag handle area (upper right corner, 16x16 pixels)
    const dragHandleSize = 16;
    const isDragHandle =
      x >= rect.width - dragHandleSize && y <= dragHandleSize;

    if (isDragHandle) {
      // This is a drag handle click, don't seek
      return;
    }

    // Calculate the click position within the segment as a percentage
    const clickPercentage = x / rect.width;

    // Calculate the start time for this segment
    let segmentStartTime = 0;
    for (let i = 0; i < segmentIndex; i++) {
      segmentStartTime += highlights[i].end - highlights[i].start;
    }

    // Calculate the duration of the clicked segment
    const segmentDuration =
      highlights[segmentIndex].end - highlights[segmentIndex].start;

    // Calculate the target time within the segment
    const targetTime = segmentStartTime + clickPercentage * segmentDuration;

    console.log(
      `Segment click: index=${segmentIndex}, clickPos=${clickPercentage.toFixed(2)}, targetTime=${targetTime.toFixed(2)}s`
    );

    // Seek to the calculated time
    handleTimelineSeek(targetTime);
  }

  // Progress percentage for timeline
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
      color: updatedHighlight.color
    };
    
    await editHighlight(updatedHighlight.id, updatedHighlight.videoClipId, updates);
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
      const success = await deleteHighlight(highlightToDelete.id, highlightToDelete.videoClipId);
      
      if (success) {
        deleteDialogOpen = false;
        highlightToDelete = null;
      }
    } catch (error) {
      console.error('Error deleting highlight:', error);
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
    // Check if the drag started from the drag handle
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    // Define drag handle area (upper right corner, 16x16 pixels)
    const dragHandleSize = 16;
    const isDragHandle =
      x >= rect.width - dragHandleSize && y <= dragHandleSize;

    if (!isDragHandle) {
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
      lastKnownOrder = newHighlights.map(h => h.id).join(',');
      // Reinitialize the video player with new order (only if save was successful)
      await reinitializeWithNewOrder(newHighlights);
    }
    
    // Reset the internal reorder flag
    isInternalReorder = false;

    handleDragEnd();
  }

  // Reinitialize video player with new segment order
  async function reinitializeWithNewOrder(newHighlights = highlights) {
    console.log("Reinitializing video player with new order:", newHighlights.map(h => h.id));
    
    // Pause and clean up existing movie
    if (movie) {
      movie.pause();
      stopProgressTracking();
      movie = null;
    }

    // Reset state
    isInitialized = false;
    currentTime = 0;
    currentHighlightIndex = 0;
    initializationError = null;

    // Recreate the movie with the new order
    if (allVideosLoaded && newHighlights.length > 0 && canvasElement) {
      const success = await createEtroMovieWithOrder(newHighlights);
      if (success) {
        console.log("Video player successfully reinitialized with new order");
      } else {
        console.error("Failed to reinitialize video player with new order");
      }
    }
  }

  // Create Etro movie with custom highlight order
  async function createEtroMovieWithOrder(highlightOrder) {
    if (!canvasElement || !allVideosLoaded || highlightOrder.length === 0) {
      console.error("Cannot create Etro movie: missing requirements");
      return false;
    }

    // Load etro library if not already loaded
    const etroLib = await loadEtro();
    if (!etroLib) {
      console.error("Failed to load etro library");
      return false;
    }

    try {
      console.log(
        "Creating Etro movie with",
        highlightOrder.length,
        "video layers in custom order"
      );

      // Set canvas dimensions
      const canvasWidth = 1280;
      const canvasHeight = 720;
      canvasElement.width = canvasWidth;
      canvasElement.height = canvasHeight;

      // Get video dimensions from the first video
      const firstHighlight = highlightOrder[0];
      console.log("First highlight in order:", firstHighlight);
      console.log("Looking for video URL with filePath:", firstHighlight.filePath);
      console.log("Available video URLs:", Array.from(videoURLs.keys()));
      
      const firstVideoURL = videoURLs.get(firstHighlight.filePath);
      if (!firstVideoURL) {
        throw new Error(`No video URL for first highlight. FilePath: ${firstHighlight.filePath}`);
      }

      console.log("Getting video dimensions from first video...");
      const videoDimensions = await getVideoDimensions(firstVideoURL);
      console.log("Video dimensions:", videoDimensions);

      // Create movie first (Etro determines dimensions from canvas)
      movie = new etroLib.Movie({
        canvas: canvasElement,
      });

      // Now calculate scaled dimensions using movie dimensions
      const scaledDims = calculateScaledDimensions(
        videoDimensions.width,
        videoDimensions.height,
        movie.width || canvasWidth,
        movie.height || canvasHeight
      );
      console.log("Scaled dimensions:", scaledDims);
      console.log(
        "Movie dimensions after creation:",
        movie.width,
        "x",
        movie.height
      );

      let currentStartTime = 0;

      // Create video layers for each highlight in the specified order
      for (let i = 0; i < highlightOrder.length; i++) {
        const highlight = highlightOrder[i];
        const videoURL = videoURLs.get(highlight.filePath);

        if (!videoURL) {
          console.warn(
            `Skipping highlight ${i}: no video URL for ${highlight.filePath}`
          );
          continue;
        }

        const segmentDuration = highlight.end - highlight.start;

        console.log(
          `Creating layer ${i}: ${highlight.videoClipName} (${segmentDuration}s)`
        );

        // Create video layer with proper destination sizing
        const videoLayer = new etroLib.layer.Video({
          startTime: currentStartTime,
          duration: segmentDuration,
          source: videoURL,
          sourceStartTime: highlight.start,
          x: 0,
          y: 0,
          width: movie.width || canvasWidth,
          height: movie.height || canvasHeight,
          destX: scaledDims.x,
          destY: scaledDims.y,
          destWidth: scaledDims.width,
          destHeight: scaledDims.height,
        });

        movie.layers.push(videoLayer);
        currentStartTime += segmentDuration;
      }

      totalDuration = currentStartTime;
      console.log(`Etro movie created with total duration: ${totalDuration}s`);

      isInitialized = true;
      return true;
    } catch (err) {
      console.error("Failed to create Etro movie with custom order:", err);
      initializationError = err.message;
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
      createEtroMovieWithOrder(highlights);
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
      const needsVideoURLs = highlights.some(h => !videoURLs.has(h.filePath));
      
      if (!allVideosLoaded || needsVideoURLs) {
        console.log(
          "Effect: Starting video URL loading for",
          highlights.length,
          "highlights",
          needsVideoURLs ? "(missing URLs detected)" : ""
        );
        // Reset the loaded state to force reload
        allVideosLoaded = false;
        loadVideoURLs();
      }
    }
  });

  // Watch for external highlight order changes (from timeline component)
  $effect(() => {
    if (browser && highlights.length > 0 && isInitialized && allVideosLoaded && !isInternalReorder) {
      const currentOrder = highlights.map(h => h.id).join(',');
      
      // If we have a previous order and it's different, refresh the video
      if (lastKnownOrder && lastKnownOrder !== currentOrder) {
        console.log("External highlight order change detected, refreshing video");
        console.log("Previous order:", lastKnownOrder);
        console.log("New order:", currentOrder);
        
        // Update our known order and refresh video
        lastKnownOrder = currentOrder;
        reinitializeWithNewOrder(highlights);
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
        <br />
        <span class="text-xs">
          URLs: {videoURLs.size}/{highlights.length} • Ready: {allVideosLoaded
            ? "Yes"
            : "No"} • Init: {isInitialized ? "Yes" : "No"}
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
        <div class="text-xs text-muted-foreground mb-2">
          💡 Click segments to seek, drag handle (⚫) to reorder
        </div>

        <!-- Clip segments with drag and drop -->
        <div class="flex gap-1 w-full">
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
            {#if isDragging && dragOverIndex === index}
              {@render dropIndicator()}
            {/if}

            <button
              class="group relative h-8 rounded transition-all duration-200 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-primary/50 {isActive
                ? 'ring-2 ring-primary'
                : ''} {isDragging && dragStartIndex === index
                ? 'opacity-50 scale-95'
                : ''} cursor-pointer"
              style="width: {segmentWidth}%; background-color: {highlight.color}; min-width: 20px;"
              title="{highlight.videoClipName}: {formatTime(
                highlight.start
              )} - {formatTime(
                highlight.end
              )} (click to seek, drag handle to reorder)"
              draggable="true"
              ondragstart={(e) => handleDragStart(e, index)}
              ondragend={handleDragEnd}
              ondragover={(e) => handleDragOver(e, index)}
              ondrop={(e) => handleDrop(e, index)}
              onclick={(e) => handleSegmentClick(e, index)}
            >
              <!-- Progress indicator for active segment -->
              {#if isActive}
                {@const segmentStartTime = highlights
                  .slice(0, index)
                  .reduce((sum, h) => sum + (h.end - h.start), 0)}
                {@const segmentProgress = Math.max(
                  0,
                  Math.min(
                    1,
                    (currentTime - segmentStartTime) / segmentDuration
                  )
                )}
                <div
                  class="absolute left-0 top-0 h-full bg-white/30 rounded transition-all duration-100"
                  style="width: {segmentProgress * 100}%;"
                ></div>
              {/if}

              <!-- Segment label and eye icon -->
              <div
                class="absolute inset-0 flex items-center justify-center text-xs font-medium text-white drop-shadow pointer-events-none"
              >
                <!-- Number label -->
                <span>{index + 1}</span>
                
                <!-- Eye icon (only show on hover and if enabled) -->
                {#if enableEyeButton}
                  <span class="ml-1 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto">
                    <HighlightMenu 
                      {highlight}
                      onEdit={handleEditHighlight}
                      onDelete={handleDeleteConfirm}
                      popoverOpen={isPopoverOpen(highlight.id)}
                      onPopoverOpenChange={(open) => {
                        if (open) {
                          openPopover(highlight.id);
                        } else {
                          closePopover(highlight.id);
                        }
                      }}
                      iconSize="w-2.5 h-2.5"
                      triggerSize="w-4 h-4"
                    />
                  </span>
                {/if}
              </div>

              <!-- Drag handle -->
              <div
                class="absolute top-0 right-0 w-4 h-4 bg-black/80 rounded-bl rounded-tr opacity-0 group-hover:opacity-100 transition-opacity cursor-move flex items-center justify-center"
                title="Drag to reorder"
              >
                <div class="w-1.5 h-1.5 bg-white rounded-full"></div>
              </div>
            </button>

            <!-- Drop indicator after the last segment -->
            {#if index === highlights.length - 1 && isDragging && dragOverIndex === highlights.length}
              {@render dropIndicator()}
            {/if}
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
      {#key isPlaying}
        <Button
          onclick={playPause}
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

      <!-- Debug info -->
      <div class="text-xs text-muted-foreground ml-4">
        State: {isPlaying ? "Playing" : "Paused"} | Time: {currentTime.toFixed(
          1
        )}s / {totalDuration.toFixed(1)}s
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
        Are you sure you want to delete this highlight? This action cannot be undone.
      </DialogDescription>
    </DialogHeader>
    
    {#if highlightToDelete}
      <div class="space-y-3">
        <div class="flex items-center gap-3 p-3 rounded-lg border" style="background-color: {highlightToDelete.color}20; border-left: 4px solid {highlightToDelete.color};">
          <Film class="w-6 h-6 flex-shrink-0" style="color: {highlightToDelete.color}" />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">{highlightToDelete.videoClipName}</h3>
            <p class="text-sm text-muted-foreground">
              {formatTime(highlightToDelete.start)} - {formatTime(highlightToDelete.end)}
            </p>
            {#if highlightToDelete.text}
              <p class="text-sm mt-1 italic line-clamp-2">"{highlightToDelete.text}"</p>
            {/if}
          </div>
        </div>
      </div>
    {/if}
    
    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={cancelDelete} disabled={deleting}>
        Cancel
      </Button>
      <Button variant="destructive" onclick={handleDeleteHighlight} disabled={deleting}>
        {#if deleting}
          Deleting...
        {:else}
          Delete Highlight
        {/if}
      </Button>
    </div>
  </DialogContent>
</Dialog>
