<script>
  import "../app.css";
  import { initializeTheme } from "$lib/stores/theme.js";
  import { onMount, onDestroy } from "svelte";
  import { Toaster } from "$lib/components/ui/sonner/index.js";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";
  import { CheckFFmpegStatus } from "$lib/wailsjs/go/main/App";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";

  onMount(async () => {
    initializeTheme();
    
    // Listen for FFmpeg events
    EventsOn('ffmpeg_ready', () => {
      console.log('FFmpeg initialized successfully');
    });
    
    EventsOn('ffmpeg_not_found', (message) => {
      console.log('FFmpeg not found:', message);
      // Redirect to installation page instead of showing toast
      if (!$page.url.pathname.includes('/install-ffmpeg')) {
        goto('/install-ffmpeg');
      }
    });
    
    EventsOn('ffmpeg_error', (...args) => {
      console.error('FFmpeg Error args:', args);
      // Redirect to installation page for FFmpeg errors
      if (!$page.url.pathname.includes('/install-ffmpeg')) {
        goto('/install-ffmpeg');
      }
    });

    // After event listeners are set up, check FFmpeg status
    // This ensures events are properly received by the frontend
    setTimeout(() => {
      CheckFFmpegStatus();
    }, 100);
  });
  
  onDestroy(() => {
    EventsOff('ffmpeg_ready');
    EventsOff('ffmpeg_not_found');
    EventsOff('ffmpeg_error');
  });
</script>

<Toaster closeButton richColors expand={false} />
<slot />
