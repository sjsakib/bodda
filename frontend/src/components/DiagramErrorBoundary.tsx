import React, { Component, ReactNode } from 'react';
import { RetryButton } from './ErrorBoundary';

/**
 * Props interface for DiagramErrorBoundary
 */
interface DiagramErrorBoundaryProps {
  children: ReactNode;
  /** The diagram content for debugging */
  content: string;
  /** The type of diagram (mermaid or vega-lite) */
  diagramType: 'mermaid' | 'vega-lite';
  /** Optional CSS class name */
  className?: string;
  /** Custom fallback component */
  fallback?: ReactNode;
  /** Callback when error occurs */
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
  /** Callback when user retries */
  onRetry?: () => void;
  /** Whether to show retry button */
  showRetry?: boolean;
  /** Whether to show raw content in error state */
  showRawContent?: boolean;
}

/**
 * State interface for DiagramErrorBoundary
 */
interface DiagramErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: React.ErrorInfo;
  retryCount: number;
  lastErrorTime: number;
}

/**
 * Enhanced error boundary specifically designed for diagram rendering failures.
 * Provides detailed error information, retry functionality, and graceful fallbacks.
 */
export class DiagramErrorBoundary extends Component<
  DiagramErrorBoundaryProps,
  DiagramErrorBoundaryState
