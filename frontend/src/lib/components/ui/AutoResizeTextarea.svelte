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

  let textareaElement = $state();

  // Function to adjust the height dynamically
  function autoGrow(event) {
    const textarea = event.target;
    textarea.style.height = 'auto'; // Reset height to recalculate
    textarea.style.height = `${textarea.scrollHeight + 5}px`; // Set height to match content
  }

  // Function to resize using bound element
  function resizeTextarea() {
    if (textareaElement) {
      textareaElement.style.height = 'auto';
      textareaElement.style.height = `${textareaElement.scrollHeight + 5}px`;
    }
  }

  // Watch for value changes and resize immediately
  $effect(() => {
    if (textareaElement && value !== undefined) {
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
</script>

<Textarea
  bind:this={textareaElement}
  class={`resize-none ${className}`}
  style="min-height: 120px;"
  {placeholder}
  bind:value
  oninput={handleInput}
  {disabled}
  {...restProps}
/>