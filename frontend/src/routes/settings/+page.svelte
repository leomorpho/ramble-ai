<script>
  import OpenAIConfig from "$lib/components/settings/OpenAIConfig.svelte";
  import OpenRouterConfig from "$lib/components/settings/OpenRouterConfig.svelte";
  import { ArrowLeft } from "@lucide/svelte";
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-2xl mx-auto space-y-6">
    <div class="flex items-center gap-3">
      <a href="/" class="text-muted-foreground hover:text-foreground" aria-label="Back to home">
        <ArrowLeft class="w-4 h-4" />
      </a>
      <h1 class="text-2xl font-semibold">Settings</h1>
    </div>

    <OpenAIConfig />
    <OpenRouterConfig />
  </div>
</main>