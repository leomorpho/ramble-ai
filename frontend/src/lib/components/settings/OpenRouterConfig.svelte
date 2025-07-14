<script>
  import { Button } from "$lib/components/ui/button";
  import { GetOpenRouterApiKey, SaveOpenRouterApiKey, DeleteOpenRouterApiKey } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let openrouterApiKey = $state("");
  let openrouterSaved = $state(false);
  let openrouterLoading = $state(false);
  let openrouterError = $state("");
  let showOpenRouterApiKey = $state(false);

  onMount(() => {
    loadOpenRouterApiKey();
  });

  async function loadOpenRouterApiKey() {
    try {
      openrouterLoading = true;
      openrouterError = "";
      const savedKey = await GetOpenRouterApiKey();
      if (savedKey) {
        openrouterApiKey = savedKey;
      }
    } catch (err) {
      console.error("Failed to load OpenRouter API key:", err);
      openrouterError = "Failed to load OpenRouter API key";
    } finally {
      openrouterLoading = false;
    }
  }

  async function saveOpenRouterApiKey() {
    try {
      openrouterLoading = true;
      openrouterError = "";
      await SaveOpenRouterApiKey(openrouterApiKey);
      openrouterSaved = true;
      
      setTimeout(() => {
        openrouterSaved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to save OpenRouter API key:", err);
      openrouterError = "Failed to save OpenRouter API key";
    } finally {
      openrouterLoading = false;
    }
  }

  async function clearOpenRouterApiKey() {
    try {
      openrouterLoading = true;
      openrouterError = "";
      openrouterApiKey = "";
      await DeleteOpenRouterApiKey();
      openrouterSaved = true;
      
      setTimeout(() => {
        openrouterSaved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to clear OpenRouter API key:", err);
      openrouterError = "Failed to clear OpenRouter API key";
    } finally {
      openrouterLoading = false;
    }
  }

  function toggleOpenRouterApiKeyVisibility() {
    showOpenRouterApiKey = !showOpenRouterApiKey;
  }

</script>

<div class="bg-card border rounded-lg p-6 space-y-6">
  <div class="space-y-2">
    <h2 class="text-xl font-semibold">OpenRouter Configuration</h2>
    <p class="text-muted-foreground text-sm">
      Configure your OpenRouter API key to enable AI-powered segment reordering features.
    </p>
  </div>

  {#if !openrouterApiKey.trim()}
    <div class="bg-secondary/50 border border-border rounded-lg p-4 space-y-3">
      <h3 class="text-sm font-medium text-foreground">🧠 First time setup? Get your OpenRouter API key:</h3>
      <ol class="text-sm text-muted-foreground space-y-2 ml-4">
        <li class="flex items-start gap-2">
          <span class="bg-primary/10 text-primary text-xs px-1.5 py-0.5 rounded font-mono">1</span>
          <div>
            Go to <a href="https://openrouter.ai/keys" target="_blank" class="text-primary hover:underline font-medium">openrouter.ai/keys</a>
          </div>
        </li>
        <li class="flex items-start gap-2">
          <span class="bg-primary/10 text-primary text-xs px-1.5 py-0.5 rounded font-mono">2</span>
          <div>Sign in with Google, GitHub, or create an account</div>
        </li>
        <li class="flex items-start gap-2">
          <span class="bg-primary/10 text-primary text-xs px-1.5 py-0.5 rounded font-mono">3</span>
          <div>Click "Create Key" and give it a name</div>
        </li>
        <li class="flex items-start gap-2">
          <span class="bg-primary/10 text-primary text-xs px-1.5 py-0.5 rounded font-mono">4</span>
          <div>Copy the key (starts with "sk-or-") and paste it below</div>
        </li>
        <li class="flex items-start gap-2">
          <span class="bg-primary/10 text-primary text-xs px-1.5 py-0.5 rounded font-mono">5</span>
          <div>💡 <strong>Pro tip:</strong> Add credits to your account - you get $1 free to start!</div>
        </li>
      </ol>
      <div class="text-xs text-muted-foreground mt-3">
        <strong>What you'll get:</strong> Intelligent reordering of video segments, content-aware highlight organization, and AI-powered editing suggestions.
        <br>
        <strong>Cost:</strong> Pay-per-use, typically $0.001-0.01 per request. <a href="https://openrouter.ai/pricing" target="_blank" class="text-primary hover:underline">See pricing details</a>
      </div>
    </div>
  {/if}

  <div class="space-y-4">
    <div class="space-y-2">
      <label for="openrouter-api-key" class="text-sm font-medium">
        OpenRouter API Key
      </label>
      <div class="relative">
        <input
          id="openrouter-api-key"
          type={showOpenRouterApiKey ? "text" : "password"}
          bind:value={openrouterApiKey}
          placeholder="sk-or-..."
          disabled={openrouterLoading}
          class="w-full px-3 py-2 pr-10 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
        />
        <button
          type="button"
          onclick={toggleOpenRouterApiKeyVisibility}
          class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground transition-colors"
          disabled={openrouterLoading}
        >
          {#if showOpenRouterApiKey}
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
        Your API key is stored securely in the local database and never sent to external servers.
      </p>
    </div>

    <div class="flex gap-2">
      <Button 
        onclick={saveOpenRouterApiKey}
        disabled={openrouterLoading || !openrouterApiKey.trim()}
        class="flex-1"
      >
        {openrouterLoading ? "Saving..." : openrouterSaved ? "Saved!" : "Save API Key"}
      </Button>
      
      
      {#if openrouterApiKey}
        <Button 
          variant="outline" 
          onclick={clearOpenRouterApiKey}
          disabled={openrouterLoading}
          class="flex-1"
        >
          {openrouterLoading ? "Clearing..." : "Clear"}
        </Button>
      {/if}
    </div>

    {#if openrouterError}
      <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-md p-3 text-sm">
        {openrouterError}
      </div>
    {/if}


    {#if openrouterSaved && !openrouterError}
      <div class="bg-green-50 border border-green-200 text-green-800 rounded-md p-3 text-sm">
        OpenRouter API key saved successfully!
      </div>
    {/if}
  </div>
</div>