# Design Document

## Overview

The empty state suggestion pills feature enhances the ChatInterface component by displaying interactive suggestion buttons when both the chat session is empty (no messages) and the text input field is empty. These pills provide users with quick-start prompts relevant to fitness coaching, complete with icons/emojis for visual appeal and better user experience.

## Architecture

### Component Structure

The suggestion pills will be implemented as a new component `SuggestionPills` that integrates into the existing `ChatInterface` component. The pills will be positioned between the messages area and the input area, appearing conditionally based on the empty state criteria.

### State Management

The suggestion pills visibility will be controlled by existing state in `ChatInterface`:
- `messages.length === 0` (empty session)
- `inputText === ''` (empty input field)

### Integration Points

- **ChatInterface.tsx**: Main integration point where `SuggestionPills` component will be rendered
- **Existing styling system**: Leverages current Tailwind CSS classes and design patterns
- **Input handling**: Integrates with existing `setInputText` state setter

## Components and Interfaces

### SuggestionPills Component

```typescript
interface SuggestionPill {
  id: string;
  text: string;
  icon: string; // emoji or icon character
  category: 'training' | 'goals' | 'progress' | 'help';
}

interface SuggestionPillsProps {
  onPillClick: (text: string) => void;
  className?: string;
}
```

### Predefined Suggestions

The component will include a curated set of coaching-relevant suggestions:

**Training Category (💪, 🏃‍♂️, 🏋️‍♀️)**
- "💪 Help me plan my next training week"
- "🏃‍♂️ Analyze my recent running performance"
- "🏋️‍♀️ What strength training should I focus on?"

**Goals Category (🎯, 📈, ⭐)**
- "🎯 Help me set realistic training goals"
- "📈 How can I improve my race times?"

**Progress Category (📊, 📅, 🔥)**
- "📊 Show me my training progress trends"
- "🔥 What's my current fitness level?"

**Help Category (❓, 💡)**
- "❓ What can you help me with?"
- "💡 Give me training tips for beginners"

## Data Models

### Suggestion Configuration

```typescript
const SUGGESTION_PILLS: SuggestionPill[] = [
  {
    id: 'training-plan',
    text: 'Help me plan my next training week',
    icon: '💪',
    category: 'training'
  },
  {
    id: 'performance-analysis',
    text: 'Analyze my recent running performance',
    icon: '🏃‍♂️',
    category: 'training'
  },
  // ... additional suggestions
];
```

## User Interface Design

### Visual Layout

```
┌─────────────────────────────────────────┐
│              Messages Area              │
│                                         │
│  [Empty state message when no msgs]     │
│                                         │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│           Suggestion Pills              │
│  [💪 Help me plan...]  [🏃‍♂️ Analyze...]   │
│  [🎯 Help me set...]   [📊 Show me...]    │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│              Input Area                 │
│  [Text Input Field]        [Send Button]│
└─────────────────────────────────────────┘
```

### Styling Specifications

**Container**
- Padding: `p-4` (consistent with input area)
- Background: `bg-gray-50` (matches main background)
- Border: `border-t border-gray-200` (subtle separation)

**Individual Pills**
- Background: `bg-white` with `hover:bg-gray-50`
- Border: `border border-gray-200` with `hover:border-gray-300`
- Padding: `px-4 py-2`
- Border radius: `rounded-lg`
- Text: `text-sm text-gray-700`
- Transition: `transition-colors duration-200`
- Shadow: `shadow-sm hover:shadow-md`

**Responsive Layout**
- Mobile: Single column, full-width pills
- Tablet/Desktop: Grid layout with 2 columns
- Grid gap: `gap-2`

### Accessibility Features

- **Keyboard Navigation**: Pills are focusable with `tabindex="0"`
- **ARIA Labels**: Each pill has `role="button"` and descriptive `aria-label`
- **Focus Indicators**: Clear focus ring with `focus:ring-2 focus:ring-blue-500`
- **Screen Reader Support**: Proper semantic markup and labels

## Interaction Design

### User Flow

1. **Initial State**: User opens chat with empty session and empty input
2. **Display**: Suggestion pills appear below empty state message
3. **Interaction**: User clicks/taps a suggestion pill
4. **Action**: Text populates input field, pills hide, input gains focus
5. **Continuation**: User can edit text before sending or send immediately

### State Transitions

```
Empty Session + Empty Input → Show Pills
Empty Session + Text Input → Hide Pills
Has Messages + Any Input → Hide Pills
Empty Session + Clear Input → Show Pills
```

### Animation

- **Fade In**: Pills appear with `opacity-0` to `opacity-100` transition
- **Fade Out**: Pills disappear with smooth opacity transition
- **Hover Effects**: Subtle scale and shadow changes on hover
- **Duration**: 200ms for smooth, responsive feel

## Error Handling

### Graceful Degradation

- If suggestion data fails to load, component renders nothing
- No error boundaries needed as this is a pure UI enhancement
- Fallback to standard empty state if component fails

### Edge Cases

- **Rapid Typing**: Debounced input changes prevent flickering
- **Network Issues**: Pills work offline as they're static data
- **Accessibility**: Keyboard users can skip pills and use input directly

## Testing Strategy

### Unit Tests

**SuggestionPills Component**
- Renders correct number of pills
- Displays proper icons and text
- Handles click events correctly
- Applies correct CSS classes
- Supports keyboard navigation

**Integration with ChatInterface**
- Shows pills when session empty and input empty
- Hides pills when input has text
- Hides pills when messages exist
- Populates input correctly on pill click

### Visual Tests

- Responsive layout across screen sizes
- Hover and focus states
- Color contrast compliance
- Icon rendering consistency

### Accessibility Tests

- Screen reader compatibility
- Keyboard navigation flow
- Focus management
- ARIA label accuracy

### User Experience Tests

- Pill click populates input correctly
- Input focus after pill selection
- Smooth show/hide transitions
- Mobile touch target sizes (minimum 44px)

## Performance Considerations

### Optimization Strategies

- **Static Data**: Suggestions are compile-time constants
- **Conditional Rendering**: Component only renders when needed
- **Minimal Re-renders**: Uses React.memo for optimization
- **CSS Transitions**: Hardware-accelerated animations

### Bundle Impact

- Minimal JavaScript footprint (~2KB)
- No external dependencies
- Leverages existing Tailwind classes
- Emoji characters (no icon library needed)

## Implementation Notes

### Integration Points

1. **ChatInterface.tsx**: Add conditional rendering logic
2. **New Component**: Create `SuggestionPills.tsx` in components directory
3. **Styling**: Use existing Tailwind design system
4. **Testing**: Add tests in `__tests__` directory

### Development Approach

- Create component in isolation first
- Integrate with ChatInterface
- Add comprehensive tests
- Ensure accessibility compliance
- Test responsive behavior