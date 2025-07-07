<script>
  import '../app.css';
  import { initializeTheme } from '$lib/stores/theme.js';
  import { onMount } from 'svelte';
  import { Toaster } from "$lib/components/ui/sonner/index.js";

  onMount(() => {
    initializeTheme();
  });
</script>

<Toaster />
<slot />