<script>
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { ExternalLink, Key, Wand2, PlayCircle, ChevronRight, Zap } from "@lucide/svelte";
  import { BrowserOpenURL } from "$lib/wailsjs/runtime/runtime";
  import { GetUseRemoteAIBackend } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let { open = $bindable(false) } = $props();
  let useRemoteBackend = $state(false);
  let backendLoaded = $state(false);

  onMount(async () => {
    try {
      useRemoteBackend = await GetUseRemoteAIBackend();
    } catch (err) {
      console.error("Failed to get backend mode:", err);
      useRemoteBackend = false; // Default to local mode on error
    } finally {
      backendLoaded = true;
    }
  });
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle>Welcome to RambleAI</DialogTitle>
      <DialogDescription>
        Set up your API keys to start transforming your videos
      </DialogDescription>
    </DialogHeader>
    
    <div class="space-y-4 py-4">
      {#if backendLoaded}
        {#if useRemoteBackend}
          <!-- RambleAI API Key Section (Remote Mode) -->
          <div class="border rounded p-4 space-y-3">
            <div class="flex items-center gap-2">
              <Zap class="w-4 h-4" />
              <h3 class="font-medium">RambleAI API Key Setup</h3>
            </div>
            
            <p class="text-sm text-muted-foreground">
              Your RambleAI API key enables all AI features including transcription and content suggestions
            </p>
            
            <div class="space-y-2">
              <ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
                <li>Get your API key from your RambleAI account</li>
                <li>Copy the key (starts with "ra-")</li>
                <li>Go to Settings → Remote AI Backend</li>
                <li>Paste your API key</li>
              </ol>
            </div>
          </div>
        {:else}
          <!-- OpenAI Section (Local Mode) -->
          <div class="border rounded p-4 space-y-3">
            <div class="flex items-center gap-2">
              <Key class="w-4 h-4" />
              <h3 class="font-medium">OpenAI Setup (Required)</h3>
            </div>
            
            <p class="text-sm text-muted-foreground">
              Enables automatic audio transcription for your videos
            </p>
            
            <div class="space-y-2">
              <ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
                <li>
                  Visit <button onclick={() => BrowserOpenURL("https://platform.openai.com/api-keys")} class="text-foreground hover:underline inline-flex items-center gap-1">
                    platform.openai.com <ExternalLink class="w-3 h-3" />
                  </button>
                </li>
                <li>Create new secret key</li>
                <li>Copy key (starts with "sk-")</li>
                <li>Add in Settings</li>
              </ol>
              <p class="text-xs text-muted-foreground">
                Cost: ~$0.006 per minute
              </p>
            </div>
          </div>

          <!-- OpenRouter Section (Local Mode) -->
          <div class="border rounded p-4 space-y-3">
            <div class="flex items-center gap-2">
              <Wand2 class="w-4 h-4" />
              <h3 class="font-medium">OpenRouter Setup (Optional)</h3>
            </div>
            
            <p class="text-sm text-muted-foreground">
              AI-powered reordering and editing suggestions
            </p>
            
            <div class="space-y-2">
              <ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
                <li>
                  Visit <button onclick={() => BrowserOpenURL("https://openrouter.ai/keys")} class="text-foreground hover:underline inline-flex items-center gap-1">
                    openrouter.ai <ExternalLink class="w-3 h-3" />
                  </button>
                </li>
                <li>Sign in or create account</li>
                <li>Create new key</li>
                <li>Copy key (starts with "sk-or-")</li>
                <li>Add in Settings</li>
              </ol>
            </div>
          </div>
        {/if}
      {:else}
        <!-- Loading state -->
        <div class="border rounded p-4 space-y-3">
          <div class="text-center text-muted-foreground">
            <p>Loading setup guide...</p>
          </div>
        </div>
      {/if}

      <!-- Quick Start -->
      <div class="border rounded p-4 space-y-3">
        <div class="flex items-center gap-2">
          <PlayCircle class="w-4 h-4" />
          <h3 class="font-medium">Ready to Start</h3>
        </div>
        
        <div class="text-sm text-muted-foreground space-y-2">
          <div class="flex items-start gap-2">
            <ChevronRight class="w-3 h-3 mt-0.5" />
            <span>Create project</span>
          </div>
          <div class="flex items-start gap-2">
            <ChevronRight class="w-3 h-3 mt-0.5" />
            <span>Upload videos</span>
          </div>
          <div class="flex items-start gap-2">
            <ChevronRight class="w-3 h-3 mt-0.5" />
            <span>AI transcribes content</span>
          </div>
          <div class="flex items-start gap-2">
            <ChevronRight class="w-3 h-3 mt-0.5" />
            <span>Export polished results</span>
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-2">
      <Button variant="outline" onclick={() => open = false} size="sm">
        Later
      </Button>
      <Button asChild size="sm">
        <a href="/settings">Go to Settings</a>
      </Button>
    </div>
  </DialogContent>
</Dialog>