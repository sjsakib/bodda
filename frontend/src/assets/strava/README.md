# Strava Brand Assets

This directory contains official Strava brand assets for use in compliance with Strava's Brand Guidelines.

## Available Assets

### Logos (Full Circle Logo)
- `strava-logo-orange.svg` - Orange logo on transparent background
- `strava-logo-white.svg` - White logo on transparent background

### Wordmarks (Text Logo)
- `strava-wordmark-orange.svg` - Orange wordmark on transparent background
- `strava-wordmark-white.svg` - White wordmark on transparent background

### Icons (Small Symbol)
- `strava-icon-orange.svg` - Orange icon for compact spaces
- `strava-icon-white.svg` - White icon for compact spaces

## Brand Guidelines Compliance

### Colors
- **Orange**: `#FC4C02` (Primary brand color)
- **White**: `#FFFFFF` (For dark backgrounds)
- **Black**: `#000000` (Alternative, use sparingly)

### Minimum Sizes
- **Icons**: 16px height minimum
- **Wordmarks**: 20px height minimum
- **Logos**: 32px height minimum

### Usage Requirements
1. Always maintain proper aspect ratio
2. Ensure adequate clear space around logos
3. Use appropriate variant for background contrast
4. Include "Powered by Strava" attribution text
5. Link to strava.com when interactive

### Attribution Requirements
Different data types require specific attribution:

- **Activity Data**: "Powered by Strava" with logo
- **Segment Data**: "Segment data by Strava" with logo
- **Athlete Data**: "Powered by Strava" with logo
- **General**: "Powered by Strava" with logo

## Implementation

Import and use assets through the index.ts module:

```typescript
import { getStravaLogo, getStravaWordmark, getStravaIcon, STRAVA_COLORS } from './assets/strava';

// Get orange logo
const orangeLogo = getStravaLogo('orange');

// Get white wordmark for dark backgrounds
const whiteWordmark = getStravaWordmark('white');

// Get small icon
const icon = getStravaIcon('orange');
```

## Legal Compliance

These assets are used in compliance with:
- Strava API Terms of Service
- Strava Brand Guidelines
- Strava Developer Agreement

All usage must include proper attribution and linking as specified in the brand guidelines.