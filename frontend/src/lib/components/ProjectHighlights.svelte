<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL } from '$lib/wailsjs/go/main/App';
  import { draggable } from '@neodrag/svelte';
  import { toast } from 'svelte-sonner';
  import { Play, Film, X, Edit3, Trash2, Eye } from '@lucide/svelte';
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import ClipEditor from "$lib/components/ClipEditor.svelte";
  import HighlightItem from "$lib/components/HighlightItem.svelte";
  import { 
    orderedHighlights, 
    highlightsLoading, 
    loadProjectHighlights, 
    updateHighlightOrder, 
    clearHighlights,
    deleteHighlight,
    editHighlight
  } from '$lib/stores/projectHighlights.js';

  let { projectId, onHighlightClick = () => {} } = $props();
  
  // Local state
  let error = $state('');
  
  // New multiselect and drag state
  let selectedHighlights = $state(new Set());
  let isDragging = $state(false);
  let draggedHighlights = $state([]);
  let dropPosition = $state(null);
  let dragStartPosition = $state(null);
  let isDropping = $state(false); // Prevent concurrent drops
  
  // Video player dialog state
  let videoDialogOpen = $state(false);
  let currentHighlight = $state(null);
  let videoURL = $state('');
  let videoElement = $state(null);
  let videoLoading = $state(false);

  // Clip editor state
  let clipEditorOpen = $state(false);
  let editingHighlight = $state(null);

  // Delete confirmation state
  let deleteDialogOpen = $state(false);
  let highlightToDelete = $state(null);
  let deleting = $state(false);

  // Popover state management
  let popoverStates = $state(new Map());

  // Initialize on mount and watch for project changes
  onMount(() => {
    if (projectId) {
      loadProjectHighlights(projectId);
    }
  });

  // Watch for project ID changes
  $effect(() => {
    if (projectId) {
      loadProjectHighlights(projectId);
    }
  });

  // Cleanup on unmount
  onDestroy(() => {
    clearHighlights();
  });

  // These functions are now handled by the centralized store

  // Format timestamp for display
  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  // Format time range
  function formatTimeRange(start, end) {
    return `${formatTimestamp(start)} - ${formatTimestamp(end)}`;
  }

  // Legacy drag handlers removed - using new inline drag system

  // Handle highlight click
  async function handleHighlightClick(highlight) {
    closePopover(highlight.id);
    currentHighlight = highlight;
    videoLoading = true;
    videoDialogOpen = true;
    
    try {
      // Get video URL for playback
      const url = await GetVideoURL(highlight.filePath);
      videoURL = url;
    } catch (err) {
      console.error('Failed to get video URL:', err);
      toast.error('Failed to load video', {
        description: 'Could not load the video file for playback'
      });
      videoURL = '';
    } finally {
      videoLoading = false;
    }
    
    // Also call the original callback
    onHighlightClick({
      videoClipId: highlight.videoClipId,
      filePath: highlight.filePath,
      start: highlight.start,
      end: highlight.end
    });
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
    videoURL = '';
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

  // New multiselect and drag handlers
  
  // Handle highlight selection with multiselect support
  function handleHighlightSelect(event, highlight) {
    const isCtrlOrCmd = event.ctrlKey || event.metaKey;
    
    if (isCtrlOrCmd) {
      // Toggle selection for this highlight
      const newSelection = new Set(selectedHighlights);
      if (newSelection.has(highlight.id)) {
        newSelection.delete(highlight.id);
      } else {
        newSelection.add(highlight.id);
      }
      selectedHighlights = newSelection;
    } else {
      // Single select - clear others and select this one, or play if already selected
      if (selectedHighlights.has(highlight.id) && selectedHighlights.size === 1) {
        // If it's the only selected item, play it
        handleHighlightClick(highlight);
      } else {
        // Single select this highlight
        selectedHighlights = new Set([highlight.id]);
      }
    }
  }

  // Handle new drag start with multiselect support
  function handleNewDragStart(event, highlight, index) {
    event.dataTransfer.effectAllowed = 'move';
    
    // If the dragged highlight is not selected, select only it
    if (!selectedHighlights.has(highlight.id)) {
      selectedHighlights = new Set([highlight.id]);
    }
    
    // Set up drag state
    isDragging = true;
    dragStartPosition = index;
    draggedHighlights = Array.from(selectedHighlights);
    
    // Store the highlight IDs in dataTransfer for the drag operation
    event.dataTransfer.setData('text/plain', JSON.stringify(draggedHighlights));
  }

  // Handle container-level drag over
  function handleContainerDragOver(event) {
    event.preventDefault();
    if (isDragging) {
      event.dataTransfer.dropEffect = 'move';
    }
  }

  // Handle container-level drop
  async function handleContainerDrop(event) {
    event.preventDefault();
    
    if (isDragging) {
      // Default to dropping at the end if no position set
      if (dropPosition === null) {
        dropPosition = $orderedHighlights.length;
      }
      console.log('handleContainerDrop: triggering drop', { dropPosition });
      await performDrop();
    }
  }

  // Handle container drag leave
  function handleContainerDragLeave(event) {
    // Only clear if we're leaving the container entirely
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX;
    const y = event.clientY;
    
    if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
      dropPosition = null;
    }
  }

  // Handle drop zone drag over
  function handleDropZoneDragOver(event, position) {
    event.preventDefault();
    event.stopPropagation();
    
    if (isDragging) {
      event.dataTransfer.dropEffect = 'move';
      dropPosition = position;
    }
  }

  // Handle span drag over
  function handleSpanDragOver(event, index) {
    event.preventDefault();
    
    if (isDragging) {
      event.dataTransfer.dropEffect = 'move';
      
      // Calculate drop position based on mouse position within the span
      const rect = event.currentTarget.getBoundingClientRect();
      const mouseX = event.clientX;
      const centerX = rect.left + rect.width / 2;
      
      // If mouse is in the left half, drop before this item, otherwise after
      dropPosition = mouseX < centerX ? index : index + 1;
    }
  }

  // Handle span drop
  async function handleSpanDrop(event, index) {
    event.preventDefault();
    event.stopPropagation();
    
    if (isDragging) {
      // Calculate final drop position based on mouse position
      const rect = event.currentTarget.getBoundingClientRect();
      const mouseX = event.clientX;
      const centerX = rect.left + rect.width / 2;
      
      dropPosition = mouseX < centerX ? index : index + 1;
      console.log('handleSpanDrop: triggering drop', { index, dropPosition });
      await performDrop();
    }
  }

  // Handle drop zone drop
  async function handleDropZoneDrop(event, position) {
    event.preventDefault();
    event.stopPropagation();
    
    if (isDragging) {
      dropPosition = position;
      await performDrop();
    }
  }

  // Perform the actual drop operation
  async function performDrop() {
    if (!isDragging || draggedHighlights.length === 0 || dropPosition === null || isDropping) {
      console.log('performDrop: early return', { 
        isDragging, 
        draggedHighlights: draggedHighlights.length, 
        dropPosition, 
        isDropping 
      });
      return;
    }

    // Prevent concurrent drops
    isDropping = true;

    // Store current state before cleanup
    const draggedIds = [...draggedHighlights];
    const insertPosition = dropPosition;
    
    console.log('performDrop: starting', { draggedIds, insertPosition, totalHighlights: $orderedHighlights.length });

    try {
      const currentHighlights = [...$orderedHighlights]; // Create a copy
      
      // Validate that we have valid data
      if (currentHighlights.length === 0) {
        console.error('performDrop: no highlights to reorder');
        return;
      }
      
      // Create new order using a simpler, more reliable algorithm
      const newOrder = [];
      const draggedItems = [];
      const remainingItems = [];
      
      // Separate dragged items from remaining items, preserving order
      for (const highlight of currentHighlights) {
        if (draggedIds.includes(highlight.id)) {
          draggedItems.push(highlight);
        } else {
          remainingItems.push(highlight);
        }
      }
      
      // Validate we found all dragged items
      if (draggedItems.length !== draggedIds.length) {
        console.error('performDrop: could not find all dragged items', { 
          expected: draggedIds.length, 
          found: draggedItems.length 
        });
        return;
      }
      
      // Insert dragged items at the correct position
      const adjustedInsertPosition = Math.min(insertPosition, remainingItems.length);
      
      // Build the new order
      for (let i = 0; i <= remainingItems.length; i++) {
        if (i === adjustedInsertPosition) {
          newOrder.push(...draggedItems);
        }
        if (i < remainingItems.length) {
          newOrder.push(remainingItems[i]);
        }
      }
      
      // Validate the new order has the correct length
      if (newOrder.length !== currentHighlights.length) {
        console.error('performDrop: new order has wrong length', {
          original: currentHighlights.length,
          newOrder: newOrder.length
        });
        return;
      }
      
      // Check if order actually changed
      const orderChanged = !newOrder.every((item, index) => item.id === currentHighlights[index].id);
      
      if (!orderChanged) {
        console.log('performDrop: order unchanged, skipping update');
        return;
      }
      
      console.log('performDrop: updating order', { 
        oldOrder: currentHighlights.map(h => h.id),
        newOrder: newOrder.map(h => h.id)
      });
      
      // Update via store
      await updateHighlightOrder(newOrder);
      
    } catch (error) {
      console.error('performDrop: error during drop operation:', error);
    } finally {
      // Clean up drag state
      isDropping = false;
      handleNewDragEnd();
    }
  }

  // Handle drag end cleanup
  function handleNewDragEnd() {
    isDragging = false;
    draggedHighlights = [];
    dropPosition = null;
    dragStartPosition = null;
  }

  // Expose refresh method
  export function refresh() {
    loadProjectHighlights(projectId);
  }
