<script>
  import "../app.css";
  import { initializeTheme } from "$lib/stores/theme.js";
  import { onMount, onDestroy } from "svelte";
  import { Toaster } from "$lib/components/ui/sonner/index.js";
  import { toast } from "svelte-sonner";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";
  import { InstallFFmpeg } from "$lib/wailsjs/go/main/App.js";
  
  async function installFFmpeg() {
    try {
      toast.info('Starting FFmpeg installation...', {
        description: 'This may take a few minutes.',
        duration: 3000
      });
      
      await InstallFFmpeg();
    } catch (error) {
      console.error('FFmpeg installation failed:', error);
      toast.error('Installation failed', {
        description: `Failed to install FFmpeg: ${error}`,
        duration: 10000,
        action: {
          label: 'Try again',
          onClick: () => installFFmpeg()
        }
      });
    }
  }

  onMount(() => {
    initializeTheme();
    
    // Listen for FFmpeg events
    EventsOn('ffmpeg_ready', () => {
      console.log('FFmpeg initialized successfully');
    });
    
    EventsOn('ffmpeg_not_found', (message) => {
      console.log('FFmpeg not found:', message);
      toast.error('Video processing requires FFmpeg', {
        description: 'FFmpeg is required for video transcription and processing. Would you like to install it?',
        duration: 0, // Don't auto-dismiss
        action: {
          label: 'Install FFmpeg',
          onClick: () => installFFmpeg()
        }
      });
    });
    
    EventsOn('ffmpeg_error', (...args) => {
      console.error('FFmpeg Error args:', args);
      let errorMessage = args[0];
      if (Array.isArray(errorMessage) && errorMessage.length > 0) {
        errorMessage = errorMessage[0];
      }
      
      toast.error('FFmpeg Error', {
        description: errorMessage || 'FFmpeg encountered an error.',
        duration: 10000,
        action: {
          label: 'Retry',
          onClick: () => window.location.reload()
        }
      });
    });
    
    // Listen for installation events
    EventsOn('ffmpeg_install_progress', (message) => {
      toast.info('Installing FFmpeg', {
        description: message,
        duration: 3000
      });
    });
    
    EventsOn('ffmpeg_install_complete', (path) => {
      toast.success('FFmpeg installed successfully!', {
        description: `FFmpeg is now installed at ${path}. Please restart the application.`,
        duration: 10000,
        action: {
          label: 'Restart',
          onClick: () => window.location.reload()
        }
      });
    });
  });
  
  onDestroy(() => {
    EventsOff('ffmpeg_ready');
    EventsOff('ffmpeg_not_found');
    EventsOff('ffmpeg_error');
    EventsOff('ffmpeg_install_progress');
    EventsOff('ffmpeg_install_complete');
  });
</script>

<Toaster closeButton richColors expand={false} />
<slot />
