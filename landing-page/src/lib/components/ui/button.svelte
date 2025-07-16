<script lang="ts">
	import { tv, type VariantProps } from "tailwind-variants";
	import { cn } from "$lib/utils.js";

	const buttonVariants = tv({
		base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 disabled:pointer-events-none disabled:opacity-50 cursor-pointer",
		variants: {
			variant: {
				default: "gradient-primary text-white shadow hover:shadow-xl hover:shadow-blue-500/25 hover:scale-105",
				secondary: "glass text-slate-300 hover:bg-slate-700/50",
				outline: "border border-slate-600 bg-transparent text-slate-300 hover:bg-slate-800/50",
				ghost: "text-slate-300 hover:bg-slate-800/50"
			},
			size: {
				default: "h-12 px-6 py-3",
				sm: "h-9 px-4 py-2 text-xs",
				lg: "h-14 px-8 py-4 text-base",
				icon: "h-9 w-9"
			}
		},
		defaultVariants: {
			variant: "default",
			size: "default"
		}
	});

	type Variant = VariantProps<typeof buttonVariants>["variant"];
	type Size = VariantProps<typeof buttonVariants>["size"];

	export let variant: Variant = "default";
	export let size: Size = "default";
	export let href: string | undefined = undefined;
	export let disabled: boolean = false;

	let className: string = "";
	export { className as class };
</script>

{#if href}
	<a
		{href}
		class={cn(buttonVariants({ variant, size }), className)}
		role="button"
		tabindex={disabled ? -1 : 0}
		aria-disabled={disabled}
		{...$$restProps}
		on:click
		on:keydown
	>
		<slot />
	</a>
{:else}
	<button
		class={cn(buttonVariants({ variant, size }), className)}
		{disabled}
		{...$$restProps}
		on:click
		on:keydown
	>
		<slot />
	</button>
{/if}