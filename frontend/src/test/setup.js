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
  DeleteSuggestedHighlight: vi.fn().mockResolvedValue(undefined),
  GetProjectHighlights: vi.fn().mockResolvedValue([]),
  GetProjectHighlightOrder: vi.fn().mockResolvedValue([]),
  UpdateProjectHighlightOrder: vi.fn().mockResolvedValue(undefined),
  DeleteHighlight: vi.fn().mockResolvedValue(undefined),
  UpdateVideoClipHighlights: vi.fn().mockResolvedValue(undefined),
  UndoOrderChange: vi.fn().mockResolvedValue([]),
  RedoOrderChange: vi.fn().mockResolvedValue([]),
  GetOrderHistoryStatus: vi.fn().mockResolvedValue({ canUndo: false, canRedo: false }),
  UndoHighlightsChange: vi.fn().mockResolvedValue(undefined),
  RedoHighlightsChange: vi.fn().mockResolvedValue(undefined),
  GetHighlightsHistoryStatus: vi.fn().mockResolvedValue({ canUndo: false, canRedo: false })
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