<script>
  import { ChevronDown, ChevronUp } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Banner } from "$lib/components/ui/banner";
  import { cn } from "$lib/utils";

  let { 
    banners = [],
    class: className,
    ...restProps
  } = $props();

  let collapsed = $state(false);
  let dismissedBanners = $state([]);

  // Load dismissed banners from localStorage
  $effect(() => {
    try {
      const stored = localStorage.getItem('dismissedBanners');
      if (stored) {
        dismissedBanners = JSON.parse(stored);
      }
    } catch (e) {
      console.warn('Failed to load dismissed banners from localStorage:', e);
    }
  });

  // Save dismissed banners to localStorage when they change
  $effect(() => {
    try {
      localStorage.setItem('dismissedBanners', JSON.stringify(dismissedBanners));
    } catch (e) {
      console.warn('Failed to save dismissed banners to localStorage:', e);
    }
  });

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

  // Filter out dismissed banners
  let visibleBanners = $derived(
    banners.filter(banner => !dismissedBanners.includes(banner.id))
  );

  function dismissBanner(bannerId) {
    if (!dismissedBanners.includes(bannerId)) {
      dismissedBanners = [...dismissedBanners, bannerId];
    }
  }

  function toggleCollapsed() {
    collapsed = !collapsed;
  }

</script>

{#if visibleBanners.length > 0}
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
            onDismiss={() => dismissBanner(banner.id)}
          />
        {/each}
      </div>
    {/if}
  </div>
{/if}