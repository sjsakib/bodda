# Design Document

## Overview

This design extends the existing markdown rendering system to support interactive diagrams through Mermaid and Vega-Lite libraries. The solution builds upon the current `MarkdownRenderer` component and implements lazy loading to ensure optimal performance. The design focuses on seamless integration with the existing chat interface while providing rich visual capabilities for AI coaching responses.

## Architecture

### Current System Analysis

The application currently uses:
- `react-markdown` v9.0.1 for markdown parsing
- Custom `MarkdownRenderer` component with styled components
- Tailwind CSS for consistent styling
- TypeScript for type safety

**Integration Points:**
- `ChatInterface` component renders AI responses
- `MarkdownRenderer` handles all markdown content
- Custom component overrides provide consistent styling

### Technology Stack

**Core Libraries:**
- `mermaid` v10.x - For flowcharts, sequence diagrams, and other diagram types
- `vega-lite` v5.x - For data visualizations and charts
- `react-vega` v7.x - React wrapper for Vega-Lite integration

**Supporting Libraries:**
- `react-markdown` (existing) - Markdown parsing
- `remark-gfm` (existing) - GitHub Flavored Markdown support

### Lazy Loading Strategy

```tsx
// Dynamic imports for diagram libraries
const loadMermaid = () => import('mermaid');
const loadVegaLite = () => Promise.all([
  import('vega-lite'),
  import('react-vega')
]);

// Detection patterns
const MERMAID_PATTERN = /```mermaid\s*\n([\s\S]*?)\n```/g;
const VEGA_PATTERN = /```vega-lite\s*\n([\s\S]*?)\n```/g;
```

## Components and Interfaces

### Enhanced Markdown Renderer

Extend the existing `MarkdownRenderer` to detect and handle diagram content:

```tsx
interface DiagramRendererProps {
  content: string;
  type: 'mermaid' | 'vega-lite';
  className?: string;
}

interface MarkdownRendererProps {
  content: string;
  className?: string;
  enableDiagrams?: boolean; // Default: true
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '',
  enableDiagrams = true 
}) => {
  const [diagramLibsLoaded, setDiagramLibsLoaded] = useState(false);
  const hasDiagrams = useMemo(() => 
    enableDiagrams && (
      MERMAID_PATTERN.test(content) || 
      VEGA_PATTERN.test(content)
    ), [content, enableDiagrams]);

  // Lazy load diagram libraries when needed
  useEffect(() => {
    if (hasDiagrams && !diagramLibsLoaded) {
      loadDiagramLibraries().then(() => setDiagramLibsLoaded(true));
    }
  }, [hasDiagrams, diagramLibsLoaded]);

  const customComponents = useMemo(() => ({
    ...existingComponents,
    code: ({ children, className }: any) => {
      const language = className?.replace('language-', '');
      
      if (language === 'mermaid' && diagramLibsLoaded) {
        return <MermaidDiagram content={children} />;
      }
      
      if (language === 'vega-lite' && diagramLibsLoaded) {
        return <VegaLiteDiagram content={children} />;
      }
      
      // Fallback to existing code rendering
      return <CodeBlock className={className}>{children}</CodeBlock>;
    }
  }), [diagramLibsLoaded]);

  return (
    <div className={`markdown-content ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={customComponents}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
```

### Mermaid Diagram Component

```tsx
interface MermaidDiagramProps {
  content: string;
  className?: string;
  theme?: 'light' | 'dark';
}

const MermaidDiagram: React.FC<MermaidDiagramProps> = ({ 
  content, 
  className = '',
  theme = 'light' 
}) => {
  const [svg, setSvg] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const diagramRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const renderDiagram = async () => {
      try {
        setLoading(true);
        setError('');
        
        const mermaid = await import('mermaid');
        
        // Configure mermaid with theme
        mermaid.default.initialize({
          theme: theme === 'dark' ? 'dark' : 'default',
          themeVariables: {
            primaryColor: theme === 'dark' ? '#3B82F6' : '#1E40AF',
            primaryTextColor: theme === 'dark' ? '#F3F4F6' : '#1F2937',
            primaryBorderColor: theme === 'dark' ? '#6B7280' : '#D1D5DB',
            lineColor: theme === 'dark' ? '#6B7280' : '#374151',
          },
          fontFamily: 'ui-sans-serif, system-ui, sans-serif',
          fontSize: 14,
          securityLevel: 'strict',
        });

        const { svg } = await mermaid.default.render(
          `mermaid-${Date.now()}`, 
          content.trim()
        );
        
        setSvg(svg);
      } catch (err) {
        console.error('Mermaid rendering error:', err);
        setError(err instanceof Error ? err.message : 'Failed to render diagram');
      } finally {
        setLoading(false);
      }
    };

    renderDiagram();
  }, [content, theme]);

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

  if (error) {
    return (
      <div className={`mermaid-error ${className}`}>
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-sm text-red-800 font-medium">Diagram Error:</p>
          <p className="text-sm text-red-600 mt-1">{error}</p>
          <details className="mt-2">
            <summary className="text-xs text-red-500 cursor-pointer">Show raw content</summary>
            <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap">{content}</pre>
          </details>
        </div>
      </div>
    );
  }

  return (
    <div 
      ref={diagramRef}
      className={`mermaid-diagram ${className}`}
    >
      <div className="relative bg-white border border-gray-200 rounded-lg p-4 overflow-auto">
        <div 
          dangerouslySetInnerHTML={{ __html: svg }}
          className="mermaid-svg-container"
        />
      </div>
    </div>
  );
};
```

### Vega-Lite Chart Component

```tsx
interface VegaLiteDiagramProps {
  content: string;
  className?: string;
  theme?: 'light' | 'dark';
}

