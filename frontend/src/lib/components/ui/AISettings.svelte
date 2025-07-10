<script>
  import * as Collapsible from "$lib/components/ui/collapsible/index.js";
  import { Sparkles, Settings, Play, RotateCcw } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import AIModelSelector from "./AIModelSelector.svelte";
  import AIPromptEditor from "./AIPromptEditor.svelte";

  let {
    open = $bindable(false),
    selectedModel = $bindable("google/gemini-2.5-flash-preview-05-20"),
    customModelValue = $bindable(""),
    customPrompt = $bindable(""),
    defaultPrompt = "",
    title = "AI Settings",
    modelLabel = "AI Model",
    modelDescription = "Choose the AI model for processing. Different models have varying strengths in content analysis.",
    promptLabel = "AI Instructions",
    promptDescription = "Customize how AI processes your content.",
    promptPlaceholder = "AI instructions...",
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
    showResetButton = true,
    loading = false,
    hasRun = false,
    onRun = () => {},
    settingsContent
  } = $props();

  // Get display name for current model
  function getModelDisplayName(model, customValue) {
    if (model === "custom") {
      return customValue || "Custom Model";
    }
    const foundModel = availableModels.find(m => m.value === model);
    return foundModel ? foundModel.label : model;
  }
</script>

<Collapsible.Root bind:open>
  <div class="space-y-3">
    <!-- Main control bar - no hover effects -->
    <div class="flex items-center justify-between p-3 rounded-md border border-input bg-background">
      <div class="flex items-center gap-3">
        <Sparkles class="w-4 h-4" />
        <div class="flex flex-col">
          <span class="font-medium">{title}</span>
          <span class="text-sm text-muted-foreground">
            {getModelDisplayName(selectedModel, customModelValue)}
          </span>
        </div>
      </div>
      
      <div class="flex items-center gap-2">
        <!-- Main Run/Rerun button -->
        <Button
          onclick={onRun}
          disabled={loading}
          variant="default"
          size="sm"
          class="gap-2"
        >
          {#if loading}
            <div class="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
          {:else if hasRun}
            <RotateCcw class="w-4 h-4" />
          {:else}
            <Play class="w-4 h-4" />
          {/if}
          {hasRun ? "Rerun" : "Run"}
        </Button>
        
        <!-- Settings trigger - only this opens the collapsible -->
        <Collapsible.Trigger class="p-2 rounded-md hover:bg-accent hover:text-accent-foreground">
          <Settings class="w-4 h-4" />
        </Collapsible.Trigger>
      </div>
    </div>
    
    <!-- Settings content that expands below the entire component -->
    <Collapsible.Content class="space-y-4 p-4 rounded-md border border-input bg-card">
      <!-- Model Selection -->
      <AIModelSelector
        bind:selectedModel
        bind:customModelValue
        label={modelLabel}
        description={modelDescription}
        {availableModels}
      />

      <!-- Custom Prompt Input -->
      <AIPromptEditor
        bind:customPrompt
        {defaultPrompt}
        label={promptLabel}
        description={promptDescription}
        placeholder={promptPlaceholder}
        {showResetButton}
      />

      <!-- Additional content -->
      {#if settingsContent}
        {@render settingsContent()}
      {/if}
    </Collapsible.Content>
  </div>
</Collapsible.Root>