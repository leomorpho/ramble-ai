<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';
	import { slide } from 'svelte/transition';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import FeatureCard from '$lib/components/FeatureCard.svelte';
	import VideoEditorPreview from '$lib/components/VideoEditorPreview.svelte';
	import Navigation from '$lib/components/Navigation.svelte';

	const features = [
		{
			title: 'AI-Powered Highlights',
			description: 'Automatically detect and extract the most engaging moments from your videos using advanced AI algorithms.',
			benefits: ['Save hours of manual editing', 'Never miss important moments', 'Smart content suggestions'],
			icon: '🤖',
			screenshot: 'ai-highlights-preview.png',
			isNew: true
		},
		{
			title: 'Precision Timeline Editing',
			description: 'Edit with timestamp precision using drag-to-resize controls and frame-perfect accuracy.',
			benefits: ['Frame-perfect cuts', 'Drag-to-resize highlights', 'Non-destructive editing'],
			icon: '✂️',
			screenshot: 'timeline-editing-preview.png'
		},
		{
			title: 'Automatic Transcription',
			description: 'Generate accurate transcripts with word-level timestamps for easy navigation and editing.',
			benefits: ['Word-level timestamps', 'Search within videos', 'Accessibility support'],
			icon: '🎤',
			screenshot: 'transcription-preview.png'
		},
		{
			title: 'Smart Clip Organization',
			description: 'Organize and manage your video clips with intelligent tagging and project management.',
			benefits: ['Auto-tagging', 'Project organization', 'Quick search and filter'],
			icon: '📚',
			screenshot: 'organization-preview.png'
		},
		{
			title: 'Flexible Export Options',
			description: 'Export individual clips or create stitched compilations in multiple formats and resolutions.',
			benefits: ['Multiple formats', 'Custom resolutions', 'Batch export'],
			icon: '📤',
			screenshot: 'export-options-preview.png'
		},
		{
			title: 'Native Performance',
			description: 'Built with Wails for true native desktop performance that handles large video files smoothly.',
			benefits: ['Lightning fast', 'Large file support', 'Memory efficient'],
			icon: '⚡',
		}
	];


	// FAQ data and state
	let openFaqId = $state<string | null>(null);

	const faqs = [
		{
			id: 'pricing',
			question: 'How much does VidKing cost?',
			answer: 'VidKing offers a free trial with full features for 14 days. After that, we have flexible subscription plans starting at $29/month for individual creators, with team and enterprise options available. All plans include free updates and email support.'
		},
		{
			id: 'system-requirements',
			question: 'What are the system requirements?',
			answer: 'VidKing runs on Windows 10+, macOS 10.15+, and modern Linux distributions. You\'ll need at least 4GB RAM (8GB recommended), 2GB of free disk space, and a graphics card that supports hardware acceleration for optimal performance with large video files.'
		},
		{
			id: 'file-formats',
			question: 'What video formats does VidKing support?',
			answer: 'VidKing supports all major video formats including MP4, MOV, AVI, MKV, WebM, and more. We use FFmpeg under the hood, which means if your computer can play it, VidKing can edit it. Export options include MP4, MOV, WebM in various resolutions from 720p to 4K.'
		},
		{
			id: 'ai-accuracy',
			question: 'How accurate is the AI highlight detection?',
			answer: 'Our AI achieves 85-90% accuracy in detecting engaging moments, speaker changes, and key segments. The system learns from user feedback and gets better over time. You always have full control to accept, reject, or modify AI suggestions to match your creative vision.'
		},
		{
			id: 'large-files',
			question: 'Can VidKing handle large video files?',
			answer: 'Yes! VidKing is optimized for large files and long-form content. We support videos up to 10+ hours and 50GB+ file sizes. The app uses smart caching and preview generation to maintain smooth performance even with 4K footage.'
		},
		{
			id: 'collaboration',
			question: 'Can multiple people work on the same project?',
			answer: 'Currently, VidKing is designed for individual use, but team collaboration features are coming soon. You can easily export and share project files, highlights, and exported clips with team members.'
		},
		{
			id: 'support',
			question: 'What kind of support do you provide?',
			answer: 'We provide email support for all users, with video tutorials and documentation available online. Premium subscribers get priority support with faster response times. We also have an active community forum where users share tips and workflows.'
		},
		{
			id: 'data-privacy',
			question: 'What about data privacy and security?',
			answer: 'VidKing processes all videos locally on your computer - nothing is uploaded to our servers unless you explicitly choose to use cloud features. Your content stays private and secure. We follow industry-standard security practices and never access your video content.'
		}
	];

	// FAQ accordion functionality
	const toggleFaq = (faqId: string) => {
		openFaqId = openFaqId === faqId ? null : faqId;
	};


	onMount(() => {
		// Detect mobile for performance optimization
		const isMobile = window.innerWidth < 1024;
		const enableHeroAnimations = !isMobile;

		if (enableHeroAnimations) {
			// Hero animations (desktop only)
			anime
				.timeline({
					easing: 'easeOutExpo',
					duration: 400
				})
				.add({
					targets: '.hero-number',
					scale: [0, 1],
					opacity: [0, 1],
					duration: 600,
					easing: 'easeOutBack'
				})
				.add(
					{
						targets: '.hero-title',
						translateY: [50, 0],
						opacity: [0, 1],
						duration: 700,
						easing: 'easeOutCubic'
					},
					'-=400'
				)
				.add(
					{
						targets: '.hero-subtitle',
						translateY: [30, 0],
						opacity: [0, 1],
						duration: 500
					},
					'-=500'
				)
				.add(
					{
						targets: '.hero-desc',
						opacity: [0, 1],
						duration: 400
					},
					'-=300'
				)
				.add(
					{
						targets: '.hero-buttons',
						translateY: [20, 0],
						opacity: [0, 1],
						duration: 400
					},
					'-=200'
				);
		} else {
			// Mobile: Make hero elements immediately visible without animation
			document.querySelectorAll('.hero-title, .hero-subtitle, .hero-desc, .hero-buttons').forEach((element) => {
				(element as HTMLElement).style.opacity = '1';
				(element as HTMLElement).style.transform = 'translateY(0px) scale(1)';
			});
		}

		// Animate pink blob floating movement (desktop only)
		if (enableHeroAnimations) {
			anime({
				targets: '.pink-blob',
				translateX: function () {
					return anime.random(-200, 100);
				},
				translateY: function () {
					return anime.random(-150, 150);
				},
				duration: function () {
					return anime.random(8000, 12000);
				},
				easing: 'easeInOutSine',
				delay: 1000,
				loop: true
			});
		}

		// Use the same mobile detection for scroll animations
		const enableScrollAnimations = !isMobile;

		if (enableScrollAnimations) {
			// Scroll animations (desktop only)
			const observerOptions = {
				root: null,
				rootMargin: '0px 0px -15% 0px',
				threshold: 0.1
			};

			const observer = new IntersectionObserver((entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						anime({
							targets: entry.target,
							translateY: 0,
							opacity: 1,
							duration: 800,
							easing: 'easeOutQuad'
						});
						observer.unobserve(entry.target);
					}
				});
			}, observerOptions);

			document.querySelectorAll('.fade-up').forEach((element) => {
				// Check if this is an FAQ item
				if (element.closest('#faq')) {
					// For FAQ items, use a special observer that animates them all at once
					element.classList.add('faq-item');
				} else {
					observer.observe(element);
				}
			});

			// Special observer for FAQ items - animates all at once
			const faqObserver = new IntersectionObserver((entries) => {
				const faqItems = document.querySelectorAll('.faq-item');
				if (entries[0].isIntersecting && faqItems.length > 0) {
					anime({
						targets: faqItems,
						translateY: 0,
						opacity: 1,
						duration: 600,
						delay: anime.stagger(50), // Minimal 50ms stagger
						easing: 'easeOutQuad'
					});
					entries.forEach(entry => faqObserver.unobserve(entry.target));
				}
			}, observerOptions);

			document.querySelectorAll('.faq-item').forEach((element) => {
				faqObserver.observe(element);
			});

			// Feature cards stagger (desktop only)
			const cardObserver = new IntersectionObserver(
				(entries) => {
					entries.forEach((entry) => {
						if (entry.isIntersecting) {
							const cards = document.querySelectorAll('.feature-card');
							const cardIndex = Array.from(cards).indexOf(entry.target);

							anime({
								targets: entry.target,
								translateY: 0,
								opacity: 1,
								duration: 1000,
								delay: cardIndex * 200,
								easing: 'easeOutQuad'
							});
							cardObserver.unobserve(entry.target);
						}
					});
				},
				{
					root: null,
					rootMargin: '0px 0px -20% 0px',
					threshold: 0.1
				}
			);

			document.querySelectorAll('.feature-card').forEach((card) => {
				cardObserver.observe(card);
			});
		} else {
			// Mobile: Make all elements immediately visible without animation
			document.querySelectorAll('.fade-up').forEach((element) => {
				(element as HTMLElement).style.opacity = '1';
				(element as HTMLElement).style.transform = 'translateY(0px)';
			});

			document.querySelectorAll('.feature-card').forEach((card) => {
				(card as HTMLElement).style.opacity = '1';
				(card as HTMLElement).style.transform = 'translateY(0px)';
			});
		}

	});
