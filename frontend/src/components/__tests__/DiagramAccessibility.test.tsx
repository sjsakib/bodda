import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { MermaidDiagram } from '../MermaidDiagram';
import { VegaLiteDiagram } from '../VegaLiteDiagram';
import { MarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock diagram libraries with accessibility features
vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockResolvedValue({
      svg: `<svg role="img" aria-labelledby="diagram-title-123" aria-describedby="diagram-desc-123">
        <title id="diagram-title-123">Training Workflow</title>
        <desc id="diagram-desc-123">A flowchart showing the training process from start to finish</desc>
        <g><text>Start → Warm Up → Exercise → Cool Down → End</text></g>
      </svg>`
    }),
  },
}));

vi.mock('react-vega', () => ({
  VegaLite: ({ spec, ...props }: any) => (
    <div 
      data-testid="vega-lite-chart" 
      role="img" 
      aria-label={`${spec?.mark || 'Chart'} visualization`}
      tabIndex={0}
      {...props}
    >
      <title>Progress Chart</title>
      <desc>Bar chart showing weekly training progress</desc>
      <svg>
        <g><text>Week 1: 15km, Week 2: 18km, Week 3: 22km</text></g>
      </svg>
    </div>
  ),
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

describe('Diagram Accessibility Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('MermaidDiagram Accessibility', () => {
    it('provides proper ARIA attributes', async () => {
      renderWithProvider(
        <MermaidDiagram 
          content="graph TD\nA[Start]-->B[End]" 
          alt="Training workflow diagram"
        />
      );

      await waitFor(() => {
        const diagram = screen.getByRole('img');
        expect(diagram).toBeInTheDocument();
        expect(diagram).toHaveAttribute('aria-labelledby');
        expect(diagram).toHaveAttribute('aria-describedby');
      });
    });

    it('includes descriptive title and description', async () => {
      renderWithProvider(
        <MermaidDiagram 
          content="graph TD\nA-->B" 
          alt="Process flow"
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Training Workflow')).toBeInTheDocument();
        expect(screen.getByText(/A flowchart showing the training process/)).toBeInTheDocument();
      });
    });

    it('provides default alt text when none specified', async () => {
      renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      await waitFor(() => {
        const svgContainer = screen.getByText(/Start.*Warm Up.*Exercise/);
        expect(svgContainer.closest('div')?.innerHTML).toContain('Mermaid Diagram');
      });
    });

    it('maintains accessibility during error states', async () => {
      const mockMermaid = await import('mermaid');
      (mockMermaid.default.render as any).mockRejectedValueOnce(new Error('Invalid syntax'));

      renderWithProvider(
        <MermaidDiagram content="invalid syntax" alt="Broken diagram" />
      );

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
        expect(screen.getByText('Show raw content')).toBeInTheDocument();
      });

      // Error state should be accessible
      const errorContainer = screen.getByText('Diagram Error:').closest('.mermaid-error');
      expect(errorContainer).toBeInTheDocument();

      // Expandable content should be keyboard accessible
      const expandButton = screen.getByText('Show raw content');
      expect(expandButton).toBeInTheDocument();
      
      fireEvent.click(expandButton);
      expect(screen.getByText('invalid syntax')).toBeInTheDocument();
    });

    it('supports keyboard navigation', async () => {
      renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      await waitFor(() => {
        const diagram = screen.getByRole('img');
        expect(diagram).toBeInTheDocument();
        
        // Should be focusable
        diagram.focus();
        expect(document.activeElement).toBe(diagram);
      });
    });

    it('provides screen reader compatible content', async () => {
      renderWithProvider(
        <MermaidDiagram 
          content="graph TD\nA[Assessment]-->B[Planning]-->C[Execution]" 
          alt="Training methodology flowchart"
        />
      );

      await waitFor(() => {
        // Should have proper semantic structure
        const diagram = screen.getByRole('img');
        expect(diagram).toHaveAttribute('aria-labelledby');
        
        const titleId = diagram.getAttribute('aria-labelledby');
        const title = document.getElementById(titleId!);
        expect(title).toBeInTheDocument();
        expect(title?.textContent).toContain('Training Workflow');
      });
    });
  });

  describe('VegaLiteDiagram Accessibility', () => {
    const validSpec = {
      mark: 'bar',
      data: { values: [{ week: 'Week 1', distance: 15 }] },
      encoding: {
        x: { field: 'week', type: 'ordinal' },
        y: { field: 'distance', type: 'quantitative' },
      },
    };

    it('provides proper ARIA attributes', async () => {
      renderWithProvider(
        <VegaLiteDiagram 
          content={JSON.stringify(validSpec)} 
          alt="Weekly progress chart"
        />
      );

      await waitFor(() => {
        const chart = screen.getByRole('img');
        expect(chart).toBeInTheDocument();
        expect(chart).toHaveAttribute('aria-label', 'Weekly progress chart');
        expect(chart).toHaveAttribute('tabIndex', '0');
      });
    });

    it('includes descriptive content for screen readers', async () => {
      renderWithProvider(
        <VegaLiteDiagram 
          content={JSON.stringify(validSpec)} 
          alt="Training progress visualization"
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Chart description: Training progress visualization')).toBeInTheDocument();
        expect(screen.getByText('Progress Chart')).toBeInTheDocument();
        expect(screen.getByText(/Bar chart showing weekly training progress/)).toBeInTheDocument();
      });
    });

    it('provides default accessibility when alt not specified', async () => {
      renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(validSpec)} />
      );

      await waitFor(() => {
        const chart = screen.getByRole('img');
        expect(chart).toHaveAttribute('aria-label', 'bar visualization');
        expect(screen.getByText('Chart description: Interactive Vega-Lite visualization')).toBeInTheDocument();
      });
    });

    it('maintains accessibility during error states', async () => {
      renderWithProvider(
        <VegaLiteDiagram content="{ invalid json }" alt="Broken chart" />
      );

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText('Show raw JSON')).toBeInTheDocument();
      });

      // Error state should be accessible
      const errorContainer = screen.getByText('Chart Error:').closest('.vega-lite-error');
      expect(errorContainer).toBeInTheDocument();

      // Expandable content should be keyboard accessible
      const expandButton = screen.getByText('Show raw JSON');
      fireEvent.click(expandButton);
      expect(screen.getByText('{ invalid json }')).toBeInTheDocument();
    });

    it('supports keyboard interaction', async () => {
      renderWithProvider(
        <VegaLiteDiagram content={JSON.stringify(validSpec)} />
      );

      await waitFor(() => {
        const chart = screen.getByRole('img');
        expect(chart).toBeInTheDocument();
        
        // Should be focusable
        chart.focus();
        expect(document.activeElement).toBe(chart);
        
        // Should support keyboard navigation
        fireEvent.keyDown(chart, { key: 'Enter' });
        fireEvent.keyDown(chart, { key: ' ' });
        fireEvent.keyDown(chart, { key: 'Tab' });
      });
    });

    it('provides meaningful chart data description', async () => {
      const complexSpec = {
        mark: 'line',
        data: { 
          values: [
            { date: '2024-01-01', distance: 10, pace: 8.5 },
            { date: '2024-01-08', distance: 12, pace: 8.2 },
            { date: '2024-01-15', distance: 15, pace: 8.0 },
          ]
        },
        encoding: {
          x: { field: 'date', type: 'temporal' },
          y: { field: 'distance', type: 'quantitative' },
        },
      };

      renderWithProvider(
        <VegaLiteDiagram 
          content={JSON.stringify(complexSpec)} 
          alt="Running progress over time showing increasing distance and improving pace"
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Chart description: Running progress over time showing increasing distance and improving pace')).toBeInTheDocument();
      });
    });
  });

  describe('MarkdownRenderer Accessibility Integration', () => {
    it('maintains heading hierarchy with diagrams', async () => {
      const content = `# Main Training Guide

## Weekly Overview

\`\`\`mermaid
graph TD
    A[Week 1] --> B[Week 2]
\`\`\`

### Detailed Schedule

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\`

#### Daily Activities`;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      // Check heading hierarchy
      expect(screen.getByRole('heading', { level: 1, name: 'Main Training Guide' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 2, name: 'Weekly Overview' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 3, name: 'Detailed Schedule' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 4, name: 'Daily Activities' })).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });
    });

    it('provides accessible table structure with diagrams', async () => {
      const content = `# Training Data

| Week | Distance | Pace |
|------|----------|------|
| 1    | 15km     | 8:30 |
| 2    | 18km     | 8:15 |

\`\`\`mermaid
graph TD
    A[Data] --> B[Analysis]
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      // Check table accessibility
      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();

      const headers = screen.getAllByRole('columnheader');
      expect(headers).toHaveLength(3);
      expect(headers[0]).toHaveTextContent('Week');

      const cells = screen.getAllByRole('cell');
      expect(cells).toHaveLength(6);

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });
    });

    it('maintains link accessibility with diagrams', async () => {
      const content = `# Resources

Check out [this guide](https://example.com/guide) for more information.

\`\`\`mermaid
graph TD
    A[Guide] --> B[Practice]
\`\`\`

Also see [internal section](#section) below.`;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      const externalLink = screen.getByText('this guide');
      expect(externalLink).toHaveAttribute('href', 'https://example.com/guide');
      expect(externalLink).toHaveAttribute('target', '_blank');
      expect(externalLink).toHaveAttribute('rel', 'noopener noreferrer');

      const internalLink = screen.getByText('internal section');
      expect(internalLink).toHaveAttribute('href', '#section');

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });
    });

    it('provides accessible error states for diagrams', async () => {
      const mockUseDiagramLoader = vi.mocked(require('../../hooks/useDiagramLoader').useDiagramLoader);
      mockUseDiagramLoader.mockReturnValue({
        hasDiagrams: true,
        allRequiredLibrariesLoaded: false,
        isLoading: false,
        errors: ['Failed to load Mermaid library'],
      });

      const content = `# Training Plan

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      // Error message should be accessible
      expect(screen.getByText('Diagram Library Errors:')).toBeInTheDocument();
      expect(screen.getByText('• Failed to load Mermaid library')).toBeInTheDocument();

      // Fallback content should be accessible
      expect(screen.getByText('Mermaid Diagram')).toBeInTheDocument();
      expect(screen.getByText('Diagram libraries are not available. Showing raw content:')).toBeInTheDocument();
    });

    it('supports high contrast mode', async () => {
      // Mock high contrast preference
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: vi.fn().mockImplementation(query => ({
          matches: query === '(prefers-contrast: high)',
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
      });

      const content = `# High Contrast Test

\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        expect(screen.getByRole('img')).toBeInTheDocument();
      });

      // Should render without accessibility issues in high contrast mode
      expect(screen.getByText('High Contrast Test')).toBeInTheDocument();
    });
  });

  describe('Screen Reader Compatibility', () => {
    it('provides meaningful content for screen readers', async () => {
      const content = `# AI Coach Analysis

Based on your training data:

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

The trend shows improvement. Here's the recommended workflow:

\`\`\`mermaid
graph TD
    A[Current Level] --> B[Target Level]
    B --> C[Achievement]
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        // Should have accessible content structure
        expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
        expect(screen.getAllByRole('img')).toHaveLength(2);
        
        // Charts should have descriptions
        expect(screen.getByText(/Chart description:/)).toBeInTheDocument();
        
        // Diagrams should have titles
        expect(screen.getByText('Training Workflow')).toBeInTheDocument();
      });
    });

    it('handles complex nested content accessibly', async () => {
      const content = `# Complex Document

> ## Important Note
> 
> \`\`\`mermaid
> graph TD
>     A[Note] --> B[Action]
> \`\`\`
> 
> This is critical information.

## Data Table

| Metric | Value |
|--------|-------|
| Speed  | 12    |

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      // Should maintain proper semantic structure
      expect(screen.getByRole('heading', { level: 1, name: 'Complex Document' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 2, name: 'Important Note' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 2, name: 'Data Table' })).toBeInTheDocument();
      
      expect(screen.getByRole('table')).toBeInTheDocument();
      
      await waitFor(() => {
        expect(screen.getAllByRole('img')).toHaveLength(2);
      });
    });

    it('provides alternative text for complex diagrams', async () => {
      const content = `# System Architecture

\`\`\`mermaid
graph TB
    subgraph "Frontend"
        A[React App]
        B[Components]
    end
    subgraph "Backend"
        C[API Server]
        D[Database]
    end
    A --> C
    C --> D
\`\`\``;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        const diagram = screen.getByRole('img');
        expect(diagram).toBeInTheDocument();
        
        // Should have descriptive content
        expect(screen.getByText(/A flowchart showing the training process/)).toBeInTheDocument();
      });
    });
  });

  describe('Keyboard Navigation', () => {
    it('supports tab navigation through diagram elements', async () => {
      const content = `# Interactive Content

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\`

[Link to more info](https://example.com)`;

      renderWithProvider(
        <MarkdownRenderer content={content} />
      );

      await waitFor(() => {
        expect(screen.getAllByRole('img')).toHaveLength(2);
      });

      // Should be able to tab through interactive elements
      const diagrams = screen.getAllByRole('img');
      const link = screen.getByText('Link to more info');

      // Focus first diagram
      diagrams[0].focus();
      expect(document.activeElement).toBe(diagrams[0]);

      // Tab to second diagram
      fireEvent.keyDown(document.activeElement!, { key: 'Tab' });
      diagrams[1].focus();
      expect(document.activeElement).toBe(diagrams[1]);

      // Tab to link
      fireEvent.keyDown(document.activeElement!, { key: 'Tab' });
      link.focus();
      expect(document.activeElement).toBe(link);
    });

    it('handles keyboard events on diagram elements', async () => {
      renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      await waitFor(() => {
        const diagram = screen.getByRole('img');
        diagram.focus();
        
        // Should handle keyboard events without errors
        fireEvent.keyDown(diagram, { key: 'Enter' });
        fireEvent.keyDown(diagram, { key: ' ' });
        fireEvent.keyDown(diagram, { key: 'Escape' });
        fireEvent.keyDown(diagram, { key: 'ArrowRight' });
        fireEvent.keyDown(diagram, { key: 'ArrowLeft' });
      });
    });
  });

  describe('Focus Management', () => {
    it('maintains focus during diagram loading', async () => {
      const mockMermaid = await import('mermaid');
      (mockMermaid.default.render as any).mockImplementation(() => 
        new Promise(resolve => 
          setTimeout(() => resolve({ 
            svg: '<svg role="img"><text>Loaded</text></svg>' 
          }), 100)
        )
      );

      renderWithProvider(
        <MermaidDiagram content="graph TD\nA-->B" />
      );

      // Focus should be manageable during loading
      const loadingElement = screen.getByText('Rendering diagram...');
      expect(loadingElement).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.getByText('Loaded')).toBeInTheDocument();
      });

      const diagram = screen.getByRole('img');
      diagram.focus();
      expect(document.activeElement).toBe(diagram);
    });

    it('handles focus during error states', async () => {
      const mockMermaid = await import('mermaid');
      (mockMermaid.default.render as any).mockRejectedValue(new Error('Error'));

      renderWithProvider(
        <MermaidDiagram content="invalid" />
      );

      await waitFor(() => {
        const errorButton = screen.getByText('Show raw content');
        errorButton.focus();
        expect(document.activeElement).toBe(errorButton);
        
        fireEvent.click(errorButton);
        expect(screen.getByText('invalid')).toBeInTheDocument();
      });
    });
  });
});