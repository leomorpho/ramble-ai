<script>
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";

  // Props
  let {
    open = $bindable(false),
    title = "",
    description = "",
    icon,
    children,
    footer,
    headerActions,
    onClose = () => {},
  } = $props();

  // Animation state
  let isVisible = $state(false);
  let isAnimating = $state(false);
  let shouldAnimate = $state(false);

  // Handle opening and closing animations
  $effect(() => {
    if (open) {
      // Opening
      isVisible = true;
      // Prevent body scroll
      document.body.style.overflow = 'hidden';
      // Start animation after DOM update
      setTimeout(() => {
        shouldAnimate = true;
      }, 10);
    } else if (isVisible) {
      // Closing
      shouldAnimate = false;
      // Restore body scroll
      document.body.style.overflow = '';
      // Hide after animation
      setTimeout(() => {
        isVisible = false;
        isAnimating = false;
      }, 300);
    }
  });

  // Close function with animation
  function closeSheet() {
    open = false;
    onClose();
  }

  // Cleanup on unmount
  onMount(() => {
    return () => {
      document.body.style.overflow = '';
    };
  });
</script>

<!-- Custom Sheet -->
{#if isVisible}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm transition-opacity duration-300"
    class:opacity-0={!shouldAnimate}
    class:opacity-100={shouldAnimate}
    onclick={closeSheet}
  ></div>

  <!-- Sheet Container -->
  <div class="fixed inset-0 z-50 flex justify-end pointer-events-none">
    <!-- Sheet Content -->
    <div
      class="w-[90vw] h-screen bg-background border-l shadow-2xl flex flex-col pointer-events-auto transition-transform duration-300 ease-in-out"
      class:translate-x-full={!shouldAnimate}
      class:translate-x-0={shouldAnimate}
    >
      <!-- Header -->
      <div class="flex-shrink-0 border-b px-6 py-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-semibold flex items-center gap-2">
              {#if icon}
                {@render icon()}
              {/if}
              {title}
            </h2>
            {#if description}
              <p class="text-sm text-muted-foreground mt-1">
                {description}
              </p>
            {/if}
          </div>
          <div class="flex items-center gap-2">
            {#if headerActions}
              {@render headerActions()}
            {/if}
            <Button
              variant="ghost"
              size="sm"
              onclick={closeSheet}
              class="h-8 w-8 p-0"
            >
              <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </Button>
          </div>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-hidden flex flex-col">
        {@render children()}
      </div>

      <!-- Footer -->
      {#if footer}
        <div class="flex-shrink-0 border-t px-6 py-4">
          {@render footer({ closeSheet })}
        </div>
      {/if}
    </div>
  </div>
{/if}