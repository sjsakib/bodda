import React, { useState, useEffect, useRef, useMemo } from 'react';

/**
 * Props interface for the MermaidDiagram component
 */
interface MermaidDiagramProps {
  /** Mermaid diagram content */
  content: string;
  /** Optional CSS class name */
  className?: string;
  /** Theme for the diagram */
  theme?: 'light' | 'dark';
  /** Alternative text description for screen readers */
  alt?: string;
  /** Callback when diagram is successfully rendered */
  onRenderSuccess?: (svg: string) => void;
  /** Callback when diagram rendering fails */
  onRenderError?: (error: string) => void;
}

/**
 * Core Mermaid diagram component with proper TypeScript interfaces,
 * theme configuration, SVG rendering with error boundaries,
 * loading states and error handling for invalid syntax.
 * 
 * This component provides the fundamental Mermaid rendering functionality
 * as specified in the design document, with mobile optimization and
 * responsive styling.
 */
export const MermaidDiagram: React.FC<MermaidDiagramProps> = ({
  content,
  className = '',
  theme = 'light',
  alt,
  onRenderSuccess,
  onRenderError
}) => {
  // Component state
  const [svg, setSvg] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState(true);
  
  // Refs
  const diagramRef = useRef<HTMLDivElement>(null);
  const renderTimeoutRef = useRef<NodeJS.Timeout>();

  // Generate unique diagram ID
  const diagramId = useMemo(() => 
    `mermaid-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`, 
    [content]
  );

  // Mermaid configuration with theme support
  const getMermaidConfig = useMemo(() => {
    const getThemeVariables = () => {
      if (theme === 'dark') {
        return {
          primaryColor: '#3B82F6',
          primaryTextColor: '#F3F4F6',
          primaryBorderColor: '#6B7280',
          lineColor: '#6B7280',
          secondaryColor: '#374151',
          tertiaryColor: '#1F2937',
          background: '#1F2937',
          mainBkg: '#1F2937',
          secondBkg: '#374151',
          tertiaryBkg: '#4B5563',
        };
      } else {
        return {
          primaryColor: '#1E40AF',
          primaryTextColor: '#1F2937',
          primaryBorderColor: '#D1D5DB',
          lineColor: '#374151',
          secondaryColor: '#F3F4F6',
          tertiaryColor: '#FFFFFF',
          background: '#FFFFFF',
          mainBkg: '#FFFFFF',
          secondBkg: '#F9FAFB',
          tertiaryBkg: '#F3F4F6',
        };
      }
    };

    return {
      theme: (theme === 'dark' ? 'dark' : 'default') as 'dark' | 'default',
      themeVariables: getThemeVariables(),
      fontFamily: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
      fontSize: 14,
      securityLevel: 'strict' as const,
      maxTextSize: 50000,
      htmlLabels: false,
      flowchart: {
        useMaxWidth: true,
        htmlLabels: false,
        curve: 'basis',
        padding: 20,
      },
      sequence: {
        useMaxWidth: true,
        diagramMarginX: 50,
        diagramMarginY: 20,
        actorMargin: 50,
        width: 150,
        height: 65,
        boxMargin: 10,
        boxTextMargin: 5,
        noteMargin: 10,
        messageMargin: 35,
      },
      gantt: {
        useMaxWidth: true,
        leftPadding: 75,
        gridLineStartPadding: 35,
        fontSize: 11,
        sectionFontSize: 24,
        numberSectionStyles: 4,
      },
      journey: {
        useMaxWidth: true,
        diagramMarginX: 50,
        diagramMarginY: 20,
        leftMargin: 150,
        width: 150,
        height: 65,
        boxMargin: 10,
        boxTextMargin: 5,
        noteMargin: 10,
        messageMargin: 35,
      },
      class: {
        useMaxWidth: true,
      },
      state: {
        useMaxWidth: true,
      },
      er: {
        useMaxWidth: true,
        diagramPadding: 20,
        layoutDirection: 'TB',
        minEntityWidth: 100,
        minEntityHeight: 75,
        entityPadding: 15,
        stroke: theme === 'dark' ? '#6B7280' : '#374151',
        fill: theme === 'dark' ? '#1F2937' : '#FFFFFF',
      },
      pie: {
        useMaxWidth: true,
        textPosition: 0.75,
      },
    };
  }, [theme]);

  /**
   * Render the Mermaid diagram
   */
  const renderDiagram = async () => {
    if (!content.trim()) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Clear any existing timeout
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }

      // Set rendering timeout
      const timeoutPromise = new Promise<never>((_, reject) => {
        renderTimeoutRef.current = setTimeout(() => {
          reject(new Error('Diagram rendering timeout'));
        }, 10000); // 10 second timeout
      });

      // Dynamic import and render
      const renderPromise = (async () => {
        const mermaid = await import('mermaid');
        
        // Initialize with theme configuration
        mermaid.default.initialize(getMermaidConfig);

        // Render the diagram
        const { svg: renderedSvg } = await mermaid.default.render(diagramId, content.trim());
        
        return renderedSvg;
      })();

      const renderedSvg = await Promise.race([renderPromise, timeoutPromise]);
      
      // Clear timeout on success
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }

      // Apply responsive styling and accessibility attributes to SVG
      const responsiveSvg = renderedSvg
        .replace(/<svg/, `<svg class="w-full h-auto max-w-full" role="img" aria-labelledby="diagram-title-${diagramId}"`)
        .replace(/width="[^"]*"/, '')
        .replace(/height="[^"]*"/, '');

      // Add title element for accessibility
      const accessibleSvg = responsiveSvg.replace(
        /<svg([^>]*)>/,
        `<svg$1>
          <title id="diagram-title-${diagramId}">${alt || 'Mermaid Diagram'}</title>`
      );

      setSvg(accessibleSvg);
      setError('');
      onRenderSuccess?.(accessibleSvg);
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to render Mermaid diagram';
      console.error('Mermaid rendering error:', err);
      
      setError(errorMessage);
      setSvg('');
      onRenderError?.(errorMessage);
      
      // Clear timeout on error
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }
    } finally {
      setLoading(false);
    }
  };

  /**
   * Render diagram when content or theme changes
   */
  useEffect(() => {
    renderDiagram();
  }, [content, getMermaidConfig]);

  /**
   * Cleanup timeout on unmount
   */
  useEffect(() => {
    return () => {
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }
    };
  }, []);

  // Loading state
  if (loading) {
    return (
      <div className={`mermaid-loading ${className}`}>
        <div className="flex items-center justify-center p-8 bg-gray-50 rounded-lg border border-gray-200">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-sm text-gray-600">Rendering diagram...</span>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={`mermaid-error ${className}`}>
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-sm text-red-800 font-medium">Diagram Error:</p>
          <p className="text-sm text-red-600 mt-1">{error}</p>
          <details className="mt-2">
            <summary className="text-xs text-red-500 cursor-pointer">Show raw content</summary>
            <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto">
              {content}
            </pre>
          </details>
        </div>
      </div>
    );
  }

  // Rendered diagram
  if (svg) {
    return (
      <div className={`mermaid-diagram ${className}`}>
        <div className="relative overflow-auto">
          <div 
            ref={diagramRef}
            dangerouslySetInnerHTML={{ __html: svg }}
            className="mermaid-svg-container"
            style={{
              minWidth: 'fit-content',
              minHeight: 'fit-content',
            }}
          />
        </div>
      </div>
    );
  }

  // Empty state
  return (
    <div className={`mermaid-empty ${className}`}>
      <div className="flex items-center justify-center p-8 bg-gray-50 border border-gray-200 rounded-lg">
        <p className="text-sm text-gray-500">No diagram content to display</p>
      </div>
    </div>
  );
};

/**
 * Error boundary component for Mermaid diagram rendering
 */
class MermaidDiagramErrorBoundary extends React.Component<
  { 
    children: React.ReactNode; 
    content: string; 
    className: string;
    fallback?: React.ReactNode;
  },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { 
    children: React.ReactNode; 
    content: string; 
    className: string;
    fallback?: React.ReactNode;
  }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Mermaid diagram rendering error:', {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString()
    });
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className={`diagram-error ${this.props.className}`}>
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium">Failed to render diagram</p>
            <p className="text-xs text-red-600 mt-1">{this.state.error?.message}</p>
            <details className="mt-2">
              <summary className="text-xs text-red-500 cursor-pointer">Show raw content</summary>
              <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto">
                {this.props.content}
              </pre>
            </details>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * Safe wrapper component that handles Mermaid diagram rendering errors gracefully.
 * Falls back to error display if diagram rendering fails.
 */
export const SafeMermaidDiagram: React.FC<MermaidDiagramProps> = (props) => {
  return (
    <MermaidDiagramErrorBoundary 
      content={props.content} 
      className={props.className || ''}
    >
      <MermaidDiagram {...props} />
    </MermaidDiagramErrorBoundary>
  );
};

export default MermaidDiagram;