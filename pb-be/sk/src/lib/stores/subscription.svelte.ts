import { browser } from '$app/environment';
import { pb } from '$lib/pocketbase.js';
import { authStore } from './authClient.svelte.js';

interface Product {
	id: string;
	product_id: string;
	active: boolean;
	name: string;
	description: string;
	image?: string;
	metadata?: any;
	product_order?: number;
}

interface Price {
	id: string;
	price_id: string;
	product_id: string;
	active: boolean;
	currency: string;
	unit_amount: number;
	type: string;
	interval?: string;
	interval_count?: number;
	trial_period_days?: number;
	metadata?: any;
}

interface Subscription {
	id: string;
	subscription_id: string;
	user_id: string;
	status: string;
	price_id: string;
	quantity: number;
	cancel_at_period_end: boolean;
	current_period_start: number;
	current_period_end: number;
	ended_at?: number;
	cancel_at?: number;
	canceled_at?: number;
	trial_start?: number;
	trial_end?: number;
	metadata?: any;
}

class SubscriptionStore {
	#products = $state<Product[]>([]);
	#prices = $state<Price[]>([]);
	#userSubscription = $state<Subscription | null>(null);
	#isLoading = $state(false);

	constructor() {
		// Initialize will be called from components
	}

	// Initialize the store with effect tracking - call this from components
	initialize() {
		if (browser) {
			// Load products and prices immediately (public data)
			this.loadData();
			
			// Watch auth state changes to load/clear user-specific data
			$effect(() => {
				if (authStore.isLoggedIn) {
					this.loadUserSubscription();
				} else {
					this.#userSubscription = null;
				}
			});
		}
	}

	get products() {
		return this.#products;
	}

	get prices() {
		return this.#prices;
	}

	get userSubscription() {
		return this.#userSubscription;
	}

	get isLoading() {
		return this.#isLoading;
	}

	get isSubscribed() {
		return this.#userSubscription?.status === 'active' || 
		       this.#userSubscription?.status === 'trialing';
	}

	get subscriptionStatus() {
		return this.#userSubscription?.status || 'none';
	}

	async loadData() {
		if (!browser) return;

		this.#isLoading = true;

		try {
			// Load products and prices in parallel (public data - no auth required)
			const [productsResult, pricesResult] = await Promise.all([
				pb.collection('products').getFullList({
					filter: 'active = true',
					sort: 'product_order,name'
				}).catch(() => []), // Return empty array if collection doesn't exist
				pb.collection('prices').getFullList({
					filter: 'active = true',
					sort: 'unit_amount'
				}).catch(() => []) // Return empty array if collection doesn't exist
			]);

			this.#products = productsResult as Product[];
			this.#prices = pricesResult as Price[];

			// Load user subscription only if logged in
			if (authStore.isLoggedIn) {
				await this.loadUserSubscription();
			} else {
				this.#userSubscription = null;
			}
		} catch (error) {
			console.debug('Subscription collections not available yet.');
			// Set empty state
			this.#products = [];
			this.#prices = [];
			this.#userSubscription = null;
		} finally {
			this.#isLoading = false;
		}
	}

	async loadUserSubscription() {
		if (!browser || !authStore.user) return;

		try {
			// Find the subscription directly for the user
			const subscription = await pb.collection('subscriptions').getFirstListItem(
				`user_id = "${authStore.user.id}" && (status = "active" || status = "trialing")`
			);
			this.#userSubscription = subscription as Subscription;
		} catch (error: any) {
			// No active subscription found or collection doesn't exist
			this.#userSubscription = null;
			if (error?.status === 404) {
				if (error?.message?.includes('Missing collection')) {
					// Silently handle missing collection - it's expected during development
					console.debug('Subscriptions collection not available yet.');
				} else {
					console.debug('No active subscription found for user');
				}
			} else {
				console.warn('Error loading subscription:', error?.message || error);
			}
		}
	}

	// Get prices for a specific product
	getPricesForProduct(productId: string): Price[] {
		return this.#prices.filter(price => price.product_id === productId);
	}

	// Get product by ID
	getProduct(productId: string): Product | undefined {
		return this.#products.find(product => product.product_id === productId);
	}

	// Get price by ID
	getPrice(priceId: string): Price | undefined {
		return this.#prices.find(price => price.price_id === priceId);
	}

	// Check if user has access to a specific feature/product
	hasAccess(productId?: string): boolean {
		if (!this.isSubscribed) return false;
		
		if (!productId) return true; // Any active subscription
		
		const userPrice = this.getPrice(this.#userSubscription?.price_id || '');
		return userPrice?.product_id === productId;
	}

	// Refresh subscription data
	async refresh() {
		await this.loadData();
	}
}

export const subscriptionStore = new SubscriptionStore();