</script>

<!-- Drop indicator snippet (adapted from EtroVideoPlayer) -->
{#snippet dropIndicator()}
  <div class="w-0.5 h-8 bg-black dark:bg-white rounded flex-shrink-0"></div>
{/snippet}

<div class="highlights-timeline space-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-xl font-semibold">Highlight Timeline</h2>
    <div class="text-sm text-muted-foreground">
      {$orderedHighlights.length} {$orderedHighlights.length === 1 ? 'highlight' : 'highlights'}
    </div>
  </div>

  {#if $highlightsLoading}
    <div class="text-center py-8 text-muted-foreground">
      <p>Loading highlights...</p>
    </div>
  {:else if error}
    <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4">
      <p class="font-medium">Error</p>
      <p class="text-sm">{error}</p>
    </div>
  {:else if $orderedHighlights.length === 0}
    <div class="text-center py-8 text-muted-foreground">
      <p class="text-lg">No highlights yet</p>
      <p class="text-sm">Create highlights in your video transcriptions to see them here</p>
    </div>
  {:else}
    <!-- Natural text flow highlight timeline -->
    <div class="highlights-paragraph">
      <div 
        class="p-4 bg-muted/30 rounded-lg min-h-[80px] relative leading-relaxed text-base"
        ondragover={(e) => handleContainerDragOver(e)}
        ondrop={(e) => handleContainerDrop(e)}
        ondragleave={handleContainerDragLeave}
      >
        {#if $orderedHighlights.length === 0}
          <div class="text-center py-4 text-muted-foreground">
            <p class="text-sm">No highlights yet. Create highlights in your video transcriptions to see them here.</p>
          </div>
        {:else}
          {#each $orderedHighlights as highlight, index}
            <HighlightItem 
              {highlight}
              {index}
              isSelected={selectedHighlights.has(highlight.id)}
              {isDragging}
              isBeingDragged={isDragging && draggedHighlights.includes(highlight.id) && draggedHighlights[0] === highlight.id}
              showDropIndicatorBefore={isDragging && dropPosition === index}
              onSelect={handleHighlightSelect}
              onDragStart={handleNewDragStart}
              onDragEnd={handleNewDragEnd}
              onDragOver={handleSpanDragOver}
              onDrop={handleSpanDrop}
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
            />
          {/each}
          
          <!-- Drop indicator at the end -->
          {#if isDragging && dropPosition === $orderedHighlights.length}
            <span class="drop-indicator">|</span>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
  
  <!-- Etro Video Player -->
  <EtroVideoPlayer highlights={$orderedHighlights} {projectId} />
</div>

<!-- Video Player Dialog -->
<Dialog bind:open={videoDialogOpen}>
  <DialogContent class="sm:max-w-[900px] max-h-[90vh]">
    <DialogHeader>
      <DialogTitle>Highlight Playback</DialogTitle>
      <DialogDescription>
        {#if currentHighlight}
          Playing highlight from {currentHighlight.videoClipName} ({formatTimeRange(currentHighlight.start, currentHighlight.end)})
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    {#if currentHighlight}
      <div class="space-y-4">
        <!-- Highlight info -->
        <div class="flex items-center gap-3 p-3 rounded-lg" style="background-color: {currentHighlight.color}20; border-left: 4px solid {currentHighlight.color};">
          <Film class="w-6 h-6 flex-shrink-0" style="color: {currentHighlight.color}" />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">{currentHighlight.videoClipName}</h3>
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
              <div class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-spin">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
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
                Your browser doesn't support video playback or the video format is not supported.
              </p>
            </video>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.864-.833-2.634 0L4.18 16.5c-.77.833.192 2.5 1.732 2.5z" />
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
                <span class="w-2 h-2 rounded-full" style="background-color: {currentHighlight.color}"></span>
                <span>Highlight: {formatTimeRange(currentHighlight.start, currentHighlight.end)}</span>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
    
    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={closeVideoDialog}>
        Close
      </Button>
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
              {formatTimeRange(highlightToDelete.start, highlightToDelete.end)}
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

<style>
  
  /* Paragraph layout container */
  .highlights-paragraph {
    line-height: 1.8;
    word-spacing: 2px;
  }
  
  /* Natural text wrapping */
  .highlights-paragraph > div {
    word-break: break-word;
    hyphens: auto;
    text-align: justify;
  }
  
</style>