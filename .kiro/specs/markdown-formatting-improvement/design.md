# Design Document

## Overview

The chat interface currently uses `react-markdown` to render AI coach responses, but the markdown rendering lacks proper styling and visual hierarchy. The current implementation renders markdown elements as unstyled HTML, making headings, lists, tables, and other formatted content appear as plain text. This design addresses the styling and configuration improvements needed to make markdown content visually distinct and readable.

## Architecture

### Current Implementation Analysis

The ChatInterface component already includes:
- `react-markdown` library (v9.0.1) for markdown parsing
- Basic prose styling with Tailwind CSS classes
- Proper message structure with role-based rendering

**Current Code:**
```tsx
{message.role === 'assistant' ? (
  <div className='prose prose-sm max-w-none'>
    <ReactMarkdown>{message.content}</ReactMarkdown>
  </div>
) : (
  <div className='whitespace-pre-wrap'>{message.content}</div>
)}
```

**Issues Identified:**
1. Limited Tailwind prose styling doesn't cover all markdown elements
2. No custom styling for coaching-specific content
3. Missing responsive design considerations
4. Tables may not render well on mobile devices

### Technology Stack

**Frontend Libraries:**
- React 18 with TypeScript
- `react-markdown` v9.0.1 for markdown parsing
- Tailwind CSS with `@tailwindcss/typography` plugin (needs verification/addition)
- Potential additions: `remark-gfm` for GitHub Flavored Markdown support

## Components and Interfaces

### Enhanced Markdown Renderer Component

Create a dedicated `MarkdownRenderer` component to handle all markdown styling and configuration:

```tsx
interface MarkdownRendererProps {
  content: string;
  className?: string;
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '' 
}) => {
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

### Custom Component Overrides

Define custom renderers for specific markdown elements to ensure proper styling:

```tsx
const customComponents = {
  h1: ({ children }: any) => (
    <h1 className="text-xl font-bold text-gray-900 mb-3 mt-4 first:mt-0">
      {children}
    </h1>
  ),
  h2: ({ children }: any) => (
    <h2 className="text-lg font-semibold text-gray-800 mb-2 mt-3 first:mt-0">
      {children}
    </h2>
  ),
  h3: ({ children }: any) => (
    <h3 className="text-base font-medium text-gray-800 mb-2 mt-3 first:mt-0">
      {children}
    </h3>
  ),
  ul: ({ children }: any) => (
    <ul className="list-disc list-inside mb-3 space-y-1 text-gray-700">
      {children}
    </ul>
  ),
  ol: ({ children }: any) => (
    <ol className="list-decimal list-inside mb-3 space-y-1 text-gray-700">
      {children}
    </ol>
  ),
  li: ({ children }: any) => (
    <li className="text-gray-700 leading-relaxed">{children}</li>
  ),
  p: ({ children }: any) => (
    <p className="mb-3 text-gray-700 leading-relaxed last:mb-0">
      {children}
    </p>
  ),
  strong: ({ children }: any) => (
    <strong className="font-semibold text-gray-900">{children}</strong>
  ),
  em: ({ children }: any) => (
    <em className="italic text-gray-800">{children}</em>
  ),
  code: ({ children, className }: any) => {
    const isInline = !className;
    return isInline ? (
      <code className="bg-gray-100 text-gray-800 px-1.5 py-0.5 rounded text-sm font-mono">
        {children}
      </code>
    ) : (
      <code className={className}>{children}</code>
    );
  },
  pre: ({ children }: any) => (
    <pre className="bg-gray-50 border border-gray-200 rounded-lg p-3 mb-3 overflow-x-auto">
      {children}
    </pre>
  ),
  blockquote: ({ children }: any) => (
    <blockquote className="border-l-4 border-blue-200 pl-4 py-2 mb-3 bg-blue-50 text-gray-700 italic">
      {children}
    </blockquote>
  ),
  table: ({ children }: any) => (
    <div className="overflow-x-auto mb-3">
      <table className="min-w-full border border-gray-200 rounded-lg">
        {children}
      </table>
    </div>
  ),
  thead: ({ children }: any) => (
    <thead className="bg-gray-50">{children}</thead>
  ),
  tbody: ({ children }: any) => (
    <tbody className="divide-y divide-gray-200">{children}</tbody>
  ),
  tr: ({ children }: any) => (
    <tr className="hover:bg-gray-50">{children}</tr>
  ),
  th: ({ children }: any) => (
    <th className="px-4 py-2 text-left text-sm font-medium text-gray-900 border-b border-gray-200">
      {children}
    </th>
  ),
  td: ({ children }: any) => (
    <td className="px-4 py-2 text-sm text-gray-700 border-b border-gray-200">
      {children}
    </td>
  ),
  hr: () => (
    <hr className="my-4 border-t border-gray-200" />
  ),
  a: ({ href, children }: any) => (
    <a 
      href={href} 
      target="_blank" 
      rel="noopener noreferrer"
      className="text-blue-600 hover:text-blue-800 underline"
    >
      {children}
    </a>
  ),
};
```

## Data Models

### Markdown Configuration

```tsx
interface MarkdownConfig {
  enableGfm: boolean;           // GitHub Flavored Markdown support
  enableTables: boolean;        // Table rendering
  enableTaskLists: boolean;     // Checkbox lists
  enableStrikethrough: boolean; // ~~strikethrough~~ text
  maxWidth: string;            // Maximum width for content
  mobileOptimized: boolean;    // Mobile-specific optimizations
}

