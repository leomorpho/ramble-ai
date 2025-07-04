<script>
  import { onMount, onDestroy } from "svelte";

  let {
    onPlayPause = () => {},
    isEnabled = true,
    children,
  } = $props();

  let containerElement = $state(null);
  let hasFocus = $state(false);

  // Handle spacebar key press
  function handleKeydown(event) {
    // Only handle spacebar and only when we have focus and are enabled
    if (event.code === 'Space' && hasFocus && isEnabled) {
      // Prevent default scrolling behavior
      event.preventDefault();
      event.stopPropagation();
      
      // Call the play/pause handler
      onPlayPause();
    }
  }

  // Handle focus events on the container
  function handleFocusIn(event) {
    // Check if the focus is within our container
    if (containerElement && containerElement.contains(event.target)) {
      hasFocus = true;
    }
  }

  function handleFocusOut(event) {
    // Check if the focus is moving outside our container
    if (containerElement && !containerElement.contains(event.relatedTarget)) {
      hasFocus = false;
    }
  }

  // Handle clicks within the container to establish focus
  function handleClick(event) {
    if (containerElement && containerElement.contains(event.target)) {
      hasFocus = true;
    }
  }

  // Handle clicks outside the container to lose focus
  function handleDocumentClick(event) {
    if (containerElement && !containerElement.contains(event.target)) {
      hasFocus = false;
    }
  }

  // Handle any user interaction within the container (including button clicks)
  function handleUserInteraction(event) {
    if (containerElement && containerElement.contains(event.target)) {
      hasFocus = true;
    }
  }

  onMount(() => {
    // Add global event listeners
    document.addEventListener('keydown', handleKeydown);
    document.addEventListener('focusin', handleFocusIn);
    document.addEventListener('focusout', handleFocusOut);
    document.addEventListener('click', handleDocumentClick);
    document.addEventListener('mousedown', handleUserInteraction);
  });

  onDestroy(() => {
    // Clean up global event listeners
    document.removeEventListener('keydown', handleKeydown);
    document.removeEventListener('focusin', handleFocusIn);
    document.removeEventListener('focusout', handleFocusOut);
    document.removeEventListener('click', handleDocumentClick);
    document.removeEventListener('mousedown', handleUserInteraction);
  });
</script>

<!-- 
  This component wraps the video player and manages spacebar play/pause functionality.
  It tracks focus state and only responds to spacebar when the user has interacted with the player.
-->
<div 
  bind:this={containerElement}
  onclick={handleClick}
  onmousedown={handleUserInteraction}
  class="video-player-key-handler focus-within:outline-none"
  tabindex="-1"
>
  {@render children()}
</div>