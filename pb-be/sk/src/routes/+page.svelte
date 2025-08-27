<script lang="ts">
	import { onMount } from 'svelte';
	import { gsap } from 'gsap';
	import { ScrollTrigger } from 'gsap/dist/ScrollTrigger';
	import { CustomEase } from 'gsap/dist/CustomEase';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { config } from '$lib/config.js';
	import { authStore } from '$lib/stores/authClient.svelte.js';
	import { subscriptionStore } from '$lib/stores/subscription.svelte.js';
	import { Zap, Shield, Users, Rocket, Star, CheckCircle } from 'lucide-svelte';

	gsap.registerPlugin(ScrollTrigger, CustomEase);

	// Initialize subscription store in component context
	subscriptionStore.initialize();

	const features = [
		{
			icon: Rocket,
			title: 'Production Ready',
			subtitle: 'Zero Configuration',
			description: 'Complete full-stack starter with SvelteKit + PocketBase. Deploy in minutes, not weeks.',
			benefits: ['Static site generation', 'Go-powered backend', 'Database included'],
			gradient: 'from-blue-500 to-purple-600'
		},
		{
			icon: Shield,
			title: 'Modern Authentication',
			subtitle: 'WebAuthn + Passkeys',
			description: 'Passwordless authentication with WebAuthn passkeys, progressive login flow, and enterprise security.',
			benefits: ['Passwordless login', 'Progressive UX', 'Multi-factor auth'],
			gradient: 'from-green-500 to-teal-600'
		},
		{
			icon: Users,
			title: 'Payment Integration',
			subtitle: 'Stripe Built-in',
			description: 'Complete subscription management with Stripe integration, webhooks, and billing portal.',
			benefits: ['Subscription billing', 'Webhook automation', 'Customer portal'],
			gradient: 'from-pink-500 to-rose-600'
		},
		{
			icon: Zap,
			title: 'Developer Experience',
			subtitle: 'Everything Included',
			description: 'TypeScript, Tailwind, shadcn-svelte, testing, i18n, file uploads - all configured and ready.',
			benefits: ['Pre-commit hooks', 'Vitest + Playwright', 'TUS file uploads'],
			gradient: 'from-yellow-500 to-orange-600'
		}
	];

	const stats = [
		{ number: '24+', label: 'Features Included' },
		{ number: '10x', label: 'Faster Development' },
		{ number: '100%', label: 'Open Source' },
		{ number: '0 min', label: 'Setup Time' }
	];

	onMount(() => {
		// Enable all animations with GSAP
		
		// Smooth blob animations - faster pace
		gsap.to('.blob-1', {
			x: 'random(-200, 200)',
			y: 'random(-200, 200)',
			rotation: 'random(-180, 180)',
			duration: 'random(2, 5)',
			repeat: -1,
			yoyo: true,
			ease: 'sine.inOut'
		});

		gsap.to('.blob-2', {
			x: 'random(-250, 250)',
			y: 'random(-150, 150)',
			rotation: 'random(-180, 180)',
			duration: 'random(2, 7)',
			repeat: -1,
			yoyo: true,
			ease: 'sine.inOut',
			delay: 1
		});

		gsap.to('.blob-3', {
			x: 'random(-180, 180)',
			y: 'random(-220, 220)',
			rotation: 'random(-180, 180)',
			duration: 'random(2, 6)',
			repeat: -1,
			yoyo: true,
			ease: 'sine.inOut',
			delay: 1
		});

		// Hero content animations
		gsap.from('.hero-badge', { opacity: 0, y: 30, duration: 1, ease: 'power2.out' });
		gsap.from('.hero-title', { opacity: 0, y: 50, duration: 1.2, delay: 0.2, ease: 'power2.out' });
		gsap.from('.hero-subtitle', { opacity: 0, y: 30, duration: 1, delay: 0.4, ease: 'power2.out' });
		gsap.from('.hero-desc', { opacity: 0, y: 20, duration: 1, delay: 0.6, ease: 'power2.out' });
		gsap.from('.hero-buttons', { opacity: 0, y: 20, duration: 1, delay: 0.8, ease: 'power2.out' });
		gsap.from('.hero-stats', { opacity: 0, y: 20, duration: 1, delay: 1, ease: 'power2.out' });

		// Feature cards gentle floating animation only
		gsap.to('.feature-card', {
			y: -8,
			duration: 3,
			ease: 'sine.inOut',
			repeat: -1,
			yoyo: true,
			stagger: 0.3
		});

		// Tech stack cards gentle floating
		gsap.to('.tech-card', {
			y: -6,
			duration: 2.5,
			ease: 'sine.inOut',
			repeat: -1,
			yoyo: true,
			stagger: 0.2
		});

		// Section headings animation
		gsap.from('.section-heading', {
			opacity: 0,
			y: 30,
			duration: 1,
			ease: 'power2.out',
			scrollTrigger: {
				trigger: '.section-heading',
				start: 'top 80%'
			}
		});

		// CTA section animation
		gsap.from('.cta-content', {
			opacity: 0,
			y: 40,
			duration: 1.2,
			ease: 'power2.out',
			scrollTrigger: {
				trigger: '.cta-content',
				start: 'top 80%'
			}
		});

		// Stats floating animation
		gsap.to('.stat-item', {
			y: -5,
			duration: 1.5,
			ease: 'sine.inOut',
			repeat: -1,
			yoyo: true,
			stagger: 0.2
		});
	});
