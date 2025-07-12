<script>
  import CustomSheet from "$lib/components/CustomSheet.svelte";
  import { Button } from "$lib/components/ui/button";
  import { Brain, RefreshCw, Trash2, Settings } from "@lucide/svelte";
  import ChatInterface from "./ChatInterface.svelte";
  import { CHATBOT_ENDPOINTS, CHATBOT_POSITIONS, ENDPOINT_CONFIGS } from "$lib/constants/chatbot.js";
  import { getChatbotMessages } from "$lib/stores/chatbotRealtime.js";
  
  let {
    endpointId = CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING,
    projectId,
    contextData = {}, // highlight data, etc.
    position = CHATBOT_POSITIONS.FLOATING,
    open = $bindable(false),
    size = "default", // "sm", "default", "lg"
    className = "",
    buttonText = "AI Assistant",
    side = "right" // sheet side: "top", "right", "bottom", "left"
  } = $props();
  
  // Get configuration for current endpoint
  let config = $derived(ENDPOINT_CONFIGS[endpointId] || ENDPOINT_CONFIGS[CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING]);
  
  // Header actions state
  let chatInterface = $state();
  
  // Access real-time state for button states
  let realtimeMessages = $derived(getChatbotMessages(projectId, endpointId));
  
  // Size configurations
  const sizeConfigs = {
    sm: { button: "h-12 w-12", sheet: "w-[80vw] max-w-md" },
    default: { button: "h-14 w-14", sheet: "w-[90vw] max-w-lg" },
    lg: { button: "h-16 w-16", sheet: "w-[95vw] max-w-xl" }
  };
  
  let sizeConfig = $derived(sizeConfigs[size] || sizeConfigs.default);
  
  // Action handlers (delegated to ChatInterface)
  function handleRefresh() {
    if (chatInterface?.handleRefresh) {
      chatInterface.handleRefresh();
    }
  }
  
  function handleClearHistory() {
    if (chatInterface?.handleClearHistory) {
      chatInterface.handleClearHistory();
    }
  }
  
  function toggleSettings() {
    if (chatInterface?.toggleSettings) {
      chatInterface.toggleSettings();
    }
  }
</script>

{#if position === CHATBOT_POSITIONS.FLOATING}
  <!-- Floating brain button -->
  <div class="fixed bottom-6 right-6 z-50 {className}">
    <Button 
      class="{sizeConfig.button} rounded-full shadow-lg bg-primary hover:bg-primary/90 text-primary-foreground"
      aria-label="Open {config.title}"
      onclick={() => open = true}
    >
      <Brain class="w-6 h-6" />
    </Button>
  </div>
{:else if position === CHATBOT_POSITIONS.INLINE}
  <!-- Inline version for replacing AISettings -->
  <Button 
    class="inline-flex items-center gap-2 justify-center rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2 {className}"
    onclick={() => open = true}
  >
    <Brain class="w-4 h-4" />
    {buttonText}
  </Button>
{/if}

<!-- CustomSheet for all positions -->
<CustomSheet 
  bind:open 
  title={config.title}
  description={config.description}
  icon={config.icon ? () => config.icon : undefined}
>
  {#snippet headerActions()}
    <Button
      variant="ghost"
      size="sm"
      onclick={handleRefresh}
      disabled={chatInterface?.loading || false}
      aria-label="Refresh chat"
      class="h-8 w-8 p-0"
    >
      <RefreshCw class="w-4 h-4" />
    </Button>
    
    <Button
      variant="ghost"
      size="sm"
      onclick={handleClearHistory}
      disabled={(chatInterface?.loading || false) || ($realtimeMessages?.length || 0) === 0}
      aria-label="Clear history"
      class="h-8 w-8 p-0"
    >
      <Trash2 class="w-4 h-4" />
    </Button>
    
    <Button
      variant="ghost"
      size="sm"
      onclick={toggleSettings}
      aria-label="Settings"
      class="h-8 w-8 p-0"
    >
      <Settings class="w-4 h-4" />
    </Button>
  {/snippet}
  
  {#snippet children()}
    <ChatInterface 
      bind:this={chatInterface}
      {endpointId} 
      {projectId} 
      {contextData} 
      hideHeader={true}
    />
  {/snippet}
</CustomSheet>