import React, { useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useDiagramLoader } from '../hooks/useDiagramLoader';
import ResponsiveDiagramRenderer from './ResponsiveDiagramRenderer';
import { DiagramLoadingIndicator } from './DiagramLoadingIndicator';

/**
 * Props interface for the MarkdownRenderer component
 */
export interface MarkdownRendererProps {
  /** The markdown content to render */
  content: string;
  /** Optional CSS class name to apply to the wrapper */
  className?: string;
  /** Whether to enable diagram rendering (default: true) */
  enableDiagrams?: boolean;
  /** Theme for diagrams (default: 'auto') */
  diagramTheme?: 'light' | 'dark' | 'auto';
  /** Whether to enable zoom and pan for diagrams (default: true) */
  enableDiagramZoomPan?: boolean;
  /** Whether to show actions for Vega-Lite charts (default: false) */
  showVegaActions?: boolean;
}

/**
 * Creates custom component overrides for markdown elements with diagram support
 */
const createCustomComponents = (
  enableDiagrams: boolean,
  diagramTheme: 'light' | 'dark' | 'auto',
  enableDiagramZoomPan: boolean,
  showVegaActions: boolean,
  allRequiredLibrariesLoaded: boolean,
  isLoading: boolean
) => ({
  h1: ({ children }: any) => (
    <h1 className='text-xl sm:text-2xl md:text-3xl font-bold text-gray-900 mb-3 sm:mb-4 mt-4 sm:mt-6 first:mt-0 leading-tight'>
      {children}
    </h1>
  ),
  h2: ({ children }: any) => (
    <h2 className='text-lg sm:text-xl md:text-2xl font-semibold text-gray-800 mb-2 sm:mb-3 mt-3 sm:mt-5 first:mt-0 leading-tight'>
      {children}
    </h2>
  ),
  h3: ({ children }: any) => (
    <h3 className='text-base sm:text-lg md:text-xl font-medium text-gray-800 mb-2 mt-3 sm:mt-4 first:mt-0 leading-tight'>
      {children}
    </h3>
  ),
  ul: ({ children }: any) => (
    <ul className='list-disc list-outside mb-3 sm:mb-4 space-y-1 text-gray-700 pl-4 sm:pl-6 marker:text-gray-400'>
      {children}
    </ul>
  ),
  ol: ({ children }: any) => (
    <ol className='list-decimal list-outside mb-3 sm:mb-4 space-y-1 text-gray-700 pl-4 sm:pl-6 marker:text-gray-600 marker:font-medium'>
      {children}
    </ol>
  ),
  li: ({ children }: any) => (
    <li className='text-gray-700 leading-relaxed pl-1 text-sm sm:text-base'>
      {children}
    </li>
  ),
  // Paragraph component with responsive spacing
  p: ({ children }: any) => (
    <p className='mb-3 sm:mb-4 text-gray-700 leading-relaxed last:mb-0 text-sm sm:text-base'>
      {children}
    </p>
  ),
  // Text formatting components
  strong: ({ children }: any) => (
    <strong className='font-semibold text-gray-900'>{children}</strong>
  ),
  em: ({ children }: any) => <em className='italic text-gray-800'>{children}</em>,
  code: ({ children, className }: any) => {
    const isInline = !className;

    if (isInline) {
      return (
        <code className='bg-gray-100 text-gray-800 px-1 sm:px-1.5 py-0.5 rounded text-xs sm:text-sm font-mono border border-gray-200'>
          {children}
        </code>
      );
    }

    // Handle diagram code blocks
    if (enableDiagrams && className) {
      const language = className.replace('language-', '');
      const content = String(children).trim();

      if (language === 'mermaid') {
        // Show loading indicator if libraries are still loading
        if (isLoading) {
          return (
            <DiagramLoadingIndicator
              type='mermaid'
              message='Loading diagram libraries...'
              className='diagram-code-block mermaid-code-block mb-3 sm:mb-4'
            />
          );
        }

        // Render Mermaid diagram if libraries are loaded
        if (allRequiredLibrariesLoaded) {
          return (
            <div className='diagram-code-block mermaid-code-block mb-3 sm:mb-4'>
              <ResponsiveDiagramRenderer
                type='mermaid'
                content={content}
                theme={diagramTheme}
                enableZoomPan={enableDiagramZoomPan}
                className=''
              />
            </div>
          );
        }

        // Fallback to code block if libraries failed to load
        return (
          <div className='diagram-fallback mermaid-fallback mb-3 sm:mb-4'>
            <div className='bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-2'>
              <p className='text-sm text-yellow-800 font-medium'>Mermaid Diagram</p>
              <p className='text-xs text-yellow-600'>
                Diagram libraries are not available. Showing raw content:
              </p>
            </div>
            <pre className='bg-gray-50 border border-gray-200 rounded-lg p-2 sm:p-4 overflow-x-auto text-xs sm:text-sm font-mono leading-relaxed'>
              <code className={className}>{children}</code>
            </pre>
          </div>
        );
      }

      if (language === 'vega-lite') {
        // Show loading indicator if libraries are still loading
        if (isLoading) {
          return (
            <DiagramLoadingIndicator
              type='vega-lite'
              message='Loading chart libraries...'
              className='diagram-code-block vega-lite-code-block mb-3 sm:mb-4'
            />
          );
        }

        // Render Vega-Lite chart if libraries are loaded
        if (allRequiredLibrariesLoaded) {
          return (
            <div className='diagram-code-block vega-lite-code-block mb-3 sm:mb-4'>
              <ResponsiveDiagramRenderer
                type='vega-lite'
                content={content}
                theme={diagramTheme}
                enableZoomPan={enableDiagramZoomPan}
                showActions={showVegaActions}
                className=''
              />
            </div>
          );
        }

        // Fallback to code block if libraries failed to load
        return (
          <div className='diagram-fallback vega-lite-fallback mb-3 sm:mb-4'>
            <div className='bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-2'>
              <p className='text-sm text-yellow-800 font-medium'>Vega-Lite Chart</p>
              <p className='text-xs text-yellow-600'>
                Chart libraries are not available. Showing raw content:
              </p>
            </div>
            <pre className='bg-gray-50 border border-gray-200 rounded-lg p-2 sm:p-4 overflow-x-auto text-xs sm:text-sm font-mono leading-relaxed'>
              <code className={className}>{children}</code>
            </pre>
          </div>
        );
      }
    }

    // Default code block rendering
    return <code className={className}>{children}</code>;
  },
  pre: ({ children }: any) => {
    // Check if this is a diagram code block by examining the code element
    const childrenArray = React.Children.toArray(children);
    const codeChild = childrenArray.find(
      (child: any) =>
        typeof child === 'object' &&
        child?.props?.className &&
        (child.props.className.includes('language-mermaid') ||
          child.props.className.includes('language-vega-lite'))
    );

    // If this is a diagram code block and diagrams are enabled, render the code element directly without pre wrapper
    if (enableDiagrams && codeChild) {
      return <>{codeChild}</>;
    }

    // Default pre rendering for regular code blocks
    return (
      <pre className='bg-gray-50 border border-gray-200 rounded-lg p-2 sm:p-4 mb-3 sm:mb-4 overflow-x-auto text-xs sm:text-sm font-mono leading-relaxed'>
        {children}
      </pre>
    );
  },
  // Table components with enhanced responsive design and proper styling
  table: ({ children }: any) => (
    <div className='overflow-x-auto mb-3 sm:mb-4 rounded-lg border border-gray-200 shadow-sm'>
      <table className='min-w-full divide-y divide-gray-200'>{children}</table>
    </div>
  ),
  thead: ({ children }: any) => <thead className='bg-gray-50'>{children}</thead>,
  tbody: ({ children }: any) => (
    <tbody className='bg-white divide-y divide-gray-200'>{children}</tbody>
  ),
  tr: ({ children }: any) => (
    <tr className='hover:bg-gray-50 transition-colors duration-150'>{children}</tr>
  ),
  th: ({ children }: any) => (
    <th className='px-2 sm:px-4 py-2 sm:py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b border-gray-200'>
      {children}
    </th>
  ),
  td: ({ children }: any) => (
    <td className='px-2 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm text-gray-700 break-words'>
      {children}
    </td>
  ),
  // Special elements with consistent design patterns and responsive spacing
  blockquote: ({ children }: any) => (
    <blockquote className='border-l-4 border-blue-200 pl-3 sm:pl-4 py-2 mb-3 sm:mb-4 bg-blue-50 text-gray-700 italic rounded-r-md text-sm sm:text-base'>
      {children}
    </blockquote>
  ),
  a: ({ href, children }: any) => (
    <a
      href={href}
      target='_blank'
      rel='noopener noreferrer'
      className='text-blue-600 hover:text-blue-800 underline decoration-blue-300 hover:decoration-blue-500 transition-colors duration-150 break-words'
    >
      {children}
    </a>
  ),
  hr: () => <hr className='my-4 sm:my-6 border-t border-gray-200' />,
});

