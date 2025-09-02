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
        duration: 10000
      });
    });
    
    EventsOn('ffmpeg_error', (error) => {
      toast.dismiss('ffmpeg');
      
      // Parse error for better user messaging
      let title = 'Media processing setup failed';
      let description = error;
      let duration = 10000; // Longer duration for error messages
      
      if (error.includes('Platform not supported')) {
        title = 'Unsupported platform';
        description = `Your system (${error.split(': ')[1]}) is not yet supported. Please contact support.`;
      } else if (error.includes('network issue')) {
        title = 'Download failed';
        description = 'Please check your internet connection and try again.';
      } else if (error.includes('Permission error')) {
        title = 'Permission denied';
        description = 'The app needs permission to download files. Please check your system settings.';
      } else if (error.includes('Insufficient disk space')) {
        title = 'Not enough space';
        description = 'Please free up at least 100MB of disk space and try again.';
      } else if (error.includes('FFmpeg verification failed')) {
        // Show detailed diagnostics for verification failures
        title = 'FFmpeg verification failed';
        const details = error.split(': ').slice(1).join(': ');
        description = details || 'The downloaded file could not be verified. Please try again.';
        duration = 15000; // Even longer for detailed errors
      }
      
      toast.error(title, {
        description: description,
        duration: duration,
        action: {
          label: 'Retry',
          onClick: () => window.location.reload()
        }
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
