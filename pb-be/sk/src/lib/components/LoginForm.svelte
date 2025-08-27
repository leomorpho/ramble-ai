<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import { authStore } from '$lib/stores/authClient.svelte';
	import { goto } from '$app/navigation';
	import { webauthnLogin } from '$lib/pocketbase';
	import { Key, ArrowLeft, Eye, EyeOff } from 'lucide-svelte';
	import PasskeyRegistration from './PasskeyRegistration.svelte';
	import SignupForm from './SignupForm.svelte';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();

	let activeTab = $state('login');
	
	// Progressive login flow state
	type LoginStep = 'email' | 'method-select' | 'password' | 'passkey';
	let currentStep = $state<LoginStep>('email');
	
	// Form state
	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
	let isLoading = $state(false);
	let error = $state<string | null>(null);
	
	// User state
	let hasPasskeys = $state(false);
	let webAuthnSupported = $state(false);
	
	// Initialize WebAuthn support check on mount
	$effect(() => {
		webAuthnSupported = typeof navigator !== 'undefined' && 
		   !!(navigator.credentials?.create && navigator.credentials?.get);
	});

	// Generate unique ID for form fields
	let formId = $state(Math.random().toString(36).substr(2, 9));

	async function checkUserPasskeys(emailAddress: string): Promise<boolean> {
		try {
			const baseURL = 'http://localhost:8090'; // Use same as pocketbase.ts
			const response = await fetch(`${baseURL}/api/webauthn/login-options?usernameOrEmail=${encodeURIComponent(emailAddress)}`);
			return response.ok;
		} catch {
			return false;
		}
	}

	async function handleEmailContinue(emailAddress: string) {
		email = emailAddress;
		isLoading = true;
		error = null;

		try {
			// Check if user has passkeys registered
			hasPasskeys = await checkUserPasskeys(emailAddress);
			currentStep = 'method-select';
		} catch (err) {
			error = 'Failed to check user information. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	function handleMethodSelect(method: 'password' | 'passkey') {
		error = null;
		currentStep = method;
		
		// Auto-trigger passkey authentication
		if (method === 'passkey') {
			handlePasskeyLogin();
		}
	}

	function handleBackToEmail() {
		currentStep = 'email';
		error = null;
	}

	function handleBackToMethodSelect() {
		currentStep = 'method-select';
		error = null;
	}

	async function handlePasswordSignIn(emailAddress: string, userPassword: string) {
		isLoading = true;
		error = null;

		try {
			const result = await authStore.login(emailAddress, userPassword);
			
			if (result.success) {
				goto('/dashboard');
			} else {
				error = result.error || 'Invalid email or password';
			}
		} catch (err) {
			error = 'Login failed. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	async function handlePasskeyLogin() {
		isLoading = true;
		error = null;

		try {
			await webauthnLogin(email);
			// webauthnLogin handles saving to authStore internally
			goto('/dashboard');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Passkey authentication failed';
			console.error('Passkey login error:', err);
		} finally {
			isLoading = false;
		}
	}

	function validateEmail(email: string): boolean {
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		return emailRegex.test(email);
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}
</script>

<div class={cn('flex flex-col gap-6', className)} bind:this={ref} {...restProps}>
	<!-- App branding -->
	<div class="flex flex-col items-center space-y-2">
		<a href="/" class="flex items-center gap-2 font-medium">
			<div class="flex size-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-4">
					<rect width="7" height="18" x="3" y="3" rx="1"></rect>
					<rect width="7" height="7" x="14" y="3" rx="1"></rect>
					<rect width="7" height="7" x="14" y="14" rx="1"></rect>
				</svg>
			</div>
			<span class="sr-only">App Name</span>
		</a>
	</div>

	<Tabs bind:value={activeTab} class="w-full">
		<TabsList class="grid w-full grid-cols-2">
			<TabsTrigger value="login">Sign In</TabsTrigger>
			<TabsTrigger value="register">Create Account</TabsTrigger>
		</TabsList>
		<TabsContent value="login">
			<div class="flex flex-col gap-6">
				{#if currentStep === 'email'}
					<!-- Email Step -->
					<div class="flex flex-col items-center gap-3 text-center">
						<h1 class="text-2xl font-semibold tracking-tight">Welcome back</h1>
						<p class="text-sm text-muted-foreground">
							Enter your email to continue to your account
						</p>
					</div>

					{#if error}
						<div class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-600">
							{error}
						</div>
					{/if}

					<form onsubmit={(e) => { e.preventDefault(); if (email && validateEmail(email)) handleEmailContinue(email); }} class="flex flex-col gap-4">
						<div class="grid gap-2">
							<Label for="email-{formId}">Email</Label>
							<Input
								id="email-{formId}"
								type="email"
								placeholder="name@example.com"
								bind:value={email}
								disabled={isLoading}
								required
								autocomplete="email"
								class="h-10"
							/>
						</div>

						<Button 
							type="submit" 
							class="w-full h-10" 
							disabled={isLoading || !email || !validateEmail(email)}
						>
							{isLoading ? 'Please wait...' : 'Continue'}
						</Button>
					</form>

				{:else if currentStep === 'method-select'}
					<!-- Method Selection Step -->
					<div class="flex items-center gap-3">
						<Button variant="ghost" size="sm" onclick={handleBackToEmail} disabled={isLoading}>
							<ArrowLeft class="h-4 w-4" />
						</Button>
						<div class="flex-1">
							<h1 class="text-xl font-semibold">How would you like to sign in?</h1>
							<p class="text-sm text-muted-foreground">{email}</p>
						</div>
					</div>

					{#if error}
						<div class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-600">
							{error}
						</div>
					{/if}

					<div class="flex flex-col gap-3">
						{#if hasPasskeys && webAuthnSupported}
							<Button onclick={() => handleMethodSelect('passkey')} class="w-full h-12" disabled={isLoading}>
								<Key class="h-5 w-5 mr-3" />
								<div class="flex flex-col items-start">
									<span class="font-medium">Use your passkey</span>
									<span class="text-xs opacity-75">Touch ID, Face ID, or security key</span>
								</div>
							</Button>
						{/if}
						
						<Button variant="outline" onclick={() => handleMethodSelect('password')} class="w-full h-12" disabled={isLoading}>
							<Key class="h-5 w-5 mr-3" />
							<div class="flex flex-col items-start">
								<span class="font-medium">Use your password</span>
								<span class="text-xs opacity-75">Sign in with your password</span>
							</div>
						</Button>
					</div>

				{:else if currentStep === 'password'}
					<!-- Password Step -->
					<div class="flex items-center gap-3">
						<Button variant="ghost" size="sm" onclick={handleBackToMethodSelect} disabled={isLoading}>
							<ArrowLeft class="h-4 w-4" />
						</Button>
						<div class="flex-1">
							<h1 class="text-xl font-semibold">Enter your password</h1>
							<p class="text-sm text-muted-foreground">{email}</p>
						</div>
					</div>

					{#if error}
						<div class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-600">
							{error}
						</div>
					{/if}

					<form onsubmit={(e) => { e.preventDefault(); if (password) handlePasswordSignIn(email, password); }} class="flex flex-col gap-4">
						<div class="grid gap-2">
							<Label for="password-{formId}">Password</Label>
							<div class="relative">
								<Input
									id="password-{formId}"
									type={showPassword ? "text" : "password"}
									placeholder="Enter your password"
									bind:value={password}
									disabled={isLoading}
									required
									autocomplete="current-password"
									class="h-10 pr-10"
								/>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									class="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
									onclick={togglePasswordVisibility}
									disabled={isLoading}
								>
									{#if showPassword}
										<EyeOff class="h-4 w-4" />
									{:else}
										<Eye class="h-4 w-4" />
									{/if}
								</Button>
							</div>
						</div>

						<Button 
							type="submit" 
							class="w-full h-10" 
							disabled={isLoading || !password}
						>
							{isLoading ? 'Signing in...' : 'Sign in'}
						</Button>
					</form>

					<div class="text-center">
						<a 
							href="/forgot-password" 
							class="text-sm text-muted-foreground hover:text-primary hover:underline"
						>
							Forgot your password?
						</a>
					</div>

					<!-- Show passkey registration option after password login -->
					{#if email && webAuthnSupported && !isLoading}
						<div class="mt-4">
							<div class="text-sm font-medium text-foreground mb-3">Optional: Enhanced Security</div>
							<PasskeyRegistration 
								{email}
								onSuccess={() => {
									console.log('Passkey registered successfully');
								}}
								onError={(error) => {
									console.error('Passkey registration failed:', error);
								}}
							/>
						</div>
					{/if}

				{:else if currentStep === 'passkey'}
					<!-- Passkey Authentication Step -->
					<div class="flex items-center gap-3">
						<Button variant="ghost" size="sm" onclick={handleBackToMethodSelect} disabled={isLoading}>
							<ArrowLeft class="h-4 w-4" />
						</Button>
						<div class="flex-1">
							<h1 class="text-xl font-semibold">Sign in with passkey</h1>
							<p class="text-sm text-muted-foreground">{email}</p>
						</div>
					</div>

					{#if error}
						<div class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-600">
							{error}
						</div>
					{/if}

					<div class="text-center space-y-4">
						{#if isLoading}
							<!-- Loading state -->
							<div class="flex justify-center">
								<div class="relative">
									<Key class="h-16 w-16 text-primary animate-pulse" />
									<div class="absolute -inset-2 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
								</div>
							</div>
							
							<div class="space-y-2">
								<p class="text-lg font-medium">Use your passkey to sign in</p>
								<p class="text-sm text-muted-foreground">
									Use Touch ID, Face ID, Windows Hello, or your security key
								</p>
							</div>
						{:else}
							<!-- Ready state -->
							<div class="space-y-4">
								<div class="flex justify-center">
									<Key class="h-16 w-16 text-primary" />
								</div>
								
								<div class="space-y-2">
									<p class="text-lg font-medium">Ready to authenticate</p>
									<p class="text-sm text-muted-foreground">
										Click below to use your passkey
									</p>
								</div>

								<Button onclick={handlePasskeyLogin} class="w-full" disabled={isLoading}>
									<Key class="h-4 w-4 mr-2" />
									Authenticate with passkey
								</Button>

								<Button variant="outline" onclick={() => handleMethodSelect('password')} class="w-full">
									Use password instead
								</Button>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</TabsContent>
		<TabsContent value="register">
			<SignupForm />
		</TabsContent>
	</Tabs>
</div>