<script>
	import GripVerticalIcon from "@lucide/svelte/icons/grip-vertical";
	import * as ResizablePrimitive from "paneforge";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		withHandle = false,
		...restProps
	} = $props();
</script>

<ResizablePrimitive.PaneResizer
	bind:ref
	data-slot="resizable-handle"
	class={cn(
		"bg-border hover:bg-primary/10 focus-visible:ring-ring focus-visible:outline-hidden relative flex w-px items-center justify-center after:absolute after:inset-y-0 after:left-1/2 after:w-3 after:-translate-x-1/2 focus-visible:ring-1 focus-visible:ring-offset-1 data-[direction=vertical]:h-px data-[direction=vertical]:w-full data-[direction=vertical]:after:left-0 data-[direction=vertical]:after:h-3 data-[direction=vertical]:after:w-full data-[direction=vertical]:after:-translate-y-1/2 data-[direction=vertical]:after:translate-x-0 [&[data-direction=vertical]>div]:rotate-90 cursor-col-resize data-[direction=vertical]:cursor-row-resize transition-all duration-200 group",
		className
	)}
	{...restProps}
>
	{#if withHandle}
		<div class="bg-muted-foreground/30 hover:bg-muted-foreground/60 group-hover:bg-primary/80 rounded-sm z-10 flex h-6 w-4 items-center justify-center border border-border/50 hover:border-primary/50 transition-all duration-200 shadow-sm hover:shadow-md">
			<GripVerticalIcon class="size-3 text-muted-foreground group-hover:text-primary-foreground transition-colors duration-200" />
		</div>
	{/if}
</ResizablePrimitive.PaneResizer>