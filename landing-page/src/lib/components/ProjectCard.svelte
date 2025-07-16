<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Project {
		title: string;
		subtitle: string;
		client: string;
		description: string;
		impact: string;
		tech: string[];
		score: string;
		year: string;
		image?: string;
		url?: string | null;
	}

	let { project, index }: { project: Project; index: number } = $props();

	let cardElement: HTMLElement;

	onMount(() => {
		if (!cardElement) {
			console.warn('Card element not found in onMount');
			return;
		}

		// Add a simple hover animation
		const handleMouseEnter = (e: Event) => {
			anime({
				targets: cardElement,
				translateY: -10,
				scale: 1.03,
				duration: 300,
				easing: 'easeOutQuad'
			});
		};

		const handleMouseLeave = (e: Event) => {
			anime({
				targets: cardElement,
				translateY: 0,
				scale: 1,
				duration: 300,
				easing: 'easeOutQuad'
			});
		};

		cardElement.addEventListener('mouseenter', handleMouseEnter);
		cardElement.addEventListener('mouseleave', handleMouseLeave);

		return () => {
			cardElement?.removeEventListener('mouseenter', handleMouseEnter);
			cardElement?.removeEventListener('mouseleave', handleMouseLeave);
		};
	});

</script>

<div bind:this={cardElement} class="project-card">
	<Card class="group hover:shadow-xl transition-shadow duration-300 h-full border-none rounded-md bg-slate-100 dark:bg-slate-900">
	<CardContent class="p-6 h-full">
		<div class="flex flex-col h-full">
			<!-- Header -->
			<div class="flex items-center justify-between mb-4">
				{#if project.url}
					<a 
						href={project.url} 
						target="_blank" 
						rel="noopener noreferrer"
						aria-label="View {project.title} project (opens in new tab)"
						class="w-8 h-8 bg-blue-100 hover:bg-blue-200 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 rounded-lg flex items-center justify-center hover:scale-105 transition-all"
						onclick={(e) => e.stopPropagation()}
					>
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue-600 dark:text-blue-400">
							<path d="M7 17L17 7"/>
							<path d="M7 7h10v10"/>
						</svg>
					</a>
				{:else}
					<div class="w-8 h-8"></div>
				{/if}
				<Badge variant="outline">{project.year}</Badge>
			</div>

			<!-- Project Image -->
			{#if project.image}
				<div class="w-full h-32 rounded-lg overflow-hidden mb-4">
					<img 
						src={project.image} 
						alt="{project.title} screenshot"
						loading="lazy"
						decoding="async"
						class="project-image w-full h-full object-contain transition-transform duration-300"
					/>
				</div>
			{/if}

			<!-- Content - grows to fill available space -->
			<div class="flex-1 space-y-3 mb-4 min-h-0">
				<h3 class="project-title text-xl font-bold text-foreground group-hover:text-primary transition-colors">
					{project.title}
				</h3>
				<p class="text-muted-foreground text-xs uppercase tracking-wider">{project.client}</p>
				<p class="text-muted-foreground text-sm leading-relaxed">{project.description}</p>
			</div>

			<!-- Impact - fixed at bottom -->
			<div class="bg-green-500/80 dark:bg-green-500/70 p-3 rounded-lg mb-4">
				<p class="text-background font-bold text-xs">{project.impact}</p>
			</div>

			<!-- Tech Stack - fixed at bottom -->
			<!-- <div class="flex flex-wrap gap-1">
				{#each project.tech as tech, techIndex}
					<span 
						class="tech-badge bg-secondary/50 text-secondary-foreground px-2 py-1 rounded-md text-xs"
					>
						{tech}
					</span>
				{/each}
			</div> -->
		</div>
	</CardContent>
	</Card>
</div>