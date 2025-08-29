import React, { useState } from 'react';
import { DiagramLibraryProvider } from '../contexts/DiagramLibraryContext';
import { useDiagramLoader } from '../hooks/useDiagramLoader';
import { DiagramLoadingIndicator, LibraryLoadingIndicator } from '../components/DiagramLoadingIndicator';

/**
 * Demo component that shows the lazy loading system in action
 */
const DiagramContent: React.FC<{ content: string }> = ({ content }) => {
  const {
    hasDiagrams,
    diagramInfo,
    libraryState,
    isLoading,
    allRequiredLibrariesLoaded,
    errors,
    loadRequiredLibraries
  } = useDiagramLoader(content);

  const loadingLibraries: Array<'mermaid' | 'vega-lite'> = [];
  if (diagramInfo.mermaidCount > 0 && libraryState.mermaid.loading) {
    loadingLibraries.push('mermaid');
  }
  if (diagramInfo.vegaLiteCount > 0 && libraryState.vegaLite.loading) {
    loadingLibraries.push('vega-lite');
  }

  return (
    <div className="space-y-4">
      {/* Status Information */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <h3 className="font-semibold text-gray-800 mb-2">Lazy Loading Status</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <strong>Has Diagrams:</strong> {hasDiagrams ? 'Yes' : 'No'}
          </div>
          <div>
            <strong>Total Diagrams:</strong> {diagramInfo.totalCount}
          </div>
          <div>
            <strong>Mermaid Count:</strong> {diagramInfo.mermaidCount}
          </div>
          <div>
            <strong>Vega-Lite Count:</strong> {diagramInfo.vegaLiteCount}
          </div>
          <div>
            <strong>Libraries Loading:</strong> {isLoading ? 'Yes' : 'No'}
          </div>
          <div>
            <strong>All Libraries Ready:</strong> {allRequiredLibrariesLoaded ? 'Yes' : 'No'}
          </div>
        </div>
        
        {/* Library States */}
        <div className="mt-4 grid grid-cols-2 gap-4">
          <div className="bg-white p-3 rounded border">
            <h4 className="font-medium text-gray-700 mb-1">Mermaid Library</h4>
            <div className="text-xs space-y-1">
              <div>Loaded: {libraryState.mermaid.loaded ? '✅' : '❌'}</div>
              <div>Loading: {libraryState.mermaid.loading ? '⏳' : '❌'}</div>
              <div>Error: {libraryState.mermaid.error || 'None'}</div>
            </div>
          </div>
          <div className="bg-white p-3 rounded border">
            <h4 className="font-medium text-gray-700 mb-1">Vega-Lite Library</h4>
            <div className="text-xs space-y-1">
              <div>Loaded: {libraryState.vegaLite.loaded ? '✅' : '❌'}</div>
              <div>Loading: {libraryState.vegaLite.loading ? '⏳' : '❌'}</div>
              <div>Error: {libraryState.vegaLite.error || 'None'}</div>
            </div>
          </div>
        </div>
      </div>

      {/* Loading Indicators */}
      {loadingLibraries.length > 0 && (
        <LibraryLoadingIndicator loadingLibraries={loadingLibraries} />
      )}

      {/* Errors */}
      {errors.length > 0 && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h4 className="font-medium text-red-800 mb-2">Loading Errors:</h4>
          <ul className="text-sm text-red-600 space-y-1">
            {errors.map((error, index) => (
              <li key={index}>• {error}</li>
            ))}
          </ul>
        </div>
      )}

      {/* Manual Load Button */}
      {hasDiagrams && !allRequiredLibrariesLoaded && !isLoading && (
        <button
          onClick={loadRequiredLibraries}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors"
        >
          Load Required Libraries
        </button>
      )}

      {/* Content Preview */}
      <div className="bg-white border rounded-lg p-4">
        <h4 className="font-medium text-gray-800 mb-2">Content Preview:</h4>
        <pre className="text-xs text-gray-600 whitespace-pre-wrap overflow-x-auto">
          {content.substring(0, 500)}{content.length > 500 ? '...' : ''}
        </pre>
      </div>
    </div>
  );
};

/**
 * Main demo component with different content examples
 */
export const DiagramLazyLoadingDemo: React.FC = () => {
  const [selectedContent, setSelectedContent] = useState('none');

  const contentExamples = {
    none: 'Regular markdown content without any diagrams.\n\nThis should not trigger any library loading.',
    mermaid: `# Mermaid Diagram Example

Here's a flowchart:

\`\`\`mermaid
graph TD
    A[Start] --> B{Is it working?}
    B -->|Yes| C[Great!]
    B -->|No| D[Debug]
    D --> B
    C --> E[End]
\`\`\`

This content should trigger Mermaid library loading.`,
    vegaLite: `# Vega-Lite Chart Example

Here's a simple bar chart:

\`\`\`vega-lite
{
  "mark": "bar",
  "data": {
    "values": [
      {"category": "A", "value": 28},
      {"category": "B", "value": 55},
      {"category": "C", "value": 43}
    ]
  },
  "encoding": {
    "x": {"field": "category", "type": "nominal"},
    "y": {"field": "value", "type": "quantitative"}
  }
}
\`\`\`

This content should trigger Vega-Lite library loading.`,
    mixed: `# Mixed Diagrams Example

Here's a flowchart:

\`\`\`mermaid
graph LR
    A[Data] --> B[Process]
    B --> C[Chart]
\`\`\`

And here's a chart:

\`\`\`vega-lite
{
  "mark": "point",
  "data": {
    "values": [
      {"x": 1, "y": 2},
      {"x": 2, "y": 5},
      {"x": 3, "y": 3}
    ]
  },
  "encoding": {
    "x": {"field": "x", "type": "quantitative"},
    "y": {"field": "y", "type": "quantitative"}
  }
}
\`\`\`

This content should trigger both library loadings.`
  };

  return (
    <DiagramLibraryProvider>
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Diagram Lazy Loading Demo
          </h1>
          <p className="text-gray-600">
            This demo shows how diagram libraries are loaded only when needed
          </p>
        </div>

        {/* Content Selector */}
        <div className="bg-white border rounded-lg p-4">
          <h2 className="font-semibold text-gray-800 mb-3">Select Content Type:</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
            {Object.entries(contentExamples).map(([key, _]) => (
              <button
                key={key}
                onClick={() => setSelectedContent(key)}
                className={`px-3 py-2 rounded text-sm font-medium transition-colors ${
                  selectedContent === key
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {key.charAt(0).toUpperCase() + key.slice(1)}
              </button>
            ))}
          </div>
        </div>

        {/* Demo Content */}
        <DiagramContent content={contentExamples[selectedContent as keyof typeof contentExamples]} />
      </div>
    </DiagramLibraryProvider>
  );
};

export default DiagramLazyLoadingDemo;