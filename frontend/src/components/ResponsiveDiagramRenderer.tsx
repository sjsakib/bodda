import React from 'react';
import ResponsiveMermaidDiagram from './ResponsiveMermaidDiagram';
import ResponsiveVegaLiteDiagram from './ResponsiveVegaLiteDiagram';

/**
 * Props interface for the ResponsiveDiagramRenderer component
 */
interface ResponsiveDiagramRendererProps {
  /** Type of diagram to render */
  type: 'mermaid' | 'vega-lite';
  /** Diagram content */
  content: string;
  /** Optional CSS class name */
  className?: string;
  /** Theme for the diagram */
  theme?: 'light' | 'dark' | 'auto';
  /** Whether to enable zoom and pan controls */
  enableZoomPan?: boolean;
  /** Whether to enable touch gestures on mobile */
  enableTouchGestures?: boolean;
  /** Whether to fit diagram to container initially */
  fitToContainer?: boolean;
  /** Whether to show actions for Vega-Lite charts */
  showActions?: boolean;
  /** Callback when diagram is successfully rendered */
  onRenderSuccess?: (result: any) => void;
  /** Callback when diagram rendering fails */
  onRenderError?: (error: string) => void;
}

/**
 * Unified responsive diagram renderer that handles both Mermaid and Vega-Lite diagrams
 * Provides consistent interface and responsive behavior across diagram types
 */
export const ResponsiveDiagramRenderer: React.FC<ResponsiveDiagramRendererProps> = ({
  type,
  content,
  className = '',
  theme = 'auto',
  enableZoomPan = true,
  enableTouchGestures = true,
  fitToContainer = true,
  showActions = false,
  onRenderSuccess,
  onRenderError
}) => {
  // Common props for both diagram types
  const commonProps = {
    content,
    className,
    theme,
    enableZoomPan,
    enableTouchGestures,
    fitToContainer,
    onRenderSuccess,
    onRenderError,
  };

  // Render appropriate diagram component based on type
  switch (type) {
    case 'mermaid':
      return <ResponsiveMermaidDiagram {...commonProps} />;
      
    case 'vega-lite':
      return (
        <ResponsiveVegaLiteDiagram 
          {...commonProps} 
          showActions={showActions}
        />
      );
      
    default:
      return (
        <div className={`diagram-error ${className}`}>
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium">Unknown Diagram Type</p>
            <p className="text-sm text-red-600 mt-1">
              Unsupported diagram type: {type}. Supported types are 'mermaid' and 'vega-lite'.
            </p>
          </div>
        </div>
      );
  }
};

export default ResponsiveDiagramRenderer;