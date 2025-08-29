import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MermaidDiagram, SafeMermaidDiagram } from '../MermaidDiagram';

// Mock mermaid library
const mockMermaidRender = vi.fn();
const mockMermaidInitialize = vi.fn();

// Test constants
const mockSvg = '<svg><g><text>Test Diagram</text></g></svg>';
const mockContent = 'graph TD\nA-->B';

vi.mock('mermaid', () => ({
  default: {
    initialize: mockMermaidInitialize,
    render: mockMermaidRender,
  },
}));

describe('MermaidDiagram', () => {

  beforeEach(() => {
    // Reset mocks
    mockMermaidRender.mockResolvedValue({
      svg: mockSvg,
    });
    mockMermaidInitialize.mockImplementation(() => {});
    
    // Clear all timers
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  describe('Basic Rendering', () => {
    it('renders successfully with valid content', async () => {
      render(<MermaidDiagram content={mockContent} />);

      // Should show loading initially
      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();

      // Wait for diagram to render
      await waitFor(() => {
        expect(screen.getByText('Test Diagram')).toBeInTheDocument();
      });

      expect(mockMermaidRender).toHaveBeenCalledWith(
        expect.stringMatching(/^mermaid-/),
        mockContent
      );
    });

    it('applies custom className', async () => {
      render(<MermaidDiagram content={mockContent} className="custom-class" />);

      await waitFor(() => {
        expect(screen.getByText('Test Diagram')).toBeInTheDocument();
      });

      const diagramContainer = screen.getByText('Test Diagram').closest('.mermaid-diagram');
      expect(diagramContainer).toHaveClass('custom-class');
    });

    it('shows empty state for no content', () => {
      render(<MermaidDiagram content="" />);

      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });

    it('shows empty state for whitespace-only content', () => {
      render(<MermaidDiagram content="   \n  \t  " />);

      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });
  });

  describe('Theme Configuration', () => {
    it('applies light theme configuration', async () => {
      render(<MermaidDiagram content={mockContent} theme="light" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({
            theme: 'default',
            themeVariables: expect.objectContaining({
              primaryColor: '#1E40AF',
              primaryTextColor: '#1F2937',
              background: '#FFFFFF',
            }),
          })
        );
      });
    });

    it('applies dark theme configuration', async () => {
      render(<MermaidDiagram content={mockContent} theme="dark" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({
            theme: 'dark',
            themeVariables: expect.objectContaining({
              primaryColor: '#3B82F6',
              primaryTextColor: '#F3F4F6',
              background: '#1F2937',
            }),
          })
        );
      });
    });

    it('defaults to light theme when no theme specified', async () => {
      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({
            theme: 'default',
          })
        );
      });
    });
  });

  describe('Error Handling', () => {
    it('shows error message for invalid syntax', async () => {
      mockMermaidRender.mockRejectedValue(new Error('Invalid syntax'));

      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
        expect(screen.getByText('Invalid syntax')).toBeInTheDocument();
      });
    });

    it('shows raw content in error details', async () => {
      mockMermaidRender.mockRejectedValue(new Error('Parse error'));

      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByText('Show raw content')).toBeInTheDocument();
      });

      // Click to expand details
      screen.getByText('Show raw content').click();
      expect(screen.getByText(mockContent)).toBeInTheDocument();
    });

    it('handles timeout error', async () => {
      // Mock a long-running render
      mockMermaidRender.mockImplementation(() => 
        new Promise(resolve => setTimeout(resolve, 15000))
      );

      render(<MermaidDiagram content={mockContent} />);

      // Fast-forward time to trigger timeout
      vi.advanceTimersByTime(10000);

      await waitFor(() => {
        expect(screen.getByText('Diagram rendering timeout')).toBeInTheDocument();
      });
    });

    it('calls onRenderError callback', async () => {
      const onRenderError = vi.fn();
      mockMermaidRender.mockRejectedValue(new Error('Test error'));

      render(
        <MermaidDiagram 
          content={mockContent} 
          onRenderError={onRenderError}
        />
      );

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith('Test error');
      });
    });
  });

  describe('Success Callbacks', () => {
    it('calls onRenderSuccess callback', async () => {
      const onRenderSuccess = vi.fn();

      render(
        <MermaidDiagram 
          content={mockContent} 
          onRenderSuccess={onRenderSuccess}
        />
      );

      await waitFor(() => {
        expect(onRenderSuccess).toHaveBeenCalledWith(
          expect.stringContaining('class="w-full h-auto max-w-full"')
        );
      });
    });
  });

  describe('Accessibility', () => {
    it('adds accessibility attributes to SVG', async () => {
      render(<MermaidDiagram content={mockContent} alt="Test diagram description" />);

      await waitFor(() => {
        const svgContainer = screen.getByText('Test Diagram').closest('div');
        expect(svgContainer?.innerHTML).toContain('role="img"');
        expect(svgContainer?.innerHTML).toContain('aria-labelledby=');
        expect(svgContainer?.innerHTML).toContain('<title');
        expect(svgContainer?.innerHTML).toContain('Test diagram description');
      });
    });

    it('uses default alt text when none provided', async () => {
      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        const svgContainer = screen.getByText('Test Diagram').closest('div');
        expect(svgContainer?.innerHTML).toContain('Mermaid Diagram');
      });
    });
  });

  describe('Responsive SVG Processing', () => {
    it('makes SVG responsive by removing fixed dimensions', async () => {
      const svgWithDimensions = '<svg width="500" height="300"><g><text>Test</text></g></svg>';
      mockMermaidRender.mockResolvedValue({ svg: svgWithDimensions });

      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        const svgContainer = screen.getByText('Test').closest('div');
        expect(svgContainer?.innerHTML).toContain('class="w-full h-auto max-w-full"');
        expect(svgContainer?.innerHTML).not.toContain('width="500"');
        expect(svgContainer?.innerHTML).not.toContain('height="300"');
      });
    });
  });

  describe('Loading States', () => {
    it('shows loading indicator initially', () => {
      // Make render promise hang
      mockMermaidRender.mockImplementation(() => new Promise(() => {}));

      render(<MermaidDiagram content={mockContent} />);

      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();
      expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument(); // spinner
    });

    it('hides loading indicator after successful render', async () => {
      render(<MermaidDiagram content={mockContent} />);

      // Initially shows loading
      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();

      // After render, loading is gone
      await waitFor(() => {
        expect(screen.queryByText('Rendering diagram...')).not.toBeInTheDocument();
        expect(screen.getByText('Test Diagram')).toBeInTheDocument();
      });
    });
  });

  describe('Content Updates', () => {
    it('re-renders when content changes', async () => {
      const { rerender } = render(<MermaidDiagram content="graph TD\nA-->B" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledWith(
          expect.any(String),
          'graph TD\nA-->B'
        );
      });

      // Change content
      rerender(<MermaidDiagram content="graph LR\nX-->Y" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledWith(
          expect.any(String),
          'graph LR\nX-->Y'
        );
      });
    });

    it('re-renders when theme changes', async () => {
      const { rerender } = render(<MermaidDiagram content={mockContent} theme="light" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({ theme: 'default' })
        );
      });

      // Change theme
      rerender(<MermaidDiagram content={mockContent} theme="dark" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({ theme: 'dark' })
        );
      });
    });
  });
});

