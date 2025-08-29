import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useResponsiveLayout } from '../hooks/useResponsiveLayout';
import { useDiagramLibrary } from '../contexts/DiagramLibraryContext';
import { DiagramLoadingIndicator, DiagramErrorIndicator } from './DiagramLoadingIndicator';
import ResponsiveDiagramContainer from './ResponsiveDiagramContainer';

/**
 * Props interface for the ResponsiveMermaidDiagram component
 */
interface ResponsiveMermaidDiagramProps {
  /** Mermaid diagram content */
  content: string;
  /** Optional CSS class name */
  className?: string;
  /** Theme for the diagram */
  theme?: 'light' | 'dark' | 'auto' | 'high-contrast';
  /** Whether to enable zoom and pan controls */
  enableZoomPan?: boolean;
  /** Whether to enable touch gestures on mobile */
  enableTouchGestures?: boolean;
  /** Whether to fit diagram to container initially */
  fitToContainer?: boolean;
  /** Alternative text description for screen readers */
  alt?: string;
  /** Detailed description for accessibility */
  description?: string;
  /** ARIA label for the diagram */
  ariaLabel?: string;
  /** Whether to enable keyboard navigation */
  enableKeyboardNavigation?: boolean;
  /** Callback when diagram is successfully rendered */
  onRenderSuccess?: (svg: string) => void;
  /** Callback when diagram rendering fails */
  onRenderError?: (error: string) => void;
}

/**
 * Responsive Mermaid diagram component with mobile optimization
 * Provides touch-friendly controls and viewport-based rendering
 */
