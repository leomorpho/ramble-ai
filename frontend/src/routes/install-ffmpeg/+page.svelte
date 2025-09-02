<script>
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { Separator } from "$lib/components/ui/separator";
  import { InstallFFmpeg, IsFFmpegReady } from "$lib/wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";
  import { onMount, onDestroy } from "svelte";
  import { goto } from "$app/navigation";
  import { Download, CheckCircle, AlertCircle, Loader2, Terminal, ExternalLink } from "@lucide/svelte";
  import { BrowserOpenURL } from "$lib/wailsjs/runtime/runtime";

  let installationState = $state("not_started"); // not_started, installing, success, error
  let installationStep = $state("");
  let errorMessage = $state("");
  let isCheckingFFmpeg = $state(true);

  async function checkFFmpegStatus() {
    try {
      const ready = await IsFFmpegReady();
      if (ready) {
        // FFmpeg is available, redirect to home
        goto("/");
        return;
      }
      isCheckingFFmpeg = false;
    } catch (error) {
      console.error("Error checking FFmpeg status:", error);
      isCheckingFFmpeg = false;
    }
  }

  async function startInstallation() {
    installationState = "installing";
    installationStep = "Starting installation...";
    errorMessage = "";

    try {
      await InstallFFmpeg();
    } catch (error) {
      console.error("FFmpeg installation failed:", error);
      installationState = "error";
      errorMessage = error.toString();
    }
  }

  function openHomebrewGuide() {
    BrowserOpenURL("https://brew.sh/");
  }

  function openFFmpegGuide() {
    BrowserOpenURL("https://ffmpeg.org/download.html#build-mac");
  }

  onMount(() => {
    checkFFmpegStatus();

    // Listen for installation events
    EventsOn("ffmpeg_install_progress", (message) => {
      installationStep = message;
    });

    EventsOn("ffmpeg_install_complete", (path) => {
      installationState = "success";
      installationStep = `FFmpeg installed successfully at ${path}`;
      
      // Redirect to home after a brief delay
      setTimeout(() => {
        goto("/");
      }, 2000);
    });

    EventsOn("ffmpeg_error", (message) => {
      installationState = "error";
      errorMessage = Array.isArray(message) ? message[0] : message;
    });
  });

  onDestroy(() => {
    EventsOff("ffmpeg_install_progress");
    EventsOff("ffmpeg_install_complete");
    EventsOff("ffmpeg_error");
  });
</script>

<svelte:head>
  <title>Install FFmpeg - RambleAI</title>
</svelte:head>

