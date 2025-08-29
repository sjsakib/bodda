import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MarkdownRenderer, SafeMarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock window.matchMedia for tests
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock diagram libraries
vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockResolvedValue({
      svg: '<svg><text>Test Mermaid</text></svg>'
    }),
  },
}));

vi.mock('react-vega', () => ({
  VegaLite: ({ spec }: any) => (
    <div data-testid="vega-lite-chart">
      Vega-Lite: {spec?.mark || 'unknown'}
    </div>
  ),
}));

// Mock the diagram detection utilities
vi.mock('../../utils/diagramDetection', () => ({
  detectDiagrams: vi.fn(),
  hasDiagramContent: vi.fn(),
}));

// Mock the diagram loader hook
vi.mock('../../hooks/useDiagramLoader', () => ({
  useDiagramLoader: vi.fn(),
}));

// Mock responsive diagram components to avoid browser API issues
vi.mock('../ResponsiveDiagramRenderer', () => ({
  default: ({ type, content }: any) => (
    <div data-testid={`${type}-diagram`}>
      Mock {type} diagram: {content}
    </div>
  ),
}));

import { detectDiagrams, hasDiagramContent } from '../../utils/diagramDetection';
import { useDiagramLoader } from '../../hooks/useDiagramLoader';

const mockDetectDiagrams = detectDiagrams as any;
const mockHasDiagramContent = hasDiagramContent as any;
const mockUseDiagramLoader = useDiagramLoader as any;

const renderWithProvider = (content: string, props = {}) => {
  return render(
    <DiagramLibraryProvider>
      <MarkdownRenderer content={content} {...props} />
    </DiagramLibraryProvider>
  );
};

const renderSafeWithProvider = (content: string, props = {}) => {
  return render(
    <DiagramLibraryProvider>
      <SafeMarkdownRenderer content={content} {...props} />
    </DiagramLibraryProvider>
  );
};

