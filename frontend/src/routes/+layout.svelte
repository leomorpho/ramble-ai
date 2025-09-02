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
    
    EventsOn('ffmpeg_error', (...args) => {
      // Show detailed error information from backend
      console.error('FFmpeg Error args:', args);
      
      // Extract the actual error message (might be in args[0] if passed as variadic)
      let errorMessage = args[0];
      if (Array.isArray(errorMessage) && errorMessage.length > 0) {
        errorMessage = errorMessage[0];
      }
      
      console.error('FFmpeg Error message:', errorMessage);
      
      toast.error('Video processing unavailable', {
        description: errorMessage || 'FFmpeg not found in app bundle. Please reinstall the application.',
        duration: 20000, // Longer duration for detailed messages
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
