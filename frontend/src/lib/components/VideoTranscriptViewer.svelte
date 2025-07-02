<script>
  import { Button } from "$lib/components/ui/button";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { 
    Tabs, 
    TabsContent, 
    TabsList, 
    TabsTrigger 
  } from "$lib/components/ui/tabs";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import TextHighlighter from "$lib/components/TextHighlighter.svelte";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import { toast } from "svelte-sonner";

  let { 
    open = $bindable(false),
    video = $bindable(null),
    projectId,
    onHighlightsChange
  } = $props();

  // Transcript video player state (separate from main highlights)
  let transcriptPlayerHighlights = $state([]);
  
  // Derived highlights formatted for EtroVideoPlayer (adds filePath from video)
  let formattedTranscriptHighlights = $derived(
    video && transcriptPlayerHighlights.length > 0 
      ? transcriptPlayerHighlights.map(highlight => ({
          ...highlight,
          filePath: video.filePath,
          videoClipId: video.id,
          videoClipName: video.name
        }))
      : []
  );

  // When video changes, update the transcript player highlights
  $effect(() => {
    if (video) {
      transcriptPlayerHighlights = video.highlights ? [...video.highlights] : [];
    }
  });

  async function handleHighlightsChangeInternal(highlights) {
    if (!video) return;
    
    try {
      // Update the transcript player highlights (local state)
      transcriptPlayerHighlights = [...highlights];
      
      // Call the parent's handler
      if (onHighlightsChange) {
        await onHighlightsChange(highlights);
      }
    } catch (err) {
      console.error("Failed to save highlights:", err);
      toast.error("Failed to save highlights", {
        description: "An error occurred while saving your highlights"
      });
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
      toast.success("Copied to clipboard");
    }
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[1200px] max-h-[90vh] flex flex-col">
    <DialogHeader>
      <DialogTitle>Video Transcript</DialogTitle>
      <DialogDescription>
        {#if video}
          Transcript for {video.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    <ScrollArea class="h-[70vh]">
      {#snippet children()}
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 pr-4 pb-4">
          <!-- Video Player Column -->
          <div class="space-y-4">
            {#if video}
              <!-- Video Player -->
              <div class="bg-background border rounded-lg p-4">
                <h3 class="font-medium mb-3">Video Preview</h3>
                <div class="aspect-video">
                  <EtroVideoPlayer 
                    highlights={formattedTranscriptHighlights}
                    projectId={projectId}
                    enableEyeButton={false}
                    enableReordering={false}
                  />
                </div>
              </div>
            {/if}
          </div>
          
          <!-- Transcript Column -->
          <div class="space-y-4">
            {#if video}
              <!-- Transcript content with tabs -->
              {#if video.transcription}
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <h3 class="font-medium">Transcript</h3>
                    <div class="flex gap-2">
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

                  <Tabs value="full-text" class="w-full">
                    <TabsList class="grid w-full grid-cols-2">
                      <TabsTrigger value="full-text">Full Text</TabsTrigger>
                      <TabsTrigger value="word-by-word" disabled={!video.transcriptionWords || video.transcriptionWords.length === 0}>
                        Word by Word
                      </TabsTrigger>
                    </TabsList>
                    
                    <TabsContent value="full-text" class="space-y-3">
                      <ScrollArea class="h-80 bg-background border rounded-lg">
                        {#snippet children()}
                          <div class="p-4 text-sm leading-relaxed">
                            <TextHighlighter 
                              text={video.transcription} 
                              words={video.transcriptionWords || []} 
                              initialHighlights={video.highlights || []}
                              onHighlightsChange={handleHighlightsChangeInternal}
                            />
                          </div>
                        {/snippet}
                      </ScrollArea>
                      <div class="text-xs text-muted-foreground">
                        Character count: {video.transcription.length}
                      </div>
                    </TabsContent>
                    
                    <TabsContent value="word-by-word" class="space-y-3">
                      {#if video.transcriptionWords && video.transcriptionWords.length > 0}
                        <ScrollArea class="h-80 bg-background border rounded-lg">
                          {#snippet children()}
                            <div class="p-4 space-y-1">
                              {#each video.transcriptionWords as word, index}
                                <div class="flex items-center gap-3 p-2 hover:bg-secondary/30 rounded-md group">
                                  <div class="flex-shrink-0 text-xs text-muted-foreground font-mono bg-secondary px-2 py-1 rounded">
                                    {formatTimestamp(word.start)}
                                  </div>
                                  <div class="flex-1">
                                    <span class="text-sm">{word.word.trim()}</span>
                                  </div>
                                  <div class="flex-shrink-0 text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                                    {(word.end - word.start).toFixed(1)}s
                                  </div>
                                </div>
                              {/each}
                            </div>
                          {/snippet}
                        </ScrollArea>
                        <div class="text-xs text-muted-foreground flex-shrink-0">
                          Word count: {video.transcriptionWords.length}
                        </div>
                      {:else}
                        <div class="flex-1 flex items-center justify-center text-muted-foreground">
                          <div class="text-center">
                            <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <p class="text-lg font-medium">No word-level timing available</p>
                            <p class="text-sm">Word timestamps weren't generated for this transcription.</p>
                          </div>
                        </div>
                      {/if}
                    </TabsContent>
                  </Tabs>
                </div>
              {:else}
                <div class="flex-1 flex items-center justify-center text-muted-foreground">
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
          </div>
        </div>
      {/snippet}
    </ScrollArea>
    
    <!-- Fixed footer buttons -->
    <div class="flex justify-end gap-2 pt-1.5 border-t flex-shrink-0">
      <Button variant="outline" onclick={() => open = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>