const VegaLiteDiagram: React.FC<VegaLiteDiagramProps> = ({ 
  content, 
  className = '',
  theme = 'light' 
}) => {
  const [spec, setSpec] = useState<any>(null);
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const parseSpec = async () => {
      try {
        setLoading(true);
        setError('');
        
        const parsedSpec = JSON.parse(content.trim());
        
        // Apply theme-specific styling
        const themedSpec = {
          ...parsedSpec,
          config: {
            ...parsedSpec.config,
            background: theme === 'dark' ? '#1F2937' : '#FFFFFF',
            title: {
              color: theme === 'dark' ? '#F3F4F6' : '#1F2937',
              ...parsedSpec.config?.title,
            },
            axis: {
              labelColor: theme === 'dark' ? '#D1D5DB' : '#374151',
              titleColor: theme === 'dark' ? '#F3F4F6' : '#1F2937',
              gridColor: theme === 'dark' ? '#374151' : '#E5E7EB',
              domainColor: theme === 'dark' ? '#6B7280' : '#9CA3AF',
              ...parsedSpec.config?.axis,
            },
            legend: {
              labelColor: theme === 'dark' ? '#D1D5DB' : '#374151',
              titleColor: theme === 'dark' ? '#F3F4F6' : '#1F2937',
              ...parsedSpec.config?.legend,
            },
          },
        };
        
        setSpec(themedSpec);
      } catch (err) {
        console.error('Vega-Lite parsing error:', err);
        setError(err instanceof Error ? err.message : 'Failed to parse chart specification');
      } finally {
        setLoading(false);
      }
    };

    parseSpec();
  }, [content, theme]);

  if (loading) {
    return (
      <div className={`vega-loading ${className}`}>
        <div className="flex items-center justify-center p-8 bg-gray-50 rounded-lg border border-gray-200">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-sm text-gray-600">Rendering chart...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`vega-error ${className}`}>
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-sm text-red-800 font-medium">Chart Error:</p>
          <p className="text-sm text-red-600 mt-1">{error}</p>
          <details className="mt-2">
            <summary className="text-xs text-red-500 cursor-pointer">Show raw JSON</summary>
            <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap">{content}</pre>
          </details>
        </div>
      </div>
    );
  }

  return (
    <div className={`vega-diagram ${className}`}>
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <VegaLite 
          spec={spec} 
          actions={false}
          renderer="svg"
          className="vega-chart"
        />
      </div>
    </div>
  );
};
```

## Data Models

### Diagram Configuration

```tsx
interface DiagramConfig {
  mermaid: {
    theme: 'light' | 'dark' | 'neutral';
    fontFamily: string;
    fontSize: number;
    securityLevel: 'strict' | 'loose';
    maxTextSize: number;
  };
  vegaLite: {
    theme: 'light' | 'dark';
    renderer: 'canvas' | 'svg';
    actions: boolean;
    tooltip: boolean;
    maxDataPoints: number;
  };
  performance: {
    lazyLoad: boolean;
    maxDiagramsPerMessage: number;
    renderTimeout: number;
  };
}