</script>

<svelte:head>
	<title>VidKing • Professional Video Editing Made Simple</title>
	<meta
		name="description"
		content="Powerful desktop video editor with AI-powered highlights, precision timestamp editing, and seamless export options. Transform your video workflow."
	/>
</svelte:head>

<Navigation />

<!-- Hero Section -->
<section class="relative min-h-screen flex items-center overflow-hidden bg-background">
	<!-- Background Effects -->
	<div class="absolute inset-0 z-0">
		<div
			class="absolute top-20 left-20 w-96 h-96 gradient-green opacity-20 rounded-full blur-3xl"
		></div>
		<div
			class="pink-blob absolute bottom-20 right-20 w-80 h-80 gradient-pink opacity-20 rounded-full blur-3xl"
		></div>
		<div
			class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] gradient-cyan opacity-10 rounded-full blur-3xl"
		></div>
	</div>

	<div class="relative z-10 max-w-7xl mx-auto px-6 grid lg:grid-cols-12 gap-12 items-center">
		<!-- Left Side -->
		<div class="lg:col-span-7 space-y-8">
			<div class="space-y-8">
				<h1 class="hero-title text-7xl lg:text-8xl xl:text-9xl text-display leading-none">
					<span class="block text-foreground">PROFESSIONAL</span>
					<span class="block gradient-text font-black">VIDEO</span>
					<span class="block text-foreground">EDITING</span>
				</h1>

				<div class="space-y-8 max-w-3xl">
					<p
						class="hero-subtitle text-2xl lg:text-3xl text-foreground/90 font-medium leading-relaxed"
					>
						AI-powered highlights, precision editing, and seamless export in one powerful desktop app.
					</p>
				</div>
			</div>

			<div class="hero-desc"></div>

			<div class="hero-buttons flex flex-wrap gap-6 pt-4">
				<Button
					size="lg"
					class="shadow-lg shadow-green-500/25 hover:shadow-green-500/40 transition-all duration-300"
					onclick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}
				>
					Download Free Trial
					<svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M7 10l5 5 5-5"
						></path>
					</svg>
				</Button>
				<Button
					variant="outline"
					size="lg"
					onclick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}
				>
					See Features in Action
				</Button>
			</div>
		</div>

		<!-- Right Side - Video Editor Preview -->
		<div class="hidden lg:block lg:col-span-5 relative">
			<div class="relative w-full h-96 overflow-visible">
				<!-- Video Editor Interface Preview -->
				<VideoEditorPreview />
			</div>
		</div>
	</div>
