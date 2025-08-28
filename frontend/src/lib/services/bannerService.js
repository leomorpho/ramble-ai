// Banner service for fetching banners from PocketBase backend

const REMOTE_BACKEND_URL = 'http://localhost:8090'; // Default for development

/**
 * Fetch public banners (no authentication required)
 */
export async function fetchPublicBanners() {
  try {
    const response = await fetch(`${REMOTE_BACKEND_URL}/api/banners`);
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return data.banners || [];
  } catch (error) {
    console.warn('Failed to fetch public banners:', error);
    return [];
  }
}

/**
 * Fetch authenticated banners (requires API key)
 */
export async function fetchAuthenticatedBanners(apiKey) {
  if (!apiKey) {
    return [];
  }

  try {
    const response = await fetch(`${REMOTE_BACKEND_URL}/api/banners/authenticated`, {
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return data.banners || [];
  } catch (error) {
    console.warn('Failed to fetch authenticated banners:', error);
    return [];
  }
}

/**
 * Fetch all banners (public + authenticated if API key available)
 */
export async function fetchBanners(apiKey = null) {
  try {
    // Always fetch public banners
    const publicBanners = await fetchPublicBanners();
    
    // If we have an API key, fetch authenticated banners too
    if (apiKey) {
      const authenticatedBanners = await fetchAuthenticatedBanners(apiKey);
      
      // Merge banners, removing duplicates by ID
      const allBanners = [...publicBanners];
      authenticatedBanners.forEach(authBanner => {
        if (!allBanners.find(banner => banner.id === authBanner.id)) {
          allBanners.push(authBanner);
        }
      });
      
      return allBanners;
    }
    
    return publicBanners;
  } catch (error) {
    console.warn('Failed to fetch banners:', error);
    return [];
  }
}