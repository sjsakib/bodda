import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MermaidDiagram, SafeMermaidDiagram } from '../MermaidDiagram';

// Mock mermaid library
const mockMermaidRender = vi.fn();
const mockMermaidInitialize = vi.fn();

vi.mock('mermaid', () => ({
  default: {
    initialize: mockMermaidInitialize,
    render: mockMermaidRender,
  },
}));

describe('MermaidDiagram - Comprehensive Unit Tests', () => {
  const mockSvg = '<svg role="img" aria-labelledby="diagram-title"><title id="diagram-title">Test Diagram</title><g><text>Test Content</text></g></svg>';
  const mockContent = 'graph TD\nA[Start]-->B[End]';

  beforeEach(() => {
    mockMermaidRender.mockResolvedValue({ svg: mockSvg });
    mockMermaidInitialize.mockImplementation(() => {});
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  describe('Core Rendering Functionality', () => {
    it('renders diagram with proper SVG processing', async () => {
      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByText('Test Content')).toBeInTheDocument();
      });

      expect(mockMermaidRender).toHaveBeenCalledWith(
        expect.stringMatching(/^mermaid-/),
        mockContent
      );
    });

    it('generates unique diagram IDs for multiple instances', async () => {
      const { rerender } = render(<MermaidDiagram content="graph TD\nA-->B" />);
      
      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledTimes(1);
      });

      const firstCallId = mockMermaidRender.mock.calls[0][0];

      rerender(<MermaidDiagram content="graph LR\nX-->Y" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledTimes(2);
      });

      const secondCallId = mockMermaidRender.mock.calls[1][0];
      expect(firstCallId).not.toBe(secondCallId);
    });

    it('processes SVG for responsive design', async () => {
      const svgWithDimensions = '<svg width="800" height="600"><g><text>Content</text></g></svg>';
      mockMermaidRender.mockResolvedValue({ svg: svgWithDimensions });

      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        const svgContainer = screen.getByText('Content').closest('div');
        expect(svgContainer?.innerHTML).toContain('class="w-full h-auto max-w-full"');
        expect(svgContainer?.innerHTML).not.toContain('width="800"');
        expect(svgContainer?.innerHTML).not.toContain('height="600"');
      });
    });

    it('adds accessibility attributes to rendered SVG', async () => {
      render(<MermaidDiagram content={mockContent} alt="Custom diagram description" />);

      await waitFor(() => {
        const svgContainer = screen.getByText('Test Content').closest('div');
        expect(svgContainer?.innerHTML).toContain('role="img"');
        expect(svgContainer?.innerHTML).toContain('aria-labelledby=');
        expect(svgContainer?.innerHTML).toContain('Custom diagram description');
      });
    });
  });

  describe('Theme Configuration', () => {
    it('applies comprehensive light theme settings', async () => {
      render(<MermaidDiagram content={mockContent} theme="light" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({
            theme: 'default',
            themeVariables: expect.objectContaining({
              primaryColor: '#1E40AF',
              primaryTextColor: '#1F2937',
              background: '#FFFFFF',
              mainBkg: '#FFFFFF',
              secondBkg: '#F9FAFB',
              tertiaryBkg: '#F3F4F6',
            }),
            fontFamily: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
            fontSize: 14,
            securityLevel: 'strict',
            htmlLabels: false,
          })
        );
      });
    });

    it('applies comprehensive dark theme settings', async () => {
      render(<MermaidDiagram content={mockContent} theme="dark" />);

      await waitFor(() => {
        expect(mockMermaidInitialize).toHaveBeenCalledWith(
          expect.objectContaining({
            theme: 'dark',
            themeVariables: expect.objectContaining({
              primaryColor: '#3B82F6',
              primaryTextColor: '#F3F4F6',
              background: '#1F2937',
              mainBkg: '#1F2937',
              secondBkg: '#374151',
              tertiaryBkg: '#4B5563',
            }),
          })
        );
      });
    });

    it('configures diagram-specific settings for all types', async () => {
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
              actorMargin: 50,
            }),
            gantt: expect.objectContaining({
              useMaxWidth: true,
              leftPadding: 75,
              gridLineStartPadding: 35,
            }),
            er: expect.objectContaining({
              useMaxWidth: true,
              diagramPadding: 20,
              minEntityWidth: 100,
            }),
          })
        );
      });
    });
  });

  describe('Error Handling and Recovery', () => {
    it('handles mermaid parsing errors gracefully', async () => {
      mockMermaidRender.mockRejectedValue(new Error('Invalid diagram syntax'));

      render(<MermaidDiagram content="invalid syntax" />);

      await waitFor(() => {
        expect(screen.getByText('Diagram Error:')).toBeInTheDocument();
        expect(screen.getByText('Invalid diagram syntax')).toBeInTheDocument();
      });
    });

    it('shows expandable raw content in error state', async () => {
      mockMermaidRender.mockRejectedValue(new Error('Parse error'));

      render(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByText('Show raw content')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Show raw content'));
      expect(screen.getByText(mockContent)).toBeInTheDocument();
    });

    it('handles timeout errors appropriately', async () => {
      mockMermaidRender.mockImplementation(() => 
        new Promise(resolve => setTimeout(resolve, 15000))
      );

      render(<MermaidDiagram content={mockContent} />);

      vi.advanceTimersByTime(10000);

      await waitFor(() => {
        expect(screen.getByText('Diagram rendering timeout')).toBeInTheDocument();
      });
    });

    it('calls error callback with proper error details', async () => {
      const onRenderError = vi.fn();
      mockMermaidRender.mockRejectedValue(new Error('Custom error'));

      render(
        <MermaidDiagram 
          content={mockContent} 
          onRenderError={onRenderError}
        />
      );

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith('Custom error');
      });
    });

    it('recovers from errors when content is updated', async () => {
      mockMermaidRender.mockRejectedValueOnce(new Error('First error'));
      mockMermaidRender.mockResolvedValueOnce({ svg: mockSvg });

      const { rerender } = render(<MermaidDiagram content="invalid" />);

      await waitFor(() => {
        expect(screen.getByText('First error')).toBeInTheDocument();
      });

      rerender(<MermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByText('Test Content')).toBeInTheDocument();
        expect(screen.queryByText('First error')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading States and Performance', () => {
    it('shows loading indicator with proper accessibility', () => {
      mockMermaidRender.mockImplementation(() => new Promise(() => {}));

      render(<MermaidDiagram content={mockContent} />);

      const loadingIndicator = screen.getByText('Rendering diagram...');
      expect(loadingIndicator).toBeInTheDocument();
      
      const spinner = loadingIndicator.parentElement?.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });

    it('transitions from loading to rendered state smoothly', async () => {
      render(<MermaidDiagram content={mockContent} />);

      expect(screen.getByText('Rendering diagram...')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByText('Rendering diagram...')).not.toBeInTheDocument();
        expect(screen.getByText('Test Content')).toBeInTheDocument();
      });
    });

    it('handles rapid content changes efficiently', async () => {
      const { rerender } = render(<MermaidDiagram content="graph TD\nA-->B" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledTimes(1);
      });

      // Rapid content changes
      rerender(<MermaidDiagram content="graph LR\nX-->Y" />);
      rerender(<MermaidDiagram content="graph TB\nP-->Q" />);
      rerender(<MermaidDiagram content="graph RL\nM-->N" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledTimes(4);
      });

      // Should render the final content
      expect(mockMermaidRender).toHaveBeenLastCalledWith(
        expect.any(String),
        'graph RL\nM-->N'
      );
    });
  });

  describe('Callback Integration', () => {
    it('calls onRenderSuccess with processed SVG', async () => {
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

    it('provides both success and error callbacks in workflow', async () => {
      const onRenderSuccess = vi.fn();
      const onRenderError = vi.fn();

      mockMermaidRender.mockRejectedValueOnce(new Error('First error'));
      mockMermaidRender.mockResolvedValueOnce({ svg: mockSvg });

      const { rerender } = render(
        <MermaidDiagram 
          content="invalid" 
          onRenderSuccess={onRenderSuccess}
          onRenderError={onRenderError}
        />
      );

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith('First error');
      });

      rerender(
        <MermaidDiagram 
          content={mockContent} 
          onRenderSuccess={onRenderSuccess}
          onRenderError={onRenderError}
        />
      );

      await waitFor(() => {
        expect(onRenderSuccess).toHaveBeenCalled();
      });
    });
  });

  describe('Edge Cases and Boundary Conditions', () => {
    it('handles empty content gracefully', () => {
      render(<MermaidDiagram content="" />);
      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });

    it('handles whitespace-only content', () => {
      render(<MermaidDiagram content="   \n\t  " />);
      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });

    it('handles very large diagram content', async () => {
      const largeContent = 'graph TD\n' + 
        Array.from({ length: 1000 }, (_, i) => `A${i}-->B${i}`).join('\n');

      render(<MermaidDiagram content={largeContent} />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledWith(
          expect.any(String),
          largeContent
        );
      });
    });

    it('handles special characters in content', async () => {
      const specialContent = 'graph TD\nA["Special & <chars>"]-->B["More & <special>"]';

      render(<MermaidDiagram content={specialContent} />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalledWith(
          expect.any(String),
          specialContent
        );
      });
    });

    it('handles component unmounting during render', () => {
      mockMermaidRender.mockImplementation(() => 
        new Promise(resolve => setTimeout(() => resolve({ svg: mockSvg }), 1000))
      );

      const { unmount } = render(<MermaidDiagram content={mockContent} />);

      // Unmount before render completes
      unmount();

      // Should not throw errors
      vi.advanceTimersByTime(1000);
    });
  });
});

