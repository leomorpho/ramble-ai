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
  import AISettings from "$lib/components/ui/AISettings.svelte";
  import CustomSheet from "$lib/components/ui/CustomSheet.svelte";
  import CompoundVideoPlayer from "$lib/components/videoplayback/CompoundVideoPlayer.svelte";
  import ReorderableHighlights from "$lib/components/ReorderableHighlights.svelte";
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
  let selectedModel = $state("anthropic/claude-sonnet-4");
  let hasCachedSuggestion = $state(false);
  let cachedSuggestionDate = $state(null);
  let cachedSuggestionModel = $state("");
  let showOriginalForm = $state(false);
  let originalHighlights = $state([]);

  // AI dialog independent state
  let aiDialogHighlights = $state([]);
  let aiSelectedHighlights = $state(new Set());

  let customModelValue = $state("");

  // Collapsible state for AI instructions
  let instructionsOpen = $state(false);

  // Available AI models (needed for logic in this component)
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
      value: "anthropic/claude-3.5-haiku-20241022",
      label: "Claude 3.5 Haiku (Fast)",
    },
    { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
    { value: "mistralai/mistral-nemo", label: "Mistral Nemo" },
    { value: "custom", label: "Custom Model" },
  ];

  // Default YouTube expert prompt
  const defaultPrompt = `You are an expert video editor focused on creating well-structured, engaging content. Your goal is to organize these highlight segments into a balanced, coherent video with natural pacing and flow.

Key principles for structuring the video:
- Create an adaptive structure that fits the specific content
- Balance sections by total text length, not highlight count
- Build natural rhythm with highs and lows throughout
- Ensure smooth transitions between different topics
- Maintain viewer engagement through variety and pacing

Section balancing guidelines:
- Analyze the cumulative text length in each section
- No single section should contain more than 30% of total content
- Short highlights can be grouped together, long ones may stand alone
- Think in terms of speaking time and content weight

Create sections that feel complete yet connected, with clear but simple titles that describe their purpose. The structure should emerge from the content itself, not force content into a rigid template.`;

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

    // Load project AI settings
    try {
      const aiSettings = await GetProjectAISettings(projectId);
      selectedModel = aiSettings.aiModel || "anthropic/claude-sonnet-4";
      customPrompt = aiSettings.aiPrompt || defaultPrompt;

      // If using custom model, extract the value
      if (!availableModels.find((m) => m.value === selectedModel)) {
        customModelValue = selectedModel;
        selectedModel = "custom";
      }
    } catch (error) {
      console.error("Failed to load AI settings:", error);
      selectedModel = "anthropic/claude-sonnet-4";
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

        // Flatten consecutive 'N' characters before processing cached suggestion
        const flattenedCachedOrder = flattenConsecutiveNewlines(cachedSuggestion.order);
        
        // Reorder based on cached suggestion - preserve 'N' characters for newlines
        for (const id of flattenedCachedOrder) {
          if (isNewline(id)) {
            // Preserve newline characters in the format expected by ReorderableHighlights
            reorderedHighlights.push(createNewlineFromDb(id));
          } else {
            const highlight = highlightsMap.get(id);
            if (highlight) {
              reorderedHighlights.push(highlight);
            }
          }
        }

        // Add any highlights that weren't in the cached order (exclude newlines - they're handled separately)
        for (const highlight of aiDialogHighlights) {
          if (!flattenedCachedOrder.includes(highlight.id) && !isNewline(highlight)) {
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

      // Flatten consecutive 'N' characters before processing
      const flattenedReorderedIds = flattenConsecutiveNewlines(reorderedIds);
      
      // Build reordered array - preserve 'N' characters for newlines
      for (const id of flattenedReorderedIds) {
        if (isNewline(id)) {
          // Preserve newline characters in the format expected by ReorderableHighlights
          reorderedHighlights.push(createNewlineFromDb(id));
        } else {
          const highlight = highlightsMap.get(id);
          if (highlight) {
            reorderedHighlights.push(highlight);
          }
        }
      }

      // Add any highlights that weren't in the AI response (exclude newlines - they're handled separately)
      for (const highlight of aiDialogHighlights) {
        if (!flattenedReorderedIds.includes(highlight.id) && !isNewline(highlight)) {
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
      // Get the exact AI suggestion order from the database and use it unchanged
      // This ensures we preserve section objects EXACTLY as they are in ai_suggestion_order
      const cachedSuggestion = await GetProjectAISuggestion(projectId);
      
      if (cachedSuggestion && cachedSuggestion.order && cachedSuggestion.order.length > 0) {
        // Use the exact AI suggestion order without any conversion
        const success = await updateHighlightOrder(cachedSuggestion.order);

        if (success) {
          open = false; // Close the sheet
          toast.success("AI reordering applied successfully!");
          onApply(aiReorderedHighlights);
        }
      } else {
        // Fallback: convert current aiReorderedHighlights if no cached suggestion
        const highlightIds = aiReorderedHighlights.map(item => {
          if (isNewline(item)) {
            // Convert display format to database format - preserve title exactly as is
            return item.title ? { type: 'N', title: item.title } : 'N';
          } else {
            return item.id;
          }
        });

        const success = await updateHighlightOrder(highlightIds);

        if (success) {
          open = false; // Close the sheet
          toast.success("AI reordering applied successfully!");
          onApply(aiReorderedHighlights);
        }
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

  // Handle reordering from the ReorderableHighlights component
  async function handleAIReorder(newHighlights) {
    // Update both AI dialog highlights and reordered highlights for sync
    aiDialogHighlights = newHighlights;
    aiReorderedHighlights = newHighlights;
  }

  // Handle reordering from the CompoundVideoPlayer in AI preview mode
  async function handleAIVideoReorder(newHighlights) {
    // Update both AI dialog highlights and reordered highlights
    aiDialogHighlights = newHighlights;
    aiReorderedHighlights = newHighlights;
    return Promise.resolve(); // Return resolved promise for success
  }

  // Handle title change for newlines in AI dialog
  function handleAITitleChange(index, newTitle) {
    if (index < aiDialogHighlights.length && isNewline(aiDialogHighlights[index])) {
      // Update the title in the current highlight (local state only)
      // Titles will be saved to backend when user applies the reorder
      const updatedHighlights = [...aiDialogHighlights];
      updatedHighlights[index] = {
        ...updatedHighlights[index],
        title: newTitle
      };
      aiDialogHighlights = updatedHighlights;
      aiReorderedHighlights = updatedHighlights;
    }
  }

  // Utility functions for newline handling with titles
  function isNewline(item) {
    return item === 'N' || item === 'n' || (typeof item === 'object' && item.type === 'N') || (typeof item === 'object' && item.type === 'newline');
  }

  function getNewlineTitle(item) {
    if (typeof item === 'object' && (item.type === 'N' || item.type === 'newline')) {
      return item.title || '';
    }
    return '';
  }

  function createNewlineFromDb(dbItem) {
    if (dbItem === 'N') {
      return {
        id: `newline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        type: 'newline',
        title: ''
      };
    }
    if (typeof dbItem === 'object' && dbItem.type === 'N') {
      return {
        id: `newline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        type: 'newline',
        title: dbItem.title || ''
      };
    }
    return dbItem;
  }

  // Utility function to flatten consecutive 'N' characters
  function flattenConsecutiveNewlines(ids) {
    if (!ids || ids.length <= 1) {
      return ids;
    }

    const result = [];
    let lastWasNewline = false;

    for (const id of ids) {
      if (isNewline(id)) {
        if (!lastWasNewline) {
          result.push(id);
          lastWasNewline = true;
        } else {
          // When flattening consecutive newlines, preserve the title if the current one has one
          const lastNewline = result[result.length - 1];
          const currentTitle = getNewlineTitle(id);
          const lastTitle = getNewlineTitle(lastNewline);
          
          if (currentTitle && !lastTitle) {
            // Replace the last newline with the current one that has a title
            result[result.length - 1] = id;
          }
        }
        // Skip consecutive newlines
      } else {
        result.push(id);
        lastWasNewline = false;
      }
    }

    return result;
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
      <!-- AI Instructions & Settings -->
      <AISettings
        bind:open={instructionsOpen}
        bind:selectedModel
        bind:customModelValue
        bind:customPrompt
        {defaultPrompt}
        title="AI Reordering"
        modelDescription="Choose the AI model for highlight reordering. Different models have varying strengths in content analysis and reasoning."
        promptDescription="Modify the prompt above to customize how AI reorders your highlights. The default focuses on YouTube best practices for maximum engagement."
        promptPlaceholder="AI instructions for reordering highlights..."
        {availableModels}
        showResetButton={false}
        loading={aiReorderLoading}
        hasRun={hasCachedSuggestion}
        onRun={startAIReordering}
      />

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
                    <CompoundVideoPlayer 
                      videoHighlights={aiReorderedHighlights.length > 0 ? aiReorderedHighlights.filter(h => h.id && h.id.startsWith('highlight_')) : highlights.filter(h => h.id && h.id.startsWith('highlight_'))} 
                      {projectId} 
                      enableEyeButton={false}
                      onReorder={handleAIVideoReorder}
                    />
                  </div>

                  <!-- AI Dialog Timeline -->
                  <div class="bg-muted/30 rounded-lg p-4">
                    <h3 class="text-sm font-medium mb-3">
                      AI Suggested Order (read-only preview):
                    </h3>

                    <!-- Timeline-style highlight display -->
                    <ReorderableHighlights
                      highlights={aiDialogHighlights}
                      bind:selectedHighlights={aiSelectedHighlights}
                      onReorder={() => {}}
                      onSelect={null}
                      onEdit={() => {}}
                      onDelete={() => {}}
                      onPopoverOpenChange={() => {}}
                      getHighlightWords={() => []}
                      isPopoverOpen={() => false}
                      onTitleChange={() => {}}
                      enableMultiSelect={false}
                      enableNewlines={true}
                      enableSelection={false}
                      enableEdit={false}
                      enableDelete={false}
                      enableDrag={false}
                      showAddNewLineButtons={false}
                      containerClass="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border"
                    />
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