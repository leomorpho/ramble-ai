<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { theme } from '$lib/stores/theme';
	import { browser } from '$app/environment';

	let { children } = $props();

	// Initialize theme on mount
	onMount(() => {
		if (browser) {
			const stored = localStorage.getItem('theme') as 'light' | 'dark';
			const systemPreference = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
			const initialTheme = stored || systemPreference;
			
			theme.set(initialTheme);
			document.documentElement.classList.toggle('dark', initialTheme === 'dark');
		}
	});
</script>

{@render children()}
