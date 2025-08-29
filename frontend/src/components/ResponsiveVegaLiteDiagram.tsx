import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useResponsiveLayout } from '../hooks/useResponsiveLayout';
import { useDiagramLibrary } from '../contexts/DiagramLibraryContext';
import {
  DiagramLoadingIndicator,
  DiagramErrorIndicator,
} from './DiagramLoadingIndicator';
import ResponsiveDiagramContainer from './ResponsiveDiagramContainer';

/**
 * Props interface for the ResponsiveVegaLiteDiagram component
 */
interface ResponsiveVegaLiteDiagramProps {
  /** Vega-Lite JSON specification as string */
  content: string;
  /** Optional CSS class name */
  className?: string;
  /** Theme for the chart */
  theme?: 'light' | 'dark' | 'auto';
  /** Whether to enable zoom and pan controls */
  enableZoomPan?: boolean;
  /** Whether to enable touch gestures on mobile */
  enableTouchGestures?: boolean;
  /** Whether to fit chart to container initially */
  fitToContainer?: boolean;
  /** Whether to show chart actions (download, etc.) */
  showActions?: boolean;
  /** Callback when chart is successfully rendered */
  onRenderSuccess?: (spec: any) => void;
  /** Callback when chart rendering fails */
  onRenderError?: (error: string) => void;
}

/**
 * Responsive Vega-Lite chart component with mobile optimization
 * Provides touch-friendly controls and viewport-based rendering
 */
