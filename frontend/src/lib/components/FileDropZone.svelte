<script>
  let { dragActive, addingClip, onFileInputChange, onOpenFileDialog, onKeyDown } = $props();

  let fileInput = $state();

  function openFileDialog() {
    onOpenFileDialog?.(fileInput);
  }

  function handleKeyDown(event) {
    onKeyDown?.(event, fileInput);
  }

  function handleFileInputChange(event) {
    onFileInputChange?.(event);
  }
</script>

<div class="mb-6">
  <!-- Hidden file input -->
  <input
    bind:this={fileInput}
    type="file"
    multiple
    accept="video/*"
    onchange={handleFileInputChange}
    class="hidden"
  />

  <!-- Drop zone with Wails drop target support -->
  <div
    role="button"
    tabindex="0"
    aria-label="Drop video files from file manager or click to browse"
    onclick={openFileDialog}
    onkeydown={handleKeyDown}
    style="--wails-drop-target: drop"
    class="border-2 border-dashed rounded-lg p-8 text-center transition-all duration-200 cursor-pointer
           {dragActive
      ? 'border-primary bg-primary/5'
      : 'border-border hover:border-primary'}
           {addingClip ? 'pointer-events-none opacity-50' : ''}"
  >
    <div class="flex flex-col items-center gap-4">
      <svg
        class="w-12 h-12 text-muted-foreground/50"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
        />
      </svg>
      <div>
        <p class="text-lg font-medium">
          {#if addingClip}
            Adding video clips...
          {:else if dragActive}
            Drop video files now
          {:else}
            Drop video files here or click to browse
          {/if}
        </p>
        <p class="text-sm text-muted-foreground">
          Supports MP4, MOV, AVI, MKV, WMV, FLV, WebM, and more
        </p>
        <p class="text-xs text-muted-foreground mt-1">
          {#if dragActive}
            Release to add files to your project
          {:else}
            Drag video files from your file manager anywhere in this
            window
          {/if}
        </p>
      </div>
    </div>
  </div>
</div>