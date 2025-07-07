<script>
  import { Button } from "$lib/components/ui/button";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import TextHighlighter from "$lib/components/TextHighlighter.svelte";
  import { 
    updateVideoHighlights, 
  } from "$lib/stores/projectHighlights.js";

  let { 
    open = $bindable(false),
    video = $bindable(null),
    projectId,
    highlights = [],
    onHighlightsChange
  } = $props();

  // Get highlights for this video
  let videoHighlights = $derived(
    video && highlights
      ? highlights.filter(h => h.videoClipId === video.id && h.filePath === video.filePath)
          .map(h => ({
            id: h.id,
            start: h.start,
            end: h.end,
            color: h.color,
            text: h.text
          }))
      : []
  );

  async function handleHighlightsChange(newHighlights) {
    if (!video) return;
    
    try {
      // Use the store function to update highlights
      await updateVideoHighlights(video.id, newHighlights);
      
      // Call parent handler if provided
      if (onHighlightsChange) {
        await onHighlightsChange(newHighlights);
      }
    } catch (err) {
      console.error("Failed to save highlights:", err);
    }
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, '0')}`;
  }

  async function copyTranscript() {
    if (video?.transcription) {
      await navigator.clipboard.writeText(video.transcription);
    }
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[800px] max-h-[90vh] flex flex-col">
    <DialogHeader>
      <DialogTitle>Video Transcript</DialogTitle>
      <DialogDescription>
        {#if video}
          {video.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    <ScrollArea class="flex-1">
      {#snippet children()}
        {#if video}
          {#if video.transcription}
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="font-medium">Transcript</h3>
                <div class="flex gap-2 items-center">
                  {#if video.transcriptionLanguage}
                    <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                      {video.transcriptionLanguage.toUpperCase()}
                    </span>
                  {/if}
                  {#if video.transcriptionDuration}
                    <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                      {formatTimestamp(video.transcriptionDuration)}
                    </span>
                  {/if}
                  <Button 
                    variant="outline" 
                    size="sm"
                    onclick={copyTranscript}
                    class="text-xs"
                  >
                    <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Copy
                  </Button>
                </div>
              </div>

              <div class="bg-background border rounded-lg p-4">
                <TextHighlighter 
                  text={video.transcription} 
                  words={video.transcriptionWords || []} 
                  highlights={videoHighlights}
                  suggestedHighlights={[]}
                  videoId={video.id}
                  onHighlightsChange={handleHighlightsChange}
                />
              </div>
              
              <div class="text-xs text-muted-foreground">
                Character count: {video.transcription.length}
                {#if video.transcriptionWords}
                  | Word count: {video.transcriptionWords.length}
                {/if}
              </div>
            </div>
          {:else}
            <div class="flex items-center justify-center h-64 text-muted-foreground">
              <div class="text-center">
                <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="text-lg font-medium">No transcript available</p>
                <p class="text-sm">This video hasn't been transcribed yet.</p>
              </div>
            </div>
          {/if}
        {/if}
      {/snippet}
    </ScrollArea>
    
    <div class="flex justify-end gap-2 pt-1.5 border-t flex-shrink-0">
      <Button variant="outline" onclick={() => open = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>