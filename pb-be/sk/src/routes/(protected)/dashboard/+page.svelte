<script lang="ts">
	import { authStore } from '$lib/stores/authClient.svelte.js';
	import { subscriptionStore } from '$lib/stores/subscription.svelte.js';
	import { config } from '$lib/config.js';
	import { pb } from '$lib/pocketbase.js';
	import { Crown, User, Mail, Calendar, Edit3, Upload, X } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { getAvatarUrl } from '$lib/files.js';
	import PersonalAccount from '$lib/components/dashboard/PersonalAccount.svelte';
	import APIKeyManager from '$lib/components/APIKeyManager.svelte';

	// State for avatar upload
	let showAvatarUploadDialog = $state(false);
	let isUploading = $state(false);
	let isDragOver = $state(false);
	let fileInput: HTMLInputElement;

	// Subscription store is initialized in root layout


	// Helper to format date
	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	// State for error messages
	let uploadError = $state<string | null>(null);

	// Function to detect actual file type from file header and check for animated PNG
	async function detectFileType(file: File): Promise<{ type: string; isAnimated: boolean }> {
		return new Promise((resolve) => {
			const reader = new FileReader();
			reader.onload = (e) => {
				const arr = new Uint8Array(e.target?.result as ArrayBuffer);
				let header = '';
				for (let i = 0; i < Math.min(arr.length, 8); i++) {
					header += arr[i].toString(16).padStart(2, '0');
				}
				
				// Detect file type by header
				if (header.startsWith('89504e47')) {
					// PNG file detected, check if it's animated (APNG)
					// Look for acTL chunk which indicates animation
					const fullArray = new Uint8Array(e.target?.result as ArrayBuffer);
					const isAnimated = checkForAPNG(fullArray);
					resolve({ type: 'image/png', isAnimated });
				} else if (header.startsWith('ffd8ff')) {
					resolve({ type: 'image/jpeg', isAnimated: false });
				} else if (header.startsWith('47494638')) {
					resolve({ type: 'image/gif', isAnimated: true }); // GIFs can be animated
				} else if (header.startsWith('52494646')) {
					resolve({ type: 'image/webp', isAnimated: false });
				} else {
					resolve({ type: 'unknown', isAnimated: false });
				}
			};
			reader.readAsArrayBuffer(file);
		});
	}

	// Function to check if PNG is animated (APNG)
	function checkForAPNG(data: Uint8Array): boolean {
		// Look for acTL chunk signature in PNG file
		// acTL = 61 63 54 4C
		const acTLSignature = [0x61, 0x63, 0x54, 0x4C];
		
		for (let i = 0; i < data.length - 4; i++) {
			if (data[i] === acTLSignature[0] && 
				data[i + 1] === acTLSignature[1] && 
				data[i + 2] === acTLSignature[2] && 
				data[i + 3] === acTLSignature[3]) {
				return true;
			}
		}
		return false;
	}

	// Handle file upload
	async function handleFileUpload(file: File) {
		if (!authStore.user) {
			uploadError = 'User not authenticated';
			return;
		}

		// Clear previous errors
		uploadError = null;

		// Enhanced file validation
		console.log('File details:', {
			name: file.name,
			type: file.type,
			size: file.size,
			lastModified: file.lastModified
		});

		// Check if we have a valid user and auth
		console.log('Auth details:', {
			isLoggedIn: authStore.isLoggedIn,
			userId: authStore.user?.id,
			userEmail: authStore.user?.email,
			authValid: pb.authStore.isValid
		});

		// Check file size
		const maxSize = 5 * 1024 * 1024; // 5MB
		if (file.size > maxSize) {
			uploadError = `File size (${(file.size / 1024 / 1024).toFixed(2)}MB) exceeds 5MB limit`;
			return;
		}

		if (file.size === 0) {
			uploadError = 'File appears to be empty';
			return;
		}

		// Check file type
		const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp', 'image/gif'];
		if (!allowedTypes.includes(file.type)) {
			uploadError = `File type "${file.type}" not supported. Please use JPEG, PNG, WebP, or GIF`;
			return;
		}

		// Additional check for file extension
		const allowedExtensions = ['.jpg', '.jpeg', '.png', '.webp', '.gif'];
		const fileExtension = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
		if (!allowedExtensions.includes(fileExtension)) {
			uploadError = `File extension "${fileExtension}" not supported`;
			return;
		}

		// Verify actual file type by reading file header
		try {
			const fileAnalysis = await detectFileType(file);
			console.log('File analysis:', fileAnalysis, 'vs reported type:', file.type);
			
			if (fileAnalysis.type === 'unknown') {
				uploadError = 'File does not appear to be a valid image';
				return;
			}
			
			// Check for animated PNG (APNG) which might not be supported
			if (fileAnalysis.isAnimated && fileAnalysis.type === 'image/png') {
				uploadError = 'Animated PNG files (APNG) are not supported. Please use a static PNG or convert to GIF for animations.';
				return;
			}
			
			// Check for animated files in general if we want to restrict them
			if (fileAnalysis.isAnimated && fileAnalysis.type === 'image/gif') {
				console.log('Animated GIF detected - this should be supported');
			}
			
			// If detected type doesn't match reported type, log it but continue
			if (fileAnalysis.type !== file.type) {
				console.warn('File type mismatch:', { detected: fileAnalysis.type, reported: file.type });
			}
		} catch (typeDetectionError) {
			console.error('Error detecting file type:', typeDetectionError);
			// Continue anyway - this is just extra validation
		}

		try {
			isUploading = true;

			// Create FormData and upload directly to PocketBase users collection
			const formData = new FormData();
			formData.append('avatar', file);

			console.log('Attempting to upload avatar...', {
				userId: authStore.user.id,
				fileName: file.name,
				fileType: file.type,
				fileSize: file.size
			});

			// Update user record with new avatar
			const result = await pb.collection('users').update(authStore.user.id, formData);
			
			console.log('Upload successful:', result);
			
			// Close dialog and refresh
			showAvatarUploadDialog = false;
			authStore.syncState();
		} catch (error: any) {
			console.error('Failed to upload avatar:', error);
			console.error('Error type:', typeof error);
			console.error('Error keys:', Object.keys(error));
			
			// Log the full error structure for debugging
			if (error?.response) {
				console.error('Error response:', error.response);
				console.error('Response status:', error.response?.status);
				console.error('Response data:', error.response?.data);
			}
			
			// Parse PocketBase error for better user feedback
			if (error?.response?.data) {
				const errorData = error.response.data;
				console.error('PocketBase error details:', errorData);
				
				// Check for field-specific errors
				if (errorData.avatar) {
					const avatarError = errorData.avatar;
					if (typeof avatarError === 'object') {
						uploadError = `Avatar error: ${avatarError.message || avatarError.code || JSON.stringify(avatarError)}`;
					} else {
						uploadError = `Avatar error: ${avatarError}`;
					}
				} else if (errorData.message) {
					uploadError = `Upload failed: ${errorData.message}`;
				} else if (errorData.code) {
					uploadError = `Upload failed (${errorData.code}): ${errorData.message || 'Server error'}`;
				} else {
					// Show the raw error data for debugging
					uploadError = `Upload failed: ${JSON.stringify(errorData)}`;
				}
			} else if (error?.status) {
				uploadError = `HTTP ${error.status}: ${error.message || 'Server error'}`;
			} else if (error?.message) {
				uploadError = `Upload failed: ${error.message}`;
			} else {
				uploadError = `Upload failed: ${JSON.stringify(error)}`;
			}
		} finally {
			isUploading = false;
		}
	}

	// Handle file input change
	function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (file) {
			handleFileUpload(file);
		}
	}

	// Handle drag and drop
	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
		
		const files = event.dataTransfer?.files;
		if (files && files.length > 0) {
			handleFileUpload(files[0]);
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		isDragOver = true;
	}

	function handleDragLeave(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
	}

