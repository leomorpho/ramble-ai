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
  import { GetProjectByID, UpdateProject, DeleteProject } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";

  let project = $state(null);
  let loading = $state(false);
  let error = $state("");
  let editDialogOpen = $state(false);
  let deleteDialogOpen = $state(false);
  let editName = $state("");
  let editDescription = $state("");
  let deleting = $state(false);

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  onMount(async () => {
    await loadProject();
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
      // Set edit form values
      editName = project.name;
      editDescription = project.description;
    } catch (err) {
      console.error("Failed to load project:", err);
      error = "Failed to load project";
    } finally {
      loading = false;
    }
  }

  async function handleUpdateProject() {
    if (!editName.trim()) return;
    
    try {
      loading = true;
      error = "";
      
      const updatedProject = await UpdateProject(
        projectId,
        editName.trim(), 
        editDescription.trim()
      );
      
      // Update local project data
      project = updatedProject;
      editDialogOpen = false;
    } catch (err) {
      console.error("Failed to update project:", err);
      error = "Failed to update project";
    } finally {
      loading = false;
    }
  }

  async function handleDeleteProject() {
    try {
      deleting = true;
      error = "";
      
      await DeleteProject(projectId);
      
      // Navigate back to projects list
      goto("/");
    } catch (err) {
      console.error("Failed to delete project:", err);
      error = "Failed to delete project";
      deleting = false;
    }
  }

  function goBack() {
    goto("/");
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <!-- Header with back button -->
    <div class="flex items-center gap-4">
      <Button variant="outline" onclick={goBack} class="flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back to Projects
      </Button>
    </div>

    <!-- Error display -->
    {#if error}
      <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4">
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
      <!-- Project details -->
      <div class="bg-card text-card-foreground rounded-lg border shadow-sm">
        <div class="p-6">
          <div class="flex justify-between items-start mb-6">
            <div class="space-y-2">
              <h1 class="text-3xl font-bold text-primary">{project.name}</h1>
              {#if project.description}
                <p class="text-muted-foreground text-lg">{project.description}</p>
              {/if}
            </div>
            
            <div class="flex gap-2">
              <!-- Edit button -->
              <Dialog bind:open={editDialogOpen}>
                <DialogTrigger>
                  <Button variant="outline" class="flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                    Edit
                  </Button>
                </DialogTrigger>
                <DialogContent class="sm:max-w-[425px]">
                  <DialogHeader>
                    <DialogTitle>Edit Project</DialogTitle>
                    <DialogDescription>
                      Update the project details below.
                    </DialogDescription>
                  </DialogHeader>
                  <div class="grid gap-4 py-4">
                    <div class="grid grid-cols-4 items-center gap-4">
                      <label for="edit-name" class="text-right">Name</label>
                      <input
                        id="edit-name"
                        bind:value={editName}
                        class="col-span-3 px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="Enter project name"
                      />
                    </div>
                    <div class="grid grid-cols-4 items-center gap-4">
                      <label for="edit-description" class="text-right">Description</label>
                      <textarea
                        id="edit-description"
                        bind:value={editDescription}
                        class="col-span-3 px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                        rows="3"
                        placeholder="Enter project description"
                      ></textarea>
                    </div>
                  </div>
                  <DialogFooter>
                    <Button onclick={handleUpdateProject} disabled={!editName.trim() || loading}>
                      {loading ? "Saving..." : "Save Changes"}
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>

              <!-- Delete button -->
              <Dialog bind:open={deleteDialogOpen}>
                <DialogTrigger>
                  <Button variant="destructive" class="flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    Delete
                  </Button>
                </DialogTrigger>
                <DialogContent class="sm:max-w-[425px]">
                  <DialogHeader>
                    <DialogTitle>Delete Project</DialogTitle>
                    <DialogDescription>
                      Are you sure you want to delete "{project.name}"? This action cannot be undone.
                    </DialogDescription>
                  </DialogHeader>
                  <DialogFooter>
                    <Button variant="outline" onclick={() => deleteDialogOpen = false}>
                      Cancel
                    </Button>
                    <Button variant="destructive" onclick={handleDeleteProject} disabled={deleting}>
                      {deleting ? "Deleting..." : "Delete Project"}
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          <!-- Project metadata -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 p-4 bg-secondary/30 rounded-lg">
            <div class="space-y-2">
              <h3 class="font-semibold text-sm text-muted-foreground uppercase tracking-wide">Project Path</h3>
              <p class="font-mono text-sm bg-background px-3 py-2 rounded border">{project.path}</p>
            </div>
            <div class="space-y-2">
              <h3 class="font-semibold text-sm text-muted-foreground uppercase tracking-wide">Created</h3>
              <p class="text-sm">{project.createdAt}</p>
            </div>
            <div class="space-y-2">
              <h3 class="font-semibold text-sm text-muted-foreground uppercase tracking-wide">Last Updated</h3>
              <p class="text-sm">{project.updatedAt}</p>
            </div>
            <div class="space-y-2">
              <h3 class="font-semibold text-sm text-muted-foreground uppercase tracking-wide">Project ID</h3>
              <p class="text-sm font-mono">{project.id}</p>
            </div>
          </div>

          <!-- Video Clips section (placeholder) -->
          <div class="mt-8">
            <div class="flex justify-between items-center mb-4">
              <h2 class="text-xl font-semibold">Video Clips</h2>
              <Button class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Add Video Clip
              </Button>
            </div>
            
            <!-- Placeholder for video clips -->
            <div class="text-center py-8 text-muted-foreground border-2 border-dashed border-border rounded-lg">
              <svg class="w-12 h-12 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              <p class="text-lg">No video clips yet</p>
              <p class="text-sm">Add your first video clip to get started</p>
            </div>
          </div>
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