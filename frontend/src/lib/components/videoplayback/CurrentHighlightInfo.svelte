<script>
  let { 
    isPlaying = false,
    currentHighlightIndex = 0,
    highlights = []
  } = $props();

  // Format time for display
  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  let currentHighlight = $derived(highlights[currentHighlightIndex]);
</script>

<!-- Current Highlight Info -->
{#if isPlaying && currentHighlight}
  <div class="current-highlight-info flex items-center gap-3 p-3 mb-4 rounded-lg" style="background-color: {currentHighlight.color}20; border-left: 4px solid {currentHighlight.color};">
    <div class="flex-shrink-0">
      <span class="inline-flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground text-sm font-medium">
        {currentHighlightIndex + 1}
      </span>
    </div>
    <div class="flex-1 min-w-0">
      <h4 class="font-medium truncate">{currentHighlight.videoClipName}</h4>
      <p class="text-sm text-muted-foreground">
        {formatTime(currentHighlight.start)} - {formatTime(currentHighlight.end)}
        {#if currentHighlight.text}
          • "{currentHighlight.text}"
        {/if}
      </p>
    </div>
    <div class="text-sm text-muted-foreground">
      {currentHighlightIndex + 1} / {highlights.length}
    </div>
  </div>
{/if}