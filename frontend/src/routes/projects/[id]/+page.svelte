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
  import { 
    Tabs, 
    TabsContent, 
    TabsList, 
    TabsTrigger 
  } from "$lib/components/ui/tabs";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index.js";
  import { GetProjectByID, UpdateProject, DeleteProject, CreateVideoClip, GetVideoClipsByProject, UpdateVideoClip, DeleteVideoClip, SelectVideoFiles, GetVideoFileInfo, GetVideoURL, TranscribeVideoClip, UpdateVideoClipHighlights, GetOpenAIApiKey, SelectExportFolder, ExportStitchedHighlights, ExportIndividualHighlights, GetExportProgress, CancelExport, GetProjectExportJobs } from "$lib/wailsjs/go/main/App";
  import { OnFileDrop, OnFileDropOff, EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";
  import { onMount, onDestroy } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";
  import TextHighlighter from "$lib/components/TextHighlighter.svelte";
  import ProjectHighlights from "$lib/components/ProjectHighlights.svelte";
  import ThemeSwitcher from "$lib/components/ui/theme-switcher/theme-switcher.svelte";
  import VideoClipCard from "$lib/components/VideoClipCard.svelte";
  import EtroVideoPlayer from "$lib/components/videoplayback/EtroVideoPlayer.svelte";
  import VideoTranscriptViewer from "$lib/components/VideoTranscriptViewer.svelte";
  import VideoPreviewDialog from "$lib/components/VideoPreviewDialog.svelte";
  import { Captions, Mic, Video, Download, FolderOpen } from "@lucide/svelte";

  let project = $state(null);
  let loading = $state(false);
  let error = $state("");
  let editDialogOpen = $state(false);
  let deleteDialogOpen = $state(false);
  let editName = $state("");
  let editDescription = $state("");
  let deleting = $state(false);
  
  // Video clips state
  let videoClips = $state([]);
  let loadingClips = $state(false);
  let addingClip = $state(false);
  let clipError = $state("");
  let dragActive = $state(false);
  let fileInput = $state();
  
  // Video preview state
  let previewDialogOpen = $state(false);
  let previewVideo = $state(null);
  
  // Transcription state
  let transcriptionDialogOpen = $state(false);
  let transcriptionVideo = $state(null);
  let transcribingClips = $state(new Set());
  
  // Highlights component reference
  let projectHighlightsComponent = $state(null);
  
  // Export state
  let exporting = $state(false);
  let exportError = $state("");
  let currentExportJob = $state(null);
  let exportProgress = $state(null);
  let progressInterval = $state(null);
  let exportHistory = $state([]);
  let showExportHistory = $state(false);

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  onMount(async () => {
    await loadProject();
    await loadVideoClips();
    await loadExportHistory();
    
    // Check if there are any active exports to resume tracking
    await checkForActiveExports();
    
    // Set up Wails drag and drop listeners
    OnFileDrop(handleWailsFileDrop, true);
    EventsOn("files-dropped", handleFilesDroppedEvent);
    
    // Watch for Wails drop target class changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
          const target = mutation.target;
          if (target.classList.contains('wails-drop-target-active')) {
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
      attributeFilter: ['class']
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
      'video/mp4', 'video/quicktime', 'video/x-msvideo', 'video/x-matroska',
      'video/x-ms-wmv', 'video/x-flv', 'video/webm', 'video/mpeg'
    ];
    return videoTypes.includes(file.type) || file.name.match(/\.(mp4|mov|avi|mkv|wmv|flv|webm|m4v|mpg|mpeg)$/i);
  }

  async function handleFiles(files) {
    if (!files || files.length === 0) return;
    
    addingClip = true;
    clipError = "";
    
    console.log("Processing files:", Array.from(files).map(f => ({ 
      name: f.name, 
      type: f.type, 
      size: f.size, 
      path: f.path,
      webkitRelativePath: f.webkitRelativePath 
    })));
    
    const fileArray = Array.from(files);
    const videoFiles = fileArray.filter(isVideoFile);
    
    if (videoFiles.length === 0) {
      clipError = "No valid video files found. Please select video files (MP4, MOV, AVI, etc.)";
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
          console.warn(`No file path available for ${file.name}. This may happen in web browsers.`);
          clipError = `Cannot access file system path for ${file.name}. Please use "Select Video Files" button instead.`;
          break;
        }
        
        console.log(`Processing file: ${file.name} with path: ${filePath}`);
        
        const newClip = await CreateVideoClip(projectId, filePath);
        // Check if this clip is already in our list
        if (!videoClips || !videoClips.some(clip => clip.id === newClip.id)) {
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
      toast.success(`Added ${successCount} video file${successCount === 1 ? '' : 's'}`, {
        description: "Video files have been added to your project"
      });
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
    
    const videoFiles = filePaths.filter(path => {
      const ext = path.toLowerCase().split('.').pop();
      return ['mp4', 'mov', 'avi', 'mkv', 'wmv', 'flv', 'webm', 'm4v', 'mpg', 'mpeg'].includes(ext);
    });
    
    if (videoFiles.length === 0) {
      clipError = "No valid video files found. Please drop video files (MP4, MOV, AVI, etc.)";
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
        if (!videoClips || !videoClips.some(clip => clip.id === newClip.id)) {
          videoClips = [...(videoClips || []), newClip]; // Trigger reactivity
          successCount++;
          console.log(`Successfully added: ${newClip.name}`);
        }
      } catch (err) {
        console.error("Failed to add video clip:", err);
        if (err.message && err.message.includes("already added")) {
          clipError = `${filePath.split('/').pop()} is already added to this project`;
        } else {
          clipError = `Failed to add ${filePath.split('/').pop()}: ${err.message || err}`;
        }
        break;
      }
    }
    
    // Show success message if files were added
    if (successCount > 0 && !clipError) {
      toast.success(`Added ${successCount} video file${successCount === 1 ? '' : 's'}`, {
        description: "Video files have been added to your project"
      });
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
      console.log("Files from drop:", Array.from(files).map(f => ({ 
        name: f.name, 
        type: f.type, 
        size: f.size,
        path: f.path,
        webkitRelativePath: f.webkitRelativePath
      })));
      handleFiles(files);
    } else {
      console.log("No files in drop event");
      
      // Try alternative methods to get file information
      const items = event.dataTransfer?.items;
      if (items && items.length > 0) {
        console.log("DataTransfer items:", Array.from(items).map(item => ({
          kind: item.kind,
          type: item.type
        })));
        
        // Try to process items
        const filePromises = [];
        for (let i = 0; i < items.length; i++) {
          const item = items[i];
          if (item.kind === 'file') {
            const file = item.getAsFile();
            if (file) {
              filePromises.push(file);
            }
          }
        }
        
        if (filePromises.length > 0) {
          console.log("Files from items:", filePromises.map(f => ({ 
            name: f.name, 
            type: f.type, 
            size: f.size,
            path: f.path
          })));
          handleFiles(filePromises);
          return;
        }
      }
      
      toast.error("No files detected in drop", {
        description: "Please try using the 'Select Video Files' button instead."
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
          if (!videoClips || !videoClips.some(clip => clip.id === newClip.id)) {
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
      event.target.value = '';
    }
  }

  function openFileDialog() {
    fileInput?.click();
  }

  function handleKeyDown(event) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      openFileDialog();
    }
  }

  async function handleDeleteClip(clipId) {
    try {
      await DeleteVideoClip(clipId);
      videoClips = videoClips.filter(clip => clip.id !== clipId);
      
      // Refresh highlights timeline after deletion
      if (projectHighlightsComponent) {
        projectHighlightsComponent.refresh();
      }
    } catch (err) {
      console.error("Failed to delete video clip:", err);
      clipError = "Failed to delete video clip";
    }
  }

  async function openPreview(clip) {
    previewVideo = clip;
    previewDialogOpen = true;
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function goBack() {
    goto("/");
  }

  async function startTranscription(clip) {
    try {
      // Check if OpenAI API key is configured
      const apiKey = await GetOpenAIApiKey();
      if (!apiKey || apiKey.trim() === '') {
        toast.error("OpenAI API Key Required", {
          description: "Please configure your OpenAI API key in settings to use transcription."
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
        description: "Extracting audio and sending to OpenAI..."
      });

      const result = await TranscribeVideoClip(clip.id);
      
      if (result.success) {
        // Update the clip with transcription and words data
        const clipIndex = videoClips.findIndex(c => c.id === clip.id);
        if (clipIndex !== -1) {
          videoClips[clipIndex] = { 
            ...videoClips[clipIndex], 
            transcription: result.transcription,
            transcriptionWords: result.words || [],
            transcriptionLanguage: result.language,
            transcriptionDuration: result.duration
          };
          videoClips = [...videoClips]; // Trigger reactivity
        }
        
        // Show success toast
        toast.success(`Transcription completed for ${clip.name}`, {
          description: "Transcript is now available to view"
        });
        
        // Refresh highlights timeline since new transcription might have highlights
        if (projectHighlightsComponent) {
          projectHighlightsComponent.refresh();
        }
      } else {
        // Show error toast
        toast.error(`Transcription failed for ${clip.name}`, {
          description: result.message
        });
      }
    } catch (err) {
      console.error("Transcription error:", err);
      toast.error(`Transcription failed for ${clip.name}`, {
        description: "An unexpected error occurred"
      });
    } finally {
      // Remove clip from transcribing set
      transcribingClips.delete(clip.id);
      transcribingClips = new Set(transcribingClips); // Trigger reactivity
    }
  }

  function viewTranscription(clip) {
    transcriptionVideo = clip;
    transcriptionDialogOpen = true;
  }

  async function handleHighlightsChange(highlights) {
    if (!transcriptionVideo) return;
    
    try {
      await UpdateVideoClipHighlights(transcriptionVideo.id, highlights);
      
      // Update the local video clip data
      const clipIndex = videoClips.findIndex(c => c.id === transcriptionVideo.id);
      if (clipIndex !== -1) {
        videoClips[clipIndex] = { 
          ...videoClips[clipIndex], 
          highlights: highlights
        };
        videoClips = [...videoClips]; // Trigger reactivity
      }
      
      // Update the transcription video as well
      transcriptionVideo = {
        ...transcriptionVideo,
        highlights: highlights
      };
      
      // Refresh the highlights timeline
      if (projectHighlightsComponent) {
        projectHighlightsComponent.refresh();
      }
    } catch (err) {
      console.error("Failed to save highlights:", err);
      toast.error("Failed to save highlights", {
        description: "An error occurred while saving your highlights"
      });
    }
  }

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, '0')}`;
  }
  
  function formatExportDateTime(jobId) {
    // Extract timestamp from jobId format: "stitched_1_1641234567890123456" or "individual_1_1641234567890123456"
    const parts = jobId.split('_');
    if (parts.length >= 3) {
      // Convert nanoseconds to milliseconds (divide by 1000000)
      const timestampNano = parseInt(parts[2]);
      const timestampMs = Math.floor(timestampNano / 1000000);
      const date = new Date(timestampMs);
      
      // Format as: Jan 15, 2:30 PM
      return date.toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
        hour12: true
      });
    }
    return 'Unknown';
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
    const activeExport = exportHistory.find(job => !job.isComplete && !job.isCancelled);
    if (activeExport) {
      currentExportJob = activeExport.jobId;
      exporting = true;
      startProgressTracking(activeExport.jobId);
      
      toast.info("Resuming export", {
        description: `Continuing ${activeExport.stage} export from before app restart.`
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
        description: "Your stitched video export has begun. Progress will be shown below."
      });
      
    } catch (err) {
      console.error("Failed to start stitched video export:", err);
      exportError = "Failed to start stitched video export";
      toast.error("Export failed", {
        description: "An error occurred while starting the stitched video export"
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
        description: "Your individual clips export has begun. Progress will be shown below."
      });
      
    } catch (err) {
      console.error("Failed to start individual clips export:", err);
      exportError = "Failed to start individual clips export";
      toast.error("Export failed", {
        description: "An error occurred while starting the individual clips export"
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
              description: progress.errorMessage
            });
          } else if (progress.isCancelled) {
            toast.info("Export cancelled", {
              description: "The export operation was cancelled."
            });
          } else {
            toast.success("Export completed!", {
              description: "Your video export has finished successfully."
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
          description: "The export operation has been cancelled."
        });
      } catch (err) {
        console.error("Failed to cancel export:", err);
        toast.error("Failed to cancel export", {
          description: "An error occurred while cancelling the export."
        });
      }
    }
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <!-- Header with back button and theme switcher -->
    <div class="flex items-center justify-between">
      <Button variant="outline" onclick={goBack} class="flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back to Projects
      </Button>
      <ThemeSwitcher />
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

          <!-- Video Clips section -->
          <div class="mt-8">
            <div class="flex justify-between items-center mb-4">
              <h2 class="text-xl font-semibold">Video Clips</h2>
              <div class="flex items-center gap-4">
                <div class="text-sm text-muted-foreground">
                  {videoClips.length} {videoClips.length === 1 ? 'clip' : 'clips'}
                </div>
                <Button onclick={selectVideoFiles} disabled={addingClip}>
                  {addingClip ? "Adding..." : "Select Video Files"}
                </Button>
              </div>
            </div>

            <!-- Video clip error display -->
            {#if clipError}
              <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4 mb-4">
                <p class="font-medium">Error</p>
                <p class="text-sm">{clipError}</p>
                <Button variant="outline" size="sm" class="mt-2" onclick={() => clipError = ""}>
                  Dismiss
                </Button>
              </div>
            {/if}
            
            <!-- File drop zone -->
            <div class="mb-6">
              <!-- Hidden file input -->
              <input
                bind:this={fileInput}
                type="file"
                multiple
                accept="video/*"
                onchange={handleFileInputChange}
                class="hidden"
              />
              
              <!-- Drop zone with Wails drop target support -->
              <div
                role="button"
                tabindex="0"
                aria-label="Drop video files from file manager or click to browse"
                onclick={openFileDialog}
                onkeydown={handleKeyDown}
                style="--wails-drop-target: drop"
                class="border-2 border-dashed rounded-lg p-8 text-center transition-all duration-200 cursor-pointer
                       {dragActive ? 'border-primary bg-primary/5' : 'border-border hover:border-primary'}
                       {addingClip ? 'pointer-events-none opacity-50' : ''}"
              >
                <div class="flex flex-col items-center gap-4">
                  <svg class="w-12 h-12 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  <div>
                    <p class="text-lg font-medium">
                      {#if addingClip}
                        Adding video clips...
                      {:else if dragActive}
                        Drop video files now
                      {:else}
                        Drop video files here or click to browse
                      {/if}
                    </p>
                    <p class="text-sm text-muted-foreground">
                      Supports MP4, MOV, AVI, MKV, WMV, FLV, WebM, and more
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                      {#if dragActive}
                        Release to add files to your project
                      {:else}
                        Drag video files from your file manager anywhere in this window
                      {/if}
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <!-- Video clips list -->
            {#if loadingClips}
              <div class="text-center py-8 text-muted-foreground">
                <p class="text-lg">Loading video clips...</p>
              </div>
            {:else if videoClips.length === 0}
              <div class="text-center py-8 text-muted-foreground">
                <p class="text-lg">No video clips yet</p>
                <p class="text-sm">Drag and drop video files above or use "Select Video Files" to get started</p>
              </div>
            {:else}
              <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {#each videoClips as clip (clip.id)}
                  <VideoClipCard 
                    {clip}
                    isTranscribing={transcribingClips.has(clip.id)}
                    onPreview={openPreview}
                    onDelete={handleDeleteClip}
                    onViewTranscription={viewTranscription}
                    onStartTranscription={startTranscription}
                    {formatFileSize}
                  />
                {/each}
              </div>
            {/if}
          </div>
          
          <!-- Highlights Timeline section -->
          {#if project}
            <div class="mt-8">
              <ProjectHighlights 
                bind:this={projectHighlightsComponent}
                projectId={projectId}
                onHighlightClick={(highlight) => {
                  console.log('Highlight clicked:', highlight);
                  // The video playback is now handled internally by the ProjectHighlights component
                }}
              />
            </div>
          {/if}
          
          <!-- Export section -->
          {#if project && videoClips.length > 0}
            <div class="mt-8">
              <div class="bg-card text-card-foreground rounded-lg border shadow-sm">
                <div class="p-6">
                  <div class="flex items-center gap-3 mb-6">
                    <Download class="w-6 h-6 text-primary" />
                    <div>
                      <h2 class="text-xl font-semibold">Export Highlights</h2>
                      <p class="text-sm text-muted-foreground">Export your highlighted video segments</p>
                    </div>
                  </div>
                  
                  <!-- Export error display -->
                  {#if exportError}
                    <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-lg p-4 mb-6">
                      <p class="font-medium">Export Error</p>
                      <p class="text-sm">{exportError}</p>
                      <Button variant="outline" size="sm" class="mt-2" onclick={() => exportError = ""}>
                        Dismiss
                      </Button>
                    </div>
                  {/if}
                  
                  <!-- Export options -->
                  <div class="grid gap-4 md:grid-cols-2">
                    <!-- Stitched video option -->
                    <div class="border rounded-lg p-4 space-y-3">
                      <div class="flex items-center gap-3">
                        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                          <Video class="w-5 h-5 text-primary" />
                        </div>
                        <div>
                          <h3 class="font-medium">Single Stitched Video</h3>
                          <p class="text-sm text-muted-foreground">Combine all highlights into one video file</p>
                        </div>
                      </div>
                      <Button 
                        variant="outline" 
                        class="w-full" 
                        onclick={handleStitchedExport}
                        disabled={exporting}
                      >
                        {#if exporting}
                          <svg class="w-4 h-4 mr-2 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
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
                        <div class="w-10 h-10 bg-secondary/50 rounded-lg flex items-center justify-center">
                          <FolderOpen class="w-5 h-5 text-foreground" />
                        </div>
                        <div>
                          <h3 class="font-medium">Individual Clip Files</h3>
                          <p class="text-sm text-muted-foreground">Export each highlight as a separate numbered file</p>
                        </div>
                      </div>
                      <Button 
                        variant="outline" 
                        class="w-full" 
                        onclick={handleIndividualExport}
                        disabled={exporting}
                      >
                        {#if exporting}
                          <svg class="w-4 h-4 mr-2 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
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
                    <div class="mt-6 p-4 bg-primary/5 border border-primary/20 rounded-lg">
                      <div class="flex items-center justify-between mb-3">
                        <div class="flex items-center gap-2">
                          <svg class="w-5 h-5 text-primary animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                          </svg>
                          <h3 class="font-medium">Export Progress</h3>
                        </div>
                        <Button variant="ghost" size="sm" onclick={handleCancelExport}>
                          Cancel
                        </Button>
                      </div>
                      
                      <!-- Progress bar -->
                      <div class="space-y-2 mb-3">
                        <div class="flex justify-between text-sm">
                          <span class="capitalize">{exportProgress.stage}</span>
                          <span>{Math.round(exportProgress.progress * 100)}%</span>
                        </div>
                        <div class="w-full bg-secondary rounded-full h-2">
                          <div 
                            class="bg-primary h-2 rounded-full transition-all duration-300" 
                            style="width: {exportProgress.progress * 100}%"
                          ></div>
                        </div>
                      </div>
                      
                      <!-- Current file info -->
                      {#if exportProgress.currentFile}
                        <div class="text-sm text-muted-foreground mb-2">
                          Processing: {exportProgress.currentFile}
                        </div>
                      {/if}
                      
                      <!-- Files count -->
                      {#if exportProgress.totalFiles > 0}
                        <div class="text-sm text-muted-foreground">
                          {exportProgress.processedFiles} of {exportProgress.totalFiles} files processed
                        </div>
                      {/if}
                    </div>
                  {/if}
                  
                  <!-- Export History -->
                  {#if exportHistory.length > 0}
                    <div class="mt-6">
                      <div class="flex items-center justify-between mb-3">
                        <h3 class="font-medium">Export History</h3>
                        <Button 
                          variant="ghost" 
                          size="sm" 
                          onclick={() => showExportHistory = !showExportHistory}
                        >
                          {showExportHistory ? 'Hide' : 'Show'} History
                        </Button>
                      </div>
                      
                      {#if showExportHistory}
                        <div class="space-y-2 max-h-48 overflow-y-auto">
                          {#each exportHistory as job (job.jobId)}
                            <div class="border rounded-lg p-3 bg-secondary/20">
                              <div class="flex items-center justify-between mb-2">
                                <div class="flex items-center gap-2">
                                  <span class="text-sm font-medium capitalize">{job.stage}</span>
                                  {#if job.isComplete}
                                    {#if job.hasError}
                                      <span class="text-xs bg-destructive text-destructive-foreground px-2 py-1 rounded">Failed</span>
                                    {:else if job.isCancelled}
                                      <span class="text-xs bg-muted text-muted-foreground px-2 py-1 rounded">Cancelled</span>
                                    {:else}
                                      <span class="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">Completed</span>
                                    {/if}
                                  {:else}
                                    <span class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">In Progress</span>
                                  {/if}
                                </div>
                                <span class="text-xs text-muted-foreground">{Math.round(job.progress * 100)}%</span>
                              </div>
                              
                              <div class="text-xs text-muted-foreground space-y-1">
                                <div class="flex justify-between">
                                  <span>Type:</span>
                                  <span>{job.jobId.startsWith('stitched') ? 'Stitched Video' : 'Individual Clips'}</span>
                                </div>
                                <div class="flex justify-between">
                                  <span>Started:</span>
                                  <span>{formatExportDateTime(job.jobId)}</span>
                                </div>
                                {#if job.currentFile}
                                  <div class="flex justify-between">
                                    <span>Current:</span>
                                    <span class="truncate max-w-32" title={job.currentFile}>{job.currentFile}</span>
                                  </div>
                                {/if}
                                {#if job.totalFiles > 0}
                                  <div class="flex justify-between">
                                    <span>Progress:</span>
                                    <span>{job.processedFiles}/{job.totalFiles} segments</span>
                                  </div>
                                {/if}
                                {#if job.hasError && job.errorMessage}
                                  <div class="text-destructive">Error: {job.errorMessage}</div>
                                {/if}
                              </div>
                            </div>
                          {/each}
                        </div>
                      {/if}
                    </div>
                  {/if}
                  
                  <!-- Export info -->
                  <div class="mt-6 p-4 bg-secondary/30 rounded-lg">
                    <div class="flex items-start gap-3">
                      <svg class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <div class="text-sm text-muted-foreground">
                        <p class="font-medium text-foreground mb-1">Export Information:</p>
                        <ul class="space-y-1">
                          <li>• Only video clips with highlights will be exported</li>
                          <li>• A timestamped project folder will be created in your chosen location</li>
                          <li>• Individual clips will be numbered sequentially with time spans</li>
                          <li>• Original video quality will be preserved (H.264/AAC)</li>
                          <li>• Export progress persists across app restarts</li>
                          <li>• Files are organized by project name and export timestamp</li>
                        </ul>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {/if}
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

<!-- Video Preview Dialog -->
<VideoPreviewDialog 
  bind:open={previewDialogOpen}
  bind:video={previewVideo}
/>

<!-- Transcription Viewer Dialog -->
<VideoTranscriptViewer 
  bind:open={transcriptionDialogOpen}
  bind:video={transcriptionVideo}
  projectId={projectId}
  onHighlightsChange={handleHighlightsChange}
/>