import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { MermaidDiagram } from '../MermaidDiagram';
import { VegaLiteDiagram } from '../VegaLiteDiagram';
import { MarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock diagram libraries with consistent visual output
vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockImplementation((id: string, content: string) => {
      // Generate consistent SVG based on content
      const hash = content.split('').reduce((a, b) => {
        a = ((a << 5) - a) + b.charCodeAt(0);
        return a & a;
      }, 0);
      
      return Promise.resolve({
        svg: `<svg width="400" height="300" viewBox="0 0 400 300" class="mermaid-diagram-${Math.abs(hash)}">
          <title>Mermaid Diagram</title>
          <rect x="10" y="10" width="100" height="50" fill="#e1f5fe" stroke="#01579b" stroke-width="2"/>
          <text x="60" y="40" text-anchor="middle" font-family="Arial" font-size="14">Node A</text>
          <rect x="200" y="10" width="100" height="50" fill="#e8f5e8" stroke="#2e7d32" stroke-width="2"/>
          <text x="250" y="40" text-anchor="middle" font-family="Arial" font-size="14">Node B</text>
          <path d="M110 35 L200 35" stroke="#333" stroke-width="2" marker-end="url(#arrowhead)"/>
          <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0, 10 3.5, 0 7" fill="#333"/>
            </marker>
          </defs>
        </svg>`
      });
    }),
  },
}));

vi.mock('react-vega', () => ({
  VegaLite: ({ spec, ...props }: any) => {
    // Generate consistent chart based on spec
    const mark = spec?.mark || 'bar';
    const dataLength = spec?.data?.values?.length || 0;
    
    return (
      <div 
        data-testid="vega-lite-chart" 
        className={`vega-chart-${mark}`}
        style={{ width: '100%', height: '300px' }}
        {...props}
      >
        <svg width="400" height="300" viewBox="0 0 400 300">
          <title>Vega-Lite Chart</title>
          <rect x="0" y="0" width="400" height="300" fill="#fafafa" stroke="#e0e0e0"/>
          {mark === 'bar' && (
            <>
              <rect x="50" y="200" width="60" height="80" fill="#1976d2"/>
              <rect x="150" y="150" width="60" height="130" fill="#1976d2"/>
              <rect x="250" y="100" width="60" height="180" fill="#1976d2"/>
            </>
          )}
          {mark === 'line' && (
            <path d="M50 250 L150 200 L250 150 L350 100" stroke="#1976d2" stroke-width="3" fill="none"/>
          )}
          {mark === 'point' && (
            <>
              <circle cx="100" cy="200" r="5" fill="#1976d2"/>
              <circle cx="200" cy="150" r="5" fill="#1976d2"/>
              <circle cx="300" cy="100" r="5" fill="#1976d2"/>
            </>
          )}
          <text x="200" y="290" text-anchor="middle" font-family="Arial" font-size="12" fill="#666">
            {mark.charAt(0).toUpperCase() + mark.slice(1)} Chart ({dataLength} data points)
          </text>
        </svg>
      </div>
    );
  },
}));

// Mock diagram loader
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

const renderWithProvider = (component: React.ReactElement) => {
  return render(
    <DiagramLibraryProvider>
      {component}
    </DiagramLibraryProvider>
  );
};

// Helper function to get computed styles
const getComputedStyleSnapshot = (element: Element) => {
  const computedStyle = window.getComputedStyle(element);
  return {
    display: computedStyle.display,
    position: computedStyle.position,
    width: computedStyle.width,
    height: computedStyle.height,
    margin: computedStyle.margin,
    padding: computedStyle.padding,
    border: computedStyle.border,
    borderRadius: computedStyle.borderRadius,
    backgroundColor: computedStyle.backgroundColor,
    color: computedStyle.color,
    fontSize: computedStyle.fontSize,
    fontFamily: computedStyle.fontFamily,
    textAlign: computedStyle.textAlign,
    overflow: computedStyle.overflow,
  };
};