const defaultDiagramConfig: DiagramConfig = {
  mermaid: {
    theme: 'light',
    fontFamily: 'ui-sans-serif, system-ui, sans-serif',
    fontSize: 14,
    securityLevel: 'strict',
    maxTextSize: 50000,
  },
  vegaLite: {
    theme: 'light',
    renderer: 'svg',
    actions: false,
    tooltip: true,
    maxDataPoints: 5000,
  },
  performance: {
    lazyLoad: true,
    maxDiagramsPerMessage: 10,
    renderTimeout: 10000,
  },
};
```

### Library Loading State

```tsx
interface DiagramLibraryState {
  mermaid: {
    loaded: boolean;
    loading: boolean;
    error: string | null;
  };
  vegaLite: {
    loaded: boolean;
    loading: boolean;
    error: string | null;
  };
}
```

## Error Handling

### Library Loading Errors

```tsx
const DiagramLibraryProvider: React.FC<{ children: React.ReactNode }> = ({ 
  children 
}) => {
  const [libraryState, setLibraryState] = useState<DiagramLibraryState>({
    mermaid: { loaded: false, loading: false, error: null },
    vegaLite: { loaded: false, loading: false, error: null },
  });

  const loadMermaid = useCallback(async () => {
    if (libraryState.mermaid.loaded || libraryState.mermaid.loading) return;
    
    setLibraryState(prev => ({
      ...prev,
      mermaid: { ...prev.mermaid, loading: true, error: null }
    }));

    try {
      await import('mermaid');
      setLibraryState(prev => ({
        ...prev,
        mermaid: { loaded: true, loading: false, error: null }
      }));
    } catch (error) {
      setLibraryState(prev => ({
        ...prev,
        mermaid: { 
          loaded: false, 
          loading: false, 
          error: 'Failed to load Mermaid library' 
        }
      }));
    }
  }, [libraryState.mermaid]);

  // Similar implementation for Vega-Lite...

  return (
    <DiagramLibraryContext.Provider value={{ libraryState, loadMermaid, loadVegaLite }}>
      {children}
    </DiagramLibraryContext.Provider>
  );
};
```

### Diagram Rendering Errors

```tsx
const withDiagramErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>
) => {
  return class DiagramErrorBoundary extends React.Component<
    P & { fallback?: React.ReactNode },
    { hasError: boolean; error?: Error }
  > {
    constructor(props: P & { fallback?: React.ReactNode }) {
      super(props);
      this.state = { hasError: false };
    }

    static getDerivedStateFromError(error: Error) {
      return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
      console.error('Diagram rendering error:', {
        error: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
      });
    }

    render() {
      if (this.state.hasError) {
        return this.props.fallback || (
          <div className="diagram-error p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800">Failed to render diagram</p>
            <p className="text-xs text-red-600 mt-1">{this.state.error?.message}</p>
          </div>
        );
      }

      return <Component {...this.props} />;
    }
  };
};
```

## Testing Strategy

### Unit Tests

```tsx
describe('DiagramRenderer', () => {
  test('detects mermaid diagrams in markdown', () => {
    const content = '```mermaid\ngraph TD\nA-->B\n```';
    const { container } = render(<MarkdownRenderer content={content} />);
    
    expect(container.querySelector('.mermaid-diagram')).toBeInTheDocument();
  });

  test('detects vega-lite charts in markdown', () => {
    const content = '```vega-lite\n{"mark": "bar", "data": {"values": []}}\n```';
    const { container } = render(<MarkdownRenderer content={content} />);
    
    expect(container.querySelector('.vega-diagram')).toBeInTheDocument();
  });

  test('handles invalid mermaid syntax gracefully', () => {
    const content = '```mermaid\ninvalid syntax here\n```';
    const { container } = render(<MarkdownRenderer content={content} />);
    
    expect(container.querySelector('.mermaid-error')).toBeInTheDocument();
  });

  test('handles invalid vega-lite JSON gracefully', () => {
    const content = '```vega-lite\n{invalid json}\n```';
    const { container } = render(<MarkdownRenderer content={content} />);
    
    expect(container.querySelector('.vega-error')).toBeInTheDocument();
  });

  test('lazy loads libraries only when diagrams are present', async () => {
    const loadMermaidSpy = jest.fn();
    const content = 'Regular markdown without diagrams';
    
    render(<MarkdownRenderer content={content} />);
    
    expect(loadMermaidSpy).not.toHaveBeenCalled();
  });
});
```

### Integration Tests

```tsx
describe('ChatInterface Diagram Integration', () => {
  test('renders AI responses with mermaid diagrams', async () => {
    const mockMessage = {
      id: '1',
      role: 'assistant' as const,
      content: 'Here is your training plan:\n\n```mermaid\ngraph TD\nA[Start] --> B[Warm up]\nB --> C[Main workout]\n```',
      created_at: new Date().toISOString(),
      session_id: 'session-1',
    };

    render(<ChatInterface />);
    
    // Test that mermaid diagram renders properly
    await waitFor(() => {
      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();
    });
  });

  test('handles streaming content with partial diagrams', async () => {
    // Test that partial diagram content doesn't break rendering
    const partialContent = 'Here is a chart:\n\n```vega-lite\n{"mark": "bar"';
    
    render(<MarkdownRenderer content={partialContent} />);
    
    // Should not attempt to render incomplete diagram
    expect(screen.queryByText('Rendering chart...')).not.toBeInTheDocument();
  });
});
```

### Performance Tests

```tsx
describe('Diagram Performance', () => {
  test('does not load libraries when no diagrams present', () => {
    const content = 'Regular markdown content without any diagrams';
    
    render(<MarkdownRenderer content={content} />);
    
    // Verify no dynamic imports were triggered
    expect(mockDynamicImport).not.toHaveBeenCalled();
  });

  test('loads libraries only once per session', async () => {
    const content1 = '```mermaid\ngraph TD\nA-->B\n```';
    const content2 = '```mermaid\ngraph LR\nX-->Y\n```';
    
    const { rerender } = render(<MarkdownRenderer content={content1} />);
    await waitFor(() => expect(mockMermaidImport).toHaveBeenCalledTimes(1));
    
    rerender(<MarkdownRenderer content={content2} />);
    
    // Should not load again
    expect(mockMermaidImport).toHaveBeenCalledTimes(1);
  });
});
```

## Performance Considerations

### Bundle Size Optimization

1. **Dynamic Imports**
   ```tsx
   // Lazy load only when needed
   const mermaid = await import('mermaid');
   const { VegaLite } = await import('react-vega');
   ```

2. **Code Splitting**
   ```tsx
   // Separate chunk for diagram components
   const DiagramRenderer = lazy(() => import('./DiagramRenderer'));
   ```

3. **Library Size Impact**
   - Mermaid: ~800KB gzipped
   - Vega-Lite + React-Vega: ~600KB gzipped
   - Total impact: ~1.4MB only when diagrams are used

### Rendering Performance

1. **Memoization**
   ```tsx
   const MermaidDiagram = React.memo<MermaidDiagramProps>(({ content, theme }) => {
     const memoizedSvg = useMemo(() => renderMermaid(content, theme), [content, theme]);
     return <div dangerouslySetInnerHTML={{ __html: memoizedSvg }} />;
   });
   ```

2. **Virtualization**
   - Consider virtual scrolling for messages with many diagrams
   - Render diagrams only when visible in viewport

3. **Caching**
   ```tsx
   const diagramCache = new Map<string, string>();
   
   const getCachedDiagram = (content: string, type: string) => {
     const key = `${type}:${hashContent(content)}`;
     return diagramCache.get(key);
   };
   ```

## Security Considerations

### Content Sanitization

```tsx
const sanitizeMermaidContent = (content: string): string => {
  // Remove potentially dangerous directives
  const sanitized = content
    .replace(/%%\{.*?\}%%/g, '') // Remove config directives
    .replace(/click\s+\w+\s+href/gi, '') // Remove click handlers
    .replace(/javascript:/gi, ''); // Remove javascript: URLs
  
  return sanitized;
};

