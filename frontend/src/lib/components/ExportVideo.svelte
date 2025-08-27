<script>
  import { Button } from "$lib/components/ui/button";
  import { Slider } from "$lib/components/ui/slider/index.js";
  import {
    SelectExportFolder,
    ExportStitchedHighlights,
    ExportIndividualHighlights,
    GetExportProgress,
    CancelExport,
    GetProjectExportJobs,
  } from "$lib/wailsjs/go/main/App";
  import { onMount, onDestroy } from "svelte";
  import { toast } from "svelte-sonner";
  import {
    Video,
    Download,
    FolderOpen,
  } from "@lucide/svelte";

  // Props
  let { project, projectId, videoClips } = $props();

  // Export state
  let exporting = $state(false);
  let exportError = $state("");
  let currentExportJob = $state(null);
  let exportProgress = $state(null);
  let progressInterval = $state(null);
  let exportHistory = $state([]);
  let showExportHistory = $state(false);
  
  // Padding settings
  let paddingValue = $state([0]); // Slider component expects an array
  let showPaddingOptions = $state(false);

  onMount(async () => {
    await loadExportHistory();
    // Check if there are any active exports to resume tracking
    await checkForActiveExports();
  });

  onDestroy(() => {
    // Clean up progress tracking if active
    if (progressInterval) {
      clearInterval(progressInterval);
    }
  });

  async function loadExportHistory() {
    try {
      const jobs = await GetProjectExportJobs(projectId);
      exportHistory = jobs || [];
    } catch (err) {
      console.error("Failed to load export history:", err);
    }
  }

  async function checkForActiveExports() {
    // Check if there are any incomplete exports for this project
    const activeExport = exportHistory.find(
      (job) => !job.isComplete && !job.isCancelled
    );
    if (activeExport) {
      currentExportJob = activeExport.jobId;
      exporting = true;
      startProgressTracking(activeExport.jobId);

      toast.info("Resuming export", {
        description: `Continuing ${activeExport.stage} export from before app restart.`,
      });
    }
  }

  async function handleStitchedExport() {
    try {
      exporting = true;
      exportError = "";

      // Get export folder from user
      const exportFolder = await SelectExportFolder();
      if (!exportFolder) {
        exporting = false;
        return; // User cancelled
      }

      // Start export job
      const paddingSeconds = paddingValue[0] / 10; // Convert from slider value to seconds
      const jobId = await ExportStitchedHighlights(projectId, exportFolder, paddingSeconds);
      currentExportJob = jobId;

      // Start progress tracking
      startProgressTracking(jobId);

      toast.info("Export started", {
        description:
          "Your stitched video export has begun. Progress will be shown below.",
      });
    } catch (err) {
      console.error("Failed to start stitched video export:", err);
      exportError = "Failed to start stitched video export";
      toast.error("Export failed", {
        description:
          "An error occurred while starting the stitched video export",
      });
      exporting = false;
    }
  }

  async function handleIndividualExport() {
    try {
      exporting = true;
      exportError = "";

      // Get export folder from user
      const exportFolder = await SelectExportFolder();
      if (!exportFolder) {
        exporting = false;
        return; // User cancelled
      }

      // Start export job
      const paddingSeconds = paddingValue[0] / 10; // Convert from slider value to seconds
      const jobId = await ExportIndividualHighlights(projectId, exportFolder, paddingSeconds);
      currentExportJob = jobId;

      // Start progress tracking
      startProgressTracking(jobId);

      toast.info("Export started", {
        description:
          "Your individual clips export has begun. Progress will be shown below.",
      });
    } catch (err) {
      console.error("Failed to start individual clips export:", err);
      exportError = "Failed to start individual clips export";
      toast.error("Export failed", {
        description:
          "An error occurred while starting the individual clips export",
      });
      exporting = false;
    }
  }

  function startProgressTracking(jobId) {
    progressInterval = setInterval(async () => {
      try {
        const progress = await GetExportProgress(jobId);
        exportProgress = progress;

        if (progress.isComplete) {
          clearInterval(progressInterval);
          progressInterval = null;
          exporting = false;
          currentExportJob = null;

          // Refresh export history
          await loadExportHistory();

          if (progress.hasError) {
            exportError = progress.errorMessage;
            toast.error("Export failed", {
              description: progress.errorMessage,
            });
          } else if (progress.isCancelled) {
            toast.info("Export cancelled", {
              description: "The export operation was cancelled.",
            });
          } else {
            toast.success("Export completed!", {
              description: "Your video export has finished successfully.",
            });
            exportProgress = null;
          }
        }
      } catch (err) {
        console.error("Failed to get export progress:", err);
        clearInterval(progressInterval);
        progressInterval = null;
        exporting = false;
      }
    }, 1000); // Check progress every second
  }

  async function handleCancelExport() {
    if (currentExportJob && progressInterval) {
      try {
        await CancelExport(currentExportJob);
        clearInterval(progressInterval);
        progressInterval = null;
        exporting = false;
        currentExportJob = null;
        exportProgress = null;

        toast.info("Export cancelled", {
          description: "The export operation has been cancelled.",
        });
      } catch (err) {
        console.error("Failed to cancel export:", err);
        toast.error("Failed to cancel export", {
          description: "An error occurred while cancelling the export.",
        });
      }
    }
  }

  function formatExportDateTime(jobId) {
    // Extract timestamp from jobId format: "stitched_1_1641234567890123456" or "individual_1_1641234567890123456"
    const parts = jobId.split("_");
    if (parts.length >= 3) {
      // Convert nanoseconds to milliseconds (divide by 1000000)
      const timestampNano = parseInt(parts[2]);
      const timestampMs = Math.floor(timestampNano / 1000000);
      const date = new Date(timestampMs);

      // Format as: Jan 15, 2:30 PM
      return date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
      });
    }
    return "Unknown";
  }
