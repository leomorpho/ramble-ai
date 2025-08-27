<script lang="ts">
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Edit3 } from 'lucide-svelte';
	import { pb } from '$lib/pocketbase.js';
	import { subscriptionStore } from '$lib/stores/subscription.svelte.js';

	interface Props {
		companyData: any;
		employeeData: any;
	}

	let { companyData = $bindable(), employeeData }: Props = $props();

	// Company editing state
	let isEditingCompany = $state(false);
	let editCompanyName = $state(companyData?.name || '');
	let editCompanyDomain = $state(companyData?.domain || '');
	let isSavingCompany = $state(false);
	let companyEditError = $state<string | null>(null);

	// Update edit values when company data changes
	$effect(() => {
		if (companyData) {
			editCompanyName = companyData.name || '';
			editCompanyDomain = companyData.domain || '';
		}
	});

	// Handle company save
	async function handleSaveCompany() {
		if (!editCompanyName.trim()) {
			companyEditError = 'Company name is required';
			return;
		}

		isSavingCompany = true;
		companyEditError = null;

		try {
			// Update company data
			const updatedCompany = await pb.collection('companies').update(companyData.id, {
				name: editCompanyName.trim(),
				domain: editCompanyDomain.trim() || null
			});

			// Update local state
			companyData = updatedCompany;
			isEditingCompany = false;
		} catch (error: any) {
			console.error('Failed to update company:', error);
			companyEditError = error.message || 'Failed to update company';
		} finally {
			isSavingCompany = false;
		}
	}

	// Handle cancel editing
	function handleCancelCompanyEdit() {
		isEditingCompany = false;
		editCompanyName = companyData?.name || '';
		editCompanyDomain = companyData?.domain || '';
		companyEditError = null;
	}
</script>

<div class="bg-card rounded-xl border border-border p-6 shadow-sm">
	<div class="flex items-center justify-between mb-4">
		<h3 class="text-lg font-semibold text-foreground">Company Overview</h3>
		{#if employeeData?.role === 'owner'}
			<button
				onclick={() => isEditingCompany = !isEditingCompany}
				class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
				disabled={isSavingCompany}
			>
				<Edit3 class="w-4 h-4" />
				Edit
			</button>
		{/if}
	</div>

	{#if isEditingCompany}
		<!-- Edit Form -->
		<form onsubmit={(e) => { e.preventDefault(); handleSaveCompany(); }} class="space-y-4">
			{#if companyEditError}
				<div class="p-3 bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800 rounded-lg">
					<p class="text-sm text-red-600 dark:text-red-400">{companyEditError}</p>
				</div>
			{/if}

			<div class="space-y-2">
				<Label for="edit-company-name">Company Name</Label>
				<Input
					id="edit-company-name"
					type="text"
					bind:value={editCompanyName}
					disabled={isSavingCompany}
					required
				/>
			</div>

			<div class="space-y-2">
				<Label for="edit-company-domain">Company Domain</Label>
				<Input
					id="edit-company-domain"
					type="text"
					bind:value={editCompanyDomain}
					placeholder="example.com"
					disabled={isSavingCompany}
				/>
			</div>

			<div class="flex gap-2">
				<Button
					type="submit"
					size="sm"
					disabled={isSavingCompany}
				>
					{isSavingCompany ? 'Saving...' : 'Save Changes'}
				</Button>
				<Button
					type="button"
					variant="outline"
					size="sm"
					onclick={handleCancelCompanyEdit}
					disabled={isSavingCompany}
				>
					Cancel
				</Button>
			</div>
		</form>
	{:else}
		<!-- Display Mode -->
		<div class="grid gap-4 sm:grid-cols-2">
			<div class="p-4 bg-muted/50 rounded-lg">
				<h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-1">Company Name</h4>
				<p class="text-foreground">{companyData?.name}</p>
			</div>
			
			<div class="p-4 bg-muted/50 rounded-lg">
				<h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-1">Your Role</h4>
				<p class="text-foreground capitalize">{employeeData?.role || 'Member'}</p>
			</div>
			
			{#if companyData?.domain}
				<div class="p-4 bg-muted/50 rounded-lg">
					<h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-1">Domain</h4>
					<p class="text-foreground">{companyData.domain}</p>
				</div>
			{/if}
			
			<div class="p-4 bg-muted/50 rounded-lg">
				<h4 class="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-1">Plan Status</h4>
				<p class="text-foreground">
					{subscriptionStore.isSubscribed ? 'Premium' : 'Free'}
				</p>
			</div>
		</div>
	{/if}
</div>