const defaultConfig: MarkdownConfig = {
  enableGfm: true,
  enableTables: true,
  enableTaskLists: true,
  enableStrikethrough: true,
  maxWidth: 'none',
  mobileOptimized: true,
};
```

### Message Rendering Context

```tsx
interface MessageRenderingContext {
  role: 'user' | 'assistant';
  isStreaming: boolean;
  timestamp: string;
  sessionId: string;
}
```

## Implementation Strategy

### Phase 1: Enhanced Styling System

1. **Install Required Dependencies**
   ```bash
   npm install remark-gfm @tailwindcss/typography
   ```

2. **Update Tailwind Configuration**
   ```js
   // tailwind.config.js
   module.exports = {
     plugins: [
       require('@tailwindcss/typography'),
     ],
   }
   ```

3. **Create MarkdownRenderer Component**
   - Implement custom component overrides
   - Add responsive design considerations
   - Include accessibility features

### Phase 2: Integration with Chat Interface

1. **Replace Existing Markdown Rendering**
   ```tsx
   // Before
   <div className='prose prose-sm max-w-none'>
     <ReactMarkdown>{message.content}</ReactMarkdown>
   </div>

   // After
   <MarkdownRenderer 
     content={message.content}
     className="max-w-none"
   />
   ```

2. **Add Streaming Support**
   - Ensure partial markdown renders correctly during streaming
   - Handle incomplete markdown gracefully

### Phase 3: Mobile Optimization

1. **Responsive Table Handling**
   - Horizontal scroll for wide tables
   - Stack table data on very small screens
   - Maintain readability across devices

2. **Touch-Friendly Elements**
   - Appropriate spacing for touch targets
   - Readable font sizes on mobile
   - Proper line height for mobile reading

## Error Handling

### Markdown Parsing Errors

```tsx
const SafeMarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className 
}) => {
  try {
    return (
      <MarkdownRenderer content={content} className={className} />
    );
  } catch (error) {
    console.error('Markdown rendering failed:', error);
    return (
      <div className={`fallback-content ${className}`}>
        <pre className="whitespace-pre-wrap text-gray-700">
          {content}
        </pre>
      </div>
    );
  }
};
```

### Streaming Content Handling

- Handle incomplete markdown during streaming
- Gracefully render partial tables and lists
- Prevent layout shifts during content updates

### Malformed Markdown

- Sanitize potentially dangerous HTML
- Handle edge cases in markdown syntax
- Provide fallback rendering for unsupported elements

## Testing Strategy

### Unit Tests

```tsx
describe('MarkdownRenderer', () => {
  test('renders headings with proper hierarchy', () => {
    const content = '# H1\n## H2\n### H3';
    render(<MarkdownRenderer content={content} />);
    
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 3 })).toBeInTheDocument();
  });

  test('renders lists with proper styling', () => {
    const content = '- Item 1\n- Item 2\n\n1. Numbered 1\n2. Numbered 2';
    render(<MarkdownRenderer content={content} />);
    
    expect(screen.getByRole('list')).toBeInTheDocument();
  });

  test('renders tables responsively', () => {
    const content = '| Col 1 | Col 2 |\n|-------|-------|\n| A | B |';
    render(<MarkdownRenderer content={content} />);
    
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  test('handles malformed markdown gracefully', () => {
    const content = '# Incomplete header\n**unclosed bold';
    render(<MarkdownRenderer content={content} />);
    
    // Should not crash and should render something
    expect(screen.getByText(/Incomplete header/)).toBeInTheDocument();
  });
});
```

### Integration Tests

```tsx
describe('ChatInterface Markdown Integration', () => {
  test('renders AI responses with proper markdown formatting', async () => {
    const mockMessage = {
      id: '1',
      role: 'assistant' as const,
      content: '# Training Plan\n\n- **Week 1**: Base building\n- **Week 2**: Intensity',
      created_at: new Date().toISOString(),
      session_id: 'session-1',
    };

    render(<ChatInterface />);
    
    // Simulate receiving a message with markdown
    // Test that headings and lists render properly
  });

  test('handles streaming markdown content', async () => {
    // Test partial markdown rendering during streaming
    // Ensure no layout breaks with incomplete content
  });
});
```

### Visual Regression Tests

- Screenshot comparisons for different markdown elements
- Mobile vs desktop rendering verification
- Dark mode compatibility (if applicable)

## Performance Considerations

### Rendering Optimization

1. **Memoization**
   ```tsx
   const MarkdownRenderer = React.memo<MarkdownRendererProps>(({ 
     content, 
     className 
   }) => {
     const memoizedComponents = useMemo(() => customComponents, []);
     
     return (
       <ReactMarkdown components={memoizedComponents}>
         {content}
       </ReactMarkdown>
     );
   });
   ```

2. **Lazy Loading**
   - Consider code splitting for markdown renderer if bundle size becomes an issue

3. **Content Optimization**
   - Limit maximum content length for performance
   - Implement virtual scrolling for very long conversations

### Bundle Size Impact

- Monitor bundle size increase from additional dependencies
- Consider alternative lightweight markdown parsers if needed
- Tree-shake unused remark plugins

## Accessibility Considerations

### Semantic HTML

- Ensure proper heading hierarchy (h1 → h2 → h3)
- Use semantic list elements (ul, ol, li)
- Proper table structure with headers

### Screen Reader Support

```tsx
const accessibleComponents = {
  table: ({ children }: any) => (
    <div className="overflow-x-auto mb-3">
      <table 
        className="min-w-full border border-gray-200 rounded-lg"
        role="table"
        aria-label="Data table"
      >
        {children}
      </table>
    </div>
  ),
  // ... other components with proper ARIA labels
};
```

### Keyboard Navigation

- Ensure all interactive elements are keyboard accessible
- Proper focus management for links and buttons
- Skip links for long content sections

## Security Considerations

### Content Sanitization

```tsx
import { remark } from 'remark';
import remarkGfm from 'remark-gfm';
import remarkHtml from 'remark-html';

const sanitizeMarkdown = (content: string): string => {
  // Remove potentially dangerous HTML
  // Sanitize links and images
  // Validate markdown structure
  return content;
};
```

### XSS Prevention

- Disable HTML rendering in markdown by default
- Sanitize any user-generated content
- Use allowlist for permitted HTML elements

### Link Security

- Add `rel="noopener noreferrer"` to external links
- Validate link protocols (http/https only)
- Consider link preview warnings for external sites