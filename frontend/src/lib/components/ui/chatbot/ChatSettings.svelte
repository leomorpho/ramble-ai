<script>
  import AIModelSelector from "../AIModelSelector.svelte";
  
  let {
    selectedModel = $bindable("anthropic/claude-sonnet-4"),
    customModelValue = $bindable(""),
    availableModels = []
  } = $props();
</script>

<div class="p-4 space-y-4">
  <div>
    <h3 class="font-medium text-sm mb-3">Chat Settings</h3>
    
    <AIModelSelector
      bind:selectedModel
      bind:customModelValue
      label="AI Model"
      description="Choose the AI model for this conversation"
      {availableModels}
    />
  </div>
</div>