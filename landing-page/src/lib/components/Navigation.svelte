<script lang="ts">
	import { onMount } from 'svelte';
	import ThemeToggle from './ThemeToggle.svelte';

	let isMenuOpen = $state(false);

	const toggleMenu = () => {
		isMenuOpen = !isMenuOpen;
	};

	const closeMenu = () => {
		isMenuOpen = false;
	};

	// Close menu when clicking outside or on links
	onMount(() => {
		const handleClickOutside = (event: MouseEvent) => {
			const nav = document.querySelector('nav');
			if (nav && !nav.contains(event.target as Node)) {
				closeMenu();
			}
		};

		document.addEventListener('click', handleClickOutside);
		return () => document.removeEventListener('click', handleClickOutside);
	});
</script>

<!-- Navigation -->
<nav
	class="fixed top-0 left-0 right-0 z-50 bg-background/95 backdrop-blur-xl border-b border-border/50 shadow-lg"
>
	<div class="max-w-7xl mx-auto px-6">
		<div class="flex items-center justify-between h-16">
			<!-- Logo -->
			<a
				href="/"
				class="flex items-center space-x-3 hover:opacity-80 transition-opacity cursor-pointer"
			>
				<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
					<span class="text-primary-foreground font-bold text-sm">R</span>
				</div>
				<span class="font-bold tracking-tight text-xl">
					<span class="gradient-text">RAMBLE</span>
				</span>
			</a>

			<!-- Desktop Menu -->
			<div class="hidden md:flex items-center space-x-8">
				<a
					href="#features"
					class="text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-all duration-300 hover:scale-105"
					>Features</a
				>
				<a
					href="#feature-gallery"
					class="text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-all duration-300 hover:scale-105"
					>Gallery</a
				>
				<a
					href="#workflow"
					class="text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-all duration-300 hover:scale-105"
					>Workflow</a
				>
				<a
					href="#faq"
					class="text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-all duration-300 hover:scale-105"
					>FAQ</a
				>
				<ThemeToggle />
				<button class="bg-primary text-white dark:text-background px-6 py-2.5 rounded-xl text-sm font-semibold hover:scale-105 hover:shadow-lg hover:bg-primary/90 transition-all duration-300 shadow-lg shadow-primary/30">
					<svg class="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
					</svg>
					Optimize Scripts
				</button>
			</div>

			<!-- Mobile Menu Controls -->
			<div class="md:hidden flex items-center space-x-4">
				<!-- Theme Toggle (always visible on mobile) -->
				<ThemeToggle />
				
				<!-- Hamburger Button -->
				<button
					onclick={toggleMenu}
					class="relative w-6 h-6 flex flex-col justify-center items-center space-y-1 focus:outline-none focus:ring-0 outline-none border-none rounded"
					aria-label="Toggle menu"
				>
					<span 
						class="w-6 h-0.5 bg-foreground transition-all duration-300 {isMenuOpen ? 'rotate-45 translate-y-1.5' : ''}"
					></span>
					<span 
						class="w-6 h-0.5 bg-foreground transition-all duration-300 {isMenuOpen ? 'opacity-0' : 'opacity-100'}"
					></span>
					<span 
						class="w-6 h-0.5 bg-foreground transition-all duration-300 {isMenuOpen ? '-rotate-45 -translate-y-1.5' : ''}"
					></span>
				</button>
			</div>
		</div>

		<!-- Mobile Menu Dropdown -->
		<div 
			class="md:hidden border-t border-border bg-background/95 backdrop-blur-xl overflow-hidden transition-all duration-300 ease-out {isMenuOpen ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'}"
		>
			<div class="px-6 py-4 space-y-4 transform transition-transform duration-300 ease-out {isMenuOpen ? 'translate-y-0' : '-translate-y-4'}">
				<a
					href="#features"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-colors py-3"
					>Features</a
				>
				<a
					href="#feature-gallery"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-colors py-3"
					>Gallery</a
				>
				<a
					href="#workflow"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-colors py-3"
					>Workflow</a
				>
				<a
					href="#faq"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-primary text-sm font-semibold tracking-wide transition-colors py-3"
					>FAQ</a
				>
				<div class="pt-4 border-t border-border">
					<button class="w-full bg-primary text-white dark:text-background px-4 py-3 rounded-xl text-sm font-semibold hover:scale-105 hover:bg-primary/90 transition-all duration-300 shadow-lg shadow-primary/30">
						<svg class="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
						</svg>
						Start Optimizing
					</button>
				</div>
			</div>
		</div>
	</div>
</nav>