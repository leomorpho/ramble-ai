/**
 * Banner Service - Custom API Endpoint Access
 * 
 * IMPORTANT: Wails App Architecture
 * 
 * This service is critical for the Wails desktop app architecture.
 * Users of the Wails app:
 * - Do NOT have PocketBase accounts or authentication
 * - Cannot use PocketBase SDK collection methods (pb.collection())
 * - Authenticate ONLY via API keys stored in the app settings
 * 
 * Therefore, this service:
 * - Uses custom API endpoints (/api/banners) instead of PocketBase collections
 * - Sends API keys as Bearer tokens for authenticated requests
 * - The backend API endpoints internally handle PocketBase operations
 * 
 * Never attempt to use pb.collection() methods in the Wails frontend!
 */

import { GetBanners, DismissBanner } from "$lib/wailsjs/go/main/App";

/**
 * Fetch banners with optional authentication and filtering
 * @param {string|null} apiKey - API key for authentication (optional)
 * @param {boolean} includeDismissed - Whether to include dismissed banners (default: false)
 * @returns {Promise<Array>} Array of banner objects
 */
export async function fetchBanners(apiKey = null, includeDismissed = false) {
  try {
    console.log('🔍 Fetching banners via Wails native method');
    console.log('🔑 API Key provided:', !!apiKey);
    console.log('📝 Include dismissed:', includeDismissed);

    // Use native Wails method instead of HTTP request
    const banners = await GetBanners();
    
    console.log('📦 Native banners response:', banners);
    
    // Filter out dismissed banners if requested
    if (!includeDismissed) {
      const filteredBanners = banners.filter(banner => !banner.dismissed);
      console.log('📋 Filtered banners (non-dismissed):', filteredBanners);
      return filteredBanners;
    }
    
    return banners || [];
  } catch (error) {
    console.error('❌ Failed to fetch banners via native method:', error);
    return [];
  }
}

/**
 * Fetch ALL banners (including dismissed ones) with dismissal status
 * This is used for the banner management page
 */
export async function fetchAllBanners(apiKey) {
  return fetchBanners(apiKey, true); // includeDismissed = true
}

/**
 * Dismiss a banner for the current API key
 * This uses the native Wails method instead of HTTP
 */
export async function dismissBanner(bannerId, apiKey) {
  if (!bannerId) {
    throw new Error('Banner ID is required');
  }

  try {
    console.log('🗑️ Dismissing banner via native method:', bannerId);
    
    // Use native Wails method instead of HTTP request
    await DismissBanner(bannerId);
    
    console.log('✅ Banner dismissed successfully:', bannerId);
    return { success: true, message: 'Banner dismissed successfully' };
  } catch (error) {
    console.error('❌ Failed to dismiss banner via native method:', error);
    throw error;
  }
}