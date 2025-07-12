<script>
  import { onMount, onDestroy, tick } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Settings, RefreshCw, Trash2 } from "@lucide/svelte";
  import MessageList from "./MessageList.svelte";
  import MessageInput from "./MessageInput.svelte";
  import ChatSettings from "./ChatSettings.svelte";
  import { ENDPOINT_CONFIGS, AVAILABLE_MODELS } from "$lib/constants/chatbot.js";
  import { SendChatMessage, GetChatHistory, ClearChatHistory, GetProjectHighlights, GetProjectHighlightOrderWithTitles } from "$lib/wailsjs/go/main/App.js";
  import { toast } from "svelte-sonner";
  import { 
    connectChatbotSession, 
    getChatbotMessages, 
    getChatbotSessionId,
    clearChatbotMessages,
    updateChatbotSessionId,
    addChatbotMessage
  } from "$lib/stores/chatbotRealtime.js";
  
  let {
    endpointId,
    projectId,
    contextData = {},
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
  
  // Real-time stores
  let realtimeMessages = $derived(getChatbotMessages(projectId, endpointId));
  let realtimeSessionId = $derived(getChatbotSessionId(projectId, endpointId));
  let unsubscribeRealtime = null;
  
  // Set default model when config changes
  $effect(() => {
    if (config.defaultModel && selectedModel === AVAILABLE_MODELS[0].value) {
      selectedModel = config.defaultModel;
    }
  });
  
  // Load chat history and set up real-time connection when component mounts
  onMount(async () => {
    if (projectId && endpointId) {
      await loadChatHistory();
    }
  });
  
  // Clean up real-time connection when component unmounts
  onDestroy(() => {
    if (unsubscribeRealtime) {
      unsubscribeRealtime();
      unsubscribeRealtime = null;
    }
  });
  
  // Reload and reconnect when projectId or endpointId changes
  $effect(() => {
    if (projectId && endpointId) {
      loadChatHistory();
      setupRealtimeConnection();
    }
    
    // Cleanup function for effect
    return () => {
      if (unsubscribeRealtime) {
        unsubscribeRealtime();
        unsubscribeRealtime = null;
      }
    };
  });
  
  async function loadChatHistory() {
    try {
      const history = await GetChatHistory(projectId, endpointId);
      if (history && history.messages) {
        // Initialize the real-time store with the loaded messages
        setupRealtimeConnection(history.messages, history.sessionId);
      } else {
        // Initialize with empty messages
        setupRealtimeConnection([], null);
      }
    } catch (error) {
      console.warn("Could not load chat history:", error);
      // Initialize with empty messages even if loading fails
      setupRealtimeConnection([], null);
    }
  }
  
  function setupRealtimeConnection(initialMessages = [], initialSessionId = null) {
    // Disconnect existing connection if any
    if (unsubscribeRealtime) {
      unsubscribeRealtime();
    }
    
    // Connect to real-time updates
    unsubscribeRealtime = connectChatbotSession(
      projectId, 
      endpointId, 
      initialMessages, 
      initialSessionId
    );
    
    console.log(`Set up real-time connection for chatbot session ${projectId}_${endpointId}`);
  }
  
  async function handleSendMessage(messageText) {
    if (!messageText.trim() || loading) return;
    
    loading = true;
    
    try {
      // Get current session ID from real-time store
      const currentSessionId = $realtimeSessionId;
      
      // Prepare context data with endpoint-specific information
      let enrichedContextData = { ...contextData };
      
      // Inject highlight context for highlight_ordering endpoint
      if (endpointId === "highlight_ordering") {
        try {
          // Load current highlights and order in parallel
          const [highlightsData, currentOrder] = await Promise.all([
            GetProjectHighlights(projectId),
            GetProjectHighlightOrderWithTitles(projectId)
          ]);
          
          // Build highlight map (ID -> text) like in buildReorderingPrompt
          const highlightMap = {};
          if (highlightsData && highlightsData.length > 0) {
            for (const videoHighlights of highlightsData) {
              if (videoHighlights.highlights) {
                for (const highlight of videoHighlights.highlights) {
                  highlightMap[highlight.id] = highlight.text;
                }
              }
            }
          }
          
          // Add highlights context to the data
          enrichedContextData.highlights = {
            highlightMap,
            currentOrder,
            totalHighlights: Object.keys(highlightMap).length
          };
          
          console.log(`Injected ${Object.keys(highlightMap).length} highlights into context for reordering`);
        } catch (error) {
          console.warn("Failed to load highlights context:", error);
          // Continue without context if loading fails
        }
      }
      
      // Send message to backend (real-time events will handle message updates)
      const response = await SendChatMessage({
        projectId,
        endpointId,
        message: messageText,
        sessionId: currentSessionId,
        contextData: enrichedContextData,
        model: selectedModel === "custom" ? customModelValue : selectedModel
      });
      
      // Update session ID if we got a new one
      if (response.sessionId && response.sessionId !== currentSessionId) {
        updateChatbotSessionId(projectId, endpointId, response.sessionId);
      }
      
      // If there was an error, add error message locally
      if (!response.success && response.error) {
        const errorMessage = {
          id: `error_${Date.now()}`,
          role: "error",
          content: response.error,
          timestamp: new Date().toISOString()
        };
        
        // Add error message to real-time store
        addChatbotMessage(projectId, endpointId, errorMessage);
        toast.error("Failed to send message");
      }
      
    } catch (error) {
      console.error("Failed to send message:", error);
      
      // Add error message to chat
      const errorMessage = {
        id: `error_${Date.now()}`,
        role: "error",
        content: "Failed to process message. Please try again.",
        timestamp: new Date().toISOString()
      };
      
      // Add error message to real-time store
      addChatbotMessage(projectId, endpointId, errorMessage);
      toast.error("Failed to send message");
    } finally {
      loading = false;
    }
  }
  
  async function handleClearHistory() {
    try {
      await ClearChatHistory(projectId, endpointId);
      // Real-time event will handle clearing the messages
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

<div class="relative flex flex-col h-full">
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
          disabled={loading || $realtimeMessages.length === 0}
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
  <div class="flex-1 overflow-y-auto px-6 scrollbar-thin">
    <MessageList 
      messages={$realtimeMessages} 
      {loading} 
      {config} 
      {endpointId}
      {projectId}
      onSendMessage={handleSendMessage}
    />
  </div>
  
  <!-- Message Input -->
  <div class="border-t border-border bg-background px-6 py-4">
    <MessageInput
      onSendMessage={handleSendMessage}
      {loading}
      placeholder="Ask me about your {config.title?.toLowerCase() || 'project'}..."
    />
  </div>
</div>