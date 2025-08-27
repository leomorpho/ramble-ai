<script>
  import { Button } from "$lib/components/ui/button";
  import { Cloud, Server, Key } from "@lucide/svelte";
  import { 
    GetUseRemoteAIBackend, 
    SaveUseRemoteAIBackend, 
    GetRemoteAIBackendURL,
    SaveRemoteAIBackendURL,
    GetRambleAIApiKey, 
    SaveRambleAIApiKey, 
    DeleteRambleAIApiKey 
  } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let useRemoteBackend = $state(false);
  let backendURL = $state("");
  let rambleApiKey = $state("");
  let loading = $state(false);
  let saved = $state(false);
  let error = $state("");
  let showRambleApiKey = $state(false);

  onMount(() => {
    loadSettings();
  });

  async function loadSettings() {
    try {
      loading = true;
      error = "";
      
      const [remoteEnabled, url, apiKey] = await Promise.all([
        GetUseRemoteAIBackend(),
        GetRemoteAIBackendURL(),
        GetRambleAIApiKey()
      ]);
      
      useRemoteBackend = remoteEnabled;
      backendURL = url || "";
      rambleApiKey = apiKey || "";
    } catch (err) {
      console.error("Failed to load remote AI settings:", err);
      error = "Failed to load remote AI settings";
    } finally {
      loading = false;
    }
  }

  async function saveSettings() {
    try {
      loading = true;
      error = "";
      
      await Promise.all([
        SaveUseRemoteAIBackend(useRemoteBackend),
        SaveRemoteAIBackendURL(backendURL),
        SaveRambleAIApiKey(rambleApiKey)
      ]);
      
      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to save remote AI settings:", err);
      error = "Failed to save remote AI settings";
    } finally {
      loading = false;
    }
  }

  async function clearSettings() {
    try {
      loading = true;
      error = "";
      
      useRemoteBackend = false;
      backendURL = "";
      rambleApiKey = "";
      
      await Promise.all([
        SaveUseRemoteAIBackend(false),
        SaveRemoteAIBackendURL(""),
        DeleteRambleAIApiKey()
      ]);
      
      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to clear remote AI settings:", err);
      error = "Failed to clear remote AI settings";
    } finally {
      loading = false;
    }
  }

  function toggleApiKeyVisibility() {
    showRambleApiKey = !showRambleApiKey;
  }
</script>

<div class="bg-card border rounded-lg p-6 space-y-6">
  <div class="space-y-2">
    <h2 class="text-xl font-semibold flex items-center gap-2">
      <Cloud class="w-5 h-5 text-primary" />
      Remote AI Backend
    </h2>
    <p class="text-muted-foreground text-sm">
      Use Ramble AI's managed backend service for AI processing instead of your own API keys.
    </p>
  </div>

  <div class="space-y-4">
    <!-- Toggle Switch -->
    <div class="flex items-center justify-between p-4 bg-secondary/50 rounded-lg">
      <div class="space-y-1">
        <div class="font-medium">Use Remote Backend</div>
        <div class="text-sm text-muted-foreground">
          Enable to use Ramble AI's managed service
        </div>
      </div>
      <label class="relative inline-flex items-center cursor-pointer">
        <input
          type="checkbox"
          bind:checked={useRemoteBackend}
          disabled={loading}
          class="sr-only peer"
        />
        <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
      </label>
    </div>

    {#if useRemoteBackend}
      <!-- Backend URL -->
      <div class="space-y-2">
        <label for="backend-url" class="text-sm font-medium flex items-center gap-2">
          <Server class="w-4 h-4" />
          Backend URL
        </label>
        <input
          id="backend-url"
          type="url"
          bind:value={backendURL}
          placeholder="https://your-backend.com"
          disabled={loading}
          class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
        />
        <p class="text-xs text-muted-foreground">
          Enter the URL of your Ramble AI backend service
        </p>
      </div>

      <!-- API Key -->
      <div class="space-y-2">
        <label for="ramble-api-key" class="text-sm font-medium flex items-center gap-2">
          <Key class="w-4 h-4" />
          Ramble AI API Key
        </label>
        <div class="relative">
          <input
            id="ramble-api-key"
            type={showRambleApiKey ? "text" : "password"}
            bind:value={rambleApiKey}
            placeholder="ra-..."
            disabled={loading}
            class="w-full px-3 py-2 pr-10 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
          />
          <button
            type="button"
            onclick={toggleApiKeyVisibility}
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground transition-colors"
            disabled={loading}
          >
            {#if showRambleApiKey}
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
              </svg>
            {:else}
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            {/if}
          </button>
        </div>
        <p class="text-xs text-muted-foreground">
          Get your API key from the Ramble AI dashboard after subscribing
        </p>
      </div>

      <!-- How to get API key -->
      {#if !rambleApiKey.trim()}
        <div class="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 space-y-3">
          <h3 class="text-sm font-medium text-blue-900 dark:text-blue-100">
            Get your Ramble AI API Key:
          </h3>
          <ol class="text-sm text-blue-700 dark:text-blue-300 space-y-1 ml-4">
            <li>1. Visit the Ramble AI dashboard</li>
            <li>2. Sign up and choose a subscription plan</li>
            <li>3. Generate an API key in your account settings</li>
            <li>4. Copy and paste the key above</li>
          </ol>
        </div>
      {/if}
    {:else}
      <div class="text-sm text-muted-foreground bg-secondary/30 rounded-lg p-4">
        Remote backend is disabled. You'll use your own OpenAI/OpenRouter API keys for AI processing.
      </div>
    {/if}

    <!-- Save/Clear buttons -->
    <div class="flex gap-2">
      <Button 
        onclick={saveSettings}
        disabled={loading || (useRemoteBackend && (!backendURL.trim() || !rambleApiKey.trim()))}
        class="flex-1"
      >
        {loading ? "Saving..." : saved ? "Saved!" : "Save Settings"}
      </Button>
      
      <Button 
        variant="outline" 
        onclick={clearSettings}
        disabled={loading}
        class="flex-1"
      >
        {loading ? "Clearing..." : "Clear"}
      </Button>
    </div>

    {#if error}
      <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-md p-3 text-sm">
        {error}
      </div>
    {/if}

    {#if saved && !error}
      <div class="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 rounded-md p-3 text-sm">
        Remote AI settings saved successfully!
      </div>
    {/if}
  </div>
</div>