<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogHeader,
		DialogTitle
	} from '$lib/components/ui/dialog';
	import emailjs from '@emailjs/browser';

	let {
		triggerText = 'Start Project',
		variant = 'default',
		size = 'default'
	}: {
		triggerText?: string;
		variant?: 'default' | 'outline' | 'destructive' | 'secondary' | 'ghost' | 'link';
		size?: 'default' | 'sm' | 'lg' | 'icon';
	} = $props();

	let open = $state(false);
	
	let isSubmitting = $state(false);
	let submitted = $state(false);

	// Form data with localStorage persistence
	const STORAGE_KEY = 'goosebytes-project-intake-form';
	
	// Initialize form data with default values
	const defaultFormData = {
		// Contact Information
		name: '',
		email: '',
		company: '',
		phone: '',
		
		// Project Details
		projectType: '',
		projectSize: '',
		budget: '',
		timeline: '',
		description: '',
		
		// Technical Requirements
		platforms: '',
		existingTech: '',
		integrations: '',
		
		// Additional Information
		inspiration: '',
		additionalNotes: ''
	};

	// Load saved data from localStorage and merge with defaults
	function getInitialFormData() {
		if (typeof window !== 'undefined') {
			const saved = localStorage.getItem(STORAGE_KEY);
			if (saved) {
				try {
					const savedData = JSON.parse(saved);
					return { ...defaultFormData, ...savedData };
				} catch (error) {
					console.warn('Failed to load saved form data:', error);
				}
			}
		}
		return defaultFormData;
	}

	// Initialize form data with saved data or defaults
	let formData = $state(getInitialFormData());

	// Save data to localStorage
	function saveFormData() {
		if (typeof window !== 'undefined') {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(formData));
		}
	}

	// Clear saved data
	function clearSavedData() {
		if (typeof window !== 'undefined') {
			localStorage.removeItem(STORAGE_KEY);
		}
	}

	// Define options arrays first (before derived values that use them)
	const projectTypes = [
		{ value: 'ai-solutions', label: 'AI Solutions' },
		{ value: 'apps', label: 'Apps' },
		{ value: 'automation', label: 'Automation' },
		{ value: 'web-design', label: 'Web Design' },
		{ value: 'web-development', label: 'Web Development' }
	];

	const projectSizes = [
		{ value: 'small', label: 'Small (1-3 months)' },
		{ value: 'medium', label: 'Medium (3-6 months)' },
		{ value: 'large', label: 'Large (6-12 months)' },
		{ value: 'enterprise', label: 'Enterprise (12+ months)' }
	];

	const budgetRanges = [
		{ value: 'under-25k', label: 'Under $25,000' },
		{ value: '25k-50k', label: '$25,000 - $50,000' },
		{ value: '50k-100k', label: '$50,000 - $100,000' },
		{ value: '100k-250k', label: '$100,000 - $250,000' },
		{ value: '250k-500k', label: '$250,000 - $500,000' },
		{ value: 'over-500k', label: 'Over $500,000' },
		{ value: 'discuss', label: 'Let\'s discuss' }
	];

	const timelines = [
		{ value: 'asap', label: 'ASAP' },
		{ value: '1-month', label: 'Within 1 month' },
		{ value: '3-months', label: 'Within 3 months' },
		{ value: '6-months', label: 'Within 6 months' },
		{ value: 'flexible', label: 'Flexible timeline' }
	];

	// Auto-save effect to save data whenever formData changes
	$effect(() => {
		if (formData) {
			saveFormData();
		}
	});

	// Derived values for select trigger content (after arrays are defined)
	const projectTypeLabel = $derived(
		projectTypes.find((t) => t.value === formData.projectType)?.label ?? "Select project type"
	);
	
	const projectSizeLabel = $derived(
		projectSizes.find((s) => s.value === formData.projectSize)?.label ?? "Select project size"
	);
	
	const budgetLabel = $derived(
		budgetRanges.find((b) => b.value === formData.budget)?.label ?? "Select budget range"
	);
	
	const timelineLabel = $derived(
		timelines.find((t) => t.value === formData.timeline)?.label ?? "Select timeline"
	);

	async function handleSubmit(event: Event) {
		event.preventDefault();
		isSubmitting = true;
		
		try {
			// Send email using EmailJS (you'll need to set this up)
			// For now, we'll use a placeholder implementation
			
			// Format the email content
			const emailData = {
				to_email: import.meta.env.VITE_CONTACT_EMAIL || 'hello@goosebyteshq.com',
				from_name: formData.name,
				from_email: formData.email,
				company: formData.company || 'Not provided',
				phone: formData.phone || 'Not provided',
				project_type: projectTypes.find(t => t.value === formData.projectType)?.label || formData.projectType,
				project_size: projectSizes.find(s => s.value === formData.projectSize)?.label || formData.projectSize,
				budget: budgetRanges.find(b => b.value === formData.budget)?.label || formData.budget,
				timeline: timelines.find(t => t.value === formData.timeline)?.label || formData.timeline,
				description: formData.description,
				platforms: formData.platforms || 'Not specified',
				existing_tech: formData.existingTech || 'Not specified',
				integrations: formData.integrations || 'Not specified',
				inspiration: formData.inspiration || 'Not provided',
				additional_notes: formData.additionalNotes || 'None',
				submission_date: new Date().toISOString()
			};

			// Send email using EmailJS or fallback to mailto
			const EMAILJS_SERVICE_ID = import.meta.env.VITE_EMAILJS_SERVICE_ID || 'YOUR_SERVICE_ID';
			const EMAILJS_TEMPLATE_ID = import.meta.env.VITE_EMAILJS_TEMPLATE_ID || 'YOUR_TEMPLATE_ID';
			const EMAILJS_USER_ID = import.meta.env.VITE_EMAILJS_USER_ID || 'YOUR_USER_ID';

			// Send email using EmailJS
			await emailjs.send(
				EMAILJS_SERVICE_ID,
				EMAILJS_TEMPLATE_ID,
				emailData,
				EMAILJS_USER_ID
			);

			submitted = true;
			clearSavedData(); // Clear localStorage after successful submission
			
			setTimeout(() => {
				open = false;
				submitted = false;
				// Reset form
				formData = { ...defaultFormData };
			}, 2000);
		} catch (error) {
			console.error('Failed to submit form:', error);
			
			// Provide specific error messages based on the error type
			let errorMessage = 'Failed to submit form. Please try again or contact us directly.';
			
			if (error instanceof Error) {
				if (error.message.includes('EmailJS')) {
					errorMessage = `Email service is not properly configured. Please contact us directly at ${import.meta.env.VITE_CONTACT_EMAIL || 'hello@goosebyteshq.com'}`;
				} else if (error.message.includes('network') || error.message.includes('fetch')) {
					errorMessage = 'Network error. Please check your connection and try again.';
				}
			}
			
			alert(errorMessage);
		} finally {
			isSubmitting = false;
		}
	}

