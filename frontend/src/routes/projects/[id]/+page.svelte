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
  } from "$lib/stores/projectHighlights.js";
  import ThemeSwitcher from "$lib/components/ui/theme-switcher/theme-switcher.svelte";
  import ProjectDetails from "$lib/components/ProjectDetails.svelte";
  import ExportVideo from "$lib/components/ExportVideo.svelte";
  import ProjectInfo from "$lib/components/ProjectInfo.svelte";
  import VideoClips from "$lib/components/VideoClips.svelte";
  import {
    Info,
    Film,
    Clock,
    Upload,
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

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  onMount(async () => {
    await loadProject();
    await loadHighlights();
  });

  onDestroy(() => {
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

  // Watch for changes in the highlights store
  $effect(() => {
    highlights = $orderedHighlights;
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
        variant="outline"
        onclick={goBack}
        class="flex items-center gap-2"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 19l-7-7 7-7"
          />
        </svg>
        Back to Projects
      </Button>
      <ThemeSwitcher />
    </div>

    <!-- Error display -->
    {#if error}
      <div
        class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
      >
        <p class="font-medium">Error</p>
        <p class="text-sm">{error}</p>
        <Button variant="outline" size="sm" class="mt-2" onclick={loadProject}>
          Try Again
        </Button>
      </div>
    {/if}

    <!-- Loading state -->
    {#if loading && !project}
      <div class="text-center py-12 text-muted-foreground">
        <p class="text-lg">Loading project...</p>
      </div>
    {:else if project}
      <!-- Project title only -->
      <div class="mb-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-3xl font-bold text-primary">{project.name}</h1>
            {#if project.description}
              <p class="text-muted-foreground text-lg mt-1">
                {project.description}
              </p>
            {/if}
          </div>
        </div>
      </div>

      <!-- Main content with tabs -->
      <div class="bg-card text-card-foreground rounded-lg border shadow-sm">
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
      <div class="text-center py-12 text-muted-foreground">
        <p class="text-lg">Project not found</p>
        <p class="text-sm">The project you're looking for doesn't exist</p>
        <Button class="mt-4" onclick={goBack}>Go Back</Button>
      </div>
    {/if}
  </div>
</main>

