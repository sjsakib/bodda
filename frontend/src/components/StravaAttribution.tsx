import React from 'react';
import { 
  getStravaWordmark, 
  getStravaIcon, 
  getAttributionRequirements,
  STRAVA_COLORS 
} from '../assets/strava';

export interface StravaAttributionProps {
  /** Type of data being displayed to determine attribution requirements */
  dataType?: 'activity_data' | 'segment_data' | 'athlete_data' | 'general';
  /** Visual variant of the attribution */
  variant?: 'inline' | 'footer' | 'badge';
  /** Size of the attribution */
  size?: 'small' | 'medium' | 'large';
  /** Theme for logo selection */
  theme?: 'light' | 'dark';
  /** Additional CSS classes */
  className?: string;
  /** Whether to show the logo alongside text */
  showLogo?: boolean;
  /** Custom attribution text (overrides default) */
  customText?: string;
}

const StravaAttribution: React.FC<StravaAttributionProps> = ({
  dataType = 'general',
  variant = 'inline',
  size = 'medium',
  theme = 'light',
  className = '',
  showLogo = true,
  customText,
}) => {
  const attributionReqs = getAttributionRequirements(dataType);
  const logoVariant = theme === 'dark' ? 'white' : 'orange';
  
  // Get appropriate logo based on size and variant
  const logo = size === 'small' || variant === 'inline'
    ? getStravaIcon(logoVariant)
    : getStravaWordmark(logoVariant);

  const attributionText = customText || attributionReqs.text;

  // Base styles for different variants
  const variantStyles = {
    inline: 'inline-flex items-center gap-2',
    footer: 'flex items-center justify-left gap-2 text-sm',
    badge: 'inline-flex items-center gap-1 px-2 py-1 rounded-md bg-gray-100 dark:bg-gray-800',
  };

  // Size styles
  const sizeStyles = {
    small: 'text-xs',
    medium: 'text-sm',
    large: 'text-base',
  };

  // Logo size styles
  const logoSizeStyles = {
    small: 'h-4 w-auto',
    medium: 'h-5 w-auto',
    large: 'h-6 w-auto',
  };

  const baseClasses = `
    ${variantStyles[variant]} 
    ${sizeStyles[size]} 
    text-gray-600 dark:text-gray-300
    ${className}
  `.trim();

  return (
    <div 
      className={baseClasses}
      role="contentinfo"
      aria-label="Strava attribution"
    >
      {showLogo && logo && (
        <a href="https://www.strava.com"
          target="_blank"
          rel="noopener noreferrer">
        <img
          src={logo.url}
          alt="Strava logo"
          className={`${logoSizeStyles[size]} flex-shrink-0`}
          style={{
            minHeight: `${logo.minHeight}px`,
            minWidth: `${logo.minWidth}px`,
          }}
        />
        </a>
      )}
    </div>
  );
};

export default StravaAttribution;