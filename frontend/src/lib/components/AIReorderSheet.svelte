<script>
  import { onMount } from "svelte";
  import {
    ReorderHighlightsWithAI,
    GetProjectAISettings,
    SaveProjectAISettings,
    GetProjectAISuggestion,
  } from "$lib/wailsjs/go/main/App";
  import { toast } from "svelte-sonner";
  import { Play, Sparkles } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Label } from "$lib/components/ui/label";
  import * as Select from "$lib/components/ui/select/index.js";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index.js";
  import * as Collapsible from "$lib/components/ui/collapsible/index.js";
  import CustomSheet from "$lib/components/ui/CustomSheet.svelte";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import HighlightItem from "$lib/components/HighlightItem.svelte";
  import { updateHighlightOrder } from "$lib/stores/projectHighlights.js";

  // Props
  let {
    open = $bindable(false),
    projectId,
    highlights = [],
    onApply = () => {},
  } = $props();


  // AI reordering state
  let aiReorderLoading = $state(false);
  let aiReorderedHighlights = $state([]);
  let aiReorderError = $state("");
  let customPrompt = $state("");
  let selectedModel = $state("anthropic/claude-3-haiku-20240307");
  let hasCachedSuggestion = $state(false);
  let cachedSuggestionDate = $state(null);
  let cachedSuggestionModel = $state("");
  let showOriginalForm = $state(false);
  let originalHighlights = $state([]);

  // AI dialog independent state
  let aiDialogHighlights = $state([]);
  let aiSelectedHighlights = $state(new Set());
  let aiIsDragging = $state(false);
  let aiDraggedHighlights = $state([]);
  let aiDropPosition = $state(null);
  let aiDragStartPosition = $state(null);
  let aiIsDropping = $state(false);

  // AI dialog drag state
  let aiDragStartIndex = $state(-1);
  let aiDragOverIndex = $state(-1);

  // Available AI models
  const availableModels = [
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
  ];

  let customModelValue = $state("");

  // Collapsible state for AI instructions
  let instructionsOpen = $state(false);

  // Derived value for model selection display
  const selectedModelDisplay = $derived(
    availableModels.find((m) => m.value === selectedModel)?.label ?? "Select a model"
  );

  // Default YouTube expert prompt
  const defaultPrompt = `You are an expert YouTuber and content creator with millions of subscribers, known for creating highly engaging videos that maximize viewer retention and satisfaction. Your task is to reorder these video highlight segments to create the highest quality video possible.

Reorder these segments using your expertise in:
- Hook creation and audience retention
- Storytelling and narrative structure
- Pacing and rhythm for maximum engagement
- Building emotional connections with viewers
- Creating viral-worthy content flow
- Strategic placement of key moments

Feel free to completely restructure the order - move any segment to any position if it will improve video quality and viewer experience.`;

  // Initialize AI reordering when sheet opens
  $effect(() => {
    if (open && projectId && highlights.length > 0) {
      initializeAIReordering();
    }
  });


  // Initialize AI reordering
  async function initializeAIReordering() {
    // Reset state
    aiReorderLoading = false;
    aiReorderError = "";
    aiReorderedHighlights = [];
    hasCachedSuggestion = false;
    cachedSuggestionDate = null;
    cachedSuggestionModel = "";
    showOriginalForm = false;

    // Initialize AI dialog with independent copy of current highlights
    aiDialogHighlights = [...highlights];
    originalHighlights = [...highlights];
    aiSelectedHighlights.clear();
    aiIsDragging = false;
    aiDraggedHighlights = [];
    aiDropPosition = null;
    aiDragStartPosition = null;
    aiIsDropping = false;

    // Load project AI settings
    try {
      const aiSettings = await GetProjectAISettings(projectId);
      selectedModel = aiSettings.aiModel || "anthropic/claude-3-haiku-20240307";
      customPrompt = aiSettings.aiPrompt || defaultPrompt;

      // If using custom model, extract the value
      if (!availableModels.find((m) => m.value === selectedModel)) {
        customModelValue = selectedModel;
        selectedModel = "custom";
      }
    } catch (error) {
      console.error("Failed to load AI settings:", error);
      selectedModel = "anthropic/claude-3-haiku-20240307";
      customPrompt = defaultPrompt;
    }

    // Try to load cached AI suggestion
    try {
      const cachedSuggestion = await GetProjectAISuggestion(projectId);
      if (
        cachedSuggestion &&
        cachedSuggestion.order &&
        cachedSuggestion.order.length > 0
      ) {
        // Reorder AI dialog highlights based on cached suggestion
        const reorderedHighlights = [];
        const highlightsMap = new Map();

        // Create a map for quick lookup from current highlights
        for (const highlight of aiDialogHighlights) {
          highlightsMap.set(highlight.id, highlight);
        }

        // Reorder based on cached suggestion
        for (const id of cachedSuggestion.order) {
          const highlight = highlightsMap.get(id);
          if (highlight) {
            reorderedHighlights.push(highlight);
          }
        }

        // Add any highlights that weren't in the cached order
        for (const highlight of aiDialogHighlights) {
          if (!cachedSuggestion.order.includes(highlight.id)) {
            reorderedHighlights.push(highlight);
          }
        }

        if (reorderedHighlights.length > 0) {
          aiDialogHighlights = reorderedHighlights;
          aiReorderedHighlights = reorderedHighlights; // Set this to show the Apply button
          hasCachedSuggestion = true;
          cachedSuggestionDate = new Date(cachedSuggestion.createdAt);
          cachedSuggestionModel = cachedSuggestion.model || "";
          showOriginalForm = true; // Show reset option when cached suggestion is loaded

          // Preselect the last used model if available
          if (cachedSuggestion.model) {
            // Check if the cached model is in available models
            if (
              availableModels.find((m) => m.value === cachedSuggestion.model)
            ) {
              selectedModel = cachedSuggestion.model;
            } else {
              // It's a custom model
              customModelValue = cachedSuggestion.model;
              selectedModel = "custom";
            }
          }

          console.log(
            "Loaded cached AI suggestion from",
            cachedSuggestionDate,
            "with model:",
            cachedSuggestion.model
          );
        }
      }
    } catch (error) {
      console.log("No cached AI suggestion found:", error);
      // Not an error - just means no cached suggestion exists
      hasCachedSuggestion = false;
      cachedSuggestionDate = null;
      cachedSuggestionModel = "";
    }
  }

  // Start the actual AI reordering process
  async function startAIReordering() {
    aiReorderLoading = true;
    aiReorderError = "";

    try {
      // Save AI settings before processing
      const modelToSave =
        selectedModel === "custom" ? customModelValue : selectedModel;
      await SaveProjectAISettings(projectId, {
        aiModel: modelToSave,
        aiPrompt: customPrompt,
      });

      // Call the AI reordering API
      const reorderedIds = await ReorderHighlightsWithAI(
        projectId,
        customPrompt
      );

      // Reorder the AI dialog highlights based on AI suggestion
      const reorderedHighlights = [];
      const highlightsMap = new Map();

      // Create a map for quick lookup from current AI dialog highlights
      for (const highlight of aiDialogHighlights) {
        highlightsMap.set(highlight.id, highlight);
      }

      // Build reordered array
      for (const id of reorderedIds) {
        const highlight = highlightsMap.get(id);
        if (highlight) {
          reorderedHighlights.push(highlight);
        }
      }

      // Add any highlights that weren't in the AI response
      for (const highlight of aiDialogHighlights) {
        if (!reorderedIds.includes(highlight.id)) {
          reorderedHighlights.push(highlight);
        }
      }

      // Update AI dialog highlights with the reordered list
      aiDialogHighlights = reorderedHighlights;
      aiReorderedHighlights = reorderedHighlights;

      // Update cache state - we now have a fresh suggestion
      hasCachedSuggestion = true;
      cachedSuggestionDate = new Date();
      showOriginalForm = true; // Show reset option after AI generation

      toast.success("AI reordering completed!");
    } catch (error) {
      console.error("AI reordering error:", error);
      aiReorderError = error.message || "Failed to reorder highlights with AI";
      toast.error("Failed to reorder highlights with AI");
    } finally {
      aiReorderLoading = false;
    }
  }

  // Apply AI reordering to global state and close sheet
  async function applyAIReordering() {
    if (aiReorderedHighlights.length === 0) return;

    try {
      // Update via centralized store
      const success = await updateHighlightOrder(aiReorderedHighlights);

      if (success) {
        open = false; // Close the sheet
        toast.success("AI reordering applied successfully!");
        onApply(aiReorderedHighlights);
      }
    } catch (error) {
      console.error("Error applying AI reordering:", error);
      toast.error("Failed to apply AI reordering");
    }
  }

  // Reset to original highlights before AI generation
  function resetToOriginal() {
    aiDialogHighlights = [...originalHighlights];
    aiReorderedHighlights = [];
    showOriginalForm = false;
    hasCachedSuggestion = false;
    cachedSuggestionDate = null;
    cachedSuggestionModel = "";
    toast.success("Reset to original highlight order");
  }

  // AI dialog drag handlers
  function handleAIDragStart(event, index) {
    aiDragStartIndex = index;
    aiIsDragging = true;
    aiDraggedHighlights = [aiDialogHighlights[index].id];
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", index.toString());
  }

  function handleAIDragEnd() {
    aiDragStartIndex = -1;
    aiDragOverIndex = -1;
    aiIsDragging = false;
    aiDraggedHighlights = [];
    aiDropPosition = null;
  }

  function handleAIDragOver(event, targetIndex) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    aiDragOverIndex = targetIndex;
    aiDropPosition = targetIndex;
  }

  async function handleAIDrop(event, targetIndex) {
    event.preventDefault();

    if (aiDragStartIndex === -1 || aiDragStartIndex === targetIndex) {
      handleAIDragEnd();
      return;
    }

    // Reorder the AI dialog highlights array
    const newHighlights = [...aiDialogHighlights];
    const draggedItem = newHighlights[aiDragStartIndex];

    // Remove dragged item
    newHighlights.splice(aiDragStartIndex, 1);

    // Insert at new position
    const insertIndex =
      aiDragStartIndex < targetIndex ? targetIndex - 1 : targetIndex;
    newHighlights.splice(insertIndex, 0, draggedItem);

    // Update both AI dialog highlights and reordered highlights for sync
    aiDialogHighlights = newHighlights;
    aiReorderedHighlights = newHighlights;

    handleAIDragEnd();
  }

  // Handle reordering from the EtroVideoPlayer in AI preview mode
  async function handleAIVideoReorder(newHighlights) {
    // Update both AI dialog highlights and reordered highlights
    aiDialogHighlights = newHighlights;
    aiReorderedHighlights = newHighlights;
    return Promise.resolve(); // Return resolved promise for success
  }

  // AI dialog container drag functions
  function handleAIContainerDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }

  function handleAIContainerDrop(event) {
    event.preventDefault();
    // Handle drop at end of timeline
    if (aiDragStartIndex !== -1) {
      const targetIndex = aiDialogHighlights.length;
      handleAIDrop(event, targetIndex);
    }
  }

  function handleAIContainerDragLeave(event) {
    // Reset drag indicators when leaving container
    if (!event.currentTarget.contains(event.relatedTarget)) {
      aiDropPosition = null;
    }
  }
