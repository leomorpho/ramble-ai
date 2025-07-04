import { updateTimeAndHighlight } from "./timelineUtils.js";

// Progress tracking utilities
export function createProgressTracker() {
  let animationFrame = null;

  function startProgressTracking(movie, highlights, updateCallbacks, stopCallback) {
    stopProgressTracking();
    console.log("Starting progress tracking");

    function updateProgress() {
      if (!movie) {
        console.log("No movie in updateProgress");
        return;
      }

      const result = updateTimeAndHighlight(
        movie, 
        highlights, 
        updateCallbacks.setCurrentTime,
        updateCallbacks.setIsPlaying,
        updateCallbacks.setCurrentHighlightIndex
      );

      // Check if playback has ended
      if (result.ended) {
        stopCallback();
        console.log("Playback ended");
        return;
      }

      // Continue tracking if movie is actually playing (not paused)
      if (!movie.paused && !movie.ended) {
        animationFrame = requestAnimationFrame(updateProgress);
      } else {
        console.log(
          "Stopping progress tracking - paused:",
          movie.paused,
          "ended:",
          movie.ended
        );
      }
    }

    animationFrame = requestAnimationFrame(updateProgress);
  }

  function stopProgressTracking() {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame);
      animationFrame = null;
    }
  }

  return { startProgressTracking, stopProgressTracking };
}