<div class="min-h-screen bg-background flex items-center justify-center p-4">
  <div class="max-w-2xl w-full space-y-8">
    <!-- Header -->
    <div class="text-center space-y-4">
      <div class="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mx-auto">
        {#if isCheckingFFmpeg}
          <Loader2 class="w-8 h-8 text-primary animate-spin" />
        {:else if installationState === "installing"}
          <Download class="w-8 h-8 text-primary animate-pulse" />
        {:else if installationState === "success"}
          <CheckCircle class="w-8 h-8 text-green-500" />
        {:else if installationState === "error"}
          <AlertCircle class="w-8 h-8 text-red-500" />
        {:else}
          <Download class="w-8 h-8 text-primary" />
        {/if}
      </div>
      
      <div class="space-y-2">
        <h1 class="text-3xl font-bold tracking-tight">FFmpeg Required</h1>
        <p class="text-lg text-muted-foreground max-w-lg mx-auto">
          RambleAI needs FFmpeg to process video files and extract audio for transcription.
        </p>
      </div>
    </div>

    {#if isCheckingFFmpeg}
      <div class="text-center">
        <p class="text-muted-foreground">Checking for FFmpeg installation...</p>
      </div>
    {:else}
      <!-- Installation Options -->
      <div class="bg-card border rounded-lg p-6 space-y-6">
        <div class="space-y-4">
          <h2 class="text-xl font-semibold">Installation Options</h2>
          
          <!-- Option 1: Automatic Installation -->
          <div class="border rounded-lg p-4 space-y-3">
            <div class="flex items-center justify-between">
              <div class="space-y-1">
                <h3 class="font-medium flex items-center gap-2">
                  Automatic Installation
                  <Badge variant="secondary">Recommended</Badge>
                </h3>
                <p class="text-sm text-muted-foreground">
                  Download and install FFmpeg automatically
                </p>
              </div>
            </div>

            {#if installationState === "installing"}
              <div class="space-y-2">
                <div class="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 class="w-4 h-4 animate-spin" />
                  {installationStep}
                </div>
                <div class="w-full bg-secondary rounded-full h-2">
                  <div class="bg-primary h-2 rounded-full animate-pulse" style="width: 100%"></div>
                </div>
              </div>
            {:else if installationState === "success"}
              <div class="flex items-center gap-2 text-sm text-green-600">
                <CheckCircle class="w-4 h-4" />
                {installationStep}
              </div>
            {:else if installationState === "error"}
              <div class="space-y-2">
                <div class="flex items-center gap-2 text-sm text-red-600">
                  <AlertCircle class="w-4 h-4" />
                  Installation failed
                </div>
                {#if errorMessage}
                  <p class="text-xs text-muted-foreground bg-secondary p-2 rounded">
                    {errorMessage}
                  </p>
                {/if}
                <Button onclick={startInstallation} size="sm" variant="outline">
                  Retry Installation
                </Button>
              </div>
            {:else}
              <Button onclick={startInstallation} class="w-full">
                <Download class="w-4 h-4 mr-2" />
                Install FFmpeg
              </Button>
            {/if}
          </div>

          <Separator />

          <!-- Option 2: Manual Installation -->
          <div class="space-y-4">
            <h3 class="font-medium">Manual Installation</h3>
            <p class="text-sm text-muted-foreground">
              If automatic installation doesn't work, you can install FFmpeg manually:
            </p>

            <!-- Homebrew Option -->
            <div class="border rounded-lg p-4 space-y-3">
              <div class="flex items-center gap-2">
                <Terminal class="w-4 h-4" />
                <h4 class="font-medium">Using Homebrew</h4>
                <Badge variant="outline">macOS</Badge>
              </div>
              <div class="space-y-2">
                <p class="text-sm text-muted-foreground">
                  If you have Homebrew installed, run this command in Terminal:
                </p>
                <code class="block text-sm bg-secondary p-2 rounded">
                  brew install ffmpeg
                </code>
                <Button variant="ghost" size="sm" onclick={openHomebrewGuide}>
                  <ExternalLink class="w-3 h-3 mr-1" />
                  Install Homebrew
                </Button>
              </div>
            </div>

            <!-- Direct Download Option -->
            <div class="border rounded-lg p-4 space-y-3">
              <div class="flex items-center gap-2">
                <Download class="w-4 h-4" />
                <h4 class="font-medium">Direct Download</h4>
              </div>
              <div class="space-y-2">
                <p class="text-sm text-muted-foreground">
                  Download FFmpeg directly from the official website
                </p>
                <Button variant="ghost" size="sm" onclick={openFFmpegGuide}>
                  <ExternalLink class="w-3 h-3 mr-1" />
                  Download from FFmpeg.org
                </Button>
              </div>
            </div>
          </div>

          <Separator />

          <!-- After Manual Installation -->
          <div class="text-center space-y-2">
            <p class="text-sm text-muted-foreground">
              After manual installation, restart RambleAI to continue
            </p>
            <Button variant="outline" onclick={() => window.location.reload()}>
              Check Again
            </Button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Footer -->
    <div class="text-center text-sm text-muted-foreground">
      <p>FFmpeg is a free, open-source multimedia framework used for video processing.</p>
    </div>
  </div>
</div>