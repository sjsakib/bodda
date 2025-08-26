import React, { useState } from 'react';
import SuggestionPills from '../SuggestionPills';

/**
 * Visual test component for SuggestionPills
 * This component demonstrates the visual behavior, responsiveness, and accessibility features
 */
export const SuggestionPillsVisualTest: React.FC = () => {
  const [selectedText, setSelectedText] = useState<string>('');
  const [clickCount, setClickCount] = useState<number>(0);

  const handlePillClick = (text: string) => {
    setSelectedText(text);
    setClickCount(prev => prev + 1);
  };

  const resetDemo = () => {
    setSelectedText('');
    setClickCount(0);
  };

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">
            SuggestionPills Visual Test
          </h1>
          <p className="text-gray-600 max-w-2xl mx-auto">
            This page demonstrates the SuggestionPills component across different screen sizes, 
            interaction states, and accessibility features. Resize your browser window to test responsiveness.
          </p>
        </div>

        {/* Interactive Demo */}
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-blue-50 px-6 py-4 border-b">
            <h2 className="text-xl font-semibold text-blue-900">Interactive Demo</h2>
            <p className="text-blue-700 text-sm mt-1">
              Click on pills to see interaction behavior. Use keyboard navigation (Tab, Arrow keys, Enter/Space).
            </p>
          </div>
          
          <div className="p-6">
            {/* Demo State Display */}
            <div className="mb-6 p-4 bg-gray-50 rounded-lg">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <strong>Selected Text:</strong>
                  <div className="mt-1 p-2 bg-white border rounded text-sm">
                    {selectedText || 'No pill selected yet'}
                  </div>
                </div>
                <div>
                  <strong>Click Count:</strong>
                  <div className="mt-1 p-2 bg-white border rounded text-sm">
                    {clickCount} clicks
                  </div>
                </div>
              </div>
              <button
                onClick={resetDemo}
                className="mt-3 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
              >
                Reset Demo
              </button>
            </div>

            {/* Chat Interface Mockup */}
            <div className="border border-gray-200 rounded-lg overflow-hidden">
              {/* Mock Messages Area */}
              <div className="h-32 bg-gray-50 flex items-center justify-center border-b">
                <p className="text-gray-500 text-sm">
                  Empty chat session - suggestion pills should appear below
                </p>
              </div>

              {/* SuggestionPills Component */}
              <SuggestionPills onPillClick={handlePillClick} />

              {/* Mock Input Area */}
              <div className="p-4 bg-white border-t">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={selectedText}
                    onChange={(e) => setSelectedText(e.target.value)}
                    placeholder="Type your message or click a suggestion above..."
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                  <button className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
                    Send
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Responsive Layout Tests */}
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-green-50 px-6 py-4 border-b">
            <h2 className="text-xl font-semibold text-green-900">Responsive Layout Tests</h2>
            <p className="text-green-700 text-sm mt-1">
              These examples show how the component adapts to different container widths.
            </p>
          </div>
          
          <div className="p-6 space-y-8">
            {/* Mobile Width Simulation */}
            <div>
              <h3 className="text-lg font-medium mb-3">Mobile Layout (375px width)</h3>
              <div className="w-full max-w-sm mx-auto border border-gray-300 rounded-lg overflow-hidden">
                <div className="h-16 bg-gray-50 flex items-center justify-center text-sm text-gray-500">
                  Mobile Chat View
                </div>
                <SuggestionPills onPillClick={handlePillClick} />
                <div className="p-3 bg-white border-t">
                  <input
                    type="text"
                    placeholder="Mobile input..."
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded"
                  />
                </div>
              </div>
            </div>

            {/* Tablet Width Simulation */}
            <div>
              <h3 className="text-lg font-medium mb-3">Tablet Layout (768px width)</h3>
              <div className="w-full max-w-2xl mx-auto border border-gray-300 rounded-lg overflow-hidden">
                <div className="h-16 bg-gray-50 flex items-center justify-center text-sm text-gray-500">
                  Tablet Chat View
                </div>
                <SuggestionPills onPillClick={handlePillClick} />
                <div className="p-4 bg-white border-t">
                  <input
                    type="text"
                    placeholder="Tablet input..."
                    className="w-full px-3 py-2 border border-gray-300 rounded"
                  />
                </div>
              </div>
            </div>

            {/* Desktop Width Simulation */}
            <div>
              <h3 className="text-lg font-medium mb-3">Desktop Layout (1024px+ width)</h3>
              <div className="w-full border border-gray-300 rounded-lg overflow-hidden">
                <div className="h-16 bg-gray-50 flex items-center justify-center text-sm text-gray-500">
                  Desktop Chat View
                </div>
                <SuggestionPills onPillClick={handlePillClick} />
                <div className="p-4 bg-white border-t">
                  <input
                    type="text"
                    placeholder="Desktop input..."
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Accessibility Features Demo */}
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-purple-50 px-6 py-4 border-b">
            <h2 className="text-xl font-semibold text-purple-900">Accessibility Features</h2>
            <p className="text-purple-700 text-sm mt-1">
              Test keyboard navigation, screen reader support, and focus management.
            </p>
          </div>
          
          <div className="p-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              {/* Keyboard Navigation Guide */}
              <div>
                <h3 className="text-lg font-medium mb-3">Keyboard Navigation</h3>
                <div className="space-y-2 text-sm">
                  <div className="flex items-center gap-2">
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">Tab</kbd>
                    <span>Navigate to pills</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">Arrow Keys</kbd>
                    <span>Navigate between pills</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">Enter</kbd>
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">Space</kbd>
                    <span>Activate pill</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">Home</kbd>
                    <kbd className="px-2 py-1 bg-gray-100 border rounded text-xs">End</kbd>
                    <span>Jump to first/last pill</span>
                  </div>
                </div>
              </div>

              {/* Accessibility Features List */}
              <div>
                <h3 className="text-lg font-medium mb-3">Accessibility Features</h3>
                <ul className="space-y-2 text-sm">
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>ARIA labels and roles for screen readers</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>Keyboard navigation support</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>Focus indicators and management</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>Minimum 44px touch targets</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>High contrast colors</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">✓</span>
                    <span>Responsive grid layout</span>
                  </li>
                </ul>
              </div>
            </div>

            {/* Test Component */}
            <div className="mt-6 p-4 border border-gray-200 rounded-lg">
              <h4 className="font-medium mb-3">Try the accessibility features:</h4>
              <SuggestionPills onPillClick={handlePillClick} />
            </div>
          </div>
        </div>

        {/* Visual States Demo */}
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-orange-50 px-6 py-4 border-b">
            <h2 className="text-xl font-semibold text-orange-900">Visual States</h2>
            <p className="text-orange-700 text-sm mt-1">
              Hover over pills to see hover effects. Focus with keyboard to see focus rings.
            </p>
          </div>
          
          <div className="p-6">
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium mb-3">Normal State</h3>
                <SuggestionPills onPillClick={handlePillClick} />
              </div>

              <div>
                <h3 className="text-lg font-medium mb-3">With Custom Styling</h3>
                <SuggestionPills 
                  onPillClick={handlePillClick} 
                  className="bg-blue-50 border-blue-200" 
                />
              </div>
            </div>
          </div>
        </div>

        {/* Testing Instructions */}
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          <div className="bg-gray-50 px-6 py-4 border-b">
            <h2 className="text-xl font-semibold text-gray-900">Testing Instructions</h2>
          </div>
          
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <h3 className="font-medium mb-3">Visual Testing</h3>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li>• Resize browser window to test responsive behavior</li>
                  <li>• Hover over pills to see hover effects</li>
                  <li>• Check color contrast in different lighting</li>
                  <li>• Verify touch targets are adequate on mobile</li>
                  <li>• Test with different zoom levels (100%, 150%, 200%)</li>
                </ul>
              </div>
              
              <div>
                <h3 className="font-medium mb-3">Accessibility Testing</h3>
                <ul className="space-y-2 text-sm text-gray-600">
                  <li>• Navigate using only keyboard</li>
                  <li>• Test with screen reader (VoiceOver, NVDA, etc.)</li>
                  <li>• Verify focus indicators are visible</li>
                  <li>• Check tab order is logical</li>
                  <li>• Test with high contrast mode enabled</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SuggestionPillsVisualTest;