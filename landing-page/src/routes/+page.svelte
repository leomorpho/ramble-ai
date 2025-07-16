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
			title: 'Smart Clip Selection',
			description: 'Quickly identify and select the best parts of your talking head videos with our intelligent analysis system.',
			benefits: ['Instant clip identification', 'Content quality scoring', 'Speaker change detection'],
			icon: '🎯',
			screenshot: 'clip-selection-preview.png',
			isNew: true
		},
		{
			title: 'AI Script Reordering',
			description: 'Transform selected clips into coherent, high-quality scripts using advanced AI that understands narrative flow.',
			benefits: ['Optimal content flow', 'Narrative coherence', 'Professional structure'],
			icon: '🧠',
			screenshot: 'script-reordering-preview.png',
			isNew: true
		},
		{
			title: 'Speech-Optimized Transcription',
			description: 'Generate word-perfect transcripts specifically tuned for talking head content with speaker timestamps.',
			benefits: ['Word-level timing', 'Speaker identification', 'Content searchability'],
			icon: '📝',
			screenshot: 'transcription-preview.png'
		},
		{
			title: 'Preprocessing Pipeline',
			description: 'Seamlessly integrate into your existing video production workflow as the essential first step.',
			benefits: ['Workflow integration', 'File format flexibility', 'Fast processing'],
			icon: '🔄',
			screenshot: 'pipeline-preview.png'
		},
		{
			title: 'Script Export & Handoff',
			description: 'Export optimized scripts and clip sequences ready for your video editor of choice.',
			benefits: ['Multiple export formats', 'Editor compatibility', 'Metadata preservation'],
			icon: '📤',
			screenshot: 'export-preview.png'
		},
		{
			title: 'Production Speed Boost',
			description: 'Reduce post-production time by 60-80% with pre-optimized content structure and flow.',
			benefits: ['Massive time savings', 'Quality improvement', 'Stress reduction'],
			icon: '⚡',
		}
	];


	// FAQ data and state
	let openFaqId = $state<string | null>(null);

	const faqs = [
		{
			id: 'preprocessing-vs-editing',
			question: 'How is Ramble different from traditional video editors?',
			answer: 'Ramble is a preprocessing tool, not a video editor. We focus on the step BEFORE editing - optimizing your talking head content into perfect scripts. You then take our optimized output to your favorite editor (Premiere, Final Cut, DaVinci) for final production. Think of us as the essential first step that saves you 60-80% of your editing time.'
		},
		{
			id: 'workflow-integration',
			question: 'How does Ramble integrate with my existing workflow?',
			answer: 'Ramble slots perfectly into your workflow as step zero. Upload your raw talking head footage to Ramble first, let us optimize the script and clip selection, then export the results to your preferred video editor. We support all major editing platforms and maintain full compatibility with your existing tools.'
		},
		{
			id: 'output-format',
			question: 'What do I get when I export from Ramble?',
			answer: 'You get a structured script file with optimized clip sequences, precise timestamps, and metadata that imports seamlessly into any video editor. Plus individual video clips ready for assembly, transcripts with speaker notes, and a detailed content map showing the logical flow of your optimized script.'
		},
		{
			id: 'ai-script-quality',
			question: 'How good is the AI at reordering my content?',
			answer: 'Our AI specializes in talking head content and understands narrative flow, logical progression, and audience engagement. It achieves 90%+ accuracy in creating coherent scripts from disorganized footage. You can always review and adjust the AI suggestions, but most users find the output ready for immediate use.'
		},
		{
			id: 'time-savings',
			question: 'How much time will Ramble actually save me?',
			answer: 'Most users report 60-80% reduction in post-production time. Instead of spending hours manually scrubbing through footage and figuring out the best order, Ramble does this automatically. A 2-hour raw recording that used to take 8 hours to edit now takes 2-3 hours total with Ramble preprocessing.'
		},
		{
			id: 'content-types',
			question: 'What types of talking head videos work best?',
			answer: 'Ramble excels with any spoken content: educational videos, interviews, presentations, course recordings, vlogs, podcasts with video, and business communications. The more speech-heavy your content, the better our AI performs at optimizing the script flow.'
		},
		{
			id: 'pricing',
			question: 'How much does Ramble cost?',
			answer: 'Ramble offers a free trial to test the preprocessing workflow with your content. Our plans start at $29/month for individual creators, with team and enterprise options available. Given the massive time savings, most users see ROI within their first project.'
		},
		{
			id: 'data-privacy',
			question: 'Is my content secure?',
			answer: 'Absolutely. Ramble processes all videos locally on your computer - nothing is uploaded to our servers. Your raw footage and optimized scripts stay completely private and secure on your machine. We never access, store, or analyze your content.'
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
	<title>Ramble • AI-Powered Script Optimization for Talking Head Videos</title>
	<meta
		name="description"
		content="Transform rambling talking head videos into compelling scripts. AI-powered clip selection and reordering for better content and faster post-production."
	/>
</svelte:head>

<Navigation />

<!-- Hero Section -->
<section class="relative min-h-screen flex items-center overflow-hidden bg-background">
	<!-- Background Effects -->
	<div class="absolute inset-0 z-0">
		<!-- Video Editor themed background blobs -->
		<div class="absolute top-20 left-20 w-96 h-96 bg-gradient-video opacity-15 rounded-full blur-3xl pulse-glow"></div>
		<div class="pink-blob absolute bottom-20 right-20 w-80 h-80 bg-gradient-accent opacity-20 rounded-full blur-3xl"></div>
		<div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-timeline opacity-10 rounded-full blur-3xl"></div>
		
		<!-- Grid pattern overlay -->
		<div class="absolute inset-0 opacity-[0.02]" style="background-image: radial-gradient(circle at 1px 1px, rgb(255,255,255) 1px, transparent 0); background-size: 20px 20px;"></div>
	</div>

	<div class="relative z-10 max-w-7xl mx-auto px-6 grid lg:grid-cols-12 gap-16 items-center">
		<!-- Left Side -->
		<div class="lg:col-span-7 space-y-12">
			<div class="space-y-8">
				<!-- Script Optimizer Badge -->
				<div class="hero-title inline-flex items-center gap-3 bg-gradient-video/10 border border-primary/20 rounded-full px-6 py-3 backdrop-blur-sm">
					<div class="w-2 h-2 bg-primary rounded-full animate-pulse"></div>
					<span class="text-sm font-semibold text-primary">For Talking Head Videos</span>
				</div>
				
				<h1 class="hero-title text-6xl lg:text-7xl xl:text-8xl text-display leading-none space-y-2">
					<div class="block text-foreground">TURN <span class="gradient-text font-black">RAMBLING</span></div>
					<div class="block text-foreground">INTO <span class="gradient-text font-black">COMPELLING</span></div>
				</h1>

				<div class="space-y-6 max-w-2xl">
					<p class="hero-subtitle text-xl lg:text-2xl text-muted-foreground font-medium leading-relaxed">
						AI that finds your best clips and reorders them into perfect scripts. Cut post-production time by 80%.
					</p>
					
				</div>
			</div>

			<div class="hero-buttons flex flex-col sm:flex-row gap-4">
				<Button
					size="lg"
					class="bg-gradient-video hover:opacity-90 shadow-2xl shadow-primary/25 transition-all duration-300 text-white font-semibold px-8 py-4 text-lg"
					onclick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}
				>
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
					</svg>
					Start Optimizing Scripts
				</Button>
				<Button
					variant="outline"
					size="lg"
					class="border-primary/30 hover:bg-primary/5 px-8 py-4 text-lg font-semibold"
					onclick={() => document.getElementById('feature-gallery')?.scrollIntoView({ behavior: 'smooth' })}
				>
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
					</svg>
					See How It Works
				</Button>
			</div>
		</div>

		<!-- Right Side - Video Editor Preview -->
		<div class="hidden lg:block lg:col-span-5 relative">
			<div class="relative w-full overflow-visible">
				<!-- Video Editor Interface Preview -->
				<VideoEditorPreview />
			</div>
		</div>
	</div>
</section>

<!-- Features Section -->
<section id="features" class="py-20 bg-gradient-to-b from-background/0 via-muted/10 to-background/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				POWERFUL<br />
				<span class="gradient-text">FEATURES</span>
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				Everything you need to transform talking head videos into perfect scripts before post-production.
			</p>
		</div>

		<div class="text-center">
			<p class="text-2xl lg:text-3xl text-foreground leading-relaxed max-w-5xl mx-auto">
				<span class="font-semibold">Smart Clip Selection</span> • 
				<span class="font-semibold">AI Script Reordering</span> • 
				<span class="font-semibold">Speech-Optimized Transcription</span> • 
				<span class="font-semibold">Preprocessing Pipeline</span> • 
				<span class="font-semibold">Script Export & Handoff</span> • 
				<span class="font-semibold">Production Speed Boost</span> • 
				<span class="font-semibold">Content Quality Optimization</span> • 
				<span class="font-semibold">Workflow Integration</span>
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
				Whether you're creating talking head content for social media, education, or business, Ramble optimizes your scripts before editing and dramatically reduces post-production time.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
			<!-- Content Creators -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-green-500 to-blue-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Talking Head Creators</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					YouTubers, course creators, and podcasters who record talking head content and need to transform rambling footage into polished, structured scripts that flow perfectly.
				</p>
			</div>

			<!-- Marketing Teams -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Marketing Teams</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Digital marketers creating talking head videos for campaigns who need to extract the best content quickly and reorder it into compelling narratives that convert.
				</p>
			</div>

			<!-- Educators -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Online Educators</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Course creators and trainers who record talking head lessons and need to restructure their content into clear, logical sequences that enhance learning outcomes.
				</p>
			</div>

			<!-- Business Teams -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-pink-500 to-green-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Business Communicators</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Executives and teams recording presentations, training videos, and announcements who need to turn long-winded recordings into concise, impactful messages.
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
					Every feature is designed to optimize your content before you start editing. From intelligent
					clip selection to AI-powered script reordering, Ramble transforms your raw talking head
					footage into production-ready content that saves hours of post-production work.
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
				From raw talking head footage to production-ready scripts in minutes. Ramble's preprocessing
				pipeline prepares your content before you even start editing.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
			<!-- Step 1: Import -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-green-500 to-blue-500 flex items-center justify-center text-white text-sm font-bold">1</div>
							<h3 class="text-lg font-bold text-foreground">Upload Footage</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							Upload your raw talking head video. Ramble automatically analyzes speech patterns and content structure.
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
							<h3 class="text-lg font-bold text-foreground">Smart Selection</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							AI identifies the best clips based on content quality, speech clarity, and narrative value - no manual scrubbing required.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Step 3: Script Optimization -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 flex items-center justify-center text-white text-sm font-bold">3</div>
							<h3 class="text-lg font-bold text-foreground">Script Reordering</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							AI reorders your selected clips into a coherent, logical script that flows naturally and engages your audience.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Step 4: Handoff -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-gradient-to-r from-pink-500 to-green-500 flex items-center justify-center text-white text-sm font-bold">4</div>
							<h3 class="text-lg font-bold text-foreground">Export & Handoff</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							Export your optimized script and clip sequences to your favorite video editor. Start post-production with perfect content.
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
				Common questions about using Ramble as your preprocessing tool.
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
				Start Optimizing Scripts
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
					<span class="text-primary-foreground font-bold text-sm">R</span>
				</div>
				<div>
					<span class="font-bold tracking-tight text-xl">
						<span class="gradient-text">RAMBLE</span>
					</span>
					<p class="text-muted-foreground text-sm">Script Optimization Tool</p>
				</div>
			</div>
			
			<!-- Contact -->
			<div class="text-center">
				<h4 class="font-semibold text-foreground mb-2">Support</h4>
				<a 
					href="mailto:support@ramble.app" 
					class="text-primary hover:text-primary/80 transition-colors"
				>
					support@ramble.app
				</a>
			</div>
			
			<!-- Copyright -->
			<div class="text-right text-sm text-muted-foreground">
				<p>&copy; {new Date().getFullYear()} Ramble</p>
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
