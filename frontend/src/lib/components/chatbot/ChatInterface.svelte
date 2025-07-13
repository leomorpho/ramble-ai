<script>
  import { onMount, onDestroy, tick } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Settings, RefreshCw, Trash2 } from "@lucide/svelte";
  import MessageList from "./MessageList.svelte";
  import MessageInput from "./MessageInput.svelte";
  import ChatSettings from "./ChatSettings.svelte";
  import { ENDPOINT_CONFIGS, AVAILABLE_MODELS } from "$lib/constants/chatbot.js";
  import { SendChatMessage, GetChatHistory, ClearChatHistory, SaveChatModelSelection } from "$lib/wailsjs/go/main/App.js";
  import { toast } from "svelte-sonner";
  import { 
    connectChatbotSession, 
    getChatbotMessages, 
    getChatbotSessionId,
    clearChatbotMessages,
    updateChatbotSessionId,
    addChatbotMessage,
    getChatbotProgress
  } from "$lib/stores/chatbotRealtime.js";
  
  let {
    endpointId,
    projectId,
    contextData = {},
    title = "AI Assistant",
    description = "Chat with AI about your project",
    icon = "🤖",
    hideHeader = false
  } = $props();
  
  // Get endpoint configuration
  let config = $derived(ENDPOINT_CONFIGS[endpointId] || {});
  
  // Component state
  let loading = $state(false);
  let settingsOpen = $state(false);
  let selectedModel = $state(AVAILABLE_MODELS[0].value);
  let customModelValue = $state("");
  let messagesContainer;
  let hasLoadedHistory = $state(false);
  
  // Set default model from config when it becomes available
  $effect(() => {
    if (config.defaultModel && !hasLoadedHistory) {
      selectedModel = config.defaultModel;
    }
  });
  
  // Real-time stores
  let realtimeMessages = $derived(getChatbotMessages(projectId, endpointId));
  let realtimeSessionId = $derived(getChatbotSessionId(projectId, endpointId));
  let progressMessage = $derived(getChatbotProgress(projectId, endpointId));
  let unsubscribeRealtime = null;
  
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
      // Reset the flag when endpoint changes so we can load the saved model
      hasLoadedHistory = false;
      loadChatHistory(); // This already calls setupRealtimeConnection
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
      if (history) {
        // Load saved model if available - only if we haven't loaded history before
        if (history.selectedModel) {
          // Check if it's a custom model
          if (!AVAILABLE_MODELS.find(m => m.value === history.selectedModel)) {
            selectedModel = "custom";
            customModelValue = history.selectedModel;
          } else {
            selectedModel = history.selectedModel;
          }
        }
        
        // Initialize the real-time store with the loaded messages
        if (history.messages) {
          setupRealtimeConnection(history.messages, history.sessionId);
        } else {
          setupRealtimeConnection([], history.sessionId);
        }
        
        hasLoadedHistory = true;
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
      
      // Send message to backend - no need to inject context, backend handles it automatically
      const modelToSend = selectedModel === "custom" ? customModelValue : selectedModel;
      
      const response = await SendChatMessage({
        projectId,
        endpointId,
        message: messageText,
        sessionId: currentSessionId,
        contextData: { ...contextData }, // Pass through any provided context data
        model: modelToSend
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
  
  function toggleSettings() {
    settingsOpen = !settingsOpen;
  }
  
  // Save model selection when it changes
  async function saveModelSelection() {
    if (!projectId || !endpointId) return;
    
    try {
      const modelToSave = selectedModel === "custom" ? customModelValue : selectedModel;
      await SaveChatModelSelection(projectId, endpointId, modelToSave);
    } catch (error) {
      console.error('Failed to save model selection:', error);
    }
  }
  
  // Watch for model changes and save them
  $effect(() => {
    // Only save if we have loaded history (to avoid saving during initialization)
    if (hasLoadedHistory && selectedModel) {
      saveModelSelection();
    }
  });
  
  // Also watch for custom model value changes
  $effect(() => {
    if (hasLoadedHistory && selectedModel === "custom" && customModelValue) {
      saveModelSelection();
    }
  });
  
  // Scroll to bottom function
  function scrollToBottom() {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  }
  
  // Auto-scroll when messages change or component mounts
  $effect(() => {
    // Scroll when messages change
    if ($realtimeMessages.length > 0) {
      // Use setTimeout to ensure DOM has updated
      setTimeout(scrollToBottom, 0);
    }
  });
  
  // Scroll to bottom when chat is first opened (when component mounts)
  onMount(() => {
    setTimeout(scrollToBottom, 100); // Small delay to ensure everything is rendered
  });
  
  // Also scroll to bottom when hideHeader prop changes (indicating chat was opened)
  $effect(() => {
    if (!hideHeader) {
      // This means we're showing our own header (standalone mode)
      setTimeout(scrollToBottom, 100);
    }
  });
  
  // Expose methods and state for parent component binding
  export { handleRefresh, handleClearHistory, toggleSettings, loading };
</script>

<div class="relative flex flex-col h-full">
  <!-- Header (only shown when not hidden) -->
  {#if !hideHeader}
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
  {/if}
  
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
  <div bind:this={messagesContainer} class="flex-1 overflow-y-auto px-6 scrollbar-thin">
    <!-- Progress indicator -->
    {#if $progressMessage}
      <div class="sticky top-0 z-10 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg px-4 py-2 mb-4 mt-4">
        <div class="flex items-center gap-3">
          <div class="animate-spin w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full"></div>
          <span class="text-sm text-blue-700 dark:text-blue-300 font-medium">{$progressMessage}</span>
        </div>
      </div>
    {/if}
    
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