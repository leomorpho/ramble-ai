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