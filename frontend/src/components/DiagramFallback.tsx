import React, { useState } from 'react';
import { RetryButton } from './ErrorBoundary';

/**
 * Props interface for DiagramFallback components
 */
interface DiagramFallbackProps {
  /** The diagram content that failed to render */
  content: string;
  /** The type of diagram */
  diagramType: 'mermaid' | 'vega-lite';
  /** The error message */
  error?: string;
  /** Optional CSS class name */
  className?: string;
  /** Callback when user retries */
  onRetry?: () => void;
  /** Whether retry is available */
  canRetry?: boolean;
  /** Whether to show the raw content by default */
  showContentByDefault?: boolean;
}

/**
 * Fallback component for when diagram libraries fail to load
 */
export const LibraryLoadFailureFallback: React.FC<DiagramFallbackProps> = ({
  content,
  diagramType,
  error,
  className = '',
  onRetry,
  canRetry = true,
}) => {
  const [showContent, setShowContent] = useState(false);

  const libraryName = diagramType === 'mermaid' ? 'Mermaid' : 'Vega-Lite';

  return (
    <div className={`diagram-fallback library-load-failure ${className}`}>
      <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
        {/* Header */}
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-blue-800">
              {libraryName} Library Not Available
            </h3>
            <p className="mt-1 text-sm text-blue-700">
              The {libraryName} rendering library couldn't be loaded. This might be due to a network issue or browser compatibility.
            </p>
          </div>
        </div>

        {/* Error Details */}
        {error && (
          <div className="mt-3 text-xs text-blue-600">
            <strong>Error:</strong> {error}
          </div>
        )}

        {/* Actions */}
        <div className="mt-4 flex items-center space-x-3">
          {canRetry && onRetry && (
            <RetryButton
              onRetry={onRetry}
              className="text-xs bg-blue-600 text-white border-blue-600 hover:bg-blue-700"
            >
              Retry Loading
            </RetryButton>
          )}
          
          <button
            onClick={() => setShowContent(!showContent)}
            className="text-xs px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition-colors"
          >
            {showContent ? 'Hide' : 'Show'} Raw Content
          </button>
        </div>

        {/* Raw Content */}
        {showContent && (
          <div className="mt-4">
            <div className="text-xs text-blue-600 mb-2">
              Raw {libraryName} content:
            </div>
            <pre className="text-xs bg-blue-100 p-3 rounded border border-blue-300 overflow-x-auto max-h-60">
              {content}
            </pre>
          </div>
        )}

        {/* Help Text */}
        <div className="mt-4 text-xs text-blue-600">
          <p>
            <strong>What you can do:</strong>
          </p>
          <ul className="mt-1 space-y-1 ml-4">
            <li>• Check your internet connection and try again</li>
            <li>• Refresh the page to reload the libraries</li>
            <li>• Try using a different browser if the issue persists</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

/**
 * Fallback component for invalid diagram syntax
 */
export const InvalidSyntaxFallback: React.FC<DiagramFallbackProps> = ({
  content,
  diagramType,
  error,
  className = '',
  showContentByDefault = false,
}) => {
  const [showContent, setShowContent] = useState(showContentByDefault);

  const libraryName = diagramType === 'mermaid' ? 'Mermaid' : 'Vega-Lite';
  const docLinks = {
    mermaid: 'https://mermaid.js.org/syntax/',
    'vega-lite': 'https://vega.github.io/vega-lite/docs/',
  };

  return (
    <div className={`diagram-fallback invalid-syntax ${className}`}>
      <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg">
        {/* Header */}
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-amber-800">
              Invalid {libraryName} Syntax
            </h3>
            <p className="mt-1 text-sm text-amber-700">
              The {libraryName} diagram contains syntax errors and cannot be rendered.
            </p>
          </div>
        </div>

        {/* Error Details */}
        {error && (
          <div className="mt-3 text-xs text-amber-700 bg-amber-100 p-2 rounded">
            <strong>Syntax Error:</strong> {error}
          </div>
        )}

        {/* Actions */}
        <div className="mt-4 flex items-center space-x-3">
          <button
            onClick={() => setShowContent(!showContent)}
            className="text-xs px-3 py-1 bg-amber-100 text-amber-800 rounded hover:bg-amber-200 transition-colors"
          >
            {showContent ? 'Hide' : 'Show'} Content
          </button>
          
          <a
            href={docLinks[diagramType]}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs px-3 py-1 bg-amber-600 text-white rounded hover:bg-amber-700 transition-colors"
          >
            View Documentation
          </a>
        </div>

        {/* Raw Content */}
        {showContent && (
          <div className="mt-4">
            <div className="text-xs text-amber-700 mb-2">
              {libraryName} content with syntax errors:
            </div>
            <pre className="text-xs bg-amber-100 p-3 rounded border border-amber-300 overflow-x-auto max-h-60">
              {content}
            </pre>
          </div>
        )}

        {/* Help Text */}
        <div className="mt-4 text-xs text-amber-700">
          <p>
            <strong>Common {libraryName} syntax issues:</strong>
          </p>
          {diagramType === 'mermaid' ? (
            <ul className="mt-1 space-y-1 ml-4">
              <li>• Missing diagram type declaration (e.g., `graph TD`, `sequenceDiagram`)</li>
              <li>• Invalid node or edge syntax</li>
              <li>• Unmatched quotes or brackets</li>
              <li>• Reserved keywords used as identifiers</li>
            </ul>
          ) : (
            <ul className="mt-1 space-y-1 ml-4">
              <li>• Invalid JSON format</li>
              <li>• Missing required properties (mark, data, etc.)</li>
              <li>• Incorrect data format or structure</li>
              <li>• Invalid encoding or field references</li>
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

/**
 * Fallback component for rendering timeouts
 */
export const TimeoutFallback: React.FC<DiagramFallbackProps> = ({
  content,
  diagramType,
  className = '',
  onRetry,
  canRetry = true,
}) => {
  const [showContent, setShowContent] = useState(false);

  const libraryName = diagramType === 'mermaid' ? 'Mermaid' : 'Vega-Lite';

  return (
    <div className={`diagram-fallback timeout ${className}`}>
      <div className="p-4 bg-orange-50 border border-orange-200 rounded-lg">
        {/* Header */}
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="w-5 h-5 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-orange-800">
              {libraryName} Rendering Timeout
            </h3>
            <p className="mt-1 text-sm text-orange-700">
              The {libraryName} diagram is taking too long to render. This might be due to complex content or performance issues.
            </p>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-4 flex items-center space-x-3">
          {canRetry && onRetry && (
            <RetryButton
              onRetry={onRetry}
              className="text-xs bg-orange-600 text-white border-orange-600 hover:bg-orange-700"
            >
              Try Again
            </RetryButton>
          )}
          
          <button
            onClick={() => setShowContent(!showContent)}
            className="text-xs px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition-colors"
          >
            {showContent ? 'Hide' : 'Show'} Content
          </button>
        </div>

        {/* Raw Content */}
        {showContent && (
          <div className="mt-4">
            <div className="text-xs text-orange-700 mb-2">
              {libraryName} content that timed out:
            </div>
            <pre className="text-xs bg-orange-100 p-3 rounded border border-orange-300 overflow-x-auto max-h-60">
              {content}
            </pre>
          </div>
        )}

        {/* Help Text */}
        <div className="mt-4 text-xs text-orange-700">
          <p>
            <strong>To resolve timeout issues:</strong>
          </p>
          <ul className="mt-1 space-y-1 ml-4">
            <li>• Try simplifying the diagram structure</li>
            <li>• Reduce the amount of data or nodes</li>
            <li>• Break complex diagrams into smaller parts</li>
            <li>• Check if your device has sufficient resources</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

/**
 * Generic fallback component that can handle different error types
 */
export const GenericDiagramFallback: React.FC<DiagramFallbackProps & {
  errorType?: 'library-load' | 'invalid-syntax' | 'timeout' | 'generic';
}> = ({ errorType = 'generic', ...props }) => {
  switch (errorType) {
    case 'library-load':
      return <LibraryLoadFailureFallback {...props} />;
    case 'invalid-syntax':
      return <InvalidSyntaxFallback {...props} />;
    case 'timeout':
      return <TimeoutFallback {...props} />;
    default:
      return <GenericErrorFallback {...props} />;
  }
};

/**
 * Generic error fallback for unknown error types
 */
const GenericErrorFallback: React.FC<DiagramFallbackProps> = ({
  content,
  diagramType,
  error,
  className = '',
  onRetry,
  canRetry = true,
}) => {
  const [showContent, setShowContent] = useState(false);

  const libraryName = diagramType === 'mermaid' ? 'Mermaid' : 'Vega-Lite';

  return (
    <div className={`diagram-fallback generic-error ${className}`}>
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        {/* Header */}
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-red-800">
              {libraryName} Diagram Error
            </h3>
            <p className="mt-1 text-sm text-red-700">
              An unexpected error occurred while rendering the {libraryName} diagram.
            </p>
          </div>
        </div>

        {/* Error Details */}
        {error && (
          <div className="mt-3 text-xs text-red-700 bg-red-100 p-2 rounded">
            <strong>Error:</strong> {error}
          </div>
        )}

        {/* Actions */}
        <div className="mt-4 flex items-center space-x-3">
          {canRetry && onRetry && (
            <RetryButton
              onRetry={onRetry}
              className="text-xs bg-red-600 text-white border-red-600 hover:bg-red-700"
            >
              Try Again
            </RetryButton>
          )}
          
          <button
            onClick={() => setShowContent(!showContent)}
            className="text-xs px-3 py-1 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition-colors"
          >
            {showContent ? 'Hide' : 'Show'} Content
          </button>
          
          <button
            onClick={() => window.location.reload()}
            className="text-xs px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 transition-colors"
          >
            Refresh Page
          </button>
        </div>

        {/* Raw Content */}
        {showContent && (
          <div className="mt-4">
            <div className="text-xs text-red-700 mb-2">
              Raw {libraryName} content:
            </div>
            <pre className="text-xs bg-red-100 p-3 rounded border border-red-300 overflow-x-auto max-h-60">
              {content}
            </pre>
          </div>
        )}

        {/* Help Text */}
        <div className="mt-4 text-xs text-red-700">
          <p>
            <strong>If this problem persists:</strong>
          </p>
          <ul className="mt-1 space-y-1 ml-4">
            <li>• Try refreshing the page</li>
            <li>• Check your internet connection</li>
            <li>• Try using a different browser</li>
            <li>• Contact support if the issue continues</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

/**
 * Utility function to determine the appropriate fallback component based on error
 */
export const getDiagramFallback = (
  error: string,
  content: string,
  diagramType: 'mermaid' | 'vega-lite',
  onRetry?: () => void,
  className?: string
): React.ReactElement => {
  const props = {
    content,
    diagramType,
    error,
    className,
    onRetry,
    canRetry: !!onRetry,
  };

  if (error.toLowerCase().includes('load') || error.toLowerCase().includes('import')) {
    return <LibraryLoadFailureFallback {...props} />;
  }
  
  if (error.toLowerCase().includes('syntax') || error.toLowerCase().includes('parse')) {
    return <InvalidSyntaxFallback {...props} />;
  }
  
  if (error.toLowerCase().includes('timeout')) {
    return <TimeoutFallback {...props} />;
  }
  
  return <GenericErrorFallback {...props} />;
};

export default GenericDiagramFallback;