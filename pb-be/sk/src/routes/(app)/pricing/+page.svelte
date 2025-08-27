<script lang="ts">
	import { subscriptionStore } from '$lib/stores/subscription.svelte.js';
	import { authStore } from '$lib/stores/authClient.svelte.js';
	import { createCheckoutSession } from '$lib/stripe.js';
	import { config } from '$lib/config.js';
	import { Loader2 } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import PricingCard from '$lib/components/PricingCard.svelte';
	import CreditCard from '$lib/components/CreditCard.svelte';
	
	let isLoading = $state(false);
	let checkoutLoading = $state<string | null>(null);

	onMount(() => {
		subscriptionStore.initialize();
		subscriptionStore.refresh();
	});

	async function handleSubscribe(priceId: string) {
		if (!authStore.isLoggedIn) {
			// Redirect to login
			window.location.href = '/login?redirect=/pricing';
			return;
		}

		checkoutLoading = priceId;
		try {
			await createCheckoutSession(priceId, 'subscription');
		} catch (error) {
			console.error('Error creating checkout session:', error);
			alert('Failed to start checkout. Please try again.');
		} finally {
			checkoutLoading = null;
		}
	}

	// Group prices by product
	function getProductsWithPrices() {
		return subscriptionStore.products.map(product => ({
			...product,
			prices: subscriptionStore.getPricesForProduct(product.product_id)
				.filter(price => price.type === 'recurring')
				.sort((a, b) => a.unit_amount - b.unit_amount)
		})).filter(product => product.prices.length > 0);
	}

	function isCurrentPlan(priceId: string): boolean {
		return subscriptionStore.userSubscription?.price_id === priceId;
	}

	function getButtonText(priceId: string): string {
		if (checkoutLoading === priceId) return 'Processing...';
		if (isCurrentPlan(priceId)) return 'Current Plan';
		if (!authStore.isLoggedIn) return 'Sign Up to Subscribe';
		if (subscriptionStore.isSubscribed) return 'Switch Plan';
		return 'Subscribe';
	}

	function isButtonDisabled(priceId: string): boolean {
		return checkoutLoading !== null || isCurrentPlan(priceId);
	}
</script>

<svelte:head>
	<title>Pricing - {config.app.name}</title>
	<meta name="description" content="Choose the perfect plan for your needs" />
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="mx-auto max-w-7xl">
		<div class="text-center mb-10">
			<h1 class="text-3xl md:text-4xl font-bold mb-3">Choose Your Plan</h1>
			<p class="text-lg text-muted-foreground max-w-2xl mx-auto">
				Select the perfect plan for your needs. Cancel or change anytime.
			</p>
		</div>

		{#if subscriptionStore.isLoading}
			<div class="text-center py-8">
				<Loader2 class="h-6 w-6 animate-spin mx-auto mb-3" />
				<p class="text-sm text-muted-foreground">Loading pricing plans...</p>
			</div>
		{:else}
			<!-- Subscription Plans -->
			<div class="grid gap-6 md:grid-cols-3 mb-12">
				{#each getProductsWithPrices().filter(p => p.metadata?.category === 'subscription') as product (product.id)}
					{@const monthlyPrice = product.prices.find(p => p.interval === 'month')}
					{@const yearlyPrice = product.prices.find(p => p.interval === 'year')}
					{@const isProfessional = product.metadata?.tier === 'professional'}
					
					<PricingCard
						{product}
						{monthlyPrice}
						{yearlyPrice}
						isPopular={isProfessional}
						popularLabel="Most Popular"
						{isCurrentPlan}
						{checkoutLoading}
						{getButtonText}
						{isButtonDisabled}
						onSubscribe={handleSubscribe}
					/>
				{/each}
			</div>

			<!-- One-time Credits Section -->
			{@const creditProducts = getProductsWithPrices().filter(p => p.metadata?.category === 'one_time')}
			{#if creditProducts.length > 0}
				<div class="border-t pt-8">
					<div class="text-center mb-6">
						<h2 class="text-2xl font-bold mb-2">Need More Credits?</h2>
						<p class="text-muted-foreground">One-time credit packages for extra usage</p>
					</div>
					
					<div class="grid gap-4 md:grid-cols-3 max-w-4xl mx-auto">
						{#each creditProducts as product (product.id)}
							{#each product.prices as price (price.id)}
								<CreditCard
									{price}
									{checkoutLoading}
									{isButtonDisabled}
									onPurchase={handleSubscribe}
								/>
							{/each}
						{/each}
					</div>
				</div>
			{/if}

			{#if getProductsWithPrices().length === 0}
				<div class="text-center py-12">
					<p class="text-muted-foreground">No pricing plans available at the moment.</p>
					<p class="text-sm text-muted-foreground mt-2">
						Please check back later or contact support.
					</p>
				</div>
			{/if}
		{/if}

		{#if subscriptionStore.isSubscribed}
			<div class="mt-12 text-center">
				<div class="rounded-lg bg-green-50 border border-green-200 p-6 inline-block">
					<h3 class="text-lg font-semibold text-green-800 mb-2">You're subscribed!</h3>
					<p class="text-green-700 mb-4">
						Manage your subscription, update payment methods, and view billing history.
					</p>
					<a 
						href="/billing" 
						class="inline-flex items-center rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 transition-colors"
					>
						Manage Subscription
					</a>
				</div>
			</div>
		{/if}

		<div class="mt-16 text-center">
			<h3 class="text-lg font-semibold mb-4">Questions?</h3>
			<p class="text-muted-foreground">
				Need help choosing the right plan? 
				<a href="/about" class="text-primary hover:underline">Contact us</a>
				and we'll help you find the perfect fit.
			</p>
		</div>
	</div>
</div>