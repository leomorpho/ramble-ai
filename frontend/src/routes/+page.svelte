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
  import { Settings, Lightbulb, Video, Volume2, HelpCircle } from "@lucide/svelte";
  import { CreateProject, GetProjects, GetVideoClipsByProject } from "$lib/wailsjs/go/main/App";
  import OnboardingDialog from "$lib/components/OnboardingDialog.svelte";
  import { onMount } from "svelte";

  let projects = $state([]);
  let projectThumbnails = $state({}); // Map of projectId to thumbnailUrl
  let dialogOpen = $state(false);
  let projectName = $state("");
  let projectDescription = $state("");
  let loading = $state(false);
  let error = $state("");
  let onboardingDialogOpen = $state(false);

  // Load projects on mount
  onMount(async () => {
    await loadProjects();
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
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-primary">Projects</h1>
      
      <div class="flex items-center gap-2">
        <ThemeSwitcher />
        
        <Button variant="ghost" size="icon" class="h-9 w-9" title="Help & Setup Guide" disabled={false} onclick={() => onboardingDialogOpen = true}>
          <HelpCircle class="h-4 w-4" />
        </Button>
        
        <Button variant="ghost" size="icon" class="h-9 w-9" title="Settings" asChild>
          <a href="/settings">
            <Settings class="h-4 w-4" />
          </a>
        </Button>
        
        <Dialog bind:open={dialogOpen}>
          <DialogTrigger>
            <Button>Create New Project</Button>
          </DialogTrigger>
        <DialogContent class="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Create New Project</DialogTitle>
            <DialogDescription>
              Enter the details for your new project.
            </DialogDescription>
          </DialogHeader>
          <div class="grid gap-4 py-4">
            <div class="grid grid-cols-4 items-center gap-4">
              <label for="name" class="text-right">Name</label>
              <input
                id="name"
                bind:value={projectName}
                class="col-span-3 px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Enter project name"
              />
            </div>
            <div class="grid grid-cols-4 items-center gap-4">
              <label for="description" class="text-right">Description</label>
              <textarea
                id="description"
                bind:value={projectDescription}
                class="col-span-3 px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                rows="3"
                placeholder="Enter project description"
              ></textarea>
            </div>
          </div>
          <DialogFooter>
            <Button onclick={createProject} disabled={!projectName.trim() || loading}>
              {loading ? "Creating..." : "Create Project"}
            </Button>
          </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>

    {#if error}
      <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4">
        <p class="font-medium">Error</p>
        <p class="text-sm">{error}</p>
        <Button variant="outline" size="sm" class="mt-2" onclick={loadProjects}>
          Try Again
        </Button>
      </div>
    {/if}

    {#if loading && projects.length === 0}
      <div class="text-center py-12 text-muted-foreground">
        <p class="text-lg">Loading projects...</p>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-12 space-y-8">
        <!-- Welcome Section -->
        <div class="space-y-4">
          <h2 class="text-2xl font-bold text-foreground">Welcome to RambleAI!</h2>
          <p class="text-muted-foreground max-w-lg mx-auto">
            Transform your talking head videos into polished content. AI-powered preprocessing that saves 60-80% of your editing time.
          </p>
        </div>

        <!-- Onboarding Card -->
        <div class="max-w-md mx-auto bg-secondary/30 border border-border rounded-lg p-6 space-y-4">
          <div class="text-center">
            <HelpCircle class="w-12 h-12 mx-auto text-primary mb-3" />
            <h3 class="text-lg font-semibold">Ready to get started?</h3>
            <p class="text-sm text-muted-foreground">Set up your API keys to unlock these powerful features</p>
          </div>
          
          <div class="flex gap-3">
            <Button onclick={() => onboardingDialogOpen = true} disabled={false} class="flex-1">
              View Setup Guide
            </Button>
            <Button variant="outline" asChild class="flex-1">
              <a href="/settings">Go to Settings</a>
            </Button>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="max-w-4xl mx-auto grid md:grid-cols-2 lg:grid-cols-3 gap-4 text-left">
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">🎯</div>
            <h3 class="font-semibold">Smart Clip Selection</h3>
            <p class="text-sm text-muted-foreground">AI identifies the best parts of your videos automatically</p>
          </div>
          
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">📝</div>
            <h3 class="font-semibold">Speech Transcription</h3>
            <p class="text-sm text-muted-foreground">Word-perfect transcripts with precise timing</p>
          </div>
          
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">🧠</div>
            <h3 class="font-semibold">AI Script Reordering</h3>
            <p class="text-sm text-muted-foreground">Transform clips into coherent scripts using AI</p>
          </div>
          
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">🔄</div>
            <h3 class="font-semibold">Workflow Integration</h3>
            <p class="text-sm text-muted-foreground">Fits perfectly as the first step in your editing workflow</p>
          </div>
          
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">📤</div>
            <h3 class="font-semibold">Export & Handoff</h3>
            <p class="text-sm text-muted-foreground">Export optimized scripts ready for your editor</p>
          </div>
          
          <div class="bg-card border rounded-lg p-4 space-y-2">
            <div class="text-2xl">⚡</div>
            <h3 class="font-semibold">60-80% Time Savings</h3>
            <p class="text-sm text-muted-foreground">Dramatically reduce post-production time</p>
          </div>
        </div>
      </div>
    {:else}
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {#each projects as project (project.id)}
          <a 
            href="/projects/{project.id}" 
            class="bg-card text-card-foreground rounded-lg border shadow-sm hover:shadow-md transition-shadow duration-200 block group overflow-hidden"
          >
            <!-- Project thumbnail -->
            {#if projectThumbnails[project.id]}
              <div class="relative">
                <img 
                  src={projectThumbnails[project.id]} 
                  alt="Project thumbnail for {project.name}"
                  class="w-full h-32 object-cover bg-muted"
                  loading="lazy"
                />
                <!-- Play overlay -->
                <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
                  <div class="w-10 h-10 bg-white/80 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <svg class="w-5 h-5 text-black ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z"/>
                    </svg>
                  </div>
                </div>
              </div>
            {:else}
              <div class="w-full h-32 bg-muted flex items-center justify-center">
                <div class="text-center text-muted-foreground">
                  <svg class="w-8 h-8 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                  <p class="text-xs">No videos</p>
                </div>
              </div>
            {/if}

            <!-- Project details -->
            <div class="p-6">
              <div class="flex justify-between items-start mb-2">
                <h3 class="text-xl font-semibold group-hover:text-primary transition-colors duration-200">{project.name}</h3>
                <svg class="w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
              {#if project.description}
                <p class="text-muted-foreground mb-4 line-clamp-2">{project.description}</p>
              {/if}
              <div class="text-sm text-muted-foreground space-y-1">
                <p>Created: {project.createdAt}</p>
                <p class="text-xs truncate">Path: {project.path}</p>
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
