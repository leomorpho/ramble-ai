<script>
  import { CornerDownLeft } from "@lucide/svelte";
  import { insertNewLine } from "$lib/stores/projectHighlights.js";

  let { position } = $props();

  let isHovering = $state(false);
  let adding = $state(false);

  async function handleAddNewLine() {
    if (adding) return;
    
    adding = true;
    try {
      await insertNewLine(position);
    } catch (error) {
      console.error("Failed to add new line:", error);
    } finally {
      adding = false;
    }
  }
</script>

<span
  class="relative inline-block w-0.5 h-5 mx-0.5 cursor-pointer"
  role="button"
  tabindex="0"
  onmouseenter={() => isHovering = true}
  onmouseleave={() => isHovering = false}
  onclick={handleAddNewLine}
  title="Add line break"
>
  <!-- Minimal transparent indicator -->
  <div class="w-0.5 h-4 bg-border/20 mt-0.5"></div>
  
  <!-- Absolutely positioned newline icon on hover -->
  {#if isHovering}
    <div
      class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 
             w-4 h-4 transition-all duration-200 ease-in-out z-10
             {adding ? 'opacity-50 animate-pulse' : ''}"
    >
      <CornerDownLeft 
        class="w-4 h-4 text-primary transition-colors duration-200
               hover:text-primary/80" 
      />
    </div>
  {/if}
</span>