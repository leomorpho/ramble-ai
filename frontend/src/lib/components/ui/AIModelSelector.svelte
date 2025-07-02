<script>
  import { Label } from "$lib/components/ui/label";
  import * as Select from "$lib/components/ui/select/index.js";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index.js";

  let {
    selectedModel = $bindable("anthropic/claude-sonnet-4"),
    customModelValue = $bindable(""),
    label = "AI Model",
    description = "Choose the AI model for processing. Different models have varying strengths in content analysis.",
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
    ]
  } = $props();

  // Derived value for model selection display
  const selectedModelDisplay = $derived(
    availableModels.find((m) => m.value === selectedModel)?.label ?? "Select a model"
  );
</script>

<div class="space-y-2">
  <Label for="ai-model">{label}</Label>
  <Select.Root type="single" name="aiModel" bind:value={selectedModel}>
    <Select.Trigger class="w-full">
      {selectedModelDisplay}
    </Select.Trigger>
    <Select.Content>
      <ScrollArea class="h-72">
        <Select.Group>
          <Select.Label>Available Models</Select.Label>
          {#each availableModels as model (model.value)}
            <Select.Item value={model.value} label={model.label}>
              {model.label}
            </Select.Item>
          {/each}
        </Select.Group>
      </ScrollArea>
    </Select.Content>
  </Select.Root>

  {#if selectedModel === "custom"}
    <input
      type="text"
      bind:value={customModelValue}
      placeholder="Enter custom model (e.g., anthropic/claude-3-5-sonnet)"
      class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
    />
  {/if}

  <p class="text-xs text-muted-foreground">
    {description}
  </p>
</div>