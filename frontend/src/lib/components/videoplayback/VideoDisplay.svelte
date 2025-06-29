<script>
  import { Film } from '@lucide/svelte';
  
  let { 
    displayVideoElement = $bindable(null),
    allVideosLoaded = false,
    isPlaying = false,
    highlights = [],
    loadingProgress = 0
  } = $props();
</script>

<div class="video-display bg-background border rounded-lg overflow-hidden mb-4">
  {#if !allVideosLoaded}
    <div class="p-12 text-center text-muted-foreground">
      <div class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50 animate-pulse">
        <Film />
      </div>
      <p class="text-lg font-medium">Preparing videos...</p>
      <p class="text-sm">Loading {highlights.length} video segments for seamless playback</p>
      <div class="w-full bg-secondary rounded-full h-2 mt-4 max-w-xs mx-auto">
        <div 
          class="bg-primary h-full rounded-full transition-all duration-300"
          style="width: {loadingProgress}%"
        ></div>
      </div>
    </div>
  {:else}
    <!-- Always show video element when videos are loaded -->
    <div class="relative">
      <video 
        bind:this={displayVideoElement}
        class="w-full aspect-video bg-black"
        controls={false}
        muted={false}
      >
        <track kind="captions" />
      </video>
      
      <!-- Show ready message overlay when not playing -->
      {#if !isPlaying}
        <div class="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div class="text-center text-muted-foreground">
            <Film class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
            <p class="text-lg font-medium">Ready to play sequence</p>
            <p class="text-sm">All videos loaded and ready for seamless playback</p>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .video-display {
    min-height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>