> {
  private maxRetries = 3;
  private retryDelay = 1000; // 1 second

  constructor(props: DiagramErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      retryCount: 0,
      lastErrorTime: 0,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<DiagramErrorBoundaryState> {
    return {
      hasError: true,
      error,
      lastErrorTime: Date.now(),
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log detailed error information
    const errorDetails = {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
      },
      errorInfo: {
        componentStack: errorInfo.componentStack,
      },
      context: {
        diagramType: this.props.diagramType,
        contentLength: this.props.content.length,
        contentPreview: this.props.content.substring(0, 200),
        retryCount: this.state.retryCount,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
      },
    };

    console.error(`Diagram rendering error (${this.props.diagramType}):`, errorDetails);

    // Call custom error handler if provided
    this.props.onError?.(error, errorInfo);

    // Update state with error info
    this.setState({ errorInfo });

    // Report to monitoring service if available
    if ((window as any).gtag) {
      (window as any).gtag('event', 'exception', {
        description: `Diagram Error: ${error.message}`,
        fatal: false,
        custom_map: {
          diagram_type: this.props.diagramType,
          retry_count: this.state.retryCount,
        },
      });
    }
  }

  /**
   * Handle retry attempt
   */
  handleRetry = () => {
    const { retryCount } = this.state;
    
    if (retryCount >= this.maxRetries) {
      console.warn(`Maximum retry attempts (${this.maxRetries}) reached for ${this.props.diagramType} diagram`);
      return;
    }

    // Reset error state and increment retry count
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined,
      retryCount: retryCount + 1,
    });

    // Call custom retry handler
    this.props.onRetry?.();

    // Add delay before retry to prevent rapid retries
    setTimeout(() => {
      // Force re-render by updating a dummy state
      this.forceUpdate();
    }, this.retryDelay);
  };

  /**
   * Reset error boundary state
   */
  reset = () => {
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined,
      retryCount: 0,
      lastErrorTime: 0,
    });
  };

  /**
   * Get error severity based on error type and retry count
   */
  getErrorSeverity(): 'low' | 'medium' | 'high' {
    const { error, retryCount } = this.state;
    
    if (retryCount >= this.maxRetries) return 'high';
    if (error?.message.includes('timeout')) return 'medium';
    if (error?.message.includes('network') || error?.message.includes('load')) return 'medium';
    
    return 'low';
  }

  /**
   * Get user-friendly error message
   */
  getUserFriendlyMessage(): string {
    const { error } = this.state;
    const { diagramType } = this.props;
    
    if (!error) return `Failed to render ${diagramType} diagram`;
    
    if (error.message.includes('timeout')) {
      return `The ${diagramType} diagram is taking too long to render. This might be due to complex content or slow network.`;
    }
    
    if (error.message.includes('syntax') || error.message.includes('parse')) {
      return `The ${diagramType} diagram contains invalid syntax. Please check the diagram code.`;
    }
    
    if (error.message.includes('network') || error.message.includes('load')) {
      return `Failed to load the ${diagramType} rendering library. Please check your internet connection.`;
    }
    
    if (error.message.includes('memory') || error.message.includes('size')) {
      return `The ${diagramType} diagram is too large or complex to render.`;
    }
    
    return `An unexpected error occurred while rendering the ${diagramType} diagram.`;
  }

  /**
   * Get suggested actions for the user
   */
  getSuggestedActions(): string[] {
    const { error, retryCount } = this.state;
    const actions: string[] = [];
    
    if (retryCount < this.maxRetries) {
      actions.push('Try again by clicking the retry button');
    }
    
    if (error?.message.includes('syntax') || error?.message.includes('parse')) {
      actions.push('Check the diagram syntax for errors');
      actions.push('Refer to the documentation for correct syntax');
    }
    
    if (error?.message.includes('network') || error?.message.includes('load')) {
      actions.push('Check your internet connection');
      actions.push('Refresh the page to reload libraries');
    }
    
    if (error?.message.includes('timeout')) {
      actions.push('Try simplifying the diagram');
      actions.push('Break complex diagrams into smaller parts');
    }
    
    if (error?.message.includes('memory') || error?.message.includes('size')) {
      actions.push('Reduce the amount of data in the diagram');
      actions.push('Simplify the diagram structure');
    }
    
    if (actions.length === 0) {
      actions.push('Refresh the page and try again');
      actions.push('Contact support if the problem persists');
    }
    
    return actions;
  }

  render() {
    const { hasError, error, retryCount } = this.state;
    const { 
      children, 
      content, 
      diagramType, 
      className = '', 
      fallback, 
      showRetry = true, 
      showRawContent = true 
    } = this.props;

    if (hasError) {
      // Use custom fallback if provided
      if (fallback) {
        return fallback;
      }

      const severity = this.getErrorSeverity();
      const userMessage = this.getUserFriendlyMessage();
      const suggestedActions = this.getSuggestedActions();
      const canRetry = retryCount < this.maxRetries && showRetry;

      const severityColors = {
        low: 'bg-yellow-50 border-yellow-200 text-yellow-800',
        medium: 'bg-orange-50 border-orange-200 text-orange-800',
        high: 'bg-red-50 border-red-200 text-red-800',
      };

      const severityIcons = {
        low: (
          <svg className="w-5 h-5 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
        ),
        medium: (
          <svg className="w-5 h-5 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        ),
        high: (
          <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        ),
      };

      return (
        <div className={`diagram-error-boundary ${className}`}>
          <div className={`p-4 rounded-lg border ${severityColors[severity]}`}>
            {/* Error Header */}
            <div className="flex items-start">
              <div className="flex-shrink-0">
                {severityIcons[severity]}
              </div>
              <div className="ml-3 flex-1">
                <h3 className="text-sm font-medium">
                  {diagramType === 'mermaid' ? 'Mermaid' : 'Vega-Lite'} Diagram Error
                </h3>
                <p className="mt-1 text-sm">
                  {userMessage}
                </p>
              </div>
            </div>

            {/* Suggested Actions */}
            {suggestedActions.length > 0 && (
              <div className="mt-3">
                <p className="text-xs font-medium mb-2">Suggested actions:</p>
                <ul className="text-xs space-y-1">
                  {suggestedActions.map((action, index) => (
                    <li key={index} className="flex items-start">
                      <span className="mr-2">•</span>
                      <span>{action}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Action Buttons */}
            <div className="mt-4 flex items-center space-x-3">
              {canRetry && (
                <RetryButton
                  onRetry={this.handleRetry}
                  className="text-xs"
                >
                  Retry ({this.maxRetries - retryCount} left)
                </RetryButton>
              )}
              
              {retryCount >= this.maxRetries && (
                <button
                  onClick={() => window.location.reload()}
                  className="text-xs px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 transition-colors"
                >
                  Refresh Page
                </button>
              )}
            </div>

            {/* Technical Details (Collapsible) */}
            <details className="mt-3">
              <summary className="text-xs cursor-pointer hover:underline">
                Technical details
              </summary>
              <div className="mt-2 text-xs space-y-2">
                <div>
                  <strong>Error:</strong> {error?.message || 'Unknown error'}
                </div>
                <div>
                  <strong>Type:</strong> {diagramType}
                </div>
                <div>
                  <strong>Retry Count:</strong> {retryCount}/{this.maxRetries}
                </div>
                <div>
                  <strong>Timestamp:</strong> {new Date(this.state.lastErrorTime).toLocaleString()}
                </div>
              </div>
            </details>

            {/* Raw Content (Collapsible) */}
            {showRawContent && content && (
              <details className="mt-3">
                <summary className="text-xs cursor-pointer hover:underline">
                  Show raw {diagramType} content
                </summary>
                <pre className="mt-2 text-xs bg-gray-100 p-2 rounded border overflow-x-auto max-h-40">
                  {content}
                </pre>
              </details>
            )}
          </div>
        </div>
      );
    }

    return children;
  }
}

/**
 * Higher-order component that wraps any component with diagram error boundary
 */
export function withDiagramErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  diagramType: 'mermaid' | 'vega-lite'
) {
  const WrappedComponent = React.forwardRef<any, P & { 
    content?: string; 
    className?: string;
    onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
  }>((props, ref) => {
    const { content = '', className = '', onError, ...componentProps } = props;
    
    return (
      <DiagramErrorBoundary
        content={content}
        diagramType={diagramType}
        className={className}
        onError={onError}
      >
        <Component {...(componentProps as P)} ref={ref} />
      </DiagramErrorBoundary>
    );
  });

  WrappedComponent.displayName = `withDiagramErrorBoundary(${Component.displayName || Component.name})`;
  
  return WrappedComponent;
}

export default DiagramErrorBoundary;