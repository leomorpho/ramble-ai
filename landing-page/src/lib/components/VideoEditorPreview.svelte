<script lang="ts">
	import { onMount } from 'svelte';

	let containerElement: HTMLElement;

	onMount(() => {
		// Simple floating animation for the preview
		if (containerElement) {
			const animate = () => {
				containerElement.style.transform = `translateY(${Math.sin(Date.now() * 0.001) * 10}px)`;
				requestAnimationFrame(animate);
			};
			animate();
		}
	});
</script>

<div bind:this={containerElement} class="relative w-full h-96 flex items-center justify-center">
	<!-- Main Interface Preview -->
	<div class="relative w-full max-w-2xl mx-auto bg-card rounded-lg shadow-2xl border border-border overflow-hidden">
		<!-- App Title Bar -->
		<div class="bg-muted/50 px-4 py-2 border-b border-border flex items-center gap-2">
			<div class="flex gap-1">
				<div class="w-3 h-3 rounded-full bg-red-500"></div>
				<div class="w-3 h-3 rounded-full bg-yellow-500"></div>
				<div class="w-3 h-3 rounded-full bg-green-500"></div>
			</div>
			<span class="text-sm font-medium text-muted-foreground ml-2">VidKing - Video Editor</span>
		</div>

		<!-- Main Interface -->
		<div class="p-4 bg-background">
			<!-- Video Preview Area -->
			<div class="bg-black rounded-lg mb-4 aspect-video flex items-center justify-center">
				<div class="text-white/60 text-center">
					<svg class="w-16 h-16 mx-auto mb-2" fill="currentColor" viewBox="0 0 24 24">
						<path d="M8 5v14l11-7z"/>
					</svg>
					<div class="text-sm">Video Preview</div>
				</div>
			</div>

			<!-- Timeline Area -->
			<div class="space-y-2">
				<div class="text-xs text-muted-foreground mb-1">Timeline</div>
				<!-- Timeline tracks -->
				<div class="space-y-1">
					<!-- Main video track -->
					<div class="h-8 bg-blue-500/20 rounded flex items-center px-2 relative">
						<span class="text-xs text-blue-600 font-medium">Main Video</span>
						<!-- Highlight segments -->
						<div class="absolute left-12 top-1 bottom-1 w-16 bg-green-500 rounded-sm"></div>
						<div class="absolute left-32 top-1 bottom-1 w-12 bg-green-500 rounded-sm"></div>
						<div class="absolute left-48 top-1 bottom-1 w-20 bg-green-500 rounded-sm"></div>
					</div>
					
					<!-- Audio track -->
					<div class="h-6 bg-purple-500/20 rounded flex items-center px-2">
						<span class="text-xs text-purple-600 font-medium">Audio</span>
					</div>
				</div>

				<!-- Timeline controls -->
				<div class="flex items-center gap-2 pt-2">
					<button aria-label="Play video" class="w-6 h-6 bg-primary rounded flex items-center justify-center">
						<svg class="w-3 h-3 text-primary-foreground" fill="currentColor" viewBox="0 0 24 24">
							<path d="M8 5v14l11-7z"/>
						</svg>
					</button>
					<div class="text-xs text-muted-foreground">00:32 / 02:45</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Floating AI Badge -->
	<div class="absolute top-4 right-4 bg-gradient-to-r from-green-500 to-blue-500 text-white px-3 py-1 rounded-full text-xs font-medium shadow-lg">
		AI Powered
	</div>

	<!-- Feature callouts -->
	<div class="absolute -left-8 top-1/2 -translate-y-1/2 hidden lg:block">
		<div class="bg-card border border-border rounded-lg p-3 shadow-lg">
			<div class="text-xs font-medium text-foreground">Precision Editing</div>
			<div class="text-xs text-muted-foreground">Timestamp-based cuts</div>
		</div>
	</div>

	<div class="absolute -right-8 top-1/4 hidden lg:block">
		<div class="bg-card border border-border rounded-lg p-3 shadow-lg">
			<div class="text-xs font-medium text-foreground">AI Highlights</div>
			<div class="text-xs text-muted-foreground">Auto-detected moments</div>
		</div>
	</div>
</div>