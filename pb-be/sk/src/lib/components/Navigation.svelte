<script lang="ts">
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import DownloadButton from '$lib/components/DownloadButton.svelte';
	import { Home, LogIn, LogOut, User, CreditCard, Crown, BarChart } from 'lucide-svelte';
	import type { AuthModel } from 'pocketbase';
	import { config } from '$lib/config.js';
	import { getAvatarUrl } from '$lib/files.js';

	let {
		isLoggedIn = false,
		user = null,
		isSubscribed = false,
		onLogout
	}: {
		isLoggedIn?: boolean;
		user?: AuthModel | null;
		isSubscribed?: boolean;
		onLogout?: () => void;
	} = $props();

	function handleLogout() {
		onLogout?.();
	}
</script>

<header class="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-lg border-b">
	<div class="container mx-auto px-4 py-4">
		<nav class="flex items-center justify-between">
			<div class="flex items-center space-x-4">
				<a href="/" class="flex items-center space-x-3 hover:opacity-80 transition-opacity cursor-pointer">
					<img 
						src="/logo-128.png" 
						alt="Ramble logo" 
						class="w-8 h-8 rounded-lg"
					/>
					<span class="font-bold tracking-tight text-xl">
						<span class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">RAMBLE</span>
					</span>
				</a>
			</div>

			<div class="flex items-center space-x-2">
				<a
					href="/"
					class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
					title="Home"
				>
					<Home class="h-4 w-4" />
				</a>
				<a
					href="/pricing"
					class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
					title="Pricing"
				>
					<CreditCard class="h-4 w-4" />
				</a>

				{#if isLoggedIn}
					<a
						href="/usages"
						class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
						title="Usage Statistics"
					>
						<BarChart class="h-4 w-4" />
					</a>
					{#if isSubscribed}
						<a
							href="/premium"
							class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
							title="Premium Features"
						>
							<Crown class="h-4 w-4 text-yellow-600" />
						</a>
					{/if}
					<a
						href="/dashboard"
						class="flex items-center space-x-2 text-sm hover:bg-accent hover:text-accent-foreground px-2 py-1 rounded-md transition-colors"
					>
						{#if getAvatarUrl(user, 'small')}
							<img
								src={getAvatarUrl(user, 'small')}
								alt="Profile"
								class="w-6 h-6 rounded-full object-cover border border-border"
							/>
						{:else}
							<div class="w-6 h-6 rounded-full bg-muted border border-border flex items-center justify-center">
								<User class="h-3 w-3 text-muted-foreground" />
							</div>
						{/if}
						<span class="hidden sm:inline">{user?.name || user?.email}</span>
					</a>
					<button
						onclick={handleLogout}
						class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
						title="Sign Out"
					>
						<LogOut class="h-4 w-4" />
					</button>
				{:else}
					<a
						href="/login"
						class="hover:bg-accent hover:text-accent-foreground inline-flex h-9 w-9 items-center justify-center rounded-md text-sm font-medium whitespace-nowrap transition-colors"
						title="Sign In"
					>
						<LogIn class="h-4 w-4" />
					</a>
					<DownloadButton text="Get Started" size="sm" showIcon={false} />
				{/if}

				<ThemeToggle />
			</div>
		</nav>
	</div>
</header>