/**
 * A dedicated component for rendering markdown content with enhanced styling, GitHub Flavored Markdown support,
 * and interactive diagram rendering. This component provides consistent markdown rendering across the application
 * with proper typography, formatting, and responsive design optimizations for mobile devices.
 *
 * Features:
 * - Mermaid diagram rendering (flowcharts, sequence diagrams, etc.)
 * - Vega-Lite chart rendering (bar charts, line charts, etc.)
 * - Lazy loading of diagram libraries for optimal performance
 * - Responsive design with mobile-optimized controls
 * - Graceful fallback when diagram libraries fail to load
 */
export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({
  content,
  className = '',
  enableDiagrams = true,
  diagramTheme = 'auto',
  enableDiagramZoomPan = true,
  showVegaActions = false,
}) => {
  // Use diagram loader hook to manage library loading
  const { hasDiagrams, allRequiredLibrariesLoaded, isLoading, errors } = useDiagramLoader(
    content,
    enableDiagrams
  );

  // Create custom components with diagram support
  const customComponents = useMemo(
    () =>
      createCustomComponents(
        enableDiagrams,
        diagramTheme,
        enableDiagramZoomPan,
        showVegaActions,
        allRequiredLibrariesLoaded,
        isLoading
      ),
    [
      enableDiagrams,
      diagramTheme,
      enableDiagramZoomPan,
      showVegaActions,
      allRequiredLibrariesLoaded,
      isLoading,
    ]
  );

  return (
    <div
      className={`markdown-content text-sm sm:text-base leading-relaxed ${className} ${
        enableDiagrams && hasDiagrams ? 'has-diagrams' : ''
      }`}
    >
      {/* Show diagram loading errors if any */}
      {errors.length > 0 && (
        <div className='diagram-errors mb-4'>
          <div className='bg-red-50 border border-red-200 rounded-lg p-3'>
            <p className='text-sm text-red-800 font-medium'>Diagram Library Errors:</p>
            <ul className='text-xs text-red-600 mt-1 space-y-1'>
              {errors.map((error, index) => (
                <li key={index}>• {error}</li>
              ))}
            </ul>
          </div>
        </div>
      )}

      <ReactMarkdown remarkPlugins={[remarkGfm]} components={customComponents}>
        {content}
      </ReactMarkdown>
    </div>
  );
};

