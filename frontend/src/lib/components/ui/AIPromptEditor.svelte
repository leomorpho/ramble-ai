<script>
  import { Label } from "$lib/components/ui/label";
  import { Button } from "$lib/components/ui/button";
  import AutoResizeTextarea from "./AutoResizeTextarea.svelte";

  let {
    customPrompt = $bindable(""),
    defaultPrompt = "",
    label = "AI Instructions",
    description = "Customize the prompt for AI processing.",
    placeholder = "AI instructions...",
    rows = 6,
    showResetButton = true
  } = $props();
</script>

<div class="space-y-2">
  <Label for="custom-prompt">{label}</Label>
  <AutoResizeTextarea
    id="custom-prompt"
    bind:value={customPrompt}
    {placeholder}
    minHeight={60}
    maxHeight={500}
  />
  <p class="text-xs text-muted-foreground">
    {description}
  </p>
  
  {#if showResetButton && defaultPrompt}
    <div class="flex justify-start">
      <Button
        variant="outline"
        size="sm"
        onclick={() => {
          customPrompt = defaultPrompt;
        }}
        disabled={customPrompt === defaultPrompt}
      >
        Reset to Default
      </Button>
    </div>
  {/if}
</div>