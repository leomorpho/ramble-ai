<script lang="ts">
	import { subscriptionStore } from '$lib/stores/subscription.svelte.ts';
	import { authStore } from '$lib/stores/authClient.svelte.ts';
	import { createCheckoutSession } from '$lib/stripe.ts';
	import { config } from '$lib/config.ts';
	import { Loader2, Check, Crown, Zap } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	
	let isLoading = $state(false);
	let checkoutLoading = $state<string | null>(null);
	let billingInterval = $state<'month' | 'year'>('month');

	onMount(() => {
		subscriptionStore.initialize();
		subscriptionStore.refresh();
	});

	async function handleSubscribe(planId: string) {
		if (!authStore.isLoggedIn) {
			// Redirect to login
			window.location.href = '/login?redirect=/pricing';
			return;
		}

		// Free plan doesn't need Stripe checkout
		const plan = subscriptionStore.getPlan(planId);
		if (plan?.billing_interval === 'free') {
			// Could implement switching to free plan here if needed
			return;
		}

		checkoutLoading = planId;
		try {
			await createCheckoutSession(planId);
		} catch (error) {
			console.error('Error creating checkout session:', error);
			alert('Failed to start checkout. Please try again.');
		} finally {
			checkoutLoading = null;
		}
	}

	function isCurrentPlan(planId: string): boolean {
		return subscriptionStore.isCurrentPlan(planId);
	}

	function getButtonText(planId: string): string {
		if (checkoutLoading === planId) return 'Processing...';
		if (isCurrentPlan(planId)) return 'Current Plan';
		
		const plan = subscriptionStore.getPlan(planId);
		if (plan?.billing_interval === 'free') {
			return 'Free Plan';
		}
		
		if (!authStore.isLoggedIn) return 'Sign Up to Subscribe';
		if (subscriptionStore.isSubscribed) return 'Switch Plan';
		return 'Subscribe';
	}

	function isButtonDisabled(planId: string): boolean {
		const plan = subscriptionStore.getPlan(planId);
		if (plan?.billing_interval === 'free') return true; // Free plan doesn't need a button action
		return checkoutLoading !== null || isCurrentPlan(planId);
	}

	function getPlanIcon(planName: string) {
		if (planName.toLowerCase().includes('pro')) return Crown;
		if (planName.toLowerCase().includes('basic')) return Zap;
		return Check;
	}

	function getPlansForInterval(interval: 'month' | 'year') {
		return subscriptionStore.plans
			.filter(plan => plan.billing_interval === interval || plan.billing_interval === 'free')
			.sort((a, b) => a.display_order - b.display_order);
	}

	function calculateSavings(monthlyPrice: number, yearlyPrice: number): number {
		const monthlyTotal = monthlyPrice * 12;
		return Math.round(((monthlyTotal - yearlyPrice) / monthlyTotal) * 100);
	}

	// Get the monthly equivalent price for yearly plans
	function getMonthlyEquivalent(plan: any) {
		if (plan.billing_interval === 'year') {
			return plan.price_cents / 12;
		}
		return plan.price_cents;
	}

	function hasYearlyPlans(): boolean {
		return subscriptionStore.plans.some(plan => plan.billing_interval === 'year');
	}
</script>

<svelte:head>
	<title>Pricing - {config.app.name}</title>
	<meta name="description" content="Choose the perfect plan for your needs" />
</svelte:head>

<!-- Hero Section -->
<section class="py-20 px-6">
	<div class="max-w-4xl mx-auto text-center">
		<h1 class="text-4xl md:text-5xl font-bold mb-6">Choose Your Plan</h1>
		<p class="text-xl text-muted-foreground">
			Process more audio, get unlimited exports, and access premium features.
		</p>
	</div>
</section>

