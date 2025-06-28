<script>
  import { Button } from "$lib/components/ui/button";
  import { GetOpenAIApiKey, SaveOpenAIApiKey, DeleteOpenAIApiKey, TestOpenAIApiKey } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let openaiApiKey = $state("");
  let saved = $state(false);
  let loading = $state(false);
  let error = $state("");
  let showApiKey = $state(false);
  let testing = $state(false);
  let testResult = $state(null);

  onMount(() => {
    loadApiKey();
  });

  async function loadApiKey() {
    try {
      loading = true;
      error = "";
      const savedKey = await GetOpenAIApiKey();
      if (savedKey) {
        openaiApiKey = savedKey;
      }
    } catch (err) {
      console.error("Failed to load API key:", err);
      error = "Failed to load API key";
    } finally {
      loading = false;
    }
  }

  async function saveApiKey() {
    try {
      loading = true;
      error = "";
      testResult = null;
      await SaveOpenAIApiKey(openaiApiKey);
      saved = true;
      
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to save API key:", err);
      error = "Failed to save API key";
    } finally {
      loading = false;
    }
  }

  async function clearApiKey() {
    try {
      loading = true;
      error = "";
      testResult = null;
      openaiApiKey = "";
      await DeleteOpenAIApiKey();
      saved = true;
      
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to clear API key:", err);
      error = "Failed to clear API key";
    } finally {
      loading = false;
    }
  }

  function toggleApiKeyVisibility() {
    showApiKey = !showApiKey;
  }

  async function testApiKey() {
    try {
      testing = true;
      error = "";
      testResult = null;
      
      const result = await TestOpenAIApiKey();
      testResult = result;
    } catch (err) {
      console.error("Failed to test API key:", err);
      error = "Failed to test API key";
      testResult = null;
    } finally {
      testing = false;
    }
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-2xl mx-auto space-y-6">
    <div class="flex items-center gap-4">
      <a href="/" class="text-muted-foreground hover:text-foreground transition-colors" aria-label="Back to home">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </a>
      <h1 class="text-3xl font-bold text-primary">Settings</h1>
    </div>

    <div class="bg-card border rounded-lg p-6 space-y-6">
      <div class="space-y-2">
        <h2 class="text-xl font-semibold">OpenAI Configuration</h2>
        <p class="text-muted-foreground text-sm">
          Configure your OpenAI API key to enable Whisper transcription features.
        </p>
      </div>

      <div class="space-y-4">
        <div class="space-y-2">
          <label for="api-key" class="text-sm font-medium">
            OpenAI API Key
          </label>
          <div class="relative">
            <input
              id="api-key"
              type={showApiKey ? "text" : "password"}
              bind:value={openaiApiKey}
              placeholder="sk-..."
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
          <p class="text-xs text-muted-foreground">
            Your API key is stored securely in the local database and never sent to external servers.
          </p>
        </div>

        <div class="flex gap-2">
          <Button 
            onclick={saveApiKey}
            disabled={loading || testing || !openaiApiKey.trim()}
            class="flex-1"
          >
            {loading ? "Saving..." : saved ? "Saved!" : "Save API Key"}
          </Button>
          
          {#if openaiApiKey}
            <Button 
              variant="outline" 
              onclick={testApiKey}
              disabled={loading || testing}
              class="flex-1"
            >
              {testing ? "Testing..." : "Test API Key"}
            </Button>
          {/if}
          
          {#if openaiApiKey}
            <Button 
              variant="outline" 
              onclick={clearApiKey}
              disabled={loading || testing}
              class="flex-1"
            >
              {loading ? "Clearing..." : "Clear"}
            </Button>
          {/if}
        </div>

        {#if error}
          <div class="bg-destructive/10 text-destructive border border-destructive/20 rounded-md p-3 text-sm">
            {error}
          </div>
        {/if}

        {#if testResult}
          <div class="rounded-md p-3 text-sm border {testResult.valid ? 'bg-green-50 border-green-200 text-green-800' : 'bg-red-50 border-red-200 text-red-800'}">
            <div class="flex items-start gap-2">
              {#if testResult.valid}
                <svg class="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              {:else}
                <svg class="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              {/if}
              <div>
                <p class="font-medium">{testResult.valid ? "Success" : "Error"}</p>
                <p>{testResult.message}</p>
                {#if testResult.model}
                  <p class="text-xs mt-1 opacity-75">Available model: {testResult.model}</p>
                {/if}
              </div>
            </div>
          </div>
        {/if}

        {#if saved && !error}
          <div class="bg-green-50 border border-green-200 text-green-800 rounded-md p-3 text-sm">
            API key saved successfully!
          </div>
        {/if}
      </div>
    </div>

    <div class="bg-card border rounded-lg p-6 space-y-4">
      <h3 class="text-lg font-medium">About Whisper Integration</h3>
      <div class="text-sm text-muted-foreground space-y-2">
        <p>
          With your OpenAI API key configured, you'll be able to use Whisper for:
        </p>
        <ul class="list-disc list-inside space-y-1 ml-4">
          <li>Automatic transcription of audio files</li>
          <li>Speech-to-text conversion for video content</li>
          <li>Subtitle generation</li>
        </ul>
        <p class="text-xs">
          Learn more about OpenAI API pricing at 
          <a href="https://openai.com/pricing" target="_blank" class="text-primary hover:underline">
            openai.com/pricing
          </a>
        </p>
      </div>
    </div>
  </div>
</main>