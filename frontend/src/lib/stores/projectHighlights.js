import { writable, derived, get } from 'svelte/store';
import { GetProjectHighlights, GetProjectHighlightOrder, UpdateProjectHighlightOrder, UpdateProjectHighlightOrderWithTitles, DeleteHighlight, UpdateVideoClipHighlights, UndoOrderChange, RedoOrderChange, GetOrderHistoryStatus, UndoHighlightsChange, RedoHighlightsChange, GetHighlightsHistoryStatus, SaveSectionTitle, GetProjectHighlightOrderWithTitles } from '$lib/wailsjs/go/main/App';
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

// Utility functions for newline handling with titles
function isNewline(item) {
  if (!item) return false;
  return item === 'N' || item === 'n' || (typeof item === 'object' && item.type === 'N');
}

function getNewlineTitle(item) {
  if (!item) return '';
  if (typeof item === 'object' && item.type === 'N') {
    return item.title || '';
  }
  return '';
}

function createNewlineFromDb(dbItem) {
  if (!dbItem) return null;
  if (dbItem === 'N') {
    return {
      id: `newline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type: 'newline',
      title: ''
    };
  }
  if (typeof dbItem === 'object' && dbItem.type === 'N') {
    return {
      id: `newline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type: 'newline',
      title: dbItem.title || ''
    };
  }
  return dbItem;
}

// Derived store that combines highlights with their custom order and new line indicators
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
    
    // Apply custom ordering including new line indicators
    const orderedList = [];
    const highlightMap = new Map($rawHighlights.map(h => [h.id, h]));
    
    // Add highlights and new lines in custom order
    for (const item of $highlightOrder) {
      if (!item) continue; // Skip null/undefined items
      if (isNewline(item)) {
        // Add new line indicator with title support
        const newlineItem = createNewlineFromDb(item);
        if (newlineItem) orderedList.push(newlineItem);
      } else {
        const highlight = highlightMap.get(item);
        if (highlight) {
          orderedList.push(highlight);
          highlightMap.delete(item);
        }
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
      GetProjectHighlightOrderWithTitles(projectId)
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
    // Extract highlight IDs and preserve new line indicators with titles
    const highlightIds = newOrder.map(item => {
      if (typeof item === 'string') {
        return item;
      } else if (item.type === 'newline') {
        // Convert display format to database format
        return item.title ? { type: 'N', title: item.title } : 'N';
      } else {
        return item.id;
      }
    });

    // Flatten consecutive 'N' characters to prevent multiple blank lines
    const flattenedIds = flattenConsecutiveNewlines(highlightIds);
    
    // Store original order for potential revert
    const originalOrder = get(highlightOrder);
    
    // Update local state first (optimistic update)
    highlightOrder.set(flattenedIds);
    
    // Save to database - use UpdateProjectHighlightOrderWithTitles to preserve title information
    await UpdateProjectHighlightOrderWithTitles(projectId, flattenedIds);
    
    // Update history status after successful order change
    await updateOrderHistoryStatus();
    
    console.log('Updated highlight order in database:', flattenedIds);
    // toast.success('Highlight order updated successfully');
    
    return true;
  } catch (error) {
    console.error('Failed to update highlight order:', error);
    toast.error('Failed to save highlight order');
    
    // Revert to original order on failure
    try {
      const order = await GetProjectHighlightOrder(projectId);
      highlightOrder.set(order || []);
    } catch (revertError) {
      // If we can't load from database, revert to what we had before
      highlightOrder.set(originalOrder);
    }
    
    return false;
  }
}

