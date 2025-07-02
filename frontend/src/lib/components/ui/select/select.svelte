<script>
	import { createEventDispatcher } from 'svelte';
	import { cn } from "$lib/utils.js";
	import { ChevronDown, Check } from "@lucide/svelte";
	
	let {
		value = $bindable(),
		options = [],
		placeholder = "Select...",
		class: className = "",
		disabled = false,
		...restProps
	} = $props();
	
	let isOpen = $state(false);
	let triggerElement = $state(null);
	
	const dispatch = createEventDispatcher();
	
	function toggleOpen() {
		if (!disabled) {
			isOpen = !isOpen;
		}
	}
	
	function selectOption(option) {
		value = option.value;
		isOpen = false;
		dispatch('change', { value: option.value, label: option.label });
	}
	
	function handleKeydown(event) {
		if (event.key === 'Escape') {
			isOpen = false;
		}
	}
	
	// Close dropdown when clicking outside
	function handleClickOutside(event) {
		if (triggerElement && !triggerElement.contains(event.target)) {
			isOpen = false;
		}
	}
	
	const selectedOption = $derived(
		options.find(option => option.value === value)
	);
	
	$effect(() => {
		if (isOpen) {
			document.addEventListener('click', handleClickOutside);
			document.addEventListener('keydown', handleKeydown);
		} else {
			document.removeEventListener('click', handleClickOutside);
			document.removeEventListener('keydown', handleKeydown);
		}
		
		return () => {
			document.removeEventListener('click', handleClickOutside);
			document.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

<div class="relative" bind:this={triggerElement}>
	<!-- Trigger -->
	<button
		type="button"
		class={cn(
			"flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
			className
		)}
		onclick={toggleOpen}
		{disabled}
		aria-expanded={isOpen}
		aria-haspopup="listbox"
		{...restProps}
	>
		<span class="block truncate">
			{selectedOption?.label || placeholder}
		</span>
		<ChevronDown class="h-4 w-4 opacity-50 transition-transform {isOpen ? 'rotate-180' : ''}" />
	</button>
	
	<!-- Dropdown Content -->
	{#if isOpen}
		<div class="absolute z-50 mt-1 w-full max-h-60 overflow-auto rounded-md border bg-popover text-popover-foreground shadow-md">
			<div class="p-1">
				{#each options as option (option.value)}
					<div
						class="relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-2 pr-8 text-sm outline-none hover:bg-accent hover:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50"
						class:bg-accent={selectedOption?.value === option.value}
						onclick={() => selectOption(option)}
						onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectOption(option); } }}
						role="option"
						aria-selected={selectedOption?.value === option.value}
						tabindex="0"
					>
						<span class="block truncate">{option.label}</span>
						{#if selectedOption?.value === option.value}
							<span class="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
								<Check class="h-4 w-4" />
							</span>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>