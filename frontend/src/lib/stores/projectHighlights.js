import { writable, derived, get } from 'svelte/store';
import { GetProjectHighlights, GetProjectHighlightOrder, UpdateProjectHighlightOrder, DeleteHighlight, UpdateVideoClipHighlights, UndoOrderChange, RedoOrderChange, GetOrderHistoryStatus, UndoHighlightsChange, RedoHighlightsChange, GetHighlightsHistoryStatus } from '$lib/wailsjs/go/main/App';
import { toast } from 'svelte-sonner';

// Store for the raw highlights data from the database
export const rawHighlights = writable([]);

// Store for the custom highlight order
export const highlightOrder = writable([]);

// Store for the current project ID
export const currentProjectId = writable(null);

// Store for loading states
export const highlightsLoading = writable(false);

// Store for history status (undo/redo availability)
export const orderHistoryStatus = writable({ canUndo: false, canRedo: false });
export const highlightsHistoryStatus = writable(new Map()); // Map of clipId -> { canUndo, canRedo }

// Derived store that combines highlights with their custom order
export const orderedHighlights = derived(
  [rawHighlights, highlightOrder],
  ([$rawHighlights, $highlightOrder]) => {
    if ($rawHighlights.length === 0) return [];
    
    if ($highlightOrder.length === 0) {
      // No custom order, sort by video clip ID then by start time
      return [...$rawHighlights].sort((a, b) => {
        if (a.videoClipId !== b.videoClipId) {
          return a.videoClipId - b.videoClipId;
        }
        return a.start - b.start;
      });
    }
    
    // Apply custom ordering
    const orderedList = [];
    const highlightMap = new Map($rawHighlights.map(h => [h.id, h]));
    
    // Add highlights in custom order
    for (const id of $highlightOrder) {
      const highlight = highlightMap.get(id);
      if (highlight) {
        orderedList.push(highlight);
        highlightMap.delete(id);
      }
    }
    
    // Add any remaining highlights that weren't in the custom order
    const remaining = Array.from(highlightMap.values()).sort((a, b) => {
      if (a.videoClipId !== b.videoClipId) {
        return a.videoClipId - b.videoClipId;
      }
      return a.start - b.start;
    });
    
    return [...orderedList, ...remaining];
  }
);

// Function to load highlights and order for a project
export async function loadProjectHighlights(projectId) {
  if (!projectId) {
    console.warn('No project ID provided to loadProjectHighlights');
    return;
  }
  
  highlightsLoading.set(true);
  currentProjectId.set(projectId);
  
  try {
    // Load both highlights and order in parallel
    const [highlightsData, order] = await Promise.all([
      GetProjectHighlights(projectId),
      GetProjectHighlightOrder(projectId)
    ]);
    
    console.log('Loaded highlights data:', highlightsData?.length || 0, 'videos');
    console.log('Loaded highlight order:', order?.length || 0);
    
    // Flatten highlights from all videos into individual highlight objects
    const flattenedHighlights = [];
    if (highlightsData && highlightsData.length > 0) {
      for (const videoHighlights of highlightsData) {
        for (const highlight of videoHighlights.highlights) {
          flattenedHighlights.push({
            ...highlight,
            videoClipId: videoHighlights.videoClipId,
            videoClipName: videoHighlights.videoClipName,
            filePath: videoHighlights.filePath,
            videoDuration: videoHighlights.duration
          });
        }
      }
    }
    
    console.log('Flattened highlights:', flattenedHighlights.length, 'individual highlights');
    
    // Update stores with flattened data
    rawHighlights.set(flattenedHighlights);
    highlightOrder.set(order || []);
    
    // Initialize history status
    await updateOrderHistoryStatus();
    
  } catch (error) {
    console.error('Failed to load project highlights:', error);
    toast.error('Failed to load project highlights');
    rawHighlights.set([]);
    highlightOrder.set([]);
  } finally {
    highlightsLoading.set(false);
  }
}

