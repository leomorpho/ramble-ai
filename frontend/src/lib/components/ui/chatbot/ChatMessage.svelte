<script>
  import { Button } from "$lib/components/ui/button";
  import { Copy, User, Bot, AlertCircle } from "@lucide/svelte";
  import { MESSAGE_TYPES } from "$lib/constants/chatbot.js";
  import { toast } from "svelte-sonner";
  
  let {
    message
  } = $props();
  
  // Format timestamp for display
  function formatTime(timestamp) {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return "";
    }
  }
  
  // Copy message content to clipboard
  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(message.content);
      toast.success("Copied to clipboard");
    } catch (error) {
      console.error("Failed to copy:", error);
      toast.error("Failed to copy");
    }
  }
  
  // Render message content with basic markdown-like formatting
  function renderContent(content) {
    if (!content) return "";
    
    // Simple formatting: **bold**, *italic*, `code`
    return content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/`(.*?)`/g, '<code class="bg-secondary px-1 py-0.5 rounded text-sm">$1</code>');
  }
</script>

<div class="space-y-2">
  {#if message.role === MESSAGE_TYPES.USER}
    <!-- User message -->
    <div class="flex justify-end gap-2">
      <div class="max-w-[80%] group">
        <div class="bg-primary text-primary-foreground rounded-lg px-4 py-3">
          <div class="whitespace-pre-wrap text-sm break-words">{message.content}</div>
        </div>
        <div class="flex items-center justify-end gap-2 mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <span class="text-xs text-muted-foreground">{formatTime(message.timestamp)}</span>
          <Button
            variant="ghost"
            size="icon"
            class="h-6 w-6"
            onclick={copyToClipboard}
            aria-label="Copy message"
          >
            <Copy class="w-3 h-3" />
          </Button>
        </div>
      </div>
      <div class="flex-shrink-0 w-8 h-8 bg-primary rounded-full flex items-center justify-center">
        <User class="w-4 h-4 text-primary-foreground" />
      </div>
    </div>
  {:else if message.role === MESSAGE_TYPES.ASSISTANT}
    <!-- AI message -->
    <div class="flex justify-start gap-2">
      <div class="flex-shrink-0 w-8 h-8 bg-secondary rounded-full flex items-center justify-center">
        <Bot class="w-4 h-4 text-secondary-foreground" />
      </div>
      <div class="max-w-[80%] group">
        <div class="bg-secondary rounded-lg px-4 py-3">
          <div class="whitespace-pre-wrap text-sm break-words prose prose-sm dark:prose-invert max-w-none">
            {@html renderContent(message.content)}
          </div>
        </div>
        <div class="flex items-center gap-2 mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <span class="text-xs text-muted-foreground">{formatTime(message.timestamp)}</span>
          <Button
            variant="ghost"
            size="icon"
            class="h-6 w-6"
            onclick={copyToClipboard}
            aria-label="Copy message"
          >
            <Copy class="w-3 h-3" />
          </Button>
        </div>
      </div>
    </div>
  {:else if message.role === MESSAGE_TYPES.ERROR}
    <!-- Error message -->
    <div class="flex justify-center">
      <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg px-4 py-3 max-w-[80%]">
        <div class="flex items-center gap-2">
          <AlertCircle class="w-4 h-4 flex-shrink-0" />
          <div class="text-sm">{message.content}</div>
        </div>
      </div>
    </div>
  {:else if message.role === MESSAGE_TYPES.SYSTEM}
    <!-- System message -->
    <div class="flex justify-center">
      <div class="bg-muted text-muted-foreground rounded-lg px-4 py-2 max-w-[80%]">
        <div class="text-xs text-center">{message.content}</div>
      </div>
    </div>
  {/if}
</div>