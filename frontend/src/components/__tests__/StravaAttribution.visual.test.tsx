import React from 'react';
import { render } from '@testing-library/react';
import { describe, it } from 'vitest';
import StravaAttribution from '../StravaAttribution';

// Visual test component to demonstrate different variants
const StravaAttributionVisualTest: React.FC = () => {
  return (
    <div className="p-8 space-y-8 bg-white">
      <h1 className="text-2xl font-bold mb-6">Strava Attribution Component Examples</h1>
      
      {/* Basic Examples */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Basic Usage</h2>
        <div className="space-y-4">
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Default (General Data)</h3>
            <StravaAttribution />
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Activity Data</h3>
            <StravaAttribution dataType="activity_data" />
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Segment Data</h3>
            <StravaAttribution dataType="segment_data" />
          </div>
        </div>
      </section>

      {/* Variants */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Variants</h2>
        <div className="space-y-4">
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Inline Variant</h3>
            <p className="text-gray-600 mb-2">Use within content flows:</p>
            <div className="bg-gray-50 p-3 rounded">
              <p>Your activity data is displayed below.</p>
              <StravaAttribution variant="inline" className="mt-2" />
            </div>
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Footer Variant</h3>
            <p className="text-gray-600 mb-2">Use in page footers:</p>
            <div className="bg-gray-50 p-3 rounded">
              <StravaAttribution variant="footer" />
            </div>
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Badge Variant</h3>
            <p className="text-gray-600 mb-2">Use as a compact badge:</p>
            <div className="bg-gray-50 p-3 rounded">
              <StravaAttribution variant="badge" />
            </div>
          </div>
        </div>
      </section>

      {/* Sizes */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Sizes</h2>
        <div className="space-y-4">
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Small</h3>
            <StravaAttribution size="small" />
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Medium (Default)</h3>
            <StravaAttribution size="medium" />
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Large</h3>
            <StravaAttribution size="large" />
          </div>
        </div>
      </section>

      {/* Themes */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Themes</h2>
        <div className="space-y-4">
          <div className="p-4 border rounded bg-white">
            <h3 className="text-sm font-medium mb-2">Light Theme (Default)</h3>
            <StravaAttribution theme="light" />
          </div>
          
          <div className="p-4 border rounded bg-gray-900">
            <h3 className="text-sm font-medium mb-2 text-white">Dark Theme</h3>
            <StravaAttribution theme="dark" />
          </div>
        </div>
      </section>

      {/* Real-world Examples */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Real-world Usage Examples</h2>
        <div className="space-y-4">
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Activity Card</h3>
            <div className="bg-gray-50 p-4 rounded">
              <h4 className="font-semibold">Morning Run</h4>
              <p className="text-sm text-gray-600">5.2 miles • 42:15 • 8:07/mile</p>
              <StravaAttribution 
                dataType="activity_data" 
                variant="inline" 
                size="small" 
                className="mt-3" 
              />
            </div>
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Segment Leaderboard</h3>
            <div className="bg-gray-50 p-4 rounded">
              <h4 className="font-semibold">Hawk Hill Climb</h4>
              <p className="text-sm text-gray-600">1. John Doe - 4:23</p>
              <p className="text-sm text-gray-600">2. Jane Smith - 4:31</p>
              <StravaAttribution 
                dataType="segment_data" 
                variant="badge" 
                size="small" 
                className="mt-3" 
              />
            </div>
          </div>
          
          <div className="p-4 border rounded">
            <h3 className="text-sm font-medium mb-2">Page Footer</h3>
            <div className="bg-gray-50 p-4 rounded border-t">
              <StravaAttribution variant="footer" />
            </div>
          </div>
        </div>
      </section>

      {/* Without Logo */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Text Only (No Logo)</h2>
        <div className="p-4 border rounded">
          <StravaAttribution showLogo={false} />
        </div>
      </section>

      {/* Custom Text */}
      <section>
        <h2 className="text-lg font-semibold mb-4">Custom Attribution Text</h2>
        <div className="p-4 border rounded">
          <StravaAttribution customText="Data provided by Strava" />
        </div>
      </section>
    </div>
  );
};

describe('StravaAttribution Visual Tests', () => {
  it('renders all visual examples without errors', () => {
    render(<StravaAttributionVisualTest />);
    // This test ensures all variants render without throwing errors
  });
});

export default StravaAttributionVisualTest;