describe('Diagram Visual Regression Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('MermaidDiagram Visual Consistency', () => {
    it('renders consistent visual structure for flowcharts', async () => {
      const { container } = renderWithProvider(
        <MermaidDiagram content="graph TD\nA[Start]-->B[End]" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
        expect(screen.getByText('Node B')).toBeInTheDocument();
      });

      // Check container structure
      const diagramContainer = container.querySelector('.mermaid-diagram');
      expect(diagramContainer).toBeInTheDocument();

      const svgContainer = container.querySelector('.mermaid-svg-container');
      expect(svgContainer).toBeInTheDocument();

      // Check SVG structure
      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();
      expect(svg).toHaveAttribute('viewBox', '0 0 400 300');

      // Check visual elements
      const rectangles = container.querySelectorAll('rect');
      expect(rectangles).toHaveLength(2);
      
      const texts = container.querySelectorAll('text');
      expect(texts).toHaveLength(2);
      
      const paths = container.querySelectorAll('path');
      expect(paths).toHaveLength(1); // Arrow
    });

    it('maintains consistent styling across themes', async () => {
      const lightTheme = renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" theme="light" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
      });

      const lightContainer = lightTheme.container.querySelector('.mermaid-diagram');
      const lightStyles = lightContainer ? getComputedStyleSnapshot(lightContainer) : null;

      lightTheme.unmount();

      const darkTheme = renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" theme="dark" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
      });

      const darkContainer = darkTheme.container.querySelector('.mermaid-diagram');
      const darkStyles = darkContainer ? getComputedStyleSnapshot(darkContainer) : null;

      // Structure should be consistent
      expect(lightStyles?.display).toBe(darkStyles?.display);
      expect(lightStyles?.position).toBe(darkStyles?.position);
      expect(lightStyles?.borderRadius).toBe(darkStyles?.borderRadius);
    });

    it('renders consistent error state visuals', async () => {
      const mockMermaid = await import('mermaid');
      (mockMermaid.default.render as any).mockRejectedValueOnce(new Error('Parse error'));

      const { container } = renderWithProvider(
        <MermaidDiagram content="invalid syntax" />
      );

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
      });

      // Check error container structure
      const errorContainer = container.querySelector('.mermaid-error');
      expect(errorContainer).toBeInTheDocument();

      const errorBox = errorContainer?.querySelector('.bg-red-50');
      expect(errorBox).toBeInTheDocument();
      expect(errorBox).toHaveClass('border', 'border-red-200', 'rounded-lg');

      // Check expandable details
      const details = container.querySelector('details');
      expect(details).toBeInTheDocument();

      const summary = container.querySelector('summary');
      expect(summary).toBeInTheDocument();
      expect(summary).toHaveClass('text-xs', 'text-red-500', 'cursor-pointer');
    });

    it('renders consistent loading state visuals', async () => {
      const mockMermaid = await import('mermaid');
      (mockMermaid.default.render as any).mockImplementation(() => new Promise(() => {}));

      const { container } = renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();

      // Check loading container structure
      const loadingContainer = container.querySelector('.mermaid-loading');
      expect(loadingContainer).toBeInTheDocument();

      const loadingBox = loadingContainer?.querySelector('.bg-gray-50');
      expect(loadingBox).toBeInTheDocument();
      expect(loadingBox).toHaveClass('border', 'border-gray-200', 'rounded-lg');

      // Check spinner
      const spinner = container.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
      expect(spinner).toHaveClass('rounded-full', 'h-6', 'w-6', 'border-b-2', 'border-blue-600');
    });

    it('maintains responsive design consistency', async () => {
      const { container } = renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" className="custom-responsive" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
      });

      // Check responsive container
      const diagramContainer = container.querySelector('.mermaid-diagram');
      expect(diagramContainer).toHaveClass('custom-responsive');

      const responsiveWrapper = container.querySelector('.relative.bg-white');
      expect(responsiveWrapper).toBeInTheDocument();
      expect(responsiveWrapper).toHaveClass('border', 'border-gray-200', 'rounded-lg', 'p-4', 'overflow-auto');

      // Check SVG responsiveness
      const svg = container.querySelector('svg');
      expect(svg?.innerHTML).toContain('class="w-full h-auto max-w-full"');
    });
  });

  describe('VegaLiteDiagram Visual Consistency', () => {
    const barSpec = {
      mark: 'bar',
      data: { values: [{ x: 'A', y: 28 }, { x: 'B', y: 55 }, { x: 'C', y: 43 }] },
      encoding: {
        x: { field: 'x', type: 'ordinal' },
        y: { field: 'y', type: 'quantitative' },
      },
    };

    it('renders consistent visual structure for bar charts', async () => {
      const { container } = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(barSpec)} />
      );

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      // Check container structure
      const diagramContainer = container.querySelector('.vega-lite-diagram');
      expect(diagramContainer).toBeInTheDocument();

      const chartContainer = container.querySelector('.vega-lite-chart-container');
      expect(chartContainer).toBeInTheDocument();

      // Check chart visual elements
      const chart = screen.getByTestId('vega-lite-chart');
      expect(chart).toHaveClass('vega-chart-bar');
      expect(chart).toHaveStyle({ width: '100%', height: '300px' });

      // Check SVG structure
      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();
      expect(svg).toHaveAttribute('viewBox', '0 0 400 300');

      // Check bar elements
      const bars = container.querySelectorAll('rect[fill="#1976d2"]');
      expect(bars).toHaveLength(3);
    });

    it('renders different chart types with consistent styling', async () => {
      const lineSpec = { ...barSpec, mark: 'line' };
      const pointSpec = { ...barSpec, mark: 'point' };

      // Test line chart
      const lineChart = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(lineSpec)} />
      );

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      expect(screen.getByTestId('vega-lite-chart')).toHaveClass('vega-chart-line');
      const linePath = lineChart.container.querySelector('path[stroke="#1976d2"]');
      expect(linePath).toBeInTheDocument();

      lineChart.unmount();

      // Test point chart
      const pointChart = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(pointSpec)} />
      );

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      expect(screen.getByTestId('vega-lite-chart')).toHaveClass('vega-chart-point');
      const circles = pointChart.container.querySelectorAll('circle[fill="#1976d2"]');
      expect(circles).toHaveLength(3);
    });

    it('maintains theme consistency across light and dark modes', async () => {
      const lightChart = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(barSpec)} theme="light" />
      );

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const lightContainer = lightChart.container.querySelector('.vega-lite-diagram');
      const lightWrapper = lightChart.container.querySelector('.relative.bg-white');
      expect(lightWrapper).toBeInTheDocument();

      lightChart.unmount();

      const darkChart = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(barSpec)} theme="dark" />
      );

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const darkContainer = darkChart.container.querySelector('.vega-lite-diagram');
      const darkWrapper = darkChart.container.querySelector('.relative');
      expect(darkWrapper).toBeInTheDocument();
      expect(darkWrapper).toHaveStyle({ backgroundColor: '#1F2937' });

      // Structure should be consistent
      expect(lightContainer?.className).toBe(darkContainer?.className);
    });

    it('renders consistent error state visuals', async () => {
      const { container } = renderWithProvider(
        <VegaLiteDiagram content="{ invalid json }" />
      );

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
      });

      // Check error container structure
      const errorContainer = container.querySelector('.vega-lite-error');
      expect(errorContainer).toBeInTheDocument();

      const errorBox = errorContainer?.querySelector('.bg-red-50');
      expect(errorBox).toBeInTheDocument();
      expect(errorBox).toHaveClass('border', 'border-red-200', 'rounded-lg');

      // Check expandable details
      const details = container.querySelector('details');
      expect(details).toBeInTheDocument();

      const summary = container.querySelector('summary');
      expect(summary).toBeInTheDocument();
      expect(summary).toHaveClass('text-xs', 'text-red-500', 'cursor-pointer');
    });

    it('renders consistent loading state visuals', () => {
      const { container } = renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(barSpec)} />
      );

      // Should show loading initially
      expect(screen.getByText('Loading chart component...')).toBeInTheDocument();

      // Check loading container structure
      const loadingContainer = container.querySelector('.vega-lite-loading');
      expect(loadingContainer).toBeInTheDocument();

      const loadingBox = loadingContainer?.querySelector('.bg-gray-50');
      expect(loadingBox).toBeInTheDocument();
      expect(loadingBox).toHaveClass('border', 'border-gray-200', 'rounded-lg');

      // Check spinner
      const spinner = container.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
      expect(spinner).toHaveClass('rounded-full', 'h-6', 'w-6', 'border-b-2', 'border-blue-600');
    });
  });

  describe('MarkdownRenderer Visual Integration', () => {
    it('maintains consistent layout with mixed content', async () => {
      const content = `# Training Analysis

Regular paragraph text before diagram.

\`\`\`mermaid
graph TD
    A[Start] --> B[End]
\`\`\`

Text between diagrams.

\`\`\`vega-lite
{
  "mark": "bar",
  "data": {"values": [{"x": "A", "y": 28}]},
  "encoding": {
    "x": {"field": "x", "type": "ordinal"},
    "y": {"field": "y", "type": "quantitative"}
  }
}
\`\`\`

Final paragraph text.`;

      const { container } = renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      // Check overall structure
      const markdownContainer = container.querySelector('.markdown-content');
      expect(markdownContainer).toBeInTheDocument();
      expect(markdownContainer).toHaveClass('has-diagrams');

      // Check heading styling
      const heading = screen.getByText('Training Analysis');
      expect(heading.tagName).toBe('H1');
      expect(heading).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl', 'font-bold');

      // Check paragraph styling
      const paragraphs = container.querySelectorAll('p');
      expect(paragraphs).toHaveLength(3);
      paragraphs.forEach(p => {
        expect(p).toHaveClass('mb-3', 'sm:mb-4', 'text-gray-700', 'leading-relaxed');
      });

      // Check diagram containers
      const diagramBlocks = container.querySelectorAll('.diagram-code-block');
      expect(diagramBlocks).toHaveLength(2);
      
      diagramBlocks.forEach(block => {
        expect(block).toHaveClass('mb-3', 'sm:mb-4');
        expect(block.querySelector('.rounded-lg')).toBeInTheDocument();
        expect(block.querySelector('.border')).toBeInTheDocument();
        expect(block.querySelector('.shadow-sm')).toBeInTheDocument();
      });
    });

    it('maintains responsive design consistency', async () => {
      const content = `# Responsive Test

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

| Column 1 | Column 2 |
|----------|----------|
| Data 1   | Data 2   |

\`\`\`vega-lite
{"mark": "line", "data": {"values": []}}
\`\`\``;

      const { container } = renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      // Check responsive classes on various elements
      const heading = screen.getByText('Responsive Test');
      expect(heading).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');

      const table = screen.getByRole('table');
      const tableWrapper = table.closest('.overflow-x-auto');
      expect(tableWrapper).toBeInTheDocument();
      expect(tableWrapper).toHaveClass('mb-3', 'sm:mb-4', 'rounded-lg');

      const diagramBlocks = container.querySelectorAll('.diagram-code-block');
      diagramBlocks.forEach(block => {
        expect(block).toHaveClass('mb-3', 'sm:mb-4');
      });
    });

    it('renders consistent fallback states', () => {
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
\`\`\`

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\``;

      const { container } = renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      // Check fallback containers
      const fallbacks = container.querySelectorAll('.diagram-fallback');
      expect(fallbacks).toHaveLength(2);

      fallbacks.forEach(fallback => {
        expect(fallback).toHaveClass('mb-3', 'sm:mb-4');
        
        const warningBox = fallback.querySelector('.bg-yellow-50');
        expect(warningBox).toBeInTheDocument();
        expect(warningBox).toHaveClass('border', 'border-yellow-200', 'rounded-lg');

        const codeBlock = fallback.querySelector('pre');
        expect(codeBlock).toBeInTheDocument();
        expect(codeBlock).toHaveClass('bg-gray-50', 'border', 'border-gray-200', 'rounded-lg');
      });
    });

    it('maintains visual hierarchy with nested content', async () => {
      const content = `# Main Title

## Section with Diagram

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

### Subsection

> **Important Note**
> 
> \`\`\`vega-lite
> {"mark": "point", "data": {"values": []}}
> \`\`\`

#### Details

Final content here.`;

      const { container } = renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      // Check heading hierarchy
      const h1 = screen.getByRole('heading', { level: 1 });
      expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl', 'font-bold');

      const h2 = screen.getByRole('heading', { level: 2 });
      expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl', 'font-semibold');

      const h3 = screen.getByRole('heading', { level: 3 });
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl', 'font-medium');

      const h4 = screen.getByRole('heading', { level: 4 });
      expect(h4).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl', 'font-medium');

      // Check blockquote styling
      const blockquote = container.querySelector('blockquote');
      expect(blockquote).toBeInTheDocument();
      expect(blockquote).toHaveClass('border-l-4', 'border-blue-200', 'pl-3', 'sm:pl-4', 'bg-blue-50');
    });
  });

  describe('Cross-browser Visual Consistency', () => {
    it('renders consistent SVG elements', async () => {
      const { container } = renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
      });

      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();

      // Check SVG attributes that should be consistent across browsers
      expect(svg).toHaveAttribute('viewBox');
      expect(svg).toHaveAttribute('width');
      expect(svg).toHaveAttribute('height');

      // Check that responsive classes are applied
      expect(svg?.outerHTML).toContain('class="w-full h-auto max-w-full"');
    });

    it('maintains consistent CSS class application', async () => {
      const content = `# Test

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      const { container } = renderWithProvider(
        <MarkdownRenderer content={content} className="test-markdown" />
      );

      await waitFor(() => {
        expect(screen.getByText('Node A')).toBeInTheDocument();
      });

      // Check that all expected classes are present
      const markdownContainer = container.querySelector('.markdown-content');
      expect(markdownContainer).toHaveClass('test-markdown', 'has-diagrams');

      const diagramBlock = container.querySelector('.diagram-code-block');
      expect(diagramBlock).toHaveClass('mermaid-code-block', 'mb-3', 'sm:mb-4');

      const diagramContainer = container.querySelector('.mermaid-diagram');
      expect(diagramContainer).toHaveClass('rounded-lg', 'border', 'border-gray-200', 'shadow-sm');
    });

    it('handles font rendering consistently', async () => {
      const { container } = renderWithProvider(
        <MarkdownRenderer content="# Typography Test\n\nRegular text with **bold** and *italic*." />
      );

      // Check font classes
      const heading = screen.getByText('Typography Test');
      expect(heading).toHaveClass('font-bold');

      const bold = screen.getByText('bold');
      expect(bold).toHaveClass('font-semibold');

      const italic = screen.getByText('italic');
      expect(italic).toHaveClass('italic');

      // Check that text sizing classes are applied
      expect(heading).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');
    });
  });
});