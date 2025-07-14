<script>
  import OpenAIConfig from "$lib/components/settings/OpenAIConfig.svelte";
  import OpenRouterConfig from "$lib/components/settings/OpenRouterConfig.svelte";
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

    <OpenAIConfig />
    <OpenRouterConfig />
  </div>
</main>