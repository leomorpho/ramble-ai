// Seek buffer offset in seconds to jump back before target position for better buffering
export const SEEK_BUFFER_OFFSET = 0.0; // 100ms

// Format time for display
export function formatTime(seconds) {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

// Update time and highlight index from Etro movie
export function updateTimeAndHighlight(movie, highlights, setCurrentTime, setIsPlaying, setCurrentHighlightIndex) {
  if (!movie) return;

  const currentTime = movie.currentTime;
  setCurrentTime(currentTime);

  // Force reactivity by reassigning
  setIsPlaying(!movie.paused && !movie.ended);

  // Determine current highlight based on timeline using highlights from store
  let highlightIndex = 0;
  let accumulatedTime = 0;

  for (let i = 0; i < highlights.length; i++) {
    const segmentDuration = highlights[i].end - highlights[i].start;
    if (currentTime < accumulatedTime + segmentDuration) {
      highlightIndex = i;
      break;
    }
    accumulatedTime += segmentDuration;
    highlightIndex = i + 1; // In case we're past all segments
  }

  setCurrentHighlightIndex(Math.min(highlightIndex, highlights.length - 1));

  return { 
    currentTime, 
    isPlaying: !movie.paused && !movie.ended,
    currentHighlightIndex: Math.min(highlightIndex, highlights.length - 1),
    ended: movie.ended 
  };
}

// Calculate progress percentage for timeline
export function getProgressPercentage(currentTime, totalDuration) {
  return totalDuration > 0 ? (currentTime / totalDuration) * 100 : 0;
}

// Calculate target time for timeline seeking within a segment
export function calculateSeekTime(event, segmentIndex, highlights) {
  // Calculate the click position within the segment as a percentage
  const rect = event.currentTarget.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const clickPercentage = x / rect.width;

  // Calculate the start time for this segment
  let segmentStartTime = 0;
  for (let i = 0; i < segmentIndex; i++) {
    segmentStartTime += highlights[i].end - highlights[i].start;
  }

  // Calculate the duration of the clicked segment
  const segmentDuration = highlights[segmentIndex].end - highlights[segmentIndex].start;

  // Calculate the target time within the segment
  const clickTargetTime = segmentStartTime + clickPercentage * segmentDuration;
  
  // Apply buffer offset - jump back by the offset amount for better buffering
  const bufferedTargetTime = Math.max(0, clickTargetTime - SEEK_BUFFER_OFFSET);

  console.log(
    `Segment click: index=${segmentIndex}, clickPos=${clickPercentage.toFixed(2)}, clickTarget=${clickTargetTime.toFixed(2)}s, bufferedTarget=${bufferedTargetTime.toFixed(2)}s (offset: ${SEEK_BUFFER_OFFSET}s)`
  );

  return bufferedTargetTime;
}

// Calculate time at start of a specific highlight
export function calculateHighlightStartTime(highlightIndex, highlights) {
  let targetTime = 0;
  for (let i = 0; i < highlightIndex; i++) {
    targetTime += highlights[i].end - highlights[i].start;
  }
  return targetTime;
}

// Check if click was on drag handle
export function isDragHandleClick(event) {
  const rect = event.currentTarget.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const y = event.clientY - rect.top;

  // Define drag handle area (upper right corner, 16x16 pixels)
  const dragHandleSize = 16;
  return x >= rect.width - dragHandleSize && y <= dragHandleSize;
}