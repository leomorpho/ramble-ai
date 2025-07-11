<script>
  import { formatTime } from "./timelineUtils.js";
  import { Button } from "$lib/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$lib/components/ui/dialog";
  import { Film } from "@lucide/svelte";

  let { 
    open = $bindable(), 
    highlightToDelete = null, 
    deleting = false,
    onConfirm = () => {},
    onCancel = () => {}
  } = $props();
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[425px]">
    <DialogHeader>
      <DialogTitle>Delete Highlight</DialogTitle>
      <DialogDescription>
        Are you sure you want to delete this highlight? This action cannot be
        undone.
      </DialogDescription>
    </DialogHeader>

    {#if highlightToDelete}
      <div class="space-y-3">
        <div
          class="flex items-center gap-3 p-3 rounded-lg border"
          style="background-color: {highlightToDelete.color}20; border-left: 4px solid {highlightToDelete.color};"
        >
          <Film
            class="w-6 h-6 flex-shrink-0"
            style="color: {highlightToDelete.color}"
          />
          <div class="flex-1 min-w-0">
            <h3 class="font-medium truncate">
              {highlightToDelete.videoClipName}
            </h3>
            <p class="text-sm text-muted-foreground">
              {formatTime(highlightToDelete.start)} - {formatTime(
                highlightToDelete.end
              )}
            </p>
            {#if highlightToDelete.text}
              <p class="text-sm mt-1 italic line-clamp-2">
                "{highlightToDelete.text}"
              </p>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    <div class="flex justify-end gap-2 mt-4">
      <Button variant="outline" onclick={onCancel} disabled={deleting}>
        Cancel
      </Button>
      <Button
        variant="destructive"
        onclick={onConfirm}
        disabled={deleting}
      >
        {#if deleting}
          Deleting...
        {:else}
          Delete Highlight
        {/if}
      </Button>
    </div>
  </DialogContent>
</Dialog>