// Function to update the highlight order and save to database
export async function updateHighlightOrder(newOrder) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for updating highlight order');
    return false;
  }
  
  try {
    // Extract highlight IDs if we received highlight objects instead of IDs
    const highlightIds = newOrder.map(item => 
      typeof item === 'string' ? item : item.id
    );
    
    // Update local state first (optimistic update)
    highlightOrder.set(highlightIds);
    
    // Save to database
    await UpdateProjectHighlightOrder(projectId, highlightIds);
    
    // Update history status after successful order change
    await updateOrderHistoryStatus();
    
    console.log('Updated highlight order in database:', highlightIds);
    toast.success('Highlight order updated successfully');
    
    return true;
  } catch (error) {
    console.error('Failed to update highlight order:', error);
    toast.error('Failed to save highlight order');
    
    // Reload from database to revert optimistic update
    const order = await GetProjectHighlightOrder(projectId).catch(() => []);
    highlightOrder.set(order || []);
    
    return false;
  }
}

// Function to refresh highlights from database
export async function refreshHighlights() {
  const projectId = get(currentProjectId);
  if (projectId) {
    await loadProjectHighlights(projectId);
  }
}

// Function to clear the store (useful when navigating away from project)
export function clearHighlights() {
  rawHighlights.set([]);
  highlightOrder.set([]);
  currentProjectId.set(null);
  highlightsLoading.set(false);
  orderHistoryStatus.set({ canUndo: false, canRedo: false });
  highlightsHistoryStatus.set(new Map());
}

// Function to edit a highlight
export async function editHighlight(highlightId, videoClipId, updates) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for editing highlight');
    return false;
  }
  
  try {
    // Update backend first
    await UpdateVideoClipHighlights(videoClipId, [updates]);
    
    // Update local store by refreshing from database
    // This ensures both timeline and video player react to changes
    await loadProjectHighlights(projectId);
    
    toast.success('Highlight updated successfully');
    return true;
  } catch (error) {
    console.error('Failed to edit highlight:', error);
    toast.error('Failed to update highlight');
    return false;
  }
}

// Function to delete a highlight
export async function deleteHighlight(highlightId, videoClipId) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for deleting highlight');
    return false;
  }
  
  try {
    // Call backend to delete the highlight
    await DeleteHighlight(videoClipId, highlightId);
    
    // Remove from highlight order if it exists
    const currentOrder = get(highlightOrder);
    const updatedOrder = currentOrder.filter(id => id !== highlightId);
    if (updatedOrder.length !== currentOrder.length) {
      highlightOrder.set(updatedOrder);
      // Save updated order to database
      await UpdateProjectHighlightOrder(projectId, updatedOrder);
    }
    
    // Refresh highlights from database to get updated state
    await loadProjectHighlights(projectId);
    
    // Update history status for this clip
    await updateHighlightsHistoryStatus(videoClipId);
    
    toast.success('Highlight deleted successfully');
    return true;
  } catch (error) {
    console.error('Failed to delete highlight:', error);
    toast.error('Failed to delete highlight');
    return false;
  }
}

