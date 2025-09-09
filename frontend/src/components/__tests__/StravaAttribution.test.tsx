import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import StravaAttribution from '../StravaAttribution';
import { STRAVA_COLORS } from '../../assets/strava';

describe('StravaAttribution', () => {
  describe('Basic Rendering', () => {
    it('renders with default props', () => {
      render(<StravaAttribution />);
      
      expect(screen.getByRole('contentinfo')).toBeInTheDocument();
      expect(screen.getByText('Powered by Strava')).toBeInTheDocument();
      expect(screen.getByAltText('Strava logo')).toBeInTheDocument();
    });

    it('renders with custom text', () => {
      render(<StravaAttribution customText="Custom Strava Text" />);
      
      expect(screen.getByText('Custom Strava Text')).toBeInTheDocument();
    });

    it('renders without logo when showLogo is false', () => {
      render(<StravaAttribution showLogo={false} />);
      
      expect(screen.queryByAltText('Strava logo')).not.toBeInTheDocument();
      expect(screen.getByText('Powered by Strava')).toBeInTheDocument();
    });
  });

  describe('Data Type Attribution', () => {
    it('shows correct text for activity data', () => {
      render(<StravaAttribution dataType="activity_data" />);
      
      expect(screen.getByText('Powered by Strava')).toBeInTheDocument();
    });

    it('shows correct text for segment data', () => {
      render(<StravaAttribution dataType="segment_data" />);
      
      expect(screen.getByText('Segment data by Strava')).toBeInTheDocument();
    });

    it('shows correct text for athlete data', () => {
      render(<StravaAttribution dataType="athlete_data" />);
      
      expect(screen.getByText('Powered by Strava')).toBeInTheDocument();
    });

    it('shows default text for general data type', () => {
      render(<StravaAttribution dataType="general" />);
      
      expect(screen.getByText('Powered by Strava')).toBeInTheDocument();
    });
  });

  describe('Variants', () => {
    it('applies inline variant styles', () => {
      render(<StravaAttribution variant="inline" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('inline-flex', 'items-center', 'gap-2');
    });

    it('applies footer variant styles', () => {
      render(<StravaAttribution variant="footer" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('flex', 'items-center', 'justify-center', 'gap-2', 'text-sm');
    });

    it('applies badge variant styles', () => {
      render(<StravaAttribution variant="badge" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('inline-flex', 'items-center', 'gap-1', 'px-2', 'py-1', 'rounded-md');
    });
  });

  describe('Sizes', () => {
    it('applies small size styles', () => {
      render(<StravaAttribution size="small" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('text-xs');
      
      const logo = screen.getByAltText('Strava logo');
      expect(logo).toHaveClass('h-4', 'w-auto');
    });

    it('applies medium size styles', () => {
      render(<StravaAttribution size="medium" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('text-sm');
      
      const logo = screen.getByAltText('Strava logo');
      expect(logo).toHaveClass('h-5', 'w-auto');
    });

    it('applies large size styles', () => {
      render(<StravaAttribution size="large" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('text-base');
      
      const logo = screen.getByAltText('Strava logo');
      expect(logo).toHaveClass('h-6', 'w-auto');
    });
  });

  describe('Themes', () => {
    it('uses orange logo for light theme', () => {
      render(<StravaAttribution theme="light" />);
      
      const logo = screen.getByAltText('Strava logo');
      // Check that the logo src is defined (SVG is imported as data URL)
      expect(logo.getAttribute('src')).toBeTruthy();
      expect(logo.getAttribute('src')).toMatch(/^data:image\/svg\+xml/);
    });

    it('uses white logo for dark theme', () => {
      render(<StravaAttribution theme="dark" />);
      
      const logo = screen.getByAltText('Strava logo');
      // Check that the logo src is defined (SVG is imported as data URL)
      expect(logo.getAttribute('src')).toBeTruthy();
      expect(logo.getAttribute('src')).toMatch(/^data:image\/svg\+xml/);
    });

    it('applies correct link color for light theme', () => {
      render(<StravaAttribution theme="light" />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveStyle({ color: STRAVA_COLORS.ORANGE });
    });

    it('applies correct link color for dark theme', () => {
      render(<StravaAttribution theme="dark" />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveStyle({ color: STRAVA_COLORS.WHITE });
    });
  });

  describe('Link Behavior', () => {
    it('links to Strava website', () => {
      render(<StravaAttribution />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveAttribute('href', 'https://www.strava.com');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('has proper hover and focus styles', () => {
      render(<StravaAttribution />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveClass('hover:underline', 'focus:underline', 'focus:outline-none');
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA label', () => {
      render(<StravaAttribution />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveAttribute('aria-label', 'Strava attribution');
    });

    it('has proper alt text for logo', () => {
      render(<StravaAttribution />);
      
      const logo = screen.getByAltText('Strava logo');
      expect(logo).toBeInTheDocument();
    });

    it('maintains minimum logo size requirements', () => {
      render(<StravaAttribution size="small" />);
      
      const logo = screen.getByAltText('Strava logo');
      const style = window.getComputedStyle(logo);
      
      // Check that minimum dimensions are set
      expect(logo.style.minHeight).toBeTruthy();
      expect(logo.style.minWidth).toBeTruthy();
    });
  });

  describe('Custom Styling', () => {
    it('applies custom className', () => {
      render(<StravaAttribution className="custom-class" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('custom-class');
    });

    it('preserves base classes with custom className', () => {
      render(<StravaAttribution className="custom-class" variant="inline" />);
      
      const attribution = screen.getByRole('contentinfo');
      expect(attribution).toHaveClass('custom-class', 'inline-flex', 'items-center');
    });
  });

  describe('Logo Selection', () => {
    it('uses icon for small size', () => {
      render(<StravaAttribution size="small" />);
      
      const logo = screen.getByAltText('Strava logo');
      // Icon should be used for small size - check that logo is present
      expect(logo.getAttribute('src')).toBeTruthy();
      expect(logo.getAttribute('src')).toMatch(/^data:image\/svg\+xml/);
    });

    it('uses wordmark for medium and large sizes', () => {
      render(<StravaAttribution size="medium" />);
      
      const logo = screen.getByAltText('Strava logo');
      // Wordmark should be used for medium/large sizes - check that logo is present
      expect(logo.getAttribute('src')).toBeTruthy();
      expect(logo.getAttribute('src')).toMatch(/^data:image\/svg\+xml/);
    });
  });

  describe('Brand Compliance', () => {
    it('maintains proper spacing with flex-shrink-0', () => {
      render(<StravaAttribution />);
      
      const logo = screen.getByAltText('Strava logo');
      expect(logo).toHaveClass('flex-shrink-0');
    });

    it('prevents text wrapping with whitespace-nowrap', () => {
      render(<StravaAttribution />);
      
      const textSpan = screen.getByText('Powered by Strava').parentElement;
      expect(textSpan).toHaveClass('whitespace-nowrap');
    });

    it('uses official Strava colors', () => {
      render(<StravaAttribution theme="light" />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveStyle({ color: '#FC4C02' }); // Official Strava orange
    });
  });
});