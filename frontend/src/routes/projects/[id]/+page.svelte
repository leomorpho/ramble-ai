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
  import { GetProjectByID, UpdateProject, DeleteProject, CreateVideoClip, GetVideoClipsByProject, UpdateVideoClip, DeleteVideoClip, SelectVideoFiles, GetVideoFileInfo, GetVideoURL, TranscribeVideoClip } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { toast } from "svelte-sonner";
  import TextHighlighter from "$lib/components/TextHighlighter.svelte";

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
  let videoURL = $state("");
  
  // Transcription state
  let transcriptionDialogOpen = $state(false);
  let transcriptionVideo = $state(null);
  let transcribingClips = $state(new Set());

  // Get project ID from route params
  let projectId = $derived(parseInt($page.params.id));

  onMount(async () => {
    await loadProject();
    await loadVideoClips();
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
      videoClips = await GetVideoClipsByProject(projectId);
    } catch (err) {
      console.error("Failed to load video clips:", err);
      clipError = "Failed to load video clips";
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
    
    const fileArray = Array.from(files);
    const videoFiles = fileArray.filter(isVideoFile);
    
    if (videoFiles.length === 0) {
      clipError = "No valid video files found. Please select video files (MP4, MOV, AVI, etc.)";
      addingClip = false;
      return;
    }
    
    for (const file of videoFiles) {
      try {
        // In Wails, we can access the file path directly from the File object
        // For drag & drop, we need to use the file path if available
        let filePath = file.path || file.webkitRelativePath || file.name;
        
        // If we have a real path, use it directly
        if (file.path) {
          const newClip = await CreateVideoClip(projectId, file.path);
          // Check if this clip is already in our list
          if (!videoClips.some(clip => clip.id === newClip.id)) {
            videoClips = [...videoClips, newClip]; // Trigger reactivity
          }
        } else {
          // For files without paths (browser drag & drop), show error
          clipError = `Cannot access file system path for ${file.name}. Please use "Select Video Files" button instead.`;
          break;
        }
      } catch (err) {
        console.error("Failed to add video clip:", err);
        clipError = `Failed to add ${file.name}: ${err.message || err}`;
        break;
      }
    }
    
    addingClip = false;
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
          if (!videoClips.some(clip => clip.id === newClip.id)) {
            videoClips = [...videoClips, newClip]; // Trigger reactivity
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

  function handleDrop(event) {
    event.preventDefault();
    dragActive = false;
    
    const files = event.dataTransfer?.files;
    if (files) {
      handleFiles(files);
    }
  }

  function handleDragOver(event) {
    event.preventDefault();
  }

  function handleDragEnter(event) {
    event.preventDefault();
    dragActive = true;
  }

  function handleDragLeave(event) {
    event.preventDefault();
    // Only hide drag state if we're leaving the drop zone completely
    if (!event.currentTarget.contains(event.relatedTarget)) {
      dragActive = false;
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
    } catch (err) {
      console.error("Failed to delete video clip:", err);
      clipError = "Failed to delete video clip";
    }
  }

  async function openPreview(clip) {
    previewVideo = clip;
    
    // Get video URL for playback
    try {
      const url = await GetVideoURL(clip.filePath);
      videoURL = url;
    } catch (err) {
      console.error("Failed to get video URL:", err);
      videoURL = "";
    }
    
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

  function formatTimestamp(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = (seconds % 60).toFixed(1);
    return `${mins}:${secs.padStart(4, '0')}`;
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
              
              <!-- Drop zone -->
              <div
                role="button"
                tabindex="0"
                aria-label="Drop video files here or click to browse"
                ondrop={handleDrop}
                ondragover={handleDragOver}
                ondragenter={handleDragEnter}
                ondragleave={handleDragLeave}
                onclick={openFileDialog}
                onkeydown={handleKeyDown}
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
                      Note: For drag & drop to work, files must be from a local file manager
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
                  <div class="bg-secondary/30 rounded-lg overflow-hidden border">
                    <!-- Video thumbnail -->
                    {#if clip.exists && clip.thumbnailUrl}
                      <div 
                        class="relative group cursor-pointer" 
                        onclick={() => openPreview(clip)}
                        onkeydown={(e) => e.key === 'Enter' && openPreview(clip)}
                        role="button"
                        tabindex="0"
                        aria-label="Preview video {clip.name}"
                      >
                        <img 
                          src={clip.thumbnailUrl} 
                          alt="Video thumbnail for {clip.name}"
                          class="w-full h-48 object-cover bg-muted"
                          loading="lazy"
                        />
                        <!-- Play overlay -->
                        <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center">
                          <div class="w-16 h-16 bg-white/80 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                            <svg class="w-8 h-8 text-black ml-1" fill="currentColor" viewBox="0 0 24 24">
                              <path d="M8 5v14l11-7z"/>
                            </svg>
                          </div>
                        </div>
                      </div>
                    {:else}
                      <div class="w-full h-48 bg-muted flex items-center justify-center">
                        <div class="text-center text-muted-foreground">
                          <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                          <p class="text-sm">
                            {clip.exists ? 'Generating thumbnail...' : 'Video not found'}
                          </p>
                        </div>
                      </div>
                    {/if}

                    <div class="p-4">
                      <div class="flex justify-between items-start mb-3">
                        <div class="flex-1 min-w-0">
                          <h3 class="font-semibold truncate" title={clip.name}>{clip.name}</h3>
                          <p class="text-sm text-muted-foreground truncate" title={clip.fileName}>
                            {clip.fileName}
                          </p>
                        </div>
                        <Button 
                          variant="ghost" 
                          size="sm" 
                          onclick={() => handleDeleteClip(clip.id)}
                          class="ml-2 text-destructive hover:text-destructive"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </Button>
                      </div>

                    <div class="space-y-2 text-xs text-muted-foreground">
                      <div class="flex justify-between">
                        <span>Format:</span>
                        <span class="font-mono uppercase">{clip.format || 'unknown'}</span>
                      </div>
                      <div class="flex justify-between">
                        <span>Size:</span>
                        <span>{formatFileSize(clip.fileSize || 0)}</span>
                      </div>
                      {#if clip.width && clip.height}
                        <div class="flex justify-between">
                          <span>Resolution:</span>
                          <span>{clip.width}×{clip.height}</span>
                        </div>
                      {/if}
                      {#if clip.duration}
                        <div class="flex justify-between">
                          <span>Duration:</span>
                          <span>{Math.round(clip.duration)}s</span>
                        </div>
                      {/if}
                      <div class="flex justify-between">
                        <span>Status:</span>
                        <span class={clip.exists ? "text-green-600" : "text-destructive"}>
                          {clip.exists ? "Found" : "Missing"}
                        </span>
                      </div>
                    </div>

                    {#if clip.description}
                      <div class="mt-3 pt-3 border-t border-border">
                        <p class="text-sm text-muted-foreground">{clip.description}</p>
                      </div>
                    {/if}

                    <!-- Action buttons -->
                    <div class="mt-3 pt-3 border-t border-border space-y-2">
                      <!-- Preview button -->
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onclick={() => openPreview(clip)}
                        disabled={!clip.exists}
                        class="w-full"
                      >
                        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h8m-5-9v-.5a2.5 2.5 0 015 0V4a2 2 0 012 2v6.5" />
                        </svg>
                        {clip.exists ? 'Preview Video' : 'File Missing'}
                      </Button>

                      <!-- Transcription buttons -->
                      <div class="flex gap-2">
                        {#if clip.transcription}
                          <Button 
                            variant="outline" 
                            size="sm" 
                            onclick={() => viewTranscription(clip)}
                            class="flex-1"
                          >
                            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            View Transcript
                          </Button>
                          <Button 
                            variant="ghost" 
                            size="sm" 
                            onclick={() => startTranscription(clip)}
                            disabled={transcribingClips.has(clip.id) || !clip.exists}
                            class="px-3"
                            title="Re-transcribe"
                          >
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                            </svg>
                          </Button>
                        {:else}
                          <Button 
                            variant="outline" 
                            size="sm" 
                            onclick={() => startTranscription(clip)}
                            disabled={transcribingClips.has(clip.id) || !clip.exists}
                            class="w-full"
                          >
                            {#if transcribingClips.has(clip.id)}
                              <svg class="w-4 h-4 mr-2 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                              </svg>
                              Transcribing...
                            {:else}
                              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                              </svg>
                              Start Transcription
                            {/if}
                          </Button>
                        {/if}
                      </div>
                    </div>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
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

<!-- Video Preview Dialog -->
<Dialog bind:open={previewDialogOpen}>
  <DialogContent class="sm:max-w-[800px] max-h-[90vh]">
    <DialogHeader>
      <DialogTitle>Video Preview</DialogTitle>
      <DialogDescription>
        {#if previewVideo}
          Preview of {previewVideo.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    {#if previewVideo}
      <div class="space-y-4">
        <!-- Video player -->
        <div class="bg-background border rounded-lg overflow-hidden">
          {#if previewVideo.exists && videoURL}
            <video 
              class="w-full h-auto max-h-96" 
              controls 
              preload="metadata"
              src={videoURL}
            >
              <p class="p-4 text-center text-muted-foreground">
                Your browser doesn't support video playback or the video format is not supported.
              </p>
            </video>
          {:else if previewVideo.exists && !videoURL}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <p class="text-lg font-medium">Loading video...</p>
              <p class="text-sm">Preparing video for playback</p>
            </div>
          {:else}
            <div class="p-8 text-center text-muted-foreground">
              <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.864-.833-2.634 0L4.18 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <p class="text-lg font-medium">Video file not found</p>
              <p class="text-sm">The video file may have been moved or deleted</p>
            </div>
          {/if}
        </div>
        
        <!-- Video details -->
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-muted-foreground">Name:</span>
              <span class="font-medium">{previewVideo.name}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-muted-foreground">Format:</span>
              <span class="font-mono uppercase">{previewVideo.format}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-muted-foreground">Size:</span>
              <span>{formatFileSize(previewVideo.fileSize)}</span>
            </div>
          </div>
          <div class="space-y-2">
            {#if previewVideo.width && previewVideo.height}
              <div class="flex justify-between">
                <span class="text-muted-foreground">Resolution:</span>
                <span>{previewVideo.width}×{previewVideo.height}</span>
              </div>
            {/if}
            {#if previewVideo.duration}
              <div class="flex justify-between">
                <span class="text-muted-foreground">Duration:</span>
                <span>{Math.round(previewVideo.duration)}s</span>
              </div>
            {/if}
            <div class="flex justify-between">
              <span class="text-muted-foreground">Status:</span>
              <span class={previewVideo.exists ? "text-green-600" : "text-destructive"}>
                {previewVideo.exists ? "Available" : "Missing"}
              </span>
            </div>
          </div>
        </div>
        
        <!-- File path -->
        <div class="p-3 bg-secondary/30 rounded-lg">
          <p class="text-xs text-muted-foreground mb-1">File Path:</p>
          <p class="text-sm font-mono break-all">{previewVideo.filePath}</p>
        </div>
      </div>
    {/if}
    
    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={() => previewDialogOpen = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>

<!-- Transcription Viewer Dialog -->
<Dialog bind:open={transcriptionDialogOpen}>
  <DialogContent class="sm:max-w-[700px] max-h-[85vh]">
    <DialogHeader>
      <DialogTitle>Video Transcript</DialogTitle>
      <DialogDescription>
        {#if transcriptionVideo}
          Transcript for {transcriptionVideo.name}
        {/if}
      </DialogDescription>
    </DialogHeader>
    
    <div class="overflow-y-auto max-h-[60vh]">
      {#if transcriptionVideo}
        <div class="space-y-4">
          <!-- Video info -->
          <div class="flex items-center gap-3 p-3 bg-secondary/30 rounded-lg">
            <svg class="w-6 h-6 text-muted-foreground flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <div class="flex-1 min-w-0">
              <p class="font-medium truncate">{transcriptionVideo.name}</p>
              <p class="text-sm text-muted-foreground truncate">{transcriptionVideo.fileName}</p>
            </div>
          </div>
        
          <!-- Transcript content with tabs -->
          {#if transcriptionVideo.transcription}
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="font-medium">Transcript</h3>
                <div class="flex gap-2">
                  {#if transcriptionVideo.transcriptionLanguage}
                    <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                      {transcriptionVideo.transcriptionLanguage.toUpperCase()}
                    </span>
                  {/if}
                  {#if transcriptionVideo.transcriptionDuration}
                    <span class="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded-md">
                      {formatTimestamp(transcriptionVideo.transcriptionDuration)}
                    </span>
                  {/if}
                  <Button 
                    variant="outline" 
                    size="sm"
                    onclick={() => navigator.clipboard.writeText(transcriptionVideo.transcription)}
                    class="text-xs"
                  >
                    <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Copy
                  </Button>
                </div>
              </div>

              <Tabs value="full-text" class="w-full">
                <TabsList class="grid w-full grid-cols-2">
                  <TabsTrigger value="full-text">Full Text</TabsTrigger>
                  <TabsTrigger value="word-by-word" disabled={!transcriptionVideo.transcriptionWords || transcriptionVideo.transcriptionWords.length === 0}>
                    Word by Word
                  </TabsTrigger>
                </TabsList>
                
                <TabsContent value="full-text" class="space-y-3">
                  <div class="max-h-64 overflow-y-auto p-4 bg-background border rounded-lg">
                    <div class="text-sm leading-relaxed">
                      <TextHighlighter 
                        text={transcriptionVideo.transcription} 
                        words={transcriptionVideo.transcriptionWords || []} 
                      />
                    </div>
                  </div>
                  <div class="text-xs text-muted-foreground">
                    Character count: {transcriptionVideo.transcription.length}
                  </div>
                </TabsContent>
                
                <TabsContent value="word-by-word" class="space-y-3">
                  {#if transcriptionVideo.transcriptionWords && transcriptionVideo.transcriptionWords.length > 0}
                    <div class="max-h-64 overflow-y-auto p-4 bg-background border rounded-lg space-y-1">
                      {#each transcriptionVideo.transcriptionWords as word, index}
                        <div class="flex items-center gap-3 p-2 hover:bg-secondary/30 rounded-md group">
                          <div class="flex-shrink-0 text-xs text-muted-foreground font-mono bg-secondary px-2 py-1 rounded">
                            {formatTimestamp(word.start)}
                          </div>
                          <div class="flex-1">
                            <span class="text-sm">{word.word.trim()}</span>
                          </div>
                          <div class="flex-shrink-0 text-xs text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                            {(word.end - word.start).toFixed(1)}s
                          </div>
                        </div>
                      {/each}
                    </div>
                    <div class="text-xs text-muted-foreground flex-shrink-0">
                      Word count: {transcriptionVideo.transcriptionWords.length}
                    </div>
                  {:else}
                    <div class="flex-1 flex items-center justify-center text-muted-foreground">
                      <div class="text-center">
                        <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <p class="text-lg font-medium">No word-level timing available</p>
                        <p class="text-sm">Word timestamps weren't generated for this transcription.</p>
                      </div>
                    </div>
                  {/if}
                </TabsContent>
              </Tabs>
            </div>
          {:else}
            <div class="flex-1 flex items-center justify-center text-muted-foreground">
              <div class="text-center">
                <svg class="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="text-lg font-medium">No transcript available</p>
                <p class="text-sm">This video hasn't been transcribed yet.</p>
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    
    <!-- Fixed footer buttons -->
    <div class="flex justify-end gap-2 pt-4 border-t flex-shrink-0">
      <Button variant="outline" onclick={() => transcriptionDialogOpen = false}>
        Close
      </Button>
    </div>
  </DialogContent>
</Dialog>