<script>
  import { onMount, tick } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import { Settings, RefreshCw, Trash2 } from "@lucide/svelte";
  import MessageList from "./MessageList.svelte";
  import MessageInput from "./MessageInput.svelte";
  import ChatSettings from "./ChatSettings.svelte";
  import { ENDPOINT_CONFIGS, AVAILABLE_MODELS } from "$lib/constants/chatbot.js";
  import { SendChatMessage, GetChatHistory, ClearChatHistory } from "$lib/wailsjs/go/main/App.js";
  import { toast } from "svelte-sonner";
  
  let {
    endpointId,
    projectId,
    contextData = {},
    messages = $bindable([]),
    sessionId = $bindable(null),
    title = "AI Assistant",
    description = "Chat with AI about your project",
    icon = "🤖"
  } = $props();
  
  // Get endpoint configuration
  let config = $derived(ENDPOINT_CONFIGS[endpointId] || {});
  
  // Component state
  let loading = $state(false);
  let settingsOpen = $state(false);
  let selectedModel = $state(AVAILABLE_MODELS[0].value);
  let customModelValue = $state("");
  
  // Set default model when config changes
  $effect(() => {
    if (config.defaultModel && selectedModel === AVAILABLE_MODELS[0].value) {
      selectedModel = config.defaultModel;
    }
  });
  
  // Load chat history when component mounts
  onMount(async () => {
    if (projectId && endpointId) {
      await loadChatHistory();
    }
  });
  
  // Reload when projectId or endpointId changes
  $effect(() => {
    if (projectId && endpointId) {
      loadChatHistory();
    }
  });
  
  async function loadChatHistory() {
    try {
      const history = await GetChatHistory(projectId, endpointId);
      if (history && history.messages) {
        messages = history.messages;
        sessionId = history.sessionId;
      }
    } catch (error) {
      console.warn("Could not load chat history:", error);
      // This is not a critical error, so we don't show a toast
    }
  }
  
  async function handleSendMessage(messageText) {
    if (!messageText.trim() || loading) return;
    
    loading = true;
    
    try {
      // Add user message immediately for better UX
      const userMessage = {
        id: Date.now().toString(),
        role: "user",
        content: messageText,
        timestamp: new Date().toISOString()
      };
      
      messages = [...messages, userMessage];
      
      // Send message to backend
      const response = await SendChatMessage({
        projectId,
        endpointId,
        message: messageText,
        sessionId,
        contextData,
        model: selectedModel === "custom" ? customModelValue : selectedModel
      });
      
      // Update session ID if we got a new one
      if (response.sessionId) {
        sessionId = response.sessionId;
      }
      
      // Add AI response
      if (response.message) {
        const aiMessage = {
          id: response.messageId || (Date.now() + 1).toString(),
          role: "assistant",
          content: response.message,
          timestamp: new Date().toISOString()
        };
        
        messages = [...messages, aiMessage];
      }
      
    } catch (error) {
      console.error("Failed to send message:", error);
      
      // Add error message to chat
      const errorMessage = {
        id: (Date.now() + 2).toString(),
        role: "error",
        content: "Failed to process message. Please try again.",
        timestamp: new Date().toISOString()
      };
      
      messages = [...messages, errorMessage];
      toast.error("Failed to send message");
    } finally {
      loading = false;
    }
  }
  
  async function handleClearHistory() {
    try {
      await ClearChatHistory(projectId, endpointId);
      messages = [];
      sessionId = null;
      toast.success("Chat history cleared");
    } catch (error) {
      console.error("Failed to clear history:", error);
      toast.error("Failed to clear chat history");
    }
  }
  
  function handleRefresh() {
    loadChatHistory();
    toast.success("Chat refreshed");
  }
</script>

<div class="flex flex-col h-[600px]">
  <!-- Header -->
  <div class="px-6 py-4 border-b border-border">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-2xl">{icon}</span>
        <div>
          <h3 class="text-lg font-semibold text-left">{title}</h3>
          <p class="text-sm text-muted-foreground text-left">{description}</p>
        </div>
      </div>
      
      <div class="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          onclick={handleRefresh}
          disabled={loading}
          aria-label="Refresh chat"
        >
          <RefreshCw class="w-4 h-4" />
        </Button>
        
        <Button
          variant="ghost"
          size="icon"
          onclick={handleClearHistory}
          disabled={loading || messages.length === 0}
          aria-label="Clear history"
        >
          <Trash2 class="w-4 h-4" />
        </Button>
        
        <Button
          variant="ghost"
          size="icon"
          onclick={() => settingsOpen = !settingsOpen}
          aria-label="Settings"
        >
          <Settings class="w-4 h-4" />
        </Button>
      </div>
    </div>
  </div>
  
  <!-- Settings Panel -->
  {#if settingsOpen}
    <div class="border-b border-border">
      <ChatSettings
        bind:selectedModel
        bind:customModelValue
        availableModels={AVAILABLE_MODELS}
      />
    </div>
  {/if}
  
  <!-- Messages Area -->
  <div class="flex-1 flex flex-col min-h-0">
    <ScrollArea class="flex-1 px-6">
      <MessageList {messages} {loading} {config} />
    </ScrollArea>
    
    <!-- Message Input -->
    <div class="border-t border-border px-6 py-4">
      <MessageInput
        onSendMessage={handleSendMessage}
        {loading}
        placeholder="Ask me about your {config.title?.toLowerCase() || 'project'}..."
      />
    </div>
  </div>
</div>