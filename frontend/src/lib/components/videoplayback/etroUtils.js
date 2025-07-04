import { browser } from "$app/environment";
import { toast } from "svelte-sonner";
import { getVideoDimensions, calculateScaledDimensions } from "./videoUtils.js";

// Load etro library dynamically (client-side only)
export async function loadEtro(etro) {
  if (!browser || etro) return etro;

  try {
    const etroModule = await import("etro");
    return etroModule;
  } catch (err) {
    console.error("Failed to load etro library:", err);
    toast.error("Failed to load video library");
    return null;
  }
}

// Create Etro movie with custom highlight order
export async function createEtroMovieWithOrder(
  highlightOrder,
  canvasElement,
  videoURLs,
  allVideosLoaded
) {
  if (!canvasElement || !allVideosLoaded || highlightOrder.length === 0) {
    console.error("Cannot create Etro movie: missing requirements");
    return { success: false, movie: null, totalDuration: 0 };
  }

  // Load etro library if not already loaded
  const etroLib = await loadEtro();
  if (!etroLib) {
    console.error("Failed to load etro library");
    return { success: false, movie: null, totalDuration: 0 };
  }

  try {
    console.log(
      "Creating Etro movie with",
      highlightOrder.length,
      "video layers in custom order"
    );

    // Set canvas dimensions
    const canvasWidth = 1280;
    const canvasHeight = 720;
    canvasElement.width = canvasWidth;
    canvasElement.height = canvasHeight;

    // Get video dimensions from the first video
    const firstHighlight = highlightOrder[0];
    console.log("First highlight in order:", firstHighlight);
    console.log(
      "Looking for video URL with filePath:",
      firstHighlight.filePath
    );
    console.log("Available video URLs:", Array.from(videoURLs.keys()));

    const firstVideoURL = videoURLs.get(firstHighlight.filePath);
    if (!firstVideoURL) {
      throw new Error(
        `No video URL for first highlight. FilePath: ${firstHighlight.filePath}`
      );
    }

    console.log("Getting video dimensions from first video...");
    const videoDimensions = await getVideoDimensions(firstVideoURL);
    console.log("Video dimensions:", videoDimensions);

    // Create movie first (Etro determines dimensions from canvas)
    const movie = new etroLib.Movie({
      canvas: canvasElement,
    });

    // Now calculate scaled dimensions using movie dimensions
    const scaledDims = calculateScaledDimensions(
      videoDimensions.width,
      videoDimensions.height,
      movie.width || canvasWidth,
      movie.height || canvasHeight
    );
    console.log("Scaled dimensions:", scaledDims);
    console.log(
      "Movie dimensions after creation:",
      movie.width,
      "x",
      movie.height
    );

    let currentStartTime = 0;

    // Create video layers for each highlight in the specified order
    for (let i = 0; i < highlightOrder.length; i++) {
      const highlight = highlightOrder[i];
      const videoURL = videoURLs.get(highlight.filePath);

      if (!videoURL) {
        console.warn(
          `Skipping highlight ${i}: no video URL for ${highlight.filePath}`
        );
        continue;
      }

      const segmentDuration = highlight.end - highlight.start;

      console.log(
        `Creating layer ${i}: ${highlight.videoClipName} (${segmentDuration}s)`
      );

      // Create video layer with proper destination sizing
      const videoLayer = new etroLib.layer.Video({
        startTime: currentStartTime,
        duration: segmentDuration,
        source: videoURL,
        sourceStartTime: highlight.start,
        x: 0,
        y: 0,
        width: movie.width || canvasWidth,
        height: movie.height || canvasHeight,
        destX: scaledDims.x,
        destY: scaledDims.y,
        destWidth: scaledDims.width,
        destHeight: scaledDims.height,
      });

      movie.layers.push(videoLayer);
      currentStartTime += segmentDuration;
    }

    const totalDuration = currentStartTime;
    console.log(`Etro movie created with total duration: ${totalDuration}s`);

    return { success: true, movie, totalDuration };
  } catch (err) {
    console.error("Failed to create Etro movie with custom order:", err);
    return { success: false, movie: null, totalDuration: 0, error: err.message };
  }
}