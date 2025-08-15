import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock global window functions that may be used in components
global.window = global.window || {};
global.document = global.document || {};

// Mock ResizeObserver if needed by components
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock IntersectionObserver if needed
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock Wails JS bindings
vi.mock('$lib/wailsjs/go/main/App', () => ({
  CancelExport: vi.fn().mockResolvedValue(undefined),
  ClearAISilenceImprovements: vi.fn().mockResolvedValue(undefined),
  ClearChatHistory: vi.fn().mockResolvedValue(undefined),
  ClearSuggestedHighlights: vi.fn().mockResolvedValue(undefined),
  Close: vi.fn().mockResolvedValue(undefined),
  CreateProject: vi.fn().mockResolvedValue(undefined),
  CreateVideoClip: vi.fn().mockResolvedValue(undefined),
  DeleteHighlight: vi.fn().mockResolvedValue(undefined),
  DeleteOpenAIApiKey: vi.fn().mockResolvedValue(undefined),
  DeleteOpenRouterApiKey: vi.fn().mockResolvedValue(undefined),
  DeleteProject: vi.fn().mockResolvedValue(undefined),
  DeleteSetting: vi.fn().mockResolvedValue(undefined),
  DeleteSuggestedHighlight: vi.fn().mockResolvedValue(undefined),
  DeleteVideoClip: vi.fn().mockResolvedValue(undefined),
  ExportIndividualHighlights: vi.fn().mockResolvedValue(undefined),
  ExportStitchedHighlights: vi.fn().mockResolvedValue(undefined),
  GetChatHistory: vi.fn().mockResolvedValue([]),
  GetExportProgress: vi.fn().mockResolvedValue({}),
  GetHighlightsHistoryStatus: vi.fn().mockResolvedValue({ canUndo: false, canRedo: false }),
  GetOpenAIApiKey: vi.fn().mockResolvedValue(''),
  GetOpenRouterApiKey: vi.fn().mockResolvedValue(''),
  GetOrderHistoryStatus: vi.fn().mockResolvedValue({ canUndo: false, canRedo: false }),
  GetProjectAISettings: vi.fn().mockResolvedValue({}),
  GetProjectAISilenceResult: vi.fn().mockResolvedValue({}),
  GetProjectAISuggestion: vi.fn().mockResolvedValue(null),
  GetProjectByID: vi.fn().mockResolvedValue({}),
  GetProjectExportJobs: vi.fn().mockResolvedValue([]),
  GetProjectHighlightAISettings: vi.fn().mockResolvedValue({}),
  GetProjectHighlightOrder: vi.fn().mockResolvedValue([]),
  GetProjectHighlightOrderWithTitles: vi.fn().mockResolvedValue([]),
  GetProjectHighlights: vi.fn().mockResolvedValue([]),
  GetProjects: vi.fn().mockResolvedValue([]),
  GetSectionTitles: vi.fn().mockResolvedValue([]),
  GetSetting: vi.fn().mockResolvedValue(''),
  GetSuggestedHighlights: vi.fn().mockResolvedValue([]),
  GetThemePreference: vi.fn().mockResolvedValue('system'),
  GetVideoClipsByProject: vi.fn().mockResolvedValue([]),
  GetVideoFileInfo: vi.fn().mockResolvedValue({}),
  GetVideoURL: vi.fn().mockResolvedValue(''),
  Greet: vi.fn().mockResolvedValue('Hello'),
  ImproveHighlightSilencesWithAI: vi.fn().mockResolvedValue(undefined),
  RecoverActiveExportJobs: vi.fn().mockResolvedValue(undefined),
  RedoHighlightsChange: vi.fn().mockResolvedValue(undefined),
  RedoOrderChange: vi.fn().mockResolvedValue([]),
  ReorderHighlightsWithAI: vi.fn().mockResolvedValue(undefined),
  SaveChatModelSelection: vi.fn().mockResolvedValue(undefined),
  SaveOpenAIApiKey: vi.fn().mockResolvedValue(undefined),
  SaveOpenRouterApiKey: vi.fn().mockResolvedValue(undefined),
  SaveProjectAISettings: vi.fn().mockResolvedValue(undefined),
  SaveProjectHighlightAISettings: vi.fn().mockResolvedValue(undefined),
  SaveSectionTitle: vi.fn().mockResolvedValue(undefined),
  SaveSetting: vi.fn().mockResolvedValue(undefined),
  SaveThemePreference: vi.fn().mockResolvedValue(undefined),
  SelectExportFolder: vi.fn().mockResolvedValue(''),
  SelectVideoFiles: vi.fn().mockResolvedValue([]),
  SendChatMessage: vi.fn().mockResolvedValue(''),
  SuggestHighlightsWithAI: vi.fn().mockResolvedValue(undefined),
  TranscribeVideoClip: vi.fn().mockResolvedValue(undefined),
  UndoHighlightsChange: vi.fn().mockResolvedValue(undefined),
  UndoOrderChange: vi.fn().mockResolvedValue([]),
  UpdateProject: vi.fn().mockResolvedValue(undefined),
  UpdateProjectActiveTab: vi.fn().mockResolvedValue(undefined),
  UpdateProjectHighlightOrder: vi.fn().mockResolvedValue(undefined),
  UpdateProjectHighlightOrderWithTitles: vi.fn().mockResolvedValue(undefined),
  UpdateVideoClip: vi.fn().mockResolvedValue(undefined),
  UpdateVideoClipHighlights: vi.fn().mockResolvedValue(undefined),
  UpdateVideoClipSuggestedHighlights: vi.fn().mockResolvedValue(undefined)
}));

// Mock svelte-sonner
vi.mock('svelte-sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn()
  }
}));

// Mock UI components
vi.mock('$lib/components/ui/button', () => ({
  Button: {
    $$render: () => '<button>Mocked Button</button>'
  }
}));

vi.mock('$lib/components/ui/TimeGap.svelte', () => ({
  default: {
    $$render: () => '<span> </span>'
  }
}));

// Mock TextHighlighter utils
vi.mock('./TextHighlighter.utils.js', () => ({
  findWordByTimestamp: vi.fn(),
  addHighlight: vi.fn((highlights, start, end, usedColors) => ({
    highlights: [...highlights, { id: 'new', start, end, color: '#ffeb3b' }],
    newHighlight: { id: 'new', start, end, color: '#ffeb3b' }
  })),
  removeHighlight: vi.fn((highlights, id) => highlights.filter(h => h.id !== id)),
  updateHighlight: vi.fn((highlights, id, start, end) => 
    highlights.map(h => h.id === id ? { ...h, start, end } : h)
  ),
  findHighlightForWord: vi.fn(),
  checkOverlap: vi.fn(),
  isWordInSelection: vi.fn(),
  calculateTimestamps: vi.fn()
}));