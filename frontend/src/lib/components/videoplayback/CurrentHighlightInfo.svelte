<script>
  import { formatTime } from "./timelineUtils.js";

  let {
    currentHighlight = null,
    currentHighlightIndex = 0,
    totalHighlights = 0,
    currentTime = 0,
    totalDuration = 0
  } = $props();
</script>

{#if currentHighlight}
  <div class="bg-secondary/30 p-3 rounded-md mb-4">
    <div class="flex items-center justify-between">
      <div>
        <h4 class="font-medium text-sm">
          {currentHighlight.videoClipName}
        </h4>
        <p class="text-xs text-muted-foreground mt-1">
          Segment {currentHighlightIndex + 1} of {totalHighlights}
        </p>
      </div>
      <div class="text-right">
        <div class="text-sm font-mono">
          {formatTime(currentTime)} / {formatTime(totalDuration)}
        </div>
      </div>
    </div>
  </div>
{/if}