</script>

{#if project && videoClips.length > 0}
  <div class="space-y-6">
    <div>
      <h3 class="text-lg font-semibold">Export Highlights</h3>
      <p class="text-sm text-muted-foreground">
        Export your highlighted video segments
      </p>
    </div>

    <!-- Export error display -->
    {#if exportError}
      <div class="border border-destructive rounded p-4 text-destructive">
        <p class="font-medium">Export Error</p>
        <p class="text-sm">{exportError}</p>
        <Button
          variant="outline"
          size="sm"
          class="mt-2"
          onclick={() => (exportError = "")}
        >
          Dismiss
        </Button>
      </div>
    {/if}

    <!-- Padding Settings -->
    <div class="border rounded p-4 space-y-3">
      <div 
        class="flex items-center justify-between cursor-pointer"
        onclick={() => showPaddingOptions = !showPaddingOptions}
      >
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <h4 class="font-medium">Clip Padding</h4>
            <p class="text-sm text-muted-foreground">
              {paddingValue[0] === 0 ? 'No padding' : `${(paddingValue[0] / 10).toFixed(1)}s padding`}
            </p>
          </div>
        </div>
        
        <div class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground">
            {(paddingValue[0] / 10).toFixed(1)}s
          </span>
          <svg 
            class="w-4 h-4 text-muted-foreground {showPaddingOptions ? 'rotate-180' : ''}"
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      
      {#if showPaddingOptions}
        <div class="space-y-3 pt-2">
          <div class="flex items-center justify-between">
            <label class="text-sm font-medium">Padding duration</label>
            <span class="text-sm text-muted-foreground">
              {(paddingValue[0] / 10).toFixed(1)}s
            </span>
          </div>
          
          <Slider 
            bind:value={paddingValue} 
            max={40} 
            step={1} 
            class="w-full"
          />
          
          <div class="p-3 border rounded">
            <div class="flex gap-2">
              <svg class="w-4 h-4 text-muted-foreground flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <div class="text-sm">
                <p class="font-medium mb-1">Padding Note:</p>
                <p>Only available padding will be added if there's not enough content before/after the clip in the original video.</p>
              </div>
            </div>
          </div>
          
          <p class="text-xs text-muted-foreground">
            Adds {(paddingValue[0] / 10).toFixed(1)} seconds before and after each clip
          </p>
        </div>
      {/if}
    </div>

    <!-- Export options -->
    <div class="grid gap-4 md:grid-cols-2">
      <!-- Stitched video option -->
      <div class="border rounded p-4 space-y-3">
        <div class="flex items-center gap-2">
          <Video class="w-4 h-4" />
          <div>
            <h4 class="font-medium">Single Stitched Video</h4>
            <p class="text-sm text-muted-foreground">
              Combine all highlights into one video file
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          class="w-full"
          onclick={handleStitchedExport}
          disabled={exporting}
        >
          {#if exporting}
            <svg
              class="w-4 h-4 mr-2 animate-spin"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Exporting...
          {:else}
            <Download class="w-4 h-4 mr-2" />
            Export Stitched
          {/if}
        </Button>
      </div>

      <!-- Individual clips option -->
      <div class="border rounded p-4 space-y-3">
        <div class="flex items-center gap-2">
          <FolderOpen class="w-4 h-4" />
          <div>
            <h4 class="font-medium">Individual Clip Files</h4>
            <p class="text-sm text-muted-foreground">
              Export each highlight as a separate numbered file
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          class="w-full"
          onclick={handleIndividualExport}
          disabled={exporting}
        >
          {#if exporting}
            <svg
              class="w-4 h-4 mr-2 animate-spin"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Exporting...
          {:else}
            <FolderOpen class="w-4 h-4 mr-2" />
            Export Individual
          {/if}
        </Button>
      </div>
    </div>

    <!-- Export Progress -->
    {#if exportProgress}
      <div class="p-4 border rounded">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <svg
              class="w-4 h-4 animate-spin"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            <h4 class="font-medium">Export Progress</h4>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onclick={handleCancelExport}
          >
            Cancel
          </Button>
        </div>

        <div class="space-y-2 mb-3">
          <div class="flex justify-between text-sm">
            <span class="capitalize">{exportProgress.stage}</span>
            <span
              >{Math.round(exportProgress.progress * 100)}%</span
            >
          </div>
          <div class="w-full bg-secondary rounded-full h-2">
            <div
              class="bg-primary h-2 rounded-full transition-all duration-300"
              style="width: {exportProgress.progress * 100}%"
            ></div>
          </div>
        </div>

        {#if exportProgress.currentFile}
          <div class="text-sm text-muted-foreground mb-2">
            Processing: {exportProgress.currentFile}
          </div>
        {/if}

        {#if exportProgress.totalFiles > 0}
          <div class="text-sm text-muted-foreground">
            {exportProgress.processedFiles} of {exportProgress.totalFiles}
            files processed
          </div>
        {/if}
      </div>
    {/if}

    <!-- Export History -->
    {#if exportHistory.length > 0}
      <div>
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">Export History</h4>
          <Button
            variant="ghost"
            size="sm"
            onclick={() =>
              (showExportHistory = !showExportHistory)}
          >
            {showExportHistory ? "Hide" : "Show"} History
          </Button>
        </div>

        {#if showExportHistory}
          <div class="space-y-2 max-h-48 overflow-y-auto">
            {#each exportHistory as job (job.jobId)}
              <div class="border rounded p-3">
                <div
                  class="flex items-center justify-between mb-2"
                >
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium capitalize"
                      >{job.stage}</span
                    >
                    {#if job.isComplete}
                      {#if job.hasError}
                        <span
                          class="text-xs bg-destructive text-destructive-foreground px-2 py-1 rounded"
                          >Failed</span
                        >
                      {:else if job.isCancelled}
                        <span
                          class="text-xs bg-muted text-muted-foreground px-2 py-1 rounded"
                          >Cancelled</span
                        >
                      {:else}
                        <span
                          class="text-xs bg-green-100 text-green-800 px-2 py-1 rounded"
                          >Completed</span
                        >
                      {/if}
                    {:else}
                      <span
                        class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded"
                        >In Progress</span
                      >
                    {/if}
                  </div>
                  <span class="text-xs text-muted-foreground"
                    >{Math.round(job.progress * 100)}%</span
                  >
                </div>

                <div
                  class="text-xs text-muted-foreground space-y-1"
                >
                  <div class="flex justify-between">
                    <span>Type:</span>
                    <span
                      >{job.jobId.startsWith("stitched")
                        ? "Stitched Video"
                        : "Individual Clips"}</span
                    >
                  </div>
                  <div class="flex justify-between">
                    <span>Started:</span>
                    <span>{formatExportDateTime(job.jobId)}</span>
                  </div>
                  {#if job.currentFile}
                    <div class="flex justify-between">
                      <span>Current:</span>
                      <span
                        class="truncate max-w-32"
                        title={job.currentFile}
                        >{job.currentFile}</span
                      >
                    </div>
                  {/if}
                  {#if job.totalFiles > 0}
                    <div class="flex justify-between">
                      <span>Progress:</span>
                      <span
                        >{job.processedFiles}/{job.totalFiles} segments</span
                      >
                    </div>
                  {/if}
                  {#if job.hasError && job.errorMessage}
                    <div class="text-destructive">
                      Error: {job.errorMessage}
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Export info -->
    <div class="p-4 border rounded">
      <div class="flex items-start gap-3">
        <svg
          class="w-4 h-4 text-muted-foreground flex-shrink-0 mt-0.5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <div class="text-sm text-muted-foreground">
          <p class="font-medium text-foreground mb-1">
            Export Information:
          </p>
          <ul class="space-y-1">
            <li>
              • Only video clips with highlights will be exported
            </li>
            <li>
              • A timestamped project folder will be created in
              your chosen location
            </li>
            <li>
              • Individual clips will be numbered sequentially
              with time spans
            </li>
            <li>
              • Original video quality will be preserved
              (H.264/AAC)
            </li>
            <li>
              • Export progress persists across app restarts
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
{:else}
  <div class="text-center py-8 text-muted-foreground">
    <p>No video clips available for export</p>
    <p class="text-sm">
      Add video clips first to enable export functionality
    </p>
  </div>
{/if}