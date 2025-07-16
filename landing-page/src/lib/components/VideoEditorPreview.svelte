<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';

	let containerElement: HTMLElement;
	let playheadPosition = $state(25);
	let isPlaying = $state(false);
	let currentTime = $state('00:32');
	let totalTime = $state('02:45');

	onMount(() => {
		// Floating animation for the container
		if (containerElement) {
			anime({
				targets: containerElement,
				translateY: [-5, 5],
				duration: 4000,
				easing: 'easeInOutSine',
				loop: true,
				direction: 'alternate'
			});
		}

		// Simulate playhead movement
		const playheadInterval = setInterval(() => {
			if (isPlaying) {
				playheadPosition += 0.5;
				if (playheadPosition > 90) {
					playheadPosition = 5;
				}
				// Update time display
				const seconds = Math.floor((playheadPosition / 100) * 165); // 2:45 = 165 seconds
				const mins = Math.floor(seconds / 60);
				const secs = seconds % 60;
				currentTime = `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
			}
		}, 100);

		// Auto-start playing after a delay
		setTimeout(() => {
			isPlaying = true;
		}, 2000);

		return () => {
			clearInterval(playheadInterval);
		};
	});

	const togglePlay = () => {
		isPlaying = !isPlaying;
	};
</script>

<div bind:this={containerElement} class="relative w-full h-96 flex items-center justify-center">
	<!-- Main Interface Preview -->
	<div class="relative w-full max-w-4xl mx-auto video-card video-card-hover glow-effect rounded-2xl overflow-hidden">
		<!-- App Title Bar -->
		<div class="bg-video-timeline px-4 py-3 border-b border-border/20 flex items-center gap-3">
			<div class="flex gap-2">
				<div class="w-3 h-3 rounded-full bg-red-500 hover:bg-red-400 transition-colors cursor-pointer"></div>
				<div class="w-3 h-3 rounded-full bg-yellow-500 hover:bg-yellow-400 transition-colors cursor-pointer"></div>
				<div class="w-3 h-3 rounded-full bg-green-500 hover:bg-green-400 transition-colors cursor-pointer"></div>
			</div>
			<div class="flex items-center gap-2">
				<div class="w-6 h-6 rounded bg-primary/20 flex items-center justify-center">
					<span class="text-primary font-bold text-xs">R</span>
				</div>
				<span class="text-sm font-medium text-white">Ramble - Script Optimization Tool</span>
			</div>
			<div class="ml-auto flex items-center gap-2">
				<div class="px-2 py-1 bg-primary/20 rounded text-xs text-primary font-medium">Script Optimizer</div>
			</div>
		</div>

		<!-- Main Interface -->
		<div class="p-6 bg-background/95 backdrop-blur-sm">
			<!-- Script Optimization Interface -->
			<div class="script-optimizer bg-background border border-border/30 rounded-lg p-4 mb-6">
				<!-- Two-panel layout: Original clips vs Optimized script -->
				<div class="grid grid-cols-2 gap-4 h-64">
					<!-- Left Panel: Original Clips -->
					<div class="space-y-3">
						<div class="flex items-center gap-2 mb-3">
							<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
							<span class="text-sm font-medium text-foreground">Original Clips</span>
						</div>
						
						<!-- Raw clips -->
						<div class="space-y-2 overflow-y-auto max-h-48">
							{#each ['Introduction ramble', 'Key point buried', 'Tangent about coffee', 'Main concept', 'Another tangent', 'Conclusion attempt'] as clip, i}
								<div class="bg-muted/30 rounded p-2 text-xs flex items-center gap-2">
									<div class="w-1.5 h-1.5 rounded-full bg-muted-foreground"></div>
									<span class="text-muted-foreground">{clip}</span>
									<span class="text-xs text-muted-foreground ml-auto">{i + 1}:30</span>
								</div>
							{/each}
						</div>
					</div>
					
					<!-- Right Panel: Optimized Script -->
					<div class="space-y-3">
						<div class="flex items-center gap-2 mb-3">
							<div class="w-3 h-3 rounded-full bg-video-success animate-pulse"></div>
							<span class="text-sm font-medium text-foreground">AI-Optimized Script</span>
							<div class="ml-auto bg-primary/20 text-primary px-2 py-1 rounded text-xs font-semibold">✨ AI</div>
						</div>
						
						<!-- Optimized clips -->
						<div class="space-y-2 overflow-y-auto max-h-48">
							{#each ['Introduction (clean)', 'Main concept', 'Supporting details', 'Strong conclusion'] as clip, i}
								<div class="bg-video-success/20 border border-video-success/30 rounded p-2 text-xs flex items-center gap-2">
									<div class="w-1.5 h-1.5 rounded-full bg-video-success"></div>
									<span class="text-foreground font-medium">{clip}</span>
									<span class="text-xs text-muted-foreground ml-auto">{i + 1}:{(i + 1) * 15}</span>
								</div>
							{/each}
						</div>
					</div>
				</div>
				
				<!-- AI Processing Arrow -->
				<div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-primary/20 backdrop-blur-sm rounded-full p-3 border border-primary/30">
					<svg class="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"></path>
					</svg>
				</div>
			</div>

			<!-- Script Analysis Results -->
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<div class="text-sm font-medium text-foreground">Script Analysis</div>
					<div class="flex items-center gap-2 text-xs text-muted-foreground">
						<span>Processing Time: 45s</span>
						<div class="w-4 h-4 bg-primary/20 rounded flex items-center justify-center">
							<span class="text-primary text-xs">AI</span>
						</div>
					</div>
				</div>
				
				<!-- Progress bar for script optimization -->
				<div class="relative h-6 bg-muted/30 rounded overflow-hidden">
					<div class="absolute inset-0 flex items-center justify-between px-2 text-xs text-muted-foreground z-10">
						<span>Content Quality</span>
						<span>Narrative Flow</span>
						<span>Engagement Score</span>
					</div>
					<!-- Progress fill -->
					<div 
						class="absolute top-0 left-0 h-full bg-gradient-to-r from-video-success to-primary transition-all duration-1000"
						style="width: {playheadPosition * 0.9}%"
					></div>
				</div>
				
				<!-- Script metrics -->
				<div class="grid grid-cols-3 gap-4">
					<!-- Original content stats -->
					<div class="bg-muted/20 rounded-lg p-3">
						<div class="text-xs text-muted-foreground mb-1">Original</div>
						<div class="text-sm font-semibold text-foreground">18:42 mins</div>
						<div class="text-xs text-muted-foreground">6 segments</div>
					</div>
					
					<!-- Optimized content stats -->
					<div class="bg-video-success/20 rounded-lg p-3 border border-video-success/30">
						<div class="text-xs text-video-success mb-1">Optimized</div>
						<div class="text-sm font-semibold text-foreground">4:15 mins</div>
						<div class="text-xs text-video-success">4 key segments</div>
					</div>
					
					<!-- Time saved -->
					<div class="bg-primary/20 rounded-lg p-3 border border-primary/30">
						<div class="text-xs text-primary mb-1">Time Saved</div>
						<div class="text-sm font-semibold text-foreground">77%</div>
						<div class="text-xs text-primary">14:27 mins</div>
					</div>
				</div>

				<!-- Export options -->
				<div class="flex items-center justify-between pt-4 border-t border-border/30">
					<div class="flex items-center gap-3">
						<button 
							class="w-8 h-8 bg-video-success rounded-lg flex items-center justify-center hover:bg-video-success/80 transition-colors"
						>
							<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
							</svg>
						</button>
						<div class="text-sm text-foreground">Script optimized and ready for export</div>
					</div>
					
					<div class="flex items-center gap-2 text-xs text-muted-foreground">
						<div class="flex items-center gap-1">
							<div class="w-2 h-2 bg-video-success rounded-full"></div>
							<span>4 Selected Clips</span>
						</div>
						<div class="flex items-center gap-1">
							<div class="w-2 h-2 bg-primary rounded-full"></div>
							<span>Ready for Handoff</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Floating AI Badge -->
	<div class="absolute -top-3 -right-3 bg-gradient-video text-white px-4 py-2 rounded-full text-sm font-medium shadow-xl pulse-glow">
		✨ Script Optimizer
	</div>

	<!-- Feature callouts -->
	<div class="absolute -left-12 top-1/3 -translate-y-1/2 hidden xl:block z-20">
		<div class="video-card p-4 shadow-xl max-w-48">
			<div class="flex items-center gap-2 mb-2">
				<div class="w-2 h-2 bg-primary rounded-full animate-pulse"></div>
				<div class="text-sm font-semibold text-foreground">Smart Selection</div>
			</div>
			<div class="text-xs text-muted-foreground">Automatically identifies best content from talking head footage</div>
		</div>
	</div>

	<div class="absolute -right-12 top-1/4 hidden xl:block z-20">
		<div class="video-card p-4 shadow-xl max-w-48">
			<div class="flex items-center gap-2 mb-2">
				<div class="w-2 h-2 bg-video-accent rounded-full animate-pulse"></div>
				<div class="text-sm font-semibold text-foreground">Script Reordering</div>
			</div>
			<div class="text-xs text-muted-foreground">AI restructures clips into logical, flowing narratives</div>
		</div>
	</div>

	<div class="absolute -left-8 bottom-1/4 hidden xl:block z-20">
		<div class="video-card p-4 shadow-xl max-w-44">
			<div class="flex items-center gap-2 mb-2">
				<div class="w-2 h-2 bg-video-success rounded-full animate-pulse"></div>
				<div class="text-sm font-semibold text-foreground">Export Ready</div>
			</div>
			<div class="text-xs text-muted-foreground">Optimized scripts ready for any video editor</div>
		</div>
	</div>
</div>

<style>
	.video-timeline {
		background: linear-gradient(90deg, 
			transparent 0%, 
			oklch(from var(--primary) l c h / 0.1) 20%, 
			oklch(from var(--primary) l c h / 0.3) 50%, 
			oklch(from var(--primary) l c h / 0.1) 80%, 
			transparent 100%);
	}
	
	@keyframes float {
		0%, 100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-10px);
		}
	}
	
	.float-animation {
		animation: float 3s ease-in-out infinite;
	}
</style>