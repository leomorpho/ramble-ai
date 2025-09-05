<script>
  import { Button } from "$lib/components/ui/button";
  import { Bell, X } from "@lucide/svelte";
  import { onMount, onDestroy } from "svelte";
  import { browser } from '$app/environment';
  import { GetRambleAIApiKey } from "$lib/wailsjs/go/main/App";
  import { fetchBanners, dismissBanner } from "$lib/services/bannerService.js";
  
  /**
   * Banner Notification Dropdown Component
   * 
   * Similar to Facebook notifications - shows active banners in a dropdown
   * with a counter badge. Allows dismissing banners directly from the dropdown.
   */

  let { class: className = "" } = $props();
  
  let dropdownOpen = $state(false);
  let banners = $state([]);
  let loading = $state(true);
  let apiKey = $state(null);
  let dropdownElement = $state(null);
  let buttonElement = $state(null);

  // Load banners on mount
  onMount(async () => {
    await loadAPIKey();
    await loadBanners();
    
    // Close dropdown when clicking outside (only in browser)
    if (browser) {
      document.addEventListener('click', handleClickOutside);
    }
  });

  onDestroy(() => {
    if (browser) {
      document.removeEventListener('click', handleClickOutside);
    }
  });

  async function loadAPIKey() {
    try {
      apiKey = await GetRambleAIApiKey();
    } catch (err) {
      apiKey = null;
    }
  }

  async function loadBanners() {
    try {
      loading = true;
      // Use fetchBanners (not fetchAllBanners) to get only non-dismissed banners
      const fetchedBanners = await fetchBanners(apiKey);
      banners = fetchedBanners;
    } catch (error) {
      console.error('Failed to load banners for dropdown:', error);
      banners = [];
    } finally {
      loading = false;
    }
  }

  async function handleDismiss(bannerId) {
    if (!apiKey) {
      console.warn('Cannot dismiss banner: no API key available');
      return;
    }

    try {
      await dismissBanner(bannerId, apiKey);
      
      // Remove dismissed banner from local state immediately (like Facebook)
      banners = banners.filter(banner => banner.id !== bannerId);
      
      console.log('Banner dismissed successfully:', bannerId);
    } catch (error) {
      console.error('Failed to dismiss banner:', error);
    }
  }

  function toggleDropdown() {
    dropdownOpen = !dropdownOpen;
  }

  function handleClickOutside(event) {
    if (dropdownOpen && 
        dropdownElement && 
        buttonElement && 
        !dropdownElement.contains(event.target) && 
        !buttonElement.contains(event.target)) {
      dropdownOpen = false;
    }
  }

  function getBannerIcon(type) {
    switch (type) {
      case 'warning': return '⚠️';
      case 'error': return '❌';
      case 'success': return '✅';
      default: return '💡';
    }
  }

  function getBannerColor(type) {
    switch (type) {
      case 'warning': return 'border-l-yellow-500';
      case 'error': return 'border-l-red-500';
      case 'success': return 'border-l-green-500';
      default: return 'border-l-blue-500';
    }
  }

  // Count for badge (total non-dismissed banners)
  let bannerCount = $derived(banners.length);
</script>

<div class="relative {className}">
  <!-- Notification Bell Button -->
  <Button
    bind:this={buttonElement}
    variant="ghost"
    size="icon"
    class="h-9 w-9 relative"
    onclick={toggleDropdown}
    title="Notifications ({bannerCount})"
  >
    <Bell class="h-4 w-4" />
    
    <!-- Counter Badge -->
    {#if bannerCount > 0}
      <span class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center font-medium">
        {bannerCount > 9 ? '9+' : bannerCount}
      </span>
    {/if}
  </Button>

  <!-- Dropdown -->
  {#if dropdownOpen}
    <div
      bind:this={dropdownElement}
      class="absolute right-0 top-full mt-2 w-96 bg-background border border-border rounded-lg shadow-lg z-50"
    >
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-border">
        <h3 class="font-medium">Notifications</h3>
        {#if bannerCount > 0}
          <span class="text-sm text-muted-foreground">{bannerCount} notification{bannerCount !== 1 ? 's' : ''}</span>
        {/if}
      </div>

      <!-- Content -->
      <div class="max-h-96 overflow-y-auto">
        {#if loading}
          <div class="p-4 text-center text-muted-foreground">
            Loading notifications...
          </div>
        {:else if banners.length === 0}
          <div class="p-8 text-center">
            <Bell class="h-8 w-8 mx-auto text-muted-foreground mb-2" />
            <p class="text-muted-foreground">No new notifications</p>
            <p class="text-sm text-muted-foreground mt-1">
              {apiKey ? 'All caught up!' : 'Configure an API key to see notifications'}
            </p>
            
            <!-- Link to see all banners, even when dismissed -->
            <div class="mt-4">
              <Button variant="outline" size="sm" class="text-xs" asChild>
                <a href="/banners">View All Notifications</a>
              </Button>
            </div>
          </div>
        {:else}
          <!-- Banner List -->
          <div class="divide-y divide-border">
            {#each banners as banner (banner.id)}
              <div class="p-4 hover:bg-muted/50 relative group border-l-2 {getBannerColor(banner.type)}">
                <!-- Dismiss Button -->
                {#if apiKey}
                  <button
                    class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-muted rounded"
                    onclick={() => handleDismiss(banner.id)}
                    title="Dismiss"
                  >
                    <X class="h-3 w-3" />
                  </button>
                {/if}

                <!-- Banner Content -->
                <div class="pr-6">
                  <div class="flex items-start gap-2 mb-2">
                    <span class="text-sm">{getBannerIcon(banner.type)}</span>
                    <h4 class="font-medium text-sm leading-tight">{banner.title}</h4>
                  </div>
                  
                  <p class="text-sm text-muted-foreground mb-3 line-clamp-3">
                    {banner.message}
                  </p>

                  <!-- Action Button -->
                  {#if banner.action_url && banner.action_text}
                    <div class="flex justify-end">
                      <Button 
                        size="sm" 
                        variant="outline" 
                        class="text-xs h-7"
                        asChild
                      >
                        <a href={banner.action_url} target="_blank" rel="noopener noreferrer">
                          {banner.action_text}
                        </a>
                      </Button>
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      {#if banners.length > 0}
        <div class="p-3 border-t border-border bg-muted/30">
          <Button variant="ghost" size="sm" class="w-full text-xs" asChild>
            <a href="/banners">Manage All Notifications</a>
          </Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .line-clamp-3 {
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>