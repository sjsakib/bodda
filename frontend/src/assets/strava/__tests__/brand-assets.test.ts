import { describe, it, expect } from 'vitest';
import {
  STRAVA_BRAND_ASSETS,
  STRAVA_COLORS,
  getStravaAsset,
  getStravaLogo,
  getStravaWordmark,
  getStravaIcon,
  getAttributionRequirements,
  STRAVA_ATTRIBUTION_REQUIREMENTS,
} from '../index';

describe('Strava Brand Assets', () => {
  it('should have all required brand assets', () => {
    expect(STRAVA_BRAND_ASSETS).toHaveLength(6);
    
    // Check that we have all asset types
    const types = STRAVA_BRAND_ASSETS.map(asset => asset.type);
    expect(types).toContain('logo');
    expect(types).toContain('wordmark');
    expect(types).toContain('icon');
    
    // Check that we have both orange and white variants
    const variants = STRAVA_BRAND_ASSETS.map(asset => asset.variant);
    expect(variants).toContain('orange');
    expect(variants).toContain('white');
  });

  it('should have correct brand colors', () => {
    expect(STRAVA_COLORS.ORANGE).toBe('#FC4C02');
    expect(STRAVA_COLORS.WHITE).toBe('#FFFFFF');
    expect(STRAVA_COLORS.BLACK).toBe('#000000');
  });

  it('should enforce minimum size requirements', () => {
    STRAVA_BRAND_ASSETS.forEach(asset => {
      expect(asset.minWidth).toBeGreaterThan(0);
      expect(asset.minHeight).toBeGreaterThan(0);
      
      // Check specific minimum sizes
      if (asset.type === 'icon') {
        expect(asset.minHeight).toBeGreaterThanOrEqual(16);
      } else if (asset.type === 'wordmark') {
        expect(asset.minHeight).toBeGreaterThanOrEqual(20);
      } else if (asset.type === 'logo') {
        expect(asset.minHeight).toBeGreaterThanOrEqual(32);
      }
    });
  });

  it('should provide helper functions for asset retrieval', () => {
    // Test getStravaLogo
    const orangeLogo = getStravaLogo('orange');
    expect(orangeLogo).toBeDefined();
    expect(orangeLogo?.type).toBe('logo');
    expect(orangeLogo?.variant).toBe('orange');

    const whiteLogo = getStravaLogo('white');
    expect(whiteLogo).toBeDefined();
    expect(whiteLogo?.variant).toBe('white');

    // Test getStravaWordmark
    const orangeWordmark = getStravaWordmark('orange');
    expect(orangeWordmark).toBeDefined();
    expect(orangeWordmark?.type).toBe('wordmark');

    // Test getStravaIcon
    const orangeIcon = getStravaIcon('orange');
    expect(orangeIcon).toBeDefined();
    expect(orangeIcon?.type).toBe('icon');
  });

  it('should provide attribution requirements', () => {
    expect(STRAVA_ATTRIBUTION_REQUIREMENTS).toBeDefined();
    
    // Test specific data types
    const activityAttribution = getAttributionRequirements('activity_data');
    expect(activityAttribution.required).toBe(true);
    expect(activityAttribution.text).toBe('Powered by Strava');
    expect(activityAttribution.logoRequired).toBe(true);
    expect(activityAttribution.linkRequired).toBe(true);

    const segmentAttribution = getAttributionRequirements('segment_data');
    expect(segmentAttribution.text).toBe('Segment data by Strava');

    // Test fallback to general
    const unknownAttribution = getAttributionRequirements('unknown_type');
    expect(unknownAttribution.text).toBe('Powered by Strava');
  });

  it('should have valid asset URLs', () => {
    STRAVA_BRAND_ASSETS.forEach(asset => {
      expect(asset.url).toBeDefined();
      expect(typeof asset.url).toBe('string');
      expect(asset.url.length).toBeGreaterThan(0);
    });
  });

  it('should have unique asset IDs', () => {
    const ids = STRAVA_BRAND_ASSETS.map(asset => asset.id);
    const uniqueIds = new Set(ids);
    expect(uniqueIds.size).toBe(ids.length);
  });
});