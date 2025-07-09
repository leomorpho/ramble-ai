<script>
  import { X } from "@lucide/svelte";
  import { removeNewLine } from "$lib/stores/projectHighlights.js";

  let { 
    newlineItem, 
    index, 
    isDragging, 
    isBeingDragged, 
    showDropIndicatorBefore, 
    showDropIndicatorAfter = false,
    onDragStart, 
    onDragEnd, 
    onDragOver, 
    onDrop 
  } = $props();

  let isHovering = $state(false);
  let deleting = $state(false);

  // Debug logging
  $effect(() => {
    console.log("NewLineItem rendered with:", newlineItem);
    console.log("NewLineItem index:", index);
    console.log("NewLineItem isDragging:", isDragging);
  });

  async function handleDelete() {
    if (deleting) return;
    
    deleting = true;
    try {
      await removeNewLine(index);
    } catch (error) {
      console.error("Failed to remove new line:", error);
    } finally {
      deleting = false;
    }
  }
</script>

<!-- Drop indicator before this item -->
{#if showDropIndicatorBefore}
  <span class="drop-indicator">|</span>
{/if}

<span
  class="newline-container block relative group w-full"
  role="separator"
  draggable="true"
  ondragstart={(e) => onDragStart(e, newlineItem, index)}
  ondragend={onDragEnd}
  ondragover={(e) => onDragOver(e, index)}
  ondrop={(e) => onDrop(e, index)}
  onmouseenter={() => isHovering = true}
  onmouseleave={() => isHovering = false}
  class:being-dragged={isBeingDragged}
  class:drag-active={isDragging}
>
  <div class="newline-visual flex items-center justify-center relative my-3">
    <!-- Thin horizontal line -->
    <div class="h-[2px] bg-primary/30 flex-1"></div>
    
    <!-- Visual indicator for newline -->
    <div class="px-3 py-1 text-xs font-medium text-primary bg-primary/5 border border-primary/20 rounded-md">
      ⏎ Section Break
    </div>
    
    <!-- Right side of the line -->
    <div class="h-[2px] bg-primary/30 flex-1"></div>
    
    <!-- Delete button on hover -->
    {#if isHovering && !isDragging}
      <button
        class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 bg-destructive text-destructive-foreground rounded-full flex items-center justify-center opacity-75 hover:opacity-100 transition-opacity"
        onclick={handleDelete}
        disabled={deleting}
        title="Remove line break"
      >
        <X class="w-2.5 h-2.5" />
      </button>
    {/if}
  </div>
</span>

<!-- Drop indicator after this newline item -->
{#if showDropIndicatorAfter}
  <span class="drop-indicator">|</span>
{/if}

<style>
  .newline-container {
    display: block;
    cursor: grab;
    width: 100%;
  }

  .newline-container:active {
    cursor: grabbing;
  }

  .newline-container.being-dragged {
    opacity: 0.5;
  }

  .newline-visual {
    height: 24px;
    position: relative;
  }

  .drop-indicator {
    color: hsl(var(--primary));
    font-weight: bold;
    display: inline-block;
    animation: pulse 1s infinite;
    margin: 0 2px;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
</style>