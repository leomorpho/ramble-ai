<script>
  import { tick } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import { MESSAGE_TYPES } from "$lib/constants/chatbot.js";
  
  let {
    messages = [],
    loading = false,
    config = {}
  } = $props();
  
  let messagesContainer = $state(null);
  
  // Auto-scroll to bottom when new messages are added
  $effect(() => {
    if (messages.length > 0 && messagesContainer) {
      tick().then(() => {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      });
    }
  });
</script>

<div bind:this={messagesContainer} class="py-4 space-y-4">
  {#if messages.length === 0 && !loading}
    <!-- Empty state -->
    <div class="text-center text-muted-foreground py-12">
      <div class="text-4xl mb-4">{config.icon || "🤖"}</div>
      <h3 class="font-medium mb-2">Ready to help!</h3>
      <p class="text-sm">
        {config.description || "Start a conversation with the AI assistant"}
      </p>
      {#if config.systemPrompt}
        <div class="mt-4 p-4 bg-secondary/30 rounded-lg text-xs text-left max-w-md mx-auto">
          <p class="font-medium mb-2">I can help you with:</p>
          <p>{config.systemPrompt}</p>
        </div>
      {/if}
    </div>
  {:else}
    <!-- Messages -->
    {#each messages as message (message.id || message.timestamp)}
      <ChatMessage {message} />
    {/each}
    
    <!-- Loading indicator -->
    {#if loading}
      <div class="flex justify-start">
        <div class="bg-secondary rounded-lg px-4 py-3 max-w-[80%]">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 bg-muted-foreground rounded-full animate-pulse"></div>
            <div class="w-2 h-2 bg-muted-foreground rounded-full animate-pulse" style="animation-delay: 0.2s"></div>
            <div class="w-2 h-2 bg-muted-foreground rounded-full animate-pulse" style="animation-delay: 0.4s"></div>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>