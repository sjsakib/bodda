import React from 'react';

/**
 * Props interface for the DiagramLoadingIndicator component
 */
interface DiagramLoadingIndicatorProps {
  /** Type of diagram being loaded */
  type: 'mermaid' | 'vega-lite' | 'general';
  /** Optional custom message */
  message?: string;
  /** Optional CSS class name */
  className?: string;
  /** Size variant for the loading indicator */
  size?: 'small' | 'medium' | 'large';
}

/**
 * Loading indicator component specifically designed for diagram rendering
 * Provides visual feedback while diagram libraries are being loaded or diagrams are being rendered
 */
export const DiagramLoadingIndicator: React.FC<DiagramLoadingIndicatorProps> = ({
  type,
  message,
  className = '',
  size = 'medium'
}) => {
  const getDefaultMessage = () => {
    switch (type) {
      case 'mermaid':
        return 'Loading Mermaid diagram...';
      case 'vega-lite':
        return 'Loading chart...';
      default:
        return 'Loading diagram...';
    }
  };

  const getSizeClasses = () => {
    switch (size) {
      case 'small':
        return {
          spinner: 'h-4 w-4',
          container: 'p-4',
          text: 'text-xs'
        };
      case 'large':
        return {
          spinner: 'h-8 w-8',
          container: 'p-8',
          text: 'text-base'
        };
      default:
        return {
          spinner: 'h-6 w-6',
          container: 'p-6',
          text: 'text-sm'
        };
    }
  };

  const sizeClasses = getSizeClasses();
  const displayMessage = message || getDefaultMessage();

  return (
    <div className={`diagram-loading ${className}`}>
      <div className={`flex items-center justify-center bg-gray-50 rounded-lg border border-gray-200 ${sizeClasses.container}`}>
        <div className="flex items-center space-x-3">
          {/* Animated spinner */}
          <div 
            className={`animate-spin rounded-full border-2 border-gray-300 border-t-blue-600 ${sizeClasses.spinner}`}
            role="status"
            aria-label="Loading diagram"
          />
          
          {/* Loading message */}
          <span className={`text-gray-600 font-medium ${sizeClasses.text}`}>
            {displayMessage}
          </span>
        </div>
      </div>
    </div>
  );
};

/**
 * Props interface for the LibraryLoadingIndicator component
 */
interface LibraryLoadingIndicatorProps {
  /** Libraries currently being loaded */
  loadingLibraries: Array<'mermaid' | 'vega-lite'>;
  /** Optional CSS class name */
  className?: string;
}

/**
 * Loading indicator for diagram libraries initialization
 * Shows progress when multiple libraries are being loaded
 */
export const LibraryLoadingIndicator: React.FC<LibraryLoadingIndicatorProps> = ({
  loadingLibraries,
  className = ''
}) => {
  if (loadingLibraries.length === 0) {
    return null;
  }

  const getLibraryDisplayName = (library: 'mermaid' | 'vega-lite') => {
    switch (library) {
      case 'mermaid':
        return 'Mermaid';
      case 'vega-lite':
        return 'Vega-Lite';
      default:
        return library;
    }
  };

  return (
    <div className={`library-loading ${className}`}>
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-center space-x-3">
          <div 
            className="animate-spin rounded-full h-5 w-5 border-2 border-blue-300 border-t-blue-600"
            role="status"
            aria-label="Loading diagram libraries"
          />
          <div className="flex-1">
            <p className="text-sm font-medium text-blue-800">
              Initializing diagram libraries...
            </p>
            <p className="text-xs text-blue-600 mt-1">
              Loading: {loadingLibraries.map(getLibraryDisplayName).join(', ')}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Props interface for the DiagramErrorIndicator component
 */
interface DiagramErrorIndicatorProps {
  /** Error message to display */
  error: string;
  /** Type of diagram that failed */
  type?: 'mermaid' | 'vega-lite' | 'general';
  /** Optional retry function */
  onRetry?: () => void;
  /** Optional CSS class name */
  className?: string;
  /** Whether to show raw content in details */
  showRawContent?: boolean;
  /** Raw content to show in details */
  rawContent?: string;
}

/**
 * Error indicator component for diagram loading/rendering failures
 * Provides user-friendly error messages and optional retry functionality
 */
export const DiagramErrorIndicator: React.FC<DiagramErrorIndicatorProps> = ({
  error,
  type = 'general',
  onRetry,
  className = '',
  showRawContent = false,
  rawContent
}) => {
  const getErrorTitle = () => {
    switch (type) {
      case 'mermaid':
        return 'Mermaid Diagram Error';
      case 'vega-lite':
        return 'Chart Error';
      default:
        return 'Diagram Error';
    }
  };

  return (
    <div className={`diagram-error ${className}`}>
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <div className="flex items-start space-x-3">
          {/* Error icon */}
          <div className="flex-shrink-0">
            <svg 
              className="h-5 w-5 text-red-400" 
              viewBox="0 0 20 20" 
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                fillRule="evenodd" 
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" 
                clipRule="evenodd" 
              />
            </svg>
          </div>
          
          <div className="flex-1">
            {/* Error title */}
            <h3 className="text-sm font-medium text-red-800">
              {getErrorTitle()}
            </h3>
            
            {/* Error message */}
            <p className="text-sm text-red-600 mt-1">
              {error}
            </p>
            
            {/* Retry button */}
            {onRetry && (
              <button
                onClick={onRetry}
                className="mt-2 text-xs text-red-700 hover:text-red-900 underline focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 rounded"
              >
                Try again
              </button>
            )}
            
            {/* Raw content details */}
            {showRawContent && rawContent && (
              <details className="mt-3">
                <summary className="text-xs text-red-500 cursor-pointer hover:text-red-700">
                  Show raw content
                </summary>
                <pre className="text-xs text-red-400 mt-2 p-2 bg-red-100 rounded border overflow-x-auto whitespace-pre-wrap">
                  {rawContent}
                </pre>
              </details>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default DiagramLoadingIndicator;