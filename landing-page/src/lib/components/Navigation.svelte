<script lang="ts">
	import { onMount } from 'svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import ProjectIntakeForm from './ProjectIntakeForm.svelte';

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
	class="fixed top-0 left-0 right-0 z-50 bg-background/90 backdrop-blur-xl border-b border-border"
>
	<div class="max-w-7xl mx-auto px-6">
		<div class="flex items-center justify-between h-16">
			<!-- Logo -->
			<a
				href="/"
				class="flex items-center space-x-3 hover:opacity-80 transition-opacity cursor-pointer"
			>
				<div class="w-8 h-8 rounded-full overflow-hidden ring-2 ring-primary/20">
					<img 
						src="/logo-128.png" 
						alt="GooseBytes Logo" 
						class="w-full h-full object-cover"
					/>
				</div>
				<span class="font-bold tracking-tight">
					<span class="text-foreground">GOOSE</span><span class="gradient-text">BYTES</span>
				</span>
			</a>

			<!-- Desktop Menu -->
			<div class="hidden md:flex items-center space-x-8">
				<a
					href="#services"
					class="text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors"
					>Services</a
				>
				<a
					href="#work"
					class="text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors"
					>Work</a
				>
				<a
					href="#team"
					class="text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors"
					>Team</a
				>
				<a
					href="#process"
					class="text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors"
					>Process</a
				>
				<a
					href="#faq"
					class="text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors"
					>FAQ</a
				>
				<ThemeToggle />
				<ProjectIntakeForm triggerText="Start Project" size="sm" />
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
					href="#services"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors py-2"
					>Services</a
				>
				<a
					href="#work"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors py-2"
					>Work</a
				>
				<a
					href="#team"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors py-2"
					>Team</a
				>
				<a
					href="#process"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors py-2"
					>Process</a
				>
				<a
					href="#faq"
					onclick={closeMenu}
					class="block text-muted-foreground hover:text-foreground text-sm uppercase tracking-wider transition-colors py-2"
					>FAQ</a
				>
				<div class="pt-4 border-t border-border">
					<ProjectIntakeForm triggerText="Start Project" size="sm" />
				</div>
			</div>
		</div>
	</div>
</nav>