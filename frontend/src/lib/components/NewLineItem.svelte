<script>
  import { X, Edit3, Check } from "@lucide/svelte";
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
    onDrop,
    // New props for title editing
    onTitleChange = null,
    // Enable/disable editing and deleting
    enableEdit = true,
    enableDelete = true,
    enableDrag = true
  } = $props();

  let isHovering = $state(false);
  let deleting = $state(false);
  let editing = $state(false);
  let editingTitle = $state('');
  let titleInput = $state(null);

  // Get title from newlineItem
  $effect(() => {
    if (newlineItem && newlineItem.title) {
      editingTitle = newlineItem.title;
    }
  });

  // Helper function to get title
  function getTitle() {
    return newlineItem?.title || '';
  }

  // Debug logging
  $effect(() => {
    console.log("NewLineItem rendered with:", newlineItem);
    console.log("NewLineItem index:", index);
    console.log("NewLineItem isDragging:", isDragging);
    console.log("NewLineItem title:", getTitle());
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

  function startEditing() {
    editing = true;
    editingTitle = getTitle();
    // Focus the input after it's rendered
    setTimeout(() => {
      titleInput?.focus();
    }, 0);
  }

  function cancelEditing() {
    editing = false;
    editingTitle = getTitle();
  }

  function saveTitle() {
    if (onTitleChange) {
      onTitleChange(index, editingTitle);
    }
    editing = false;
  }

  function handleKeydown(event) {
    if (event.key === 'Enter') {
      saveTitle();
    } else if (event.key === 'Escape') {
      cancelEditing();
    }
  }

  function handleDragStart(event) {
    if (!enableDrag) {
      event.preventDefault();
      return;
    }
    onDragStart(event, newlineItem, index);
  }
</script>

<!-- Drop indicator before this item -->
{#if showDropIndicatorBefore}
  <span class="drop-indicator">|</span>
{/if}

<span
  class="newline-container block relative group w-full"
  class:non-draggable={!enableDrag}
  role="separator"
  draggable={enableDrag}
  ondragstart={handleDragStart}
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
    
    <!-- Visual indicator for newline with title -->
    <div class="px-3 py-1 text-xs font-medium text-primary bg-primary/5 border border-primary/20 rounded-md flex items-center gap-2">
      <span>⏎</span>
      {#if editing}
        <input
          bind:this={titleInput}
          bind:value={editingTitle}
          onkeydown={handleKeydown}
          onblur={saveTitle}
          class="bg-transparent border-none outline-none text-xs font-medium text-primary placeholder:text-primary/50"
          placeholder="Section title..."
          style="width: {Math.max(80, editingTitle.length * 8)}px;"
        />
      {:else}
        <span class="select-none">
          {getTitle() || 'Section Break'}
        </span>
      {/if}
    </div>
    
    <!-- Right side of the line -->
    <div class="h-[2px] bg-primary/30 flex-1"></div>
    
    <!-- Action buttons on hover -->
    {#if isHovering && !isDragging && !editing && (enableEdit || enableDelete)}
      <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
        <!-- Edit button -->
        {#if enableEdit}
          <button
            class="w-4 h-4 bg-primary text-primary-foreground rounded-full flex items-center justify-center opacity-75 hover:opacity-100 transition-opacity"
            onclick={startEditing}
            title="Edit section title"
          >
            <Edit3 class="w-2.5 h-2.5" />
          </button>
        {/if}
        
        <!-- Delete button -->
        {#if enableDelete}
          <button
            class="w-4 h-4 bg-destructive text-destructive-foreground rounded-full flex items-center justify-center opacity-75 hover:opacity-100 transition-opacity"
            onclick={handleDelete}
            disabled={deleting}
            title="Remove line break"
          >
            <X class="w-2.5 h-2.5" />
          </button>
        {/if}
      </div>
    {/if}
    
    <!-- Save button when editing -->
    {#if editing && enableEdit}
      <button
        class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 bg-green-600 text-white rounded-full flex items-center justify-center opacity-75 hover:opacity-100 transition-opacity"
        onclick={saveTitle}
        title="Save title"
      >
        <Check class="w-2.5 h-2.5" />
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

  .newline-container.non-draggable {
    cursor: default;
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