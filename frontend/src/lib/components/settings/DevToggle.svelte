<script>
  import { Button } from "$lib/components/ui/button";
  import { Settings2 } from "@lucide/svelte";
  import { 
    GetUseRemoteAIBackend,
    SaveUseRemoteAIBackend,
    IsDevMode,
    IsRemoteBackendOverriddenByEnv,
    SeedDevAPIKey
  } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let { onToggle = () => {}, onAPIKeySeeded = () => {} } = $props();

  let isDevMode = $state(false);
  let useRemoteBackend = $state(false);
  let loading = $state(false);
  let saved = $state(false);
  let error = $state("");
  let isOverriddenByEnv = $state(false);

  onMount(() => {
    loadSettings();
  });

  async function loadSettings() {
    try {
      loading = true;
      error = "";
      
      const [devMode, remoteEnabled, envOverride] = await Promise.all([
        IsDevMode(),
        GetUseRemoteAIBackend(),
        IsRemoteBackendOverriddenByEnv()
      ]);
      
      isDevMode = devMode;
      useRemoteBackend = remoteEnabled;
      isOverriddenByEnv = envOverride;
    } catch (err) {
      console.error("Failed to load development settings:", err);
      error = "Failed to load development settings";
    } finally {
      loading = false;
    }
  }

  async function toggleBackend() {
    try {
      loading = true;
      error = "";
      
      const newValue = !useRemoteBackend;
      await SaveUseRemoteAIBackend(newValue);
      useRemoteBackend = newValue;
      
      // If switching to remote mode, seed the development API key
      if (newValue) {
        try {
          await SeedDevAPIKey();
          // Notify parent that API key was seeded
          onAPIKeySeeded();
        } catch (seedErr) {
          console.warn("Failed to seed development API key:", seedErr);
          // Don't fail the toggle operation if seeding fails
        }
      }
      
      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
      
      // Notify parent component about the change
      onToggle(newValue);
    } catch (err) {
      console.error("Failed to toggle backend:", err);
      error = "Failed to toggle backend";
    } finally {
      loading = false;
    }
  }

</script>

{#if isDevMode}
  <div class="bg-orange-50 dark:bg-orange-950/20 border border-orange-200 dark:border-orange-800 rounded-lg p-6 space-y-4">
    <div class="space-y-2">
      <h2 class="text-xl font-semibold flex items-center gap-2 text-orange-900 dark:text-orange-100">
        <Settings2 class="w-5 h-5" />
        Development Settings
      </h2>
      <p class="text-orange-700 dark:text-orange-300 text-sm">
        These settings are only available in development mode for testing purposes.
      </p>
    </div>

    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <div class="space-y-1">
          <label class="text-sm font-medium text-orange-900 dark:text-orange-100">
            Backend Mode
          </label>
          <p class="text-xs text-orange-700 dark:text-orange-300">
            Switch between local API calls and remote PocketBase backend
          </p>
        </div>
        
        <Button 
          onclick={toggleBackend}
          disabled={loading}
          variant={useRemoteBackend ? "default" : "outline"}
          class="min-w-[120px]"
        >
          {#if loading}
            Switching...
          {:else if useRemoteBackend}
            Remote (PB)
          {:else}
            Local APIs
          {/if}
        </Button>
      </div>

      <div class="text-xs text-orange-600 dark:text-orange-400">
        Current mode: <strong>{useRemoteBackend ? "Remote PocketBase Backend" : "Local API Calls"}</strong>
        <br><small>Toggle freely between modes for testing</small>
        {#if isOverriddenByEnv}
          <br><em>Note: Environment variable USE_REMOTE_AI_BACKEND is set but overridden by toggle</em>
        {/if}
      </div>

      {#if error}
        <div class="bg-red-50 dark:bg-red-950/20 text-red-800 dark:text-red-200 border border-red-200 dark:border-red-800 rounded-md p-3 text-sm">
          {error}
        </div>
      {/if}

      {#if saved && !error}
        <div class="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 rounded-md p-3 text-sm">
          Backend mode saved successfully!
          {#if useRemoteBackend}
            <br><small>Development API key has been auto-populated.</small>
          {/if}
        </div>
      {/if}

    </div>
  </div>
{/if}