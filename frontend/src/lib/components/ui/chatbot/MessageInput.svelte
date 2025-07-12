<script>
  import { Button } from "$lib/components/ui/button";
  import { Send } from "@lucide/svelte";
  import AutoResizeTextarea from "../AutoResizeTextarea.svelte";
  import { tick } from "svelte";
  
  let {
    onSendMessage = () => {},
    loading = false,
    placeholder = "Type your message..."
  } = $props();
  
  let currentMessage = $state("");
  let textareaElement = $state(null);
  
  async function handleSendMessage() {
    if (!currentMessage.trim() || loading) return;
    
    const message = currentMessage.trim();
    currentMessage = "";
    
    // Focus textarea after sending
    tick().then(() => {
      if (textareaElement && typeof textareaElement.focus === 'function') {
        textareaElement.focus();
      }
    });
    
    await onSendMessage(message);
  }
  
  function handleKeydown(event) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    }
  }
  
  // Auto-focus on mount
  $effect(() => {
    if (textareaElement && typeof textareaElement.focus === 'function') {
      tick().then(() => {
        textareaElement.focus();
      });
    }
  });
</script>

<div class="flex gap-3">
  <AutoResizeTextarea
    bind:this={textareaElement}
    bind:value={currentMessage}
    {placeholder}
    class="flex-1 min-h-[44px] max-h-32 resize-none"
    onkeydown={handleKeydown}
    disabled={loading}
  />
  <Button
    onclick={handleSendMessage}
    disabled={!currentMessage.trim() || loading}
    size="icon"
    class="h-[44px] w-[44px] flex-shrink-0"
    aria-label="Send message"
  >
    {#if loading}
      <div class="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
    {:else}
      <Send class="w-4 h-4" />
    {/if}
  </Button>
</div>