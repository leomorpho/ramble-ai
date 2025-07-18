<script>
  import { Edit3, Trash2, Pen, EyeOff } from '@lucide/svelte';
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";

  let { 
    highlight,
    onEdit = () => {},
    onDelete = () => {},
    onHide = () => {},
    popoverOpen = false,
    onPopoverOpenChange = () => {},
    iconSize = "w-4 h-4",
    triggerSize = "w-5 h-5"
  } = $props();
</script>

<DropdownMenu.Root open={popoverOpen} onOpenChange={onPopoverOpenChange}>
  <DropdownMenu.Trigger 
    class="inline-flex items-center justify-center {triggerSize} rounded-full hover:bg-white/20 transition-all duration-200"
    onclick={(e) => e.stopPropagation()}
  >
    <Pen class="{iconSize} text-foreground/50 hover:text-foreground transition-all duration-200" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content class="w-48">
    <DropdownMenu.Item 
      class="flex items-center gap-2"
      onclick={() => onEdit(null, highlight)}
    >
      <Edit3 class="w-4 h-4" />
      Edit Times
    </DropdownMenu.Item>
    <DropdownMenu.Item 
      class="flex items-center gap-2"
      onclick={() => onHide(null, highlight)}
    >
      <EyeOff class="w-4 h-4" />
      Hide from Timeline
    </DropdownMenu.Item>
    <DropdownMenu.Item 
      class="flex items-center gap-2 text-destructive"
      onclick={() => onDelete(null, highlight)}
    >
      <Trash2 class="w-4 h-4" />
      Delete Highlight
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>