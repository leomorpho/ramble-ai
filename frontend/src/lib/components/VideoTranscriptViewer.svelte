<script>
  import { Button } from "$lib/components/ui/button";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { 
    Tabs, 
    TabsContent, 
    TabsList, 
    TabsTrigger 
  } from "$lib/components/ui/tabs";
  import { ScrollArea } from "$lib/components/ui/scroll-area";
  import AISettings from "$lib/components/ui/AISettings.svelte";
  import TextHighlighter from "$lib/components/TextHighlighter.svelte";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import { toast } from "svelte-sonner";
  import { Sparkles } from "@lucide/svelte";
  import {
    SuggestHighlightsWithAI,
    GetProjectHighlightAISettings,
    SaveProjectHighlightAISettings,
    GetSuggestedHighlights,
    UpdateVideoClipSuggestedHighlights,
  } from "$lib/wailsjs/go/main/App";

  let { 
    open = $bindable(false),
    video = $bindable(null),
    projectId,
    onHighlightsChange
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
  
  // Derived highlights formatted for EtroVideoPlayer (adds filePath from video)
  let formattedTranscriptHighlights = $derived(
    video && transcriptPlayerHighlights.length > 0 
      ? transcriptPlayerHighlights.map(highlight => ({
          ...highlight,
          filePath: video.filePath,
          videoClipId: video.id,
          videoClipName: video.name
        }))
      : []
  );

  // When video changes, update the transcript player highlights
  $effect(() => {
    if (video) {
      transcriptPlayerHighlights = video.highlights ? [...video.highlights] : [];
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
      suggestedHighlights = suggestions.map(suggestion => ({
        id: suggestion.id,
        start: suggestion.start,
        end: suggestion.end,
        color: suggestion.color,
        text: suggestion.text,
        isSuggestion: true
      }));
    } catch (error) {
      console.error("Failed to load suggested highlights:", error);
      // Silently fail - suggested highlights are optional
    } finally {
      loadingSuggestedHighlights = false;
    }
  }

  async function handleHighlightsChangeInternal(highlights) {
    if (!video) return;
    
    try {
      // Update the transcript player highlights (local state)
      transcriptPlayerHighlights = [...highlights];
      
      // Call the parent's handler
      if (onHighlightsChange) {
        await onHighlightsChange(highlights);
      }
    } catch (err) {
      console.error("Failed to save highlights:", err);
      toast.error("Failed to save highlights", {
        description: "An error occurred while saving your highlights"
      });
    }
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, '0')}`;
  }

  async function copyTranscript() {
    if (video?.transcription) {
      await navigator.clipboard.writeText(video.transcription);
      toast.success("Copied to clipboard");
    }
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
      videoId: video?.id
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
      const modelToSave = selectedModel === "custom" ? customModelValue : selectedModel;
      await SaveProjectHighlightAISettings(projectId, {
        aiModel: modelToSave,
        aiPrompt: customPrompt,
      });
      console.log("✅ Saved AI settings:", { model: modelToSave, prompt: customPrompt });

      console.log("🤖 Calling SuggestHighlightsWithAI...", {
        projectId,
        videoId: video.id,
        prompt: customPrompt || "default"
      });

      // Call the AI highlight suggestion API
      const suggestions = await SuggestHighlightsWithAI(
        projectId,
        video.id,
        customPrompt || ""
      );

      console.log("📝 Raw AI suggestions received:", suggestions);
      console.log("📝 Suggestions type:", typeof suggestions, Array.isArray(suggestions));

      // AI suggestions already come with timestamps from backend
      const newSuggestions = suggestions.map((suggestion, index) => {
        console.log(`🔄 Processing suggestion ${index}:`, suggestion);
        
        const converted = {
          id: suggestion.id,
          start: suggestion.start, // Already in timestamp format
          end: suggestion.end,     // Already in timestamp format
          color: suggestion.color,
          text: suggestion.text,
          isSuggestion: true
        };
        
        console.log(`✨ Converted suggestion ${index}:`, converted);
        return converted;
      });

      console.log("🎯 AI generation complete, reloading from database");
      
      // Reload suggested highlights from database
      await loadSuggestedHighlights();
      
      toast.success(`Generated ${suggestions.length} AI highlight suggestions!`);
    } catch (error) {
      console.error("💥 AI highlight suggestion error:", error);
      console.error("💥 Error details:", {
        message: error.message,
        stack: error.stack,
        name: error.name
      });
      toast.error("Failed to generate highlight suggestions", {
        description: error.message || "An error occurred while generating suggestions"
      });
    } finally {
      // Clear loading state for this specific video
      aiSuggestLoadingMap.delete(video.id);
      aiSuggestLoadingMap = new Map(aiSuggestLoadingMap); // Trigger reactivity
      console.log("✅ Setting loading state to false for video:", video.id);
    }
  }

  // Accept a suggested highlight
  async function acceptSuggestedHighlight(suggestionId) {
    const suggestion = suggestedHighlights.find(s => s.id === suggestionId);
    if (!suggestion || !video) return;

    try {
      // Get all available colors
      const availableColors = [
        '#FFEB3B', '#FF9800', '#F44336', '#E91E63', '#9C27B0',
        '#673AB7', '#3F51B5', '#2196F3', '#03A9F4', '#00BCD4',
        '#009688', '#4CAF50', '#8BC34A', '#CDDC39', '#FFC107'
      ];
      
      // Get used colors from existing highlights
      const usedColors = new Set(transcriptPlayerHighlights?.map(h => h.color) || []);
      
      // Find an available color
      let color = availableColors.find(c => !usedColors.has(c)) || availableColors[0];
      
      // Create new highlight with proper color
      const newHighlight = {
        id: `highlight_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        start: suggestion.start,
        end: suggestion.end,
        color: color
      };

      // Add to existing highlights
      const updatedHighlights = [...transcriptPlayerHighlights, newHighlight];
      
      // Update local state immediately for responsiveness
      transcriptPlayerHighlights = updatedHighlights;
      
      // Update video object to keep it in sync
      video = {
        ...video,
        highlights: updatedHighlights
      };
      
      // Remove from suggestions locally
      const remainingSuggestions = suggestedHighlights.filter(s => s.id !== suggestionId);
      suggestedHighlights = remainingSuggestions;
      
      // Update suggested highlights in database
      await UpdateVideoClipSuggestedHighlights(video.id, remainingSuggestions);
      
      // Save the accepted highlight to database
      if (onHighlightsChange) {
        await onHighlightsChange(updatedHighlights);
      }
      
      toast.success("Highlight suggestion accepted!");
    } catch (err) {
      console.error("Failed to accept suggestion:", err);
      toast.error("Failed to accept suggestion");
    }
  }

  // Reject a suggested highlight
  async function rejectSuggestedHighlight(suggestionId) {
    if (!video) return;
    
    try {
      // Remove from suggestions locally
      const remainingSuggestions = suggestedHighlights.filter(s => s.id !== suggestionId);
      suggestedHighlights = remainingSuggestions;
      
      // Update suggested highlights in database
      await UpdateVideoClipSuggestedHighlights(video.id, remainingSuggestions);
      
      toast.success("Highlight suggestion rejected");
    } catch (err) {
      console.error("Failed to reject suggestion:", err);
      toast.error("Failed to reject suggestion");
      // Restore the suggestion on error
      suggestedHighlights = [...suggestedHighlights, suggestedHighlights.find(s => s.id === suggestionId)];
    }
  }

  // Accept all suggested highlights
  async function acceptAllSuggestions() {
    if (suggestedHighlights.length === 0 || !video) return;

    try {
      // Get all available colors
      const availableColors = [
        '#FFEB3B', '#FF9800', '#F44336', '#E91E63', '#9C27B0',
        '#673AB7', '#3F51B5', '#2196F3', '#03A9F4', '#00BCD4',
        '#009688', '#4CAF50', '#8BC34A', '#CDDC39', '#FFC107'
      ];
      
      // Get used colors from existing highlights
      const usedColors = new Set(transcriptPlayerHighlights?.map(h => h.color) || []);
      
      // Convert all suggestions to regular highlights
      const newHighlights = suggestedHighlights.map((suggestion, index) => {
        // Find an available color
        let color = availableColors.find(c => !usedColors.has(c)) || availableColors[index % availableColors.length];
        usedColors.add(color);
        
        return {
          id: `highlight_${Date.now()}_${index}_${Math.random().toString(36).substr(2, 9)}`,
          start: suggestion.start,
          end: suggestion.end,
          color: color
        };
      });

      // Add to existing highlights
      const updatedHighlights = [...transcriptPlayerHighlights, ...newHighlights];
      
      // Update local state immediately
      transcriptPlayerHighlights = updatedHighlights;
      
      // Update video object to keep it in sync
      video = {
        ...video,
        highlights: updatedHighlights
      };
      
      // Clear suggestions locally
      suggestedHighlights = [];
      
      // Update suggested highlights in database (empty array)
      await UpdateVideoClipSuggestedHighlights(video.id, []);
      
      // Save all accepted highlights to database
      if (onHighlightsChange) {
        await onHighlightsChange(updatedHighlights);
      }
      
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
      
      toast.success(`Rejected ${count} highlight suggestion${count === 1 ? '' : 's'}`);
    } catch (err) {
      console.error("Failed to reject all suggestions:", err);
      toast.error("Failed to reject all suggestions");
    }
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[1200px] max-h-[90vh] flex flex-col">
    <DialogHeader>
      <DialogTitle>Video Transcript</DialogTitle>
      <DialogDescription>
        {#if video}
          Transcript for {video.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    <ScrollArea class="h-[70vh]">
      {#snippet children()}
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 pr-4 pb-4">
          <!-- Video Player Column -->
          <div class="space-y-4">
            {#if video}
              <!-- Video Player -->
              <div class="bg-background border rounded-lg p-4">
                <h3 class="font-medium mb-3">Video Preview</h3>
                <div class="aspect-video">
                  <EtroVideoPlayer 
                    highlights={formattedTranscriptHighlights}
                    projectId={projectId}
                    enableEyeButton={false}
                    enableReordering={false}
                  />
                </div>
              </div>
            {/if}
          </div>
          
          <!-- Transcript Column -->
          <div class="space-y-4">
            {#if video}
              <!-- Transcript content with tabs -->
              {#if video.transcription}
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <h3 class="font-medium">Transcript</h3>
                    <div class="flex gap-2 items-center">
                      {#if video.transcriptionLanguage}
                        <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                          {video.transcriptionLanguage.toUpperCase()}
                        </span>
                      {/if}
                      {#if video.transcriptionDuration}
                        <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                          {formatTimestamp(video.transcriptionDuration)}
                        </span>
                      {/if}
                      <Button 
                        variant="outline" 
                        size="sm"
                        onclick={suggestHighlightsInline}
                        class="text-xs"
                        disabled={!video.transcription || aiSuggestLoadingMap.get(video.id)}
                      >
                        <Sparkles class="w-3 h-3 mr-1" />
                        {aiSuggestLoadingMap.get(video.id) ? "AI Analyzing..." : "AI Suggest"}
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onclick={copyTranscript}
                        class="text-xs"
                      >
                        <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        Copy
                      </Button>
                    </div>
                  </div>

                  <!-- AI Settings -->
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

                  <!-- Bulk suggestion actions -->
                  {#if suggestedHighlights.length > 0}
                    <div class="flex items-center justify-between p-3 bg-secondary/30 rounded-lg">
                      <span class="text-sm text-muted-foreground">
                        {suggestedHighlights.length} AI suggestion{suggestedHighlights.length === 1 ? '' : 's'}
                      </span>
                      <div class="flex gap-2">
                        <Button 
                          variant="outline" 
                          size="sm"
                          onclick={acceptAllSuggestions}
                          class="text-xs"
                        >
                          <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                          </svg>
                          Accept All
                        </Button>
                        <Button 
                          variant="outline" 
                          size="sm"
                          onclick={rejectAllSuggestions}
                          class="text-xs"
                        >
                          <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                          </svg>
                          Reject All
                        </Button>
                      </div>
                    </div>
                  {/if}

                  <Tabs value="full-text" class="w-full">
                    <TabsList class="grid w-full grid-cols-2">
                      <TabsTrigger value="full-text">Full Text</TabsTrigger>
                      <TabsTrigger value="word-by-word" disabled={!video.transcriptionWords || video.transcriptionWords.length === 0}>
                        Word by Word
                      </TabsTrigger>
                    </TabsList>
                    
                    <TabsContent value="full-text" class="space-y-3">
                      <ScrollArea class="h-80 bg-background border rounded-lg">
                        {#snippet children()}
                          <div class="p-4 text-sm leading-relaxed">
                            <TextHighlighter 
                              text={video.transcription} 
                              words={video.transcriptionWords || []} 
                              initialHighlights={transcriptPlayerHighlights}
                              {suggestedHighlights}
                              onHighlightsChange={handleHighlightsChangeInternal}
                              onSuggestionAccept={acceptSuggestedHighlight}
                              onSuggestionReject={rejectSuggestedHighlight}
                            />
                          </div>
                        {/snippet}
                      </ScrollArea>
                      <div class="text-xs text-muted-foreground">
                        Character count: {video.transcription.length}
                      </div>
                    </TabsContent>
                    
                    <TabsContent value="word-by-word" class="space-y-3">
                      {#if video.transcriptionWords && video.transcriptionWords.length > 0}
                        <ScrollArea class="h-80 bg-background border rounded-lg">
                          {#snippet children()}
                            <div class="p-4 space-y-1">
                              {#each video.transcriptionWords as word, index}
                                <div class="flex items-center gap-3 p-2 hover:bg-secondary/30 rounded-md group">
                                  <div class="flex-shrink-0 text-xs text-muted-foreground font-mono bg-secondary px-2 py-1 rounded">
                                    {formatTimestamp(word.start)}
                                  </div>
                                  <div class="flex-1">
                                    <span class="text-sm">{word.word.trim()}</span>
                                  </div>
                                  <div class="flex-shrink-0 text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                                    {(word.end - word.start).toFixed(1)}s
                                  </div>
                                </div>
                              {/each}
                            </div>
                          {/snippet}
                        </ScrollArea>
                        <div class="text-xs text-muted-foreground flex-shrink-0">
                          Word count: {video.transcriptionWords.length}
                        </div>
                      {:else}
                        <div class="flex-1 flex items-center justify-center text-muted-foreground">
                          <div class="text-center">
                            <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <p class="text-lg font-medium">No word-level timing available</p>
                            <p class="text-sm">Word timestamps weren't generated for this transcription.</p>
                          </div>
                        </div>
                      {/if}
                    </TabsContent>
                  </Tabs>
                </div>
              {:else}
                <div class="flex-1 flex items-center justify-center text-muted-foreground">
                  <div class="text-center">
                    <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <p class="text-lg font-medium">No transcript available</p>
                    <p class="text-sm">This video hasn't been transcribed yet.</p>
                  </div>
                </div>
              {/if}
            {/if}
          </div>
        </div>
      {/snippet}
    </ScrollArea>
    
    <!-- Fixed footer buttons -->
    <div class="flex justify-end gap-2 pt-1.5 border-t flex-shrink-0">
      <Button variant="outline" onclick={() => open = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>