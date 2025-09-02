<script>
  import "../app.css";
  import { initializeTheme } from "$lib/stores/theme.js";
  import { onMount, onDestroy } from "svelte";
  import { Toaster } from "$lib/components/ui/sonner/index.js";
  import { toast } from "svelte-sonner";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";

  onMount(() => {
    initializeTheme();
    
    // Listen for FFmpeg events (bundled FFmpeg is immediately available)
    EventsOn('ffmpeg_ready', () => {
      // FFmpeg is bundled and ready immediately - no need for setup messages
      console.log('FFmpeg initialized successfully');
    });
    
    EventsOn('ffmpeg_error', (error) => {
      // Only show critical errors for bundled FFmpeg
      toast.error('Video processing unavailable', {
        description: 'FFmpeg not found in app bundle. Please reinstall the application.',
        duration: 15000,
        action: {
          label: 'Reload',
          onClick: () => window.location.reload()
        }
      });
    });
  });
  
  onDestroy(() => {
    EventsOff('ffmpeg_ready');
    EventsOff('ffmpeg_error');
  });
</script>

<Toaster closeButton richColors expand={false} />
<slot />
