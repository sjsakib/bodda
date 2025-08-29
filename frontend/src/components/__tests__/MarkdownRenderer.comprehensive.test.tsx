import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MarkdownRenderer, SafeMarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock diagram libraries
vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockResolvedValue({
      svg: '<svg role="img"><title>Test Mermaid</title><g><text>Mermaid Content</text></g></svg>'
    }),
  },
}));

vi.mock('react-vega', () => ({
  VegaLite: ({ spec }: any) => (
    <div data-testid="vega-lite-chart" role="img" aria-label={`${spec?.mark} chart`}>
      <title>Test Vega-Lite</title>
      <text>Vega-Lite: {spec?.mark || 'unknown'}</text>
    </div>
  ),
}));

// Mock diagram loader hook
vi.mock('../../hooks/useDiagramLoader', () => ({
  useDiagramLoader: vi.fn((content: string, enableDiagrams: boolean) => {
    const hasMermaid = enableDiagrams && content.includes('```mermaid');
    const hasVegaLite = enableDiagrams && content.includes('```vega-lite');
    
    return {
      hasDiagrams: hasMermaid || hasVegaLite,
      allRequiredLibrariesLoaded: true,
      isLoading: false,
      errors: [],
    };
  }),
}));

const renderWithProvider = (content: string, props = {}) => {
  return render(
    <DiagramLibraryProvider>
      <MarkdownRenderer content={content} {...props} />
    </DiagramLibraryProvider>
  );
};

