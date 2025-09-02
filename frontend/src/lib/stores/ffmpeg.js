import { goto } from "$app/navigation";
import { IsFFmpegReady } from "$lib/wailsjs/go/main/App";

/**
 * Check if FFmpeg is ready and redirect to installation page if not
 * @param {string} currentPath - Current page path to avoid redirect loops
 * @returns {Promise<boolean>} - true if FFmpeg is ready, false if redirected
 */
export async function checkFFmpegOrRedirect(currentPath = "") {
  // Don't check if we're already on the install page or checking from it
  if (currentPath.includes("/install-ffmpeg")) {
    return true;
  }

  try {
    const isReady = await IsFFmpegReady();
    if (!isReady) {
      console.log("FFmpeg not ready, redirecting to installation page");
      goto("/install-ffmpeg");
      return false;
    }
    return true;
  } catch (error) {
    console.error("Error checking FFmpeg status:", error);
    // On error, assume FFmpeg is not available and redirect
    goto("/install-ffmpeg");
    return false;
  }
}

/**
 * Check FFmpeg without redirecting - useful for conditional UI
 * @returns {Promise<boolean>}
 */
export async function checkFFmpegStatus() {
  try {
    return await IsFFmpegReady();
  } catch (error) {
    console.error("Error checking FFmpeg status:", error);
    return false;
  }
}

/**
 * Check FFmpeg before video operations and redirect if needed
 * @param {string} operation - Description of the operation for logging
 * @returns {Promise<boolean>} - true if can proceed, false if redirected
 */
export async function checkFFmpegForVideoOperation(operation = "video processing") {
  const isReady = await checkFFmpegStatus();
  if (!isReady) {
    console.log(`FFmpeg not ready for ${operation}, redirecting to installation`);
    goto("/install-ffmpeg");
    return false;
  }
  return true;
}