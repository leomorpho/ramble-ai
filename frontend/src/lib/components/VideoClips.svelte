<script>
  import { Button } from "$lib/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "$lib/components/ui/dialog";
  import {
    CreateVideoClip,
    GetVideoClipsByProject,
    DeleteVideoClip,
    SelectVideoFiles,
    TranscribeVideoClip,
    BatchTranscribeUntranscribedClips,
    GetOpenAIApiKey,
    GetUseRemoteAIBackend,
    IsFFmpegReady,
  } from "$lib/wailsjs/go/main/App";
  import {
    OnFileDrop,
    OnFileDropOff,
    EventsOn,
    EventsOff,
  } from "$lib/wailsjs/runtime/runtime";
  import { onMount, onDestroy } from "svelte";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";
  import VideoClipCard from "$lib/components/VideoClipCard.svelte";
  import FileDropZone from "$lib/components/FileDropZone.svelte";
  import { updateVideoHighlights } from "$lib/stores/projectHighlights.js";

  // Props
  let { 
    projectId, 
    highlights, 
    onHighlightsChange,
    dragActive = $bindable(),
    videoClips: exposedVideoClips = $bindable()
  } = $props();

  // Video clips state
  let videoClips = $state([]);
  let loadingClips = $state(false);
  let addingClip = $state(false);
  let clipError = $state("");
  let batchTranscribing = $state(false);


  // Delete confirmation dialog state
  let deleteDialogOpen = $state(false);
  let clipToDelete = $state(null);
  
  // Batch transcription confirmation dialog state
  let transcribeAllDialogOpen = $state(false);

  onMount(async () => {
    await loadVideoClips();

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

    return () => {
      observer.disconnect();
    };
  });

  onDestroy(() => {
    // Clean up Wails drag and drop listeners
    OnFileDropOff();
    EventsOff("files-dropped");
  });

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
    if (!files || files.length === 0) {
      console.log("No files to process, ignoring");
      return;
    }

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

    // Don't process if no files provided - this can happen with spurious Wails events
    if (!filePaths || filePaths.length === 0) {
      console.log("No file paths to process, ignoring");
      return;
    }

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

      // Check if user canceled the dialog (empty array or null/undefined)
      if (!selectedFiles || selectedFiles.length === 0) {
        // User canceled or selected no files - this is not an error
        return;
      }

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
      // Only show error if it's not a user cancellation
      if (!err.message || !err.message.toLowerCase().includes("cancel")) {
        clipError = "Failed to select video files";
      }
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

  function handleDeleteClip(clipId) {
    // Find the clip to delete
    const clip = videoClips.find(c => c.id === clipId);
    if (!clip) return;

    // Set the clip to delete and show confirmation dialog
    clipToDelete = clip;
    deleteDialogOpen = true;
  }

  async function confirmDeleteClip() {
    if (!clipToDelete) return;

    try {
      await DeleteVideoClip(clipToDelete.id);
      videoClips = videoClips.filter((clip) => clip.id !== clipToDelete.id);

      // Refresh highlights after deletion (if callback provided)
      if (onHighlightsChange) {
        // This will trigger a refresh in the parent component
        onHighlightsChange();
      }
      
      // Close dialog and reset state
      deleteDialogOpen = false;
      clipToDelete = null;
    } catch (err) {
      console.error("Failed to delete video clip:", err);
      clipError = "Failed to delete video clip";
    }
  }

  function cancelDeleteClip() {
    deleteDialogOpen = false;
    clipToDelete = null;
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  async function startTranscription(clip) {
    try {
      // Check if FFmpeg is ready
      const ffmpegReady = await IsFFmpegReady();
      if (!ffmpegReady) {
        toast.warning('Media processing not ready', {
          description: 'Please wait for setup to complete',
          duration: 3000
        });
        return;
      }
      
      // Check if using remote backend - if so, skip OpenAI API key check
      const useRemoteBackend = await GetUseRemoteAIBackend();
      
      if (!useRemoteBackend) {
        // Only check for OpenAI API key when using local backend
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
      }

      // Update clip state to transcribing
      const clipIndex = videoClips.findIndex((c) => c.id === clip.id);
      if (clipIndex !== -1) {
        videoClips[clipIndex] = {
          ...videoClips[clipIndex],
          transcriptionState: 'transcribing'
        };
        videoClips = [...videoClips]; // Trigger reactivity
      }

      // Show starting toast
      toast.info(`Starting transcription for ${clip.name}`, {
        description: "Transcribing...",
      });

      const result = await TranscribeVideoClip(clip.id);

      if (result.success) {
        // Update the clip with transcription and words data
        const successClipIndex = videoClips.findIndex((c) => c.id === clip.id);
        if (successClipIndex !== -1) {
          videoClips[successClipIndex] = {
            ...videoClips[successClipIndex],
            transcription: result.transcription,
            transcriptionWords: result.words || [],
            transcriptionLanguage: result.language,
            transcriptionDuration: result.duration,
            transcriptionState: 'completed',
          };
          videoClips = [...videoClips]; // Trigger reactivity
        }

        // Show success toast
        toast.success(`Transcription completed for ${clip.name}`, {
          description: "Transcript is now available to view",
        });

        // Refresh highlights since new transcription might have highlights
        if (onHighlightsChange) {
          onHighlightsChange();
        }
      } else {
        // Update clip state to error
        const errorClipIndex = videoClips.findIndex((c) => c.id === clip.id);
        if (errorClipIndex !== -1) {
          videoClips[errorClipIndex] = {
            ...videoClips[errorClipIndex],
            transcriptionState: 'error',
            transcriptionError: result.message
          };
          videoClips = [...videoClips]; // Trigger reactivity
        }
        
        // Show error toast
        toast.error(`Transcription failed for ${clip.name}`, {
          description: result.message,
        });
      }
    } catch (err) {
      console.error("Transcription error:", err);
      
      // Update clip state to error
      const errorClipIndex = videoClips.findIndex((c) => c.id === clip.id);
      if (errorClipIndex !== -1) {
        videoClips[errorClipIndex] = {
          ...videoClips[errorClipIndex],
          transcriptionState: 'error',
          transcriptionError: err.message || "An unexpected error occurred"
        };
        videoClips = [...videoClips]; // Trigger reactivity
      }
      
      toast.error(`Transcription failed for ${clip.name}`, {
        description: "An unexpected error occurred",
      });
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

  function showTranscribeAllConfirmation() {
    transcribeAllDialogOpen = true;
  }

  async function batchTranscribeUntranscribedClips() {
    try {
      batchTranscribing = true;
      clipError = "";

      // Check if FFmpeg is ready
      const ffmpegReady = await IsFFmpegReady();
      if (!ffmpegReady) {
        toast.warning('Media processing not ready', {
          description: 'Please wait for setup to complete',
          duration: 3000
        });
        batchTranscribing = false;
        return;
      }

      // Check if using remote backend - if so, skip OpenAI API key check
      const useRemoteBackend = await GetUseRemoteAIBackend();
      
      if (!useRemoteBackend) {
        // Only check for OpenAI API key when using local backend
        const apiKey = await GetOpenAIApiKey();
        if (!apiKey || apiKey.trim() === "") {
          toast.error("OpenAI API Key Required", {
            description: "Please configure your OpenAI API key in settings to use transcription.",
          });
          setTimeout(() => {
            goto("/settings");
          }, 2000);
          return;
        }
      }

      // Count untranscribed clips
      const untranscribedClips = videoClips.filter(clip => 
        !clip.transcription || 
        clip.transcription.trim() === "" || 
        clip.transcriptionState === 'error'
      );

      if (untranscribedClips.length === 0) {
        toast.info("No untranscribed clips found", {
          description: "All video clips in this project have already been transcribed.",
        });
        return;
      }

      // Show starting toast
      toast.info(`Starting batch transcription for ${untranscribedClips.length} video clips`, {
        description: "This may take a while...",
      });

      // Update states of clips that will be transcribed
      for (const clip of untranscribedClips) {
        const clipIndex = videoClips.findIndex(c => c.id === clip.id);
        if (clipIndex !== -1) {
          videoClips[clipIndex] = {
            ...videoClips[clipIndex],
            transcriptionState: 'transcribing'
          };
        }
      }
      videoClips = [...videoClips]; // Trigger reactivity

      // Call backend batch transcription
      const result = await BatchTranscribeUntranscribedClips(projectId);

      if (result.success) {
        // Reload video clips to get updated transcription data
        await loadVideoClips();

        // Show success toast
        toast.success("Batch transcription completed", {
          description: `${result.transcribedCount} clips transcribed successfully${result.failedCount > 0 ? `, ${result.failedCount} failed` : ''}`,
        });

        // Refresh highlights since new transcriptions might have highlights
        if (onHighlightsChange) {
          onHighlightsChange();
        }
      } else {
        // Show error toast
        toast.error("Batch transcription failed", {
          description: result.message,
        });
        
        // Reset transcription states on error
        for (const clip of untranscribedClips) {
          const clipIndex = videoClips.findIndex(c => c.id === clip.id);
          if (clipIndex !== -1) {
            videoClips[clipIndex] = {
              ...videoClips[clipIndex],
              transcriptionState: 'error'
            };
          }
        }
        videoClips = [...videoClips]; // Trigger reactivity
      }
    } catch (err) {
      console.error("Batch transcription error:", err);
      toast.error("Batch transcription failed", {
        description: "An unexpected error occurred",
      });
      
      // Reset transcription states on error
      const untranscribedClips = videoClips.filter(clip => 
        !clip.transcription || 
        clip.transcription.trim() === "" || 
        clip.transcriptionState === 'error'
      );
      
      for (const clip of untranscribedClips) {
        const clipIndex = videoClips.findIndex(c => c.id === clip.id);
        if (clipIndex !== -1) {
          videoClips[clipIndex] = {
            ...videoClips[clipIndex],
            transcriptionState: 'error'
          };
        }
      }
      videoClips = [...videoClips]; // Trigger reactivity
    } finally {
      batchTranscribing = false;
    }
  }

  // Expose videoClips for parent component
  $effect(() => {
    exposedVideoClips = videoClips;
  });
</script>

<div class="space-y-4">
  <div class="flex justify-between items-center">
    <h3 class="text-lg font-semibold">Video Clips</h3>
    <div class="flex gap-2">
      <Button
        onclick={showTranscribeAllConfirmation}
        disabled={batchTranscribing || loadingClips}
        variant="outline"
        size="sm"
      >
        {batchTranscribing ? "Transcribing..." : "Transcribe All"}
      </Button>
      <Button
        onclick={selectVideoFiles}
        disabled={addingClip}
        size="sm"
      >
        {addingClip ? "Adding..." : "Add Video Files"}
      </Button>
    </div>
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

<!-- Delete Video Clip Confirmation Dialog -->
<Dialog bind:open={deleteDialogOpen}>
  <DialogContent class="sm:max-w-[425px]">
    <DialogHeader>
      <DialogTitle>Delete Video Clip</DialogTitle>
      <DialogDescription>
        Are you sure you want to delete this video clip? This action cannot be undone.
      </DialogDescription>
    </DialogHeader>
    
    {#if clipToDelete}
      <div class="py-4">
        <div class="space-y-2">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium">Name:</span>
            <span class="text-sm text-muted-foreground">{clipToDelete.name}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium">File:</span>
            <span class="text-sm text-muted-foreground">{clipToDelete.fileName}</span>
          </div>
          {#if clipToDelete.transcription}
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium">Status:</span>
              <span class="text-sm text-muted-foreground">Has transcription</span>
            </div>
          {/if}
        </div>
      </div>
    {/if}
    
    <DialogFooter>
      <Button variant="outline" onclick={cancelDeleteClip}>Cancel</Button>
      <Button variant="destructive" onclick={confirmDeleteClip}>Delete Clip</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<!-- Batch Transcription Confirmation Dialog -->
<Dialog bind:open={transcribeAllDialogOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Transcribe All Video Clips?</DialogTitle>
      <DialogDescription>
        This will transcribe all untranscribed video clips in this project using OpenAI Whisper.
        This process may take several minutes depending on the number and length of your videos.
      </DialogDescription>
    </DialogHeader>
    
    <DialogFooter>
      <Button variant="outline" onclick={() => transcribeAllDialogOpen = false}>Cancel</Button>
      <Button onclick={() => { transcribeAllDialogOpen = false; batchTranscribeUntranscribedClips(); }}>Start Transcription</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>