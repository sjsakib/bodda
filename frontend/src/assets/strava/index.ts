// Strava Brand Assets
// Official Strava brand assets following brand guidelines
// Colors: Orange (#FC4C02), White, Black
// Minimum size: 16px height

export interface BrandAsset {
  id: string;
  type: 'logo' | 'icon' | 'wordmark';
  variant: 'orange' | 'white' | 'black';
  format: 'svg' | 'png';
  size: 'small' | 'medium' | 'large';
  url: string;
  minWidth: number;
  minHeight: number;
}

// Asset imports
import stravaLogoOrange from './strava-logo-orange.svg';
import stravaLogoWhite from './strava-logo-white.svg';
import stravaWordmarkOrange from './strava-wordmark-orange.svg';
import stravaWordmarkWhite from './strava-wordmark-white.svg';
import stravaIconOrange from './strava-icon-orange.svg';
import stravaIconWhite from './strava-icon-white.svg';

// Brand asset definitions
export const STRAVA_BRAND_ASSETS: BrandAsset[] = [
  {
    id: 'strava-logo-orange-svg',
    type: 'logo',
    variant: 'orange',
    format: 'svg',
    size: 'large',
    url: stravaLogoOrange,
    minWidth: 32,
    minHeight: 32,
  },
  {
    id: 'strava-logo-white-svg',
    type: 'logo',
    variant: 'white',
    format: 'svg',
    size: 'large',
    url: stravaLogoWhite,
    minWidth: 32,
    minHeight: 32,
  },
  {
    id: 'strava-wordmark-orange-svg',
    type: 'wordmark',
    variant: 'orange',
    format: 'svg',
    size: 'medium',
    url: stravaWordmarkOrange,
    minWidth: 100,
    minHeight: 20,
  },
  {
    id: 'strava-wordmark-white-svg',
    type: 'wordmark',
    variant: 'white',
    format: 'svg',
    size: 'medium',
    url: stravaWordmarkWhite,
    minWidth: 100,
    minHeight: 20,
  },
  {
    id: 'strava-icon-orange-svg',
    type: 'icon',
    variant: 'orange',
    format: 'svg',
    size: 'small',
    url: stravaIconOrange,
    minWidth: 16,
    minHeight: 16,
  },
  {
    id: 'strava-icon-white-svg',
    type: 'icon',
    variant: 'white',
    format: 'svg',
    size: 'small',
    url: stravaIconWhite,
    minWidth: 16,
    minHeight: 16,
  },
];

// Brand colors
export const STRAVA_COLORS = {
  ORANGE: '#FC4C02',
  WHITE: '#FFFFFF',
  BLACK: '#000000',
} as const;

// Helper functions
export const getStravaAsset = (
  type: BrandAsset['type'],
  variant: BrandAsset['variant'] = 'orange'
): BrandAsset | undefined => {
  return STRAVA_BRAND_ASSETS.find(
    asset => asset.type === type && asset.variant === variant
  );
};

export const getStravaLogo = (variant: BrandAsset['variant'] = 'orange'): BrandAsset | undefined => {
  return getStravaAsset('logo', variant);
};

export const getStravaWordmark = (variant: BrandAsset['variant'] = 'orange'): BrandAsset | undefined => {
  return getStravaAsset('wordmark', variant);
};

export const getStravaIcon = (variant: BrandAsset['variant'] = 'orange'): BrandAsset | undefined => {
  return getStravaAsset('icon', variant);
};

// Attribution requirements
export interface AttributionRequirements {
  required: boolean;
  text: string;
  logoRequired: boolean;
  linkRequired: boolean;
  placement: string[];
}

export const STRAVA_ATTRIBUTION_REQUIREMENTS: Record<string, AttributionRequirements> = {
  activity_data: {
    required: true,
    text: 'Powered by Strava',
    logoRequired: true,
    linkRequired: true,
    placement: ['footer', 'inline', 'badge'],
  },
  segment_data: {
    required: true,
    text: 'Segment data by Strava',
    logoRequired: true,
    linkRequired: true,
    placement: ['inline', 'badge'],
  },
  athlete_data: {
    required: true,
    text: 'Powered by Strava',
    logoRequired: true,
    linkRequired: true,
    placement: ['footer', 'inline'],
  },
  general: {
    required: true,
    text: 'Powered by Strava',
    logoRequired: true,
    linkRequired: true,
    placement: ['footer'],
  },
};

export const getAttributionRequirements = (dataType: string): AttributionRequirements => {
  return STRAVA_ATTRIBUTION_REQUIREMENTS[dataType] || STRAVA_ATTRIBUTION_REQUIREMENTS.general;
};