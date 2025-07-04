import { toast } from "svelte-sonner";
import { calculateHighlightStartTime } from "./timelineUtils.js";

// Playback control utilities
export async function playPause(movie, isInitialized, setIsPlaying, startProgressTracking, stopProgressTracking) {
  if (!movie || !isInitialized) {
    toast.error("Video player not ready");
    return;
  }

  startProgressTracking();

  try {
    if (movie.paused || movie.ended) {
      // Start or resume playback
      if (movie.ended) {
        movie.currentTime = 0; // Reset if ended
      }

      console.log("Starting/resuming playback");
      await movie.play();
      setIsPlaying(true);
    } else {
      // Pause playback
      console.log("Pausing playback");
      movie.pause();
      setIsPlaying(false);
      stopProgressTracking();
    }
  } catch (err) {
    console.error("Error toggling playback:", err);
    toast.error("Failed to toggle playback");
    // Sync state with actual movie state
    setIsPlaying(!movie.paused && !movie.ended);
  }
}

// Jump to a specific highlight
export async function jumpToHighlight(
  movie,
  highlightIndex,
  highlights,
  updateTimeAndHighlight,
  isPlaying,
  startProgressTracking
) {
  if (!movie || highlightIndex < 0 || highlightIndex >= highlights.length)
    return;

  // Calculate time at start of target highlight using highlights from store
  const targetTime = calculateHighlightStartTime(highlightIndex, highlights);

  console.log(
    `Jumping to highlight ${highlightIndex} at time ${targetTime}s`
  );
  movie.currentTime = targetTime;

  // Update time and highlight index immediately
  updateTimeAndHighlight();

  // Continue playing if we were already playing
  if (isPlaying && movie.paused) {
    try {
      await movie.play();
      // Ensure progress tracking continues
      startProgressTracking();
    } catch (err) {
      if (!err.message.includes("Already playing")) {
        console.error("Error resuming playback:", err);
      }
    }
  } else if (isPlaying && !movie.paused) {
    // Already playing, just ensure progress tracking is active
    startProgressTracking();
  }
}

// Timeline seeking
export async function handleTimelineSeek(
  movie,
  targetTime,
  totalDuration,
  isInitialized,
  updateTimeAndHighlight,
  isPlaying,
  startProgressTracking
) {
  if (!movie || !isInitialized) return;

  movie.currentTime = Math.max(0, Math.min(targetTime, totalDuration));
  updateTimeAndHighlight();

  // Resume playing if we were playing before seeking
  if (isPlaying && movie.paused) {
    try {
      await movie.play();
      startProgressTracking();
    } catch (err) {
      // Ignore "already playing" errors
      if (!err.message.includes("Already playing")) {
        console.error("Error resuming playback after seek:", err);
      }
    }
  }
}