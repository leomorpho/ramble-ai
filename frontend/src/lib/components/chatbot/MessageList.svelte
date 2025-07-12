<script>
  import { tick } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import { MESSAGE_TYPES, CHATBOT_ENDPOINTS } from "$lib/constants/chatbot.js";
  import { Button } from "$lib/components/ui/button";
  import { Sparkles } from "@lucide/svelte";
  
  let {
    messages = [],
    loading = false,
    config = {},
    endpointId = "",
    onSendMessage = () => {},
    projectId = null
  } = $props();
  
  let messagesContainer = $state(null);
  let showSuggestions = $state(true);
  
  // Auto-scroll to bottom when new messages are added
  $effect(() => {
    if (messages.length > 0 && messagesContainer) {
      tick().then(() => {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      });
    }
  });
  
  // Suggestion functions for highlight ordering
  const suggestions = [
    {
      id: "reorder",
      title: "Smart Reorder",
      description: "Let AI reorganize highlights for better flow",
      icon: "🔄",
      message: "Please analyze my highlights and reorder them for maximum engagement and narrative flow."
    },
    {
      id: "analyze",
      title: "Analyze Content",
      description: "Get insights about your highlights",
      icon: "📊",
      message: "Please analyze my highlights and provide insights about content themes, structure, and engagement potential."
    },
    {
      id: "hook",
      title: "Create Hook",
      description: "Optimize opening for maximum engagement",
      icon: "🎯",
      message: "Help me reorder these highlights to create a compelling hook that grabs viewers' attention from the start."
    },
    {
      id: "flow",
      title: "Improve Flow",
      description: "Enhance narrative and pacing",
      icon: "🌊",
      message: "Please reorder my highlights to improve narrative flow and create better pacing throughout the video."
    },
    {
      id: "conclusion",
      title: "Create Conclusion",
      description: "Structure highlights with a strong ending",
      icon: "🎬",
      message: "Help me reorder these highlights to create a strong conclusion section that leaves viewers satisfied and engaged."
    }
  ];
  
  function handleSuggestionClick(suggestion) {
    showSuggestions = false;
    onSendMessage(suggestion.message);
  }
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
      
      <!-- Suggestions for highlight ordering endpoint -->
      {#if endpointId === CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING && projectId && showSuggestions}
        <div class="mt-8 space-y-4">
          <div>
            <h3 class="text-sm font-medium mb-2">Quick Actions</h3>
            <p class="text-xs text-muted-foreground mb-4">Try these AI-powered suggestions to optimize your highlights</p>
          </div>
          
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-2xl mx-auto">
            {#each suggestions as suggestion}
              <Button
                variant="outline"
                class="h-auto p-4 flex flex-col items-start text-left space-y-2 hover:bg-accent transition-colors"
                onclick={() => handleSuggestionClick(suggestion)}
                disabled={loading}
              >
                <div class="flex items-center gap-2 w-full">
                  <span class="text-lg">{suggestion.icon}</span>
                  <span class="font-medium text-sm">{suggestion.title}</span>
                </div>
                <p class="text-xs text-muted-foreground text-left">{suggestion.description}</p>
              </Button>
            {/each}
          </div>
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
    
    <!-- Show suggestions toggle when there are messages -->
    {#if endpointId === CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING && projectId && !showSuggestions && !loading}
      <div class="flex justify-center pt-4">
        <Button
          variant="outline"
          size="sm"
          onclick={() => showSuggestions = true}
          class="flex items-center gap-2"
        >
          <Sparkles class="w-4 h-4" />
          Show Quick Actions
        </Button>
      </div>
    {/if}
    
    <!-- Suggestions in chat -->
    {#if endpointId === CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING && projectId && showSuggestions && messages.length > 0}
      <div class="space-y-4 pt-4 border-t border-border">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium">Quick Actions</h3>
          <Button
            variant="ghost"
            size="sm"
            onclick={() => showSuggestions = false}
            class="text-xs text-muted-foreground"
          >
            Hide
          </Button>
        </div>
        
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {#each suggestions as suggestion}
            <Button
              variant="outline"
              class="h-auto p-3 flex flex-col items-start text-left space-y-1 hover:bg-accent transition-colors"
              onclick={() => handleSuggestionClick(suggestion)}
              disabled={loading}
            >
              <div class="flex items-center gap-2 w-full">
                <span class="text-sm">{suggestion.icon}</span>
                <span class="font-medium text-xs">{suggestion.title}</span>
              </div>
              <p class="text-xs text-muted-foreground text-left">{suggestion.description}</p>
            </Button>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>