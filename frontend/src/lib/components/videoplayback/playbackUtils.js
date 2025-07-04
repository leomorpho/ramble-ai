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
  startProgressTracking,
  setIsBuffering
) {
  if (!movie || highlightIndex < 0 || highlightIndex >= highlights.length)
    return;

  // Calculate time at start of target highlight using highlights from store
  const targetTime = calculateHighlightStartTime(highlightIndex, highlights);

  console.log(
    `Jumping to highlight ${highlightIndex} at time ${targetTime}s`
  );

  const wasPlaying = isPlaying && !movie.paused;
  
  // Set buffering state if we were playing
  if (wasPlaying) {
    setIsBuffering(true);
    console.log("Starting buffering for jump to highlight", highlightIndex);
  }

  // Seek to the target time
  movie.currentTime = targetTime;

  // Update time and highlight index immediately
  updateTimeAndHighlight();

  try {
    // Use refresh() to load the frame at the new position
    await movie.refresh();
    console.log("Frame refreshed for highlight", highlightIndex);

    // Continue playing if we were already playing
    if (wasPlaying && movie.paused) {
      console.log("Resuming playback after jump buffering");
      
      // Use play() with onStart callback to ensure we wait for readiness
      await movie.play({
        onStart: () => {
          console.log("Playback started after highlight jump, clearing buffering state");
          setIsBuffering(false);
          startProgressTracking();
        }
      });
    } else if (wasPlaying && !movie.paused) {
      // Already playing, just ensure progress tracking is active
      setIsBuffering(false);
      startProgressTracking();
    } else {
      // If we weren't playing, just clear buffering state
      setIsBuffering(false);
    }
  } catch (err) {
    console.error("Error during highlight jump buffering:", err);
    setIsBuffering(false);
    
    // Fallback to old behavior if buffering fails
    if (wasPlaying && movie.paused) {
      try {
        await movie.play();
        startProgressTracking();
      } catch (fallbackErr) {
        if (!fallbackErr.message.includes("Already playing")) {
          console.error("Error in fallback playback for highlight jump:", fallbackErr);
        }
      }
    } else if (wasPlaying && !movie.paused) {
      startProgressTracking();
    }
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
  startProgressTracking,
  setIsBuffering
) {
  if (!movie || !isInitialized) return;

  const wasPlaying = isPlaying && !movie.paused;
  
  // Set buffering state if we were playing
  if (wasPlaying) {
    setIsBuffering(true);
    console.log("Starting buffering for seek to", targetTime);
  }

  // Seek to the target time
  movie.currentTime = Math.max(0, Math.min(targetTime, totalDuration));
  updateTimeAndHighlight();

  try {
    // Use refresh() to load the frame at the new position
    await movie.refresh();
    console.log("Frame refreshed at", targetTime);

    // Resume playing if we were playing before seeking
    if (wasPlaying && movie.paused) {
      console.log("Resuming playback after buffering");
      
      // Use play() with onStart callback to ensure we wait for readiness
      await movie.play({
        onStart: () => {
          console.log("Playback actually started, clearing buffering state");
          setIsBuffering(false);
          startProgressTracking();
        }
      });
    } else {
      // If we weren't playing, just clear buffering state
      setIsBuffering(false);
    }
  } catch (err) {
    console.error("Error during seek buffering:", err);
    setIsBuffering(false);
    
    // Fallback to old behavior if buffering fails
    if (wasPlaying && movie.paused) {
      try {
        await movie.play();
        startProgressTracking();
      } catch (fallbackErr) {
        if (!fallbackErr.message.includes("Already playing")) {
          console.error("Error in fallback playback:", fallbackErr);
        }
      }
    }
  }
}