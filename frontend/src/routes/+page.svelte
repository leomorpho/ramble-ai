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

  let projects = $state([]);
  let dialogOpen = $state(false);
  let projectName = $state("");
  let projectDescription = $state("");

  function createProject() {
    if (projectName.trim()) {
      const newProject = {
        id: Date.now(),
        name: projectName.trim(),
        description: projectDescription.trim(),
        createdAt: new Date().toLocaleDateString()
      };
      projects.push(newProject);
      projectName = "";
      projectDescription = "";
      dialogOpen = false;
    }
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-primary">Projects</h1>
      
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
            <Button onclick={createProject} disabled={!projectName.trim()}>
              Create Project
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>

    {#if projects.length === 0}
      <div class="text-center py-12 text-muted-foreground">
        <p class="text-lg">No projects yet</p>
        <p class="text-sm">Create your first project to get started</p>
      </div>
    {:else}
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {#each projects as project (project.id)}
          <div class="bg-card text-card-foreground p-6 rounded-lg border shadow-sm">
            <h3 class="text-xl font-semibold mb-2">{project.name}</h3>
            {#if project.description}
              <p class="text-muted-foreground mb-4">{project.description}</p>
            {/if}
            <p class="text-sm text-muted-foreground">Created: {project.createdAt}</p>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</main>
