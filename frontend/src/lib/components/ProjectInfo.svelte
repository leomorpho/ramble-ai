<script>
  import ProjectDetails from "$lib/components/ProjectDetails.svelte";
  import CopyToClipboardButton from "$lib/components/CopyToClipboardButton.svelte";

  // Props
  let { project, videoClips, highlights, onProjectUpdate, onProjectDelete } = $props();

  // Calculate highlights duration
  function calculateHighlightsDuration() {
    const totalSeconds = highlights.reduce(
      (sum, highlight) => sum + (highlight.end - highlight.start),
      0
    );
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = Math.floor(totalSeconds % 60);
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  }
</script>

<div class="space-y-8">
  <!-- Project Overview Card -->
  <div>
    <div class="flex items-start justify-between mb-4">
      <ProjectDetails
        {project}
        onUpdate={onProjectUpdate}
        onDelete={onProjectDelete}
        buttonsOnly={true}
      />
    </div>

    <!-- Statistics Grid -->
    <div class="grid gap-6 md:grid-cols-3">
      <div class="text-center p-4 bg-background rounded-lg border">
        <div class="text-2xl font-bold text-primary mb-1">
          {videoClips.length}
        </div>
        <div class="text-sm text-muted-foreground">
          Video Clips
        </div>
      </div>
      <div class="text-center p-4 bg-background rounded-lg border">
        <div class="text-2xl font-bold text-primary mb-1">
          {highlights.length}
        </div>
        <div class="text-sm text-muted-foreground">
          Highlights
        </div>
      </div>
      <div class="text-center p-4 bg-background rounded-lg border">
        <div class="text-2xl font-bold text-primary mb-1">
          {calculateHighlightsDuration()}
        </div>
        <div class="text-sm text-muted-foreground">
          Highlights Duration
        </div>
      </div>
    </div>
  </div>

  <!-- Project Details Card -->
  <div class="bg-card border rounded-lg p-6">
    <div class="flex items-center gap-3 mb-4">
      <div
        class="w-10 h-10 bg-secondary/50 rounded-lg flex items-center justify-center"
      >
        <svg
          class="w-5 h-5 text-foreground"
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
      </div>
      <div>
        <h3 class="text-lg font-semibold">Project Details</h3>
        <p class="text-muted-foreground text-sm">
          Technical information and metadata
        </p>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <div class="space-y-3">
        <div
          class="flex justify-between items-center py-2 border-b border-border/50"
        >
          <span class="text-sm font-medium text-muted-foreground">Created</span>
          <span class="text-sm font-medium">
            {new Date(project.createdAt).toLocaleDateString("en-US", {
              year: "numeric",
              month: "short",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
        </div>
        <div
          class="flex justify-between items-center py-2 border-b border-border/50"
        >
          <span class="text-sm font-medium text-muted-foreground">Last Updated</span>
          <span class="text-sm font-medium">
            {new Date(project.updatedAt).toLocaleDateString("en-US", {
              year: "numeric",
              month: "short",
              day: "numeric",
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
        </div>
      </div>
      <div class="space-y-3">
        <div
          class="flex justify-between items-center py-2 border-b border-border/50"
        >
          <span class="text-sm font-medium text-muted-foreground">Project ID</span>
          <span class="text-sm font-mono bg-secondary px-2 py-1 rounded">
            {project.id}
          </span>
        </div>
        <div class="py-2 md:col-span-2">
          <div class="text-sm font-medium text-muted-foreground mb-2">
            Project Path
          </div>
          <div class="flex items-center gap-2">
            <code class="text-sm bg-secondary px-3 py-2 rounded flex-1 break-all">
              {project.path}
            </code>
            <CopyToClipboardButton text={project.path} />
          </div>
        </div>
      </div>
    </div>
  </div>
</div>