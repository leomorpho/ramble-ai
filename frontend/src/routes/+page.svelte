<script>
  import { Button } from "$lib/components/ui/button";
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogFooter, 
    DialogHeader, 
    DialogTitle, 
    DialogTrigger 
  } from "$lib/components/ui/dialog";
  import { ThemeSwitcher } from "$lib/components/ui/theme-switcher";
  import { Settings, Lightbulb, Video, Volume2, HelpCircle, Target, FileText, Brain, RotateCcw, Upload, Zap, Plus, RefreshCw, BarChart } from "@lucide/svelte";
  import { CreateProject, GetProjects, GetVideoClipsByProject, GetRambleAIApiKey } from "$lib/wailsjs/go/main/App";
  import OnboardingDialog from "$lib/components/OnboardingDialog.svelte";
  import { BannerList } from "$lib/components/ui/banner";
  import { fetchBanners } from "$lib/services/bannerService.js";
  import { onMount } from "svelte";

  let projects = $state([]);
  let projectThumbnails = $state({}); // Map of projectId to thumbnailUrl
  let dialogOpen = $state(false);
  let projectName = $state("");
  let projectDescription = $state("");
  let loading = $state(false);
  let error = $state("");
  let onboardingDialogOpen = $state(false);
  let banners = $state([]);
  let hasApiKey = $state(false);

  // Load projects and banners on mount
  onMount(async () => {
    await loadProjects();
    await loadBanners();
    await checkApiKey();
  });

  async function loadProjects() {
    try {
      loading = true;
      error = "";
      const result = await GetProjects();
      projects = result || [];
      
      // Load thumbnails for each project
      await loadProjectThumbnails();
    } catch (err) {
      console.error("Failed to load projects:", err);
      error = "Failed to load projects";
    } finally {
      loading = false;
    }
  }

  async function loadProjectThumbnails() {
    const thumbnails = {};
    
    // Load thumbnails for all projects in parallel
    const thumbnailPromises = projects.map(async (project) => {
      try {
        const videoClips = await GetVideoClipsByProject(project.id);
        if (videoClips && videoClips.length > 0 && videoClips[0].thumbnailUrl) {
          thumbnails[project.id] = videoClips[0].thumbnailUrl;
        }
      } catch (err) {
        console.warn(`Failed to load thumbnail for project ${project.id}:`, err);
      }
    });
    
    await Promise.all(thumbnailPromises);
    projectThumbnails = thumbnails;
  }

  async function createProject() {
    if (!projectName.trim()) return;
    
    try {
      loading = true;
      error = "";
      
      const newProject = await CreateProject(
        projectName.trim(), 
        projectDescription.trim()
      );
      
      // Add the new project to the list
      projects.push(newProject);
      
      // Reset form
      projectName = "";
      projectDescription = "";
      dialogOpen = false;
    } catch (err) {
      console.error("Failed to create project:", err);
      error = "Failed to create project";
    } finally {
      loading = false;
    }
  }

  async function loadBanners() {
    try {
      // Get API key if available
      let apiKey = null;
      try {
        apiKey = await GetRambleAIApiKey();
      } catch (err) {
        console.log("No API key available for authenticated banners");
      }

      // Fetch banners (public + authenticated if API key available)
      banners = await fetchBanners(apiKey);
    } catch (err) {
      console.error("Failed to load banners:", err);
      // Don't show banner errors to user, just log them
    }
  }

  async function checkApiKey() {
    try {
      const apiKey = await GetRambleAIApiKey();
      hasApiKey = !!(apiKey && apiKey.trim());
    } catch (err) {
      console.log("No API key configured");
      hasApiKey = false;
    }
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-semibold">Projects</h1>
      
      <div class="flex items-center gap-2">
        <ThemeSwitcher />
        
        <Button variant="ghost" size="icon" class="h-9 w-9" title="Help & Setup Guide" onclick={() => onboardingDialogOpen = true}>
          <HelpCircle class="h-4 w-4" />
        </Button>
        
        {#if hasApiKey}
          <Button variant="ghost" size="icon" class="h-9 w-9" title="Usage Statistics" asChild>
            <a href="/usage">
              <BarChart class="h-4 w-4" />
            </a>
          </Button>
        {/if}
        
        <Button variant="ghost" size="icon" class="h-9 w-9" title="Settings" asChild>
          <a href="/settings">
            <Settings class="h-4 w-4" />
          </a>
        </Button>
        
        <Dialog bind:open={dialogOpen}>
          <DialogTrigger>
            <Button size="sm" class="flex items-center gap-2">
              <Plus class="w-4 h-4" />
              New Project
            </Button>
          </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Project</DialogTitle>
          </DialogHeader>
          <div class="space-y-4">
            <div>
              <label for="name" class="block text-sm mb-1">Name</label>
              <input
                id="name"
                bind:value={projectName}
                class="w-full px-3 py-2 border border-input rounded bg-background"
                placeholder="Project name"
              />
            </div>
            <div>
              <label for="description" class="block text-sm mb-1">Description</label>
              <textarea
                id="description"
                bind:value={projectDescription}
                class="w-full px-3 py-2 border border-input rounded bg-background resize-none"
                rows="2"
                placeholder="Project description"
              ></textarea>
            </div>
          </div>
          <DialogFooter>
            <Button onclick={createProject} disabled={!projectName.trim() || loading}>
              {loading ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>

    <!-- Banners -->
    <BannerList banners={banners} />

    {#if error}
      <div class="border border-destructive rounded p-4 text-destructive">
        <p class="font-medium">Error</p>
        <p class="text-sm">{error}</p>
        <Button variant="outline" size="sm" class="mt-2 flex items-center gap-2" onclick={loadProjects}>
          <RefreshCw class="w-4 h-4" />
          Try Again
        </Button>
      </div>
    {/if}

    {#if loading && projects.length === 0}
      <div class="text-center py-12">
        <p>Loading projects...</p>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-12 space-y-6">
        <div>
          <h2 class="text-xl font-semibold mb-2">Welcome to RambleAI</h2>
          <p class="text-muted-foreground">
            Transform your talking head videos into polished content.
          </p>
        </div>

        <div class="max-w-sm mx-auto border rounded p-4 space-y-3">
          <h3 class="font-medium">Ready to get started?</h3>
          <p class="text-sm text-muted-foreground">Set up your API keys first</p>
          
          <div class="flex gap-2">
            <Button size="sm" onclick={() => onboardingDialogOpen = true} class="flex-1">
              Setup Guide
            </Button>
            <Button variant="outline" size="sm" asChild class="flex-1">
              <a href="/settings">Settings</a>
            </Button>
          </div>
        </div>
      </div>
    {:else}
      <div class="space-y-4">
        {#each projects as project (project.id)}
          <a 
            href="/projects/{project.id}" 
            class="block border rounded overflow-hidden hover:bg-card h-32"
          >
            <div class="flex h-full">
              <!-- Project thumbnail -->
              {#if projectThumbnails[project.id]}
                <img 
                  src={projectThumbnails[project.id]} 
                  alt="Project thumbnail for {project.name}"
                  class="w-48 h-full object-cover bg-muted flex-shrink-0"
                  loading="lazy"
                />
              {:else}
                <div class="w-48 h-full bg-muted flex-shrink-0 flex items-center justify-center">
                  <svg class="w-8 h-8 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </div>
              {/if}

              <!-- Project details -->
              <div class="flex-1 min-w-0 p-4 flex items-center">
                <div class="flex-1 min-w-0">
                  <h3 class="font-medium text-lg">{project.name}</h3>
                  {#if project.description}
                    <p class="text-sm text-muted-foreground mt-1 line-clamp-2">{project.description}</p>
                  {/if}
                  <p class="text-xs text-muted-foreground mt-3">{project.createdAt}</p>
                </div>
                
                <!-- Arrow icon -->
                <svg class="w-5 h-5 text-muted-foreground flex-shrink-0 ml-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </div>
          </a>
        {/each}
      </div>
    {/if}

  </div>

  <!-- Onboarding Dialog -->
  <OnboardingDialog bind:open={onboardingDialogOpen} />
</main>