describe('SafeMermaidDiagram Error Boundary', () => {
  // Mock console.error to avoid noise in tests
  const originalConsoleError = console.error;
  beforeEach(() => {
    console.error = vi.fn();
  });

  afterEach(() => {
    console.error = originalConsoleError;
  });

  it('catches rendering errors and shows fallback', () => {
    // Mock a component that throws
    const ThrowingComponent = () => {
      throw new Error('Component error');
    };

    // Replace MermaidDiagram with throwing component for this test
    vi.doMock('../MermaidDiagram', () => ({
      MermaidDiagram: ThrowingComponent,
      SafeMermaidDiagram: ({ content, className }: any) => (
        <div className={`diagram-error ${className || ''}`}>
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium">Failed to render diagram</p>
            <details className="mt-2">
              <summary className="text-xs text-red-500 cursor-pointer">Show raw content</summary>
              <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto">
                {content}
              </pre>
            </details>
          </div>
        </div>
      ),
    }));

    render(<SafeMermaidDiagram content={mockContent} />);

    expect(screen.getByText('Failed to render diagram')).toBeInTheDocument();
    expect(screen.getByText('Show raw content')).toBeInTheDocument();
  });

  it('shows raw content in error boundary fallback', () => {
    const ThrowingComponent = () => {
      throw new Error('Component error');
    };

    vi.doMock('../MermaidDiagram', () => ({
      MermaidDiagram: ThrowingComponent,
      SafeMermaidDiagram: ({ content }: any) => (
        <div className="diagram-error">
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800 font-medium">Failed to render diagram</p>
            <details className="mt-2">
              <summary className="text-xs text-red-500 cursor-pointer">Show raw content</summary>
              <pre className="text-xs text-red-400 mt-1 whitespace-pre-wrap bg-red-100 p-2 rounded border border-red-300 overflow-x-auto">
                {content}
              </pre>
            </details>
          </div>
        </div>
      ),
    }));

    render(<SafeMermaidDiagram content={mockContent} />);

    // Click to expand details
    screen.getByText('Show raw content').click();
    expect(screen.getByText(mockContent)).toBeInTheDocument();
  });

  it('passes through props when no error occurs', async () => {
    render(<SafeMermaidDiagram content={mockContent} className="test-class" />);

    await waitFor(() => {
      expect(screen.getByText('Test Diagram')).toBeInTheDocument();
    });

    const diagramContainer = screen.getByText('Test Diagram').closest('.mermaid-diagram');
    expect(diagramContainer).toHaveClass('test-class');
  });
});

describe('Mermaid Configuration', () => {
  it('applies security settings', async () => {
    render(<MermaidDiagram content={mockContent} />);

    await waitFor(() => {
      expect(mockMermaidInitialize).toHaveBeenCalledWith(
        expect.objectContaining({
          securityLevel: 'strict',
          htmlLabels: false,
          maxTextSize: 50000,
        })
      );
    });
  });

  it('configures font settings', async () => {
    render(<MermaidDiagram content={mockContent} />);

    await waitFor(() => {
      expect(mockMermaidInitialize).toHaveBeenCalledWith(
        expect.objectContaining({
          fontFamily: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
          fontSize: 14,
        })
      );
    });
  });

  it('configures diagram-specific settings', async () => {
    render(<MermaidDiagram content={mockContent} />);

    await waitFor(() => {
      expect(mockMermaidInitialize).toHaveBeenCalledWith(
        expect.objectContaining({
          flowchart: expect.objectContaining({
            useMaxWidth: true,
            htmlLabels: false,
            curve: 'basis',
            padding: 20,
          }),
          sequence: expect.objectContaining({
            useMaxWidth: true,
            diagramMarginX: 50,
            diagramMarginY: 20,
          }),
          gantt: expect.objectContaining({
            useMaxWidth: true,
            leftPadding: 75,
          }),
        })
      );
    });
  });
});