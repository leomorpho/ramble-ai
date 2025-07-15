/**
 * Color theme mapping for highlight colors
 * Maps integer IDs (1-20) to CSS custom properties that work in both light and dark modes
 */

export const HIGHLIGHT_COLORS = {
  1: 'var(--highlight-1)',   // Yellow
  2: 'var(--highlight-2)',   // Orange
  3: 'var(--highlight-3)',   // Red
  4: 'var(--highlight-4)',   // Pink
  5: 'var(--highlight-5)',   // Purple
  6: 'var(--highlight-6)',   // Deep Purple
  7: 'var(--highlight-7)',   // Blue
  8: 'var(--highlight-8)',   // Light Blue
  9: 'var(--highlight-9)',   // Cyan
  10: 'var(--highlight-10)', // Teal
  11: 'var(--highlight-11)', // Green
  12: 'var(--highlight-12)', // Light Green
  13: 'var(--highlight-13)', // Lime
  14: 'var(--highlight-14)', // Amber
  15: 'var(--highlight-15)', // Brown
  16: 'var(--highlight-16)', // Rose
  17: 'var(--highlight-17)', // Indigo
  18: 'var(--highlight-18)', // Emerald
  19: 'var(--highlight-19)', // Slate
  20: 'var(--highlight-20)', // Gray
};

/**
 * Get color value from color ID
 * @param {number} colorId - Integer ID (1-20)
 * @returns {string} - CSS custom property
 */
export function getColorFromId(colorId) {
  // Handle invalid colorId (0, null, undefined, out of range)
  if (!colorId || colorId < 1 || colorId > 20) {
    console.warn('🎨 Invalid colorId:', colorId, 'falling back to color 1');
    return HIGHLIGHT_COLORS[1];
  }
  
  return HIGHLIGHT_COLORS[colorId] || HIGHLIGHT_COLORS[1]; // Fallback to first color
}

/**
 * Convert old CSS variable color to integer ID
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
  
  // Handle direct hex colors by mapping to closest theme color
  const hexMatch = oldColor.match(/^#[0-9a-fA-F]{6}$/);
  if (hexMatch) {
    // Find the closest color in our theme
    const targetColor = oldColor.toLowerCase();
    for (const [id, color] of Object.entries(HIGHLIGHT_COLORS)) {
      if (color.toLowerCase() === targetColor) {
        return parseInt(id, 10);
      }
    }
  }
  
  // Default fallback
  return 1;
}

/**
 * Map old string colors to new integer IDs based on common patterns
 * @param {string} oldColor - Old color string
 * @returns {number} - New integer ID
 */
export function mapLegacyColorToId(oldColor) {
  if (!oldColor || typeof oldColor !== 'string') return 1;
  
  const color = oldColor.toLowerCase();
  
  // Map common color patterns
  const colorMap = {
    // Red variations
    'red': 1,
    '#ef4444': 1,
    '#dc2626': 1,
    '#b91c1c': 1,
    
    // Orange variations  
    'orange': 2,
    '#f97316': 2,
    '#ea580c': 2,
    '#c2410c': 2,
    
    // Yellow variations
    'yellow': 3,
    '#eab308': 3,
    '#ca8a04': 3,
    '#a16207': 3,
    
    // Green variations
    'green': 4,
    '#22c55e': 4,
    '#16a34a': 4,
    '#15803d': 4,
    
    // Cyan variations
    'cyan': 5,
    '#06b6d4': 5,
    '#0891b2': 5,
    '#0e7490': 5,
    
    // Blue variations
    'blue': 6,
    '#3b82f6': 6,
    '#2563eb': 6,
    '#1d4ed8': 6,
    
    // Violet variations
    'violet': 7,
    '#8b5cf6': 7,
    '#7c3aed': 7,
    '#6d28d9': 7,
    
    // Pink variations
    'pink': 8,
    '#ec4899': 8,
    '#db2777': 8,
    '#be185d': 8,
  };
  
  // Check for exact matches first
  if (colorMap[color]) {
    return colorMap[color];
  }
  
  // Check for CSS variables
  const cssVarMatch = color.match(/var\(--highlight-(\d+)\)/);
  if (cssVarMatch) {
    const id = parseInt(cssVarMatch[1], 10);
    return (id >= 1 && id <= 15) ? id : 1;
  }
  
  // Check for hex colors
  if (colorMap[color]) {
    return colorMap[color];
  }
  
  // Default fallback
  return 1;
}