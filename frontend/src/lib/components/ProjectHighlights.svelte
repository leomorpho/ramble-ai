<script>
  import { onDestroy } from "svelte";
  import {
    GetVideoURL,
    GetVideoClipsByProject,
  } from "$lib/wailsjs/go/main/App";
  import { getColorFromId } from "$lib/components/texthighlighter/TextHighlighter.utils.js";
  import { toast } from "svelte-sonner";
  import { Play, Film, Clock, Undo, Redo, Ear, Sparkles } from "@lucide/svelte";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";

  import CompoundVideoPlayer from "$lib/components/videoplayback/CompoundVideoPlayer.svelte";
  import VideoPlayerKeyHandler from "$lib/components/videoplayback/VideoPlayerKeyHandler.svelte";
  import ClipEditor from "$lib/components/ClipEditor.svelte";
  import ReorderableHighlights from "$lib/components/ReorderableHighlights.svelte";
  import AIReorderSheet from "$lib/components/AIReorderSheet.svelte";
  import AIActionsSheet from "$lib/components/AIActionsSheet.svelte";
  import {
    updateHighlightOrder,
    deleteHighlight,
    editHighlight,
    undoOrderChange,
    redoOrderChange,
    orderHistoryStatus,
    updateNewLineTitle,
    updateVideoHighlights,
    refreshHighlights,
    hiddenHighlightsList,
    hideHighlight,
    unhideHighlight,
  } from "$lib/stores/projectHighlights.js";
  import { ImproveHighlightSilencesWithAI } from "$lib/wailsjs/go/main/App";

  let { projectId, highlights, loading } = $props();

  // Local state
  let error = $state("");

  // Multiselect state (managed by ReorderableHighlights component)
  let selectedHighlights = $state(new Set());

  // Video player dialog state
  let videoDialogOpen = $state(false);
  let currentHighlight = $state(null);
  let videoURL = $state("");
  let videoElement = $state(null);
  let videoLoading = $state(false);

  // Clip editor state
  let clipEditorOpen = $state(false);
  let editingHighlight = $state(null);

  // Delete confirmation state
  let deleteDialogOpen = $state(false);
  let highlightToDelete = $state(null);
  let deleting = $state(false);

  // AI silence improvement state
  let aiSilenceLoading = $state(false);
  let showAISilenceConfirmation = $state(false);
  let showAIActionsSheet = $state(false);

  // Video clips with transcription words for internal pause analysis
  let videoClipsWithWords = $state(new Map());

  // Video player play/pause function reference
  let playPauseRef = $state({ current: null });

  // Popover state management
  let popoverStates = $state(new Map());

  // Filter highlights for video player (exclude newline items)
  let videoHighlights = $derived(
    highlights.filter((item) => item.type !== "newline")
  );

  // Load video clips with transcription words for pause analysis
  $effect(() => {
    if (projectId && highlights.length > 0) {
      loadVideoClipsWithWords();
    }
  });

  async function loadVideoClipsWithWords() {
    try {
      const clips = await GetVideoClipsByProject(projectId);
      const clipsMap = new Map();
      clips.forEach((clip) => {
        if (clip.transcriptionWords && clip.transcriptionWords.length > 0) {
          clipsMap.set(clip.id, clip.transcriptionWords);
        }
      });
      videoClipsWithWords = clipsMap;
    } catch (error) {
      console.error("Failed to load video clips with words:", error);
    }
  }

  // Function to get transcription words for a highlight
  function getHighlightWords(highlight) {
    // New lines don't have transcription words
    if (highlight.type === "newline") return [];

    const words = videoClipsWithWords.get(highlight.videoClipId);
    if (!words || words.length === 0) return [];

    // Find words within this highlight's time range
    return words.filter(
      (word) => word.start >= highlight.start && word.end <= highlight.end
    );
  }

  // Cleanup on unmount
  onDestroy(() => {
    // Component-specific cleanup if needed
  });

  // These functions are now handled by the centralized store

  // Format timestamp for display
  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }

  // Format time range
  function formatTimeRange(start, end) {
    return `${formatTimestamp(start)} - ${formatTimestamp(end)}`;
  }

  // Legacy drag handlers removed - using new inline drag system

  // Handle highlight click
  async function handleHighlightClick(highlight) {
    // Don't try to play new lines
    if (highlight.type === "newline") return;

    closePopover(highlight.id);
    currentHighlight = highlight;
    videoLoading = true;
    videoDialogOpen = true;

    try {
      // Get video URL for playback
      const url = await GetVideoURL(highlight.filePath);
      videoURL = url;
    } catch (err) {
      console.error("Failed to get video URL:", err);
      toast.error("Failed to load video", {
        description: "Could not load the video file for playback",
      });
      videoURL = "";
    } finally {
      videoLoading = false;
    }
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
    videoURL = "";
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
      colorId: updatedHighlight.colorId,
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
  
  // Handle hide highlight
  async function handleHideHighlight(event, highlight) {
    if (event) {
      event.stopPropagation();
    }
    closePopover(highlight.id);
    await hideHighlight(highlight.id);
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

  // Handle reordering via the ReorderableHighlights component
  async function handleHighlightReorder(newOrder) {
    await updateHighlightOrder(newOrder);
  }

  async function handleAISilenceImprovement() {
    if (!projectId || highlights.length === 0) {
      toast.error("No highlights to improve");
      return;
    }

    aiSilenceLoading = true;

    try {
      const improvedHighlights =
        await ImproveHighlightSilencesWithAI(projectId);

      if (improvedHighlights && improvedHighlights.length > 0) {
        // Apply the improvements to each video clip's highlights
        for (const videoClip of improvedHighlights) {
          if (videoClip.highlights && videoClip.highlights.length > 0) {
            // The highlights already have the correct format from the backend
            // Update the highlights for this video clip
            await updateVideoHighlights(videoClip.videoClipId, videoClip.highlights);
          }
        }

        // Count total highlights across all video clips
        const totalHighlights = improvedHighlights.reduce((total, videoClip) => {
          return total + (videoClip.highlights ? videoClip.highlights.length : 0);
        }, 0);

        toast.success(
          `Added silence padding to ${totalHighlights} highlight${totalHighlights === 1 ? '' : 's'} across ${improvedHighlights.length} video${improvedHighlights.length === 1 ? '' : 's'}!`
        );
        
        // Refresh highlights to show the changes
        await refreshHighlights();
      } else {
        toast.info("No silence padding improvements were suggested by AI");
      }
    } catch (error) {
      console.error("Failed to improve highlights with AI:", error);
      toast.error("Failed to improve highlights with AI", {
        description: error.message || "An error occurred while processing",
      });
    } finally {
      aiSilenceLoading = false;
    }
  }

  // Undo/Redo handlers
  async function handleUndo() {
    try {
      await undoOrderChange();
    } catch (error) {
      console.error("Failed to undo:", error);
    }
  }

  async function handleRedo() {
    try {
      await redoOrderChange();
    } catch (error) {
      console.error("Failed to redo:", error);
    }
  }

  // Handle title change for newlines
  async function handleTitleChange(index, newTitle) {
    try {
      await updateNewLineTitle(index, newTitle);
    } catch (error) {
      console.error("Failed to update newline title:", error);
      toast.error("Failed to update section title");
    }
  }
</script>

<div class="highlights-timeline space-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-semibold">Highlight Timeline</h2>
    <div class="flex items-center gap-3">
      <!-- Undo/Redo buttons -->
      <div class="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          onclick={handleUndo}
          disabled={!$orderHistoryStatus.canUndo}
          class="flex items-center gap-1 px-2"
          title="Undo order change (Ctrl+Z)"
        >
          <Undo class="w-4 h-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onclick={handleRedo}
          disabled={!$orderHistoryStatus.canRedo}
          class="flex items-center gap-1 px-2"
          title="Redo order change (Ctrl+Y)"
        >
          <Redo class="w-4 h-4" />
        </Button>
      </div>

      {#if highlights.length > 0}
        <Button
          variant="outline"
          size="sm"
          onclick={() => (showAISilenceConfirmation = true)}
          disabled={aiSilenceLoading}
          class="flex items-center gap-2"
          title="Add natural silence padding around highlights for smoother transitions"
        >
          {#if aiSilenceLoading}
            <svg
              class="w-4 h-4 animate-spin"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Adding Padding...
          {:else}
            <Ear class="w-4 h-4" />
            AI Add Silence Padding
          {/if}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onclick={() => (showAIActionsSheet = true)}
          class="flex items-center gap-2"
          title="AI actions for highlight organization and optimization"
        >
          <Sparkles class="w-4 h-4" />
          AI Actions
        </Button>
      {/if}
      <div class="text-sm text-muted-foreground">
        {highlights.length}
        {highlights.length === 1 ? "highlight" : "highlights"}
      </div>
    </div>
  </div>

  {#if loading}
    <div class="text-center py-8 text-muted-foreground">
      <p>Loading highlights...</p>
    </div>
  {:else if error}
    <div
      class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
    >
      <p class="font-medium">Error</p>
      <p class="text-sm">{error}</p>
    </div>
  {:else if highlights.length === 0}
    <div class="text-center py-8 text-muted-foreground">
      <p class="text-lg">No highlights yet</p>
      <p class="text-sm">
        Create highlights in your video transcriptions to see them here
      </p>
    </div>
  {:else}
    <!-- Natural text flow highlight timeline -->
    <div class="highlights-paragraph">
      <ReorderableHighlights
        {highlights}
        bind:selectedHighlights
        onReorder={handleHighlightReorder}
        onSelect={null}
        onEdit={handleEditHighlight}
        onDelete={handleDeleteConfirm}
        onHide={handleHideHighlight}
        onPopoverOpenChange={(highlightId, isOpen) => {
          if (isOpen) {
            openPopover(highlightId);
          } else {
            closePopover(highlightId);
          }
        }}
        {getHighlightWords}
        {isPopoverOpen}
        onHighlightClick={handleHighlightClick}
        onTitleChange={handleTitleChange}
        enableMultiSelect={true}
        enableNewlines={true}
        enableSelection={true}
        enableEdit={true}
        enableDelete={true}
        showAddNewLineButtons={true}
      />
    </div>
    
    <!-- Hidden Highlights Section -->
    {#if $hiddenHighlightsList.length > 0}
      <div class="mt-8 pt-6 border-t border-border">
        <div class="flex items-center gap-2 mb-4">
          <div class="w-2 h-2 bg-muted-foreground rounded-full opacity-50"></div>
          <h3 class="text-sm font-medium text-muted-foreground">Hidden Highlights</h3>
          <div class="text-xs text-muted-foreground">
            {$hiddenHighlightsList.length} hidden
          </div>
        </div>
        <div class="space-y-2">
          {#each $hiddenHighlightsList as highlight}
            <div class="flex items-center gap-2 p-2 bg-muted/30 rounded border border-dashed opacity-60 hover:opacity-100 transition-opacity">
              <div 
                class="w-3 h-3 rounded-sm border border-muted-foreground/30" 
                style="background-color: {getColorFromId(highlight.colorId)}"
              ></div>
              <div class="flex-1 min-w-0">
                <div class="text-sm truncate text-muted-foreground">
                  {highlight.text || 'Highlight'}
                </div>
                <div class="text-xs text-muted-foreground/70">
                  {highlight.videoClipName} • {Math.round(highlight.start)}s-{Math.round(highlight.end)}s
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                class="text-xs h-auto py-1 px-2"
                onclick={() => unhideHighlight(highlight.id)}
              >
                Restore
              </Button>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}

  <!-- Etro Video Player with Keyboard Handler -->
  <VideoPlayerKeyHandler
    onPlayPause={() => {
      if (playPauseRef.current) {
        playPauseRef.current();
      }
    }}
  >
    <CompoundVideoPlayer
      videoHighlights={videoHighlights.filter(
        (h) => h.id && h.id.startsWith("highlight_")
      )}
      {projectId}
      {playPauseRef}
      enableReordering={false}
    />
  </VideoPlayerKeyHandler>
</div>

<!-- AI Add Silence Padding Confirmation Dialog -->
<Dialog bind:open={showAISilenceConfirmation}>
  <DialogContent class="z-[100] sm:max-w-lg">
    <DialogHeader>
      <DialogTitle>Add Silence Padding with AI?</DialogTitle>
      <DialogDescription>
        <div class="space-y-3 pt-2">
          <div class="bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
            <p class="text-sm font-medium text-blue-900 dark:text-blue-100 mb-2">
              Why is this needed?
            </p>
            <p class="text-sm text-blue-800 dark:text-blue-200">
              Transcription-based highlights cut off abruptly without breathing room, creating jarring transitions between clips.
            </p>
          </div>
          
          <div class="bg-secondary/50 rounded-lg p-3">
            <p class="text-sm font-medium mb-2">
              AI will extend {highlights.length} highlight{highlights.length === 1 ? "" : "s"} to include natural silence padding before and after speech.
            </p>
            <p class="text-sm text-muted-foreground">
              You can undo these changes if needed.
            </p>
          </div>
        </div>
      </DialogDescription>
    </DialogHeader>
    <div class="flex justify-end gap-2 pt-4">
      <Button
        variant="outline"
        onclick={() => (showAISilenceConfirmation = false)}
      >
        Cancel
      </Button>
      <Button
        onclick={() => {
          showAISilenceConfirmation = false;
          handleAISilenceImprovement();
        }}
      >
        <Ear class="w-4 h-4 mr-2" />
        Add Silence Padding
      </Button>
    </div>
  </DialogContent>
</Dialog>

<!-- Video Player Dialog -->
<Dialog bind:open={videoDialogOpen}>
  <DialogContent class="sm:max-w-[900px] max-h-[90vh]">
    <DialogHeader>
      <DialogTitle>Highlight Playback</DialogTitle>
      <DialogDescription>
        {#if currentHighlight}
          Playing highlight from {currentHighlight.videoClipName} ({formatTimeRange(
            currentHighlight.start,
            currentHighlight.end
          )})
        {/if}
      </DialogDescription>
    </DialogHeader>

    {#if currentHighlight}
      <div class="space-y-4">
        <!-- Highlight info -->
        <div
          class="flex items-center gap-3 p-3 rounded-lg"
          style="background-color: {getColorFromId(currentHighlight.colorId)}20; border-left: 4px solid {getColorFromId(currentHighlight.colorId)};"
        >
          <Film
            class="w-6 h-6 flex-shrink-0"
            style="color: {getColorFromId(currentHighlight.colorId)}"
          />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">
              {currentHighlight.videoClipName}
            </h3>
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
              <div
                class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-spin"
              >
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
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
                Your browser doesn't support video playback or the video format
                is not supported.
              </p>
            </video>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <svg
                class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.864-.833-2.634 0L4.18 16.5c-.77.833.192 2.5 1.732 2.5z"
                />
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
                <span
                  class="w-2 h-2 rounded-full"
                  style="background-color: {getColorFromId(currentHighlight.colorId)}"
                ></span>
                <span
                  >Highlight: {formatTimeRange(
                    currentHighlight.start,
                    currentHighlight.end
                  )}</span
                >
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={closeVideoDialog}>Close</Button>
    </div>
  </DialogContent>
</Dialog>

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
          style="background-color: {getColorFromId(highlightToDelete.colorId)}20; border-left: 4px solid {getColorFromId(highlightToDelete.colorId)};"
        >
          <Film
            class="w-6 h-6 flex-shrink-0"
            style="color: {getColorFromId(highlightToDelete.colorId)}"
          />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">
              {highlightToDelete.videoClipName}
            </h3>
            <p class="text-sm text-muted-foreground">
              {formatTimeRange(highlightToDelete.start, highlightToDelete.end)}
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

<!-- AI Actions Sheet -->
<AIActionsSheet
  bind:open={showAIActionsSheet}
  {projectId}
  {highlights}
  onApply={() => {
    // Refresh highlights after AI actions
    // The updateHighlightOrder is already called in the component
  }}
/>

<style>
  /* Paragraph layout container */
  .highlights-paragraph {
    line-height: 1.8;
    word-spacing: 2px;
  }
</style>
