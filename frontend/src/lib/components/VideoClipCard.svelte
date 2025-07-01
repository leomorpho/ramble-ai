<script>
  import { Button } from "$lib/components/ui/button";
  import { Captions, Mic, Video } from "@lucide/svelte";

  let { 
    clip,
    isTranscribing = false,
    onPreview,
    onDelete,
    onViewTranscription,
    onStartTranscription,
    formatFileSize
  } = $props();
</script>

<div class="bg-secondary/30 rounded-lg overflow-hidden border">
  <!-- Video thumbnail -->
  {#if clip.exists && clip.thumbnailUrl}
    <div 
      class="relative group cursor-pointer" 
      onclick={() => onPreview(clip)}
      onkeydown={(e) => e.key === 'Enter' && onPreview(clip)}
      role="button"
      tabindex="0"
      aria-label="Preview video {clip.name}"
    >
      <img 
        src={clip.thumbnailUrl} 
        alt="Video thumbnail for {clip.name}"
        class="w-full h-48 object-cover bg-muted"
        loading="lazy"
      />
      <!-- Play overlay -->
      <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
        <div class="w-16 h-16 bg-white/80 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
          <svg class="w-8 h-8 text-black ml-1" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </div>
      </div>
    </div>
  {:else}
    <div class="w-full h-48 bg-muted flex items-center justify-center">
      <div class="text-center text-muted-foreground">
        <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        <p class="text-sm">
          {clip.exists ? 'Generating thumbnail...' : 'Video not found'}
        </p>
      </div>
    </div>
  {/if}

  <div class="p-4">
    <div class="flex justify-between items-start mb-3">
      <div class="flex-1 min-w-0">
        <h3 class="font-semibold truncate" title={clip.name}>{clip.name}</h3>
        <p class="text-sm text-muted-foreground truncate" title={clip.fileName}>
          {clip.fileName}
        </p>
      </div>
      <Button 
        variant="ghost" 
        size="sm" 
        onclick={() => onDelete(clip.id)}
        class="ml-2 text-destructive hover:text-destructive"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
      </Button>
    </div>

    <div class="space-y-2 text-xs text-muted-foreground">
      <div class="flex justify-between">
        <span>Format:</span>
        <span class="font-mono uppercase">{clip.format || 'unknown'}</span>
      </div>
      <div class="flex justify-between">
        <span>Size:</span>
        <span>{formatFileSize(clip.fileSize || 0)}</span>
      </div>
      {#if clip.width && clip.height}
        <div class="flex justify-between">
          <span>Resolution:</span>
          <span>{clip.width}×{clip.height}</span>
        </div>
      {/if}
      {#if clip.duration}
        <div class="flex justify-between">
          <span>Duration:</span>
          <span>{Math.round(clip.duration)}s</span>
        </div>
      {/if}
      <div class="flex justify-between">
        <span>Status:</span>
        <span class={clip.exists ? "text-green-600" : "text-destructive"}>
          {clip.exists ? "Found" : "Missing"}
        </span>
      </div>
    </div>

    {#if clip.description}
      <div class="mt-3 pt-3 border-t border-border">
        <p class="text-sm text-muted-foreground">{clip.description}</p>
      </div>
    {/if}

    <!-- Action buttons -->
    <div class="mt-3 pt-3 border-t border-border space-y-2">
      <!-- Preview button -->
      <Button 
        variant="outline" 
        size="sm" 
        onclick={() => onPreview(clip)}
        disabled={!clip.exists}
        class="w-full"
      >
        <Video/>
        {clip.exists ? 'Preview Video' : 'File Missing'}
      </Button>

      <!-- Transcription buttons -->
      <div class="flex gap-2">
        {#if clip.transcription}
          <Button 
            variant="outline" 
            size="sm" 
            onclick={() => onViewTranscription(clip)}
            class="flex-1"
          >
            <Captions/>
            View Transcript
          </Button>
          <Button 
            variant="ghost" 
            size="sm" 
            onclick={() => onStartTranscription(clip)}
            disabled={isTranscribing || !clip.exists}
            class="px-3"
            title="Re-transcribe"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </Button>
        {:else}
          <Button 
            variant="outline" 
            size="sm" 
            onclick={() => onStartTranscription(clip)}
            disabled={isTranscribing || !clip.exists}
            class="w-full"
          >
            {#if isTranscribing}
              <svg class="w-4 h-4 mr-2 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Transcribing...
            {:else}
              <Mic/>
              Start Transcription
            {/if}
          </Button>
        {/if}
      </div>
    </div>
  </div>
</div>