export const ResponsiveVegaLiteDiagram: React.FC<ResponsiveVegaLiteDiagramProps> = ({
  content,
  className = '',
  theme = 'auto',
  enableZoomPan = true,
  enableTouchGestures = true,
  fitToContainer = true,
  showActions = false,
  onRenderSuccess,
  onRenderError,
}) => {
  const { isMobile } = useResponsiveLayout();
  const { libraryState, loadVegaLite } = useDiagramLibrary();

  // Component state
  const [spec, setSpec] = useState<any>(null);
  const [error, setError] = useState<string>('');
  const [isRendering, setIsRendering] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [VegaLiteComponent, setVegaLiteComponent] = useState<any>(null);
  const [componentLoadError, setComponentLoadError] = useState<string>('');

  // Determine effective theme
  const effectiveTheme = useMemo(() => {
    if (theme === 'auto') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return theme;
  }, [theme]);

  /**
   * Get mobile-optimized Vega-Lite configuration
   */
  const getVegaLiteConfig = useMemo(() => {
    const baseColors = {
      light: {
        background: '#FFFFFF',
        text: '#1F2937',
        grid: '#E5E7EB',
        axis: '#9CA3AF',
        primary: '#3B82F6',
        secondary: '#10B981',
      },
      dark: {
        background: '#1F2937',
        text: '#F3F4F6',
        grid: '#374151',
        axis: '#6B7280',
        primary: '#60A5FA',
        secondary: '#34D399',
      },
    };

    const colors = baseColors[effectiveTheme];

    return {
      background: colors.background,
      title: {
        color: colors.text,
        fontSize: isMobile ? 14 : 16,
        fontWeight: 600,
        anchor: 'start',
        offset: isMobile ? 10 : 20,
      },
      axis: {
        labelColor: colors.text,
        titleColor: colors.text,
        gridColor: colors.grid,
        domainColor: colors.axis,
        tickColor: colors.axis,
        labelFontSize: isMobile ? 10 : 11,
        titleFontSize: isMobile ? 11 : 12,
        labelLimit: isMobile ? 60 : 100,
        titleLimit: isMobile ? 80 : 120,
        labelPadding: isMobile ? 4 : 6,
        titlePadding: isMobile ? 8 : 12,
      },
      legend: {
        labelColor: colors.text,
        titleColor: colors.text,
        labelFontSize: isMobile ? 10 : 11,
        titleFontSize: isMobile ? 11 : 12,
        symbolSize: isMobile ? 60 : 100,
        labelLimit: isMobile ? 80 : 120,
        titleLimit: isMobile ? 100 : 150,
        padding: isMobile ? 8 : 12,
        offset: isMobile ? 10 : 18,
      },
      range: {
        category: [
          colors.primary,
          colors.secondary,
          '#F59E0B',
          '#EF4444',
          '#8B5CF6',
          '#06B6D4',
          '#84CC16',
          '#F97316',
        ],
        heatmap:
          effectiveTheme === 'dark' ? ['#1F2937', '#3B82F6'] : ['#F3F4F6', '#1E40AF'],
      },
      mark: {
        color: colors.primary,
        fontSize: isMobile ? 10 : 11,
        strokeWidth: isMobile ? 1 : 1.5,
      },
      text: {
        color: colors.text,
        fontSize: isMobile ? 10 : 11,
      },
      arc: {
        stroke: colors.background,
        strokeWidth: 2,
      },
      area: {
        opacity: 0.7,
      },
      line: {
        strokeWidth: isMobile ? 2 : 3,
        strokeCap: 'round',
        strokeJoin: 'round',
      },
      point: {
        size: isMobile ? 60 : 100,
        strokeWidth: isMobile ? 1 : 2,
      },
      rect: {
        stroke: colors.background,
        strokeWidth: 1,
      },
      bar: {
        stroke: colors.background,
        strokeWidth: 1,
      },
    };
  }, [effectiveTheme, isMobile]);

  /**
   * Apply responsive optimizations to Vega-Lite spec
   */
  const optimizeSpecForDevice = useCallback(
    (originalSpec: any) => {
      const optimizedSpec = { ...originalSpec };

      // Apply mobile-specific sizing
      if (isMobile) {
        // Reduce default dimensions for mobile
        if (optimizedSpec.width && optimizedSpec.width > 300) {
          optimizedSpec.width = Math.min(optimizedSpec.width, 300);
        }
        if (optimizedSpec.height && optimizedSpec.height > 200) {
          optimizedSpec.height = Math.min(optimizedSpec.height, 200);
        }

        // Use container width for responsive charts
        if (!optimizedSpec.width) {
          optimizedSpec.width = 'container';
        }

        // Optimize text and labels for mobile
        if (optimizedSpec.encoding) {
          // Reduce label lengths
          Object.keys(optimizedSpec.encoding).forEach(channel => {
            const encoding = optimizedSpec.encoding[channel];
            if (encoding && encoding.axis) {
              encoding.axis = {
                ...encoding.axis,
                labelLimit: 40,
                titleLimit: 60,
                labelAngle: encoding.axis.labelAngle || (channel === 'x' ? -45 : 0),
              };
            }
          });
        }

        // Optimize legend for mobile
        if (optimizedSpec.resolve && optimizedSpec.resolve.legend) {
          optimizedSpec.resolve.legend = { ...optimizedSpec.resolve.legend };
        }
      }

      // Apply theme configuration
      optimizedSpec.config = {
        ...optimizedSpec.config,
        ...getVegaLiteConfig,
        view: {
          ...optimizedSpec.config?.view,
          continuousWidth: isMobile ? 280 : 400,
          continuousHeight: isMobile ? 180 : 300,
          discreteWidth: isMobile ? 20 : 30,
          discreteHeight: isMobile ? 20 : 30,
        },
        autosize: {
          type: 'fit',
          contains: 'padding',
          resize: true,
        },
      };

      // Ensure responsive behavior
      if (!optimizedSpec.autosize) {
        optimizedSpec.autosize = {
          type: 'fit',
          contains: 'padding',
          resize: true,
        };
      }

      return optimizedSpec;
    },
    [isMobile, getVegaLiteConfig]
  );

  /**
   * Parse and validate Vega-Lite specification
   */
  const parseSpec = useCallback(async () => {
    if (!content.trim() || !libraryState.vegaLite.loaded) {
      return;
    }

    setIsRendering(true);
    setError('');

    try {
      // Parse JSON specification
      const parsedSpec = JSON.parse(content.trim());

      // Basic validation
      if (!parsedSpec || typeof parsedSpec !== 'object') {
        throw new Error('Invalid Vega-Lite specification: must be a valid JSON object');
      }

      if (
        !parsedSpec.mark &&
        !parsedSpec.layer &&
        !parsedSpec.concat &&
        !parsedSpec.facet &&
        !parsedSpec.repeat
      ) {
        throw new Error(
          'Invalid Vega-Lite specification: missing required mark or composition property'
        );
      }

      // Security validation - remove potentially dangerous properties
      const sanitizedSpec = { ...parsedSpec };
      delete sanitizedSpec.datasets; // Remove external data references
      delete sanitizedSpec.transform; // Remove transform functions that could be dangerous

      // Apply responsive optimizations
      const optimizedSpec = optimizeSpecForDevice(sanitizedSpec);

      setSpec(optimizedSpec);
      setError('');
      setRetryCount(0);
      onRenderSuccess?.(optimizedSpec);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to parse Vega-Lite specification';
      console.error('Vega-Lite parsing error:', err);

      setError(errorMessage);
      setSpec(null);
      onRenderError?.(errorMessage);
    } finally {
      setIsRendering(false);
    }
  }, [
    content,
    libraryState.vegaLite.loaded,
    optimizeSpecForDevice,
    onRenderSuccess,
    onRenderError,
  ]);

  /**
   * Retry chart rendering
   */
  const retryRender = () => {
    setRetryCount(prev => prev + 1);
    parseSpec();
  };

  /**
   * Load Vega-Lite library and component if needed
   */
  useEffect(() => {
    if (!libraryState.vegaLite.loaded && !libraryState.vegaLite.loading) {
      loadVegaLite();
    }
  }, [libraryState.vegaLite, loadVegaLite]);

  /**
   * Load VegaLite React component when library is ready
   */
  useEffect(() => {
    if (libraryState.vegaLite.loaded && !VegaLiteComponent && !componentLoadError) {
      // First try to get from the loaded module
      const reactVegaModule = libraryState.vegaLite.module?.reactVega;
      console.log('Context module structure:', {
        module: libraryState.vegaLite.module,
        reactVegaModule,
        VegaLite: reactVegaModule?.VegaLite,
        VegaEmbed: reactVegaModule?.VegaEmbed,
        default: reactVegaModule?.default,
      });

      // Try to get any available component from context
      const Component =
        reactVegaModule?.VegaLite ||
        reactVegaModule?.VegaEmbed ||
        reactVegaModule?.default;
      if (Component) {
        console.log('Using component from context:', Component.name || 'unnamed');
        setVegaLiteComponent(() => Component);
        setComponentLoadError('');
        return;
      }

      // Fallback to direct import with longer timeout
      console.log('Fallback: importing react-vega directly');
      const loadTimer = setTimeout(() => {
        import('react-vega')
          .then(module => {
            console.log('Direct import successful, module:', module);
            console.log('Module keys:', Object.keys(module));
            console.log('Module default:', module.default);

            // Check all available exports
            console.log('All module exports:', Object.keys(module));
            console.log('Default export:', module.default);
            console.log('VegaEmbed:', module.VegaEmbed);
            console.log('VegaLite:', module.VegaLite);

            // Try different ways to get the component
            let Component = module.VegaLite || module.VegaEmbed || module.default;

            console.log('Selected component:', Component);

            if (Component) {
              setVegaLiteComponent(() => Component);
              setComponentLoadError('');
            } else {
              throw new Error('No suitable Vega component found in react-vega module');
            }
          })
          .catch(err => {
            console.error('Direct import failed:', err);
            const errorMsg = 'Failed to load chart rendering component';
            setComponentLoadError(errorMsg);
            setError(errorMsg);
          });
      }, 200); // Longer delay

      // Set a longer timeout
      const timeoutTimer = setTimeout(() => {
        if (!VegaLiteComponent) {
          // console.error('Timeout loading chart component');
          // setComponentLoadError('Timeout loading chart component');
        }
      }, 10000); // 10 second timeout

      return () => {
        clearTimeout(loadTimer);
        clearTimeout(timeoutTimer);
      };
    }
  }, [
    libraryState.vegaLite.loaded,
    libraryState.vegaLite.module,
    VegaLiteComponent,
    componentLoadError,
  ]);

  /**
   * Parse specification when library is loaded or content changes
   */
  useEffect(() => {
    if (libraryState.vegaLite.loaded && VegaLiteComponent && content.trim()) {
      parseSpec();
    }
  }, [libraryState.vegaLite.loaded, VegaLiteComponent, content, parseSpec]);

  // Show loading state
  if (
    libraryState.vegaLite.loading ||
    isRendering ||
    (libraryState.vegaLite.loaded && !VegaLiteComponent && !componentLoadError)
  ) {
    return (
      <DiagramLoadingIndicator
        type='vega-lite'
        message={
          libraryState.vegaLite.loading
            ? 'Loading Vega-Lite library...'
            : !VegaLiteComponent
            ? 'Loading chart component...'
            : 'Rendering chart...'
        }
        size={isMobile ? 'large' : 'medium'}
        className={className}
      />
    );
  }

  // Show library loading error
  if (libraryState.vegaLite.error) {
    return (
      <DiagramErrorIndicator
        error={`Library Error: ${libraryState.vegaLite.error}`}
        type='vega-lite'
        onRetry={() => loadVegaLite()}
        className={className}
      />
    );
  }

  // Show component loading error
  if (componentLoadError) {
    return (
      <DiagramErrorIndicator
        error={componentLoadError}
        type='vega-lite'
        onRetry={() => {
          setComponentLoadError('');
          setError('');
          setVegaLiteComponent(null);
        }}
        className={className}
      />
    );
  }

  // Show parsing/rendering error
  if (error) {
    return (
      <DiagramErrorIndicator
        error={error}
        type='vega-lite'
        onRetry={retryCount < 3 ? retryRender : undefined}
        showRawContent={true}
        rawContent={content}
        className={className}
      />
    );
  }

  // Show rendered chart
  if (spec && VegaLiteComponent) {
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
          className={`vega-lite-diagram ${effectiveTheme === 'dark' ? 'dark' : 'light'}`}
        >
          <VegaLiteComponent
            spec={spec}
            actions={showActions}
            renderer='svg'
            tooltip={true}
            hover={true}
            onError={(error: any) => {
              console.error('Vega-Lite rendering error:', error);
              setError(error.message || 'Chart rendering failed');
            }}
            style={{
              width: '100%',
              height: 'auto',
            }}
          />
        </div>
      </ResponsiveDiagramContainer>
    );
  }

  // Empty state
  return (
    <div className={`vega-lite-empty ${className}`}>
      <div className='flex items-center justify-center p-8 bg-gray-50 border border-gray-200 rounded-lg'>
        <p className='text-sm text-gray-500'>No chart specification to display</p>
      </div>
    </div>
  );
};

export default ResponsiveVegaLiteDiagram;
