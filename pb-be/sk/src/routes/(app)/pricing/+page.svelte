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

	// Subscription store is initialized in root layout
	onMount(() => {
		// Refresh data to ensure it's current
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
	<div class="max-w-4xl mx-auto">
		<h1 class="text-4xl md:text-5xl font-bold mb-6">Choose Your Plan</h1>
		<p class="text-xl text-muted-foreground">
			Process more audio, get unlimited exports, and access premium features.
		</p>
	</div>
</section>

<!-- Pricing Plans -->
<section class="py-20 border-t px-6">
	<div class="max-w-4xl mx-auto">
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
					{@const isPopular = plan.name.toLowerCase().includes('basic')}
					{@const isCurrentPlan = subscriptionStore.isCurrentPlan(plan.id)}
					
					<div class="border rounded-lg p-6 {isPopular ? 'border-primary' : ''} {isCurrentPlan ? 'bg-muted/30' : ''}">
						{#if isPopular}
							<Badge class="mb-4">Most Popular</Badge>
						{/if}
						
						<div class="text-center mb-6">
							<h3 class="text-xl font-semibold mb-2">{plan.name}</h3>
							
							<div class="mb-4">
								{#if plan.billing_interval === 'free'}
									<div class="text-3xl font-bold">Free</div>
									<div class="text-sm text-muted-foreground">Always free</div>
								{:else}
									<div class="text-3xl font-bold">
										{subscriptionStore.formatPrice(plan.price_cents)}
									</div>
									<div class="text-sm text-muted-foreground">
										per {plan.billing_interval}
										{#if plan.billing_interval === 'year'}
											<div class="text-sm text-green-600 mt-1">
												({subscriptionStore.formatPrice(getMonthlyEquivalent(plan))} per month)
											</div>
										{/if}
									</div>
								{/if}
							</div>

							<div class="text-lg font-medium text-primary mb-4">
								{plan.hours_per_month} hour{plan.hours_per_month !== 1 ? 's' : ''} per month
							</div>
						</div>

						<ul class="space-y-2 mb-6">
							{#each plan.features as feature}
								<li class="flex items-start gap-2 text-sm">
									<Check class="h-4 w-4 text-green-600 mt-0.5 flex-shrink-0" />
									<span>{feature}</span>
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
					</div>
				{/each}
			</div>
		{/if}

		<!-- Current Subscription Status -->
		{#if subscriptionStore.isSubscribed}
			<div class="mt-12">
				<div class="border rounded-lg p-6 bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-800">
					<h3 class="text-lg font-semibold text-green-800 dark:text-green-200 mb-2">You're subscribed!</h3>
					<p class="text-green-700 dark:text-green-300 mb-4">
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
			<div class="mt-8">
				<div class="border rounded-lg p-4 bg-yellow-50 dark:bg-yellow-950/30 border-yellow-200 dark:border-yellow-800">
					<div class="text-yellow-800 dark:text-yellow-200">
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
		<h2 class="text-3xl md:text-4xl font-bold mb-12">All Plans Include</h2>
		
		<div class="grid md:grid-cols-2 gap-8 mb-12">
			<div>
				<h3 class="text-lg font-semibold mb-4">Core Features</h3>
				<ul class="space-y-3">
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>High-quality audio transcription</span>
					</li>
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>Unlimited video quality exports</span>
					</li>
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>Multiple export formats</span>
					</li>
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>Secure file processing</span>
					</li>
				</ul>
			</div>
			
			<div>
				<h3 class="text-lg font-semibold mb-4">Support</h3>
				<ul class="space-y-3">
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>Email support</span>
					</li>
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>Cancel anytime</span>
					</li>
					<li class="flex items-center gap-2">
						<Check class="h-4 w-4 text-green-600" />
						<span>No long-term contracts</span>
					</li>
				</ul>
			</div>
		</div>
		
		<div class="border rounded-lg p-6 text-center">
			<h3 class="text-lg font-semibold mb-2">Questions?</h3>
			<p class="text-muted-foreground">
				Need help choosing the right plan? Contact our support team for assistance.
			</p>
		</div>
	</div>
</section>