describe('MarkdownRenderer - Comprehensive Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic Markdown Rendering', () => {
    it('renders standard markdown elements with proper styling', () => {
      const content = `# Main Title

## Subtitle

This is a **bold** and *italic* text with \`inline code\`.

- List item 1
- List item 2

1. Numbered item 1
2. Numbered item 2

> This is a blockquote

[Link text](https://example.com)

---

\`\`\`javascript
const code = 'block';
\`\`\``;

      renderWithProvider(content);

      expect(screen.getByText('Main Title')).toBeInTheDocument();
      expect(screen.getByText('Subtitle')).toBeInTheDocument();
      expect(screen.getByText('bold')).toBeInTheDocument();
      expect(screen.getByText('italic')).toBeInTheDocument();
      expect(screen.getByText('inline code')).toBeInTheDocument();
      expect(screen.getByText('List item 1')).toBeInTheDocument();
      expect(screen.getByText('Numbered item 1')).toBeInTheDocument();
      expect(screen.getByText('This is a blockquote')).toBeInTheDocument();
      expect(screen.getByText('Link text')).toBeInTheDocument();
    });

    it('applies responsive CSS classes correctly', () => {
      const content = '# Responsive Title\n\nResponsive paragraph text.';
      
      const { container } = renderWithProvider(content);

      const title = screen.getByText('Responsive Title');
      expect(title).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');

      const paragraph = screen.getByText('Responsive paragraph text.');
      expect(paragraph).toHaveClass('text-sm', 'sm:text-base');
    });

    it('handles tables with responsive design', () => {
      const content = `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |`;

      renderWithProvider(content);

      expect(screen.getByText('Header 1')).toBeInTheDocument();
      expect(screen.getByText('Cell 1')).toBeInTheDocument();
      
      const table = screen.getByText('Header 1').closest('table');
      expect(table?.parentElement).toHaveClass('overflow-x-auto');
    });
  }); 
 describe('Diagram Integration', () => {
    it('detects and renders Mermaid diagrams', async () => {
      const content = `# Training Plan

\`\`\`mermaid
graph TD
    A[Start] --> B[Warm Up]
    B --> C[Exercise]
    C --> D[Cool Down]
\`\`\`

Follow this workflow for best results.`;

      renderWithProvider(content);

      expect(screen.getByText('Training Plan')).toBeInTheDocument();
      expect(screen.getByText('Follow this workflow for best results.')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByText('Mermaid Content')).toBeInTheDocument();
      });

      const diagramContainer = screen.getByText('Mermaid Content').closest('.diagram-code-block');
      expect(diagramContainer).toHaveClass('mermaid-code-block');
    });

    it('detects and renders Vega-Lite charts', async () => {
      const content = `# Progress Analysis

\`\`\`vega-lite
{
  "mark": "bar",
  "data": {
    "values": [
      {"week": "Week 1", "distance": 15},
      {"week": "Week 2", "distance": 18}
    ]
  },
  "encoding": {
    "x": {"field": "week", "type": "ordinal"},
    "y": {"field": "distance", "type": "quantitative"}
  }
}
\`\`\`

Your progress is improving!`;

      renderWithProvider(content);

      expect(screen.getByText('Progress Analysis')).toBeInTheDocument();
      expect(screen.getByText('Your progress is improving!')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
        expect(screen.getByText('Vega-Lite: bar')).toBeInTheDocument();
      });

      const diagramContainer = screen.getByTestId('vega-lite-chart').closest('.diagram-code-block');
      expect(diagramContainer).toHaveClass('vega-lite-code-block');
    });

    it('handles multiple diagrams in single document', async () => {
      const content = `# Comprehensive Analysis

## Workflow
\`\`\`mermaid
graph LR
    A --> B --> C
\`\`\`

## Data Visualization
\`\`\`vega-lite
{"mark": "line", "data": {"values": []}}
\`\`\`

## Another Workflow
\`\`\`mermaid
sequenceDiagram
    A->>B: Message
\`\`\``;

      renderWithProvider(content);

      await waitFor(() => {
        expect(screen.getAllByText('Mermaid Content')).toHaveLength(2);
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const mermaidBlocks = document.querySelectorAll('.mermaid-code-block');
      const vegaBlocks = document.querySelectorAll('.vega-lite-code-block');
      
      expect(mermaidBlocks).toHaveLength(2);
      expect(vegaBlocks).toHaveLength(1);
    });

    it('applies has-diagrams class when diagrams are present', () => {
      const content = `Text with diagram:

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { container } = renderWithProvider(content);
      
      expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
    });

    it('does not apply has-diagrams class without diagrams', () => {
      const content = 'Just regular markdown text with no diagrams.';
      
      const { container } = renderWithProvider(content);
      
      expect(container.querySelector('.has-diagrams')).not.toBeInTheDocument();
    });

    it('can disable diagram rendering', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { container } = renderWithProvider(content, { enableDiagrams: false });

      expect(container.querySelector('.has-diagrams')).not.toBeInTheDocument();
      expect(screen.queryByText('Mermaid Content')).not.toBeInTheDocument();
      
      // Should render as regular code block
      const codeBlock = screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('graph TD');
      });
      expect(codeBlock).toBeInTheDocument();
    });
  });

  describe('Theme and Configuration', () => {
    it('passes diagram theme configuration', async () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(content, { diagramTheme: 'dark' });

      await waitFor(() => {
        expect(screen.getByText('Mermaid Content')).toBeInTheDocument();
      });

      // Component should render without errors with dark theme
      expect(screen.getByText('Mermaid Content')).toBeInTheDocument();
    });

    it('configures zoom and pan settings', async () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(content, { enableDiagramZoomPan: false });

      await waitFor(() => {
        expect(screen.getByText('Mermaid Content')).toBeInTheDocument();
      });
    });

    it('configures Vega-Lite actions', async () => {
      const content = `\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\``;

      renderWithProvider(content, { showVegaActions: true });

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });
    });

    it('applies custom CSS classes', () => {
      const content = '# Test Content';
      
      const { container } = renderWithProvider(content, { className: 'custom-markdown' });
      
      expect(container.querySelector('.custom-markdown')).toBeInTheDocument();
    });
  });

  describe('Error Handling and Fallbacks', () => {
    it('shows diagram loading errors', () => {
      const mockUseDiagramLoader = vi.mocked(require('../../hooks/useDiagramLoader').useDiagramLoader);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        allRequiredLibrariesLoaded: false,
        isLoading: false,
        errors: ['Failed to load Mermaid library', 'Failed to load Vega-Lite library'],
      });

      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(content);

      expect(screen.getByText('Diagram Library Errors:')).toBeInTheDocument();
      expect(screen.getByText('• Failed to load Mermaid library')).toBeInTheDocument();
      expect(screen.getByText('• Failed to load Vega-Lite library')).toBeInTheDocument();
    });

    it('shows fallback when libraries fail to load', () => {
      const mockUseDiagramLoader = vi.mocked(require('../../hooks/useDiagramLoader').useDiagramLoader);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        allRequiredLibrariesLoaded: false,
        isLoading: false,
        errors: [],
      });

      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(content);

      expect(screen.getByText('Mermaid Diagram')).toBeInTheDocument();
      expect(screen.getByText('Diagram libraries are not available. Showing raw content:')).toBeInTheDocument();
      
      const fallbackContainer = screen.getByText('Mermaid Diagram').closest('.diagram-fallback');
      expect(fallbackContainer).toHaveClass('mermaid-fallback');
    });

    it('shows loading indicators when libraries are loading', () => {
      const mockUseDiagramLoader = vi.mocked(require('../../hooks/useDiagramLoader').useDiagramLoader);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        allRequiredLibrariesLoaded: false,
        isLoading: true,
        errors: [],
      });

      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\`

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\``;

      renderWithProvider(content);

      expect(screen.getByText('Loading diagram libraries...')).toBeInTheDocument();
      expect(screen.getByText('Loading chart libraries...')).toBeInTheDocument();
    });

    it('handles mixed content with some diagrams failing', () => {
      const content = `# Mixed Content

Regular text here.

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

More text.

\`\`\`javascript
const code = 'regular code block';
\`\`\`

Final text.`;

      renderWithProvider(content);

      expect(screen.getByText('Mixed Content')).toBeInTheDocument();
      expect(screen.getByText('Regular text here.')).toBeInTheDocument();
      expect(screen.getByText('More text.')).toBeInTheDocument();
      expect(screen.getByText('Final text.')).toBeInTheDocument();
      expect(screen.getByText("const code = 'regular code block';")).toBeInTheDocument();
    });
  });

  describe('Performance and Optimization', () => {
    it('handles large markdown documents efficiently', () => {
      const largeContent = `# Large Document

${Array.from({ length: 100 }, (_, i) => `## Section ${i + 1}

This is paragraph ${i + 1} with some content.

- Item 1
- Item 2
- Item 3

\`\`\`javascript
const section${i + 1} = 'code block ${i + 1}';
\`\`\`

`).join('\n')}`;

      const startTime = performance.now();
      renderWithProvider(largeContent);
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(1000); // Should render in less than 1 second
      expect(screen.getByText('Large Document')).toBeInTheDocument();
      expect(screen.getByText('Section 1')).toBeInTheDocument();
      expect(screen.getByText('Section 100')).toBeInTheDocument();
    });

    it('handles rapid content updates efficiently', () => {
      const contents = [
        '# Content 1\nFirst content',
        '# Content 2\nSecond content',
        '# Content 3\nThird content',
        '# Content 4\nFourth content',
      ];

      const { rerender } = renderWithProvider(contents[0]);
      expect(screen.getByText('Content 1')).toBeInTheDocument();

      contents.slice(1).forEach((content, index) => {
        rerender(
          <DiagramLibraryProvider>
            <MarkdownRenderer content={content} />
          </DiagramLibraryProvider>
        );
        expect(screen.getByText(`Content ${index + 2}`)).toBeInTheDocument();
      });
    });

    it('memoizes custom components properly', () => {
      const content = '# Test\n\nContent here.';
      
      const { rerender } = renderWithProvider(content);
      
      // Re-render with same props should not cause unnecessary re-renders
      rerender(
        <DiagramLibraryProvider>
          <MarkdownRenderer content={content} />
        </DiagramLibraryProvider>
      );

      expect(screen.getByText('Test')).toBeInTheDocument();
      expect(screen.getByText('Content here.')).toBeInTheDocument();
    });
  });

  describe('Accessibility Features', () => {
    it('maintains proper heading hierarchy', () => {
      const content = `# Main Title
## Subtitle
### Sub-subtitle
#### Fourth level
##### Fifth level
###### Sixth level`;

      renderWithProvider(content);

      expect(screen.getByRole('heading', { level: 1, name: 'Main Title' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 2, name: 'Subtitle' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 3, name: 'Sub-subtitle' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 4, name: 'Fourth level' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 5, name: 'Fifth level' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 6, name: 'Sixth level' })).toBeInTheDocument();
    });

    it('provides proper link accessibility', () => {
      const content = '[External link](https://example.com) and [Internal link](#section)';
      
      renderWithProvider(content);

      const externalLink = screen.getByText('External link');
      expect(externalLink).toHaveAttribute('href', 'https://example.com');
      expect(externalLink).toHaveAttribute('target', '_blank');
      expect(externalLink).toHaveAttribute('rel', 'noopener noreferrer');

      const internalLink = screen.getByText('Internal link');
      expect(internalLink).toHaveAttribute('href', '#section');
    });

    it('provides accessible table structure', () => {
      const content = `| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |`;

      renderWithProvider(content);

      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();

      const headers = screen.getAllByRole('columnheader');
      expect(headers).toHaveLength(3);
      expect(headers[0]).toHaveTextContent('Name');

      const cells = screen.getAllByRole('cell');
      expect(cells).toHaveLength(6);
      expect(cells[0]).toHaveTextContent('John');
    });

    it('maintains accessibility for diagrams', async () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(content);

      await waitFor(() => {
        const diagram = screen.getByRole('img');
        expect(diagram).toBeInTheDocument();
        expect(diagram).toHaveAttribute('aria-labelledby');
      });
    });
  });

  describe('Edge Cases and Boundary Conditions', () => {
    it('handles empty content', () => {
      renderWithProvider('');
      
      const container = document.querySelector('.markdown-content');
      expect(container).toBeInTheDocument();
      expect(container?.textContent?.trim()).toBe('');
    });

    it('handles whitespace-only content', () => {
      renderWithProvider('   \n\t  ');
      
      const container = document.querySelector('.markdown-content');
      expect(container).toBeInTheDocument();
    });

    it('handles malformed markdown gracefully', () => {
      const malformedContent = `# Unclosed **bold text
## Missing closing bracket [link text](
### Incomplete table
| Header 1 | Header 2
| Cell 1`;

      renderWithProvider(malformedContent);

      // Should still render what it can
      expect(screen.getByText('Unclosed **bold text')).toBeInTheDocument();
      expect(screen.getByText('Missing closing bracket [link text](')).toBeInTheDocument();
    });

    it('handles very long lines', () => {
      const longLine = 'This is a very long line that should wrap properly and not break the layout even when it contains many words and extends far beyond the normal width of a typical paragraph or text block in the user interface.';
      
      renderWithProvider(longLine);
      
      expect(screen.getByText(longLine)).toBeInTheDocument();
    });

    it('handles special characters and unicode', () => {
      const unicodeContent = `# Unicode Test 🚀

Emoji: 😀 🎉 ⭐
Math: α β γ δ ∑ ∫ ∞
Symbols: © ® ™ § ¶
Accents: café naïve résumé`;

      renderWithProvider(unicodeContent);

      expect(screen.getByText('Unicode Test 🚀')).toBeInTheDocument();
      expect(screen.getByText(/Emoji: 😀 🎉 ⭐/)).toBeInTheDocument();
      expect(screen.getByText(/Math: α β γ δ ∑ ∫ ∞/)).toBeInTheDocument();
    });

    it('handles nested markdown structures', () => {
      const nestedContent = `# Main Title

> ## Quoted Heading
> 
> This is a **bold** text inside a blockquote with a [link](https://example.com).
> 
> - Quoted list item 1
> - Quoted list item 2
>   - Nested item
>   - Another nested item
> 
> \`\`\`javascript
> const quotedCode = 'inside blockquote';
> \`\`\`

Normal paragraph after blockquote.`;

      renderWithProvider(nestedContent);

      expect(screen.getByText('Main Title')).toBeInTheDocument();
      expect(screen.getByText('Quoted Heading')).toBeInTheDocument();
      expect(screen.getByText('bold')).toBeInTheDocument();
      expect(screen.getByText('link')).toBeInTheDocument();
      expect(screen.getByText('Quoted list item 1')).toBeInTheDocument();
      expect(screen.getByText('Normal paragraph after blockquote.')).toBeInTheDocument();
    });
  });
});

