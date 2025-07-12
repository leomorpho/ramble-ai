<script>
  import { onMount } from "svelte";
  import {
    ImproveHighlightSilencesWithAI,
    GetProjectAISilenceResult,
    ClearAISilenceImprovements,
    GetProjectAISettings,
    SaveProjectAISettings,
  } from "$lib/wailsjs/go/main/App";
  import { toast } from "svelte-sonner";
  import { Sparkles, Clock } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import AISettings from "$lib/components/AISettings.svelte";
  import CustomSheet from "$lib/components/CustomSheet.svelte";
  import CompoundVideoPlayer from "$lib/components/videoplayback/CompoundVideoPlayer.svelte";
  import HighlightItem from "$lib/components/HighlightItem.svelte";
  import { Play } from "@lucide/svelte";

  // Props
  let {
    open = $bindable(false),
    projectId,
    highlights = [],
    onApply = () => {},
  } = $props();

  // AI silence improvement state
  let aiSilenceLoading = $state(false);
  let improvedHighlights = $state([]);
  let aiSilenceError = $state("");
  let customPrompt = $state("");
  let selectedModel = $state("anthropic/claude-sonnet-4");
  let hasCachedImprovements = $state(false);
  let cachedImprovementsDate = $state(null);
  let cachedImprovementsModel = $state("");
  let showOriginalForm = $state(false);
  let originalHighlights = $state([]);

  // AI settings state
  let customModelValue = $state("");
  let instructionsOpen = $state(false);

  // Available AI models (same as in AISettings component)
  const availableModels = [
    { value: "anthropic/claude-3.5-haiku-20241022", label: "Claude 3.5 Haiku" },
    { value: "anthropic/claude-3.5-sonnet-20241022", label: "Claude 3.5 Sonnet" },
    { value: "anthropic/claude-3-5-sonnet-20241022", label: "Claude 3.5 Sonnet (Latest)" },
    { value: "anthropic/claude-3-opus-20240229", label: "Claude 3 Opus" },
    { value: "openai/gpt-4o", label: "GPT-4o" },
    { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
    { value: "google/gemini-2.0-flash-exp", label: "Gemini 2.0 Flash" },
    { value: "google/gemini-exp-1206", label: "Gemini Experimental" },
    { value: "x-ai/grok-2-1212", label: "Grok 2" },
    { value: "custom", label: "Custom Model (Enter Below)" }
  ];

  // Default silence improvement prompt
  const defaultPrompt = `You are an expert video editor specializing in creating natural-sounding speech cuts. Your task is to improve highlight timings by including appropriate silence buffers that make the speech flow naturally.

For each highlight, you're given:
- The current start/end times
- The text content
- The end time of the word before the highlight starts
- The start time of the word after the highlight ends

Adjust the start and end times to include natural pauses while staying within the given boundaries. Consider:
- Include slight pauses before sentences or thoughts (100-300ms)
- Include natural breathing room after sentences (200-500ms)
- For questions, include the pause before the answer
- For dramatic statements, include the build-up pause
- Never cut into the middle of words
- Prefer to include complete breaths and natural speech rhythms

Only include highlights where you recommend changes.`;

  // Reset state when opening
  $effect(() => {
    if (open) {
      aiSilenceError = "";
      showOriginalForm = false;
      originalHighlights = [...highlights];
      loadAISettings();
      loadCachedImprovements();
    }
  });

  // Load AI settings from project
  async function loadAISettings() {
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
  }

  // Load cached silence improvements
  async function loadCachedImprovements() {
    try {
      const cachedResult = await GetProjectAISilenceResult(projectId);
      if (cachedResult) {
        improvedHighlights = cachedResult.improvements;
        hasCachedImprovements = true;
        cachedImprovementsDate = new Date(cachedResult.createdAt).toLocaleString();
        cachedImprovementsModel = cachedResult.model;
        showOriginalForm = true; // Show comparison when loading cached results
        console.log("Loaded cached improvements:", improvedHighlights.length, "videos", improvedHighlights);
      } else {
        improvedHighlights = [];
        hasCachedImprovements = false;
        cachedImprovementsDate = null;
        cachedImprovementsModel = "";
      }
    } catch (error) {
      console.error("Failed to load cached improvements:", error);
      improvedHighlights = [];
      hasCachedImprovements = false;
    }
  }

  // Generate AI silence improvements
  async function generateAISilenceImprovements() {
    if (highlights.length === 0) {
      toast.error("No highlights to improve");
      return;
    }

    aiSilenceLoading = true;
    aiSilenceError = "";
    
    try {
      // Save current AI settings before processing
      const modelToSave = selectedModel === "custom" ? customModelValue : selectedModel;
      await SaveProjectAISettings(projectId, {
        aiModel: modelToSave,
        aiPrompt: customPrompt,
      });

      // Call AI service
      const improvements = await ImproveHighlightSilencesWithAI(projectId);
      
      improvedHighlights = improvements;
      hasCachedImprovements = true;
      cachedImprovementsDate = new Date().toLocaleString();
      cachedImprovementsModel = modelToSave;
      showOriginalForm = true; // Show comparison by default after generation
      
      toast.success(`Generated AI silence improvements for ${improvements.length} video(s)!`);
    } catch (error) {
      console.error("Failed to generate AI silence improvements:", error);
      aiSilenceError = error.message || "Failed to generate improvements";
      toast.error("Failed to generate AI silence improvements", {
        description: error.message || "An error occurred while processing"
      });
    } finally {
      aiSilenceLoading = false;
    }
  }

  // Apply improvements to project
  async function applyImprovements() {
    if (improvedHighlights.length === 0) {
      toast.error("No improvements to apply");
      return;
    }

    try {
      // Clear cached improvements when applying
      await ClearAISilenceImprovements(projectId);
      
      // Call parent onApply function with improved highlights
      await onApply(improvedHighlights);
      
      // Close the sheet
      open = false;
      
      toast.success("Applied AI silence improvements!");
    } catch (error) {
      console.error("Failed to apply improvements:", error);
      toast.error("Failed to apply improvements", {
        description: error.message || "An error occurred while applying"
      });
    }
  }

  // Clear cached improvements
  async function clearCachedImprovements() {
    try {
      await ClearAISilenceImprovements(projectId);
      improvedHighlights = [];
      hasCachedImprovements = false;
      cachedImprovementsDate = null;
      cachedImprovementsModel = "";
      toast.success("Cleared cached improvements");
    } catch (error) {
      console.error("Failed to clear cached improvements:", error);
      toast.error("Failed to clear cached improvements");
    }
  }

  // Reset to original highlights before AI generation
  function resetToOriginal() {
    improvedHighlights = [];
    showOriginalForm = false;
    hasCachedImprovements = false;
    cachedImprovementsDate = null;
    cachedImprovementsModel = "";
    toast.success("Reset to original highlight timings");
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, '0')}`;
  }

  // Flatten improved highlights for display
  let flattenedImprovements = $derived(() => {
    if (!improvedHighlights || improvedHighlights.length === 0) {
      console.log("No improved highlights to flatten");
      return [];
    }
    
    console.log("Flattening improved highlights:", improvedHighlights);
    const flattened = [];
    improvedHighlights.forEach((video, videoIndex) => {
      console.log(`Video ${videoIndex}:`, video);
      console.log(`Video highlights:`, video.highlights);
      
      if (video.highlights && video.highlights.length > 0) {
        video.highlights.forEach(highlight => {
          flattened.push({
            ...highlight,
            videoClipName: video.videoClipName,
            videoClipId: video.videoClipId,
            filePath: video.filePath,
          });
        });
      } else {
        console.log(`Video ${videoIndex} has no highlights or highlights is empty/undefined`);
      }
    });
    console.log("Flattened improvements:", flattened.length, "highlights", flattened);
    return flattened;
  });

  // Flatten original highlights for comparison
  let flattenedOriginals = $derived(() => {
    if (!originalHighlights || originalHighlights.length === 0) return [];
    
    const flattened = [];
    originalHighlights.forEach(video => {
      video.highlights.forEach(highlight => {
        flattened.push({
          ...highlight,
          videoClipName: video.videoClipName,
          videoClipId: video.videoClipId,
          filePath: video.filePath,
        });
      });
    });
    return flattened;
  });
</script>

<CustomSheet 
  bind:open 
  title="AI Improved Silences" 
  description="Use AI to improve highlight timings with natural silence buffers for better flow"
>
  {#snippet icon()}
    <Clock class="w-5 h-5" />
  {/snippet}
  {#snippet children()}
    <!-- Content -->
    <div class="p-6 space-y-6">
      <!-- AI Settings -->
      <AISettings
        bind:open={instructionsOpen}
        bind:selectedModel
        bind:customModelValue
        bind:customPrompt
        {defaultPrompt}
        title="AI Silence Improvement"
        modelDescription="Choose the AI model for silence improvement analysis. Different models have varying strengths in understanding speech patterns."
        promptDescription="Customize how AI analyzes and improves highlight timings for natural speech flow."
        promptPlaceholder="AI instructions for silence improvement..."
        {availableModels}
        showResetButton={false}
        loading={aiSilenceLoading}
        hasRun={hasCachedImprovements}
        onRun={generateAISilenceImprovements}
      />

      {#if hasCachedImprovements && cachedImprovementsDate}
        <div class="bg-secondary/30 border border-dashed rounded-lg p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-medium flex items-center gap-2">
                <Sparkles class="w-4 h-4" />
                Cached AI Improvements Available
              </div>
              <div class="text-xs text-muted-foreground mt-1">
                Generated on {cachedImprovementsDate} using {cachedImprovementsModel}
              </div>
            </div>
            <Button variant="outline" size="sm" onclick={clearCachedImprovements}>
              Clear Cache
            </Button>
          </div>
        </div>
      {/if}

      {#if aiSilenceError}
        <div class="text-sm text-destructive bg-destructive/10 p-3 rounded-md border border-destructive/20">
          <p class="font-medium">Error generating improvements:</p>
          <p>{aiSilenceError}</p>
          <div class="flex justify-center gap-2 mt-3">
            <Button
              variant="outline"
              onclick={() => {
                aiSilenceError = "";
                improvedHighlights = [];
                instructionsOpen = true;
              }}
            >
              Back to Instructions
            </Button>
          </div>
        </div>
      {:else if highlights.length === 0}
        <div class="text-sm text-muted-foreground bg-muted/50 p-3 rounded-md border">
          No highlights available to improve. Create some highlights first.
        </div>
      {/if}

      <!-- Debug Info (temporary) -->
      <div class="bg-gray-100 p-2 text-xs rounded border">
        <div>Debug: improvedHighlights.length = {improvedHighlights.length}</div>
        <div>Debug: flattenedImprovements.length = {flattenedImprovements.length}</div>
        <div>Debug: hasCachedImprovements = {hasCachedImprovements}</div>
        <div>Debug: showOriginalForm = {showOriginalForm}</div>
      </div>

      <!-- Results Section -->
      <div class="space-y-6">
        {#if aiSilenceLoading}
          <div class="p-8 text-center">
            <div
              class="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-4"
            ></div>
            <p class="text-lg font-medium">
              AI is analyzing speech patterns...
            </p>
            <p class="text-sm text-muted-foreground">
              This may take a few moments
            </p>
          </div>
        {:else if aiSilenceError}
          <div class="p-6 text-center space-y-4">
            <div
              class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
            >
              <p class="font-medium">Error</p>
              <p class="text-sm">{aiSilenceError}</p>
            </div>
            <div class="flex justify-center gap-2">
              <Button
                variant="outline"
                onclick={() => {
                  aiSilenceError = "";
                  improvedHighlights = [];
                  instructionsOpen = true;
                }}
              >
                Back to Instructions
              </Button>
            </div>
          </div>
        {:else if flattenedImprovements.length > 0}
          <div class="space-y-6">
            <!-- Preview Video Player -->
            <div class="bg-card border rounded-lg p-4">
              <h3 class="text-sm font-medium mb-3 flex items-center gap-2">
                <Play class="w-4 h-4" />
                Preview AI Improved Timings
              </h3>
              <CompoundVideoPlayer 
                videoHighlights={improvedHighlights} 
                {projectId} 
                enableEyeButton={false}
              />
            </div>

            <!-- AI Improved Timeline -->
            <div class="bg-muted/30 rounded-lg p-4">
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-medium">
                  AI Improved Highlights ({flattenedImprovements.length}):
                </h3>
                {#if showOriginalForm}
                  <Button variant="outline" size="sm" onclick={() => showOriginalForm = false}>
                    Hide Comparison
                  </Button>
                {:else}
                  <Button variant="outline" size="sm" onclick={() => showOriginalForm = true}>
                    Show Original vs Improved
                  </Button>
                {/if}
              </div>

              {#if showOriginalForm && flattenedOriginals.length > 0}
                <!-- Side-by-side comparison -->
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  <!-- Original highlights -->
                  <div class="space-y-2">
                    <h4 class="text-xs font-medium text-muted-foreground">Original Timings:</h4>
                    <div class="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border">
                      {#each flattenedOriginals as highlight, index}
                        <HighlightItem
                          {highlight}
                          {index}
                          isSelected={false}
                          isDragging={false}
                          isBeingDragged={false}
                          showDropIndicatorBefore={false}
                          onSelect={() => {}}
                          onDragStart={() => {}}
                          onDragEnd={() => {}}
                          onDragOver={() => {}}
                          onDrop={() => {}}
                          onEdit={() => {}}
                          onDelete={() => {}}
                          popoverOpen={false}
                          onPopoverOpenChange={() => {}}
                        />
                        {#if index < flattenedOriginals.length - 1}
                          <span class="mx-1"> </span>
                        {/if}
                      {/each}
                    </div>
                  </div>

                  <!-- Improved highlights -->
                  <div class="space-y-2">
                    <h4 class="text-xs font-medium text-muted-foreground">AI Improved Timings:</h4>
                    <div class="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border border-primary/20">
                      {#each flattenedImprovements as highlight, index}
                        <HighlightItem
                          {highlight}
                          {index}
                          isSelected={false}
                          isDragging={false}
                          isBeingDragged={false}
                          showDropIndicatorBefore={false}
                          onSelect={() => {}}
                          onDragStart={() => {}}
                          onDragEnd={() => {}}
                          onDragOver={() => {}}
                          onDrop={() => {}}
                          onEdit={() => {}}
                          onDelete={() => {}}
                          popoverOpen={false}
                          onPopoverOpenChange={() => {}}
                        />
                        {#if index < flattenedImprovements.length - 1}
                          <span class="mx-1"> </span>
                        {/if}
                      {/each}
                    </div>
                  </div>
                </div>
              {:else}
                <!-- Single view - showing improved highlights -->
                <div class="p-4 bg-background rounded-lg min-h-[80px] relative leading-relaxed text-base border border-primary/20">
                  {#each flattenedImprovements as highlight, index}
                    <HighlightItem
                      {highlight}
                      {index}
                      isSelected={false}
                      isDragging={false}
                      isBeingDragged={false}
                      showDropIndicatorBefore={false}
                      onSelect={() => {}}
                      onDragStart={() => {}}
                      onDragEnd={() => {}}
                      onDragOver={() => {}}
                      onDrop={() => {}}
                      onEdit={() => {}}
                      onDelete={() => {}}
                      popoverOpen={false}
                      onPopoverOpenChange={() => {}}
                    />
                    {#if index < flattenedImprovements.length - 1}
                      <span class="mx-1"> </span>
                    {/if}
                  {/each}
                </div>
              {/if}
            </div>
          </div>
        {:else if highlights.length > 0}
          <div class="p-8 text-center text-muted-foreground">
            <p class="text-lg font-medium">No results yet</p>
            <p class="text-sm">
              Generate AI improvements using the instructions above to see results here
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
            disabled={aiSilenceLoading}
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
        {#if flattenedImprovements.length > 0}
          <Button onclick={applyImprovements} class="flex items-center gap-2">
            <Clock class="w-4 h-4" />
            Apply Improvements
          </Button>
        {/if}
      </div>
    </div>
  {/snippet}
</CustomSheet>