</script>

<!-- Use shadcn-svelte Button component -->
<Button 
	{variant}
	{size}
	onclick={() => {
		open = true;
	}}
>
	{triggerText}
</Button>

<Dialog bind:open>
	<DialogContent class="max-w-4xl max-h-[90vh] overflow-y-auto sm:max-w-4xl sm:max-h-[90vh] max-sm:max-w-full max-sm:max-h-full max-sm:h-screen max-sm:w-screen rounded-lg custom-scrollbar">
		<DialogHeader>
			<DialogTitle class="text-2xl">Let's Build Something Amazing Together</DialogTitle>
			<DialogDescription class="text-base">
				Tell us about your project so we can provide the best solution for your needs.
			</DialogDescription>
			<div class="text-sm text-muted-foreground bg-muted/20 p-2 rounded flex items-center justify-between">
				<span>💾 Your progress is automatically saved locally</span>
				<button 
					type="button" 
					class="text-xs underline hover:no-underline"
					onclick={() => {
						clearSavedData();
						formData = { ...defaultFormData };
					}}
				>
					Clear saved data
				</button>
			</div>
		</DialogHeader>

		{#if submitted}
			<div class="text-center py-12">
				<div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
					</svg>
				</div>
				<h3 class="text-xl font-semibold text-foreground mb-2">Thank You!</h3>
				<p class="text-muted-foreground">
					We've received your project details and will get back to you within 24 hours.
				</p>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="space-y-8">
				<!-- Contact Information -->
				<Card>
					<CardHeader>
						<CardTitle>Contact Information</CardTitle>
					</CardHeader>
					<CardContent class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div class="space-y-2">
							<Label for="name">Full Name *</Label>
							<Input id="name" bind:value={formData.name} required />
						</div>
						<div class="space-y-2">
							<Label for="email">Email Address *</Label>
							<Input id="email" type="email" bind:value={formData.email} required />
						</div>
						<div class="space-y-2">
							<Label for="company">Company/Organization</Label>
							<Input id="company" bind:value={formData.company} />
						</div>
						<div class="space-y-2">
							<Label for="phone">Phone Number</Label>
							<Input id="phone" type="tel" bind:value={formData.phone} />
						</div>
					</CardContent>
				</Card>

				<!-- Project Overview -->
				<Card>
					<CardHeader>
						<CardTitle>Project Overview</CardTitle>
					</CardHeader>
					<CardContent class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label for="projectType">What type of project is this? *</Label>
								<Select.Root type="single" name="projectType" bind:value={formData.projectType}>
									<Select.Trigger class="w-full">
										{projectTypeLabel}
									</Select.Trigger>
									<Select.Content>
										{#each projectTypes as type (type.value)}
											<Select.Item value={type.value} label={type.label}>
												{type.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
							<div class="space-y-2">
								<Label for="projectSize">Project Size *</Label>
								<Select.Root type="single" name="projectSize" bind:value={formData.projectSize}>
									<Select.Trigger class="w-full">
										{projectSizeLabel}
									</Select.Trigger>
									<Select.Content>
										{#each projectSizes as size (size.value)}
											<Select.Item value={size.value} label={size.label}>
												{size.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
						</div>
						
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label for="budget">Budget Range *</Label>
								<Select.Root type="single" name="budget" bind:value={formData.budget}>
									<Select.Trigger class="w-full">
										{budgetLabel}
									</Select.Trigger>
									<Select.Content>
										{#each budgetRanges as budget (budget.value)}
											<Select.Item value={budget.value} label={budget.label}>
												{budget.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
							<div class="space-y-2">
								<Label for="timeline">Desired Timeline *</Label>
								<Select.Root type="single" name="timeline" bind:value={formData.timeline}>
									<Select.Trigger class="w-full">
										{timelineLabel}
									</Select.Trigger>
									<Select.Content>
										{#each timelines as timeline (timeline.value)}
											<Select.Item value={timeline.value} label={timeline.label}>
												{timeline.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
						</div>

						<div class="space-y-2">
							<Label for="description">Project Description *</Label>
							<Textarea 
								id="description" 
								bind:value={formData.description}
								placeholder="Describe your project, goals, target audience, and key features you envision..."
								rows={4}
								required
							/>
						</div>
					</CardContent>
				</Card>

				<!-- Technical Requirements -->
				<Card>
					<CardHeader>
						<CardTitle>Technical Requirements</CardTitle>
					</CardHeader>
					<CardContent class="space-y-4">
						<div class="space-y-2">
							<Label for="platforms">Target Platforms</Label>
							<Input 
								id="platforms" 
								bind:value={formData.platforms}
								placeholder="e.g., Web, iOS, Android, Desktop"
							/>
						</div>
						<div class="space-y-2">
							<Label for="existingTech">Existing Technology Stack</Label>
							<Input 
								id="existingTech" 
								bind:value={formData.existingTech}
								placeholder="e.g., React, Node.js, PostgreSQL, AWS"
							/>
						</div>
						<div class="space-y-2">
							<Label for="integrations">Required Integrations</Label>
							<Input 
								id="integrations" 
								bind:value={formData.integrations}
								placeholder="e.g., Payment processing, CRM, Analytics, APIs"
							/>
						</div>
					</CardContent>
				</Card>

				<!-- Additional Information -->
				<Card>
					<CardHeader>
						<CardTitle>Additional Information</CardTitle>
					</CardHeader>
					<CardContent class="space-y-4">
						<div class="space-y-2">
							<Label for="inspiration">Inspiration/Reference Sites</Label>
							<Input 
								id="inspiration" 
								bind:value={formData.inspiration}
								placeholder="URLs of sites or apps you admire"
							/>
						</div>
						<div class="space-y-2">
							<Label for="additionalNotes">Additional Notes</Label>
							<Textarea 
								id="additionalNotes" 
								bind:value={formData.additionalNotes}
								placeholder="Anything else you'd like us to know?"
								rows={3}
							/>
						</div>
					</CardContent>
				</Card>

				<div class="flex gap-4 pt-4">
					<Button type="button" variant="outline" onclick={() => open = false} class="flex-1">
						Cancel
					</Button>
					<Button 
						type="submit" 
						disabled={isSubmitting || !formData.name || !formData.email || !formData.projectType || !formData.projectSize || !formData.budget || !formData.timeline || !formData.description}
						class="flex-1"
					>
						{#if isSubmitting}
							<svg class="animate-spin -ml-1 mr-3 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Submitting...
						{:else}
							Submit Project Request
						{/if}
					</Button>
				</div>
			</form>
		{/if}
	</DialogContent>
</Dialog>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 8px;
	}

	.custom-scrollbar::-webkit-scrollbar-track {
		background: hsl(var(--muted));
		border-radius: 4px 0 0 4px;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: hsl(var(--muted-foreground) / 0.3);
		border-radius: 4px 0 0 4px;
		transition: background 0.2s ease;
	}

	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: hsl(var(--muted-foreground) / 0.5);
	}

	/* Firefox */
	.custom-scrollbar {
		scrollbar-width: thin;
		scrollbar-color: hsl(var(--muted-foreground) / 0.3) hsl(var(--muted));
	}
</style>