</script>

<CustomSheet 
  bind:open
  title="AI Reordered Highlights"
  description="Let AI suggest a new order for your highlights to maximize video quality and viewer engagement."
>
  {#snippet icon()}
    <Sparkles class="w-5 h-5" />
  {/snippet}
  {#snippet children()}
    <!-- Content -->
    <div class="p-6 space-y-6">
      <!-- AI Instructions Collapsible -->
      <Collapsible.Root bind:open={instructionsOpen}>
        <Collapsible.Trigger class="flex w-full justify-between items-center p-3 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground">
          <span class="flex items-center gap-2">
            <Sparkles class="w-4 h-4" />
            AI Instructions & Settings
          </span>
          <svg
            class="w-4 h-4 transition-transform duration-200"
            class:rotate-180={instructionsOpen}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </Collapsible.Trigger>
        <Collapsible.Content class="space-y-4 mt-4">
          <!-- Model Selection -->
          <div class="space-y-2">
            <Label for="ai-model">AI Model</Label>
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
              Choose the AI model for highlight reordering. Different models have
              varying strengths in content analysis and reasoning.
            </p>
          </div>

          <!-- Custom Prompt Input -->
          <div class="space-y-2">
            <Label for="custom-prompt">AI Instructions</Label>
            <Textarea
              id="custom-prompt"
              bind:value={customPrompt}
              placeholder="AI instructions for reordering highlights..."
              class="min-h-[120px] resize-none"
              rows="6"
            />
            <p class="text-xs text-muted-foreground">
              Modify the prompt above to customize how AI reorders your
              highlights. The default focuses on YouTube best practices for
              maximum engagement.
            </p>
          </div>

          {#if hasCachedSuggestion && cachedSuggestionDate}
            <div class="p-3 bg-secondary rounded-lg">
              <p class="text-sm text-muted-foreground">
                <strong>Cached AI Suggestion:</strong> Loaded from {cachedSuggestionDate.toLocaleString()}
                {#if cachedSuggestionModel}
                  <br /><strong>Model used:</strong>
                  {availableModels.find((m) => m.value === cachedSuggestionModel)
                    ?.label || cachedSuggestionModel}
                {/if}
              </p>
            </div>
          {/if}

          <div class="flex justify-between">
            <Button
              variant="outline"
              onclick={() => {
                customPrompt = defaultPrompt;
              }}
              disabled={customPrompt === defaultPrompt}
            >
              Reset to Default
            </Button>
            <div class="flex gap-2">
              {#if hasCachedSuggestion}
                <Button
                  variant="outline"
                  onclick={startAIReordering}
                  class="flex items-center gap-2"
                  disabled={aiReorderLoading}
                >
                  <Sparkles class="w-4 h-4" />
                  Re-run AI
                </Button>
              {/if}
              <Button
                onclick={startAIReordering}
                class="flex items-center gap-2"
                disabled={aiReorderLoading}
              >
                <Sparkles class="w-4 h-4" />
                {hasCachedSuggestion
                  ? "Update AI Suggestions"
                  : "Generate AI Suggestions"}
              </Button>
            </div>
          </div>
        </Collapsible.Content>
      </Collapsible.Root>

      <!-- Results Section -->
      <div class="space-y-6">
              {#if aiReorderLoading}
                <div class="p-8 text-center">
                  <div
                    class="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-4"
                  ></div>
                  <p class="text-lg font-medium">
                    AI is analyzing your highlights...
                  </p>
                  <p class="text-sm text-muted-foreground">
                    This may take a few moments
                  </p>
                </div>
              {:else if aiReorderError}
                <div class="p-6 text-center space-y-4">
                  <div
                    class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
                  >
                    <p class="font-medium">Error</p>
                    <p class="text-sm">{aiReorderError}</p>
                  </div>
                  <div class="flex justify-center gap-2">
                    <Button
                      variant="outline"
                      onclick={() => {
                        aiReorderError = "";
                        aiReorderedHighlights = [];
                        aiDialogHighlights = [...highlights];
                        instructionsOpen = true;
                      }}
                    >
                      Back to Instructions
                    </Button>
                  </div>
                </div>
              {:else if aiDialogHighlights.length > 0}
                <div class="space-y-6">
                  <!-- Preview Video Player -->
                  <div class="bg-card border rounded-lg p-4">
                    <h3 class="text-sm font-medium mb-3 flex items-center gap-2">
                      <Play class="w-4 h-4" />
                      {#if aiReorderedHighlights.length > 0}
                        Preview AI Arrangement
                      {:else}
                        Current Highlight Order
                      {/if}
                    </h3>
                    <EtroVideoPlayer 
                      highlights={aiReorderedHighlights.length > 0 ? aiReorderedHighlights : highlights} 
                      {projectId} 
                      enableEyeButton={false}
                      onReorder={handleAIVideoReorder}
                    />
                  </div>

                  <!-- AI Dialog Timeline -->
                  <div class="bg-muted/30 rounded-lg p-4">
                    <h3 class="text-sm font-medium mb-3">
                      AI Suggested Order (drag to reorder):
                    </h3>

                    <!-- Timeline-style highlight display -->
                    <div
                      class="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border"
                      role="application"
                      ondragover={(e) => handleAIContainerDragOver(e)}
                      ondrop={(e) => handleAIContainerDrop(e)}
                      ondragleave={handleAIContainerDragLeave}
                    >
                      {#if aiDialogHighlights.length === 0}
                        <div class="text-center py-4 text-muted-foreground">
                          <p class="text-sm">No highlights to display.</p>
                        </div>
                      {:else}
                        {#each aiDialogHighlights as highlight, index}
                          <HighlightItem
                            {highlight}
                            {index}
                            isSelected={aiSelectedHighlights.has(highlight.id)}
                            isDragging={aiIsDragging}
                            isBeingDragged={aiIsDragging &&
                              aiDraggedHighlights.includes(highlight.id) &&
                              aiDraggedHighlights[0] === highlight.id}
                            showDropIndicatorBefore={aiIsDragging &&
                              aiDropPosition === index}
                            onSelect={() => {}}
                            onDragStart={(e, h, i) => handleAIDragStart(e, i)}
                            onDragEnd={handleAIDragEnd}
                            onDragOver={(e, i) => handleAIDragOver(e, i)}
                            onDrop={(e, i) => handleAIDrop(e, i)}
                            onEdit={() => {}}
                            onDelete={() => {}}
                            popoverOpen={false}
                            onPopoverOpenChange={() => {}}
                          />
                          {#if index < aiDialogHighlights.length - 1}
                            <span class="mx-1"> </span>
                          {/if}
                        {/each}
                      {/if}
                    </div>
                  </div>
                </div>
              {:else}
                <div class="p-8 text-center text-muted-foreground">
                  <p class="text-lg font-medium">No results yet</p>
                  <p class="text-sm">
                    Generate AI suggestions using the instructions above to see results here
                  </p>
                </div>
              {/if}
      </div>
    </div>
  {/snippet}
  
  {#snippet footer({ closeSheet })}
    <div class="flex justify-between gap-2">
      <div class="flex gap-2">
        {#if showOriginalForm}
          <Button
            variant="outline"
            onclick={resetToOriginal}
            disabled={aiReorderLoading}
          >
            Reset to Original
          </Button>
        {/if}
      </div>
      <div class="flex gap-2">
        <Button
          variant="outline"
          onclick={closeSheet}
        >
          Cancel
        </Button>
        {#if aiReorderedHighlights.length > 0}
          <Button onclick={applyAIReordering} class="flex items-center gap-2">
            <Sparkles class="w-4 h-4" />
            Apply AI Order
          </Button>
        {/if}
      </div>
    </div>
  {/snippet}
</CustomSheet>