export const ResponsiveMermaidDiagram: React.FC<ResponsiveMermaidDiagramProps> = ({
  content,
  className = '',
  theme = 'auto',
  enableZoomPan = true,
  enableTouchGestures = true,
  fitToContainer = true,
  alt,
  description,
  ariaLabel,
  enableKeyboardNavigation = true,
  onRenderSuccess,
  onRenderError
}) => {
  const { isMobile } = useResponsiveLayout();
  const { libraryState, loadMermaid } = useDiagramLibrary();
  
  // Component state
  const [svg, setSvg] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [isRendering, setIsRendering] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  
  // Refs
  const diagramRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const renderTimeoutRef = useRef<NodeJS.Timeout>();

  // Keyboard navigation state
  const [focusedElement, setFocusedElement] = useState<number>(-1);
  const [interactiveElements, setInteractiveElements] = useState<Element[]>([]);

  // Determine effective theme
  const effectiveTheme = useMemo(() => {
    if (theme === 'auto') {
      // Auto-detect based on system preference or default to light
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    if (theme === 'high-contrast') {
      return 'high-contrast';
    }
    return theme;
  }, [theme]);

  // Generate accessible description from diagram content
  const generateAccessibleDescription = useMemo(() => {
    if (description) return description;
    if (alt) return alt;
    
    // Basic description generation based on diagram type
    const trimmedContent = content.trim().toLowerCase();
    
    if (trimmedContent.includes('graph')) {
      return 'A flowchart diagram showing connected nodes and relationships';
    } else if (trimmedContent.includes('sequencediagram')) {
      return 'A sequence diagram showing interactions between participants over time';
    } else if (trimmedContent.includes('gantt')) {
      return 'A Gantt chart showing project timeline and task dependencies';
    } else if (trimmedContent.includes('pie')) {
      return 'A pie chart showing data distribution';
    } else if (trimmedContent.includes('journey')) {
      return 'A user journey diagram showing user interactions and experiences';
    } else if (trimmedContent.includes('class')) {
      return 'A class diagram showing object-oriented relationships';
    } else if (trimmedContent.includes('state')) {
      return 'A state diagram showing system states and transitions';
    } else if (trimmedContent.includes('er')) {
      return 'An entity relationship diagram showing database relationships';
    }
    
    return 'A Mermaid diagram with visual information';
  }, [content, description, alt]);

  // Mobile-optimized Mermaid configuration with accessibility support
  const getMermaidConfig = useMemo(() => {
    // High contrast theme variables
    const getThemeVariables = () => {
      if (effectiveTheme === 'high-contrast') {
        return {
          primaryColor: '#000000',
          primaryTextColor: '#000000',
          primaryBorderColor: '#000000',
          lineColor: '#000000',
          secondaryColor: '#FFFFFF',
          tertiaryColor: '#FFFFFF',
          background: '#FFFFFF',
          mainBkg: '#FFFFFF',
          secondBkg: '#F0F0F0',
          tertiaryBkg: '#E0E0E0',
        };
      } else if (effectiveTheme === 'dark') {
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

    const baseConfig = {
      theme: effectiveTheme === 'high-contrast' ? 'base' : (effectiveTheme === 'dark' ? 'dark' : 'default'),
      themeVariables: getThemeVariables(),
      fontFamily: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
      fontSize: isMobile ? 12 : 14,
      securityLevel: 'strict' as const,
      maxTextSize: 50000,
      htmlLabels: false,
      flowchart: {
        useMaxWidth: true,
        htmlLabels: false,
        curve: 'basis',
        padding: isMobile ? 10 : 20,
      },
      sequence: {
        useMaxWidth: true,
        diagramMarginX: isMobile ? 20 : 50,
        diagramMarginY: isMobile ? 10 : 20,
        actorMargin: isMobile ? 30 : 50,
        width: isMobile ? 120 : 150,
        height: isMobile ? 45 : 65,
        boxMargin: isMobile ? 5 : 10,
        boxTextMargin: isMobile ? 3 : 5,
        noteMargin: isMobile ? 5 : 10,
        messageMargin: isMobile ? 25 : 35,
      },
      gantt: {
        useMaxWidth: true,
        leftPadding: isMobile ? 50 : 75,
        gridLineStartPadding: isMobile ? 25 : 35,
        fontSize: isMobile ? 10 : 11,
        sectionFontSize: isMobile ? 18 : 24,
        numberSectionStyles: 4,
      },
      journey: {
        useMaxWidth: true,
        diagramMarginX: isMobile ? 20 : 50,
        diagramMarginY: isMobile ? 10 : 20,
        leftMargin: isMobile ? 100 : 150,
        width: isMobile ? 120 : 150,
        height: isMobile ? 45 : 65,
        boxMargin: isMobile ? 5 : 10,
        boxTextMargin: isMobile ? 3 : 5,
        noteMargin: isMobile ? 5 : 10,
        messageMargin: isMobile ? 25 : 35,
      },
      class: {
        useMaxWidth: true,
      },
      state: {
        useMaxWidth: true,
      },
      er: {
        useMaxWidth: true,
        diagramPadding: isMobile ? 10 : 20,
        layoutDirection: 'TB',
        minEntityWidth: isMobile ? 80 : 100,
        minEntityHeight: isMobile ? 60 : 75,
        entityPadding: isMobile ? 10 : 15,
        stroke: effectiveTheme === 'dark' ? '#6B7280' : '#374151',
        fill: effectiveTheme === 'dark' ? '#1F2937' : '#FFFFFF',
      },
      pie: {
        useMaxWidth: true,
        textPosition: 0.75,
      },
      quadrantChart: {
        useMaxWidth: true,
        chartWidth: isMobile ? 300 : 500,
        chartHeight: isMobile ? 300 : 400,
      },
      requirement: {
        useMaxWidth: true,
        rect_fill: effectiveTheme === 'dark' ? '#374151' : '#F3F4F6',
        text_color: effectiveTheme === 'dark' ? '#F3F4F6' : '#1F2937',
        rect_border_size: '0.5px',
        rect_border_color: effectiveTheme === 'dark' ? '#6B7280' : '#D1D5DB',
      },
      gitgraph: {
        useMaxWidth: true,
        diagramPadding: isMobile ? 5 : 8,
        nodeLabel: {
          width: isMobile ? 60 : 75,
          height: isMobile ? 20 : 30,
          x: isMobile ? -20 : -25,
          y: isMobile ? -5 : -8,
        },
      },
      c4: {
        useMaxWidth: true,
        diagramMarginX: isMobile ? 20 : 50,
        diagramMarginY: isMobile ? 10 : 20,
        c4ShapeMargin: isMobile ? 30 : 50,
        c4ShapePadding: isMobile ? 10 : 20,
        width: isMobile ? 140 : 216,
        height: isMobile ? 60 : 60,
        boxMargin: isMobile ? 5 : 10,
      },
    };

    return baseConfig;
  }, [effectiveTheme, isMobile]);

  /**
   * Render the Mermaid diagram
   */
  const renderDiagram = async () => {
    if (!content.trim() || !libraryState.mermaid.loaded) {
      return;
    }

    setIsRendering(true);
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
        }, 15000); // 15 second timeout for complex diagrams
      });

      // Generate unique ID for this diagram
      const diagramId = `mermaid-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

      // Dynamic import and render
      const renderPromise = (async () => {
        const mermaid = await import('mermaid');
        
        // Initialize with mobile-optimized config
        mermaid.default.initialize(getMermaidConfig as any);

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
        .replace(/<svg/, `<svg class="w-full h-auto max-w-full" role="img" aria-labelledby="diagram-title-${diagramId}" aria-describedby="diagram-desc-${diagramId}"`)
        .replace(/width="[^"]*"/, '')
        .replace(/height="[^"]*"/, '');

      // Add title and description elements for accessibility
      const accessibleSvg = responsiveSvg.replace(
        /<svg([^>]*)>/,
        `<svg$1>
          <title id="diagram-title-${diagramId}">${ariaLabel || alt || 'Mermaid Diagram'}</title>
          <desc id="diagram-desc-${diagramId}">${generateAccessibleDescription}</desc>`
      );

      setSvg(accessibleSvg);
      setError('');
      setRetryCount(0);
      onRenderSuccess?.(accessibleSvg);

      // Update interactive elements for keyboard navigation
      if (enableKeyboardNavigation) {
        setTimeout(() => {
          // Find interactive elements in the rendered SVG
          const svgElement = document.querySelector(`#${diagramId}`);
          if (svgElement) {
            const elements = Array.from(svgElement.querySelectorAll('g[class*="node"], g[class*="edge"]'));
            setInteractiveElements(elements);
          }
        }, 100);
      }
      
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
      setIsRendering(false);
    }
  };

  /**
   * Retry diagram rendering
   */
  const retryRender = () => {
    setRetryCount(prev => prev + 1);
    renderDiagram();
  };

  /**
   * Load Mermaid library if needed
   */
  useEffect(() => {
    if (!libraryState.mermaid.loaded && !libraryState.mermaid.loading) {
      loadMermaid();
    }
  }, [libraryState.mermaid, loadMermaid]);

  /**
   * Render diagram when library is loaded or content changes
   */
  useEffect(() => {
    if (libraryState.mermaid.loaded && content.trim()) {
      renderDiagram();
    }
  }, [libraryState.mermaid.loaded, content, getMermaidConfig]);

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

  // Show loading state
  if (libraryState.mermaid.loading || isRendering) {
    return (
      <DiagramLoadingIndicator
        type="mermaid"
        message={libraryState.mermaid.loading ? 'Loading Mermaid library...' : 'Rendering diagram...'}
        size={isMobile ? 'large' : 'medium'}
        className={className}
      />
    );
  }

  // Show library loading error
  if (libraryState.mermaid.error) {
    return (
      <DiagramErrorIndicator
        error={`Library Error: ${libraryState.mermaid.error}`}
        type="mermaid"
        onRetry={() => loadMermaid()}
        className={className}
      />
    );
  }

  // Show rendering error
  if (error) {
    return (
      <DiagramErrorIndicator
        error={error}
        type="mermaid"
        onRetry={retryCount < 3 ? retryRender : undefined}
        showRawContent={true}
        rawContent={content}
        className={className}
      />
    );
  }

  // Show rendered diagram
  if (svg) {
    return (
      <ResponsiveDiagramContainer
        className={className}
        enableZoomPan={enableZoomPan}
        enableTouchGestures={enableTouchGestures}
        fitToContainer={fitToContainer}
        minZoom={0.1}
        maxZoom={isMobile ? 3 : 5}
        initialZoom={1}
      >
        <div
          ref={diagramRef}
          className={`mermaid-diagram ${effectiveTheme === 'dark' ? 'dark' : 'light'}`}
          dangerouslySetInnerHTML={{ __html: svg }}
          style={{
            minWidth: 'fit-content',
            minHeight: 'fit-content',
          }}
        />
      </ResponsiveDiagramContainer>
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

export default ResponsiveMermaidDiagram;