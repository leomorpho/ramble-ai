import { GetThemePreference } from "$lib/wailsjs/go/main/App";

/**
 * Set theme on document element
 */
export function setTheme(theme) {
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', theme);
  }
}

/**
 * Get current theme from document element
 */
export function getTheme() {
  if (typeof document !== 'undefined') {
    return document.documentElement.getAttribute('data-theme') || 'light';
  }
  return 'light';
}

/**
 * Initialize theme from database on app startup
 */
export async function initializeTheme() {
  try {
    const savedTheme = await GetThemePreference();
    setTheme(savedTheme);
  } catch (error) {
    console.error("Failed to load theme preference:", error);
    // Fallback to light mode if there's an error
    setTheme("light");
  }
}