describe('SafeMermaidDiagram - Error Boundary Tests', () => {
  const originalConsoleError = console.error;

  beforeEach(() => {
    console.error = vi.fn();
  });

  afterEach(() => {
    console.error = originalConsoleError;
  });

  it('catches component errors and shows fallback UI', () => {
    const ThrowingComponent = () => {
      throw new Error('Component crashed');
    };

    // Mock the MermaidDiagram to throw
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

    render(<SafeMermaidDiagram content="graph TD\nA-->B" />);

    expect(screen.getByText('Failed to render diagram')).toBeInTheDocument();
    expect(screen.getByText('Show raw content')).toBeInTheDocument();
  });

  it('logs error details for debugging', () => {
    const ThrowingComponent = () => {
      throw new Error('Test error for logging');
    };

    vi.doMock('../MermaidDiagram', () => ({
      MermaidDiagram: ThrowingComponent,
      SafeMermaidDiagram: ({ content }: any) => (
        <div>Error boundary fallback</div>
      ),
    }));

    render(<SafeMermaidDiagram content="test content" />);

    expect(console.error).toHaveBeenCalled();
  });

  it('passes through props when no error occurs', async () => {
    mockMermaidRender.mockResolvedValue({ 
      svg: '<svg><text>Normal render</text></svg>' 
    });

    render(<SafeMermaidDiagram content="graph TD\nA-->B" className="test-class" />);

    await waitFor(() => {
      expect(screen.getByText('Normal render')).toBeInTheDocument();
    });

    const container = screen.getByText('Normal render').closest('.mermaid-diagram');
    expect(container).toHaveClass('test-class');
  });
});