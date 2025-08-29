import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';

/**
 * Props interface for the VegaLiteDiagram component
 */
interface VegaLiteDiagramProps {
  /** Vega-Lite JSON specification as string */
  content: string;
  /** Optional CSS class name */
  className?: string;
  /** Theme for the chart */
  theme?: 'light' | 'dark';
  /** Alternative text description for screen readers */
  alt?: string;
  /** Whether to show chart actions (download, etc.) */
  showActions?: boolean;
  /** Whether to enable tooltips */
  enableTooltips?: boolean;
  /** Whether to enable hover effects */
  enableHover?: boolean;
  /** Callback when chart is successfully rendered */
  onRenderSuccess?: (spec: any) => void;
  /** Callback when chart rendering fails */
  onRenderError?: (error: string) => void;
}

/**
 * Core Vega-Lite chart component with proper TypeScript interfaces,
 * spec parsing and validation, chart rendering with react-vega integration,
 * theme application and responsive design, and interactive features.
 *
 * This component provides the fundamental Vega-Lite rendering functionality
 * as specified in the design document, with error handling and accessibility.
 */
export const VegaLiteDiagram: React.FC<VegaLiteDiagramProps> = ({
  content,
  className = '',
  theme = 'light',
  alt,
  showActions = false,
  enableTooltips = true,
  enableHover = true,
  onRenderSuccess,
  onRenderError,
}) => {
  // Component state
  const [spec, setSpec] = useState<any>(null);
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [VegaLiteComponent, setVegaLiteComponent] = useState<any>(null);
  const [componentLoadError, setComponentLoadError] = useState<string>('');

  // Refs
  const chartRef = useRef<HTMLDivElement>(null);
  const renderTimeoutRef = useRef<NodeJS.Timeout>();

  // Generate unique chart ID
  const chartId = useMemo(
    () => `vega-lite-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
    [content]
  );

  /**
   * Get theme-specific Vega-Lite configuration
   */
  const getVegaLiteConfig = useMemo(() => {
    const getThemeColors = () => {
      if (theme === 'dark') {
        return {
          background: '#1F2937',
          text: '#F3F4F6',
          grid: '#374151',
          axis: '#6B7280',
          primary: '#60A5FA',
          secondary: '#34D399',
          tertiary: '#FBBF24',
          quaternary: '#F87171',
          quinary: '#A78BFA',
          senary: '#22D3EE',
        };
      } else {
        return {
          background: '#FFFFFF',
          text: '#1F2937',
          grid: '#E5E7EB',
          axis: '#9CA3AF',
          primary: '#3B82F6',
          secondary: '#10B981',
          tertiary: '#F59E0B',
          quaternary: '#EF4444',
          quinary: '#8B5CF6',
          senary: '#06B6D4',
        };
      }
    };

    const colors = getThemeColors();

    return {
      background: colors.background,
      title: {
        color: colors.text,
        fontSize: 16,
        fontWeight: 600,
        anchor: 'start',
        offset: 20,
        font: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
      },
      axis: {
        labelColor: colors.text,
        titleColor: colors.text,
        gridColor: colors.grid,
        domainColor: colors.axis,
        tickColor: colors.axis,
        labelFontSize: 11,
        titleFontSize: 12,
        labelFont: 'ui-sans-serif, system-ui, sans-serif',
        titleFont: 'ui-sans-serif, system-ui, sans-serif',
        labelLimit: 100,
        titleLimit: 120,
        labelPadding: 6,
        titlePadding: 12,
      },
      legend: {
        labelColor: colors.text,
        titleColor: colors.text,
        labelFontSize: 11,
        titleFontSize: 12,
        labelFont: 'ui-sans-serif, system-ui, sans-serif',
        titleFont: 'ui-sans-serif, system-ui, sans-serif',
        symbolSize: 100,
        labelLimit: 120,
        titleLimit: 150,
        padding: 12,
        offset: 18,
      },
      range: {
        category: [
          colors.primary,
          colors.secondary,
          colors.tertiary,
          colors.quaternary,
          colors.quinary,
          colors.senary,
          '#84CC16',
          '#F97316',
        ],
        heatmap: theme === 'dark' ? ['#1F2937', '#3B82F6'] : ['#F3F4F6', '#1E40AF'],
        ramp:
          theme === 'dark'
            ? ['#1F2937', '#374151', '#4B5563', '#6B7280', '#9CA3AF', '#D1D5DB']
            : ['#F9FAFB', '#F3F4F6', '#E5E7EB', '#D1D5DB', '#9CA3AF', '#6B7280'],
      },
      mark: {
        color: colors.primary,
        fontSize: 11,
        font: 'ui-sans-serif, system-ui, sans-serif',
        strokeWidth: 1.5,
        opacity: 0.8,
      },
      text: {
        color: colors.text,
        fontSize: 11,
        font: 'ui-sans-serif, system-ui, sans-serif',
      },
      arc: {
        stroke: colors.background,
        strokeWidth: 2,
      },
      area: {
        opacity: 0.7,
        stroke: null,
      },
      line: {
        strokeWidth: 3,
        strokeCap: 'round',
        strokeJoin: 'round',
      },
      point: {
        size: 100,
        strokeWidth: 2,
        stroke: colors.background,
      },
      rect: {
        stroke: colors.background,
        strokeWidth: 1,
      },
      bar: {
        stroke: colors.background,
        strokeWidth: 1,
      },
      view: {
        continuousWidth: 400,
        continuousHeight: 300,
        discreteWidth: 30,
        discreteHeight: 30,
      },
      autosize: {
        type: 'fit',
        contains: 'padding',
        resize: true,
      },
    };
  }, [theme]);

  /**
   * Sanitize and validate Vega-Lite specification
   */
  const sanitizeSpec = useCallback((rawSpec: any) => {
    if (!rawSpec || typeof rawSpec !== 'object') {
      throw new Error('Invalid Vega-Lite specification: must be a valid JSON object');
    }

    // Check for required properties
    if (
      !rawSpec.mark &&
      !rawSpec.layer &&
      !rawSpec.concat &&
      !rawSpec.facet &&
      !rawSpec.repeat
    ) {
      throw new Error(
        'Invalid Vega-Lite specification: missing required mark or composition property'
      );
    }

    // Create sanitized copy
    const sanitizedSpec = { ...rawSpec };

    // Remove potentially dangerous properties for security
    delete sanitizedSpec.datasets; // Remove external data references

    // Remove dangerous transform functions
    if (sanitizedSpec.transform) {
      sanitizedSpec.transform = sanitizedSpec.transform.filter((t: any) => {
        // Allow safe transforms only
        const safeTransforms = [
          'aggregate',
          'bin',
          'calculate',
          'density',
          'filter',
          'flatten',
          'fold',
          'impute',
          'joinaggregate',
          'loess',
          'lookup',
          'pivot',
          'quantile',
          'regression',
          'sample',
          'stack',
          'timeunit',
          'window',
        ];
        return t && typeof t === 'object' && safeTransforms.includes(Object.keys(t)[0]);
      });
    }

    // Validate data size if inline data is provided
    if (
      sanitizedSpec.data &&
      sanitizedSpec.data.values &&
      Array.isArray(sanitizedSpec.data.values)
    ) {
      if (sanitizedSpec.data.values.length > 5000) {
        throw new Error('Dataset too large: maximum 5000 data points allowed');
      }
    }

    return sanitizedSpec;
  }, []);

  /**
   * Apply theme configuration to specification
   */
  const applyThemeToSpec = useCallback(
    (originalSpec: any) => {
      // Deep merge config to preserve existing properties
      const mergeConfig = (original: any, theme: any) => {
        const merged = { ...theme };

        if (original) {
          Object.keys(original).forEach(key => {
            if (typeof original[key] === 'object' && typeof theme[key] === 'object') {
              merged[key] = { ...theme[key], ...original[key] };
            } else {
              merged[key] = original[key];
            }
          });
        }

        return merged;
      };

      const themedSpec = {
        ...originalSpec,
        config: mergeConfig(originalSpec.config, getVegaLiteConfig),
      };

      // Ensure responsive behavior
      if (!themedSpec.autosize) {
        themedSpec.autosize = {
          type: 'fit',
          contains: 'padding',
          resize: true,
        };
      }

      return themedSpec;
    },
    [getVegaLiteConfig]
  );

  /**
   * Parse and process Vega-Lite specification
   */
  const parseSpec = useCallback(async () => {
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

      // Set parsing timeout
      const timeoutPromise = new Promise<never>((_, reject) => {
        renderTimeoutRef.current = setTimeout(() => {
          reject(new Error('Chart parsing timeout'));
        }, 10000); // 10 second timeout
      });

      // Parse and process specification
      const parsePromise = (async () => {
        // Parse JSON specification
        const parsedSpec = JSON.parse(content.trim());

        // Sanitize specification
        const sanitizedSpec = sanitizeSpec(parsedSpec);

        // Apply theme configuration
        const themedSpec = applyThemeToSpec(sanitizedSpec);

        return themedSpec;
      })();

      const processedSpec = await Promise.race([parsePromise, timeoutPromise]);

      // Clear timeout on success
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }

      setSpec(processedSpec);
      setError('');
      onRenderSuccess?.(processedSpec);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to parse Vega-Lite specification';
      console.error('Vega-Lite parsing error:', err);

      setError(errorMessage);
      setSpec(null);
      onRenderError?.(errorMessage);

      // Clear timeout on error
      if (renderTimeoutRef.current) {
        clearTimeout(renderTimeoutRef.current);
      }
    } finally {
      setLoading(false);
    }
  }, [content, sanitizeSpec, applyThemeToSpec, onRenderSuccess, onRenderError]);

  /**
   * Load Vega-Lite React component
   */
  const loadVegaLiteComponent = useCallback(async () => {
    if (componentLoadError) return;
    
    try {
      console.log('Loading vega lite component');
      
      // Try to import react-vega with a shorter timeout
      const VegaLite = await Promise.race([
        import('react-vega').then(({ VegaLite }) => {
          console.log('VegaLite component loaded successfully', VegaLite);
          return VegaLite;
        }),
        new Promise((_, reject) => 
          setTimeout(() => reject(new Error('Import timeout')), 3000)
        )
      ]);
      
      setVegaLiteComponent(() => VegaLite);
      setComponentLoadError('');
    } catch (err) {
      console.error('Failed to load VegaLite component:', err);
      const errorMsg = err instanceof Error ? err.message : 'Failed to load chart rendering component';
      setComponentLoadError(errorMsg);
      setError(errorMsg);
    }
  }, [componentLoadError]);

  /**
   * Load component and parse specification when content changes
   */
  useEffect(() => {
    console.log('VegaLite component hook', { VegaLiteComponent, componentLoadError });
    if (!VegaLiteComponent && !componentLoadError) {
      loadVegaLiteComponent();
    }
  }, [VegaLiteComponent, componentLoadError, loadVegaLiteComponent]);

  /**
   * Parse specification when component is loaded or content changes
   */
  useEffect(() => {
    if (VegaLiteComponent) {
      if (content.trim()) {
        parseSpec();
      } else {
        setLoading(false);
        setSpec(null);
        setError('');
      }
    }
  }, [VegaLiteComponent, content, parseSpec]);

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
  if (loading || (!VegaLiteComponent && !componentLoadError)) {
    return (
      <div className={`vega-lite-loading ${className}`}>
        <div className='flex items-center justify-center p-8 bg-gray-50 rounded-lg border border-gray-200'>
          <div
            className='animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600'
            role='status'
            aria-label='Loading chart'
          ></div>
          <span className='ml-2 text-sm text-gray-600'>
            {!VegaLiteComponent ? 'Loading chart component...' : 'Rendering chart...'}
          </span>
        </div>
      </div>
    );
  }

  // Component loading error state
  if (componentLoadError) {
    return (
      <div className={`vega-lite-error ${className}`}>
        <div className='p-4 bg-red-50 border border-red-200 rounded-lg'>
          <p className='text-sm text-red-800 font-medium'>Component Error:</p>
          <p className='text-sm text-red-600 mt-1'>{componentLoadError}</p>
          <button
            onClick={() => {
              setComponentLoadError('');
              setError('');
              setVegaLiteComponent(null);
              loadVegaLiteComponent();
            }}
            className='mt-2 px-3 py-1 text-xs bg-red-100 text-red-700 rounded hover:bg-red-200'
          >
            Retry Loading
          </button>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={`vega-lite-error ${className}`}>
        <div className='p-4 bg-red-50 border border-red-200 rounded-lg'>
          <p className='text-sm text-red-800 font-medium'>Chart Error:</p>
          <p className='text-sm text-red-600 mt-1'>{error}</p>
          <details className='mt-2'>
            <summary className='text-xs text-red-500 cursor-pointer'>
              Show raw JSON
            </summary>
            <pre className='text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto'>
              {content}
            </pre>
          </details>
        </div>
      </div>
    );
  }

  // Rendered chart
  if (spec && VegaLiteComponent) {
    return (
      <div className={`vega-lite-diagram ${className}`}>
        <div
          className='relative overflow-auto'
          style={{ backgroundColor: theme === 'dark' ? '#1F2937' : '#FFFFFF' }}
        >
          <div
            ref={chartRef}
            className='vega-lite-chart-container'
            role='img'
            aria-label={alt || 'Vega-Lite Chart'}
          >
            <VegaLiteComponent
              spec={spec}
              actions={showActions}
              renderer='svg'
              tooltip={enableTooltips}
              hover={enableHover}
              onError={(error: any) => {
                console.error('Vega-Lite rendering error:', error);
                const errorMessage = error?.message || 'Chart rendering failed';
                setError(errorMessage);
                onRenderError?.(errorMessage);
              }}
              style={{
                width: '100%',
                height: 'auto',
              }}
            />
            {/* Screen reader description */}
            <div className='sr-only'>
              Chart description: {alt || 'Interactive Vega-Lite visualization'}
            </div>
          </div>
        </div>
      </div>
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

/**
 * Error boundary component for Vega-Lite chart rendering
 */
export class VegaLiteDiagramErrorBoundary extends React.Component<
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
    console.error('Vega-Lite chart rendering error:', {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
    });
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className={`chart-error ${this.props.className}`}>
            <div className='p-4 bg-red-50 border border-red-200 rounded-lg'>
              <p className='text-sm text-red-800 font-medium'>Failed to render chart</p>
              <p className='text-xs text-red-600 mt-1'>{this.state.error?.message}</p>
              <details className='mt-2'>
                <summary className='text-xs text-red-500 cursor-pointer'>
                  Show raw JSON
                </summary>
                <pre className='text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto'>
                  {this.props.content}
                </pre>
              </details>
            </div>
          </div>
        )
      );
    }

    return this.props.children;
  }
}

/**
 * Safe wrapper component that handles Vega-Lite chart rendering errors gracefully.
 * Falls back to error display if chart rendering fails.
 */
export const SafeVegaLiteDiagram: React.FC<VegaLiteDiagramProps> = props => {
  return (
    <VegaLiteDiagramErrorBoundary
      content={props.content}
      className={props.className || ''}
    >
      <VegaLiteDiagram {...props} />
    </VegaLiteDiagramErrorBoundary>
  );
};

export default VegaLiteDiagram;
