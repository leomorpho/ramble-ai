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
    CreateVideoClip,
    GetVideoClipsByProject,
    DeleteVideoClip,
    SelectVideoFiles,
    TranscribeVideoClip,
    GetOpenAIApiKey,
    SelectExportFolder,
    ExportStitchedHighlights,
    ExportIndividualHighlights,
    GetExportProgress,
    CancelExport,
    GetProjectExportJobs,
    UpdateProjectActiveTab,
  } from "$lib/wailsjs/go/main/App";
  import {
    OnFileDrop,
    OnFileDropOff,
    EventsOn,
    EventsOff,
  } from "$lib/wailsjs/runtime/runtime";
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
  import VideoClipCard from "$lib/components/VideoClipCard.svelte";
  import ProjectDetails from "$lib/components/ProjectDetails.svelte";
  import FileDropZone from "$lib/components/FileDropZone.svelte";
  import {
    Video,
    Download,
    FolderOpen,
    Info,
    Film,
    Clock,
    Upload,
    Copy,
    Check,
  } from "@lucide/svelte";
  import { updateVideoHighlights } from "$lib/stores/projectHighlights.js";

  let project = $state(null);
  let loading = $state(false);
  let error = $state("");

  // Video clips state
  let videoClips = $state([]);
  let loadingClips = $state(false);
  let addingClip = $state(false);
  let clipError = $state("");
  let dragActive = $state(false);

  // Transcription state
  let transcribingClips = $state(new Set());

  // Highlights component reference
  let projectHighlightsComponent = $state(null);

  // Highlights state managed at parent level
  let highlights = $state([]);
  let highlightsLoaded = $state(false);

  // Export state
  let exporting = $state(false);
  let exportError = $state("");
  let currentExportJob = $state(null);
  let exportProgress = $state(null);
  let progressInterval = $state(null);
  let exportHistory = $state([]);
  let showExportHistory = $state(false);

  // Tabs state
  let activeTab = $state("clips");
  let debounceTimer = null;

  // Copy button state
  let pathCopied = $state(false);

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  onMount(async () => {
    await loadProject();
    await loadVideoClips();
    await loadExportHistory();
    await loadHighlights();

    // Check if there are any active exports to resume tracking
    await checkForActiveExports();

    // Set up Wails drag and drop listeners
    OnFileDrop(handleWailsFileDrop, true);
    EventsOn("files-dropped", handleFilesDroppedEvent);

    // Watch for Wails drop target class changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (
          mutation.type === "attributes" &&
          mutation.attributeName === "class"
        ) {
          const target = mutation.target;
          if (target.classList.contains("wails-drop-target-active")) {
            dragActive = true;
          } else {
            dragActive = false;
          }
        }
      });
    });

    // Observe the document for class changes
    observer.observe(document.body, {
      attributes: true,
      subtree: true,
      attributeFilter: ["class"],
    });
  });

  onDestroy(() => {
    // Clean up progress tracking if active
    if (progressInterval) {
      clearInterval(progressInterval);
    }

    // Clean up Wails drag and drop listeners
    OnFileDropOff();
    EventsOff("files-dropped");

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

  async function loadVideoClips() {
    if (!projectId || isNaN(projectId)) return;

    try {
      loadingClips = true;
      clipError = "";
      const clips = await GetVideoClipsByProject(projectId);
      videoClips = clips || []; // Ensure it's always an array
    } catch (err) {
      console.error("Failed to load video clips:", err);
      clipError = "Failed to load video clips";
      videoClips = []; // Set to empty array on error
    } finally {
      loadingClips = false;
    }
  }

  function isVideoFile(file) {
    const videoTypes = [
      "video/mp4",
      "video/quicktime",
      "video/x-msvideo",
      "video/x-matroska",
      "video/x-ms-wmv",
      "video/x-flv",
      "video/webm",
      "video/mpeg",
    ];
    return (
      videoTypes.includes(file.type) ||
      file.name.match(/\.(mp4|mov|avi|mkv|wmv|flv|webm|m4v|mpg|mpeg)$/i)
    );
  }

  async function handleFiles(files) {
    if (!files || files.length === 0) return;

    addingClip = true;
    clipError = "";

    console.log(
      "Processing files:",
      Array.from(files).map((f) => ({
        name: f.name,
        type: f.type,
        size: f.size,
        path: f.path,
        webkitRelativePath: f.webkitRelativePath,
      }))
    );

    const fileArray = Array.from(files);
    const videoFiles = fileArray.filter(isVideoFile);

    if (videoFiles.length === 0) {
      clipError =
        "No valid video files found. Please select video files (MP4, MOV, AVI, etc.)";
      addingClip = false;
      return;
    }

    let successCount = 0;

    for (const file of videoFiles) {
      try {
        // In Wails desktop app, files should have a path property
        let filePath = file.path;

        // If no path property, try webkitRelativePath (some browsers)
        if (!filePath && file.webkitRelativePath) {
          filePath = file.webkitRelativePath;
        }

        // If still no path, try to construct from file name (fallback)
        if (!filePath) {
          console.warn(
            `No file path available for ${file.name}. This may happen in web browsers.`
          );
          clipError = `Cannot access file system path for ${file.name}. Please use "Select Video Files" button instead.`;
          break;
        }

        console.log(`Processing file: ${file.name} with path: ${filePath}`);

        const newClip = await CreateVideoClip(projectId, filePath);
        // Check if this clip is already in our list
        if (!videoClips || !videoClips.some((clip) => clip.id === newClip.id)) {
          videoClips = [...(videoClips || []), newClip]; // Trigger reactivity
          successCount++;
        }
      } catch (err) {
        console.error("Failed to add video clip:", err);
        if (err.message && err.message.includes("already added")) {
          clipError = `${file.name} is already added to this project`;
        } else {
          clipError = `Failed to add ${file.name}: ${err.message || err}`;
        }
        break;
      }
    }

    // Show success message if files were added
    if (successCount > 0 && !clipError) {
      toast.success(
        `Added ${successCount} video file${successCount === 1 ? "" : "s"}`,
        {
          description: "Video files have been added to your project",
        }
      );
    }

    addingClip = false;
  }

  // Wails runtime drag and drop handlers
  function handleWailsFileDrop(x, y, paths) {
    console.log(`Wails OnFileDrop: Files dropped at (${x}, ${y}):`, paths);

    if (!paths || paths.length === 0) {
      console.log("No paths received in Wails OnFileDrop");
      return;
    }

    // Set visual feedback
    dragActive = false;

    // Process the dropped files directly with full paths
    handleFilePathsFromWails(paths);
  }

  function handleFilesDroppedEvent(data) {
    console.log("Wails files-dropped event received:", data);

    if (!data || !data.paths || data.paths.length === 0) {
      console.log("No paths in files-dropped event");
      return;
    }

    // Set visual feedback
    dragActive = false;

    // Process the dropped files
    handleFilePathsFromWails(data.paths);
  }

  async function handleFilePathsFromWails(filePaths) {
    console.log("Processing file paths from Wails:", filePaths);

    addingClip = true;
    clipError = "";

    const videoFiles = filePaths.filter((path) => {
      const ext = path.toLowerCase().split(".").pop();
      return [
        "mp4",
        "mov",
        "avi",
        "mkv",
        "wmv",
        "flv",
        "webm",
        "m4v",
        "mpg",
        "mpeg",
      ].includes(ext);
    });

    if (videoFiles.length === 0) {
      clipError =
        "No valid video files found. Please drop video files (MP4, MOV, AVI, etc.)";
      addingClip = false;
      return;
    }

    console.log("Valid video files to process:", videoFiles);

    let successCount = 0;

    for (const filePath of videoFiles) {
      try {
        console.log(`Adding video clip: ${filePath}`);
        const newClip = await CreateVideoClip(projectId, filePath);

        // Check if this clip is already in our list
        if (!videoClips || !videoClips.some((clip) => clip.id === newClip.id)) {
          videoClips = [...(videoClips || []), newClip]; // Trigger reactivity
          successCount++;
          console.log(`Successfully added: ${newClip.name}`);
        }
      } catch (err) {
        console.error("Failed to add video clip:", err);
        if (err.message && err.message.includes("already added")) {
          clipError = `${filePath.split("/").pop()} is already added to this project`;
        } else {
          clipError = `Failed to add ${filePath.split("/").pop()}: ${err.message || err}`;
        }
        break;
      }
    }

    // Show success message if files were added
    if (successCount > 0 && !clipError) {
      toast.success(
        `Added ${successCount} video file${successCount === 1 ? "" : "s"}`,
        {
          description: "Video files have been added to your project",
        }
      );
    }

    addingClip = false;
  }

  // Enhanced drag and drop handlers that work with Wails
  function handleDrop(event) {
    event.preventDefault();
    dragActive = false;

    console.log("Drop event received:", event);
    console.log("DataTransfer:", event.dataTransfer);

    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      console.log(
        "Files from drop:",
        Array.from(files).map((f) => ({
          name: f.name,
          type: f.type,
          size: f.size,
          path: f.path,
          webkitRelativePath: f.webkitRelativePath,
        }))
      );
      handleFiles(files);
    } else {
      console.log("No files in drop event");

      // Try alternative methods to get file information
      const items = event.dataTransfer?.items;
      if (items && items.length > 0) {
        console.log(
          "DataTransfer items:",
          Array.from(items).map((item) => ({
            kind: item.kind,
            type: item.type,
          }))
        );

        // Try to process items
        const filePromises = [];
        for (let i = 0; i < items.length; i++) {
          const item = items[i];
          if (item.kind === "file") {
            const file = item.getAsFile();
            if (file) {
              filePromises.push(file);
            }
          }
        }

        if (filePromises.length > 0) {
          console.log(
            "Files from items:",
            filePromises.map((f) => ({
              name: f.name,
              type: f.type,
              size: f.size,
              path: f.path,
            }))
          );
          handleFiles(filePromises);
          return;
        }
      }

      toast.error("No files detected in drop", {
        description:
          "Please try using the 'Select Video Files' button instead.",
      });
    }
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "copy";
  }

  function handleDragEnter(event) {
    event.preventDefault();
    dragActive = true;
    console.log("Drag enter detected");
  }

  function handleDragLeave(event) {
    event.preventDefault();
    // Only hide drag state if we're leaving the drop zone completely
    if (!event.currentTarget.contains(event.relatedTarget)) {
      dragActive = false;
      console.log("Drag leave detected");
    }
  }

  async function selectVideoFiles() {
    try {
      addingClip = true;
      clipError = "";
      const selectedFiles = await SelectVideoFiles();

      // Add each selected file to database
      for (const file of selectedFiles) {
        try {
          const newClip = await CreateVideoClip(projectId, file.filePath);
          // Check if this clip is already in our list
          if (
            !videoClips ||
            !videoClips.some((clip) => clip.id === newClip.id)
          ) {
            videoClips = [...(videoClips || []), newClip]; // Trigger reactivity
          }
        } catch (err) {
          console.error("Failed to add video clip:", err);
          if (err.message && err.message.includes("already added")) {
            clipError = `${file.fileName} is already added to this project`;
          } else {
            clipError = `Failed to add ${file.fileName}: ${err.message || err}`;
          }
          continue; // Continue with other files instead of breaking
        }
      }
    } catch (err) {
      console.error("Failed to select video files:", err);
      clipError = "Failed to select video files";
    } finally {
      addingClip = false;
    }
  }

  function handleFileInputChange(event) {
    const files = event.target?.files;
    if (files) {
      handleFiles(files);
    }
    // Reset input value so the same file can be selected again
    if (event.target) {
      event.target.value = "";
    }
  }

  function openFileDialog(fileInput) {
    fileInput?.click();
  }

  function handleKeyDown(event, fileInput) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      fileInput?.click();
    }
  }

  async function handleDeleteClip(clipId) {
    try {
      await DeleteVideoClip(clipId);
      videoClips = videoClips.filter((clip) => clip.id !== clipId);

      // Refresh highlights after deletion
      await loadHighlights();
    } catch (err) {
      console.error("Failed to delete video clip:", err);
      clipError = "Failed to delete video clip";
    }
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  function goBack() {
    goto("/");
  }

  async function copyToClipboard(text) {
    try {
      await navigator.clipboard.writeText(text);
      pathCopied = true;
      toast.success("Copied to clipboard");
      
      // Reset the copied state after 2 seconds
      setTimeout(() => {
        pathCopied = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
      toast.error("Failed to copy to clipboard");
    }
  }

  async function startTranscription(clip) {
    try {
      // Check if OpenAI API key is configured
      const apiKey = await GetOpenAIApiKey();
      if (!apiKey || apiKey.trim() === "") {
        toast.error("OpenAI API Key Required", {
          description:
            "Please configure your OpenAI API key in settings to use transcription.",
        });

        // Redirect to settings after a short delay
        setTimeout(() => {
          goto("/settings");
        }, 2000);
        return;
      }

      // Add clip to transcribing set
      transcribingClips.add(clip.id);
      transcribingClips = new Set(transcribingClips); // Trigger reactivity

      // Show starting toast
      toast.info(`Starting transcription for ${clip.name}`, {
        description: "Extracting audio and sending to OpenAI...",
      });

      const result = await TranscribeVideoClip(clip.id);

      if (result.success) {
        // Update the clip with transcription and words data
        const clipIndex = videoClips.findIndex((c) => c.id === clip.id);
        if (clipIndex !== -1) {
          videoClips[clipIndex] = {
            ...videoClips[clipIndex],
            transcription: result.transcription,
            transcriptionWords: result.words || [],
            transcriptionLanguage: result.language,
            transcriptionDuration: result.duration,
          };
          videoClips = [...videoClips]; // Trigger reactivity
        }

        // Show success toast
        toast.success(`Transcription completed for ${clip.name}`, {
          description: "Transcript is now available to view",
        });

        // Refresh highlights since new transcription might have highlights
        await loadHighlights();
      } else {
        // Show error toast
        toast.error(`Transcription failed for ${clip.name}`, {
          description: result.message,
        });
      }
    } catch (err) {
      console.error("Transcription error:", err);
      toast.error(`Transcription failed for ${clip.name}`, {
        description: "An unexpected error occurred",
      });
    } finally {
      // Remove clip from transcribing set
      transcribingClips.delete(clip.id);
      transcribingClips = new Set(transcribingClips); // Trigger reactivity
    }
  }

  async function handleHighlightsChange(videoId, highlights) {
    try {
      // Use the store function to update highlights
      await updateVideoHighlights(videoId, highlights);

      // No need to manually update local state or refresh components
      // The store will handle updating all subscribers automatically
    } catch (err) {
      console.error("Failed to save highlights:", err);
      // Error toast is already shown by the store function
    }
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, "0")}`;
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
      const jobId = await ExportStitchedHighlights(projectId, exportFolder);
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
      const jobId = await ExportIndividualHighlights(projectId, exportFolder);
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

<main
  class="min-h-screen bg-background text-foreground p-8"
  style="--wails-drop-target: drop"
  ondrop={handleDrop}
  ondragover={handleDragOver}
  ondragenter={handleDragEnter}
  ondragleave={handleDragLeave}
>
  <div class="max-w-4xl mx-auto space-y-6">
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
              <div class="space-y-8">
                <!-- Project Overview Card -->
                <div>
                  <div class="flex items-start justify-between mb-4">
                    <ProjectDetails
                      {project}
                      onUpdate={handleProjectUpdate}
                      onDelete={handleProjectDelete}
                      buttonsOnly={true}
                    />
                  </div>

                  <!-- Statistics Grid -->
                  <div class="grid gap-6 md:grid-cols-3">
                    <div
                      class="text-center p-4 bg-background rounded-lg border"
                    >
                      <div class="text-2xl font-bold text-primary mb-1">
                        {videoClips.length}
                      </div>
                      <div class="text-sm text-muted-foreground">
                        Video Clips
                      </div>
                    </div>
                    <div
                      class="text-center p-4 bg-background rounded-lg border"
                    >
                      <div class="text-2xl font-bold text-primary mb-1">
                        {highlights.length}
                      </div>
                      <div class="text-sm text-muted-foreground">
                        Highlights
                      </div>
                    </div>
                    <div
                      class="text-center p-4 bg-background rounded-lg border"
                    >
                      <div class="text-2xl font-bold text-primary mb-1">
                        {Math.floor(
                          highlights.reduce(
                            (sum, highlight) => sum + (highlight.end - highlight.start),
                            0
                          ) / 60
                        )}:{Math.floor(
                          highlights.reduce(
                            (sum, highlight) => sum + (highlight.end - highlight.start),
                            0
                          ) % 60
                        )
                          .toString()
                          .padStart(2, "0")}
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
                        <span class="text-sm font-medium text-muted-foreground"
                          >Created</span
                        >
                        <span class="text-sm font-medium"
                          >{new Date(project.createdAt).toLocaleDateString(
                            "en-US",
                            {
                              year: "numeric",
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                            }
                          )}</span
                        >
                      </div>
                      <div
                        class="flex justify-between items-center py-2 border-b border-border/50"
                      >
                        <span class="text-sm font-medium text-muted-foreground"
                          >Last Updated</span
                        >
                        <span class="text-sm font-medium"
                          >{new Date(project.updatedAt).toLocaleDateString(
                            "en-US",
                            {
                              year: "numeric",
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                            }
                          )}</span
                        >
                      </div>
                    </div>
                    <div class="space-y-3">
                      <div
                        class="flex justify-between items-center py-2 border-b border-border/50"
                      >
                        <span class="text-sm font-medium text-muted-foreground"
                          >Project ID</span
                        >
                        <span
                          class="text-sm font-mono bg-secondary px-2 py-1 rounded"
                          >{project.id}</span
                        >
                      </div>
                      <div class="py-2 md:col-span-2">
                        <div
                          class="text-sm font-medium text-muted-foreground mb-2"
                        >
                          Project Path
                        </div>
                        <div class="flex items-center gap-2">
                          <code
                            class="text-sm bg-secondary px-3 py-2 rounded flex-1 break-all"
                            >{project.path}</code
                          >
                          <Button
                            variant="ghost"
                            size="sm"
                            onclick={() => copyToClipboard(project.path)}
                            class="flex-shrink-0 transition-all"
                            disabled={pathCopied}
                          >
                            {#if pathCopied}
                              <Check class="w-4 h-4 text-green-600" />
                            {:else}
                              <Copy class="w-4 h-4" />
                            {/if}
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </TabsContent>

            <!-- Video Clips Tab -->
            <TabsContent value="clips" class="mt-6">
              <div class="space-y-4">
                <div class="flex justify-between items-center">
                  <h3 class="text-lg font-semibold">Video Clips</h3>
                  <Button
                    onclick={selectVideoFiles}
                    disabled={addingClip}
                    size="sm"
                  >
                    {addingClip ? "Adding..." : "Add Video Files"}
                  </Button>
                </div>

                <!-- Video clip error display -->
                {#if clipError}
                  <div
                    class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
                  >
                    <p class="font-medium">Error</p>
                    <p class="text-sm">{clipError}</p>
                    <Button
                      variant="outline"
                      size="sm"
                      class="mt-2"
                      onclick={() => (clipError = "")}
                    >
                      Dismiss
                    </Button>
                  </div>
                {/if}

                <!-- File drop zone -->
                <FileDropZone
                  {dragActive}
                  {addingClip}
                  onFileInputChange={handleFileInputChange}
                  onOpenFileDialog={openFileDialog}
                  onKeyDown={handleKeyDown}
                />

                <!-- Video clips grid -->
                {#if loadingClips}
                  <div class="text-center py-8 text-muted-foreground">
                    <p>Loading video clips...</p>
                  </div>
                {:else if videoClips.length === 0}
                  <div class="text-center py-8 text-muted-foreground">
                    <p>No video clips yet</p>
                    <p class="text-sm">
                      Drag and drop video files or use "Add Video Files" to get
                      started
                    </p>
                  </div>
                {:else}
                  <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                    {#each videoClips as clip (clip.id)}
                      <VideoClipCard
                        {clip}
                        isTranscribing={transcribingClips.has(clip.id)}
                        onDelete={handleDeleteClip}
                        onStartTranscription={startTranscription}
                        {formatFileSize}
                        {projectId}
                        {highlights}
                        onHighlightsChange={(highlights) =>
                          handleHighlightsChange(clip.id, highlights)}
                      />
                    {/each}
                  </div>
                {/if}
              </div>
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
              {#if project && videoClips.length > 0}
                <div class="space-y-6">
                  <div class="flex items-center gap-3">
                    <Download class="w-5 h-5 text-primary" />
                    <div>
                      <h3 class="text-lg font-semibold">Export Highlights</h3>
                      <p class="text-sm text-muted-foreground">
                        Export your highlighted video segments
                      </p>
                    </div>
                  </div>

                  <!-- Export error display -->
                  {#if exportError}
                    <div
                      class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4"
                    >
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

                  <!-- Export options -->
                  <div class="grid gap-4 md:grid-cols-2">
                    <!-- Stitched video option -->
                    <div class="border rounded-lg p-4 space-y-3">
                      <div class="flex items-center gap-3">
                        <div
                          class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center"
                        >
                          <Video class="w-5 h-5 text-primary" />
                        </div>
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
                          Export Stitched Video
                        {/if}
                      </Button>
                    </div>

                    <!-- Individual clips option -->
                    <div class="border rounded-lg p-4 space-y-3">
                      <div class="flex items-center gap-3">
                        <div
                          class="w-10 h-10 bg-secondary/50 rounded-lg flex items-center justify-center"
                        >
                          <FolderOpen class="w-5 h-5 text-foreground" />
                        </div>
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
                          Export Individual Clips
                        {/if}
                      </Button>
                    </div>
                  </div>

                  <!-- Export Progress -->
                  {#if exportProgress}
                    <div
                      class="p-4 bg-primary/5 border border-primary/20 rounded-lg"
                    >
                      <div class="flex items-center justify-between mb-3">
                        <div class="flex items-center gap-2">
                          <svg
                            class="w-5 h-5 text-primary animate-spin"
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
                            <div class="border rounded-lg p-3 bg-secondary/20">
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
                  <div class="p-4 bg-secondary/30 rounded-lg">
                    <div class="flex items-start gap-3">
                      <svg
                        class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5"
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
