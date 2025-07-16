# Claude Development Guidelines

## ⚠️ IMPORTANT: This project uses Svelte 5

### Svelte 5 Runes Mode Requirements
- **DO NOT use `export let`** - Use `$props()` instead
- **DO NOT use `$:`** - Use `$derived()` or `$effect()` instead
- **DO NOT use `let` for reactive state** - Use `$state()` instead
- **DO NOT use stores** - Use runes instead

### Svelte 5 Syntax Examples
```typescript
// Props - OLD (Svelte 4)
export let value = 'default';

// Props - NEW (Svelte 5)
let { value = 'default' } = $props();

// Reactive statements - OLD
$: doubled = count * 2;

// Reactive statements - NEW
const doubled = $derived(count * 2);

// Reactive state - OLD
let count = 0;

// Reactive state - NEW
let count = $state(0);
```

## Component Library: shadcn-svelte

**ALWAYS prefer components from [shadcn-svelte](https://www.shadcn-svelte.com/docs/components) instead of creating custom components.**

If a component doesn't exist in the project, install it using:
```bash
npx shadcn-svelte@latest add [component-name]
```

### Available shadcn-svelte Components

#### Layout & Navigation
- **Accordion** - Collapsible content sections
- **Breadcrumb** - Navigation breadcrumbs
- **Navigation Menu** - Complex navigation with dropdowns
- **Pagination** - Page navigation controls
- **Tabs** - Tabbed content interface
- **Sidebar** - Application sidebar layout

#### Form Components
- **Button** - Primary action buttons ✅ (already installed)
- **Input** - Text input fields
- **Textarea** - Multi-line text input
- **Select** - Dropdown selection
- **Combobox** - Searchable select dropdown
- **Checkbox** - Boolean input controls
- **Radio Group** - Single selection from options
- **Switch** - Toggle controls
- **Slider** - Range input controls
- **Form** - Form validation and layout
- **Label** - Form field labels

#### Display Components
- **Badge** - Status and category indicators ✅ (already installed)
- **Card** - Content containers ✅ (already installed)
- **Avatar** - User profile images
- **Separator** - Visual content dividers
- **Skeleton** - Loading placeholders
- **Progress** - Progress indicators
- **Table** - Data tables
- **Calendar** - Date selection
- **Command** - Command palette interface

#### Feedback & Overlays
- **Alert** - Important messages
- **Alert Dialog** - Modal confirmations
- **Dialog** - Modal dialogs
- **Sheet** - Slide-out panels
- **Popover** - Floating content panels
- **Tooltip** - Hover information
- **Toast** - Notification messages
- **Drawer** - Mobile-friendly slide panels

#### Advanced Components
- **Data Table** - Advanced table with sorting/filtering
- **Date Picker** - Calendar-based date selection
- **Carousel** - Image/content sliders
- **Chart** - Data visualization
- **Sonner** - Modern toast notifications
- **Resizable** - Resizable panel layouts

### Installation Examples
```bash
# Install commonly needed components
npx shadcn-svelte@latest add input
npx shadcn-svelte@latest add select
npx shadcn-svelte@latest add textarea
npx shadcn-svelte@latest add dialog
npx shadcn-svelte@latest add toast
npx shadcn-svelte@latest add alert
```

### Design System Integration
- All shadcn-svelte components work seamlessly with the existing Tailwind CSS setup
- Components automatically respect the configured theme (dark/light mode)
- Consistent styling and behavior across the application
- Built-in accessibility features and best practices

### Component Import Pattern
```typescript
import { Button } from '$lib/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
import { Badge } from '$lib/components/ui/badge';
```

## Performance Guidelines

### Animation Optimization
- Avoid duplicate animations when possible
- Use responsive design patterns instead of separate mobile/desktop implementations
- Prefer CSS animations for simple effects, GSAP for complex interactions
- Consider animation performance impact on mobile devices

### GSAP Best Practices
- Use direct DOM manipulation for colors to avoid parsing issues
- Implement error handling for animation functions
- Use specific selectors rather than broad class queries
- Clean up animations appropriately to prevent memory leaks