<script>
  import { ChevronDown, ChevronUp } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Banner } from "$lib/components/ui/banner";
  import { cn } from "$lib/utils";
  import { onMount } from "svelte";
  import { GetRambleAIApiKey } from "$lib/wailsjs/go/main/App";
  import { fetchBanners, dismissBanner } from "$lib/services/bannerService.js";

  /**
   * IMPORTANT: Wails App Architecture Note
   * 
   * This component does NOT use PocketBase SDK for authentication or data fetching.
   * In the Wails desktop app:
   * - Users authenticate with an API key, NOT PocketBase login
   * - Direct PocketBase collection access (pb.collection()) doesn't work
   * - All data must be fetched through custom API endpoints via bannerService.js
   * - The backend exposes specific API routes that handle PocketBase internally
   * 
   * This is why we use fetchBanners() from bannerService.js instead of 
   * pb.collection('banners').getFullList()
   */

  let { 
    class: className,
    ...restProps
  } = $props();

  let collapsed = $state(false);
  let banners = $state([]);
  let loading = $state(true);
  let apiKey = $state(null);

  // Load banners through the custom API endpoint
  // NOT through direct PocketBase SDK collection access
  async function loadBanners() {
    try {
      loading = true;
      
      // Get API key if user has one configured
      // This is how Wails users authenticate, not through PocketBase login
      try {
        apiKey = await GetRambleAIApiKey();
      } catch (err) {
        // User may not have API key configured, that's OK
        // We'll still fetch public banners
        console.log('No API key configured, fetching public banners only');
        apiKey = null;
      }
      
      // Use the bannerService to fetch from custom API endpoints
      // This goes through /api/banners endpoints, not PocketBase collections
      const fetchedBanners = await fetchBanners(apiKey);
      
      banners = fetchedBanners;
    } catch (error) {
      console.error('Failed to load banners:', error);
      banners = [];
    } finally {
      loading = false;
    }
  }

  // Load collapsed state from localStorage
  $effect(() => {
    try {
      const storedCollapsed = localStorage.getItem('bannersCollapsed');
      if (storedCollapsed !== null) {
        collapsed = JSON.parse(storedCollapsed);
      }
    } catch (e) {
      console.warn('Failed to load banner collapsed state:', e);
    }
  });

  // Save collapsed state when it changes
  $effect(() => {
    try {
      localStorage.setItem('bannersCollapsed', JSON.stringify(collapsed));
    } catch (e) {
      console.warn('Failed to save banner collapsed state:', e);
    }
  });

  // All banners are visible since the API already filters dismissed ones
  // The backend now handles dismissal filtering per API key
  let visibleBanners = $derived(() => {
    return banners; // No need to filter locally anymore
  });

  async function handleDismissBanner(bannerId) {
    if (!apiKey) {
      console.warn('Cannot dismiss banner: no API key available');
      return;
    }

    try {
      // Call the API to dismiss the banner
      await dismissBanner(bannerId, apiKey);
      
      // Reload banners to get updated list (without dismissed banner)
      await loadBanners();
      
      console.log('Banner dismissed successfully:', bannerId);
    } catch (error) {
      console.error('Failed to dismiss banner:', error);
      // Could show a toast notification here
    }
  }

  function toggleCollapsed() {
    collapsed = !collapsed;
  }

  // Load banners on mount
  onMount(() => {
    loadBanners();
  });

</script>

{#if loading}
  <div class={cn("space-y-2", className)} {...restProps}>
    <div class="flex items-center justify-between">
      <h3 class="text-sm font-medium text-muted-foreground">Loading announcements...</h3>
    </div>
  </div>
{:else if visibleBanners.length > 0}
  <div class={cn("space-y-2", className)} {...restProps}>
    <!-- Header with collapse toggle -->
    <div class="flex items-center justify-between">
      <h3 class="text-sm font-medium text-muted-foreground">
        Updates & Announcements
        {#if visibleBanners.length > 0}
          <span class="ml-1 text-xs bg-muted text-muted-foreground rounded-full px-2 py-0.5">
            {visibleBanners.length}
          </span>
        {/if}
      </h3>
      
      <Button
        variant="ghost"
        size="sm"
        class="h-8 px-2 text-muted-foreground hover:text-foreground"
        onclick={toggleCollapsed}
      >
        {collapsed ? 'Show' : 'Hide'}
        {#if collapsed}
          <ChevronDown class="h-4 w-4 ml-1" />
        {:else}
          <ChevronUp class="h-4 w-4 ml-1" />
        {/if}
      </Button>
    </div>

    <!-- Banner list - collapsible -->
    {#if !collapsed}
      <div class="space-y-2">
        {#each visibleBanners as banner (banner.id)}
          <Banner
            type={banner.type}
            title={banner.title}
            message={banner.message}
            actionText={banner.action_text}
            actionUrl={banner.action_url}
            onDismiss={() => handleDismissBanner(banner.id)}
          />
        {/each}
      </div>
    {/if}
  </div>
{/if}