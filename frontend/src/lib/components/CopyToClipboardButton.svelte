<script lang="ts">
  import { Check, Copy } from "@lucide/svelte";
  import { Button } from "./ui/button";
  import { toast } from "svelte-sonner";

  let { text, confirmationText="Copied to clipboard", failureText="Failed to copy to clipboard" } = $props();

  let isCopied = $state(false);

  async function copyToClipboard(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      isCopied = true;
      toast.success(confirmationText);

      // Reset the copied state after 2 seconds
      setTimeout(() => {
        isCopied = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
      toast.error(failureText);
    }
  }
</script>

<Button
  variant="ghost"
  size="sm"
  onclick={() => copyToClipboard(text)}
  class="flex-shrink-0 transition-all"
  disabled={isCopied}
>
  {#if isCopied}
    <Check class="w-4 h-4 text-green-600" />
  {:else}
    <Copy class="w-4 h-4" />
  {/if}
</Button>