const sanitizeVegaLiteSpec = (spec: any): any => {
  // Remove potentially dangerous properties
  const sanitized = { ...spec };
  delete sanitized.datasets; // Remove external data references
  delete sanitized.transform; // Remove transform functions
  
  return sanitized;
};
```

### XSS Prevention

```tsx
// Configure mermaid with strict security
mermaid.initialize({
  securityLevel: 'strict',
  htmlLabels: false,
  maxTextSize: 50000,
});

// Validate Vega-Lite specs
const isValidVegaSpec = (spec: any): boolean => {
  try {
    // Basic validation
    return typeof spec === 'object' && 
           spec.mark && 
           !spec.datasets && 
           !spec.transform;
  } catch {
    return false;
  }
};
```

## Accessibility Considerations

### Screen Reader Support

```tsx
const MermaidDiagram: React.FC<MermaidDiagramProps> = ({ content, alt }) => {
  return (
    <div 
      role="img" 
      aria-label={alt || 'Mermaid diagram'}
      className="mermaid-diagram"
    >
      <div dangerouslySetInnerHTML={{ __html: svg }} />
      <div className="sr-only">
        Diagram description: {generateDiagramDescription(content)}
      </div>
    </div>
  );
};
```

### Keyboard Navigation

```tsx
const InteractiveDiagram: React.FC = ({ children }) => {
  return (
    <div 
      tabIndex={0}
      role="application"
      aria-label="Interactive diagram"
      onKeyDown={handleKeyboardNavigation}
    >
      {children}
    </div>
  );
};
```

### High Contrast Support

```tsx
const getAccessibleTheme = (userPreferences: UserPreferences) => {
  if (userPreferences.highContrast) {
    return {
      primaryColor: '#000000',
      primaryTextColor: '#000000',
      primaryBorderColor: '#000000',
      lineColor: '#000000',
      backgroundColor: '#FFFFFF',
    };
  }
  
  return defaultTheme;
};
```