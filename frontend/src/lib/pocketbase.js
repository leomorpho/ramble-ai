import PocketBase from "pocketbase";
import { browser } from '$app/environment';

// Determine the PocketBase URL based on environment
function getPocketBaseURL() {
  if (!browser) return undefined;
  
  // In development, use localhost:8090
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:8090';
  }
  
  // In production, use the same origin as the frontend
  return window.location.origin;
}

export const pb = new PocketBase(getPocketBaseURL());