// Function to add a highlight to a video clip
export async function addHighlight(videoClipId, highlight) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for adding highlight');
    return false;
  }
  
  try {
    // Get current highlights for the video clip
    const currentHighlights = get(rawHighlights);
    const videoHighlights = currentHighlights.filter(h => h.videoClipId === videoClipId);
    
    // Prepare the new highlight (ensure it has an ID)
    const newHighlight = {
      id: highlight.id || `highlight_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      ...highlight
    };
    
    // Prepare highlights array for the backend (without extra fields)
    const backendHighlights = [...videoHighlights, newHighlight].map(h => ({
      id: h.id,
      start: h.start,
      end: h.end,
      color: h.color,
      text: h.text
    }));
    
    // Update backend
    await UpdateVideoClipHighlights(videoClipId, backendHighlights);
    
    // Refresh from database to ensure consistency
    await loadProjectHighlights(projectId);
    
    // Update history status for this clip
    await updateHighlightsHistoryStatus(videoClipId);
    
    toast.success('Highlight added successfully');
    return true;
  } catch (error) {
    console.error('Failed to add highlight:', error);
    toast.error('Failed to add highlight');
    return false;
  }
}

// Function to update all highlights for a video clip
export async function updateVideoHighlights(videoClipId, highlights) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for updating video highlights');
    return false;
  }
  
  try {
    // Prepare highlights for backend (remove extra fields)
    const backendHighlights = highlights.map(h => ({
      id: h.id,
      start: h.start,
      end: h.end,
      color: h.color,
      text: h.text
    }));
    
    // Update backend
    await UpdateVideoClipHighlights(videoClipId, backendHighlights);
    
    // Refresh from database to ensure consistency
    await loadProjectHighlights(projectId);
    
    // Update history status for this clip
    await updateHighlightsHistoryStatus(videoClipId);
    
    toast.success('Highlights updated successfully');
    return true;
  } catch (error) {
    console.error('Failed to update video highlights:', error);
    toast.error('Failed to update highlights');
    return false;
  }
}

// History Management Functions

// Function to update order history status
export async function updateOrderHistoryStatus() {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    orderHistoryStatus.set({ canUndo: false, canRedo: false });
    return;
  }
  
  try {
    const status = await GetOrderHistoryStatus(projectId);
    orderHistoryStatus.set(status);
  } catch (error) {
    console.error('Failed to get order history status:', error);
    orderHistoryStatus.set({ canUndo: false, canRedo: false });
  }
}

// Function to update highlights history status for a video clip
export async function updateHighlightsHistoryStatus(clipId) {
  try {
    const status = await GetHighlightsHistoryStatus(clipId);
    const currentMap = get(highlightsHistoryStatus);
    currentMap.set(clipId, status);
    highlightsHistoryStatus.set(new Map(currentMap));
  } catch (error) {
    console.error('Failed to get highlights history status:', error);
    const currentMap = get(highlightsHistoryStatus);
    currentMap.set(clipId, { canUndo: false, canRedo: false });
    highlightsHistoryStatus.set(new Map(currentMap));
  }
}

// Function to undo order change
export async function undoOrderChange() {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for undo order change');
    return false;
  }
  
  try {
    const newOrder = await UndoOrderChange(projectId);
    
    // Update local store
    highlightOrder.set(newOrder);
    
    // Update history status
    await updateOrderHistoryStatus();
    
    toast.success('Order change undone');
    return true;
  } catch (error) {
    console.error('Failed to undo order change:', error);
    toast.error('Failed to undo order change');
    return false;
  }
}

// Function to redo order change
export async function redoOrderChange() {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for redo order change');
    return false;
  }
  
  try {
    const newOrder = await RedoOrderChange(projectId);
    
    // Update local store
    highlightOrder.set(newOrder);
    
    // Update history status
    await updateOrderHistoryStatus();
    
    toast.success('Order change redone');
    return true;
  } catch (error) {
    console.error('Failed to redo order change:', error);
    toast.error('Failed to redo order change');
    return false;
  }
}

// Function to undo highlights change for a video clip
export async function undoHighlightsChange(clipId) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for undo highlights change');
    return false;
  }
  
  try {
    await UndoHighlightsChange(clipId);
    
    // Refresh highlights from database to reflect changes
    await loadProjectHighlights(projectId);
    
    // Update history status for this clip
    await updateHighlightsHistoryStatus(clipId);
    
    toast.success('Highlights change undone');
    return true;
  } catch (error) {
    console.error('Failed to undo highlights change:', error);
    toast.error('Failed to undo highlights change');
    return false;
  }
}

// Function to redo highlights change for a video clip
export async function redoHighlightsChange(clipId) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for redo highlights change');
    return false;
  }
  
  try {
    await RedoHighlightsChange(clipId);
    
    // Refresh highlights from database to reflect changes
    await loadProjectHighlights(projectId);
    
    // Update history status for this clip
    await updateHighlightsHistoryStatus(clipId);
    
    toast.success('Highlights change redone');
    return true;
  } catch (error) {
    console.error('Failed to redo highlights change:', error);
    toast.error('Failed to redo highlights change');
    return false;
  }
}