<script>
  import { Button } from "$lib/components/ui/button";
  import { Key, ExternalLink } from "@lucide/svelte";
  import { 
    GetRambleAIApiKey, 
    SaveRambleAIApiKey, 
    DeleteRambleAIApiKey 
  } from "$lib/wailsjs/go/main/App";
  import { BrowserOpenURL } from "$lib/wailsjs/runtime/runtime";
  import { onMount } from "svelte";

  let rambleApiKey = $state("");
  let loading = $state(false);
  let saved = $state(false);
  let error = $state("");
  let showApiKey = $state(false);

  onMount(() => {
    loadSettings();
  });

  async function loadSettings() {
    try {
      loading = true;
      error = "";
      
      const apiKey = await GetRambleAIApiKey();
      rambleApiKey = apiKey || "";
    } catch (err) {
      console.error("Failed to load Ramble AI settings:", err);
      error = "Failed to load Ramble AI settings";
    } finally {
      loading = false;
    }
  }

  async function saveApiKey() {
    try {
      loading = true;
      error = "";
      
      await SaveRambleAIApiKey(rambleApiKey);
      
      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to save Ramble AI API key:", err);
      error = "Failed to save Ramble AI API key";
    } finally {
      loading = false;
    }
  }

  async function clearApiKey() {
    try {
      loading = true;
      error = "";
      
      rambleApiKey = "";
      await DeleteRambleAIApiKey();
      
      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to clear Ramble AI API key:", err);
      error = "Failed to clear Ramble AI API key";
    } finally {
      loading = false;
    }
  }

  function toggleApiKeyVisibility() {
    showApiKey = !showApiKey;
  }

  function openRambleAILink() {
    BrowserOpenURL("https://ramble.goosebyteshq.com");
  }
</script>

<div class="bg-card border rounded-lg p-6 space-y-6">
    <div class="space-y-2">
      <h2 class="text-xl font-semibold flex items-center gap-2">
        <Key class="w-5 h-5 text-primary" />
        Ramble AI API Key
      </h2>
      <p class="text-muted-foreground text-sm">
        Enter your Ramble AI API key to access the remote backend service.
      </p>
    </div>

    <div class="space-y-4">
      <!-- API Key Input -->
      <div class="space-y-2">
        <label for="ramble-api-key" class="text-sm font-medium">
          API Key
        </label>
        <div class="relative">
          <input
            id="ramble-api-key"
            type={showApiKey ? "text" : "password"}
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
            {#if showApiKey}
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
        <div class="flex justify-end">
          <button
            type="button"
            onclick={openRambleAILink}
            class="text-xs text-primary hover:underline flex items-center gap-1"
          >
            Get new API key here
            <ExternalLink class="w-3 h-3" />
          </button>
        </div>
      </div>

      <!-- How to get API key -->
      {#if !rambleApiKey.trim()}
        <div class="bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 space-y-3">
          <h3 class="text-sm font-medium text-blue-900 dark:text-blue-100">
            Get your Ramble AI API Key:
          </h3>
          <div class="space-y-3">
            <Button 
              variant="outline" 
              onclick={openRambleAILink}
              class="w-full justify-center gap-2 text-blue-700 dark:text-blue-300 border-blue-300 dark:border-blue-700 hover:bg-blue-100 dark:hover:bg-blue-900/30"
            >
              Get API Key at ramble.goosebytes.hq
              <ExternalLink class="w-4 h-4" />
            </Button>
            <p class="text-xs text-blue-600 dark:text-blue-400">
              Sign up, choose a plan, generate an API key, then paste it above.
            </p>
          </div>
        </div>
      {/if}

      <!-- Save/Clear buttons -->
      <div class="flex gap-2">
        <Button 
          onclick={saveApiKey}
          disabled={loading || !rambleApiKey.trim()}
          class="flex-1"
        >
          {loading ? "Saving..." : saved ? "Saved!" : "Save API Key"}
        </Button>
        
        <Button 
          variant="outline" 
          onclick={clearApiKey}
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
          Ramble AI API key saved successfully!
        </div>
      {/if}
    </div>
</div>