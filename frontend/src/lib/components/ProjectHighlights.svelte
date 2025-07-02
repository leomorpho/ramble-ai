<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetVideoURL, ReorderHighlightsWithAI, GetProjectAISettings, SaveProjectAISettings, GetProjectAISuggestion } from '$lib/wailsjs/go/main/App';
  import { draggable } from '@neodrag/svelte';
  import { toast } from 'svelte-sonner';
  import { Play, Film, X, Edit3, Trash2, Eye, Sparkles } from '@lucide/svelte';
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Label } from "$lib/components/ui/label";
  import { Select } from "$lib/components/ui/select";
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

  // AI reordering state
  let aiReorderDialogOpen = $state(false);
  let aiReorderLoading = $state(false);
  let aiReorderedHighlights = $state([]);
  let aiReorderError = $state('');
  let customPrompt = $state('');
  let selectedModel = $state('anthropic/claude-3-haiku-20240307');
  let hasCachedSuggestion = $state(false);
  let cachedSuggestionDate = $state(null);
  let cachedSuggestionModel = $state('');
  let showOriginalForm = $state(false);
  let originalHighlights = $state([]);
  
  // AI dialog independent state (separate from main page)
  let aiDialogHighlights = $state([]); // Independent copy of highlights for AI dialog
  let aiSelectedHighlights = $state(new Set()); // Selection state for AI dialog
  let aiIsDragging = $state(false);
  let aiDraggedHighlights = $state([]);
  let aiDropPosition = $state(null);
  let aiDragStartPosition = $state(null);
  let aiIsDropping = $state(false);
  
  // AI dialog drag state
  let aiDragStartIndex = $state(-1);
  let aiDragOverIndex = $state(-1);

  // Popover state management
  let popoverStates = $state(new Map());

  // Available AI models
  const availableModels = [
    { value: 'anthropic/claude-sonnet-4', label: 'Claude Sonnet 4 (Latest)' },
    { value: 'google/gemini-2.0-flash-001', label: 'Gemini 2.0 Flash' },
    { value: 'google/gemini-2.5-flash-preview-05-20', label: 'Gemini 2.5 Flash Preview' },
    { value: 'deepseek/deepseek-chat-v3-0324:free', label: 'DeepSeek Chat v3 (Free)' },
    { value: 'anthropic/claude-3.7-sonnet', label: 'Claude 3.7 Sonnet' },
    { value: 'anthropic/claude-3-haiku-20240307', label: 'Claude 3 Haiku (Fast)' },
    { value: 'openai/gpt-4o-mini', label: 'GPT-4o Mini' },
    { value: 'mistralai/mistral-nemo', label: 'Mistral Nemo' },
    { value: 'custom', label: 'Custom Model' }
  ];
  
  let customModelValue = $state('');

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

  // Default YouTube expert prompt
  const defaultPrompt = `You are an expert YouTuber and content creator with millions of subscribers, known for creating highly engaging videos that maximize viewer retention and satisfaction. Your task is to reorder these video highlight segments to create the highest quality video possible.

Reorder these segments using your expertise in:
- Hook creation and audience retention
- Storytelling and narrative structure
- Pacing and rhythm for maximum engagement
- Building emotional connections with viewers
- Creating viral-worthy content flow
- Strategic placement of key moments

Feel free to completely restructure the order - move any segment to any position if it will improve video quality and viewer experience.`;

  // Handle AI reordering - opens dialog
  async function handleAIReorder() {
    if (!projectId || $orderedHighlights.length === 0) {
      toast.error('No highlights to reorder');
      return;
    }

    // Reset state and open dialog
    aiReorderLoading = false;
    aiReorderError = '';
    aiReorderedHighlights = [];
    hasCachedSuggestion = false;
    cachedSuggestionDate = null;
    cachedSuggestionModel = '';
    showOriginalForm = false;
    
    // Initialize AI dialog with independent copy of current highlights
    aiDialogHighlights = [...$orderedHighlights];
    originalHighlights = [...$orderedHighlights]; // Store original order
    aiSelectedHighlights.clear();
    aiIsDragging = false;
    aiDraggedHighlights = [];
    aiDropPosition = null;
    aiDragStartPosition = null;
    aiIsDropping = false;
    
    // Load project AI settings
    try {
      const aiSettings = await GetProjectAISettings(projectId);
      selectedModel = aiSettings.aiModel || 'anthropic/claude-3-haiku-20240307';
      customPrompt = aiSettings.aiPrompt || defaultPrompt;
      
      // If using custom model, extract the value
      if (!availableModels.find(m => m.value === selectedModel)) {
        customModelValue = selectedModel;
        selectedModel = 'custom';
      }
    } catch (error) {
      console.error('Failed to load AI settings:', error);
      selectedModel = 'anthropic/claude-3-haiku-20240307';
      customPrompt = defaultPrompt;
    }
    
    // Try to load cached AI suggestion
    try {
      const cachedSuggestion = await GetProjectAISuggestion(projectId);
      if (cachedSuggestion && cachedSuggestion.order && cachedSuggestion.order.length > 0) {
        // Reorder AI dialog highlights based on cached suggestion
        const reorderedHighlights = [];
        const highlightsMap = new Map();
        
        // Create a map for quick lookup from current highlights
        for (const highlight of aiDialogHighlights) {
          highlightsMap.set(highlight.id, highlight);
        }
        
        // Reorder based on cached suggestion
        for (const id of cachedSuggestion.order) {
          const highlight = highlightsMap.get(id);
          if (highlight) {
            reorderedHighlights.push(highlight);
          }
        }
        
        // Add any highlights that weren't in the cached order
        for (const highlight of aiDialogHighlights) {
          if (!cachedSuggestion.order.includes(highlight.id)) {
            reorderedHighlights.push(highlight);
          }
        }
        
        if (reorderedHighlights.length > 0) {
          aiDialogHighlights = reorderedHighlights;
          hasCachedSuggestion = true;
          cachedSuggestionDate = new Date(cachedSuggestion.createdAt);
          cachedSuggestionModel = cachedSuggestion.model || '';
          
          // Preselect the last used model if available
          if (cachedSuggestion.model) {
            // Check if the cached model is in available models
            if (availableModels.find(m => m.value === cachedSuggestion.model)) {
              selectedModel = cachedSuggestion.model;
            } else {
              // It's a custom model
              customModelValue = cachedSuggestion.model;
              selectedModel = 'custom';
            }
          }
          
          console.log('Loaded cached AI suggestion from', cachedSuggestionDate, 'with model:', cachedSuggestion.model);
        }
      }
    } catch (error) {
      console.log('No cached AI suggestion found:', error);
      // Not an error - just means no cached suggestion exists
      hasCachedSuggestion = false;
      cachedSuggestionDate = null;
      cachedSuggestionModel = '';
    }
    
    aiReorderDialogOpen = true;
  }

  // Start the actual AI reordering process
  async function startAIReordering() {
    aiReorderLoading = true;
    aiReorderError = '';

    try {
      // Save AI settings before processing
      const modelToSave = selectedModel === 'custom' ? customModelValue : selectedModel;
      await SaveProjectAISettings(projectId, {
        aiModel: modelToSave,
        aiPrompt: customPrompt
      });

      // Call the AI reordering API
      const reorderedIds = await ReorderHighlightsWithAI(projectId, customPrompt);
      
      // Reorder the AI dialog highlights based on AI suggestion
      const reorderedHighlights = [];
      const highlightsMap = new Map();
      
      // Create a map for quick lookup from current AI dialog highlights
      for (const highlight of aiDialogHighlights) {
        highlightsMap.set(highlight.id, highlight);
      }
      
      // Build reordered array
      for (const id of reorderedIds) {
        const highlight = highlightsMap.get(id);
        if (highlight) {
          reorderedHighlights.push(highlight);
        }
      }
      
      // Add any highlights that weren't in the AI response
      for (const highlight of aiDialogHighlights) {
        if (!reorderedIds.includes(highlight.id)) {
          reorderedHighlights.push(highlight);
        }
      }
      
      // Update AI dialog highlights with the reordered list
      aiDialogHighlights = reorderedHighlights;
      aiReorderedHighlights = reorderedHighlights; // Keep for backward compatibility with existing logic
      
      // Update cache state - we now have a fresh suggestion
      hasCachedSuggestion = true;
      cachedSuggestionDate = new Date();
      showOriginalForm = true; // Show reset option after AI generation
      
      toast.success('AI reordering completed!');
    } catch (error) {
      console.error('AI reordering error:', error);
      aiReorderError = error.message || 'Failed to reorder highlights with AI';
      toast.error('Failed to reorder highlights with AI');
    } finally {
      aiReorderLoading = false;
    }
  }

  // Apply AI reordering to global state
  async function applyAIReordering() {
    if (aiReorderedHighlights.length === 0) return;

    try {
      // Update via centralized store
      const success = await updateHighlightOrder(aiReorderedHighlights);
      
      if (success) {
        aiReorderDialogOpen = false;
        toast.success('AI reordering applied successfully!');
      }
    } catch (error) {
      console.error('Error applying AI reordering:', error);
      toast.error('Failed to apply AI reordering');
    }
  }

  // Reset to original highlights before AI generation
  function resetToOriginal() {
    aiDialogHighlights = [...originalHighlights];
    aiReorderedHighlights = [];
    showOriginalForm = false;
    hasCachedSuggestion = false;
    cachedSuggestionDate = null;
    cachedSuggestionModel = '';
    toast.success('Reset to original highlight order');
  }

  // Cancel AI reordering
  function cancelAIReordering() {
    aiReorderDialogOpen = false;
    aiReorderedHighlights = [];
    aiReorderError = '';
    customPrompt = '';
  }

  // AI dialog drag handlers
  function handleAIDragStart(event, index) {
    aiDragStartIndex = index;
    aiIsDragging = true;
    aiDraggedHighlights = [aiDialogHighlights[index].id];
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", index.toString());
  }

  function handleAIDragEnd() {
    aiDragStartIndex = -1;
    aiDragOverIndex = -1;
    aiIsDragging = false;
    aiDraggedHighlights = [];
    aiDropPosition = null;
  }

  function handleAIDragOver(event, targetIndex) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    aiDragOverIndex = targetIndex;
    aiDropPosition = targetIndex;
  }

  async function handleAIDrop(event, targetIndex) {
    event.preventDefault();

    if (aiDragStartIndex === -1 || aiDragStartIndex === targetIndex) {
      handleAIDragEnd();
      return;
    }

    // Reorder the AI dialog highlights array
    const newHighlights = [...aiDialogHighlights];
    const draggedItem = newHighlights[aiDragStartIndex];

    // Remove dragged item
    newHighlights.splice(aiDragStartIndex, 1);

    // Insert at new position
    const insertIndex =
      aiDragStartIndex < targetIndex ? targetIndex - 1 : targetIndex;
    newHighlights.splice(insertIndex, 0, draggedItem);

    // Update both AI dialog highlights and reordered highlights for sync
    aiDialogHighlights = newHighlights;
    aiReorderedHighlights = newHighlights;

    handleAIDragEnd();
  }

  // AI dialog container drag functions
  function handleAIContainerDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }

  function handleAIContainerDrop(event) {
    event.preventDefault();
    // Handle drop at end of timeline
    if (aiDragStartIndex !== -1) {
      const targetIndex = aiDialogHighlights.length;
      handleAIDrop(event, targetIndex);
    }
  }

  function handleAIContainerDragLeave(event) {
    // Reset drag indicators when leaving container
    if (!event.currentTarget.contains(event.relatedTarget)) {
      aiDropPosition = null;
    }
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
    <div class="flex items-center gap-3">
      {#if $orderedHighlights.length > 1}
        <Button 
          variant="outline" 
          size="sm"
          onclick={handleAIReorder}
          disabled={aiReorderLoading}
          class="flex items-center gap-2"
        >
          <Sparkles class="w-4 h-4" />
          {aiReorderLoading ? 'AI Reordering...' : 'AI Reorder'}
        </Button>
      {/if}
      <div class="text-sm text-muted-foreground">
        {$orderedHighlights.length} {$orderedHighlights.length === 1 ? 'highlight' : 'highlights'}
      </div>
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
        role="application"
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

<!-- AI Reordering Dialog -->
<Dialog bind:open={aiReorderDialogOpen}>
  <DialogContent class="sm:max-w-[900px] max-h-[90vh] flex flex-col">
    <DialogHeader>
      <DialogTitle class="flex items-center gap-2">
        <Sparkles class="w-5 h-5" />
        AI Reordered Highlights
      </DialogTitle>
      <DialogDescription>
        Let AI suggest a new order for your highlights to maximize video quality and viewer engagement.
      </DialogDescription>
    </DialogHeader>
    
    <div class="flex-1 overflow-y-auto pr-2">
    
    <!-- AI Settings -->
    {#if !aiReorderLoading && aiReorderedHighlights.length === 0 && !aiReorderError}
      <div class="space-y-4">
        <!-- Model Selection -->
        <div class="space-y-2">
          <Label for="ai-model">AI Model</Label>
          <Select
            bind:value={selectedModel}
            options={availableModels}
            placeholder="Select AI model..."
            class="w-full"
          />
          
          {#if selectedModel === 'custom'}
            <input
              type="text"
              bind:value={customModelValue}
              placeholder="Enter custom model (e.g., anthropic/claude-3-5-sonnet)"
              class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            />
          {/if}
          
          <p class="text-xs text-muted-foreground">
            Choose the AI model for highlight reordering. Different models have varying strengths in content analysis and reasoning.
          </p>
        </div>
        
        <!-- Custom Prompt Input -->
        <div class="space-y-2">
          <Label for="custom-prompt">AI Instructions</Label>
          <Textarea
            id="custom-prompt"
            bind:value={customPrompt}
            placeholder="AI instructions for reordering highlights..."
            class="min-h-[120px] resize-none"
            rows="6"
          />
          <p class="text-xs text-muted-foreground">
            Modify the prompt above to customize how AI reorders your highlights. The default focuses on YouTube best practices for maximum engagement.
          </p>
        </div>
        
        {#if hasCachedSuggestion && cachedSuggestionDate}
          <div class="p-3 bg-secondary rounded-lg">
            <p class="text-sm text-muted-foreground">
              <strong>Cached AI Suggestion:</strong> Loaded from {cachedSuggestionDate.toLocaleString()}
              {#if cachedSuggestionModel}
                <br><strong>Model used:</strong> {availableModels.find(m => m.value === cachedSuggestionModel)?.label || cachedSuggestionModel}
              {/if}
            </p>
          </div>
        {/if}
        
        <div class="flex justify-between">
          <Button 
            variant="outline"
            onclick={() => { customPrompt = defaultPrompt; }}
            disabled={customPrompt === defaultPrompt}
          >
            Reset to Default
          </Button>
          <div class="flex gap-2">
            {#if hasCachedSuggestion}
              <Button 
                variant="outline" 
                onclick={startAIReordering} 
                class="flex items-center gap-2"
                disabled={aiReorderLoading}
              >
                <Sparkles class="w-4 h-4" />
                Re-run AI
              </Button>
            {/if}
            <Button 
              onclick={startAIReordering} 
              class="flex items-center gap-2"
              disabled={aiReorderLoading}
            >
              <Sparkles class="w-4 h-4" />
              {hasCachedSuggestion ? 'Update AI Suggestions' : 'Generate AI Suggestions'}
            </Button>
          </div>
        </div>
      </div>
    {/if}
    
    {#if aiReorderLoading}
      <div class="p-8 text-center">
        <div class="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-4"></div>
        <p class="text-lg font-medium">AI is analyzing your highlights...</p>
        <p class="text-sm text-muted-foreground">This may take a few moments</p>
      </div>
    {:else if aiReorderError}
      <div class="p-6 text-center space-y-4">
        <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4">
          <p class="font-medium">Error</p>
          <p class="text-sm">{aiReorderError}</p>
        </div>
        <div class="flex justify-center gap-2">
          <Button variant="outline" onclick={() => { 
            aiReorderError = ''; 
            aiReorderedHighlights = []; 
            aiDialogHighlights = [...$orderedHighlights]; // Reset to original highlights
          }}>
            Modify Prompt & Try Again
          </Button>
          <Button variant="outline" onclick={cancelAIReordering}>
            Cancel
          </Button>
        </div>
      </div>
    {:else if aiDialogHighlights.length > 0}
      <div class="space-y-4">
        <!-- Preview Video Player -->
        <div class="bg-card border rounded-lg p-4">
          <h3 class="text-sm font-medium mb-3 flex items-center gap-2">
            <Play class="w-4 h-4" />
            Preview AI Arrangement
          </h3>
          <EtroVideoPlayer highlights={aiDialogHighlights} {projectId} />
        </div>
        
        <!-- AI Dialog Timeline (same style as main page) -->
        <div class="bg-muted/30 rounded-lg p-4">
          <h3 class="text-sm font-medium mb-3">AI Suggested Order (drag to reorder):</h3>
          
          <!-- Timeline-style highlight display -->
          <div 
            class="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border"
            role="application"
            ondragover={(e) => handleAIContainerDragOver(e)}
            ondrop={(e) => handleAIContainerDrop(e)}
            ondragleave={handleAIContainerDragLeave}
          >
            {#if aiDialogHighlights.length === 0}
              <div class="text-center py-4 text-muted-foreground">
                <p class="text-sm">No highlights to display.</p>
              </div>
            {:else}
              {#each aiDialogHighlights as highlight, index}
                <HighlightItem 
                  {highlight}
                  {index}
                  isSelected={aiSelectedHighlights.has(highlight.id)}
                  isDragging={aiIsDragging}
                  isBeingDragged={aiIsDragging && aiDraggedHighlights.includes(highlight.id) && aiDraggedHighlights[0] === highlight.id}
                  showDropIndicatorBefore={aiIsDragging && aiDropPosition === index}
                  onSelect={() => {}}
                  onDragStart={(e, h, i) => handleAIDragStart(e, i)}
                  onDragEnd={handleAIDragEnd}
                  onDragOver={(e, i) => handleAIDragOver(e, i)}
                  onDrop={(e, i) => handleAIDrop(e, i)}
                  onEdit={() => {}}
                  onDelete={() => {}}
                  popoverOpen={false}
                  onPopoverOpenChange={() => {}}
                />
                {#if index < aiDialogHighlights.length - 1}
                  <span class="mx-1"> </span>
                {/if}
              {/each}
            {/if}
          </div>
        </div>
      </div>
    {/if}
    
    </div>
    
    <div class="flex justify-between gap-2 mt-4 pt-2 border-t">
      <div class="flex gap-2">
        {#if showOriginalForm}
          <Button variant="outline" onclick={resetToOriginal} disabled={aiReorderLoading}>
            Reset to Original
          </Button>
        {/if}
      </div>
      <div class="flex gap-2">
        <Button variant="outline" onclick={cancelAIReordering} disabled={aiReorderLoading}>
          Cancel
        </Button>
        {#if aiReorderedHighlights.length > 0}
          <Button onclick={applyAIReordering} class="flex items-center gap-2">
            <Sparkles class="w-4 h-4" />
            Apply AI Order
          </Button>
        {/if}
      </div>
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