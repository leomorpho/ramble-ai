<script>
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { Copy, User, Bot, AlertCircle, CheckCircle, ChevronDown, ChevronUp } from "@lucide/svelte";
  import { MESSAGE_TYPES } from "$lib/constants/chatbot.js";
  import { toast } from "svelte-sonner";
  
  let {
    message
  } = $props();
  
  let showTechnicalDetails = $state(false);
  
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
        {#if message.hasActions}
          <CheckCircle class="w-4 h-4 text-green-600" />
        {:else}
          <Bot class="w-4 h-4 text-secondary-foreground" />
        {/if}
      </div>
      <div class="max-w-[80%] group">
        <!-- Action Summary (if present) -->
        {#if message.actionSummary}
          <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg px-4 py-3 mb-2">
            <div class="flex items-start gap-2 mb-2">
              <CheckCircle class="w-4 h-4 text-green-600 flex-shrink-0 mt-0.5" />
              <div class="text-sm font-medium text-green-800 dark:text-green-200">Actions Completed</div>
            </div>
            
            {#if message.actionsPerformed && message.actionsPerformed.length > 0}
              <div class="flex flex-wrap gap-1 mb-2">
                {#each message.actionsPerformed as action}
                  <Badge variant="secondary" class="text-xs bg-green-100 dark:bg-green-800 text-green-800 dark:text-green-200">
                    {action.replace(/_/g, ' ')}
                  </Badge>
                {/each}
              </div>
            {/if}
            
            <div class="text-sm text-green-700 dark:text-green-300 prose prose-sm dark:prose-invert max-w-none">
              {@html renderContent(message.actionSummary)}
            </div>
            
            <!-- Technical Details (collapsible) -->
            {#if message.functionResults && message.functionResults.length > 0}
              <div class="mt-3 pt-2 border-t border-green-200 dark:border-green-700">
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-xs h-6 p-1 text-green-700 dark:text-green-300"
                  onclick={() => showTechnicalDetails = !showTechnicalDetails}
                >
                  {showTechnicalDetails ? 'Hide' : 'Show'} Technical Details
                  {#if showTechnicalDetails}
                    <ChevronUp class="w-3 h-3 ml-1" />
                  {:else}
                    <ChevronDown class="w-3 h-3 ml-1" />
                  {/if}
                </Button>
                
                {#if showTechnicalDetails}
                  <div class="mt-2 space-y-2">
                    {#each message.functionResults as result}
                      <div class="bg-green-100 dark:bg-green-800/30 rounded p-2">
                        <div class="text-xs font-mono">
                          <div class="text-green-600 dark:text-green-400 font-medium">{result.functionName}</div>
                          <div class="text-green-700 dark:text-green-300">
                            Status: {result.success ? '✅ Success' : '❌ Failed'}
                          </div>
                          {#if result.error}
                            <div class="text-red-600 dark:text-red-400">Error: {result.error}</div>
                          {/if}
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/if}
        
        <!-- Regular Message Content -->
        {#if message.content && (!message.actionSummary || message.content !== message.actionSummary)}
          <div class="bg-secondary rounded-lg px-4 py-3">
            <div class="whitespace-pre-wrap text-sm break-words prose prose-sm dark:prose-invert max-w-none">
              {@html renderContent(message.content)}
            </div>
          </div>
        {/if}
        
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