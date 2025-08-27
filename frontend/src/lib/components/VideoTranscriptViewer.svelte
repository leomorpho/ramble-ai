<script>
  import { tick } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$lib/components/ui/dialog";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import * as Resizable from "$lib/components/ui/resizable/index.js";
  import AISettings from "$lib/components/AISettings.svelte";
  import TextHighlighter from "$lib/components/texthighlighter/TextHighlighter.svelte";
  import CompoundVideoPlayer from "$lib/components/videoplayback/CompoundVideoPlayer.svelte";
  import { toast } from "svelte-sonner";
  import { Sparkles, Undo, Redo } from "@lucide/svelte";
  import {
    SuggestHighlightsWithAI,
    GetProjectHighlightAISettings,
    SaveProjectHighlightAISettings,
    GetSuggestedHighlights,
    UpdateVideoClipSuggestedHighlights,
  } from "$lib/wailsjs/go/main/App";
  import {
    updateVideoHighlights,
    undoHighlightsChange,
    redoHighlightsChange,
    highlightsHistoryStatus,
    updateHighlightsHistoryStatus,
    rawHighlights,
  } from "$lib/stores/projectHighlights.js";
  import {
    getNextColorId,
  } from "$lib/components/texthighlighter/TextHighlighter.utils.js";
  import CopyToClipboardButton from "./CopyToClipboardButton.svelte";

  let {
    open = $bindable(false),
    video = $bindable(null),
    projectId,
    highlights = [],
    onHighlightsChange,
  } = $props();

  // Transcript video player state (separate from main highlights)
  let transcriptPlayerHighlights = $state([]);

  // AI suggestion state - Map to track loading state per video ID
  let aiSuggestLoadingMap = $state(new Map());
  let suggestedHighlights = $state([]);
  let loadingSuggestedHighlights = $state(false);

  // AI settings state
  let selectedModel = $state("anthropic/claude-sonnet-4");
  let customPrompt = $state("");
  let customModelValue = $state("");
  let instructionsOpen = $state(false);
  let showAISuggestConfirmation = $state(false);

  // Available AI models (same as in AISettings component)
  const availableModels = [
    { value: "anthropic/claude-3.5-haiku-20241022", label: "Claude 3.5 Haiku" },
    {
      value: "anthropic/claude-3.5-sonnet-20241022",
      label: "Claude 3.5 Sonnet",
    },
    {
      value: "anthropic/claude-3-5-sonnet-20241022",
      label: "Claude 3.5 Sonnet (Latest)",
    },
    { value: "anthropic/claude-3-opus-20240229", label: "Claude 3 Opus" },
    { value: "openai/gpt-4o", label: "GPT-4o" },
    { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
    { value: "google/gemini-2.0-flash-exp", label: "Gemini 2.0 Flash" },
    { value: "google/gemini-exp-1206", label: "Gemini Experimental" },
    { value: "x-ai/grok-2-1212", label: "Grok 2" },
    { value: "custom", label: "Custom Model (Enter Below)" },
  ];

  // Default highlight suggestion prompt
  const defaultPrompt = `You are an expert content creator analyzing video transcripts to identify the most compelling and engaging moments. Your task is to suggest highlight segments that would be valuable for creating shorts, clips, or key moments.

Analyze the transcript and identify segments that are:
- Emotionally impactful or surprising
- Information-dense or educational
- Entertaining or humorous
- Controversial or thought-provoking
- Action-packed or visually interesting
- Contains key insights or takeaways

Return segments that would work well as standalone content pieces.`;

  // Derived highlights formatted for CompoundVideoPlayer (adds filePath from video)
  let formattedTranscriptHighlights = $derived(
    video && transcriptPlayerHighlights.length > 0
      ? [...transcriptPlayerHighlights] // Create a copy to avoid mutation
          .sort((a, b) => a.start - b.start) // Order by start time (temporal order)
          .map((highlight) => ({
            ...highlight,
            filePath: video.filePath,
            videoClipId: video.id,
            videoClipName: video.name,
          }))
      : []
  );

  // When video changes or highlights update, update the transcript player highlights
  $effect(() => {
    if (video) {
      // Get highlights for this specific video clip from either props or global store
      // Prefer global store (rawHighlights) for latest data, fallback to props
      const sourceHighlights = $rawHighlights.length > 0 ? $rawHighlights : highlights;
      
      if (sourceHighlights) {
        // Filter by both videoClipId and filePath for extra safety
        const videoHighlights = sourceHighlights.filter(
          (h) => h.videoClipId === video.id && h.filePath === video.filePath
        );
        
        // Ensure all highlights have valid colorIds, assign them if missing
        const processedHighlights = videoHighlights.map((h, index) => {
          let validColorId = h.colorId;
          
          // If colorId is invalid (0, null, undefined, out of range), assign a new one
          if (!validColorId || validColorId < 1 || validColorId > 20) {
            console.warn('Found highlight with invalid colorId:', h.colorId, 'for highlight:', h.id);
            validColorId = getNextColorId(videoHighlights.slice(0, index).filter(vh => vh.colorId >= 1 && vh.colorId <= 20));
            console.log('Assigned new colorId:', validColorId);
          }
          
          return {
            id: h.id,
            start: h.start,
            end: h.end,
            colorId: validColorId,
            text: h.text,
          };
        });
        
        transcriptPlayerHighlights = processedHighlights;
      }
    }
  });

  // Load AI settings and suggested highlights when dialog opens
  $effect(() => {
    if (open && projectId && video) {
      loadAISettings();
      loadSuggestedHighlights();
    }
  });

  // Load AI settings from project
  async function loadAISettings() {
    try {
      const aiSettings = await GetProjectHighlightAISettings(projectId);
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

  // Load suggested highlights from database
  async function loadSuggestedHighlights() {
    if (!video?.id) return;

    loadingSuggestedHighlights = true;
    try {
      const suggestions = await GetSuggestedHighlights(video.id);
      console.log("📥 Loaded suggested highlights from DB:", suggestions);

      // Convert to frontend format (already have timestamps from backend)
      const newSuggestions = suggestions.map((suggestion) => ({
        id: suggestion.id,
        start: suggestion.start,
        end: suggestion.end,
        color: suggestion.color,
        text: suggestion.text,
        isSuggestion: true,
      }));

      // Force reactive update by creating new array
      suggestedHighlights = [...newSuggestions];

      // Ensure Svelte processes the update
      await tick();

      console.log("✅ Updated suggestedHighlights array:", {
        count: suggestedHighlights.length,
        highlights: suggestedHighlights,
        dbCount: suggestions.length,
      });
    } catch (error) {
      console.error("Failed to load suggested highlights:", error);
      // Silently fail - suggested highlights are optional
    } finally {
      loadingSuggestedHighlights = false;
    }
  }

  async function handleHighlightsChangeInternal(highlights) {
    if (!video) return;

    console.log(
      "handleHighlightsChangeInternal called with",
      highlights.length,
      "highlights"
    );

    try {
      // Update the transcript player highlights (local state)
      transcriptPlayerHighlights = [...highlights];

      // Use the store function to update highlights
      await updateVideoHighlights(video.id, highlights);

      // Still call the parent's handler if provided for backward compatibility
      if (onHighlightsChange) {
        await onHighlightsChange(highlights);
      }

      // Reload suggested highlights to ensure we're in sync with database
      // This is important when accepting suggestions
      console.log("🔄 Reloading suggested highlights after change...");

      // Add a small delay to ensure DB operations complete
      await new Promise((resolve) => setTimeout(resolve, 200));

      try {
        await loadSuggestedHighlights();
        console.log("✅ Successfully reloaded suggested highlights");
      } catch (loadError) {
        console.error("❌ Failed to reload suggested highlights:", loadError);
      }
    } catch (err) {
      console.error("Failed to save highlights:", err);
      toast.error("Failed to save highlights", {
        description: "An error occurred while saving your highlights",
      });
    }
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, "0")}`;
  }

  // Suggest highlights inline
  async function suggestHighlightsInline() {
    console.log("🔍 AI Suggest button clicked");
    console.log("📊 Current state:", {
      hasVideo: !!video,
      hasTranscription: !!video?.transcription,
      transcriptionLength: video?.transcription?.length,
      hasTranscriptionWords: !!video?.transcriptionWords,
      transcriptionWordsCount: video?.transcriptionWords?.length,
      projectId,
      videoId: video?.id,
    });

    if (!video?.transcription) {
      console.log("❌ No transcription available");
      toast.error("Video has no transcription available");
      return;
    }

    // Set loading state for this specific video
    aiSuggestLoadingMap.set(video.id, true);
    aiSuggestLoadingMap = new Map(aiSuggestLoadingMap); // Trigger reactivity
    console.log("⏳ Setting loading state to true for video:", video.id);

    try {
      // Save current AI settings before processing
      console.log("💾 Saving AI settings...");
      const modelToSave =
        selectedModel === "custom" ? customModelValue : selectedModel;
      await SaveProjectHighlightAISettings(projectId, {
        aiModel: modelToSave,
        aiPrompt: customPrompt,
      });
      console.log("✅ Saved AI settings:", {
        model: modelToSave,
        prompt: customPrompt,
      });

      console.log("🤖 Calling SuggestHighlightsWithAI...", {
        projectId,
        videoId: video.id,
        prompt: customPrompt || "default",
      });

      // Call the AI highlight suggestion API
      const suggestions = await SuggestHighlightsWithAI(
        projectId,
        video.id,
        customPrompt || ""
      );

      console.log("📝 Raw AI suggestions received:", suggestions);
      console.log(
        "📝 Suggestions type:",
        typeof suggestions,
        Array.isArray(suggestions)
      );

      // AI suggestions already come with timestamps from backend
      const newSuggestions = suggestions.map((suggestion, index) => {
        console.log(`🔄 Processing suggestion ${index}:`, suggestion);

        const converted = {
          id: suggestion.id,
          start: suggestion.start, // Already in timestamp format
          end: suggestion.end, // Already in timestamp format
          colorId: suggestion.colorId,
          text: suggestion.text,
          isSuggestion: true,
        };

        console.log(`✨ Converted suggestion ${index}:`, converted);
        return converted;
      });

      console.log("🎯 AI generation complete, reloading from database");

      // Reload suggested highlights from database
      await loadSuggestedHighlights();

      toast.success(
        `Generated ${suggestions.length} AI highlight suggestions!`
      );
    } catch (error) {
      console.error("💥 AI highlight suggestion error:", error);
      console.error("💥 Error details:", {
        message: error.message,
        stack: error.stack,
        name: error.name,
      });
      toast.error("Failed to generate highlight suggestions", {
        description:
          error.message || "An error occurred while generating suggestions",
      });
    } finally {
      // Clear loading state for this specific video
      aiSuggestLoadingMap.delete(video.id);
      aiSuggestLoadingMap = new Map(aiSuggestLoadingMap); // Trigger reactivity
      console.log("✅ Setting loading state to false for video:", video.id);
    }
  }

  // Accept all suggested highlights
  async function acceptAllSuggestions() {
    if (suggestedHighlights.length === 0 || !video) return;

    try {
      // Convert all suggestions to regular highlights using colorId system
      const newHighlights = [];
      suggestedHighlights.forEach((suggestion, index) => {
        const newHighlight = {
          id: `highlight_${Date.now()}_${index}_${Math.random().toString(36).substr(2, 9)}`,
          start: suggestion.start,
          end: suggestion.end,
          colorId: getNextColorId([...transcriptPlayerHighlights, ...newHighlights]),
        };
        newHighlights.push(newHighlight);
      });

      // Use the store function to update all highlights at once
      const updatedHighlights = [
        ...transcriptPlayerHighlights,
        ...newHighlights,
      ];
      await updateVideoHighlights(video.id, updatedHighlights);

      // Update local state immediately
      transcriptPlayerHighlights = updatedHighlights;

      // Clear suggestions locally
      suggestedHighlights = [];

      // Update suggested highlights in database (empty array)
      await UpdateVideoClipSuggestedHighlights(video.id, []);

      toast.success(`Accepted ${newHighlights.length} highlight suggestions!`);
    } catch (err) {
      console.error("Failed to accept all suggestions:", err);
      toast.error("Failed to accept all suggestions");
    }
  }

  // Reject all suggested highlights
  async function rejectAllSuggestions() {
    if (!video) return;

    const count = suggestedHighlights.length;

    try {
      // Clear suggestions locally
      suggestedHighlights = [];

      // Update suggested highlights in database (empty array)
      await UpdateVideoClipSuggestedHighlights(video.id, []);

      toast.success(
        `Rejected ${count} highlight suggestion${count === 1 ? "" : "s"}`
      );
    } catch (err) {
      console.error("Failed to reject all suggestions:", err);
      toast.error("Failed to reject all suggestions");
    }
  }

  // Undo/Redo handlers for highlights
  async function handleUndoHighlights() {
    if (!video) return;

    try {
      await undoHighlightsChange(video.id);
      
      // Wait a moment for the store to be fully updated
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Get the current highlights from the global store after undo
      const currentHighlights = $rawHighlights.filter(
        (h) => h.videoClipId === video.id && h.filePath === video.filePath
      );
      
      // Convert to the format expected by the component
      const processedHighlights = currentHighlights.map((h, index) => {
        let validColorId = h.colorId;
        
        // If colorId is invalid (0, null, undefined, out of range), assign a new one
        if (!validColorId || validColorId < 1 || validColorId > 20) {
          console.warn('🎨 Found highlight with invalid colorId:', h.colorId, 'for highlight:', h.id);
          validColorId = getNextColorId(currentHighlights.slice(0, index).filter(vh => vh.colorId >= 1 && vh.colorId <= 20));
          console.log('🎨 Assigned new colorId:', validColorId);
        }
        
        return {
          id: h.id,
          start: h.start,
          end: h.end,
          colorId: validColorId,
          text: h.text,
        };
      });
      
      // Update the local state immediately
      transcriptPlayerHighlights = processedHighlights;
      
      // Also notify the parent component if available
      if (onHighlightsChange) {
        await onHighlightsChange(processedHighlights);
      }
    } catch (error) {
      console.error("Failed to undo highlights:", error);
    }
  }

  async function handleRedoHighlights() {
    if (!video) return;

    try {
      await redoHighlightsChange(video.id);
      
      // Wait a moment for the store to be fully updated
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Get the current highlights from the global store after redo
      const currentHighlights = $rawHighlights.filter(
        (h) => h.videoClipId === video.id && h.filePath === video.filePath
      );
      
      // Convert to the format expected by the component
      const processedHighlights = currentHighlights.map((h, index) => {
        let validColorId = h.colorId;
        
        // If colorId is invalid (0, null, undefined, out of range), assign a new one
        if (!validColorId || validColorId < 1 || validColorId > 20) {
          console.warn('🎨 Found highlight with invalid colorId:', h.colorId, 'for highlight:', h.id);
          validColorId = getNextColorId(currentHighlights.slice(0, index).filter(vh => vh.colorId >= 1 && vh.colorId <= 20));
          console.log('🎨 Assigned new colorId:', validColorId);
        }
        
        return {
          id: h.id,
          start: h.start,
          end: h.end,
          colorId: validColorId,
          text: h.text,
        };
      });
      
      // Update the local state immediately
      transcriptPlayerHighlights = processedHighlights;
      
      // Also notify the parent component if available
      if (onHighlightsChange) {
        await onHighlightsChange(processedHighlights);
      }
    } catch (error) {
      console.error("Failed to redo highlights:", error);
    }
  }

  // Effect to update highlights history status when video changes
  $effect(() => {
    if (video?.id) {
      updateHighlightsHistoryStatus(video.id);
    }
  });

  // Prevent body scroll when modal is open
  $effect(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    
    // Cleanup on component destroy
    return () => {
      document.body.style.overflow = '';
    };
  });
</script>

<Dialog bind:open>
  <DialogContent class="!w-screen !h-screen !max-w-none !max-h-none !m-0 !p-4 !rounded-none !left-0 !top-0 !translate-x-0 !translate-y-0 flex flex-col">
    <DialogHeader>
      <DialogTitle>Video Transcript</DialogTitle>
      <DialogDescription>
        {#if video}
          Transcript for {video.name}
        {/if}
      </DialogDescription>
    </DialogHeader>

    <ScrollArea class="flex-1 lg:h-[calc(95vh-10rem)] h-[60vh]">
      {#snippet children()}
        <div class="p-4">
          <Resizable.PaneGroup direction="horizontal" class="h-[calc(100vh-12rem)] border rounded">
        <!-- Video Player Pane -->
        <Resizable.Pane defaultSize={50}>
          <div class="h-full p-4">
            {#if video}
              <div class="bg-background h-full flex flex-col">
                <h3 class="font-medium mb-3">Video Preview</h3>
                <div class="flex-1 min-h-0">
                  <CompoundVideoPlayer
                    videoHighlights={formattedTranscriptHighlights}
                    {projectId}
                    enableEyeButton={false}
                    enableReordering={false}
                  />
                </div>
              </div>
            {/if}
          </div>
        </Resizable.Pane>
        
        <Resizable.Handle />
        
        <!-- Transcript Pane -->
        <Resizable.Pane defaultSize={50}>
          <div class="h-full p-4 flex flex-col">
            {#if video}
              {#if video.transcription}
                <div class="flex flex-col h-full space-y-3">
                  <div class="flex-shrink-0 space-y-2">
                    <div class="flex flex-wrap items-start justify-between gap-2">
                      <h3 class="font-medium">Transcript</h3>
                      <div class="flex flex-wrap gap-2 items-center">
                        {#if video.transcriptionLanguage}
                          <span class="text-xs border rounded px-2 py-1">
                            {video.transcriptionLanguage.toUpperCase()}
                          </span>
                        {/if}
                        {#if video.transcriptionDuration}
                          <span class="text-xs border rounded px-2 py-1">
                            {formatTimestamp(video.transcriptionDuration)}
                          </span>
                        {/if}
                        <!-- Undo/Redo buttons for highlights - temporarily hidden -->
                        <!-- <div class="flex items-center gap-1">
                          <Button
                            variant="outline"
                            size="sm"
                            onclick={handleUndoHighlights}
                            disabled={!$highlightsHistoryStatus.get(video.id)
                              ?.canUndo}
                            class="text-xs px-2"
                            title="Undo highlights change (Ctrl+Z)"
                          >
                            <Undo class="w-3 h-3" />
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onclick={handleRedoHighlights}
                            disabled={!$highlightsHistoryStatus.get(video.id)
                              ?.canRedo}
                            class="text-xs px-2"
                            title="Redo highlights change (Ctrl+Y)"
                          >
                            <Redo class="w-3 h-3" />
                          </Button>
                        </div> -->
                        <Button
                          variant="outline"
                          size="sm"
                          onclick={() => showAISuggestConfirmation = true}
                          class="text-xs"
                          disabled={!video.transcription ||
                            aiSuggestLoadingMap.get(video.id)}
                        >
                          <Sparkles class="w-3 h-3 mr-1" />
                          {aiSuggestLoadingMap.get(video.id)
                            ? "AI Analyzing..."
                            : "AI Suggest"}
                        </Button>
                        <CopyToClipboardButton
                          text={video?.transcription}
                          confirmationText={"Copied transcript to clipboard"}
                          failureText={"Failed to copy transcript to clipboard"}
                        />
                      </div>
                    </div>
                  </div>

                  <!-- AI Settings -->
                  <div class="flex-shrink-0">
                    <AISettings
                      bind:open={instructionsOpen}
                      bind:selectedModel
                      bind:customModelValue
                      bind:customPrompt
                      {defaultPrompt}
                      title="AI Settings"
                      modelDescription="Choose the AI model for highlight suggestions. Different models have varying strengths in content analysis."
                      promptDescription="Customize how AI identifies highlight-worthy segments in your transcript."
                      promptPlaceholder="AI instructions for highlighting..."
                    />
                  </div>

                  <!-- Bulk suggestion actions -->
                  {#if suggestedHighlights.length > 0}
                    <div class="flex-shrink-0">
                      <div class="flex flex-wrap items-center justify-between gap-2 p-3 border rounded">
                        <span class="text-sm text-muted-foreground">
                          {suggestedHighlights.length} AI suggestion{suggestedHighlights.length ===
                          1
                            ? ""
                            : "s"}
                        </span>
                        <div class="flex flex-wrap gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onclick={acceptAllSuggestions}
                            class="text-xs"
                          >
                            <svg
                              class="w-3 h-3 mr-1"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 13l4 4L19 7"
                              />
                            </svg>
                            Accept All
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onclick={rejectAllSuggestions}
                            class="text-xs"
                          >
                            <svg
                              class="w-3 h-3 mr-1"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M6 18L18 6M6 6l12 12"
                              />
                            </svg>
                            Reject All
                          </Button>
                        </div>
                      </div>
                    </div>
                  {/if}

                  <!-- Scrollable Transcript Text -->
                  <div class="flex-1 min-h-0">
                    <ScrollArea class="h-[calc(100vh-20rem)] bg-background border rounded-lg">
                      {#snippet children()}
                        <div class="p-4 space-y-2">
                          <div class="text-sm leading-relaxed">
                            <TextHighlighter
                              text={video.transcription}
                              words={video.transcriptionWords || []}
                              highlights={transcriptPlayerHighlights}
                              {suggestedHighlights}
                              videoId={video.id}
                              onHighlightsChange={handleHighlightsChangeInternal}
                            />
                          </div>
                          <div class="text-xs text-muted-foreground">
                            Character count: {video.transcription.length}
                          </div>
                        </div>
                      {/snippet}
                    </ScrollArea>
                  </div>
                </div>
              {:else}
                <div
                  class="h-full flex items-center justify-center text-muted-foreground"
                >
                  <div class="text-center">
                    <svg
                      class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                    <p class="text-lg font-medium">No transcript available</p>
                    <p class="text-sm">
                      This video hasn't been transcribed yet.
                    </p>
                  </div>
                </div>
              {/if}
            {/if}
          </div>
        </Resizable.Pane>
          </Resizable.PaneGroup>
        </div>
      {/snippet}
    </ScrollArea>

  </DialogContent>
</Dialog>

<!-- AI Suggest Confirmation Dialog -->
<Dialog bind:open={showAISuggestConfirmation}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Generate AI Highlight Suggestions?</DialogTitle>
      <DialogDescription>
        <div class="space-y-3 pt-2">
          <p>
            The AI will analyze your transcript to identify the most compelling moments for highlights.
          </p>
          <div class="border rounded p-3 space-y-2">
            <p class="text-sm">
              <strong>Model:</strong> {selectedModel === "custom" ? customModelValue : availableModels.find(m => m.value === selectedModel)?.label || selectedModel}
            </p>
            <div class="text-sm space-y-1">
              <p class="font-medium">This will:</p>
              <ul class="list-disc list-inside text-muted-foreground">
                <li>Send your transcript to the AI model</li>
                <li>Generate suggested highlight segments</li>
                <li>Display them as preview highlights (not saved automatically)</li>
              </ul>
            </div>
          </div>
          <p class="text-sm text-muted-foreground">
            You can review, accept, or reject the suggestions before they're saved.
          </p>
        </div>
      </DialogDescription>
    </DialogHeader>
    <div class="flex justify-end gap-2 pt-4">
      <Button
        variant="outline"
        onclick={() => showAISuggestConfirmation = false}
      >
        Cancel
      </Button>
      <Button
        onclick={() => {
          showAISuggestConfirmation = false;
          suggestHighlightsInline();
        }}
      >
        <Sparkles class="w-4 h-4 mr-2" />
        Generate Suggestions
      </Button>
    </div>
  </DialogContent>
</Dialog>
