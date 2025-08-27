<script>
  import { Button } from "$lib/components/ui/button";
  import { Captions, Mic, Video, Trash2 } from "@lucide/svelte";
  import VideoPreviewDialog from "./VideoPreviewDialog.svelte";
  import VideoTranscriptViewer from "./VideoTranscriptViewer.svelte";
  import { 
    getTranscriptionState, 
    getTranscriptionButtonLabel, 
    canTranscribe, 
    isTranscribing as isTranscribingState,
    TranscriptionState
  } from "$lib/utils/transcription.js";

  let { 
    clip,
    onDelete,
    onStartTranscription,
    formatFileSize,
    projectId,
    highlights,
    onHighlightsChange
  } = $props();

  let transcriptionState = $derived(getTranscriptionState(clip));
  let transcriptionButtonLabel = $derived(getTranscriptionButtonLabel(clip));
  let canTranscribeClip = $derived(canTranscribe(clip));
  let isTranscribing = $derived(isTranscribingState(clip));

  let previewDialogOpen = $state(false);
  let transcriptionDialogOpen = $state(false);

  // Calculate highlight count for this video clip
  let highlightCount = $derived(
    highlights?.filter(h => h.videoClipId === clip.id && h.filePath === clip.filePath).length || 0
  );

  function openPreview() {
    previewDialogOpen = true;
  }

  function openTranscription() {
    transcriptionDialogOpen = true;
  }
</script>

<div class="border rounded overflow-hidden">
  <!-- Video thumbnail -->
  {#if clip.exists && clip.thumbnailUrl}
    <div 
      class="relative group cursor-pointer" 
      onclick={openPreview}
      onkeydown={(e) => e.key === 'Enter' && openPreview()}
      role="button"
      tabindex="0"
      aria-label="Preview video {clip.name}"
    >
      <img 
        src={clip.thumbnailUrl} 
        alt="Video thumbnail for {clip.name}"
        class="w-full h-24 object-cover bg-muted"
        loading="lazy"
      />
      <!-- Play overlay -->
      <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
        <div class="w-10 h-10 bg-white/80 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
          <svg class="w-5 h-5 text-black ml-0.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </div>
      </div>
    </div>
  {:else}
    <div class="w-full h-24 bg-muted flex items-center justify-center">
      <div class="text-center text-muted-foreground">
        <svg class="w-8 h-8 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        <p class="text-xs">
          {clip.exists ? 'Generating thumbnail...' : 'Video not found'}
        </p>
      </div>
    </div>
  {/if}

  <div class="p-3">
    <div class="flex justify-between items-start mb-2">
      <div class="flex-1 min-w-0">
        <h3 class="font-medium text-sm truncate" title={clip.name}>{clip.name}</h3>
        <p class="text-xs text-muted-foreground truncate" title={clip.fileName}>
          {clip.fileName}
        </p>
      </div>
      <Button 
        variant="ghost" 
        size="sm" 
        onclick={() => onDelete(clip.id)}
        class="ml-2 text-destructive hover:text-destructive h-6 w-6 p-0"
      >
        <Trash2 class="w-3 h-3" />
      </Button>
    </div>

{#snippet infoCard(label, value, valueClass = "")}
  <div class="border rounded p-1.5 text-center">
    <div class="text-xs text-muted-foreground">{label}</div>
    <div class="text-xs font-medium {valueClass}">{value}</div>
  </div>
{/snippet}

    <div class="grid grid-cols-2 md:grid-cols-4 gap-1.5 mb-2">
      <!-- Size Card -->
      {@render infoCard("Size", formatFileSize(clip.fileSize || 0))}
      
      <!-- Status Card -->
      {@render infoCard("Status", clip.exists ? "Found" : "Missing", clip.exists ? "text-green-600" : "text-destructive")}
      
      <!-- Highlights Card -->
      {@render infoCard("Highlights", highlightCount, highlightCount > 0 ? "text-green-600" : "text-muted-foreground")}
      
      <!-- Duration Card (if available) -->
      {#if clip.duration}
        {@render infoCard("Duration", `${Math.round(clip.duration)}s`)}
      {:else if clip.format}
        <!-- Format Card (if no duration but format available) -->
        {@render infoCard("Format", clip.format, "font-mono uppercase")}
      {/if}
    </div>
    
    <!-- Format Card (separate row if both duration and format are available) -->
    {#if clip.duration && clip.format}
      <div class="flex justify-start mb-2">
        <div class="w-1/2 md:w-1/4">
          {@render infoCard("Format", clip.format, "font-mono uppercase")}
        </div>
      </div>
    {/if}

    <!-- Action buttons -->
    <div class="space-y-1.5">
      <!-- Preview button -->
      <Button 
        variant="outline" 
        size="sm" 
        onclick={openPreview}
        disabled={!clip.exists}
        class="w-full h-7 text-xs"
      >
        <Video class="w-3 h-3 mr-1"/>
        {clip.exists ? 'Preview' : 'Missing'}
      </Button>

      <!-- Transcription buttons -->
      <div class="flex gap-1.5">
        {#if transcriptionState === TranscriptionState.COMPLETED}
          <Button 
            variant="outline" 
            size="sm" 
            onclick={openTranscription}
            class="flex-1 h-7 text-xs"
          >
            <Captions class="w-3 h-3 mr-1"/>
            Transcript
          </Button>
          <Button 
            variant="ghost" 
            size="sm" 
            onclick={() => onStartTranscription(clip)}
            disabled={isTranscribing || !clip.exists}
            class="h-7 w-7 p-0"
            title="Re-transcribe"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </Button>
        {:else}
          <Button 
            variant="outline" 
            size="sm" 
            onclick={() => onStartTranscription(clip)}
            disabled={!canTranscribeClip || !clip.exists}
            class="w-full h-7 text-xs {isTranscribing ? 'animate-pulse' : ''}"
          >
            {#if isTranscribing}
              <svg class="w-3 h-3 mr-1 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            {:else if transcriptionState === TranscriptionState.ERROR}
              <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            {:else}
              <Mic class="w-3 h-3 mr-1"/>
            {/if}
            {transcriptionButtonLabel}
          </Button>
        {/if}
      </div>
    </div>
  </div>
</div>

<!-- Video Preview Dialog -->
<VideoPreviewDialog bind:open={previewDialogOpen} video={clip} />

<!-- Transcription Viewer Dialog -->
<VideoTranscriptViewer
  bind:open={transcriptionDialogOpen}
  video={clip}
  {projectId}
  {highlights}
  onHighlightsChange={onHighlightsChange}
/>