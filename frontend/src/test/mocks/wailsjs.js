import { vi } from 'vitest';

// Mock implementations for all Wails API functions
export const GetProjectHighlights = vi.fn().mockResolvedValue([]);
export const GetProjectHighlightOrder = vi.fn().mockResolvedValue([]);
export const UpdateProjectHighlightOrder = vi.fn().mockResolvedValue(undefined);
export const DeleteHighlight = vi.fn().mockResolvedValue(undefined);
export const UpdateVideoClipHighlights = vi.fn().mockResolvedValue(undefined);
export const UndoOrderChange = vi.fn().mockResolvedValue([]);
export const RedoOrderChange = vi.fn().mockResolvedValue([]);
export const GetOrderHistoryStatus = vi.fn().mockResolvedValue({ canUndo: false, canRedo: false });
export const UndoHighlightsChange = vi.fn().mockResolvedValue(undefined);
export const RedoHighlightsChange = vi.fn().mockResolvedValue(undefined);
export const GetHighlightsHistoryStatus = vi.fn().mockResolvedValue({ canUndo: false, canRedo: false });
export const DeleteSuggestedHighlight = vi.fn().mockResolvedValue(undefined);