<script>
  import HighlightItem from "$lib/components/HighlightItem.svelte";
  import NewLineItem from "$lib/components/NewLineItem.svelte";
  import AddNewLineButton from "$lib/components/AddNewLineButton.svelte";

  // Props
  let {
    highlights = [],
    selectedHighlights = $bindable(new Set()),
    // Callbacks
    onReorder = (newOrder) => {},
    onSelect = (event, highlight) => {},
    onEdit = (event, highlight) => {},
    onDelete = (event, highlight) => {},
    onPopoverOpenChange = (highlightId, isOpen) => {},
    getHighlightWords = (highlight) => [],
    isPopoverOpen = (highlightId) => false,
    // New callback for title changes
    onTitleChange = (index, newTitle) => {},
    // Configuration
    enableMultiSelect = true,
    enableNewlines = true,
    enableSelection = true,
    enableEdit = true,
    enableDelete = true,
    enableDrag = true,
    showAddNewLineButtons = true,
    // Optional click handler for playing highlights
    onHighlightClick = null,
    // Container styling
    containerClass = "p-4 bg-muted/30 rounded-lg min-h-[80px] relative leading-relaxed text-base",
  } = $props();

  // Drag and drop state
  let isDragging = $state(false);
  let draggedHighlights = $state([]);
  let dropPosition = $state(null);
  let dragStartPosition = $state(null);
  let isDropping = $state(false);

  // Handle highlight selection with multiselect support
  function handleHighlightSelect(event, highlight) {
    if (!enableSelection) return;

    // Don't select new lines
    if (isNewline(highlight)) return;

    if (enableMultiSelect) {
      const isCtrlOrCmd = event.ctrlKey || event.metaKey;

      if (isCtrlOrCmd) {
        // Toggle selection for this highlight
        const newSelection = new Set(selectedHighlights);
        if (newSelection.has(highlight.id)) {
          newSelection.delete(highlight.id);
        } else {
          newSelection.add(highlight.id);
        }
        selectedHighlights = newSelection;
      } else {
        // Single select - clear others and select this one, or trigger click if already selected
        if (
          selectedHighlights.has(highlight.id) &&
          selectedHighlights.size === 1 &&
          onHighlightClick
        ) {
          // If it's the only selected item and click handler exists, trigger it
          onHighlightClick(highlight);
        } else {
          // Single select this highlight
          selectedHighlights = new Set([highlight.id]);
        }
      }
    } else {
      // Simple selection without multiselect
      selectedHighlights = new Set([highlight.id]);
    }

    // Call custom select handler if provided
    if (onSelect) {
      onSelect(event, highlight);
    }
  }

  // Handle drag start with multiselect support
  function handleDragStart(event, highlight, index) {
    if (!enableDrag) {
      event.preventDefault();
      return;
    }

    event.dataTransfer.effectAllowed = "move";

    if (isNewline(highlight)) {
      // New lines are draggable individually
      isDragging = true;
      dragStartPosition = index;
      draggedHighlights = [highlight.id];
    } else {
      if (enableMultiSelect) {
        // If the dragged highlight is not selected, select only it
        if (!selectedHighlights.has(highlight.id)) {
          selectedHighlights = new Set([highlight.id]);
        }

        // Set up drag state
        isDragging = true;
        dragStartPosition = index;
        draggedHighlights = Array.from(selectedHighlights);
      } else {
        // Simple single-item drag
        isDragging = true;
        dragStartPosition = index;
        draggedHighlights = [highlight.id];
      }
    }

    // Store the highlight IDs in dataTransfer for the drag operation
    event.dataTransfer.setData("text/plain", JSON.stringify(draggedHighlights));
  }

  // Handle container-level drag over
  function handleContainerDragOver(event) {
    event.preventDefault();
    if (isDragging) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  // Handle container-level drop
  async function handleContainerDrop(event) {
    event.preventDefault();

    if (isDragging) {
      // Default to dropping at the end if no position set
      if (dropPosition === null) {
        dropPosition = highlights.length;
      }
      await performDrop();
    }
  }

  // Handle container drag leave
  function handleContainerDragLeave(event) {
    // Only clear if we're leaving the container entirely
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX;
    const y = event.clientY;

    if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
      dropPosition = null;
    }
  }

  // Handle span drag over
  function handleSpanDragOver(event, index) {
    event.preventDefault();

    if (isDragging) {
      event.dataTransfer.dropEffect = "move";

      // Calculate drop position based on mouse position within the span
      const rect = event.currentTarget.getBoundingClientRect();
      const mouseX = event.clientX;
      const centerX = rect.left + rect.width / 2;

      // If mouse is in the left half, drop before this item, otherwise after
      dropPosition = mouseX < centerX ? index : index + 1;
    }
  }

  // Handle span drop
  async function handleSpanDrop(event, index) {
    event.preventDefault();
    event.stopPropagation();

    if (isDragging) {
      // Calculate final drop position based on mouse position
      const rect = event.currentTarget.getBoundingClientRect();
      const mouseX = event.clientX;
      const centerX = rect.left + rect.width / 2;

      dropPosition = mouseX < centerX ? index : index + 1;
      await performDrop();
    }
  }

  // Perform the actual drop operation
  async function performDrop() {
    if (
      !isDragging ||
      draggedHighlights.length === 0 ||
      dropPosition === null ||
      isDropping
    ) {
      return;
    }

    // Prevent concurrent drops
    isDropping = true;

    // Store current state before cleanup
    const draggedIds = [...draggedHighlights];
    const insertPosition = dropPosition;

    try {
      const currentHighlights = [...highlights]; // Create a copy

      // Validate that we have valid data
      if (currentHighlights.length === 0) {
        console.error("performDrop: no highlights to reorder");
        return;
      }

      // Create new order using a simpler, more reliable algorithm
      const newOrder = [];
      const draggedItems = [];
      const remainingItems = [];

      // Separate dragged items from remaining items, preserving order
      for (const highlight of currentHighlights) {
        if (draggedIds.includes(highlight.id)) {
          draggedItems.push(highlight);
        } else {
          remainingItems.push(highlight);
        }
      }

      // Validate we found all dragged items
      if (draggedItems.length !== draggedIds.length) {
        console.error("performDrop: could not find all dragged items", {
          expected: draggedIds.length,
          found: draggedItems.length,
        });
        return;
      }

      // Insert dragged items at the correct position
      const adjustedInsertPosition = Math.min(
        insertPosition,
        remainingItems.length
      );

      // Build the new order
      for (let i = 0; i <= remainingItems.length; i++) {
        if (i === adjustedInsertPosition) {
          newOrder.push(...draggedItems);
        }
        if (i < remainingItems.length) {
          newOrder.push(remainingItems[i]);
        }
      }

      // Validate the new order has the correct length
      if (newOrder.length !== currentHighlights.length) {
        console.error("performDrop: new order has wrong length", {
          original: currentHighlights.length,
          newOrder: newOrder.length,
        });
        return;
      }

      // Check if order actually changed
      const orderChanged = !newOrder.every(
        (item, index) => item.id === currentHighlights[index].id
      );

      if (!orderChanged) {
        return;
      }

      // Flatten consecutive newlines before calling the reorder callback
      const flattenedOrder = flattenConsecutiveNewlines(newOrder);

      // Call the reorder callback
      await onReorder(flattenedOrder);
    } catch (error) {
      console.error("performDrop: error during drop operation:", error);
    } finally {
      // Clean up drag state
      isDropping = false;
      handleDragEnd();
    }
  }

  // Handle drag end cleanup
  function handleDragEnd() {
    isDragging = false;
    draggedHighlights = [];
    dropPosition = null;
    dragStartPosition = null;
  }

  // Handle edit
  function handleEditHighlight(event, highlight) {
    if (enableEdit && onEdit) {
      onEdit(event, highlight);
    }
  }

  // Handle delete
  function handleDeleteHighlight(event, highlight) {
    if (enableDelete && onDelete) {
      onDelete(event, highlight);
    }
  }

  // Handle popover state change
  function handlePopoverStateChange(highlightId, isOpen) {
    if (onPopoverOpenChange) {
      onPopoverOpenChange(highlightId, isOpen);
    }
  }

  // Utility functions for newline handling with titles
  function isNewline(item) {
    return (
      item === "N" ||
      item === "n" ||
      (typeof item === "object" && item.type === "N") ||
      (typeof item === "object" && item.type === "newline")
    );
  }

  function getNewlineTitle(item) {
    if (
      typeof item === "object" &&
      (item.type === "N" || item.type === "newline")
    ) {
      return item.title || "";
    }
    return "";
  }

  function createNewline(title = "") {
    return {
      type: "newline",
      id: `newline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      title: title,
    };
  }

  // Convert database format to display format
  function convertDatabaseToDisplay(dbItem) {
    if (dbItem === "N") {
      return createNewline("");
    }
    if (typeof dbItem === "object" && dbItem.type === "N") {
      return createNewline(dbItem.title || "");
    }
    return dbItem;
  }

  // Utility function to flatten consecutive newlines
  function flattenConsecutiveNewlines(highlights) {
    if (!highlights || highlights.length <= 1) {
      return highlights;
    }

    const result = [];
    let lastWasNewline = false;

    for (const highlight of highlights) {
      if (isNewline(highlight)) {
        if (!lastWasNewline) {
          result.push(highlight);
          lastWasNewline = true;
        } else {
          // When flattening consecutive newlines, preserve the title of the last one
          // Only if the current newline has a title and the previous doesn't
          const lastNewline = result[result.length - 1];
          const currentTitle = getNewlineTitle(highlight);
          const lastTitle = getNewlineTitle(lastNewline);

          if (currentTitle && !lastTitle) {
            // Replace the last newline with the current one that has a title
            result[result.length - 1] = highlight;
          }
        }
        // Skip consecutive newlines
      } else {
        result.push(highlight);
        lastWasNewline = false;
      }
    }

    return result;
  }
</script>

<div
  class={containerClass}
  role="application"
  ondragover={(e) => handleContainerDragOver(e)}
  ondrop={(e) => handleContainerDrop(e)}
  ondragleave={handleContainerDragLeave}
>
  {#if highlights.length === 0}
    <div class="text-center py-4 text-muted-foreground">
      <p class="text-sm">No highlights to display.</p>
    </div>
  {:else}
    {#each highlights as item, index}
      {#if isNewline(item) && enableNewlines}
        <!-- New line creates actual line break -->
        <NewLineItem
          newlineItem={item}
          {index}
          {isDragging}
          isBeingDragged={isDragging && draggedHighlights.includes(item.id)}
          showDropIndicatorBefore={isDragging && dropPosition === index}
          showDropIndicatorAfter={false}
          {enableDrag}
          enableEdit={enableEdit}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDragOver={handleSpanDragOver}
          onDrop={handleSpanDrop}
          {onTitleChange}
        />
      {:else if !isNewline(item)}
        <!-- Add new line button before highlight (only if previous item is also a highlight) -->
        {#if enableNewlines && showAddNewLineButtons && !isDragging && index > 0 && !isNewline(highlights[index - 1])}
          <AddNewLineButton position={index} />
        {/if}

        <HighlightItem
          highlight={item}
          {index}
          isSelected={selectedHighlights.has(item.id)}
          {isDragging}
          isBeingDragged={isDragging &&
            draggedHighlights.includes(item.id) &&
            draggedHighlights[0] === item.id}
          showDropIndicatorBefore={isDragging && dropPosition === index}
          {enableDrag}
          {enableEdit}
          onSelect={handleHighlightSelect}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDragOver={handleSpanDragOver}
          onDrop={handleSpanDrop}
          onEdit={handleEditHighlight}
          onDelete={handleDeleteHighlight}
          popoverOpen={isPopoverOpen(item.id)}
          onPopoverOpenChange={(open) =>
            handlePopoverStateChange(item.id, open)}
          words={getHighlightWords ? getHighlightWords(item) : []}
        />
      {/if}
    {/each}

    <!-- Add new line button at the end (only if last item is a highlight) -->
    {#if enableNewlines && showAddNewLineButtons && !isDragging && highlights.length > 0 && !isNewline(highlights[highlights.length - 1])}
      <AddNewLineButton position={highlights.length} />
    {/if}

    <!-- Drop indicator at the end or after the last newline -->
    {#if isDragging && (dropPosition >= highlights.length || (highlights.length > 0 && isNewline(highlights[highlights.length - 1]) && dropPosition === highlights.length))}
      <span class="drop-indicator">|</span>
    {/if}
  {/if}
</div>

<style>
  /* Drop indicator styling */
  .drop-indicator {
    color: #3b82f6;
    font-weight: bold;
    animation: blink 1s infinite;
  }

  @keyframes blink {
    0%,
    50% {
      opacity: 1;
    }
    51%,
    100% {
      opacity: 0.3;
    }
  }
</style>
