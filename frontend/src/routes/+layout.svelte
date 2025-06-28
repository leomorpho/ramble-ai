<script>
  import '../app.css';
  import { initializeTheme } from '$lib/stores/theme.js';
  import { onMount } from 'svelte';

  onMount(() => {
    initializeTheme();
  });
</script>

<slot />