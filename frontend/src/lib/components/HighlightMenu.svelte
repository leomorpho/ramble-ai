<script>
  import { Eye, Edit3, Trash2 } from '@lucide/svelte';
  import { Popover, PopoverContent, PopoverTrigger } from "$lib/components/ui/popover";

  let { 
    highlight,
    onEdit = () => {},
    onDelete = () => {},
    popoverOpen = false,
    onPopoverOpenChange = () => {},
    iconSize = "w-2 h-2",
    triggerSize = "w-3 h-3"
  } = $props();
</script>

<Popover 
  open={popoverOpen}
  onOpenChange={onPopoverOpenChange}
>
  <PopoverTrigger 
    class="inline-flex items-center justify-center {triggerSize} rounded-full hover:bg-black/10 hover:bg-white/20 transition-all duration-200"
    onclick={(e) => e.stopPropagation()}
  >
    <Eye class="{iconSize} text-foreground/50 hover:text-foreground transition-all duration-200" />
  </PopoverTrigger>
  <PopoverContent align="start" class="w-48 p-1">
    <div class="space-y-1">
      <button
        class="w-full flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-secondary rounded transition-colors"
        onclick={() => onEdit(null, highlight)}
      >
        <Edit3 class="w-4 h-4" />
        Edit Times
      </button>
      <button
        class="w-full flex items-center gap-2 px-2 py-1.5 text-sm hover:bg-destructive/10 hover:text-destructive rounded transition-colors"
        onclick={() => onDelete(null, highlight)}
      >
        <Trash2 class="w-4 h-4" />
        Delete Highlight
      </button>
    </div>
  </PopoverContent>
</Popover>