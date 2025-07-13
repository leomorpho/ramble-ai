/**
 * Migration utilities for converting old highlight colors to new integer IDs
 */

import { MigrateHighlightColors } from '$lib/wailsjs/go/main/App';
import { toast } from 'svelte-sonner';

/**
 * Runs the highlight color migration
 * @returns {Promise<boolean>} Success status
 */
export async function runHighlightColorMigration() {
  try {
    toast.info('Starting highlight color migration...', {
      description: 'Converting existing highlights to use integer color IDs',
    });

    await MigrateHighlightColors();

    toast.success('Migration completed successfully!', {
      description: 'All highlights now use the new color system',
    });

    return true;
  } catch (error) {
    console.error('Migration failed:', error);
    toast.error('Migration failed', {
      description: error.message || 'An error occurred during migration',
    });
    return false;
  }
}

/**
 * Converts old CSS variable colors to integer IDs (for frontend use)
 * @param {string} oldColor - CSS variable like 'var(--highlight-1)'
 * @returns {number} - Integer ID (1-15)
 */
export function convertCssVariableToId(oldColor) {
  if (typeof oldColor !== 'string') return 1;
  
  // Extract number from CSS variable
  const match = oldColor.match(/var\(--highlight-(\d+)\)/);
  if (match) {
    const id = parseInt(match[1], 10);
    return (id >= 1 && id <= 15) ? id : 1;
  }
  
  // Default fallback
  return 1;
}

/**
 * Maps legacy color strings to new integer IDs
 * @param {string} oldColor - Old color string
 * @returns {number} - New integer ID
 */
export function mapLegacyColorToId(oldColor) {
  if (!oldColor || typeof oldColor !== 'string') return 1;
  
  const color = oldColor.toLowerCase();
  
  // Map common color patterns
  const colorMap = {
    // CSS variables
    'var(--highlight-1)': 1,
    'var(--highlight-2)': 2,
    'var(--highlight-3)': 3,
    'var(--highlight-4)': 4,
    'var(--highlight-5)': 5,
    'var(--highlight-6)': 6,
    'var(--highlight-7)': 7,
    'var(--highlight-8)': 8,
    'var(--highlight-9)': 9,
    'var(--highlight-10)': 10,
    'var(--highlight-11)': 11,
    'var(--highlight-12)': 12,
    'var(--highlight-13)': 13,
    'var(--highlight-14)': 14,
    'var(--highlight-15)': 15,
    
    // Common color names
    'yellow': 1,
    'orange': 2, 
    'red': 3,
    'pink': 4,
    'purple': 5,
    'blue': 7,
    'cyan': 9,
    'teal': 10,
    'green': 11,
    'lime': 13,
    'amber': 14,
    'brown': 15,
  };
  
  // Check for exact matches first
  if (colorMap[color]) {
    return colorMap[color];
  }
  
  // Check for CSS variables using regex
  const cssVarMatch = color.match(/var\(--highlight-(\d+)\)/);
  if (cssVarMatch) {
    const id = parseInt(cssVarMatch[1], 10);
    return (id >= 1 && id <= 15) ? id : 1;
  }
  
  // Default fallback
  return 1;
}