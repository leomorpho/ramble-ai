<script>
  import * as Collapsible from "$lib/components/ui/collapsible/index.js";
  import { Sparkles } from "@lucide/svelte";
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
        value: "anthropic/claude-3-haiku-20240307",
        label: "Claude 3 Haiku (Fast)",
      },
      { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
      { value: "mistralai/mistral-nemo", label: "Mistral Nemo" },
      { value: "custom", label: "Custom Model" },
    ],
    showResetButton = true,
    children
  } = $props();
</script>

<Collapsible.Root bind:open>
  <Collapsible.Trigger class="flex w-full justify-between items-center p-3 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground">
    <span class="flex items-center gap-2">
      <Sparkles class="w-4 h-4" />
      {title}
    </span>
    <svg
      class="w-4 h-4 transition-transform duration-200"
      class:rotate-180={open}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
    </svg>
  </Collapsible.Trigger>
  <Collapsible.Content class="space-y-4 mt-4">
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
    {#if children}
      {@render children()}
    {/if}
  </Collapsible.Content>
</Collapsible.Root>