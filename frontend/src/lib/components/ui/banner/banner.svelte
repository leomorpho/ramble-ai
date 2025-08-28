<script>
  import { X, Info, AlertTriangle, CheckCircle, AlertCircle, ExternalLink } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { cn } from "$lib/utils";

  let { 
    type = "info",
    title,
    message, 
    actionText,
    actionUrl,
    dismissible = true,
    onDismiss,
    class: className,
    ...restProps
  } = $props();

  const typeConfig = {
    info: {
      icon: Info,
      bgColor: "bg-blue-50 dark:bg-blue-950/50",
      borderColor: "border-blue-200 dark:border-blue-800",
      textColor: "text-blue-800 dark:text-blue-200",
      iconColor: "text-blue-600 dark:text-blue-400"
    },
    warning: {
      icon: AlertTriangle,
      bgColor: "bg-yellow-50 dark:bg-yellow-950/50",
      borderColor: "border-yellow-200 dark:border-yellow-800",
      textColor: "text-yellow-800 dark:text-yellow-200",
      iconColor: "text-yellow-600 dark:text-yellow-400"
    },
    success: {
      icon: CheckCircle,
      bgColor: "bg-green-50 dark:bg-green-950/50",
      borderColor: "border-green-200 dark:border-green-800",
      textColor: "text-green-800 dark:text-green-200",
      iconColor: "text-green-600 dark:text-green-400"
    },
    error: {
      icon: AlertCircle,
      bgColor: "bg-red-50 dark:bg-red-950/50",
      borderColor: "border-red-200 dark:border-red-800",
      textColor: "text-red-800 dark:text-red-200",
      iconColor: "text-red-600 dark:text-red-400"
    }
  };

  const config = typeConfig[type] || typeConfig.info;
  const IconComponent = config.icon;
</script>

<div
  class={cn(
    "rounded-lg border p-4 mb-4",
    config.bgColor,
    config.borderColor,
    className
  )}
  {...restProps}
>
  <div class="flex items-start gap-3">
    <IconComponent class={cn("h-5 w-5 mt-0.5 flex-shrink-0", config.iconColor)} />
    
    <div class="flex-1 min-w-0">
      {#if title}
        <h4 class={cn("font-medium text-sm mb-1", config.textColor)}>
          {title}
        </h4>
      {/if}
      
      <div class={cn("text-sm", config.textColor)}>
        {@html message}
      </div>

      {#if actionText && actionUrl}
        <div class="mt-3">
          <Button 
            variant="outline" 
            size="sm"
            class={cn(
              "h-8 text-xs",
              config.textColor,
              config.borderColor,
              "hover:bg-background/80"
            )}
            onclick={() => window.open(actionUrl, '_blank')}
          >
            {actionText}
            <ExternalLink class="h-3 w-3 ml-1" />
          </Button>
        </div>
      {/if}
    </div>

    {#if dismissible && onDismiss}
      <Button
        variant="ghost"
        size="icon"
        class={cn(
          "h-6 w-6 flex-shrink-0",
          config.textColor,
          "hover:bg-background/50"
        )}
        onclick={onDismiss}
      >
        <X class="h-4 w-4" />
      </Button>
    {/if}
  </div>
</div>