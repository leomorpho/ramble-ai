<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { pb } from '$lib/pocketbase.js';
	import { authStore } from '$lib/stores/authClient.svelte.js';
	import { Building2, ArrowRight } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	interface Props {
		onSuccess?: () => void;
	}

	let { onSuccess }: Props = $props();

	// Form state
	let companyName = $state('');
	let companyDomain = $state('');
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	async function handleSubmit(event: Event) {
		event.preventDefault();
		
		if (!companyName.trim()) {
			error = 'Company name is required';
			return;
		}

		if (!authStore.user) {
			error = 'You must be logged in to create a company';
			return;
		}

		isLoading = true;
		error = null;

		try {
			// Create the company
			const company = await pb.collection('companies').create({
				name: companyName.trim(),
				domain: companyDomain.trim() || null,
				owner_id: authStore.user.id
			});

			// Create the employee record (owner)
			await pb.collection('employees').create({
				user_id: authStore.user.id,
				company_id: company.id,
				role: 'owner',
				joined_at: new Date().toISOString()
			});

			// Success! Call the onSuccess callback or redirect
			if (onSuccess) {
				onSuccess();
			} else {
				// Refresh the page to load company data
				window.location.reload();
			}
		} catch (err: any) {
			console.error('Error creating company:', err);
			error = err.message || 'Failed to create company. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="w-full max-w-md mx-auto">
	<div class="text-center mb-8">
		<div class="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
			<Building2 class="w-8 h-8 text-primary" />
		</div>
		<h2 class="text-2xl font-bold text-foreground mb-2">Set up your company</h2>
		<p class="text-muted-foreground">Let's get your organization ready to go</p>
	</div>

	<form onsubmit={handleSubmit} class="space-y-6">
		{#if error}
			<div class="p-3 bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800 rounded-lg">
				<p class="text-sm text-red-600 dark:text-red-400">{error}</p>
			</div>
		{/if}

		<div class="space-y-2">
			<Label for="company-name">Company Name *</Label>
			<Input
				id="company-name"
				type="text"
				placeholder="Acme Corporation"
				bind:value={companyName}
				disabled={isLoading}
				required
			/>
			<p class="text-xs text-muted-foreground">This is how your company will appear throughout the platform</p>
		</div>

		<div class="space-y-2">
			<Label for="company-domain">Company Domain (optional)</Label>
			<Input
				id="company-domain"
				type="text"
				placeholder="acme.com"
				bind:value={companyDomain}
				disabled={isLoading}
			/>
			<p class="text-xs text-muted-foreground">Your company's website or domain</p>
		</div>

		<Button type="submit" class="w-full" disabled={isLoading}>
			{#if isLoading}
				<span class="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full mr-2"></span>
				Creating company...
			{:else}
				Continue
				<ArrowRight class="w-4 h-4 ml-2" />
			{/if}
		</Button>

		<p class="text-xs text-center text-muted-foreground">
			You'll be set as the company owner and can manage settings later
		</p>
	</form>
</div>