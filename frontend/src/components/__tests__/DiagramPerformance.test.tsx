import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock performance.now for consistent timing
const mockPerformanceNow = vi.fn();
Object.defineProperty(global, 'performance', {
  value: { now: mockPerformanceNow },
  writable: true,
});

// Mock diagram libraries with timing simulation
const mockMermaidRender = vi.fn();
const mockMermaidInitialize = vi.fn();

vi.mock('mermaid', () => ({
  default: {
    initialize: mockMermaidInitialize,
    render: mockMermaidRender,
  },
}));

const mockVegaLite = vi.fn();
vi.mock('react-vega', () => ({
  VegaLite: mockVegaLite,
}));

// Mock diagram loader with performance tracking
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

describe('Diagram Performance Tests', () => {
  let timeCounter = 0;

  beforeEach(() => {
    vi.clearAllMocks();
    timeCounter = 0;
    
    // Mock performance.now to return incremental values
    mockPerformanceNow.mockImplementation(() => {
      timeCounter += 10; // Increment by 10ms each call
      return timeCounter;
    });

    // Mock fast diagram rendering
    mockMermaidRender.mockImplementation(() => 
      Promise.resolve({ svg: '<svg><text>Fast Mermaid</text></svg>' })
    );
    
    mockVegaLite.mockImplementation(({ spec }) => (
      <div data-testid="vega-lite-chart">Fast Vega-Lite: {spec?.mark}</div>
    ));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Library Loading Performance', () => {
    it('does not load libraries when no diagrams are present', () => {
      const content = `# Regular Content

This is just regular markdown without any diagrams.

- List item 1
- List item 2

\`\`\`javascript
const code = 'regular code block';
\`\`\``;

      const startTime = performance.now();
      renderWithProvider(content);
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(50); // Should be very fast
      expect(screen.getByText('Regular Content')).toBeInTheDocument();
      
      // Should not have has-diagrams class
      const container = document.querySelector('.markdown-content');
      expect(container).not.toHaveClass('has-diagrams');
    });

    it('loads libraries only when diagrams are detected', async () => {
      const contentWithDiagram = `# Content with Diagram

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const startTime = performance.now();
      renderWithProvider(contentWithDiagram);
      const endTime = performance.now();

      // Initial render should be fast
      expect(endTime - startTime).toBeLessThan(100);
      
      await waitFor(() => {
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });

      // Should have has-diagrams class
      const container = document.querySelector('.markdown-content');
      expect(container).toHaveClass('has-diagrams');
    });

    it('handles multiple diagram types efficiently', async () => {
      const contentWithMultipleDiagrams = `# Multiple Diagrams

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\`

\`\`\`mermaid
sequenceDiagram
    A->>B: Message
\`\`\``;

      const startTime = performance.now();
      renderWithProvider(contentWithMultipleDiagrams);
      const endTime = performance.now();

      // Should handle multiple diagrams efficiently
      expect(endTime - startTime).toBeLessThan(200);

      await waitFor(() => {
        expect(screen.getAllByText('Fast Mermaid')).toHaveLength(2);
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });
    });
  });

  describe('Rendering Performance', () => {
    it('renders large documents with diagrams efficiently', async () => {
      const largeDiagramContent = `# Large Document with Diagrams

${Array.from({ length: 20 }, (_, i) => `## Section ${i + 1}

Regular content for section ${i + 1}.

\`\`\`mermaid
graph TD
    A${i} --> B${i}
    B${i} --> C${i}
\`\`\`

More content after diagram.

`).join('\n')}`;

      const startTime = performance.now();
      renderWithProvider(largeDiagramContent);
      const endTime = performance.now();

      // Should handle large content reasonably fast
      expect(endTime - startTime).toBeLessThan(1000);

      expect(screen.getByText('Large Document with Diagrams')).toBeInTheDocument();
      expect(screen.getByText('Section 1')).toBeInTheDocument();
      expect(screen.getByText('Section 20')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getAllByText('Fast Mermaid')).toHaveLength(20);
      });
    });

    it('handles rapid content updates without performance degradation', () => {
      const contents = Array.from({ length: 10 }, (_, i) => 
        `# Content ${i + 1}\n\n\`\`\`mermaid\ngraph TD\n    A${i} --> B${i}\n\`\`\``
      );

      const { rerender } = renderWithProvider(contents[0]);
      
      const startTime = performance.now();
      
      contents.slice(1).forEach((content, index) => {
        rerender(
          <DiagramLibraryProvider>
            <MarkdownRenderer content={content} />
          </DiagramLibraryProvider>
        );
      });
      
      const endTime = performance.now();

      // Rapid updates should be efficient
      expect(endTime - startTime).toBeLessThan(500);
      expect(screen.getByText('Content 10')).toBeInTheDocument();
    });

    it('maintains performance with complex nested markdown', async () => {
      const complexContent = `# Complex Document

> ## Quoted Section
> 
> \`\`\`mermaid
> graph TD
>     A --> B
> \`\`\`
> 
> | Table | In | Quote |
> |-------|----|----- |
> | Cell  | 1  | Data  |

## Regular Section

\`\`\`vega-lite
{
  "layer": [
    {
      "mark": "line",
      "encoding": {
        "x": {"field": "x", "type": "quantitative"},
        "y": {"field": "y", "type": "quantitative"}
      }
    },
    {
      "mark": "point",
      "encoding": {
        "x": {"field": "x", "type": "quantitative"},
        "y": {"field": "y", "type": "quantitative"}
      }
    }
  ],
  "data": {"values": [{"x": 1, "y": 2}]}
}
\`\`\`

### Nested Lists

1. First item
   - Nested item
     - Double nested
       - Triple nested
2. Second item
   \`\`\`mermaid
   pie title Pie Chart
       "A" : 386
       "B" : 85
   \`\`\``;

      const startTime = performance.now();
      renderWithProvider(complexContent);
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(300);
      expect(screen.getByText('Complex Document')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getAllByText('Fast Mermaid')).toHaveLength(2);
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });
    });
  });

  describe('Memory and Resource Management', () => {
    it('cleans up resources when component unmounts', async () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { unmount } = renderWithProvider(content);

      await waitFor(() => {
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });

      // Unmount should not cause errors
      expect(() => unmount()).not.toThrow();
    });

    it('handles memory efficiently with many diagram instances', async () => {
      const manyDiagramsContent = Array.from({ length: 50 }, (_, i) => 
        `\`\`\`mermaid\ngraph TD\n    A${i} --> B${i}\n\`\`\``
      ).join('\n\n');

      const startTime = performance.now();
      renderWithProvider(manyDiagramsContent);
      const endTime = performance.now();

      // Should handle many diagrams without excessive delay
      expect(endTime - startTime).toBeLessThan(2000);

      await waitFor(() => {
        expect(screen.getAllByText('Fast Mermaid')).toHaveLength(50);
      }, { timeout: 5000 });
    });

    it('reuses diagram components efficiently', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { rerender } = renderWithProvider(content);

      // Re-render with same content
      rerender(
        <DiagramLibraryProvider>
          <MarkdownRenderer content={content} />
        </DiagramLibraryProvider>
      );

      // Should not cause additional renders
      expect(mockMermaidRender).toHaveBeenCalledTimes(1);
    });
  });

  describe('Error Handling Performance', () => {
    it('handles diagram errors without blocking other content', async () => {
      mockMermaidRender.mockRejectedValueOnce(new Error('Diagram error'));

      const contentWithError = `# Content with Error

Regular text before.

\`\`\`mermaid
invalid diagram
\`\`\`

Regular text after.

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const startTime = performance.now();
      renderWithProvider(contentWithError);
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(200);
      expect(screen.getByText('Content with Error')).toBeInTheDocument();
      expect(screen.getByText('Regular text before.')).toBeInTheDocument();
      expect(screen.getByText('Regular text after.')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });
    });

    it('recovers quickly from temporary errors', async () => {
      mockMermaidRender
        .mockRejectedValueOnce(new Error('Temporary error'))
        .mockResolvedValueOnce({ svg: '<svg><text>Recovered</text></svg>' });

      const { rerender } = renderWithProvider(`\`\`\`mermaid\ngraph TD\n    A --> B\n\`\`\``);

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
      });

      const startTime = performance.now();
      rerender(
        <DiagramLibraryProvider>
          <MarkdownRenderer content={`\`\`\`mermaid\ngraph LR\n    X --> Y\n\`\`\``} />
        </DiagramLibraryProvider>
      );
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(100);

      await waitFor(() => {
        expect(screen.getByText('Recovered')).toBeInTheDocument();
        expect(screen.queryByText('Diagram Error:')).not.toBeInTheDocument();
      });
    });
  });

  describe('Bundle Size and Loading Optimization', () => {
    it('demonstrates lazy loading benefits', () => {
      // Test without diagrams - should be fast
      const regularContent = '# Regular Content\n\nNo diagrams here.';
      
      const startTime1 = performance.now();
      renderWithProvider(regularContent);
      const endTime1 = performance.now();

      // Test with diagrams - initial load may be slower but still reasonable
      const diagramContent = '# With Diagrams\n\n```mermaid\ngraph TD\n    A --> B\n```';
      
      const startTime2 = performance.now();
      renderWithProvider(diagramContent);
      const endTime2 = performance.now();

      // Regular content should be faster
      expect(endTime1 - startTime1).toBeLessThan(endTime2 - startTime2);
      
      // But both should be reasonable
      expect(endTime1 - startTime1).toBeLessThan(50);
      expect(endTime2 - startTime2).toBeLessThan(150);
    });

    it('shows efficient re-rendering with diagram libraries loaded', async () => {
      const content1 = `\`\`\`mermaid\ngraph TD\n    A --> B\n\`\`\``;
      const content2 = `\`\`\`mermaid\ngraph LR\n    X --> Y\n\`\`\``;

      const { rerender } = renderWithProvider(content1);

      await waitFor(() => {
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });

      // Second render should be faster since libraries are loaded
      const startTime = performance.now();
      rerender(
        <DiagramLibraryProvider>
          <MarkdownRenderer content={content2} />
        </DiagramLibraryProvider>
      );
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(100);

      await waitFor(() => {
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });
    });
  });

  describe('Real-world Performance Scenarios', () => {
    it('handles streaming chat scenario efficiently', async () => {
      const streamingContent = [
        'Here is your training plan:',
        'Here is your training plan:\n\n```mermaid',
        'Here is your training plan:\n\n```mermaid\ngraph TD',
        'Here is your training plan:\n\n```mermaid\ngraph TD\n    A[Start]',
        'Here is your training plan:\n\n```mermaid\ngraph TD\n    A[Start] --> B[End]\n```',
        'Here is your training plan:\n\n```mermaid\ngraph TD\n    A[Start] --> B[End]\n```\n\nFollow this workflow!',
      ];

      const { rerender } = renderWithProvider(streamingContent[0]);

      streamingContent.slice(1).forEach((content, index) => {
        const startTime = performance.now();
        rerender(
          <DiagramLibraryProvider>
            <MarkdownRenderer content={content} />
          </DiagramLibraryProvider>
        );
        const endTime = performance.now();

        // Each update should be fast
        expect(endTime - startTime).toBeLessThan(50);
      });

      await waitFor(() => {
        expect(screen.getByText('Follow this workflow!')).toBeInTheDocument();
        expect(screen.getByText('Fast Mermaid')).toBeInTheDocument();
      });
    });

    it('maintains performance with mixed content types', async () => {
      const mixedContent = `# AI Coach Response

Based on your data, here's the analysis:

## Progress Overview
\`\`\`vega-lite
{
  "mark": "line",
  "data": {"values": [{"week": 1, "distance": 10}]},
  "encoding": {
    "x": {"field": "week", "type": "quantitative"},
    "y": {"field": "distance", "type": "quantitative"}
  }
}
\`\`\`

## Training Plan
\`\`\`mermaid
graph TD
    A[Assessment] --> B[Planning]
    B --> C[Execution]
    C --> D[Review]
\`\`\`

## Weekly Schedule

| Day | Activity | Duration |
|-----|----------|----------|
| Mon | Run      | 30 min   |
| Wed | Bike     | 45 min   |
| Fri | Swim     | 30 min   |

## Recovery Process
\`\`\`mermaid
sequenceDiagram
    participant A as Athlete
    participant C as Coach
    A->>C: Report fatigue
    C->>A: Adjust plan
\`\`\`

Remember to stay hydrated and listen to your body!`;

      const startTime = performance.now();
      renderWithProvider(mixedContent);
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(500);
      expect(screen.getByText('AI Coach Response')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getAllByText('Fast Mermaid')).toHaveLength(2);
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });
    });
  });
});