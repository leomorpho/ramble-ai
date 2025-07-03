<script>
  import { Button } from "$lib/components/ui/button";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { GetVideoURL } from "$lib/wailsjs/go/app/App";

  let { 
    open = $bindable(false),
    video = $bindable(null)
  } = $props();
  
  let videoURL = $state("");
  let posterURL = $state("");
  let videoElement = $state(null);

  // Watch for video changes and get video URL
  $effect(async () => {
    if (video && open) {
      try {
        const url = await GetVideoURL(video.filePath);
        videoURL = url;
      } catch (err) {
        console.error("Failed to get video URL:", err);
        videoURL = "";
      }
    } else {
      videoURL = "";
      posterURL = "";
    }
  });

  // Generate poster image from first frame
  function generatePoster() {
    if (!videoElement || posterURL) return; // Don't regenerate if we already have a poster
    
    try {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      // Set canvas dimensions to match video
      canvas.width = videoElement.videoWidth;
      canvas.height = videoElement.videoHeight;
      
      // Draw the current frame
      ctx.drawImage(videoElement, 0, 0, canvas.width, canvas.height);
      
      // Convert canvas to data URL for poster
      posterURL = canvas.toDataURL('image/jpeg', 0.8);
      console.log("Generated poster successfully");
    } catch (err) {
      console.error("Failed to generate poster:", err);
      posterURL = "";
    }
  }

  // Try to generate poster when metadata is loaded
  function handleVideoLoadedMetadata() {
    if (!videoElement) return;
    
    // Set current time to 0 and try to seek to get the first frame
    videoElement.currentTime = 0.1; // Seek slightly forward to ensure frame is available
  }

  // Generate poster when data is loaded
  function handleVideoLoadedData() {
    generatePoster();
  }

  // Generate poster after seeking completes
  function handleVideoSeekComplete() {
    if (!posterURL) {
      generatePoster();
    }
  }

  // As a fallback, try to generate poster when the video can play
  function handleVideoCanPlay() {
    if (!posterURL) {
      // Set time to 0 and generate poster
      videoElement.currentTime = 0;
      setTimeout(() => generatePoster(), 100); // Small delay to ensure frame is rendered
    }
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[800px] max-h-[90vh]">
    <DialogHeader>
      <DialogTitle>Video Preview</DialogTitle>
      <DialogDescription>
        {#if video}
          Preview of {video.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    {#if video}
      <div class="space-y-4">
        <!-- Video player -->
        <div class="bg-background border rounded-lg overflow-hidden">
          {#if video.exists && videoURL}
            <video 
              bind:this={videoElement}
              class="w-full h-auto max-h-96" 
              controls 
              preload="metadata"
              src={videoURL}
              poster={posterURL}
              onloadedmetadata={handleVideoLoadedMetadata}
              onloadeddata={handleVideoLoadedData}
              onseeked={handleVideoSeekComplete}
              oncanplay={handleVideoCanPlay}
            >
              <track kind="captions" />
              <p class="p-4 text-center text-muted-foreground">
                Your browser doesn't support video playback or the video format is not supported.
              </p>
            </video>
          {:else if video.exists && !videoURL}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <p class="text-lg font-medium">Loading video...</p>
              <p class="text-sm">Preparing video for playback</p>
            </div>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.864-.833-2.634 0L4.18 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <p class="text-lg font-medium">Video file not found</p>
              <p class="text-sm">The video file may have been moved or deleted</p>
            </div>
          {/if}
        </div>
        
        <!-- Video details -->
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-muted-foreground">Name:</span>
              <span class="font-medium">{video.name}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-muted-foreground">Format:</span>
              <span class="font-mono uppercase">{video.format}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-muted-foreground">Size:</span>
              <span>{formatFileSize(video.fileSize)}</span>
            </div>
          </div>
          <div class="space-y-2">
            {#if video.width && video.height}
              <div class="flex justify-between">
                <span class="text-muted-foreground">Resolution:</span>
                <span>{video.width}×{video.height}</span>
              </div>
            {/if}
            {#if video.duration}
              <div class="flex justify-between">
                <span class="text-muted-foreground">Duration:</span>
                <span>{Math.round(video.duration)}s</span>
              </div>
            {/if}
            <div class="flex justify-between">
              <span class="text-muted-foreground">Status:</span>
              <span class={video.exists ? "text-green-600" : "text-destructive"}>
                {video.exists ? "Available" : "Missing"}
              </span>
            </div>
          </div>
        </div>
        
        <!-- File path -->
        <div class="p-3 bg-secondary/30 rounded-lg">
          <p class="text-xs text-muted-foreground mb-1">File Path:</p>
          <p class="text-sm font-mono break-all">{video.filePath}</p>
        </div>
      </div>
    {/if}
    
    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={() => open = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>