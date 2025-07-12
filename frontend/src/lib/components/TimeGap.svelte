<script>
  // Props
  let {
    duration = 0,
    showNormal = false,
    size = "sm" // "xs", "sm", "md", "lg"
  } = $props();

  // Pause detection thresholds (configurable via props if needed)
  const LONG_PAUSE_THRESHOLD = 0.8; // seconds - noticeable pause
  const VERY_LONG_PAUSE_THRESHOLD = 1.5; // seconds - significant pause

  // Get pause styling based on duration
  function getPauseStyle(pauseDuration) {
    if (pauseDuration >= VERY_LONG_PAUSE_THRESHOLD) {
      return {
        type: 'very-long',
        bgColor: 'bg-destructive/10',
        textColor: 'text-destructive',
        borderColor: 'border-destructive/20',
        hoverColor: 'hover:bg-destructive/20'
      };
    } else if (pauseDuration >= LONG_PAUSE_THRESHOLD) {
      return {
        type: 'long',
        bgColor: 'bg-orange-100',
        textColor: 'text-orange-700',
        borderColor: 'border-orange-200',
        hoverColor: 'hover:bg-orange-200'
      };
    }
    return {
      type: 'normal',
      bgColor: 'bg-muted/30',
      textColor: 'text-muted-foreground',
      borderColor: 'border-muted/40',
      hoverColor: 'hover:bg-muted/50'
    };
  }

  // Size configurations
  const sizeClasses = {
    xs: {
      container: 'px-1 py-0.5',
      text: 'text-xs',
      icon: 'w-2 h-2',
      spacing: 'mx-0.5'
    },
    sm: {
      container: 'px-1.5 py-0.5',
      text: 'text-xs',
      icon: 'w-2.5 h-2.5',
      spacing: 'mx-0.5'
    },
    md: {
      container: 'px-2 py-1',
      text: 'text-sm',
      icon: 'w-3 h-3',
      spacing: 'mx-1'
    },
    lg: {
      container: 'px-3 py-1.5',
      text: 'text-base',
      icon: 'w-4 h-4',
      spacing: 'mx-1.5'
    }
  };

  // Computed values
  let style = $derived(getPauseStyle(duration));
  let sizeConfig = $derived(sizeClasses[size] || sizeClasses.sm);
  let shouldShow = $derived(style.type !== 'normal' || showNormal);

  // Generate tooltip text
  let tooltipText = $derived(() => {
    const typeLabel = style.type === 'very-long' ? 'Very long pause' : 
                     style.type === 'long' ? 'Long pause' : 'Normal pause';
    const warning = style.type === 'very-long' ? ' - this may indicate an unnatural break' : '';
    return `${typeLabel}: ${duration.toFixed(2)}s${warning}`;
  });
</script>

{#if shouldShow && duration > 0}
  <span 
    class="inline-flex items-center {sizeConfig.spacing} {sizeConfig.container} rounded {style.bgColor} {style.textColor} {style.borderColor} {style.hoverColor} {sizeConfig.text} font-medium border transition-colors"
    title={tooltipText}
  >
    <svg class="{sizeConfig.icon} mr-0.5" fill="currentColor" viewBox="0 0 20 20">
      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
    </svg>
    {duration.toFixed(1)}s
  </span>
{:else if duration > 0}
  <!-- Normal space for short pauses when not showing all -->
  <span class="inline-block w-1"></span>
{/if}