// Function to insert a new line indicator at a specific position
export async function insertNewLine(position) {
  const currentOrder = get(highlightOrder);
  const orderedHighlightsList = get(orderedHighlights);
  const allHighlights = get(rawHighlights);
  
  // If we have no order but have highlights, create initial order with all highlights
  if (currentOrder.length === 0 && allHighlights.length > 0) {
    // Sort highlights by videoClipId then by start time to create initial order
    const sortedHighlights = [...allHighlights].sort((a, b) => {
      if (a.videoClipId !== b.videoClipId) {
        return a.videoClipId - b.videoClipId;
      }
      return a.start - b.start;
    });
    
    // Create initial order with all highlight IDs
    const initialOrder = sortedHighlights.map(h => h.id);
    
    // Insert newline at the specified position
    const newOrder = [...initialOrder];
    newOrder.splice(position, 0, 'N');
    
    return await updateHighlightOrder(newOrder);
  }
  
  // Convert the visual position (in orderedHighlights) to the position in highlightOrder
  let actualPosition = 0;
  
  // If position is 0, insert at the beginning
  if (position === 0) {
    actualPosition = 0;
  } else {
    // Find the position in currentOrder where we should insert
    // We need to map the visual position to the database position
    let orderIndex = 0;
    
    for (let i = 0; i < position && i < orderedHighlightsList.length; i++) {
      const item = orderedHighlightsList[i];
      if (item.type !== 'newline') {
        // Find this highlight in the current order
        while (orderIndex < currentOrder.length && isNewline(currentOrder[orderIndex])) {
          orderIndex++;
        }
        orderIndex++; // Move past this highlight
      } else {
        // This is a newline, move past it in the order
        orderIndex++;
      }
    }
    
    actualPosition = orderIndex;
  }
  
  const newOrder = [...currentOrder];
  newOrder.splice(actualPosition, 0, 'N');
  
  return await updateHighlightOrder(newOrder);
}

// Function to remove a new line indicator at a specific position
export async function removeNewLine(position) {
  const currentOrder = get(highlightOrder);
  const orderedHighlightsList = get(orderedHighlights);
  
  // Check if the position is valid and contains a newline
  if (position >= orderedHighlightsList.length || orderedHighlightsList[position].type !== 'newline') {
    return false;
  }
  
  // Find the actual position in highlightOrder corresponding to the visual position
  let actualPosition = -1;
  let orderIndex = 0;
  
  for (let i = 0; i <= position && i < orderedHighlightsList.length; i++) {
    const item = orderedHighlightsList[i];
    if (item.type === 'newline') {
      if (i === position) {
        actualPosition = orderIndex;
        break;
      } else {
        orderIndex++; // Move past this newline
      }
    } else {
      // This is a highlight, find it in the current order
      while (orderIndex < currentOrder.length && currentOrder[orderIndex] === 'N') {
        orderIndex++;
      }
      orderIndex++; // Move past this highlight
    }
  }
  
  if (actualPosition >= 0) {
    const newOrder = [...currentOrder];
    newOrder.splice(actualPosition, 1);
    return await updateHighlightOrder(newOrder);
  }
  
  return false;
}

// Function to update a newline title
export async function updateNewLineTitle(position, newTitle) {
  const projectId = get(currentProjectId);
  
  if (!projectId) {
    console.warn('No project ID available for updating newline title');
    return false;
  }
  
  try {
    // Use the new dedicated SaveSectionTitle backend function
    await SaveSectionTitle(projectId, position, newTitle);
    
    // Refresh highlights from database to reflect the title change
    await loadProjectHighlights(projectId);
    
    return true;
  } catch (error) {
    console.error('Failed to update newline title:', error);
    throw error; // Re-throw so the UI can handle the error
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
    
    // Remove from highlight order if it exists (preserve newlines)
    const currentOrder = get(highlightOrder);
    const updatedOrder = currentOrder.filter(id => id !== highlightId);
    if (updatedOrder.length !== currentOrder.length) {
      highlightOrder.set(updatedOrder);
      // Save updated order to database - use UpdateProjectHighlightOrderWithTitles to preserve title information
      await UpdateProjectHighlightOrderWithTitles(projectId, updatedOrder);
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

// Utility function to flatten consecutive 'N' characters
function flattenConsecutiveNewlines(ids) {
  if (!ids || ids.length <= 1) {
    return ids;
  }

  const result = [];
  let lastWasNewline = false;

  for (const id of ids) {
    if (isNewline(id)) {
      if (!lastWasNewline) {
        result.push(id);
        lastWasNewline = true;
      } else {
        // When flattening consecutive newlines, preserve the title if the current one has one
        const lastNewline = result[result.length - 1];
        const currentTitle = getNewlineTitle(id);
        const lastTitle = getNewlineTitle(lastNewline);
        
        if (currentTitle && !lastTitle) {
          // Replace the last newline with the current one that has a title
          result[result.length - 1] = id;
        }
      }
      // Skip consecutive newlines
    } else {
      result.push(id);
      lastWasNewline = false;
    }
  }

  return result;
}