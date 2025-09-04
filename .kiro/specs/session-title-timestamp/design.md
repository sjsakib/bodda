# Design Document

## Overview

This feature enhances the SessionSidebar component to display formatted timestamps as session titles instead of the current generic titles. The implementation will be entirely frontend-based, utilizing the existing `created_at` timestamp from the Session interface to generate human-readable titles in the format "3 Sep, 08:20 pm".

## Architecture

### Current State
- SessionSidebar component displays `session.title` and `session.created_at` as separate elements
- Session interface includes `created_at` as an ISO string timestamp
- Date formatting is currently done with `toLocaleDateString()` for a secondary display

### Proposed Changes
- Replace the primary session title display with a formatted timestamp
- Create a utility function for consistent timestamp formatting
- Maintain accessibility and responsive design principles
- Remove dependency on the `session.title` field for display purposes

## Components and Interfaces

### 1. Date Formatting Utility

**Location:** `frontend/src/utils/dateFormatting.ts`

```typescript
interface TimestampFormatOptions {
  includeYear?: boolean
  use24Hour?: boolean
}

function formatSessionTimestamp(
  timestamp: string, 
  options?: TimestampFormatOptions
): string
```

**Functionality:**
- Parse ISO timestamp strings
- Format to "DD MMM, HH:MM am/pm" pattern
- Handle year inclusion logic (current year vs. previous years)
- Support both 12-hour and 24-hour formats (defaulting to 12-hour)
- Handle timezone conversion to local time

### 2. SessionSidebar Component Updates

**File:** `frontend/src/components/SessionSidebar.tsx`

**Changes:**
- Import and use the new formatting utility
- Replace `session.title` display with formatted timestamp
- Update the secondary date display or remove it if redundant
- Maintain existing accessibility attributes
- Preserve hover states and truncation behavior

### 3. Type Definitions

**File:** `frontend/src/types.ts` (if needed)

No new types required - leveraging existing Session interface with `created_at: string`.

## Data Models

### Session Interface (No Changes)
The existing Session interface already provides the necessary data:

```typescript
interface Session {
  id: string
  user_id: string
  title: string        // Will not be used for display
  created_at: string   // ISO timestamp - primary data source
  updated_at: string
}
```

### Formatting Logic

**Input:** ISO timestamp string (e.g., "2024-09-03T20:20:00Z")
**Output:** Formatted string (e.g., "3 Sep, 08:20 pm")

**Rules:**
1. Current year sessions: "3 Sep, 08:20 pm"
2. Previous year sessions: "3 Sep 2023, 08:20 pm"
3. Use local timezone for display
4. Use 12-hour format with lowercase am/pm
5. Use abbreviated month names (Jan, Feb, Mar, etc.)

## Error Handling

### Invalid Timestamps
- **Scenario:** Malformed or missing `created_at` values
- **Fallback:** Display "Invalid Date" or use session ID as fallback
- **Logging:** Log warnings for debugging purposes

### Timezone Issues
- **Scenario:** Browser timezone detection fails
- **Fallback:** Use UTC time with "(UTC)" suffix
- **User Experience:** Maintain consistent formatting even with fallback

### Formatting Failures
- **Scenario:** Date formatting utility throws errors
- **Fallback:** Use basic `toLocaleDateString()` as backup
- **Recovery:** Graceful degradation without breaking the component

## Testing Strategy

### Unit Tests
1. **Date Formatting Utility Tests**
   - Test various timestamp formats and edge cases
   - Test year inclusion logic (current vs. previous years)
   - Test timezone handling
   - Test invalid input handling

2. **SessionSidebar Component Tests**
   - Test timestamp display integration
   - Test accessibility attributes
   - Test responsive behavior
   - Test error fallback scenarios

### Integration Tests
1. **Session List Rendering**
   - Test with multiple sessions from different time periods
   - Test with mixed valid/invalid timestamps
   - Test timezone consistency across sessions

### Accessibility Tests
1. **Screen Reader Compatibility**
   - Verify aria-labels include full timestamp information
   - Test with various screen reader software
   - Ensure semantic HTML structure

2. **Keyboard Navigation**
   - Verify existing keyboard navigation remains functional
   - Test focus management with new content structure

### Visual Regression Tests
1. **Layout Consistency**
   - Test text overflow handling
   - Test responsive breakpoints
   - Test with various timestamp lengths

## Implementation Considerations

### Performance
- Formatting utility should be lightweight and fast
- Consider memoization for repeated formatting of the same timestamps
- Avoid unnecessary re-renders when timestamp formatting changes

### Internationalization (Future)
- Design formatting utility to support future i18n requirements
- Use consistent date formatting patterns that can be localized
- Consider RTL language support in layout

### Browser Compatibility
- Use standard JavaScript Date APIs for broad compatibility
- Test across major browsers (Chrome, Firefox, Safari, Edge)
- Ensure graceful degradation for older browsers

### Responsive Design
- Maintain existing responsive behavior
- Ensure timestamp text fits within mobile sidebar constraints
- Consider truncation strategies for very long formatted dates