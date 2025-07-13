<script>
  import { Button } from "$lib/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "$lib/components/ui/dialog";
  import { UpdateProject, DeleteProject } from "$lib/wailsjs/go/main/App";
  import { goto } from "$app/navigation";

  let { project, onUpdate, onDelete, buttonsOnly = false } = $props();

  let editDialogOpen = $state(false);
  let deleteDialogOpen = $state(false);
  let editName = $state("");
  let editDescription = $state("");
  let loading = $state(false);
  let deleting = $state(false);
  let error = $state("");

  $effect(() => {
    if (project) {
      editName = project.name;
      editDescription = project.description;
    }
  });

  async function handleUpdateProject() {
    if (!editName.trim()) return;

    try {
      loading = true;
      error = "";

      const updatedProject = await UpdateProject(
        project.id,
        editName.trim(),
        editDescription.trim()
      );

      editDialogOpen = false;
      onUpdate?.(updatedProject);
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

      await DeleteProject(project.id);
      onDelete?.();
    } catch (err) {
      console.error("Failed to delete project:", err);
      error = "Failed to delete project";
      deleting = false;
    }
  }
</script>

{#if buttonsOnly}
  <div class="flex gap-2">
    <!-- Edit button -->
    <Dialog bind:open={editDialogOpen}>
      <DialogTrigger>
        <Button variant="outline" class="flex items-center gap-2">
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
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
            />
          </svg>
          Edit Project
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
            <label for="edit-description" class="text-right"
              >Description</label
            >
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
          <Button
            onclick={handleUpdateProject}
            disabled={!editName.trim() || loading}
          >
            {loading ? "Saving..." : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete button -->
    <Dialog bind:open={deleteDialogOpen}>
      <DialogTrigger>
        <Button variant="destructive" class="flex items-center gap-2">
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
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
            />
          </svg>
          Delete Project
        </Button>
      </DialogTrigger>
      <DialogContent class="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Delete Project</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete "{project.name}"? This action
            cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant="outline"
            onclick={() => (deleteDialogOpen = false)}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onclick={handleDeleteProject}
            disabled={deleting}
          >
            {deleting ? "Deleting..." : "Delete Project"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
{:else}
<div class="p-6">
  <div class="flex justify-between items-start mb-6">
    <div class="space-y-2">
      <h1 class="text-3xl font-bold text-primary">{project.name}</h1>
      {#if project.description}
        <p class="text-muted-foreground text-lg">
          {project.description}
        </p>
      {/if}
    </div>

    <div class="flex gap-2">
      <!-- Edit button -->
      <Dialog bind:open={editDialogOpen}>
        <DialogTrigger>
          <Button variant="outline" class="flex items-center gap-2">
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
                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
              />
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
              <label for="edit-description" class="text-right"
                >Description</label
              >
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
            <Button
              onclick={handleUpdateProject}
              disabled={!editName.trim() || loading}
            >
              {loading ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Delete button -->
      <Dialog bind:open={deleteDialogOpen}>
        <DialogTrigger>
          <Button variant="destructive" class="flex items-center gap-2">
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
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
            Delete Project
          </Button>
        </DialogTrigger>
        <DialogContent class="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete Project</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{project.name}"? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onclick={() => (deleteDialogOpen = false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onclick={handleDeleteProject}
              disabled={deleting}
            >
              {deleting ? "Deleting..." : "Delete Project"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </div>

  <!-- Project metadata -->
  <div
    class="grid grid-cols-1 md:grid-cols-2 gap-6 p-4 bg-secondary/30 rounded-lg"
  >
    <div class="space-y-2">
      <h3
        class="font-semibold text-sm text-muted-foreground uppercase tracking-wide"
      >
        Project Path
      </h3>
      <p class="font-mono text-sm bg-background px-3 py-2 rounded border">
        {project.path}
      </p>
    </div>
    <div class="space-y-2">
      <h3
        class="font-semibold text-sm text-muted-foreground uppercase tracking-wide"
      >
        Created
      </h3>
      <p class="text-sm">{project.createdAt}</p>
    </div>
    <div class="space-y-2">
      <h3
        class="font-semibold text-sm text-muted-foreground uppercase tracking-wide"
      >
        Last Updated
      </h3>
      <p class="text-sm">{project.updatedAt}</p>
    </div>
    <div class="space-y-2">
      <h3
        class="font-semibold text-sm text-muted-foreground uppercase tracking-wide"
      >
        Project ID
      </h3>
      <p class="text-sm font-mono">{project.id}</p>
    </div>
  </div>
</div>
{/if}