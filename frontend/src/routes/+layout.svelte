<script>
  import "../app.css";
  import { initializeTheme } from "$lib/stores/theme.js";
  import { onMount, onDestroy } from "svelte";
  import { Toaster } from "$lib/components/ui/sonner/index.js";
  import { toast } from "svelte-sonner";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";

  onMount(() => {
    initializeTheme();
    
    // Listen for FFmpeg download events
    EventsOn('ffmpeg_downloading', () => {
      toast.info('Downloading requirements...', {
        description: 'Setting up video processing tools (80MB)',
        duration: Infinity,
        id: 'ffmpeg'
      });
    });
    
    EventsOn('ffmpeg_ready', () => {
      toast.dismiss('ffmpeg');
      toast.success('Ready to transcribe!', {
        description: 'You can now transcribe videos',
        duration: 3000
      });
    });
    
    EventsOn('ffmpeg_error', (error) => {
      toast.dismiss('ffmpeg');
      toast.error('Media processing setup failed', {
        description: error,
        duration: 5000
      });
    });
  });
  
  onDestroy(() => {
    EventsOff('ffmpeg_downloading');
    EventsOff('ffmpeg_ready');
    EventsOff('ffmpeg_error');
  });
</script>

<Toaster closeButton richColors expand={false} />
<slot />
