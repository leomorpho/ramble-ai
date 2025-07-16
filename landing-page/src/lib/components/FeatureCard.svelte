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
	<div class="video-card video-card-hover group h-full rounded-2xl overflow-hidden border border-border/50">
		<div class="p-6 h-full flex flex-col">
			<!-- Header -->
			<div class="flex items-center justify-between mb-6">
				<div class="feature-icon">
					<span class="text-2xl">{feature.icon}</span>
				</div>
				{#if feature.isNew}
					<div class="bg-gradient-accent text-white px-2 py-1 rounded-full text-xs font-semibold">✨ New</div>
				{/if}
			</div>

			<!-- Screenshot Preview -->
			{#if feature.screenshot}
				<div class="w-full h-40 rounded-xl overflow-hidden mb-6 screenshot-shimmer bg-gradient-to-br from-video-timeline to-background border border-border/30">
					<div class="w-full h-full flex items-center justify-center relative">
						<!-- Simulated interface elements -->
						<div class="absolute inset-2 border border-white/10 rounded-lg">
							<div class="absolute top-2 left-2 right-2 h-4 bg-white/5 rounded flex items-center gap-1 px-2">
								<div class="w-1.5 h-1.5 bg-primary rounded-full"></div>
								<div class="w-1.5 h-1.5 bg-video-accent rounded-full"></div>
								<div class="w-1.5 h-1.5 bg-video-success rounded-full"></div>
							</div>
							<div class="absolute bottom-2 left-2 right-2 h-6 bg-white/5 rounded flex items-center px-2">
								<div class="flex gap-0.5 items-end">
									{#each Array(20) as _, i}
										<div class="w-0.5 bg-primary/60 rounded-full" style="height: {Math.random() * 12 + 2}px"></div>
									{/each}
								</div>
							</div>
						</div>
						<div class="text-white/40 text-xs font-medium">{feature.screenshot}</div>
					</div>
				</div>
			{/if}

			<!-- Content - grows to fill available space -->
			<div class="flex-1 space-y-4 mb-6 min-h-0">
				<h3 class="text-xl font-bold text-foreground group-hover:text-primary transition-colors duration-300">
					{feature.title}
				</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">{feature.description}</p>
			</div>

			<!-- Benefits List -->
			<div class="space-y-3">
				{#each feature.benefits as benefit, benefitIndex}
					<div class="flex items-center gap-3 group/benefit">
						<div class="w-5 h-5 rounded-full bg-video-success/20 flex items-center justify-center flex-shrink-0 group-hover/benefit:bg-video-success/30 transition-colors">
							<svg class="w-3 h-3 text-video-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path>
							</svg>
						</div>
						<span class="text-sm text-foreground font-medium">{benefit}</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style>
	.feature-card {
		opacity: 0;
		transform: translateY(60px);
		transition: all 0.6s cubic-bezier(0.16, 1, 0.3, 1);
	}
	
	.feature-card:hover {
		transform: translateY(-8px) scale(1.02);
	}
	
	/* Enhanced shimmer effect for screenshots */
	.screenshot-shimmer::before {
		animation-delay: calc(var(--index, 0) * 0.5s);
	}
</style>