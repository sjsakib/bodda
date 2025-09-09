# StravaAttribution Component

A React component that provides compliant Strava branding and attribution according to Strava's Brand Guidelines and API Terms of Service.

## Features

- ✅ Official Strava branding compliance
- ✅ Multiple variants (inline, footer, badge)
- ✅ Responsive sizing (small, medium, large)
- ✅ Theme support (light/dark)
- ✅ Accessibility compliant
- ✅ Proper attribution text for different data types
- ✅ Links to Strava website
- ✅ TypeScript support

## Basic Usage

```tsx
import StravaAttribution from './components/StravaAttribution';

// Basic usage with default settings
<StravaAttribution />

// For activity data
<StravaAttribution dataType="activity_data" />

// For segment data
<StravaAttribution dataType="segment_data" />
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `dataType` | `'activity_data' \| 'segment_data' \| 'athlete_data' \| 'general'` | `'general'` | Type of data being displayed |
| `variant` | `'inline' \| 'footer' \| 'badge'` | `'inline'` | Visual variant of the attribution |
| `size` | `'small' \| 'medium' \| 'large'` | `'medium'` | Size of the attribution |
| `theme` | `'light' \| 'dark'` | `'light'` | Theme for logo and text color |
| `className` | `string` | `''` | Additional CSS classes |
| `showLogo` | `boolean` | `true` | Whether to show the Strava logo |
| `customText` | `string` | `undefined` | Custom attribution text (overrides default) |

## Examples

### Different Data Types

```tsx
// General usage
<StravaAttribution dataType="general" />
// Shows: "Powered by Strava"

// Activity data
<StravaAttribution dataType="activity_data" />
// Shows: "Powered by Strava"

// Segment data
<StravaAttribution dataType="segment_data" />
// Shows: "Segment data by Strava"

// Athlete data
<StravaAttribution dataType="athlete_data" />
// Shows: "Powered by Strava"
```

### Variants

```tsx
// Inline variant (default) - for use within content
<div className="activity-card">
  <h3>Morning Run</h3>
  <p>5.2 miles in 42:15</p>
  <StravaAttribution variant="inline" size="small" />
</div>

// Footer variant - for page footers
<footer>
  <StravaAttribution variant="footer" />
</footer>

// Badge variant - compact display
<div className="data-source">
  <StravaAttribution variant="badge" size="small" />
</div>
```

### Themes

```tsx
// Light theme (default)
<div className="bg-white p-4">
  <StravaAttribution theme="light" />
</div>

// Dark theme
<div className="bg-gray-900 p-4">
  <StravaAttribution theme="dark" />
</div>
```

### Sizes

```tsx
// Small - for compact spaces
<StravaAttribution size="small" />

// Medium - default size
<StravaAttribution size="medium" />

// Large - for prominent display
<StravaAttribution size="large" />
```

### Custom Styling

```tsx
// With custom classes
<StravaAttribution 
  className="mt-4 opacity-75" 
  variant="inline" 
/>

// Text only (no logo)
<StravaAttribution 
  showLogo={false}
  customText="Data from Strava"
/>
```

## Real-world Usage Examples

### Activity Display Page

```tsx
function ActivityPage({ activity }) {
  return (
    <div className="activity-page">
      <h1>{activity.name}</h1>
      <div className="activity-stats">
        {/* Activity content */}
      </div>
      
      {/* Attribution at bottom of activity data */}
      <StravaAttribution 
        dataType="activity_data"
        variant="inline"
        size="small"
        className="mt-6"
      />
    </div>
  );
}
```

### Segment Leaderboard

```tsx
function SegmentLeaderboard({ segment }) {
  return (
    <div className="segment-leaderboard">
      <h2>{segment.name}</h2>
      <div className="leaderboard">
        {/* Leaderboard content */}
      </div>
      
      {/* Attribution for segment data */}
      <StravaAttribution 
        dataType="segment_data"
        variant="badge"
        size="small"
        className="mt-4"
      />
    </div>
  );
}
```

### App Footer

```tsx
function AppFooter() {
  return (
    <footer className="app-footer bg-gray-100 py-4">
      <div className="container mx-auto">
        {/* Other footer content */}
        
        {/* Strava attribution */}
        <StravaAttribution 
          variant="footer"
          size="small"
          className="mt-4"
        />
      </div>
    </footer>
  );
}
```

## Brand Compliance

This component ensures compliance with:

- **Strava Brand Guidelines**: Uses official logos, colors, and spacing
- **Strava API Terms**: Includes required attribution text and links
- **Accessibility Standards**: Proper ARIA labels and keyboard navigation
- **Minimum Size Requirements**: Maintains logo minimum dimensions

## Colors Used

- **Strava Orange**: `#FC4C02` (Primary brand color)
- **White**: `#FFFFFF` (For dark backgrounds)
- **Black**: `#000000` (Alternative text color)

## Accessibility

- Uses semantic HTML with proper ARIA labels
- Maintains color contrast ratios
- Supports keyboard navigation
- Screen reader friendly
- Respects minimum logo size requirements

## Browser Support

Works in all modern browsers that support:
- CSS Flexbox
- SVG images
- ES6+ JavaScript features