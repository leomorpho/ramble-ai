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
  import { CreateProject, GetProjects } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let projects = $state([]);
  let dialogOpen = $state(false);
  let projectName = $state("");
  let projectDescription = $state("");
  let loading = $state(false);
  let error = $state("");

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
    } catch (err) {
      console.error("Failed to load projects:", err);
      error = "Failed to load projects";
    } finally {
      loading = false;
    }
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
        <a href="/settings" class="p-2 text-muted-foreground hover:text-foreground transition-colors rounded-md hover:bg-secondary" title="Settings">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </a>
        
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
      <div class="text-center py-12 text-muted-foreground">
        <p class="text-lg">No projects yet</p>
        <p class="text-sm">Create your first project to get started</p>
      </div>
    {:else}
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {#each projects as project (project.id)}
          <a 
            href="/projects/{project.id}" 
            class="bg-card text-card-foreground p-6 rounded-lg border shadow-sm hover:shadow-md transition-shadow duration-200 block group"
          >
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
          </a>
        {/each}
      </div>
    {/if}
  </div>
</main>
