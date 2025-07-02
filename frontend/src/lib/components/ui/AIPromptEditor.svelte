<script>
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Button } from "$lib/components/ui/button";

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
  <Textarea
    id="custom-prompt"
    bind:value={customPrompt}
    {placeholder}
    class="min-h-[120px] resize-none"
    {rows}
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