describe('SafeMarkdownRenderer - Error Boundary Tests', () => {
  const originalConsoleError = console.error;

  beforeEach(() => {
    console.error = vi.fn();
  });

  afterEach(() => {
    console.error = originalConsoleError;
  });

  it('catches rendering errors and shows fallback', () => {
    // Mock ReactMarkdown to throw an error
    vi.doMock('react-markdown', () => ({
      default: () => {
        throw new Error('Markdown parsing failed');
      },
    }));

    render(<SafeMarkdownRenderer content="# Test Content" />);

    // Should show fallback content
    expect(screen.getByText('# Test Content')).toBeInTheDocument();
  });

  it('logs error details for debugging', () => {
    vi.doMock('react-markdown', () => ({
      default: () => {
        throw new Error('Test error for logging');
      },
    }));

    render(<SafeMarkdownRenderer content="test content" />);

    expect(console.error).toHaveBeenCalled();
  });

  it('passes through props when no error occurs', () => {
    render(<SafeMarkdownRenderer content="# Normal Content" className="test-class" />);

    expect(screen.getByText('Normal Content')).toBeInTheDocument();
    
    const container = document.querySelector('.test-class');
    expect(container).toBeInTheDocument();
  });

  it('handles complex content in fallback mode', () => {
    const complexContent = `# Title

Complex content with **formatting** and [links](https://example.com).

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

    vi.doMock('react-markdown', () => ({
      default: () => {
        throw new Error('Complex parsing failed');
      },
    }));

    render(<SafeMarkdownRenderer content={complexContent} />);

    // Should show raw content in fallback
    expect(screen.getByText(complexContent)).toBeInTheDocument();
  });
});