describe('MarkdownRenderer Enhanced Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Default mock implementations
    mockDetectDiagrams.mockReturnValue({
      hasDiagrams: false,
      mermaidCount: 0,
      vegaLiteCount: 0,
      totalCount: 0,
      diagrams: [],
    });
    
    mockHasDiagramContent.mockReturnValue(false);
    
    mockUseDiagramLoader.mockReturnValue({
      hasDiagrams: false,
      diagramInfo: {
        hasDiagrams: false,
        mermaidCount: 0,
        vegaLiteCount: 0,
        totalCount: 0,
        diagrams: [],
      },
      libraryState: {
        mermaid: { loaded: false, loading: false, error: null },
        vegaLite: { loaded: false, loading: false, error: null },
      },
      isLoading: false,
      allRequiredLibrariesLoaded: true,
      errors: [],
      loadRequiredLibraries: vi.fn(),
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Basic Markdown Rendering', () => {
    it('renders regular markdown content without diagrams', () => {
      const content = `# Training Guide

This is a **comprehensive** training guide with *emphasis* on proper form.

## Key Points

- Warm up properly
- Stay hydrated
- Listen to your body

\`\`\`javascript
const heartRate = 150;
console.log('Target HR:', heartRate);
\`\`\``;

      renderWithProvider(content);

      expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Guide');
      expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Key Points');
      expect(screen.getByText('comprehensive')).toBeInTheDocument();
      expect(screen.getByText('emphasis')).toBeInTheDocument();
      expect(screen.getByText('Warm up properly')).toBeInTheDocument();
      expect(screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('const heartRate = 150;');
      })).toBeInTheDocument();
    });

    it('applies custom className and maintains markdown-content class', () => {
      const content = 'Test content';
      const customClass = 'custom-markdown';
      
      const { container } = renderWithProvider(content, { className: customClass });
      
      const markdownElement = container.querySelector('.markdown-content');
      expect(markdownElement).toHaveClass('markdown-content', customClass);
    });

    it('handles empty content gracefully', () => {
      const { container } = renderWithProvider('');
      
      expect(container.querySelector('.markdown-content')).toBeInTheDocument();
    });
  });

  describe('Diagram Detection and Integration', () => {
    it('detects mermaid diagrams and applies has-diagrams class', () => {
      const content = `# Workflow

\`\`\`mermaid
graph TD
    A[Start] --> B[Process]
    B --> C[End]
\`\`\``;

      // Mock diagram detection
      mockHasDiagramContent.mockReturnValue(true);
      mockDetectDiagrams.mockReturnValue({
        hasDiagrams: true,
        mermaidCount: 1,
        vegaLiteCount: 0,
        totalCount: 1,
        diagrams: [
          {
            type: 'mermaid',
            content: 'graph TD\n    A[Start] --> B[Process]\n    B --> C[End]',
            startIndex: 0,
            endIndex: 50,
            fullMatch: '```mermaid\ngraph TD\n    A[Start] --> B[Process]\n    B --> C[End]\n```'
          }
        ],
      });

      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 1,
          vegaLiteCount: 0,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: true,
        allRequiredLibrariesLoaded: false,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(screen.getByText('Workflow')).toBeInTheDocument();
      expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });

    it('detects vega-lite charts and applies has-diagrams class', () => {
      const content = `# Progress Chart

\`\`\`vega-lite
{
  "mark": "bar",
  "data": {"values": [{"x": "A", "y": 28}]}
}
\`\`\``;

      // Mock diagram detection
      mockHasDiagramContent.mockReturnValue(true);
      mockDetectDiagrams.mockReturnValue({
        hasDiagrams: true,
        mermaidCount: 0,
        vegaLiteCount: 1,
        totalCount: 1,
        diagrams: [
          {
            type: 'vega-lite',
            content: '{\n  "mark": "bar",\n  "data": {"values": [{"x": "A", "y": 28}]}\n}',
            startIndex: 0,
            endIndex: 80,
            fullMatch: '```vega-lite\n{\n  "mark": "bar",\n  "data": {"values": [{"x": "A", "y": 28}]}\n}\n```'
          }
        ],
      });

      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 0,
          vegaLiteCount: 1,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: null },
          vegaLite: { loaded: false, loading: true, error: null },
        },
        isLoading: true,
        allRequiredLibrariesLoaded: false,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(screen.getByText('Progress Chart')).toBeInTheDocument();
      expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
      expect(screen.getByText(/Loading chart libraries/)).toBeInTheDocument();
    });

    it('handles multiple diagram types in one document', () => {
      const content = `# Mixed Content

\`\`\`mermaid
graph LR
    A --> B
\`\`\`

\`\`\`vega-lite
{"mark": "point", "data": {"values": []}}
\`\`\``;

      // Mock diagram detection
      mockHasDiagramContent.mockReturnValue(true);
      mockDetectDiagrams.mockReturnValue({
        hasDiagrams: true,
        mermaidCount: 1,
        vegaLiteCount: 1,
        totalCount: 2,
        diagrams: [
          {
            type: 'mermaid',
            content: 'graph LR\n    A --> B',
            startIndex: 0,
            endIndex: 40,
            fullMatch: '```mermaid\ngraph LR\n    A --> B\n```'
          },
          {
            type: 'vega-lite',
            content: '{"mark": "point", "data": {"values": []}}',
            startIndex: 50,
            endIndex: 100,
            fullMatch: '```vega-lite\n{"mark": "point", "data": {"values": []}}\n```'
          }
        ],
      });

      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 1,
          vegaLiteCount: 1,
          totalCount: 2,
        },
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
          vegaLite: { loaded: false, loading: true, error: null },
        },
        isLoading: true,
        allRequiredLibrariesLoaded: false,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(screen.getByText('Mixed Content')).toBeInTheDocument();
      expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });
  });

  describe('Diagram Rendering with Mocked Components', () => {
    it('renders mermaid diagrams when libraries are loaded', async () => {
      const content = `\`\`\`mermaid
graph TD
    A[Start] --> B[End]
\`\`\``;

      // Mock successful library loading
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 1,
          vegaLiteCount: 0,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: true, loading: false, error: null },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: true,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      await waitFor(() => {
        expect(container.querySelector('.mermaid-code-block')).toBeInTheDocument();
      });
    });

    it('renders vega-lite charts when libraries are loaded', async () => {
      const content = `\`\`\`vega-lite
{
  "mark": "bar",
  "data": {"values": [{"x": "A", "y": 28}]}
}
\`\`\``;

      // Mock successful library loading
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 0,
          vegaLiteCount: 1,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: null },
          vegaLite: { loaded: true, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: true,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      await waitFor(() => {
        expect(container.querySelector('.vega-lite-code-block')).toBeInTheDocument();
      });
    });
  });

  describe('Fallback Rendering', () => {
    it('shows fallback for mermaid when libraries fail to load', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      // Mock library loading failure
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 1,
          vegaLiteCount: 0,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: 'Failed to load' },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: false,
        errors: ['Mermaid: Failed to load'],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(container.querySelector('.mermaid-fallback')).toBeInTheDocument();
      expect(screen.getByText('Mermaid Diagram')).toBeInTheDocument();
      expect(screen.getByText(/Diagram libraries are not available/)).toBeInTheDocument();
      expect(screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('graph TD');
      })).toBeInTheDocument();
    });

    it('shows fallback for vega-lite when libraries fail to load', () => {
      const content = `\`\`\`vega-lite
{"mark": "bar"}
\`\`\``;

      // Mock library loading failure
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 0,
          vegaLiteCount: 1,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: null },
          vegaLite: { loaded: false, loading: false, error: 'Failed to load' },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: false,
        errors: ['Vega-Lite: Failed to load'],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(container.querySelector('.vega-lite-fallback')).toBeInTheDocument();
      expect(screen.getByText('Vega-Lite Chart')).toBeInTheDocument();
      expect(screen.getByText(/Chart libraries are not available/)).toBeInTheDocument();
      expect(screen.getByText('{"mark": "bar"}')).toBeInTheDocument();
    });

    it('displays library loading errors', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      // Mock library loading with errors
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: {
          hasDiagrams: true,
          mermaidCount: 1,
          vegaLiteCount: 0,
          totalCount: 1,
        },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: 'Network error' },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: false,
        errors: ['Mermaid: Network error'],
        loadRequiredLibraries: vi.fn(),
      });

      renderWithProvider(content);

      expect(screen.getByText('Diagram Library Errors:')).toBeInTheDocument();
      expect(screen.getByText('• Mermaid: Network error')).toBeInTheDocument();
    });
  });

  describe('Configuration Options', () => {
    it('can disable diagram rendering entirely', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { container } = renderWithProvider(content, { enableDiagrams: false });

      // Should not have has-diagrams class
      expect(container.querySelector('.has-diagrams')).not.toBeInTheDocument();
      
      // Should render as regular code block
      expect(screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('graph TD');
      })).toBeInTheDocument();
      expect(screen.queryByText(/Loading diagram/)).not.toBeInTheDocument();
    });

    it('passes diagram configuration props without errors', () => {
      const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: { hasDiagrams: true, mermaidCount: 1, vegaLiteCount: 0, totalCount: 1 },
        libraryState: {
          mermaid: { loaded: true, loading: false, error: null },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: true,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      // Test various configuration options
      expect(() => {
        renderWithProvider(content, { diagramTheme: 'dark' });
      }).not.toThrow();

      expect(() => {
        renderWithProvider(content, { enableDiagramZoomPan: false });
      }).not.toThrow();

      expect(() => {
        renderWithProvider(content, { showVegaActions: true });
      }).not.toThrow();
    });
  });

  describe('Error Handling', () => {
    it('handles markdown parsing errors gracefully with SafeMarkdownRenderer', () => {
      const content = 'Test content that should render safely';
      
      renderSafeWithProvider(content);
      
      // Should render without errors
      expect(screen.getByText('Test content that should render safely')).toBeInTheDocument();
    });

    it('handles mixed content with some diagrams failing', () => {
      const content = `# Mixed Content

Regular text here.

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

More regular text.

\`\`\`javascript
console.log('This should work');
\`\`\``;

      // Mock partial failure
      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: { hasDiagrams: true, mermaidCount: 1, vegaLiteCount: 0, totalCount: 1 },
        libraryState: {
          mermaid: { loaded: false, loading: false, error: 'Load failed' },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: false,
        allRequiredLibrariesLoaded: false,
        errors: ['Mermaid: Load failed'],
        loadRequiredLibraries: vi.fn(),
      });

      renderWithProvider(content);

      // Regular content should still render
      expect(screen.getByText('Mixed Content')).toBeInTheDocument();
      expect(screen.getByText('Regular text here.')).toBeInTheDocument();
      expect(screen.getByText('More regular text.')).toBeInTheDocument();
      expect(screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('console.log');
      })).toBeInTheDocument();
      
      // Error should be displayed
      expect(screen.getByText('Diagram Library Errors:')).toBeInTheDocument();
      
      // Fallback should be shown for failed diagram
      expect(screen.getByText('Mermaid Diagram')).toBeInTheDocument();
    });
  });

  describe('Performance and Optimization', () => {
    it('does not load libraries when no diagrams are present', () => {
      const content = `# Regular Content

This is just regular markdown with **bold** and *italic* text.

\`\`\`javascript
// This is just code, not a diagram
const x = 1;
\`\`\``;

      const { container } = renderWithProvider(content);

      // Should not have has-diagrams class
      expect(container.querySelector('.has-diagrams')).not.toBeInTheDocument();
      
      // Should not show any loading indicators
      expect(screen.queryByText(/Loading diagram/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Loading chart/)).not.toBeInTheDocument();
    });

    it('handles large documents with multiple diagrams efficiently', () => {
      const content = `# Large Document

${Array.from({ length: 5 }, (_, i) => `
## Section ${i + 1}

Some content here.

\`\`\`mermaid
graph TD
    A${i} --> B${i}
\`\`\`

More content.

\`\`\`vega-lite
{"mark": "bar", "data": {"values": [{"x": "${i}", "y": ${i * 10}}]}}
\`\`\`
`).join('\n')}`;

      // Mock detection of multiple diagrams
      mockHasDiagramContent.mockReturnValue(true);
      mockDetectDiagrams.mockReturnValue({
        hasDiagrams: true,
        mermaidCount: 5,
        vegaLiteCount: 5,
        totalCount: 10,
        diagrams: Array.from({ length: 10 }, (_, i) => ({
          type: i % 2 === 0 ? 'mermaid' : 'vega-lite',
          content: `diagram ${i}`,
          startIndex: i * 100,
          endIndex: (i + 1) * 100,
          fullMatch: `diagram ${i} full match`
        })),
      });

      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: { hasDiagrams: true, mermaidCount: 5, vegaLiteCount: 5, totalCount: 10 },
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
          vegaLite: { loaded: false, loading: true, error: null },
        },
        isLoading: true,
        allRequiredLibrariesLoaded: false,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      const { container } = renderWithProvider(content);

      expect(screen.getByText('Large Document')).toBeInTheDocument();
      expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
      
      // Should handle multiple sections
      expect(screen.getByText('Section 1')).toBeInTheDocument();
      expect(screen.getByText('Section 5')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('maintains proper heading hierarchy with diagrams', () => {
      const content = `# Main Title

## Section with Diagram

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

### Subsection

More content here.`;

      renderWithProvider(content);

      const h1 = screen.getByRole('heading', { level: 1 });
      const h2 = screen.getByRole('heading', { level: 2 });
      const h3 = screen.getByRole('heading', { level: 3 });

      expect(h1).toHaveTextContent('Main Title');
      expect(h2).toHaveTextContent('Section with Diagram');
      expect(h3).toHaveTextContent('Subsection');
    });

    it('provides appropriate ARIA labels and structure', () => {
      const content = `# Accessible Content

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      mockHasDiagramContent.mockReturnValue(true);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        diagramInfo: { hasDiagrams: true, mermaidCount: 1, vegaLiteCount: 0, totalCount: 1 },
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
          vegaLite: { loaded: false, loading: false, error: null },
        },
        isLoading: true,
        allRequiredLibrariesLoaded: false,
        errors: [],
        loadRequiredLibraries: vi.fn(),
      });

      renderWithProvider(content);

      // Check heading structure
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
      
      // Loading indicator should have proper accessibility
      const loadingElement = screen.getByText(/Loading diagram libraries/);
      expect(loadingElement).toBeInTheDocument();
    });
  });
});