</section>

<!-- Features Section -->
<section id="features" class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				POWERFUL<br />
				<span class="gradient-text">FEATURES</span>
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				Everything you need to create professional videos with AI-powered assistance and precision editing tools.
			</p>
		</div>

		<div class="text-center">
			<p class="text-2xl lg:text-3xl text-foreground leading-relaxed max-w-5xl mx-auto">
				<span class="font-semibold">AI-Powered Highlights</span> • 
				<span class="font-semibold">Precision Timeline Editing</span> • 
				<span class="font-semibold">Automatic Transcription</span> • 
				<span class="font-semibold">Smart Clip Organization</span> • 
				<span class="font-semibold">Export Flexibility</span> • 
				<span class="font-semibold">Native Desktop Performance</span> • 
				<span class="font-semibold">Dark/Light Themes</span> • 
				<span class="font-semibold">Drag-to-Resize Editing</span>
			</p>
		</div>
	</div>
</section>

<!-- Target Users Section -->
<section class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				PERFECT FOR<br />
				<span class="gradient-text">CREATORS</span>
			</h2>
			<p class="text-muted-foreground text-lg max-w-3xl mx-auto">
				Whether you're creating content for social media, education, or business, VidKing adapts to your workflow and makes professional video editing accessible to everyone.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
			<!-- Content Creators -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-green-500 to-blue-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Content Creators</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					YouTubers, streamers, podcasters, and social media influencers who need to quickly create engaging highlights and clips from longer content.
				</p>
			</div>

			<!-- Marketing Teams -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Marketing Teams</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Digital marketers and agencies creating video content for campaigns, social media, and brand storytelling with tight deadlines.
				</p>
			</div>

			<!-- Educators -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Educators</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Teachers, trainers, and course creators who need to extract key moments from lectures and create digestible learning materials.
				</p>
			</div>

			<!-- Business Teams -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-pink-500 to-green-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Business Teams</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Corporate teams creating internal communications, training videos, and meeting highlights for better knowledge sharing.
				</p>
			</div>
		</div>
	</div>
