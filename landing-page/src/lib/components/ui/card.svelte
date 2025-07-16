<script lang="ts">
	import { cn } from "$lib/utils.js";

	let className: string = "";
	export { className as class };
</script>

<div class={cn("glass rounded-xl border border-slate-700/50 p-8 card-hover", className)}>
	<slot />
</div>