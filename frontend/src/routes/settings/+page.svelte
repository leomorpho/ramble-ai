<script>
  import OpenAIConfig from "$lib/components/settings/OpenAIConfig.svelte";
  import OpenRouterConfig from "$lib/components/settings/OpenRouterConfig.svelte";
  import RemoteAIConfig from "$lib/components/settings/RemoteAIConfig.svelte";
  import DevToggle from "$lib/components/settings/DevToggle.svelte";
  import { ArrowLeft } from "@lucide/svelte";
  import { GetUseRemoteAIBackend } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let useRemoteBackend = $state(false);
  let loading = $state(true);

  onMount(() => {
    loadSettings();
  });

  async function loadSettings() {
    try {
      const remoteEnabled = await GetUseRemoteAIBackend();
      useRemoteBackend = remoteEnabled;
    } catch (err) {
      console.error("Failed to load remote backend setting:", err);
    } finally {
      loading = false;
    }
  }

  function handleBackendToggle(newValue) {
    useRemoteBackend = newValue;
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-2xl mx-auto space-y-6">
    <div class="flex items-center gap-3">
      <a href="/" class="text-muted-foreground hover:text-foreground" aria-label="Back to home">
        <ArrowLeft class="w-4 h-4" />
      </a>
      <h1 class="text-2xl font-semibold">Settings</h1>
    </div>

    {#if !loading}
      <DevToggle onToggle={handleBackendToggle} />
      
      {#if useRemoteBackend}
        <RemoteAIConfig />
      {:else}
        <OpenAIConfig />
        <OpenRouterConfig />
      {/if}
    {/if}
  </div>
</main>