</section>

<!-- Feature Gallery Section -->
<section id="feature-gallery" class="py-20 bg-background">
	<div class="max-w-7xl mx-auto px-6">
		<div class="grid lg:grid-cols-12 gap-12 mb-20">
			<div class="lg:col-span-4">
				<div class="fade-up space-y-4">
					<h2 class="text-4xl lg:text-5xl text-headline text-foreground">
						POWERFUL<br />
						<span class="gradient-text">FEATURES</span>
					</h2>
				</div>
			</div>
			<div class="lg:col-span-8">
				<p class="fade-up text-muted-foreground text-lg leading-relaxed">
					Every feature is designed to save you time and help you create better videos. From AI-powered
					highlight detection to precision editing controls, VidKing handles the technical complexity
					so you can focus on your creativity.
				</p>
			</div>
		</div>

		<div
			class="grid md:grid-cols-2 lg:grid-cols-3 gap-6"
			style="grid-auto-rows: 1fr;"
		>
			{#each features as feature, i (feature.title)}
				<FeatureCard {feature} index={i} />
			{/each}
		</div>
	</div>
</section>


<!-- Workflow Section -->
<section id="workflow" class="py-20 bg-background">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				SIMPLE<br />
				<span class="gradient-text">WORKFLOW</span>
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				From raw footage to polished highlights in just a few clicks. VidKing streamlines your
				entire video editing process.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
			<!-- Step 1: Import -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-green-500 to-blue-500 flex items-center justify-center text-white text-sm font-bold">1</div>
							<h3 class="text-lg font-bold text-foreground">Import</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							Drag and drop your video files. VidKing automatically processes and prepares them for editing.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Step 2: AI Analysis -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center text-white text-sm font-bold">2</div>
							<h3 class="text-lg font-bold text-foreground">AI Analysis</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							AI automatically transcribes your video and suggests the most engaging highlights and moments.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Step 3: Edit -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center text-white text-sm font-bold">3</div>
							<h3 class="text-lg font-bold text-foreground">Edit</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							Refine your highlights with precision controls. Drag to resize, trim, and perfect your clips.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Step 4: Export -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-pink-500 to-green-500 flex items-center justify-center text-white text-sm font-bold">4</div>
							<h3 class="text-lg font-bold text-foreground">Export</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							Export individual clips or create stitched compilations in your preferred format and resolution.
						</p>
					</div>
				</CardContent>
			</Card>
		</div>
	</div>
</section>

<!-- FAQ Section -->
<section id="faq" class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-4xl mx-auto px-6">
		<div class="fade-up text-center mb-16">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				QUESTIONS YOU<br />
				<span class="gradient-text">MIGHT HAVE</span>
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				We get these questions a lot. Here are our honest answers.
			</p>
		</div>

		<div class="space-y-4">
			{#each faqs as faq, index (faq.id)}
				<div class="fade-up border border-border rounded-lg overflow-hidden">
					<button 
						class="w-full px-6 py-4 text-left bg-background hover:bg-muted/50 transition-colors duration-200 flex items-center justify-between"
						onclick={() => toggleFaq(faq.id)}
					>
						<h3 class="text-lg font-semibold text-foreground pr-4">{faq.question}</h3>
						<svg 
							class="w-5 h-5 text-muted-foreground transition-transform duration-200 {openFaqId === faq.id ? 'rotate-180' : ''}"
							fill="none" 
							stroke="currentColor" 
							viewBox="0 0 24 24"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
						</svg>
					</button>
					{#if openFaqId === faq.id}
						<div class="px-6 pb-4 bg-muted/20" transition:slide={{ duration: 300 }}>
							<p class="text-muted-foreground leading-relaxed">{faq.answer}</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Call to Action -->
		<div class="text-center mt-16">
			<Button size="lg" class="shadow-lg shadow-green-500/25 hover:shadow-green-500/40 transition-all duration-300">
				Start Free Trial
			</Button>
		</div>
	</div>
</section>

<!-- Footer -->
<footer class="bg-background border-t border-border py-12">
	<div class="max-w-7xl mx-auto px-6">
		<div class="grid md:grid-cols-3 gap-8 items-center">
			<!-- Logo/Brand -->
			<div class="flex items-center space-x-3">
				<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
					<span class="text-primary-foreground font-bold text-sm">VK</span>
				</div>
				<div>
					<span class="font-bold tracking-tight">
						<span class="text-foreground">VID</span><span class="gradient-text">KING</span>
					</span>
					<p class="text-muted-foreground text-sm">Professional Video Editing</p>
				</div>
			</div>
			
			<!-- Contact -->
			<div class="text-center">
				<h4 class="font-semibold text-foreground mb-2">Support</h4>
				<a 
					href="mailto:support@vidking.app" 
					class="text-primary hover:text-primary/80 transition-colors"
				>
					support@vidking.app
				</a>
			</div>
			
			<!-- Copyright -->
			<div class="text-right text-sm text-muted-foreground">
				<p>&copy; {new Date().getFullYear()} VidKing</p>
				<p>Made for creators</p>
			</div>
		</div>
	</div>
</footer>

<style>
	/* Hide animated elements initially to prevent flash */
	.fade-up {
		opacity: 0;
		transform: translateY(60px);
	}

	:global(.feature-card) {
		opacity: 0;
		transform: translateY(100px);
	}

	/* Hide hero elements initially to prevent flash */
	.hero-title,
	.hero-subtitle,
	.hero-desc,
	.hero-buttons {
		opacity: 0;
		transform: translateY(50px);
	}
</style>
