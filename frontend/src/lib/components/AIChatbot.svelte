<script>
  import { onMount, tick } from "svelte";
  import * as Collapsible from "$lib/components/ui/collapsible/index.js";
  import { Sparkles, Settings, Send, RotateCcw, MessageSquare, Edit3, CheckCircle, AlertCircle } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import AutoResizeTextarea from "./AutoResizeTextarea.svelte";
  import AIModelSelector from "./AIModelSelector.svelte";
  import { SendChatMessage } from "$lib/wailsjs/go/main/App";

  let {
    open = $bindable(false),
    selectedModel = $bindable("anthropic/claude-sonnet-4"),
    customModelValue = $bindable(""),
    title = "AI Assistant",
    defaultPrompt = "",
    availableModels = [
      { value: "anthropic/claude-sonnet-4", label: "Claude Sonnet 4 (Latest)" },
      { value: "google/gemini-2.0-flash-001", label: "Gemini 2.0 Flash" },
      {
        value: "google/gemini-2.5-flash-preview-05-20",
        label: "Gemini 2.5 Flash Preview",
      },
      {
        value: "deepseek/deepseek-chat-v3-0324:free",
        label: "DeepSeek Chat v3 (Free)",
      },
      { value: "anthropic/claude-3.7-sonnet", label: "Claude 3.7 Sonnet" },
      {
        value: "anthropic/claude-3.5-haiku-20241022",
        label: "Claude 3.5 Haiku (Fast)",
      },
      { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
      { value: "mistralai/mistral-nemo", label: "Mistral Nemo" },
      { value: "custom", label: "Custom Model" },
    ],
    loading = false,
    onSendMessage = async (message) => {},
    chatHistory = $bindable([]),
    modelLabel = "AI Model",
    modelDescription = "Choose the AI model for processing",
    showModelSettings = true,
    settingsContent,
    // New props for reorder mode
    mode = "chat", // "chat" or "reorder"
    projectId = null,
    endpointId = "general_chat"
  } = $props();

  // Local state
  let currentMessage = $state("");
  let settingsOpen = $state(false);
  let scrollAreaElement = $state(null);
  let textareaElement = $state(null);
  let editingPrompt = $state(false);
  let editablePrompt = $state("");

  // Initialize the chat with the default prompt if there's no history
  $effect(() => {
    if (chatHistory.length === 0 && defaultPrompt) {
      editablePrompt = defaultPrompt;
      editingPrompt = true;
    }
  });

  // Auto-scroll to bottom when new messages are added
  $effect(() => {
    if (chatHistory.length > 0 && scrollAreaElement) {
      tick().then(() => {
        const scrollContainer = scrollAreaElement.querySelector('[data-radix-scroll-area-viewport]');
        if (scrollContainer) {
          scrollContainer.scrollTop = scrollContainer.scrollHeight;
        }
      });
    }
  });

  // Get display name for current model
  function getModelDisplayName(model, customValue) {
    if (model === "custom") {
      return customValue || "Custom Model";
    }
    const foundModel = availableModels.find(m => m.value === model);
    return foundModel ? foundModel.label : model;
  }

  async function handleSendMessage() {
    if (!currentMessage.trim() || loading) return;

    const message = currentMessage.trim();
    currentMessage = "";
    loading = true;

    // Add user message to history
    chatHistory = [...chatHistory, {
      type: "user",
      content: message,
      timestamp: new Date().toISOString()
    }];

    try {
      if (mode === "reorder" && projectId) {
        // Use new chatbot service with function calling
        const modelToUse = selectedModel === "custom" ? customModelValue : selectedModel;
        
        const chatRequest = {
          projectId: projectId,
          endpointId: endpointId,
          message: message,
          sessionId: `session_${projectId}_${endpointId}`,
          contextData: {},
          model: modelToUse,
          enableFunctionCalls: true,
          mode: "reorder"
        };
        
        const response = await SendChatMessage(chatRequest);
        
        if (response.success) {
          // Add AI response to chat
          chatHistory = [...chatHistory, {
            type: "assistant",
            content: response.message,
            timestamp: new Date().toISOString(),
            functionResults: response.functionResults || []
          }];
          
          // If there were function results, show them
          if (response.functionResults && response.functionResults.length > 0) {
            for (const result of response.functionResults) {
              chatHistory = [...chatHistory, {
                type: "function_result",
                content: result.message || `Executed ${result.functionName}`,
                timestamp: new Date().toISOString(),
                functionName: result.functionName,
                success: result.success,
                result: result.result
              }];
            }
          }
        } else {
          throw new Error(response.error || "Chat request failed");
        }
      } else {
        // Use legacy callback for regular chat mode
        await onSendMessage(message);
      }
    } catch (error) {
      console.error("Failed to send message:", error);
      // Add error message to chat
      chatHistory = [...chatHistory, {
        type: "error",
        content: "Failed to process message. Please try again.",
        timestamp: new Date().toISOString()
      }];
    } finally {
      loading = false;
    }
  }

  function handleKeydown(event) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    }
  }

  function handleEditPrompt() {
    editingPrompt = true;
    editablePrompt = defaultPrompt;
  }

  function handleSendInitialPrompt() {
    if (editablePrompt.trim()) {
      currentMessage = editablePrompt.trim();
      editingPrompt = false;
      handleSendMessage();
    }
  }

  function handleCancelEdit() {
    editingPrompt = false;
    editablePrompt = "";
  }

  // Focus textarea when component opens
  $effect(() => {
    if (open && textareaElement) {
      tick().then(() => {
        textareaElement.focus();
      });
    }
  });
