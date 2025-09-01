<script>
  import { onMount } from "svelte";
  import {
    ReorderHighlightsWithAIOptions,
    GetProjectAISettings,
    SaveProjectAISettings,
    GetUseRemoteAIBackend,
  } from "$lib/wailsjs/go/main/App";
  import { toast } from "svelte-sonner";
  import { Sparkles, Settings2, Play, Loader2 } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import * as Select from "$lib/components/ui/select/index.js";
  import { Separator } from "$lib/components/ui/separator";
  import CustomSheet from "$lib/components/CustomSheet.svelte";
  import { updateHighlightOrder } from "$lib/stores/projectHighlights.js";

  // Props
  let {
    open = $bindable(false),
    projectId,
    highlights = [],
    onApply = () => {},
  } = $props();

  // AI actions state
  let aiActionLoading = $state(false);
  let aiActionError = $state("");
  let customPrompt = $state("");
  let selectedModel = $state("anthropic/claude-sonnet-4");
  let useRemoteBackend = $state(false);
  let extraInfoOpen = $state(false);
  
  // AI action options
  let useCurrentOrder = $state(false);
  let keepAllHighlights = $state(true);
  let optimizeForEngagement = $state(false);
  let createSections = $state(true);
  let balanceLength = $state(false);
  let improveTransitions = $state(false);

  // Available AI models
  const availableModels = [
    { value: "anthropic/claude-sonnet-4", label: "Claude Sonnet 4 (Latest)" },
    { value: "google/gemini-2.0-flash-001", label: "Gemini 2.0 Flash" },
    { value: "anthropic/claude-3.5-haiku-20241022", label: "Claude 3.5 Haiku (Fast)" },
    { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
    { value: "mistralai/mistral-nemo", label: "Mistral Nemo" },
  ];

  // Default prompt
  const defaultPrompt = `You are an expert video editor focused on creating well-structured, engaging content. Your goal is to organize these highlight segments into a balanced, coherent video with natural pacing and flow.`;

  // Derived value for model selection display
  const selectedModelDisplay = $derived(
    availableModels.find((m) => m.value === selectedModel)?.label ?? "Select a model"
  );

  // Initialize settings when sheet opens
  $effect(() => {
    if (open && projectId) {
      loadAISettings();
      checkRemoteBackend();
    }
  });

  // Check if using remote AI backend
  async function checkRemoteBackend() {
    try {
      useRemoteBackend = await GetUseRemoteAIBackend();
    } catch (error) {
      console.error("Failed to check remote backend status:", error);
    }
  }

  // Load AI settings
  async function loadAISettings() {
    try {
      const settings = await GetProjectAISettings(projectId);
      selectedModel = settings.aiModel || "anthropic/claude-sonnet-4";
      customPrompt = settings.aiPrompt || defaultPrompt;
    } catch (error) {
      console.error("Failed to load AI settings:", error);
      customPrompt = defaultPrompt;
    }
  }

  // Save AI settings
  async function saveAISettings() {
    try {
      await SaveProjectAISettings(projectId, {
        aiModel: selectedModel,
        aiPrompt: customPrompt,
      });
    } catch (error) {
      console.error("Failed to save AI settings:", error);
    }
  }

  // Build AI action options based on selected checkboxes
  function buildActionOptions() {
    return {
      useCurrentOrder,
      keepAllHighlights,
      optimizeForEngagement,
      createSections,
      balanceLength,
      improveTransitions,
    };
  }

  // Build prompt preview based on selected options - this is just for display
  function buildPromptPreview() {
    let prompt = customPrompt || defaultPrompt;
    
    const instructions = [];
    
    if (useCurrentOrder) {
      instructions.push("- Use the current highlight order as a starting point for organization");
    }
    
    if (keepAllHighlights) {
      instructions.push("- Keep all highlights in the reorder, do not drop any highlights");
    } else {
      instructions.push("- Focus on QUALITY over quantity - remove repetitive or low-value highlights");
      instructions.push("- Remove repetitive highlights that cover the same points or topics");
      instructions.push("- Keep only the best version of similar highlights for a smoother script");
      instructions.push("- Prioritize highlights that advance the narrative or provide unique value");
    }
    
    if (optimizeForEngagement) {
      instructions.push("- Optimize the sequence for maximum viewer engagement and retention");
    }
    
    if (createSections) {
      instructions.push("- Create logical sections with clear titles to organize the content");
    }
    
    if (balanceLength) {
      instructions.push("- Balance section lengths to ensure no single section is too long or short");
    }
    
    if (improveTransitions) {
      instructions.push("- Focus on smooth transitions between different topics and sections");
    }
    
    if (instructions.length > 0) {
      prompt += "\n\nAdditional instructions:\n" + instructions.join("\n");
    }
    
    return prompt;
  }

  // Execute AI action
  async function executeAIAction() {
    if (!projectId || highlights.length === 0) {
      toast.error("No highlights available for AI processing");
      return;
    }

    aiActionLoading = true;
    aiActionError = "";

    try {
      // Use the complete prompt that includes all the instructions based on selected options
      const finalPrompt = buildPromptPreview();
      const options = buildActionOptions();
      
      // Save settings first
      await saveAISettings();
      
      // Execute AI reordering with options
      const reorderedHighlights = await ReorderHighlightsWithAIOptions(projectId, finalPrompt, options);
      
      if (reorderedHighlights && reorderedHighlights.length > 0) {
        // Apply the reordered highlights
        await updateHighlightOrder(reorderedHighlights);
        
        toast.success("AI actions applied successfully!");
        onApply(reorderedHighlights);
        open = false;
      } else {
        throw new Error("AI returned empty results");
      }
    } catch (error) {
      console.error("AI action failed:", error);
      aiActionError = error.message || "Failed to execute AI action";
      toast.error("AI action failed: " + aiActionError);
    } finally {
      aiActionLoading = false;
    }
  }

  // Reset options to defaults
  function resetOptions() {
    useCurrentOrder = false;
    keepAllHighlights = true;
    optimizeForEngagement = false;
    createSections = true;
    balanceLength = false;
    improveTransitions = false;
    customPrompt = defaultPrompt;
    selectedModel = "anthropic/claude-sonnet-4";
  }
</script>

<CustomSheet bind:open title="AI Reorder" class="w-full max-w-2xl">
  <div class="h-full overflow-y-auto">
    <div class="p-6 pb-8 space-y-6">

    <!-- Action Options -->
    <div class="border rounded p-4 space-y-4">
      <div class="flex items-start justify-between">
        <div>
          <h3 class="font-medium">Action Options</h3>
          <p class="text-sm text-muted-foreground">Choose what the AI should focus on.</p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onclick={resetOptions}
          disabled={aiActionLoading}
          class="text-xs"
        >
          <Settings2 class="w-3 h-3 mr-1" />
          Reset to Defaults
        </Button>
      </div>
      
      <!-- Use Current Order -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="use-current-order" 
            bind:checked={useCurrentOrder}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="use-current-order" class="text-sm font-medium">
            Use current highlights order as starting point
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          AI will consider your current arrangement when making improvements
        </p>
      </div>

      <!-- Keep All Highlights -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="keep-all-highlights" 
            bind:checked={keepAllHighlights}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="keep-all-highlights" class="text-sm font-medium">
            Keep all highlights in reorder, do not drop any
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          {#if keepAllHighlights}
            Ensures no highlights are removed during reorganization
          {:else}
            AI will remove repetitive or low-quality highlights for a smoother script
          {/if}
        </p>
      </div>

      <Separator />

      <!-- Optimize for Engagement -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="optimize-engagement" 
            bind:checked={optimizeForEngagement}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="optimize-engagement" class="text-sm font-medium">
            Optimize for viewer engagement
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          Arrange highlights to maximize viewer retention and interest
        </p>
      </div>

      <!-- Create Sections -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="create-sections" 
            bind:checked={createSections}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="create-sections" class="text-sm font-medium">
            Create logical sections with titles
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          Group related highlights into named sections
        </p>
      </div>

      <!-- Balance Length -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="balance-length" 
            bind:checked={balanceLength}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="balance-length" class="text-sm font-medium">
            Balance section lengths
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          Ensure no single section is too long or short
        </p>
      </div>

      <!-- Improve Transitions -->
      <div class="space-y-1">
        <div class="flex items-center space-x-2">
          <input 
            type="checkbox" 
            id="improve-transitions" 
            bind:checked={improveTransitions}
            class="w-4 h-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-2"
          />
          <Label for="improve-transitions" class="text-sm font-medium">
            Focus on smooth transitions
          </Label>
        </div>
        <p class="text-xs text-muted-foreground ml-6">
          Optimize flow between different topics and sections
        </p>
      </div>
    </div>

    <!-- AI Model Selection (hide when using remote backend) -->
    {#if !useRemoteBackend}
      <div class="border rounded p-4 space-y-3">
        <div>
          <h3 class="font-medium">AI Model</h3>
          <p class="text-sm text-muted-foreground">Choose the AI model to use.</p>
        </div>
        <Select.Root type="single" name="aiModel" bind:value={selectedModel}>
          <Select.Trigger class="w-full">
            {selectedModelDisplay}
          </Select.Trigger>
          <Select.Content>
            <Select.Group>
              <Select.Label>Available Models</Select.Label>
              {#each availableModels as model}
                <Select.Item value={model.value} label={model.label}>
                  {model.label}
                </Select.Item>
              {/each}
            </Select.Group>
          </Select.Content>
        </Select.Root>
      </div>
    {/if}

    <!-- Extra Info Section (Collapsible) -->
    <div class="border rounded p-4 space-y-3">
      <button
        type="button"
        onclick={() => extraInfoOpen = !extraInfoOpen}
        class="w-full flex items-center justify-between text-left hover:bg-accent/50 -m-1 p-1 rounded transition-colors"
      >
        <span class="font-medium text-sm">Extra Information</span>
        <svg
          class="w-4 h-4 transition-transform duration-200"
          class:rotate-180={extraInfoOpen}
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      
      {#if extraInfoOpen}
        <div class="space-y-4 pt-2">
          <!-- Custom Prompt -->
          <div class="space-y-2">
            <h4 class="text-sm font-medium">Custom Instructions</h4>
            <p class="text-xs text-muted-foreground">Provide additional context for the AI.</p>
            <Textarea
              bind:value={customPrompt}
              placeholder="Enter custom instructions for the AI..."
              class="min-h-[80px] text-sm"
            />
          </div>

          <!-- Prompt Preview -->
          <div class="space-y-2">
            <h4 class="text-sm font-medium">Generated Prompt Preview</h4>
            <p class="text-xs text-muted-foreground">The prompt that will be sent to the AI.</p>
            <div class="border rounded p-2 text-xs font-mono text-muted-foreground max-h-[150px] overflow-y-auto whitespace-pre-wrap">
              {buildPromptPreview()}
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Error Display -->
    {#if aiActionError}
      <div class="border border-destructive rounded p-3 text-destructive">
        <p class="font-medium">Error</p>
        <p class="text-sm">{aiActionError}</p>
      </div>
    {/if}

    <!-- Action Buttons -->
    <div class="flex gap-2 pt-4">
      <Button
        onclick={executeAIAction}
        disabled={aiActionLoading || highlights.length === 0}
        class="flex-1"
      >
        {#if aiActionLoading}
          <Loader2 class="w-4 h-4 mr-2 animate-spin" />
          Processing...
        {:else}
          <Sparkles class="w-4 h-4 mr-2" />
          Reorder now!
        {/if}
      </Button>
      
      <Button
        variant="outline"
        onclick={() => (open = false)}
        disabled={aiActionLoading}
      >
        Cancel
      </Button>
    </div>
    </div>
  </div>
</CustomSheet>