/**
 * Fallback component that renders plain text when markdown parsing fails
 */
const FallbackRenderer: React.FC<{ content: string; className?: string }> = ({
  content,
  className = '',
}) => {
  return (
    <div className={`fallback-content ${className}`}>
      <pre className='whitespace-pre-wrap text-gray-700 font-sans leading-relaxed'>
        {content}
      </pre>
    </div>
  );
};

/**
 * Error boundary component that catches rendering errors and provides fallback UI
 */
class MarkdownErrorBoundary extends React.Component<
  {
    children: React.ReactNode;
    content: string;
    className: string;
  },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { children: React.ReactNode; content: string; className: string }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log the error for debugging purposes
    console.error('Markdown rendering failed:', {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
    });

    // Log content info when fallback is used
    const truncatedContent =
      this.props.content.substring(0, 100) +
      (this.props.content.length > 100 ? '...' : '');
    console.error('Markdown rendering failed, using fallback renderer:', {
      contentLength: this.props.content.length,
      contentPreview: truncatedContent,
      timestamp: new Date().toISOString(),
    });
  }

  render() {
    if (this.state.hasError) {
      return (
        <FallbackRenderer content={this.props.content} className={this.props.className} />
      );
    }

    return this.props.children;
  }
}

/**
 * Safe wrapper component that handles markdown rendering errors gracefully.
 * Falls back to plain text rendering if markdown parsing fails.
 * Includes error logging for debugging purposes and supports all diagram features.
 */
export const SafeMarkdownRenderer: React.FC<MarkdownRendererProps> = ({
  content,
  className = '',
  enableDiagrams = true,
  diagramTheme = 'auto',
  enableDiagramZoomPan = true,
  showVegaActions = false,
}) => {
  return (
    <MarkdownErrorBoundary content={content} className={className}>
      <MarkdownRenderer
        content={content}
        className={className}
        enableDiagrams={enableDiagrams}
        diagramTheme={diagramTheme}
        enableDiagramZoomPan={enableDiagramZoomPan}
        showVegaActions={showVegaActions}
      />
    </MarkdownErrorBoundary>
  );
};

export default SafeMarkdownRenderer;
