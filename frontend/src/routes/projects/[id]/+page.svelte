<script>
  import { Button } from "$lib/components/ui/button";
  import {
    Tabs,
    TabsList,
    TabsTrigger,
    TabsContent,
  } from "$lib/components/ui/tabs";
  import {
    GetProjectByID,
    UpdateProject,
    DeleteProject,
    UpdateProjectActiveTab,
  } from "$lib/wailsjs/go/main/App";
  import { onMount, onDestroy } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";
  import ProjectHighlights from "$lib/components/ProjectHighlights.svelte";
  import {
    orderedHighlights,
    highlightsLoading,
    loadProjectHighlights,
    clearHighlights,
    rawHighlights,
    highlightOrder,
    currentProjectId,
  } from "$lib/stores/projectHighlights.js";
  import { 
    connectToProject, 
    disconnect, 
    onRealtimeEvent, 
    EVENT_TYPES 
  } from "$lib/stores/realtime.js";
  import { get } from "svelte/store";
  import ThemeSwitcher from "$lib/components/ui/theme-switcher/theme-switcher.svelte";
  import ProjectDetails from "$lib/components/ProjectDetails.svelte";
  import ExportVideo from "$lib/components/ExportVideo.svelte";
  import ProjectInfo from "$lib/components/ProjectInfo.svelte";
  import VideoClips from "$lib/components/VideoClips.svelte";
  import { AIChatbot } from "$lib/components/chatbot";
  import { CHATBOT_ENDPOINTS } from "$lib/constants/chatbot.js";
  import {
    Info,
    Film,
    Clock,
    Upload,
    ArrowLeft,
    RefreshCw,
  } from "@lucide/svelte";

  let project = $state(null);
  let loading = $state(false);
  let error = $state("");

  // Video clips state
  let videoClips = $state([]);
  let dragActive = $state(false);

  // Highlights component reference
  let projectHighlightsComponent = $state(null);

  // Highlights state managed at parent level
  let highlights = $state([]);
  let highlightsLoaded = $state(false);

  // Tabs state
  let activeTab = $state("clips");
  let debounceTimer = null;

  // Chatbot configuration - easy to enable/disable for different tabs
  const chatbotConfig = {
    enabled: true,
    timeline: true,     // Show in Timeline tab
    clips: false,       // Could enable for Clips tab later
    info: false,        // Could enable for Info tab later  
    export: false       // Could enable for Export tab later
  };

  // To add chatbot to other tabs, follow this pattern:
  // 1. Set the tab to true in chatbotConfig above
  // 2. Add this block inside the TabsContent for that tab:
  /*
    {#if chatbotConfig.enabled && chatbotConfig.TABNAME}
      <AIChatbot 
        endpointId={CHATBOT_ENDPOINTS.APPROPRIATE_ENDPOINT}
        {projectId}
        contextData={chatbotContextData}
        position="floating"
        size="default"
      />
    {/if}
  */

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  // Store unsubscribe functions for real-time events
  let realtimeUnsubscribers = [];

  onMount(async () => {
    await loadProject();
    await loadHighlights();
    
    // Set up real-time event handlers
    setupRealtimeHandlers();
  });

  onDestroy(() => {
    // Clean up real-time event handlers
    realtimeUnsubscribers.forEach(unsubscribe => unsubscribe());
    realtimeUnsubscribers = [];
    
    // Clean up highlights store
    clearHighlights();

    // Clean up tab save timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
  });

  async function loadProject() {
    if (!projectId || isNaN(projectId)) {
      error = "Invalid project ID";
      return;
    }

    try {
      loading = true;
      error = "";
      project = await GetProjectByID(projectId);

      // Set the active tab from project settings
      if (project && project.activeTab) {
        activeTab = project.activeTab;
      }
    } catch (err) {
      console.error("Failed to load project:", err);
      error = "Failed to load project";
    } finally {
      loading = false;
    }
  }

  async function loadHighlights() {
    if (!projectId || isNaN(projectId)) return;

    try {
      await loadProjectHighlights(projectId);
      highlightsLoaded = true;
    } catch (err) {
      console.error("Failed to load highlights:", err);
    }
  }

  function setupRealtimeHandlers() {
    // Set up real-time event handlers for highlights updates
    const unsubscribeUpdated = onRealtimeEvent(EVENT_TYPES.HIGHLIGHTS_UPDATED, (data) => {
      console.log('Received highlights update:', data);
      
      // Only process if this update is for the current project
      if (data.projectId === projectId?.toString()) {
        try {
          // Access highlights from the nested data structure
          const highlights = data.data?.highlights;
          if (highlights && Array.isArray(highlights)) {
            // Flatten highlights from all videos into individual highlight objects
            const flattenedHighlights = [];
            for (const videoHighlights of highlights) {
              for (const highlight of videoHighlights.highlights) {
                flattenedHighlights.push({
                  ...highlight,
                  videoClipId: videoHighlights.videoClipId,
                  videoClipName: videoHighlights.videoClipName,
                  filePath: videoHighlights.filePath,
                  videoDuration: videoHighlights.duration
                });
              }
            }
            
            console.log('Updating highlights store with real-time data:', flattenedHighlights.length, 'highlights');
            rawHighlights.set(flattenedHighlights);
          }
        } catch (error) {
          console.error('Error processing real-time highlights update:', error);
        }
      }
    });

    const unsubscribeReordered = onRealtimeEvent(EVENT_TYPES.HIGHLIGHTS_REORDERED, (data) => {
      console.log('Received highlights reorder:', data);
      
      // Only process if this update is for the current project
      if (data.projectId === projectId?.toString()) {
        try {
          // Access order from the nested data structure
          const newOrder = data.data?.newOrder;
          if (newOrder && Array.isArray(newOrder)) {
            console.log('Updating highlight order with real-time data:', newOrder.length, 'items');
            highlightOrder.set(newOrder);
          }
        } catch (error) {
          console.error('Error processing real-time highlights reorder:', error);
        }
      }
    });

    const unsubscribeDeleted = onRealtimeEvent(EVENT_TYPES.HIGHLIGHTS_DELETED, (data) => {
      console.log('Received highlights deletion:', data);
      
      // Only process if this update is for the current project
      if (data.projectId === projectId?.toString()) {
        try {
          // Refresh highlights from the backend to ensure consistency
          loadProjectHighlights(projectId);
        } catch (error) {
          console.error('Error processing real-time highlights deletion:', error);
        }
      }
    });

    // Store unsubscribe functions for cleanup
    realtimeUnsubscribers = [unsubscribeUpdated, unsubscribeReordered, unsubscribeDeleted];
  }

  // Watch for changes in the highlights store
  $effect(() => {
    highlights = $orderedHighlights;
  });

  // Prepare context data for chatbot
  let chatbotContextData = $derived({
    highlights: highlights,
    order: highlights.map(h => h.id || h), // Extract IDs for order
    projectInfo: {
      id: projectId,
      name: project?.name || '',
      description: project?.description || '',
      totalHighlights: highlights.length,
      videoClipsCount: videoClips.length
    },
    videoClips: videoClips.map(clip => ({
      id: clip.id,
      name: clip.name,
      duration: clip.duration,
      highlightCount: highlights.filter(h => h.videoClipId === clip.id).length
    }))
  });

  function handleProjectUpdate(updatedProject) {
    project = updatedProject;
  }

  function handleProjectDelete() {
    goto("/");
  }



  function goBack() {
    goto("/");
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, "0")}`;
  }


  async function saveActiveTab(tab) {
    if (!projectId || isNaN(projectId)) return;

    try {
      await UpdateProjectActiveTab(projectId, tab);
    } catch (err) {
      console.error("Failed to save active tab:", err);
      // Don't show error toast for this - it's not critical to user experience
    }
  }

  // Watch for active tab changes
  $effect(() => {
    // Clear any existing timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    // Only save if project is loaded
    if (project && activeTab) {
      // Debounce the save to avoid too many API calls
      debounceTimer = setTimeout(() => {
        saveActiveTab(activeTab);
      }, 500);
    }
  });
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="w-full mx-auto space-y-6">
    <!-- Header with back button and theme switcher -->
    <div class="flex items-center justify-between">
      <Button
        variant="ghost"
        size="sm"
        onclick={goBack}
        class="flex items-center gap-2"
      >
        <ArrowLeft class="w-4 h-4" />
        Projects
      </Button>
      <ThemeSwitcher />
    </div>

    <!-- Error display -->
    {#if error}
      <div class="border border-destructive rounded p-4 text-destructive">
        <p class="font-medium">Error</p>
        <p class="text-sm">{error}</p>
        <Button variant="outline" size="sm" class="mt-2 flex items-center gap-2" onclick={loadProject}>
          <RefreshCw class="w-4 h-4" />
          Try Again
        </Button>
      </div>
    {/if}

    <!-- Loading state -->
    {#if loading && !project}
      <div class="text-center py-12">
        <p>Loading project...</p>
      </div>
    {:else if project}
      <!-- Project title only -->
      <div class="mb-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-semibold">{project.name}</h1>
            {#if project.description}
              <p class="text-muted-foreground mt-1">
                {project.description}
              </p>
            {/if}
          </div>
        </div>
      </div>

      <!-- Main content with tabs -->
      <div class="border rounded">
        <div class="p-6">
          <Tabs bind:value={activeTab} class="w-full">
            <TabsList class="grid w-full grid-cols-4">
              <TabsTrigger value="info" class="flex items-center gap-2">
                <Info class="w-4 h-4" />
                Info
              </TabsTrigger>
              <TabsTrigger value="clips" class="flex items-center gap-2">
                <Film class="w-4 h-4" />
                Clips ({videoClips.length})
              </TabsTrigger>
              <TabsTrigger value="timeline" class="flex items-center gap-2">
                <Clock class="w-4 h-4" />
                Timeline
              </TabsTrigger>
              <TabsTrigger value="export" class="flex items-center gap-2">
                <Upload class="w-4 h-4" />
                Export
              </TabsTrigger>
            </TabsList>

            <!-- Project Info Tab -->
            <TabsContent value="info" class="mt-6">
              <ProjectInfo
                {project}
                {videoClips}
                {highlights}
                onProjectUpdate={handleProjectUpdate}
                onProjectDelete={handleProjectDelete}
              />
            </TabsContent>

            <!-- Video Clips Tab -->
            <TabsContent value="clips" class="mt-6">
              <VideoClips
                {projectId}
                {highlights}
                bind:dragActive
                bind:videoClips
                onHighlightsChange={loadHighlights}
              />
            </TabsContent>

            <!-- Timeline Tab -->
            <TabsContent value="timeline" class="mt-6">
              {#if project && highlightsLoaded}
                <ProjectHighlights
                  bind:this={projectHighlightsComponent}
                  {projectId}
                  {highlights}
                  loading={$highlightsLoading}
                />
                
                <!-- Chatbot for Timeline tab -->
                <!-- TODO: Refactor this to use the new AI Actions menu instead -->
                <!-- {#if chatbotConfig.enabled && chatbotConfig.timeline}
                  <AIChatbot 
                    endpointId={CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING}
                    {projectId}
                    contextData={chatbotContextData}
                    position="floating"
                    size="default"
                  />
                {/if} -->
              {:else}
                <div class="text-center py-8 text-muted-foreground">
                  <p>Timeline will appear here once highlights are loaded</p>
                </div>
              {/if}
            </TabsContent>

            <!-- Export Tab -->
            <TabsContent value="export" class="mt-6">
              <ExportVideo {project} {projectId} {videoClips} />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    {:else if !loading}
      <!-- Project not found -->
      <div class="text-center py-12">
        <p>Project not found</p>
        <p class="text-sm text-muted-foreground">The project you're looking for doesn't exist</p>
        <Button size="sm" class="mt-4" onclick={goBack}>Go Back</Button>
      </div>
    {/if}
  </div>
</main>

