<script lang="ts">
	import { tv, type VariantProps } from "tailwind-variants";
	import { cn } from "$lib/utils.js";

	const badgeVariants = tv({
		base: "inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
		variants: {
			variant: {
				default: "bg-slate-700/50 text-slate-300 border-slate-600/50",
				secondary: "bg-slate-800/50 text-slate-400 border-slate-700/50",
				destructive: "bg-red-500/10 text-red-400 border-red-500/20",
				outline: "text-slate-300 border-slate-600"
			}
		},
		defaultVariants: {
			variant: "default"
		}
	});

	type Variant = VariantProps<typeof badgeVariants>["variant"];

	type $$Props = {
		variant?: Variant;
		class?: string;
	};

	export let variant: Variant = "default";

	let className: string = "";
	export { className as class };
</script>

<div class={cn(badgeVariants({ variant }), className)} {...$$restProps}>
	<slot />
</div>