</script>

<Collapsible.Root bind:open>
  <div class="space-y-3">
    <!-- Main control bar -->
    <div class="flex flex-wrap items-center justify-between gap-2 p-3 rounded-md border border-input bg-background">
      <div class="flex items-center gap-3 min-w-0">
        <MessageSquare class="w-4 h-4 flex-shrink-0" />
        <div class="flex flex-col min-w-0">
          <span class="font-medium">{title}</span>
          <span class="text-sm text-muted-foreground truncate">
            {#if mode === \"reorder\"}\n              Reorder Mode - {getModelDisplayName(selectedModel, customModelValue)}\n            {:else}\n              {getModelDisplayName(selectedModel, customModelValue)}\n            {/if}
          </span>
        </div>
      </div>
      
      <div class="flex items-center gap-2 flex-shrink-0">
        <!-- Chat toggle button -->
        <Collapsible.Trigger>
          <Button variant="default" size="sm" class="gap-2">
            <Sparkles class="w-4 h-4" />
            {open ? "Close Chat" : "Open Chat"}
          </Button>
        </Collapsible.Trigger>
        
        <!-- Settings button -->
        {#if showModelSettings}
          <Button
            variant="outline"
            size="sm"
            onclick={() => settingsOpen = !settingsOpen}
            class="p-2"
          >
            <Settings class="w-4 h-4" />
          </Button>
        {/if}
      </div>
    </div>
    
    <!-- Chat interface -->
    <Collapsible.Content class="space-y-4">
      <div class="border border-input rounded-md bg-card">
        <!-- Settings panel -->
        {#if settingsOpen && showModelSettings}
          <div class="p-4 border-b border-border space-y-4">
            <AIModelSelector
              bind:selectedModel
              bind:customModelValue
              label={modelLabel}
              description={modelDescription}
              {availableModels}
            />
            
            <!-- Additional settings content -->
            {#if settingsContent}
              {@render settingsContent()}
            {/if}
          </div>
        {/if}

        <!-- Chat messages area -->
        <div class="h-96 flex flex-col">
          <ScrollArea bind:this={scrollAreaElement} class="flex-1 p-4">
            <div class="space-y-4">
              {#if editingPrompt}
                <!-- Initial prompt editing -->
                <div class="bg-secondary/50 rounded-lg p-4 space-y-3">
                  <div class="flex items-center gap-2">
                    <Edit3 class="w-4 h-4" />
                    <span class="font-medium text-sm">Initial Instructions</span>
                  </div>
                  <AutoResizeTextarea
                    bind:value={editablePrompt}
                    placeholder="Enter your instructions for the AI..."
                    class="min-h-24"
                    onkeydown={(e) => {
                      if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
                        e.preventDefault();
                        handleSendInitialPrompt();
                      }
                    }}
                  />
                  <div class="flex gap-2">
                    <Button size="sm" onclick={handleSendInitialPrompt} disabled={!editablePrompt.trim()}>
                      <Send class="w-4 h-4 mr-2" />
                      Send Instructions
                    </Button>
                    <Button size="sm" variant="outline" onclick={handleCancelEdit}>
                      Cancel
                    </Button>
                  </div>
                </div>
              {:else if chatHistory.length === 0}
                <!-- Empty state -->
                <div class="text-center text-muted-foreground py-8">
                  <MessageSquare class="w-8 h-8 mx-auto mb-2 opacity-50" />
                  <p>Start a conversation with the AI assistant</p>
                  {#if defaultPrompt}
                    <Button size="sm" variant="outline" class="mt-2" onclick={handleEditPrompt}>
                      <Edit3 class="w-4 h-4 mr-2" />
                      Use Default Instructions
                    </Button>
                  {/if}
                </div>
              {:else}
                <!-- Chat messages -->
                {#each chatHistory as message (message.timestamp)}
                  <div class="space-y-2">
                    {#if message.type === "user"}
                      <!-- User message -->
                      <div class="flex justify-end">
                        <div class="max-w-[80%] bg-primary text-primary-foreground rounded-lg px-4 py-2">
                          <div class="whitespace-pre-wrap text-sm">{message.content}</div>
                        </div>
                      </div>
                    {:else if message.type === "assistant"}
                      <!-- AI message -->
                      <div class="flex justify-start">
                        <div class="max-w-[80%] bg-secondary rounded-lg px-4 py-2">
                          <div class="whitespace-pre-wrap text-sm">{message.content}</div>
                        </div>
                      </div>
                    {:else if message.type === "function_result"}
                      <!-- Function execution result -->
                      <div class="flex justify-center">
                        <div class="bg-card border rounded-lg px-4 py-2 max-w-[90%]">
                          <div class="flex items-center gap-2 mb-2">
                            {#if message.success}
                              <CheckCircle class="w-4 h-4 text-green-600" />
                            {:else}
                              <AlertCircle class="w-4 h-4 text-red-600" />
                            {/if}
                            <span class="text-sm font-medium">{message.functionName}</span>
                          </div>
                          <div class="text-sm text-muted-foreground">{message.content}</div>
                          {#if message.result && typeof message.result === 'object'}
                            <div class="mt-2 text-xs bg-muted/50 rounded p-2">
                              {#if message.result.reason}
                                <div><strong>Reason:</strong> {message.result.reason}</div>
                              {/if}
                              {#if message.result.count}
                                <div><strong>Items affected:</strong> {message.result.count}</div>
                              {/if}
                            </div>
                          {/if}
                        </div>
                      </div>
                    {:else if message.type === "error"}
                      <!-- Error message -->
                      <div class="flex justify-center">
                        <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg px-4 py-2">
                          <div class="text-sm">{message.content}</div>
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}

                <!-- Loading indicator -->
                {#if loading}
                  <div class="flex justify-start">
                    <div class="bg-secondary rounded-lg px-4 py-2">
                      <div class="flex items-center gap-2">
                        <div class="w-2 h-2 bg-current rounded-full animate-pulse"></div>
                        <div class="w-2 h-2 bg-current rounded-full animate-pulse" style="animation-delay: 0.2s"></div>
                        <div class="w-2 h-2 bg-current rounded-full animate-pulse" style="animation-delay: 0.4s"></div>
                      </div>
                    </div>
                  </div>
                {/if}
              {/if}
            </div>
          </ScrollArea>

          <!-- Message input -->
          {#if !editingPrompt}
            <div class="p-4 border-t border-border">
              <div class="flex gap-2">
                <AutoResizeTextarea
                  bind:this={textareaElement}
                  bind:value={currentMessage}
                  placeholder="Type your message... (Enter to send, Shift+Enter for new line)"
                  class="flex-1 min-h-[40px] max-h-32"
                  onkeydown={handleKeydown}
                  disabled={loading}
                />
                <Button
                  onclick={handleSendMessage}
                  disabled={!currentMessage.trim() || loading}
                  size="sm"
                  class="self-end"
                >
                  {#if loading}
                    <div class="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
                  {:else}
                    <Send class="w-4 h-4" />
                  {/if}
                </Button>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </Collapsible.Content>
  </div>
</Collapsible.Root>