<!-- Pricing Plans -->
<section class="py-20 border-t px-6">
	<div class="max-w-7xl mx-auto">
		{#if subscriptionStore.isLoading}
			<div class="text-center py-8">
				<Loader2 class="h-6 w-6 animate-spin mx-auto mb-3" />
				<p class="text-sm text-muted-foreground">Loading pricing plans...</p>
			</div>
		{:else if subscriptionStore.plans.length === 0}
			<div class="text-center py-12">
				<p class="text-muted-foreground">No pricing plans available at the moment.</p>
				<p class="text-sm text-muted-foreground mt-2">
					Please check back later or contact support.
				</p>
			</div>
		{:else}
			<!-- Billing Toggle (only show if yearly plans exist) -->
			{#if hasYearlyPlans()}
				<div class="flex justify-center mb-12">
					<div class="flex items-center bg-muted p-1 rounded-lg">
						<button
							class="px-6 py-2 rounded-md text-sm font-medium transition-colors {billingInterval === 'month' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}"
							onclick={() => billingInterval = 'month'}
						>
							Monthly
						</button>
						<button
							class="px-6 py-2 rounded-md text-sm font-medium transition-colors {billingInterval === 'year' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}"
							onclick={() => billingInterval = 'year'}
						>
							Yearly
							<Badge variant="secondary" class="ml-2">Save 20%</Badge>
						</button>
					</div>
				</div>
			{/if}

			<!-- Plans Grid -->
			<div class="grid gap-6 md:grid-cols-3">
				{#each getPlansForInterval(billingInterval) as plan (plan.id)}
					{@const Icon = getPlanIcon(plan.name)}
					{@const isPopular = plan.name.toLowerCase().includes('basic')}
					{@const isCurrentPlan = subscriptionStore.isCurrentPlan(plan.id)}
					
					<Card class="relative {isPopular ? 'ring-2 ring-primary' : ''} {isCurrentPlan ? 'bg-muted/50' : ''}">
						{#if isPopular}
							<Badge class="absolute -top-3 left-1/2 -translate-x-1/2">
								Most Popular
							</Badge>
						{/if}
						
						<CardHeader class="text-center">
							<div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
								<Icon class="h-6 w-6 text-primary" />
							</div>
							
							<CardTitle class="text-2xl">{plan.name}</CardTitle>
							
							<div class="mt-4">
								{#if plan.billing_interval === 'free'}
									<div class="text-4xl font-bold">Free</div>
									<div class="text-muted-foreground">Always free</div>
								{:else}
									<div class="text-4xl font-bold">
										{subscriptionStore.formatPrice(plan.price_cents)}
									</div>
									<div class="text-muted-foreground">
										per {plan.billing_interval}
										{#if plan.billing_interval === 'year'}
											<div class="text-sm text-green-600 mt-1">
												({subscriptionStore.formatPrice(getMonthlyEquivalent(plan))} per month)
											</div>
										{/if}
									</div>
								{/if}
							</div>
						</CardHeader>
						
						<CardContent class="space-y-6">
							<div class="text-center">
								<div class="text-2xl font-semibold text-primary">
									{plan.hours_per_month} hour{plan.hours_per_month !== 1 ? 's' : ''}
								</div>
								<div class="text-sm text-muted-foreground">of media processing per month</div>
							</div>

							<ul class="space-y-3">
								{#each plan.features as feature}
									<li class="flex items-center">
										<Check class="h-4 w-4 text-green-600 mr-3 flex-shrink-0" />
										<span class="text-sm">{feature}</span>
									</li>
								{/each}
							</ul>

							<Button 
								class="w-full" 
								variant={isCurrentPlan ? "secondary" : "default"}
								disabled={isButtonDisabled(plan.id)}
								onclick={() => handleSubscribe(plan.id)}
							>
								{getButtonText(plan.id)}
							</Button>
						</CardContent>
					</Card>
				{/each}
			</div>
		{/if}

		<!-- Current Subscription Status -->
		{#if subscriptionStore.isSubscribed}
			<div class="mt-16 text-center">
				<div class="rounded-lg bg-green-50 border border-green-200 p-8 inline-block">
					<h3 class="text-2xl font-semibold text-green-800 mb-4">You're subscribed!</h3>
					<p class="text-green-700 mb-6 text-lg">
						Manage your subscription, update payment methods, and view billing history.
					</p>
					<Button variant="outline" onclick={() => window.location.href = '/billing'}>
						Manage Subscription
					</Button>
				</div>
			</div>
		{/if}

		<!-- Usage Warning -->
		{#if subscriptionStore.usageWarning}
			<div class="mt-8 mx-auto max-w-2xl">
				<div class="rounded-lg border border-yellow-200 bg-yellow-50 p-4">
					<div class="text-yellow-800">
						<strong>Usage Notice:</strong> {subscriptionStore.usageWarning.message}
					</div>
				</div>
			</div>
		{/if}
	</div>
</section>

<!-- FAQ/Features Section -->
<section class="py-20 border-t px-6">
	<div class="max-w-4xl mx-auto">
		<div class="text-center mb-12">
			<h2 class="text-3xl md:text-4xl font-bold mb-6">All Plans Include</h2>
		</div>
		
		<div class="grid md:grid-cols-2 gap-8">
			<div class="space-y-4">
				<h3 class="text-xl font-semibold mb-4">Core Features</h3>
				<ul class="space-y-3">
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>High-quality audio transcription</span>
					</li>
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>Unlimited video quality exports</span>
					</li>
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>Multiple export formats</span>
					</li>
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>Secure file processing</span>
					</li>
				</ul>
			</div>
			
			<div class="space-y-4">
				<h3 class="text-xl font-semibold mb-4">Support</h3>
				<ul class="space-y-3">
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>Email support</span>
					</li>
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>Cancel anytime</span>
					</li>
					<li class="flex items-center">
						<Check class="h-5 w-5 text-green-600 mr-3" />
						<span>No long-term contracts</span>
					</li>
				</ul>
			</div>
		</div>
		
		<div class="text-center mt-12">
			<h3 class="text-xl font-semibold mb-4">Questions?</h3>
			<p class="text-muted-foreground">
				Need help choosing the right plan? Contact our support team for assistance.
			</p>
		</div>
	</div>
</section>