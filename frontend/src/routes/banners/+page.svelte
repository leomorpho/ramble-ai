<script>
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { ArrowLeft, Eye, EyeOff, RefreshCw, Bell } from "@lucide/svelte";
  import { onMount } from "svelte";
  import { GetRambleAIApiKey } from "$lib/wailsjs/go/main/App";
  import { fetchAllBanners, dismissBanner } from "$lib/services/bannerService.js";

  /**
   * IMPORTANT: Wails App Architecture Note
   * 
   * This banner management page has been updated for Wails app compatibility.
   * It NO LONGER uses PocketBase SDK (pb.collection()) calls.
   * Instead, it uses the bannerService.js which calls custom API endpoints.
   * 
   * This is necessary because:
   * - Wails users authenticate with API keys, not PocketBase accounts
   * - Direct PocketBase collection access doesn't work in Wails
   * - All data must flow through custom API endpoints that handle PocketBase internally
   */

  let banners = $state([]);
  let loading = $state(true);
  let error = $state("");
  let apiKey = $state(null);

  // Load API key and banners on mount
  onMount(async () => {
    await loadAPIKey();
    await loadBanners();
  });

  async function loadAPIKey() {
    try {
      apiKey = await GetRambleAIApiKey();
    } catch (err) {
      console.log('No API key configured');
      apiKey = null;
    }
  }

  // Load all banners using the API service
  async function loadBanners() {
    try {
      loading = true;
      error = "";
      
      // Use the bannerService to fetch ALL banners (including dismissed ones)
      // This endpoint includes dismissal status for each banner
      const fetchedBanners = await fetchAllBanners(apiKey);
      banners = fetchedBanners;
      
      console.log('Loaded banners for management page:', fetchedBanners);
    } catch (err) {
      console.error('Failed to load banners:', err);
      error = "Failed to load banners: " + (err.message || "Unknown error");
    } finally {
      loading = false;
    }
  }

  // Handle banner dismissal through API
  async function handleDismiss(banner) {
    if (!apiKey) {
      console.warn('Cannot dismiss banner: no API key available');
      error = "API key required to dismiss banners";
      return;
    }

    try {
      error = "";
      await dismissBanner(banner.id, apiKey);
      
      // Reload banners to show updated state
      await loadBanners();
      
      console.log('Banner dismissed:', banner.title);
    } catch (err) {
      console.error('Failed to dismiss banner:', err);
      error = "Failed to dismiss banner: " + (err.message || "Unknown error");
    }
  }

  // Helper functions
  function getBannerTypeColor(type) {
    switch (type) {
      case 'info': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'warning': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'success': return 'bg-green-100 text-green-800 border-green-200';
      case 'error': return 'bg-red-100 text-red-800 border-red-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  }

  function formatDate(dateString) {
    if (!dateString) return 'Unknown';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return 'Invalid date';
    }
  }

  function isExpired(expiresAt) {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <Button variant="ghost" size="icon" asChild>
        <a href="/" class="flex items-center">
          <ArrowLeft class="h-4 w-4" />
        </a>
      </Button>
      
      <div class="flex-1">
        <h1 class="text-2xl font-semibold flex items-center gap-2">
          <Bell class="h-6 w-6" />
          Banner Management
        </h1>
        <p class="text-sm text-muted-foreground mt-1">
          View and manage all banners (both active and dismissed)
        </p>
      </div>

      <Button onclick={loadBanners} variant="outline" size="sm" disabled={loading}>
        <RefreshCw class={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
        Refresh
      </Button>
    </div>

    <!-- Summary Info -->
    {#if !loading && banners.length > 0}
      {@const activeBanners = banners.filter(b => !b.dismissed)}
      {@const dismissedBanners = banners.filter(b => b.dismissed)}
      <div class="border rounded p-4 bg-muted/50">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">Banner Summary</h3>
            <p class="text-sm text-muted-foreground">
              {activeBanners.length} active, {dismissedBanners.length} dismissed
            </p>
          </div>
          <div class="flex gap-2">
            <Badge variant="default">{activeBanners.length} Active</Badge>
            <Badge variant="secondary">{dismissedBanners.length} Dismissed</Badge>
          </div>
        </div>
      </div>
    {/if}

    <!-- Error Display -->
    {#if error}
      <div class="border border-destructive rounded p-4 bg-destructive/10">
        <p class="text-destructive font-medium">Error</p>
        <p class="text-sm text-destructive/80">{error}</p>
      </div>
    {/if}

    <!-- Loading State -->
    {#if loading}
      <div class="text-center py-8">
        <RefreshCw class="h-6 w-6 animate-spin mx-auto mb-2" />
        <p class="text-muted-foreground">Loading banners...</p>
      </div>
    {:else if banners.length === 0}
      <!-- Empty State -->
      <div class="text-center py-12 space-y-4">
        <Bell class="h-12 w-12 mx-auto text-muted-foreground" />
        <div>
          <h3 class="text-lg font-medium">No banners found</h3>
          <p class="text-muted-foreground">
            {apiKey 
              ? 'There are no banners available at the moment.'
              : 'Configure an API key in settings to see personalized banners.'
            }
          </p>
        </div>
        {#if !apiKey}
          <Button asChild variant="outline">
            <a href="/settings">Configure API Key</a>
          </Button>
        {/if}
      </div>
    {:else}
      <!-- Banner List -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-medium">
            Current Banners ({banners.length})
          </h2>
          <p class="text-sm text-muted-foreground">
            All banners available to {apiKey ? 'your API key' : 'public users'}
          </p>
        </div>

        <div class="grid gap-4">
          {#each banners as banner (banner.id)}
            <div class="border rounded p-4 space-y-3">
              <!-- Header -->
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-2">
                    <h3 class="font-medium">{banner.title}</h3>
                    <Badge class={getBannerTypeColor(banner.type)} variant="outline">
                      {banner.type}
                    </Badge>
                    {#if !banner.active}
                      <Badge variant="secondary">Inactive</Badge>
                    {/if}
                    {#if banner.requires_auth}
                      <Badge variant="outline">Auth Required</Badge>
                    {/if}
                    {#if isExpired(banner.expires_at)}
                      <Badge variant="destructive">Expired</Badge>
                    {/if}
                    {#if banner.dismissed}
                      <Badge variant="secondary">Dismissed</Badge>
                    {/if}
                  </div>
                  
                  <p class="text-sm text-muted-foreground mb-2">{banner.message}</p>
                  
                  <div class="flex items-center gap-4 text-xs text-muted-foreground">
                    <span>Created: {formatDate(banner.created)}</span>
                    {#if banner.expires_at}
                      <span>Expires: {formatDate(banner.expires_at)}</span>
                    {:else}
                      <span>No expiration</span>
                    {/if}
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex items-center gap-2">
                  {#if banner.action_url && banner.action_text}
                    <Button size="sm" variant="outline" asChild>
                      <a href={banner.action_url} target="_blank" rel="noopener noreferrer">
                        {banner.action_text}
                      </a>
                    </Button>
                  {/if}
                  
                  {#if apiKey && !banner.dismissed}
                    <Button 
                      size="sm" 
                      variant="ghost" 
                      onclick={() => handleDismiss(banner)}
                      title="Dismiss this banner"
                    >
                      <EyeOff class="h-4 w-4" />
                    </Button>
                  {:else if banner.dismissed}
                    <Button 
                      size="sm" 
                      variant="ghost" 
                      disabled
                      title="Banner already dismissed"
                    >
                      <Eye class="h-4 w-4 opacity-50" />
                    </Button>
                  {:else}
                    <Button 
                      size="sm" 
                      variant="ghost" 
                      disabled
                      title="API key required to dismiss banners"
                    >
                      <EyeOff class="h-4 w-4" />
                    </Button>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Help Text -->
    <div class="border-t pt-6 text-sm text-muted-foreground">
      <p>
        <strong>Note:</strong> This page shows all banners available to {apiKey ? 'your API key' : 'public users'}, 
        including both active and dismissed banners. 
        {apiKey 
          ? 'Dismissed banners will not appear in the main app notification dropdown.' 
          : 'Configure an API key in settings to see authenticated banners and dismiss functionality.'
        }
      </p>
    </div>
  </div>
</main>