import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

/**
 * Props interface for the MarkdownRenderer component
 */
export interface MarkdownRendererProps {
  /** The markdown content to render */
  content: string;
  /** Optional CSS class name to apply to the wrapper */
  className?: string;
}

/**
 * Custom component overrides for markdown elements with proper Tailwind styling
 */
const customComponents = {
  h1: ({ children }: any) => (
    <h1 className="text-xl sm:text-2xl md:text-3xl font-bold text-gray-900 mb-3 sm:mb-4 mt-4 sm:mt-6 first:mt-0 leading-tight">
      {children}
    </h1>
  ),
  h2: ({ children }: any) => (
    <h2 className="text-lg sm:text-xl md:text-2xl font-semibold text-gray-800 mb-2 sm:mb-3 mt-3 sm:mt-5 first:mt-0 leading-tight">
      {children}
    </h2>
  ),
  h3: ({ children }: any) => (
    <h3 className="text-base sm:text-lg md:text-xl font-medium text-gray-800 mb-2 mt-3 sm:mt-4 first:mt-0 leading-tight">
      {children}
    </h3>
  ),
  ul: ({ children }: any) => (
    <ul className="list-disc list-outside mb-3 sm:mb-4 space-y-1 text-gray-700 pl-4 sm:pl-6 marker:text-gray-400">
      {children}
    </ul>
  ),
  ol: ({ children }: any) => (
    <ol className="list-decimal list-outside mb-3 sm:mb-4 space-y-1 text-gray-700 pl-4 sm:pl-6 marker:text-gray-600 marker:font-medium">
      {children}
    </ol>
  ),
  li: ({ children }: any) => (
    <li className="text-gray-700 leading-relaxed pl-1 text-sm sm:text-base">
      {children}
    </li>
  ),
  // Paragraph component with responsive spacing
  p: ({ children }: any) => (
    <p className="mb-3 sm:mb-4 text-gray-700 leading-relaxed last:mb-0 text-sm sm:text-base">
      {children}
    </p>
  ),
  // Text formatting components
  strong: ({ children }: any) => (
    <strong className="font-semibold text-gray-900">
      {children}
    </strong>
  ),
  em: ({ children }: any) => (
    <em className="italic text-gray-800">
      {children}
    </em>
  ),
  code: ({ children, className }: any) => {
    const isInline = !className;
    return isInline ? (
      <code className="bg-gray-100 text-gray-800 px-1 sm:px-1.5 py-0.5 rounded text-xs sm:text-sm font-mono border border-gray-200">
        {children}
      </code>
    ) : (
      <code className={className}>
        {children}
      </code>
    );
  },
  pre: ({ children }: any) => (
    <pre className="bg-gray-50 border border-gray-200 rounded-lg p-2 sm:p-4 mb-3 sm:mb-4 overflow-x-auto text-xs sm:text-sm font-mono leading-relaxed">
      {children}
    </pre>
  ),
  // Table components with enhanced responsive design and proper styling
  table: ({ children }: any) => (
    <div className="overflow-x-auto mb-3 sm:mb-4 rounded-lg border border-gray-200 shadow-sm">
      <table className="min-w-full divide-y divide-gray-200">
        {children}
      </table>
    </div>
  ),
  thead: ({ children }: any) => (
    <thead className="bg-gray-50">
      {children}
    </thead>
  ),
  tbody: ({ children }: any) => (
    <tbody className="bg-white divide-y divide-gray-200">
      {children}
    </tbody>
  ),
  tr: ({ children }: any) => (
    <tr className="hover:bg-gray-50 transition-colors duration-150">
      {children}
    </tr>
  ),
  th: ({ children }: any) => (
    <th className="px-2 sm:px-4 py-2 sm:py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b border-gray-200">
      {children}
    </th>
  ),
  td: ({ children }: any) => (
    <td className="px-2 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm text-gray-700 break-words">
      {children}
    </td>
  ),
  // Special elements with consistent design patterns and responsive spacing
  blockquote: ({ children }: any) => (
    <blockquote className="border-l-4 border-blue-200 pl-3 sm:pl-4 py-2 mb-3 sm:mb-4 bg-blue-50 text-gray-700 italic rounded-r-md text-sm sm:text-base">
      {children}
    </blockquote>
  ),
  a: ({ href, children }: any) => (
    <a 
      href={href} 
      target="_blank" 
      rel="noopener noreferrer"
      className="text-blue-600 hover:text-blue-800 underline decoration-blue-300 hover:decoration-blue-500 transition-colors duration-150 break-words"
    >
      {children}
    </a>
  ),
  hr: () => (
    <hr className="my-4 sm:my-6 border-t border-gray-200" />
  ),
};

/**
 * A dedicated component for rendering markdown content with enhanced styling and GitHub Flavored Markdown support.
 * This component provides consistent markdown rendering across the application with proper typography and formatting.
 * Includes responsive design optimizations for mobile devices and touch-friendly spacing.
 */
export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '' 
}) => {
  return (
    <div className={`markdown-content text-sm sm:text-base leading-relaxed ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={customComponents}
      >
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
  className = '' 
}) => {
  return (
    <div className={`fallback-content ${className}`}>
      <pre className="whitespace-pre-wrap text-gray-700 font-sans leading-relaxed">
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
      timestamp: new Date().toISOString()
    });

    // Log content info when fallback is used
    const truncatedContent = this.props.content.substring(0, 100) + (this.props.content.length > 100 ? '...' : '');
    console.error('Markdown rendering failed, using fallback renderer:', {
      contentLength: this.props.content.length,
      contentPreview: truncatedContent,
      timestamp: new Date().toISOString()
    });
  }

  render() {
    if (this.state.hasError) {
      return <FallbackRenderer content={this.props.content} className={this.props.className} />;
    }

    return this.props.children;
  }
}

/**
 * Safe wrapper component that handles markdown rendering errors gracefully.
 * Falls back to plain text rendering if markdown parsing fails.
 * Includes error logging for debugging purposes.
 */
export const SafeMarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '' 
}) => {
  return (
    <MarkdownErrorBoundary content={content} className={className}>
      <MarkdownRenderer content={content} className={className} />
    </MarkdownErrorBoundary>
  );
};

export default SafeMarkdownRenderer;