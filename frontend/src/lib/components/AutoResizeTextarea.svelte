<script>
  import { Textarea } from "$lib/components/ui/textarea";

  let {
    oninput,
    class: className = "",
    value = $bindable(""),
    placeholder = "",
    disabled = false,
    ...restProps
  } = $props();

  let containerElement = $state();

  // Function to adjust the height dynamically
  function autoGrow(event) {
    const textarea = event.target;
    textarea.style.height = 'auto'; // Reset height to recalculate
    textarea.style.height = `${textarea.scrollHeight + 5}px`; // Set height to match content
  }

  // Function to resize using bound element
  function resizeTextarea() {
    if (containerElement) {
      const textarea = containerElement.querySelector('textarea');
      if (textarea) {
        textarea.style.height = 'auto';
        textarea.style.height = `${textarea.scrollHeight + 5}px`;
      }
    }
  }

  // Watch for value changes and resize immediately
  $effect(() => {
    if (containerElement && value !== undefined) {
      // Use requestAnimationFrame to ensure DOM is updated
      requestAnimationFrame(() => {
        resizeTextarea();
      });
    }
  });

  // Function to handle input changes
  function handleInput(event) {
    autoGrow(event);
    oninput?.(event);
  }

  // Find the actual textarea DOM element and expose focus method
  function focus() {
    if (containerElement) {
      const textarea = containerElement.querySelector('textarea');
      if (textarea) {
        textarea.focus();
      }
    }
  }

  // Expose the focus method to parent components
  $effect(() => {
    if (containerElement) {
      containerElement.focus = focus;
    }
  });
</script>

<div bind:this={containerElement} class="w-full">
  <Textarea
    class={`resize-none w-full ${className}`}
    {placeholder}
    bind:value
    oninput={handleInput}
    {disabled}
    {...restProps}
  />
</div>