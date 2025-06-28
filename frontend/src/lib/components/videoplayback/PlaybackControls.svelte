<script>
  import { Play, Pause, SkipForward, SkipBack, Square } from '@lucide/svelte';
  import { Button } from "$lib/components/ui/button";
  
  let { 
    isPlaying = false,
    isPaused = false,
    allVideosLoaded = false,
    currentHighlightIndex = 0,
    highlights = [],
    onStartPlayback = () => {},
    onTogglePlayback = () => {},
    onPreviousHighlight = () => {},
    onNextHighlight = () => {},
    onStopPlayback = () => {}
  } = $props();
</script>

<div class="playback-controls flex items-center justify-center gap-2">
  {#if !isPlaying}
    <Button 
      onclick={onStartPlayback} 
      disabled={!allVideosLoaded}
      class="flex items-center gap-2"
    >
      <Play class="w-4 h-4" />
      Play Sequence
    </Button>
  {:else}
    <Button 
      variant="outline" 
      onclick={onPreviousHighlight}
      disabled={currentHighlightIndex === 0}
      title="Previous highlight"
    >
      <SkipBack class="w-4 h-4" />
    </Button>
    
    <Button onclick={onTogglePlayback} class="flex items-center gap-2">
      {#if !isPaused}
        <Pause class="w-4 h-4" />
        Pause
      {:else}
        <Play class="w-4 h-4" />
        Play
      {/if}
    </Button>
    
    <Button 
      variant="outline" 
      onclick={onNextHighlight}
      disabled={currentHighlightIndex >= highlights.length - 1}
      title="Next highlight"
    >
      <SkipForward class="w-4 h-4" />
    </Button>
    
    <Button variant="outline" onclick={onStopPlayback} title="Stop sequence">
      <Square class="w-4 h-4" />
    </Button>
  {/if}
</div>