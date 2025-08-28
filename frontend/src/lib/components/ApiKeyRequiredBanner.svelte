<script>
  import { Button } from "$lib/components/ui/button";
  import { ExternalLink, Key, Settings } from "@lucide/svelte";
  import { GetRambleFrontendURL } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let frontendUrl = $state("https://app.ramble.ai"); // fallback

  onMount(async () => {
    try {
      frontendUrl = await GetRambleFrontendURL();
    } catch (err) {
      console.warn("Failed to get frontend URL:", err);
    }
  });
</script>

<div class="border border-orange-200 bg-orange-50 rounded-lg p-6 space-y-4 dark:border-orange-800 dark:bg-orange-950">
  <div class="flex items-start gap-3">
    <div class="flex-shrink-0">
      <Key class="w-5 h-5 text-orange-600 dark:text-orange-400" />
    </div>
    <div class="flex-1 space-y-3">
      <div>
        <h3 class="font-medium text-foreground">API Key Required</h3>
        <p class="text-sm text-muted-foreground mt-1">
          You need a Ramble AI API key to create projects and start transcribing videos. 
          Get 10 transcription hours per month free!
        </p>
      </div>
      
      <div class="flex flex-col sm:flex-row gap-3">
        <Button 
          asChild 
          size="sm" 
          class="flex-1 sm:flex-none"
        >
          <a href={frontendUrl} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2">
            Get Free API Key
            <ExternalLink class="w-3 h-3" />
          </a>
        </Button>
        
        <Button 
          variant="outline" 
          size="sm"
          asChild
          class="flex-1 sm:flex-none"
        >
          <a href="/settings" class="inline-flex items-center gap-2">
            <Settings class="w-4 h-4" />
            I already have a key
          </a>
        </Button>
      </div>
    </div>
  </div>
</div>