<script>
  import { Button } from "$lib/components/ui/button";
  import { Sun, Moon } from "@lucide/svelte";
  import { SaveThemePreference } from "$lib/wailsjs/go/main/App";
  import { getTheme, setTheme } from "$lib/stores/theme.js";

  let currentTheme = $state("light");

  // Initialize theme on mount
  $effect(() => {
    currentTheme = getTheme();
  });

  async function toggleTheme() {
    const newTheme = currentTheme === "dark" ? "light" : "dark";
    setTheme(newTheme);
    currentTheme = newTheme;
    
    try {
      await SaveThemePreference(newTheme);
    } catch (error) {
      console.error("Failed to save theme preference:", error);
    }
  }
</script>

<Button
  variant="ghost"
  size="icon"
  onclick={toggleTheme}
  class="h-9 w-9"
  title={currentTheme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
>
  {#if currentTheme === "dark"}
    <Sun class="h-4 w-4" />
  {:else}
    <Moon class="h-4 w-4" />
  {/if}
  <span class="sr-only">Toggle theme</span>
</Button>