</script>

<svelte:head>
	<title>{config.app.name} • Powerful SaaS Starter Kit</title>
	<meta
		name="description"
		content="Launch your SaaS in days, not months. Complete full-stack starter with SvelteKit, PocketBase, WebAuthn, Stripe, TUS uploads, and more."
	/>
</svelte:head>

<!-- Hero Section -->
<section class="relative min-h-screen flex items-center overflow-hidden bg-background">
	<!-- Background Effects -->
	<div class="absolute inset-0 overflow-hidden z-0">
		<div class="blob-1 absolute top-20 left-20 w-96 h-96 bg-gradient-to-r from-blue-400/20 to-purple-600/20 rounded-full blur-3xl"></div>
		<div class="blob-2 absolute bottom-20 right-20 w-80 h-80 bg-gradient-to-r from-pink-400/20 to-rose-600/20 rounded-full blur-3xl"></div>
		<div class="blob-3 absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-to-r from-green-400/10 to-teal-600/10 rounded-full blur-3xl"></div>
		
		<!-- Grid Pattern -->
		<div class="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]"></div>
	</div>

	<div class="relative z-10 max-w-7xl mx-auto px-6 grid lg:grid-cols-12 gap-12 items-center">
		<!-- Left Side -->
		<div class="lg:col-span-7 space-y-8">
			<div class="hero-badge">
				<Badge class="bg-gradient-to-r from-blue-500 to-purple-600 text-white border-0">
					{#if subscriptionStore.isSubscribed}
						✨ Premium User
					{:else}
						🚀 Now Available
					{/if}
				</Badge>
			</div>

			<div class="space-y-6">
				<h1 class="hero-title text-6xl lg:text-8xl xl:text-9xl font-black leading-none">
					<span class="block text-foreground">POWERFUL</span>
					<span class="block bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
						STARTER
					</span>
					<span class="block text-foreground">KIT</span>
				</h1>

				<p class="hero-subtitle text-xl lg:text-2xl text-muted-foreground font-light max-w-2xl">
					Launch your SaaS in days, not months. Complete full-stack starter with authentication, payments, file uploads, and more.
				</p>
			</div>

			<div class="hero-desc">
				<p class="text-muted-foreground/80 max-w-xl text-sm leading-relaxed">
					SvelteKit + PocketBase + TypeScript + Tailwind + shadcn-svelte + WebAuthn + Stripe + TUS uploads + Testing + i18n - everything you need to ship fast.
				</p>
			</div>

			<div class="hero-buttons flex flex-wrap gap-4">
				{#if authStore.isLoggedIn}
					<Button href="/dashboard" size="lg" class="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
						Go to Dashboard
					</Button>
				{:else}
					<Button href="/signup" size="lg" class="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
						Start Building Now
					</Button>
				{/if}
				<Button href="/about" variant="outline" size="lg">
					See All Features
				</Button>
			</div>

			<!-- Stats -->
			<div class="hero-stats grid grid-cols-2 lg:grid-cols-4 gap-6 pt-8 border-t border-border/50">
				{#each stats as stat}
					<div class="stat-item text-center lg:text-left">
						<div class="text-2xl lg:text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
							{stat.number}
						</div>
						<div class="text-sm text-muted-foreground">{stat.label}</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Right Side - Visual Element -->
		<div class="lg:col-span-5 mt-12 lg:mt-0">
			<div class="relative max-w-md mx-auto lg:max-w-none">
				<div class="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-600/20 rounded-3xl transform rotate-6"></div>
				<div class="relative bg-background/90 backdrop-blur-xl border border-border/50 rounded-3xl p-8 transform -rotate-3 hover:rotate-0 transition-transform duration-700">
					<div class="space-y-4">
						<div class="flex items-center space-x-3">
							<div class="w-3 h-3 bg-red-500 rounded-full"></div>
							<div class="w-3 h-3 bg-yellow-500 rounded-full"></div>
							<div class="w-3 h-3 bg-green-500 rounded-full"></div>
						</div>
						<div class="space-y-3">
							<div class="h-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded w-3/4"></div>
							<div class="h-3 bg-muted rounded w-1/2"></div>
							<div class="h-3 bg-muted rounded w-2/3"></div>
							<div class="grid grid-cols-2 gap-3 mt-6">
								<div class="h-20 bg-gradient-to-br from-green-400 to-teal-600 rounded-lg"></div>
								<div class="h-20 bg-gradient-to-br from-pink-400 to-rose-600 rounded-lg"></div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>

<!-- Features Section -->
<section class="py-32 bg-background">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center mb-20">
			<div class="section-heading space-y-4">
				<Badge variant="secondary">Starter Kit Features</Badge>
				<h2 class="text-4xl lg:text-6xl font-black text-foreground">
					EVERYTHING YOU<br />
					<span class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">NEED TO SHIP</span>
				</h2>
				<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
					Production-ready features that eliminate months of development and configuration
				</p>
			</div>
		</div>

		<div class="grid md:grid-cols-2 gap-8">
			{#each features as feature, i}
				<Card class="feature-card group hover:shadow-2xl transition-all duration-500 border-0 bg-gradient-to-br from-background to-muted/30">
					<CardContent class="p-8">
						<div class="space-y-6">
							<!-- Icon -->
							<div class="feature-icon relative">
								<div class="w-16 h-16 bg-gradient-to-r {feature.gradient} rounded-2xl flex items-center justify-center transform group-hover:scale-110 transition-transform duration-300">
									<svelte:component this={feature.icon} class="h-8 w-8 text-white" />
								</div>
								<div class="absolute -top-1 -right-1 w-6 h-6 bg-gradient-to-r {feature.gradient} rounded-full opacity-60 animate-pulse"></div>
							</div>

							<!-- Content -->
							<div class="space-y-4">
								<div>
									<h3 class="text-2xl font-bold text-foreground">
										{feature.title}
									</h3>
									<p class="text-sm font-medium bg-gradient-to-r {feature.gradient} bg-clip-text text-transparent">
										{feature.subtitle}
									</p>
								</div>
								<p class="text-muted-foreground leading-relaxed">
									{feature.description}
								</p>
							</div>

							<!-- Benefits -->
							<div class="space-y-2">
								{#each feature.benefits as benefit}
									<div class="flex items-center space-x-3">
										<CheckCircle class="h-4 w-4 text-green-500" />
										<span class="text-sm text-muted-foreground">{benefit}</span>
									</div>
								{/each}
							</div>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	</div>
</section>

<!-- Social Proof Section -->
<section class="py-32 bg-muted/20">
	<div class="max-w-7xl mx-auto px-6">
		<div class="text-center mb-16">
			<div class="section-heading">
				<Badge variant="secondary">Tech Stack</Badge>
				<h2 class="text-4xl lg:text-5xl font-black text-foreground mt-8 mb-6">
					BUILT WITH<br />
					<span class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">MODERN STACK</span>
				</h2>
			</div>
		</div>

		<div class="grid md:grid-cols-3 gap-8">
			<Card class="tech-card group hover:shadow-xl transition-all duration-300">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="text-4xl">⚡</div>
						<h3 class="text-xl font-bold">Frontend</h3>
						<p class="text-muted-foreground text-sm">
							SvelteKit 5 with runes, TypeScript, Tailwind CSS, shadcn-svelte components, Paraglide i18n
						</p>
					</div>
				</CardContent>
			</Card>
			<Card class="tech-card group hover:shadow-xl transition-all duration-300">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="text-4xl">🚀</div>
						<h3 class="text-xl font-bold">Backend</h3>
						<p class="text-muted-foreground text-sm">
							PocketBase with Go, SQLite database, real-time subscriptions, WebAuthn passkeys, Stripe integration
						</p>
					</div>
				</CardContent>
			</Card>
			<Card class="tech-card group hover:shadow-xl transition-all duration-300">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="text-4xl">🧪</div>
						<h3 class="text-xl font-bold">Developer Tools</h3>
						<p class="text-muted-foreground text-sm">
							Vitest + Playwright testing, Storybook, ESLint, Prettier, pre-commit hooks, TUS file uploads
						</p>
					</div>
				</CardContent>
			</Card>
		</div>
	</div>
</section>

<!-- CTA Section -->
<section class="py-32 bg-background">
	<div class="max-w-5xl mx-auto px-6 text-center">
		<div class="cta-content space-y-12">
			<Badge variant="secondary">Ready to Start?</Badge>

			<h2 class="text-5xl lg:text-7xl font-black text-foreground">
				CLONE AND<br />
				<span class="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
					START BUILDING
				</span>
			</h2>

			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				Skip months of setup. Get a production-ready SaaS foundation with authentication, payments, file uploads, and more.
			</p>

			<div class="flex flex-col sm:flex-row gap-6 justify-center">
				{#if authStore.isLoggedIn}
					{#if subscriptionStore.isSubscribed}
						<Button href="/dashboard" size="lg" class="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
							Go to Dashboard
						</Button>
						<Button href="/premium" variant="outline" size="lg">
							Explore Premium
						</Button>
					{:else}
						<Button href="/pricing" size="lg" class="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
							Upgrade Now
						</Button>
						<Button href="/dashboard" variant="outline" size="lg">
							Go to Dashboard
						</Button>
					{/if}
				{:else}
					<Button href="/signup" size="lg" class="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
						Get Started Free
					</Button>
					<Button href="/pricing" variant="outline" size="lg">
						View Pricing
					</Button>
				{/if}
			</div>

			<!-- Final Stats -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-8 pt-16 border-t border-border/50">
				{#each stats as stat}
					<div class="stat-item">
						<div class="text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
							{stat.number}
						</div>
						<div class="text-sm text-muted-foreground mt-1">{stat.label}</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</section>

<style>
	/* Allow GSAP animations to work properly */
	:global(.feature-card) {
		will-change: transform;
		/* Ensure cards are visible by default */
		opacity: 1;
	}
	
	:global(.tech-card) {
		will-change: transform;
		/* Ensure cards are visible by default */
		opacity: 1;
	}
	
	:global(.stat-item) {
		will-change: transform;
	}
	
	:global(.section-heading) {
		will-change: opacity, transform;
	}
	
	:global(.cta-content) {
		will-change: opacity, transform;
	}

	:global(.hero-badge),
	:global(.hero-title),
	:global(.hero-subtitle),
	:global(.hero-desc),
	:global(.hero-buttons),
	:global(.hero-stats) {
		will-change: opacity, transform;
	}
</style>