</script>

<svelte:head>
	<title>Dashboard - {config.app.name}</title>
	<meta name="description" content="User dashboard" />
</svelte:head>

<!-- Hero Section -->
<section class="py-20 px-6">
	<div class="max-w-4xl mx-auto">
		<h1 class="text-4xl md:text-5xl font-bold mb-6">Dashboard</h1>
		<p class="text-xl text-muted-foreground">
			Welcome back, {authStore.user?.name || 'User'}
		</p>
	</div>
</section>

<!-- Dashboard Content -->
<section class="py-20 border-t px-6">
	<div class="max-w-6xl mx-auto">

		<div class="grid gap-8 lg:grid-cols-3">
		<!-- Profile Section -->
		<div class="lg:col-span-1">
				<div class="bg-card rounded-xl border border-border p-6 shadow-sm">
					<div class="text-center">
						<!-- Avatar Section - Hidden for now -->
						<!-- <div class="relative mb-6">
							<div class="relative inline-block">
								{#if getAvatarUrl(authStore.user, 'large')}
									<img
										src={getAvatarUrl(authStore.user, 'large')}
										alt="Profile"
										class="w-24 h-24 rounded-full object-cover border-4 border-background shadow-lg"
									/>
								{:else}
									<div class="w-24 h-24 rounded-full bg-muted border-4 border-background shadow-lg flex items-center justify-center">
										<User class="w-8 h-8 text-muted-foreground" />
									</div>
								{/if}
								
								<button
									onclick={() => {
										uploadError = null;
										showAvatarUploadDialog = true;
									}}
									class="absolute -bottom-1 -right-1 w-8 h-8 bg-primary text-primary-foreground rounded-full shadow-lg hover:bg-primary/90 transition-colors flex items-center justify-center"
									title="Upload avatar"
								>
									<Edit3 class="w-4 h-4" />
								</button>
							</div>
						</div> -->


						<!-- User Name & Email -->
						<div class="space-y-2">
							<h2 class="text-xl font-semibold text-foreground">
								{authStore.user?.name || 'User'}
							</h2>
							<p class="text-muted-foreground flex items-center justify-center gap-2">
								<Mail class="w-4 h-4" />
								{authStore.user?.email}
							</p>
							{#if authStore.user?.created}
								<p class="text-sm text-muted-foreground flex items-center justify-center gap-2">
									<Calendar class="w-4 h-4" />
									Member since {formatDate(authStore.user.created)}
								</p>
							{/if}
						</div>

						<!-- Subscription Status -->
						{#if subscriptionStore.isSubscribed}
							<div class="mt-4 inline-flex items-center gap-2 px-3 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 rounded-full text-sm font-medium">
								<Crown class="w-4 h-4" />
								Premium Member
							</div>
						{:else}
							<div class="mt-4">
								<a
									href="/pricing"
									class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors text-sm font-medium"
								>
									<Crown class="w-4 h-4" />
									Upgrade to Premium
								</a>
							</div>
						{/if}
					</div>
				</div>
			</div>

		<!-- Main Content -->
		<div class="lg:col-span-2 space-y-6">
			<!-- Quick Actions -->
				<div class="bg-card rounded-xl border border-border p-6 shadow-sm">
					<h3 class="text-lg font-semibold text-foreground mb-4">Quick Actions</h3>
					<div class="grid gap-4 sm:grid-cols-1">
						{#if subscriptionStore.isSubscribed}
							<a
								href="/billing"
								class="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-950/50 rounded-lg border border-green-200 dark:border-green-800/50 hover:bg-green-100 dark:hover:bg-green-950/70 transition-colors group"
							>
								<div class="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center">
									<Crown class="w-5 h-5 text-white" />
								</div>
								<div>
									<h4 class="font-medium text-green-900 dark:text-green-100">Manage Billing</h4>
									<p class="text-sm text-green-700 dark:text-green-300">View subscription</p>
								</div>
							</a>
						{:else}
							<a
								href="/pricing"
								class="flex items-center gap-3 p-4 bg-purple-50 dark:bg-purple-950/50 rounded-lg border border-purple-200 dark:border-purple-800/50 hover:bg-purple-100 dark:hover:bg-purple-950/70 transition-colors group"
							>
								<div class="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center">
									<Crown class="w-5 h-5 text-white" />
								</div>
								<div>
									<h4 class="font-medium text-purple-900 dark:text-purple-100">Go Premium</h4>
									<p class="text-sm text-purple-700 dark:text-purple-300">Unlock all features</p>
								</div>
							</a>
						{/if}
					</div>
				</div>


			<!-- Personal Account -->
				<PersonalAccount />

			<!-- API Key Management -->
				<APIKeyManager />

			</div>
		</div>
	</div>
</section>

<!-- Avatar Upload Dialog -->
{#if showAvatarUploadDialog}
	<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={(e) => e.target === e.currentTarget && !isUploading && (showAvatarUploadDialog = false)}>
		<div class="bg-background rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-lg font-semibold">Upload Avatar</h3>
				{#if !isUploading}
					<button
						onclick={() => showAvatarUploadDialog = false}
						class="text-muted-foreground hover:text-foreground"
					>
						<X class="h-5 w-5" />
					</button>
				{/if}
			</div>
			
			<!-- Error Message -->
			{#if uploadError}
				<div class="mb-4 p-3 bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800/50 rounded-lg">
					<p class="text-sm text-red-700 dark:text-red-300 mb-2">{uploadError}</p>
					<button
						onclick={() => uploadError = null}
						class="text-xs text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-200 underline"
					>
						Try again
					</button>
				</div>
			{/if}

			<!-- Upload Area -->
			<div 
				class="border-2 border-dashed rounded-lg p-8 text-center transition-colors {isDragOver ? 'border-primary bg-primary/5' : uploadError ? 'border-red-300 bg-red-50/50 dark:border-red-700 dark:bg-red-950/20' : 'border-muted-foreground/20'}"
				ondrop={handleDrop}
				ondragover={handleDragOver}
				ondragleave={handleDragLeave}
			>
				{#if isUploading}
					<div class="space-y-4">
						<div class="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent mx-auto"></div>
						<p class="text-sm text-muted-foreground">Uploading avatar...</p>
					</div>
				{:else}
					<div class="space-y-4">
						<div class="w-12 h-12 bg-muted rounded-full flex items-center justify-center mx-auto">
							<Upload class="h-6 w-6 text-muted-foreground" />
						</div>
						<div>
							<p class="text-sm font-medium mb-1">Drop your image here, or click to browse</p>
							<p class="text-xs text-muted-foreground">JPEG, PNG (static), WebP or GIF up to 5MB</p>
							<p class="text-xs text-muted-foreground mt-1">✨ Upload starts automatically when file is selected</p>
							<p class="text-xs text-muted-foreground">Note: Animated PNG files (APNG) are not supported</p>
						</div>
						<button
							onclick={() => fileInput.click()}
							class="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors text-sm"
							disabled={uploadError !== null}
						>
							Choose File
						</button>
					</div>
				{/if}
			</div>

			<!-- Hidden file input -->
			<input
				bind:this={fileInput}
				type="file"
				accept="image/jpeg,image/jpg,image/png,image/webp,image/gif"
				onchange={handleFileChange}
				class="hidden"
				disabled={isUploading}
			/>

		</div>
	</div>
{/if}
