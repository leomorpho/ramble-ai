import { toast } from "svelte-sonner";

// Load video URLs using direct file paths (gahara approach)
export async function loadVideoURLs(highlights, videoURLs, setProgress, setAllVideosLoaded) {
  if (highlights.length === 0) {
    console.warn("No highlights provided to load video URLs");
    return;
  }

  console.log(
    "Starting to load video URLs for",
    highlights.length,
    "highlights"
  );
  setProgress(0);
  videoURLs.clear();

  const uniqueVideos = new Map();
  for (const highlight of highlights) {
    if (!uniqueVideos.has(highlight.filePath)) {
      uniqueVideos.set(highlight.filePath, highlight);
    }
  }

  const videoFiles = Array.from(uniqueVideos.values());
  console.log(
    "Setting up direct file URLs for",
    videoFiles.length,
    "unique video files:",
    videoFiles.map((h) => h.filePath)
  );

  let loadedCount = 0;

  for (const highlight of videoFiles) {
    try {
      console.log("Setting up direct URL for:", highlight.filePath);

      // Use direct file path as URL (gahara approach)
      // Encode the file path to handle special characters in external drive names
      const videoURL = encodeURI(highlight.filePath);

      console.log(
        "Using direct file path for",
        highlight.filePath,
        ":",
        videoURL ? "SUCCESS" : "EMPTY"
      );

      if (videoURL) {
        videoURLs.set(highlight.filePath, videoURL);
        loadedCount++;
        const progress = (loadedCount / videoFiles.length) * 100;
        setProgress(progress);
        console.log(
          `Progress: ${loadedCount}/${videoFiles.length} (${Math.round(progress)}%)`
        );
      } else {
        throw new Error("Empty file path");
      }
    } catch (err) {
      console.error("Error setting up video URL for:", highlight.filePath, err);
      toast.error("Failed to load video", {
        description: `Could not load ${highlight.videoClipName}: ${err.message}`,
      });
    }
  }

  console.log(
    "Finished setting up video URLs. Loaded:",
    loadedCount,
    "out of",
    videoFiles.length
  );

  if (loadedCount === videoFiles.length) {
    setAllVideosLoaded(true);
    console.log("All video URLs loaded successfully");
    toast.success("All video URLs loaded!");
  } else if (loadedCount > 0) {
    setAllVideosLoaded(true); // Allow partial loading
    console.log(
      "Partial video URLs loaded:",
      loadedCount,
      "/",
      videoFiles.length
    );
    toast.warning(`Loaded ${loadedCount} out of ${videoFiles.length} videos`);
  } else {
    console.error("No video URLs could be loaded");
    toast.error("Failed to load any video URLs");
  }
}

// Get video dimensions from a test video element
export async function getVideoDimensions(videoURL) {
  return new Promise((resolve, reject) => {
    console.log("Getting video dimensions for URL:", videoURL);
    
    const video = document.createElement("video");
    video.onloadedmetadata = () => {
      console.log("Video metadata loaded successfully. Dimensions:", {
        width: video.videoWidth,
        height: video.videoHeight,
      });
      resolve({
        width: video.videoWidth,
        height: video.videoHeight,
      });
    };
    video.onerror = (error) => {
      console.error("Video failed to load for dimension detection:", error);
      console.error("Failed URL:", videoURL);
      reject(new Error(`Failed to load video for dimension detection: ${videoURL}`));
    };
    video.onabort = () => {
      console.error("Video loading was aborted for:", videoURL);
      reject(new Error(`Video loading aborted: ${videoURL}`));
    };
    video.src = videoURL;
  });
}

// Calculate aspect ratio preserving dimensions
export function calculateScaledDimensions(
  videoWidth,
  videoHeight,
  canvasWidth,
  canvasHeight
) {
  const videoAspect = videoWidth / videoHeight;
  const canvasAspect = canvasWidth / canvasHeight;

  let scaledWidth, scaledHeight, x, y;

  if (videoAspect > canvasAspect) {
    // Video is wider than canvas - fit by width
    scaledWidth = canvasWidth;
    scaledHeight = canvasWidth / videoAspect;
    x = 0;
    y = (canvasHeight - scaledHeight) / 2;
  } else {
    // Video is taller than canvas - fit by height
    scaledHeight = canvasHeight;
    scaledWidth = canvasHeight * videoAspect;
    x = (canvasWidth - scaledWidth) / 2;
    y = 0;
  }

  return { width: scaledWidth, height: scaledHeight, x, y };
}

// Preload video for next highlight to ensure smooth transitions
export async function preloadNextHighlight(
  currentHighlightIndex,
  highlights,
  videoURLs,
  preloadedHighlights,
  setIsPreloading
) {
  // Don't preload if we're at the last highlight
  if (currentHighlightIndex >= highlights.length - 1) {
    return;
  }

  const nextHighlightIndex = currentHighlightIndex + 1;
  const nextHighlight = highlights[nextHighlightIndex];
  
  if (!nextHighlight) {
    return;
  }

  // Check if this highlight is already preloaded
  if (preloadedHighlights.has(nextHighlight.id)) {
    console.log(`Next highlight ${nextHighlight.id} already preloaded`);
    return;
  }

  // Check if we already have the video URL
  if (videoURLs.has(nextHighlight.filePath)) {
    console.log(`Video URL for next highlight ${nextHighlight.id} already available`);
    // Mark as preloaded (caller will update reactive state)
    preloadedHighlights.add(nextHighlight.id);
    return;
  }

  console.log(`Starting preload for next highlight: ${nextHighlight.videoClipName} (${nextHighlight.id})`);
  setIsPreloading(true);

  try {
    // Use direct file path as URL (gahara approach)
    // Encode the file path to handle special characters in external drive names
    const videoURL = encodeURI(nextHighlight.filePath);

    if (videoURL) {
      videoURLs.set(nextHighlight.filePath, videoURL);
      preloadedHighlights.add(nextHighlight.id);
      console.log(`Successfully preloaded next highlight: ${nextHighlight.videoClipName}`);
    } else {
      console.warn(`Empty file path for preload: ${nextHighlight.filePath}`);
    }
  } catch (err) {
    console.warn(`Failed to preload next highlight ${nextHighlight.videoClipName}:`, err.message);
    // Don't show toast for preload failures as they're not critical
  } finally {
    setIsPreloading(false);
  }
}

// Clear preloaded highlights cache (useful when highlights change)
export function clearPreloadCache(setPreloadedHighlights) {
  setPreloadedHighlights(new Set());
  console.log("Preload cache cleared");
}