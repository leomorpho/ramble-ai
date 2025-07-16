<script lang="ts">
	import { theme, toggleTheme } from '$lib/stores/theme';
</script>

<button
	on:click={toggleTheme}
	class="inline-flex items-center justify-center w-9 h-9 rounded-lg bg-muted hover:bg-muted/80 transition-colors cursor-pointer"
	aria-label="Toggle theme"
>
	{#if $theme === 'light'}
		<!-- Moon icon for dark mode -->
		<svg
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			class="text-muted-foreground"
		>
			<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
		</svg>
	{:else}
		<!-- Sun icon for light mode -->
		<svg
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			class="text-muted-foreground"
		>
			<circle cx="12" cy="12" r="5" />
			<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
		</svg>
	{/if}
</button>