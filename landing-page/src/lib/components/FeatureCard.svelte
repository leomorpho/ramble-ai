<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	interface Feature {
		title: string;
		description: string;
		benefits: string[];
		icon: string;
		screenshot?: string;
		isNew?: boolean;
	}

	let { feature, index }: { feature: Feature; index: number } = $props();

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

<div bind:this={cardElement} class="feature-card">
	<Card class="group hover:shadow-xl transition-shadow duration-300 h-full border-none rounded-md bg-slate-100 dark:bg-slate-900">
		<CardContent class="p-6 h-full">
			<div class="flex flex-col h-full">
				<!-- Header -->
				<div class="flex items-center justify-between mb-4">
					<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center">
						<span class="text-2xl">{feature.icon}</span>
					</div>
					{#if feature.isNew}
						<Badge variant="outline" class="bg-green-500/10 text-green-600 border-green-500/20">New</Badge>
					{/if}
				</div>

				<!-- Screenshot Preview -->
				{#if feature.screenshot}
					<div class="w-full h-32 rounded-lg overflow-hidden mb-4 bg-muted/50 flex items-center justify-center">
						<div class="text-muted-foreground text-sm">Screenshot Preview</div>
					</div>
				{/if}

				<!-- Content - grows to fill available space -->
				<div class="flex-1 space-y-3 mb-4 min-h-0">
					<h3 class="feature-title text-xl font-bold text-foreground group-hover:text-primary transition-colors">
						{feature.title}
					</h3>
					<p class="text-muted-foreground text-sm leading-relaxed">{feature.description}</p>
				</div>

				<!-- Benefits List -->
				<div class="space-y-2">
					{#each feature.benefits as benefit, benefitIndex}
						<div class="flex items-center gap-2">
							<svg class="w-4 h-4 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
							</svg>
							<span class="text-xs text-muted-foreground">{benefit}</span>
						</div>
					{/each}
				</div>
			</div>
		</CardContent>
	</Card>
</div>

<style>
	.feature-card {
		opacity: 0;
		transform: translateY(60px);
	}
</style>