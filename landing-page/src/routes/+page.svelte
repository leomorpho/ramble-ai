<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';
	import { slide } from 'svelte/transition';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import ProjectIntakeForm from '$lib/components/ProjectIntakeForm.svelte';
	import ProjectCard from '$lib/components/ProjectCard.svelte';
	import TeamMemberCard from '$lib/components/TeamMemberCard.svelte';
	import CodeAnimation from '$lib/components/CodeAnimation.svelte';
	import Navigation from '$lib/components/Navigation.svelte';

	const projects = [
		{
			title: 'BONFIRE',
			subtitle: 'Community Platform',
			client: 'Creators & Communities',
			description:
				'Community platform designed to help creators build meaningful connections through effortless event hosting.',
			impact: 'Real community building • Simplified event hosting',
			tech: ['React', 'Node.js', 'AI/ML', 'PostgreSQL'],
			score: '92',
			year: '2024',
			image: 'https://tobyluxembourg.com/img/bonfire.png',
			url: 'https://bnfr-events.app/'
		},
		{
			title: 'GOSHIP',
			subtitle: 'Business Launch Tool',
			client: 'Entrepreneurs',
			description:
				'Business launch tools that help entrepreneurs get their ideas off the ground with less technical complexity.',
			impact: 'Faster launches • Lower technical barriers',
			tech: ['Go', 'HTMX', 'PostgreSQL', 'Docker'],
			score: '89',
			year: '2024',
			image: 'https://tobyluxembourg.com/img/goship-icon.png',
			url: 'https://goship.run/'
		},
		{
			title: 'ECLAIR',
			subtitle: 'Delivery Tracking',
			client: 'Local Business',
			description:
				'Delivery tracking automation that integrates with Slack and reduces manual customer communication.',
			impact: 'Reduced manual work • Better customer visibility',
			tech: ['Python', 'Slack API', 'Image Processing', 'AWS'],
			score: '87',
			year: '2024',
			image: 'https://tobyluxembourg.com/img/eclair-icon.png',
			url: null
		},
		{
			title: 'CHÉRIE',
			subtitle: 'Connection App',
			client: 'Personal Project',
			description:
				'Simple relationship app designed to help couples maintain daily connection habits.',
			impact: 'Stronger relationships • Simple daily habits',
			tech: ['React Native', 'Firebase', 'Push Notifications'],
			score: '85',
			year: '2023',
			image: 'https://tobyluxembourg.com/img/cherie-icon.png',
			url: 'https://cherie.chatbond.app/'
		},
		{
			title: 'CANACOMPOST',
			subtitle: 'Agricultural Innovation',
			client: 'NASA/CSA Challenge',
			description:
				'Automated insect-farming solution for space food production, contributed to NASA Deep Space Food Challenge.',
			impact: '$30,000 NASA grant • Breakthrough innovation',
			tech: ['IoT', 'Automation', 'Sensors', 'Data Analytics'],
			score: '94',
			year: '2022',
			image: 'https://tobyluxembourg.com/img/Canacompost.png',
			url: 'https://www.youtube.com/watch?v=xuLiqHc1crs'
		}
	];

	// Contact email from environment variables
	const contactEmail = import.meta.env.VITE_CONTACT_EMAIL || 'hello@goosebyteshq.com';

	// FAQ data and state
	let openFaqId = $state<string | null>(null);

	const faqs = [
		{
			id: 'cost',
			question: 'How much does a typical project cost?',
			answer: 'We offer flexible payment structures on a sliding scale from traditional freelancing to equity partnerships. We take on a wide range of projects, from $10k to $100k+, depending on scope and complexity. We pour our heart into every project regardless of the model. With equity partnerships, the unique advantage is that you gain us as your long-term technical team - we\'re there to help your product grow and evolve. Having skin in the game makes the work exhilarating and aligns our success with yours. With equity arrangements, upfront costs can be significantly reduced. We\'ll provide transparent estimates based on your preferred model.'
		},
		{
			id: 'timeline',
			question: 'How long does it take to build something?',
			answer: 'Most projects take 2-6 months depending on complexity. We break work into short cycles so you see progress every 2 weeks. Simple websites might be done in 4-6 weeks, while custom applications typically take 3-4 months. We\'ll give you a realistic timeline estimate upfront.'
		},
		{
			id: 'process',
			question: 'What\'s your development process like?',
			answer: 'We start with deep conversations to understand what you really need. Then we build in short cycles, showing you progress every 2 weeks and getting your feedback. This prevents misunderstandings and ensures we\'re building exactly what will help your business succeed.'
		},
		{
			id: 'maintenance',
			question: 'What happens after the project is done?',
			answer: 'When we partner through equity arrangements, the project is never truly "done" - we\'re your technical team for the long haul. We\'ll be there to help your product grow, scale, and evolve as your business needs change. For traditional freelance projects, we provide ongoing support plans including security updates, bug fixes, and feature development. Either way, we\'re committed to your success beyond the initial launch.'
		},
		{
			id: 'team-size',
			question: 'Do I need a big budget to work with you?',
			answer: 'Not at all! We offer equity-based partnerships that can dramatically reduce upfront costs. In these arrangements, you\'ll only need to provide a base deposit (which is returned once the app is live) as we\'re putting significant skin in the game ourselves. This deposit ensures commitment from all parties. We also offer traditional payment models and can help you prioritize features to fit your budget. We believe great software should be accessible to visionary entrepreneurs regardless of their current funding.'
		},
		{
			id: 'technology',
			question: 'What technologies do you use?',
			answer: 'We choose technology based on your specific needs, not trends. We\'re experienced with modern web frameworks (React, Svelte, Vue), backend systems (Node.js, Python, Go), databases (PostgreSQL, MongoDB), and cloud platforms (AWS, Vercel, Cloudflare). We prioritize proven, maintainable solutions.'
		},
		{
			id: 'communication',
			question: 'How do we stay in touch during the project?',
			answer: 'We provide regular updates via your preferred communication method - email, Slack, or scheduled calls. You\'ll see working versions of your project every 2 weeks, and we\'re always available for questions. We believe in transparent, frequent communication.'
		},
		{
			id: 'changes',
			question: 'What if I want to change something mid-project?',
			answer: 'Changes are normal and expected! Our short development cycles make it easy to adjust direction. We\'ll discuss the impact on timeline and budget upfront, but we build flexibility into our process specifically to accommodate evolving needs.'
		}
	];

	// FAQ accordion functionality
	const toggleFaq = (faqId: string) => {
		openFaqId = openFaqId === faqId ? null : faqId;
	};

	const team = [
		{
			name: 'Léo Audibert',
			role: 'Founder & Full-Stack Engineer',
			bio: 'Passionate about building products that matter. Léo brings full-stack engineering expertise with deep AI/ML knowledge, helping teams navigate from initial idea to successful launch.',
			image:
				'/team/leo.jpg',
			skills: ['Full-Stack Development', 'AI/ML Integration', 'Project Leadership']
		},
		{
			name: 'Christa Klingensmith',
			role: 'Product Owner',
			bio: 'Expert at figuring out what clients actually need versus what they think they need. Christa ensures we build the right thing, not just the requested thing.',
			image: '/team/christa.jpeg',
			skills: ['Needs Discovery', 'Client Partnership', 'Project Management']
		},
		{
			name: 'John Buonassissi',
			role: 'Backend Engineer',
			bio: 'Believes great software should work reliably in the background. John builds systems that grow with your business without breaking.',
			image: '/team/john.jpeg',
			skills: ['Reliable Systems', 'API Development', 'Problem Solving']
		},
		{
			name: 'Martin Kuerbis',
			role: 'Technical Lead',
			bio: 'Passionate about keeping things simple. Martin brings years of experience helping businesses avoid over-engineered solutions that create more problems than they solve.',
			image: '/team/martin.jpg',
			skills: ['Simplicity Focus', 'Honest Feedback', 'Technical Leadership']
		},
		{
			name: 'Sahil Asthana',
			role: 'Legal & Strategy Advisor',
			bio: 'Bridges technology and legal clarity. Sahil brings deep expertise in AI governance, data privacy, and business architecture to ensure solutions are secure, compliant, and built to scale.',
			image: '/team/sahil.jpeg',
			skills: ['AI Governance', 'Privacy Compliance', 'Business Strategy']
		}
		// {
		// 	name: 'Adam Spilchen',
		// 	role: 'Software Developer',
		// 	bio: 'Believes in building things that last. Adam approaches software development like route development - carefully planned, well-tested, and safe for everyone to use.',
		// 	image:
		// 		'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=400&h=400&fit=crop&crop=face',
		// 	skills: ['Quality Focus', 'Team Collaboration', 'Careful Planning']
		// }
	];

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

			// Project cards stagger (desktop only)
			const cardObserver = new IntersectionObserver(
				(entries) => {
					entries.forEach((entry) => {
						if (entry.isIntersecting) {
							const cards = document.querySelectorAll('.project-card');
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

			document.querySelectorAll('.project-card').forEach((card) => {
				cardObserver.observe(card);
			});
		} else {
			// Mobile: Make all elements immediately visible without animation
			document.querySelectorAll('.fade-up').forEach((element) => {
				(element as HTMLElement).style.opacity = '1';
				(element as HTMLElement).style.transform = 'translateY(0px)';
			});

			document.querySelectorAll('.project-card').forEach((card) => {
				(card as HTMLElement).style.opacity = '1';
				(card as HTMLElement).style.transform = 'translateY(0px)';
			});
		}

	});
</script>

<svelte:head>
	<title>GooseBytes • Your Technical Co-Pilot for Business Growth</title>
	<meta
		name="description"
		content="Passionate engineers helping small businesses and creators build the right software solutions. We'll tell you honestly what won't work."
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
					<span class="block text-foreground">YOUR</span>
					<span class="block gradient-text font-black">DIGITAL</span>
					<span class="block text-foreground">ARTISANS</span>
				</h1>

				<div class="space-y-8 max-w-3xl">
					<p
						class="hero-subtitle text-2xl lg:text-3xl text-foreground/90 font-medium leading-relaxed"
					>
						Every visionary deserves high-quality software, built simply and effectively.
					</p>
				</div>
			</div>

			<div class="hero-desc"></div>

			<div class="hero-buttons flex flex-wrap gap-6 pt-4">
				<Button
					size="lg"
					class="shadow-lg shadow-green-500/25 hover:shadow-green-500/40 transition-all duration-300"
					onclick={() => document.getElementById('work')?.scrollIntoView({ behavior: 'smooth' })}
				>
					See What We've Built
					<svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 14l-7 7m0 0l-7-7m7 7V3"
						></path>
					</svg>
				</Button>
				<ProjectIntakeForm
					triggerText="Let's Talk About Your Challenges"
					variant="outline"
					size="lg"
				/>
			</div>
		</div>

		<!-- Right Side - Animation Container -->
		<div class="hidden lg:block lg:col-span-5 relative">
			<div class="relative w-full h-96 overflow-visible">
				<!-- Code Animation - Centered -->
				<CodeAnimation maxSnippets={1} />
			</div>
		</div>
	</div>
</section>

<!-- Services Section -->
<section id="services" class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				WHAT WE'RE<br />
				<span class="gradient-text">EXPERTS</span> AT
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				We've built deep expertise in the technologies and strategies that help businesses grow.
			</p>
		</div>

		<div class="text-center">
			<p class="text-2xl lg:text-3xl text-foreground leading-relaxed max-w-5xl mx-auto">
				<span class="font-semibold">AI Solutions</span> • 
				<span class="font-semibold">Advanced Data Architecture</span> • 
				<span class="font-semibold">Apps</span> • 
				<span class="font-semibold">Automation</span> • 
				<span class="font-semibold">Digital Innovation & Strategy</span> • 
				<span class="font-semibold">Enterprise Digital Transformation</span> • 
				<span class="font-semibold">Performance Technology</span> • 
				<span class="font-semibold">Web Design</span> • 
				<span class="font-semibold">Web Development</span>
			</p>
		</div>
	</div>
</section>

<!-- Who We Work With Section -->
<section class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				WHO WE<br />
				<span class="gradient-text">PARTNER</span> WITH
			</h2>
			<p class="text-muted-foreground text-lg max-w-3xl mx-auto">
				We work with ambitious people who have great ideas but need the right technical partner to bring them to life. Whether you're bootstrapping your first startup or scaling an established business, we meet you where you are.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
			<!-- Entrepreneurs -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-green-500 to-blue-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Entrepreneurs</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Building your first product or scaling to new markets. We help you navigate technical decisions and avoid costly mistakes.
				</p>
			</div>

			<!-- Small Business Owners -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Small Business Owners</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Looking to automate processes, improve efficiency, or create better customer experiences through technology.
				</p>
			</div>

			<!-- Creators & Communities -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Creators & Communities</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Building platforms, tools, and experiences that help you connect with your audience in meaningful ways.
				</p>
			</div>

			<!-- Innovation Teams -->
			<div class="space-y-4">
				<div class="h-1 w-12 bg-gradient-to-r from-pink-500 to-green-500 rounded-full"></div>
				<h3 class="text-xl font-bold text-foreground">Innovation Teams</h3>
				<p class="text-muted-foreground text-sm leading-relaxed">
					Exploring new technologies or tackling ambitious challenges that require deep technical expertise and creative problem-solving.
				</p>
			</div>
		</div>
	</div>
</section>

<!-- Projects Section -->
<section id="work" class="py-20 bg-background">
	<div class="max-w-7xl mx-auto px-6">
		<div class="grid lg:grid-cols-12 gap-12 mb-20">
			<div class="lg:col-span-4">
				<div class="fade-up space-y-4">
					<h2 class="text-4xl lg:text-5xl text-headline text-foreground">
						WHAT WE'VE<br />
						<span class="gradient-text">BUILT</span>
					</h2>
				</div>
			</div>
			<div class="lg:col-span-8">
				<p class="fade-up text-muted-foreground text-lg leading-relaxed">
					These projects show our technical capabilities, but we're excited to apply this expertise
					to the challenges small businesses face every day. We believe in building solutions that
					actually work for real people.
				</p>
			</div>
		</div>

		<div
			class="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-6"
			style="grid-auto-rows: 1fr;"
		>
			{#each projects as project, i (project.title)}
				<ProjectCard {project} index={i} />
			{/each}
		</div>
	</div>
</section>

<!-- Team Section -->
<section id="team" class="py-20 bg-gradient-to-b from-muted/0 via-muted/20 to-muted/0">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				MEET YOUR<br />
				<span class="gradient-text">DIGITAL ARCHITECTS</span>
			</h2>
		</div>

		<div class="flex flex-wrap justify-center gap-12">
			{#each team as member, i (member.name)}
				<div class="w-full sm:w-[calc(50%-1.5rem)] lg:w-[calc(33.333%-2rem)]">
					<TeamMemberCard {member} index={i} />
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- Process Section -->
<section id="process" class="py-20 bg-background">
	<div class="max-w-7xl mx-auto px-6">
		<div class="fade-up text-center mb-20">
			<h2 class="text-5xl lg:text-6xl text-headline text-foreground mt-8 mb-6">
				LET'S FIGURE OUT<br />
				<span class="gradient-text">WHAT YOU</span><br />
				REALLY NEED
			</h2>
			<p class="text-muted-foreground text-lg max-w-2xl mx-auto">
				We're not claiming to know your business better than you do — but we know technology, and
				we'll help you figure out what actually makes sense.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
			<!-- Understanding Together -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<span class="text-xl">💬</span>
							<h3 class="text-xl font-bold text-foreground">Understanding Together</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							We collaborate through deep conversations to uncover what you really need, not just
							what you initially request. This prevents costly detours and ensures we build what
							truly matters.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Build Smart, Test Early -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<span class="text-xl">⚡</span>
							<h3 class="text-xl font-bold text-foreground">Build Smart, Test Early</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							We choose simple solutions, build in short cycles, and get your feedback at every
							step. This saves time, reduces risk, and keeps you in control.
						</p>
					</div>
				</CardContent>
			</Card>

			<!-- Advice That Saves You Money -->
			<Card class="group hover:shadow-xl transition-all duration-300 border-0 rounded-md bg-slate-100 dark:bg-slate-900">
				<CardContent class="p-6">
					<div class="space-y-4">
						<div class="flex items-center gap-3">
							<span class="text-xl">💡</span>
							<h3 class="text-xl font-bold text-foreground">Advice That Saves You Money</h3>
						</div>

						<p class="text-muted-foreground text-sm leading-relaxed">
							We'll tell you when a simpler solution exists, when to pivot, and if something
							probably won't work. Our honest feedback prevents costly mistakes.
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
			<ProjectIntakeForm triggerText="Let's Talk About Your Challenge" size="lg" />
		</div>
	</div>
</section>

<!-- Footer -->
<footer class="bg-background border-t border-border py-12">
	<div class="max-w-7xl mx-auto px-6">
		<div class="grid md:grid-cols-3 gap-8 items-center">
			<!-- Logo/Brand -->
			<div class="flex items-center space-x-3">
				<div class="w-8 h-8 rounded-full overflow-hidden ring-2 ring-primary/20">
					<img 
						src="/logo-128.png" 
						alt="GooseBytes Logo" 
						class="w-full h-full object-cover"
					/>
				</div>
				<div>
					<span class="font-bold tracking-tight">
						<span class="text-foreground">GOOSE</span><span class="gradient-text">BYTES</span>
					</span>
					<p class="text-muted-foreground text-sm">Your Digital Artisans</p>
				</div>
			</div>
			
			<!-- Contact -->
			<div class="text-center">
				<h4 class="font-semibold text-foreground mb-2">Get In Touch</h4>
				<a 
					href="mailto:{contactEmail}" 
					class="text-primary hover:text-primary/80 transition-colors"
				>
					{contactEmail}
				</a>
			</div>
			
			<!-- Copyright -->
			<div class="text-right text-sm text-muted-foreground">
				<p>&copy; {new Date().getFullYear()} GooseBytes</p>
				<p>Built with care</p>
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

	:global(.project-card) {
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
