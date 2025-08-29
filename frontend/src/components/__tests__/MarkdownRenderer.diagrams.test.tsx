import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { MarkdownRenderer } from '../MarkdownRenderer';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

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

const renderWithProvider = (content: string, props = {}) => {
  return render(
    <DiagramLibraryProvider>
      <MarkdownRenderer content={content} {...props} />
    </DiagramLibraryProvider>
  );
};

describe('MarkdownRenderer Diagram Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders regular markdown without diagrams', () => {
    const content = '# Hello World\n\nThis is regular markdown.';
    renderWithProvider(content);
    
    expect(screen.getByText('Hello World')).toBeInTheDocument();
    expect(screen.getByText('This is regular markdown.')).toBeInTheDocument();
  });

  it('detects and shows loading for mermaid diagrams', async () => {
    const content = `# Training Plan

\`\`\`mermaid
graph TD
    A[Start] --> B[End]
\`\`\``;

    renderWithProvider(content);
    
    expect(screen.getByText('Training Plan')).toBeInTheDocument();
    
    // Should show loading indicator initially
    await waitFor(() => {
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });
  });

  it('detects and shows loading for vega-lite charts', async () => {
    const content = `# Progress Chart

\`\`\`vega-lite
{
  "mark": "bar",
  "data": {"values": []}
}
\`\`\``;

    renderWithProvider(content);
    
    expect(screen.getByText('Progress Chart')).toBeInTheDocument();
    
    // Should show loading indicator initially
    await waitFor(() => {
      expect(screen.getByText(/Loading chart libraries/)).toBeInTheDocument();
    });
  });

  it('applies has-diagrams class when diagrams are present', () => {
    const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

    const { container } = renderWithProvider(content);
    
    expect(container.querySelector('.has-diagrams')).toBeInTheDocument();
  });

  it('does not apply has-diagrams class when no diagrams are present', () => {
    const content = 'Just regular text';
    
    const { container } = renderWithProvider(content);
    
    expect(container.querySelector('.has-diagrams')).not.toBeInTheDocument();
  });

  it('handles multiple diagram types in one document', async () => {
    const content = `# Mixed Diagrams

\`\`\`mermaid
graph TD
    A --> B
\`\`\`

\`\`\`vega-lite
{"mark": "point", "data": {"values": []}}
\`\`\``;

    renderWithProvider(content);
    
    expect(screen.getByText('Mixed Diagrams')).toBeInTheDocument();
    
    // Should show loading for both types
    await waitFor(() => {
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });
  });

  it('can disable diagram rendering', () => {
    const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

    const { container } = renderWithProvider(content, { enableDiagrams: false });
    
    // Should not have has-diagrams class
    expect(container.querySelector('.has-diagrams')).toBeNull();
    
    // Should render as regular code block
    expect(screen.getByText((content, element) => {
      return element?.tagName === 'CODE' && content.includes('graph TD');
    })).toBeInTheDocument();
  });

  it('passes through diagram theme prop', () => {
    const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

    renderWithProvider(content, { diagramTheme: 'dark' });
    
    // Component should render without errors
    expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
  });

  it('handles diagram zoom and pan settings', () => {
    const content = `\`\`\`mermaid
graph TD
    A --> B
\`\`\``;

    renderWithProvider(content, { enableDiagramZoomPan: false });
    
    // Component should render without errors
    expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
  });

  it('handles vega actions setting', () => {
    const content = `\`\`\`vega-lite
{"mark": "bar", "data": {"values": []}}
\`\`\``;

    renderWithProvider(content, { showVegaActions: true });
    
    // Component should render without errors
    expect(screen.getByText(/Loading chart libraries/)).toBeInTheDocument();
  });
});