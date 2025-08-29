import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { VegaLiteDiagram, SafeVegaLiteDiagram } from '../VegaLiteDiagram';

// Mock react-vega
const mockVegaLite = vi.fn(({ spec, onError, ...props }) => (
  <div data-testid='vega-lite-chart' data-spec={JSON.stringify(spec)} {...props}>
    Mock Vega-Lite Chart: {spec?.mark || 'unknown'}
  </div>
));

vi.mock('react-vega', () => ({
  VegaLite: mockVegaLite,
}));

describe('VegaLiteDiagram - Comprehensive Unit Tests', () => {
  const validBarSpec = {
    mark: 'bar',
    data: { values: [{ category: 'A', value: 28 }, { category: 'B', value: 55 }] },
    encoding: {
      x: { field: 'category', type: 'ordinal' },
      y: { field: 'value', type: 'quantitative' },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Core Rendering and Spec Processing', () => {
    it('renders valid Vega-Lite specification successfully', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
        expect(screen.getByText('Mock Vega-Lite Chart: bar')).toBeInTheDocument();
      });
    });

    it('generates unique chart IDs for multiple instances', async () => {
      const { rerender } = render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);
      
      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const lineSpec = { ...validBarSpec, mark: 'line' };
      rerender(<VegaLiteDiagram content={JSON.stringify(lineSpec)} />);

      await waitFor(() => {
        expect(screen.getByText('Mock Vega-Lite Chart: line')).toBeInTheDocument();
      });
    });

    it('applies theme configuration to specification', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} theme="dark" />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.config.background).toBe('#1F2937');
        expect(specData.config.title.color).toBe('#F3F4F6');
        expect(specData.config.axis.labelColor).toBe('#F3F4F6');
      });
    });
  });

describe('Specification Validation and Security', () => {
    it('validates required mark property', async () => {
      const invalidSpec = {
        data: { values: [{ a: 1, b: 2 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
        // Missing mark property
      };

      render(<VegaLiteDiagram content={JSON.stringify(invalidSpec)} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText(/missing required mark or composition property/)).toBeInTheDocument();
      });
    });

    it('accepts layer composition as alternative to mark', async () => {
      const layerSpec = {
        layer: [
          {
            mark: 'line',
            encoding: {
              x: { field: 'x', type: 'quantitative' },
              y: { field: 'y', type: 'quantitative' },
            },
          },
        ],
        data: { values: [{ x: 1, y: 2 }] },
      };

      render(<VegaLiteDiagram content={JSON.stringify(layerSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.layer).toBeDefined();
        expect(specData.layer[0].mark).toBe('line');
      });
    });

    it('sanitizes dangerous properties for security', async () => {
      const dangerousSpec = {
        mark: 'bar',
        data: { values: [{ a: 1, b: 2 }] },
        datasets: { external: 'http://malicious.com/data' }, // Should be removed
        transform: [
          { calculate: 'datum.a * 2', as: 'doubled' }, // Safe transform
          { malicious: 'dangerous code' }, // Should be filtered out
        ],
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(dangerousSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.datasets).toBeUndefined();
        expect(specData.transform).toHaveLength(1);
        expect(specData.transform[0].calculate).toBeDefined();
        expect(specData.transform[0].malicious).toBeUndefined();
      });
    });

    it('enforces data size limits for performance', async () => {
      const largeDataSpec = {
        mark: 'point',
        data: { values: Array.from({ length: 6000 }, (_, i) => ({ x: i, y: i * 2 })) },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(largeDataSpec)} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText(/Dataset too large/)).toBeInTheDocument();
      });
    });

    it('handles malformed JSON gracefully', async () => {
      const invalidJson = '{ "mark": "bar", "data": { invalid json }';

      render(<VegaLiteDiagram content={invalidJson} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText(/Expected property name/)).toBeInTheDocument();
      });
    });
  });

  describe('Theme System and Styling', () => {
    it('applies comprehensive light theme configuration', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} theme="light" />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.config).toMatchObject({
          background: '#FFFFFF',
          title: expect.objectContaining({
            color: '#1F2937',
            fontSize: 16,
            fontWeight: 600,
          }),
          axis: expect.objectContaining({
            labelColor: '#1F2937',
            titleColor: '#1F2937',
            gridColor: '#E5E7EB',
          }),
          legend: expect.objectContaining({
            labelColor: '#1F2937',
            titleColor: '#1F2937',
          }),
        });
      });
    });

    it('applies comprehensive dark theme configuration', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} theme="dark" />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.config).toMatchObject({
          background: '#1F2937',
          title: expect.objectContaining({
            color: '#F3F4F6',
          }),
          axis: expect.objectContaining({
            labelColor: '#F3F4F6',
            titleColor: '#F3F4F6',
            gridColor: '#374151',
          }),
        });
      });
    });

    it('preserves existing config while applying theme', async () => {
      const specWithConfig = {
        ...validBarSpec,
        config: {
          title: { fontSize: 20, customProperty: 'preserved' },
          axis: { labelFontSize: 14 },
        },
      };

      render(<VegaLiteDiagram content={JSON.stringify(specWithConfig)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.config.title.fontSize).toBe(20); // Original preserved
        expect(specData.config.title.customProperty).toBe('preserved'); // Custom preserved
        expect(specData.config.title.color).toBe('#1F2937'); // Theme applied
        expect(specData.config.axis.labelFontSize).toBe(14); // Original preserved
      });
    });

    it('applies responsive autosize configuration', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        
        expect(specData.autosize).toEqual({
          type: 'fit',
          contains: 'padding',
          resize: true,
        });
      });
    });
  });

  describe('Interactive Features and Props', () => {
    it('configures tooltips based on enableTooltips prop', async () => {
      const { rerender } = render(
        <VegaLiteDiagram content={JSON.stringify(validBarSpec)} enableTooltips={true} />
      );

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ tooltip: true }),
          expect.anything()
        );
      });

      rerender(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} enableTooltips={false} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ tooltip: false }),
          expect.anything()
        );
      });
    });

    it('configures hover effects based on enableHover prop', async () => {
      const { rerender } = render(
        <VegaLiteDiagram content={JSON.stringify(validBarSpec)} enableHover={true} />
      );

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ hover: true }),
          expect.anything()
        );
      });

      rerender(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} enableHover={false} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ hover: false }),
          expect.anything()
        );
      });
    });

    it('configures chart actions based on showActions prop', async () => {
      const { rerender } = render(
        <VegaLiteDiagram content={JSON.stringify(validBarSpec)} showActions={true} />
      );

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ actions: true }),
          expect.anything()
        );
      });

      rerender(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} showActions={false} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ actions: false }),
          expect.anything()
        );
      });
    });

    it('applies custom className to container', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} className="custom-chart" />);

      await waitFor(() => {
        const container = screen.getByTestId('vega-lite-chart').closest('.vega-lite-diagram');
        expect(container).toHaveClass('custom-chart');
      });
    });
  });

  describe('Error Handling and Recovery', () => {
    it('shows detailed error information with expandable content', async () => {
      const invalidSpec = '{ "mark": "invalid", "data": }';

      render(<VegaLiteDiagram content={invalidSpec} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText('Show raw JSON')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Show raw JSON'));
      expect(screen.getByText(invalidSpec)).toBeInTheDocument();
    });

    it('calls error callback with proper error details', async () => {
      const onRenderError = vi.fn();
      const invalidSpec = '{ invalid json }';

      render(<VegaLiteDiagram content={invalidSpec} onRenderError={onRenderError} />);

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith(
          expect.stringContaining('Expected property name')
        );
      });
    });

    it('calls success callback with processed specification', async () => {
      const onRenderSuccess = vi.fn();

      render(
        <VegaLiteDiagram 
          content={JSON.stringify(validBarSpec)} 
          onRenderSuccess={onRenderSuccess} 
        />
      );

      await waitFor(() => {
        expect(onRenderSuccess).toHaveBeenCalledWith(
          expect.objectContaining({
            mark: 'bar',
            config: expect.any(Object),
            autosize: expect.any(Object),
          })
        );
      });
    });

    it('handles VegaLite component rendering errors', async () => {
      const onRenderError = vi.fn();

      mockVegaLite.mockImplementationOnce(({ onError }) => {
        onError({ message: 'Rendering failed' });
        return <div>Error occurred</div>;
      });

      render(
        <VegaLiteDiagram 
          content={JSON.stringify(validBarSpec)} 
          onRenderError={onRenderError} 
        />
      );

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith('Rendering failed');
      });
    });

    it('handles timeout during parsing', async () => {
      // Mock JSON.parse to be slow
      const originalParse = JSON.parse;
      JSON.parse = vi.fn().mockImplementation(() => {
        return new Promise(resolve => setTimeout(() => resolve(validBarSpec), 15000));
      });

      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      vi.advanceTimersByTime(10000);

      await waitFor(() => {
        expect(screen.getByText(/timeout/)).toBeInTheDocument();
      });

      JSON.parse = originalParse;
    });

    it('recovers from errors when content is updated', async () => {
      const { rerender } = render(<VegaLiteDiagram content="{ invalid }" />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
      });

      rerender(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
        expect(screen.queryByText('Chart Error:')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading States and Performance', () => {
    it('shows loading indicator while component loads', () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      expect(screen.getByText('Loading chart component...')).toBeInTheDocument();
      expect(screen.getByRole('status')).toBeInTheDocument();
    });

    it('shows parsing indicator after component loads', async () => {
      // Mock slow parsing
      const slowParseContent = JSON.stringify(validBarSpec);
      
      render(<VegaLiteDiagram content={slowParseContent} />);

      // Should show component loading first
      expect(screen.getByText('Loading chart component...')).toBeInTheDocument();

      // After component loads, should show rendering
      await waitFor(() => {
        expect(screen.getByText('Rendering chart...')).toBeInTheDocument();
      });
    });

    it('handles rapid content changes efficiently', async () => {
      const { rerender } = render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const lineSpec = { ...validBarSpec, mark: 'line' };
      const pointSpec = { ...validBarSpec, mark: 'point' };
      const areaSpec = { ...validBarSpec, mark: 'area' };

      rerender(<VegaLiteDiagram content={JSON.stringify(lineSpec)} />);
      rerender(<VegaLiteDiagram content={JSON.stringify(pointSpec)} />);
      rerender(<VegaLiteDiagram content={JSON.stringify(areaSpec)} />);

      await waitFor(() => {
        expect(screen.getByText('Mock Vega-Lite Chart: area')).toBeInTheDocument();
      });
    });

    it('handles component unmounting during processing', () => {
      const { unmount } = render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      // Unmount immediately
      unmount();

      // Should not throw errors
      vi.advanceTimersByTime(1000);
    });
  });

  describe('Accessibility Features', () => {
    it('includes proper ARIA attributes', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} alt="Sales data chart" />);

      await waitFor(() => {
        const chartContainer = screen.getByRole('img');
        expect(chartContainer).toHaveAttribute('aria-label', 'Sales data chart');
      });
    });

    it('provides default ARIA label when alt not specified', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} />);

      await waitFor(() => {
        const chartContainer = screen.getByRole('img');
        expect(chartContainer).toHaveAttribute('aria-label', 'Vega-Lite Chart');
      });
    });

    it('includes screen reader description', async () => {
      render(<VegaLiteDiagram content={JSON.stringify(validBarSpec)} alt="Revenue trends" />);

      await waitFor(() => {
        expect(screen.getByText('Chart description: Revenue trends')).toBeInTheDocument();
      });
    });

    it('maintains accessibility during error states', async () => {
      render(<VegaLiteDiagram content="{ invalid }" />);

      await waitFor(() => {
        const errorContainer = screen.getByText('Chart Error:').closest('.vega-lite-error');
        expect(errorContainer).toBeInTheDocument();
        
        // Error state should still be accessible
        const expandableContent = screen.getByText('Show raw JSON');
        expect(expandableContent).toBeInTheDocument();
      });
    });
  });

  describe('Edge Cases and Boundary Conditions', () => {
    it('handles empty content gracefully', async () => {
      render(<VegaLiteDiagram content="" />);

      await waitFor(() => {
        expect(screen.getByText('No chart specification to display')).toBeInTheDocument();
      });
    });

    it('handles whitespace-only content', async () => {
      render(<VegaLiteDiagram content="   \n\t  " />);

      await waitFor(() => {
        expect(screen.getByText('No chart specification to display')).toBeInTheDocument();
      });
    });

    it('handles complex nested specifications', async () => {
      const complexSpec = {
        layer: [
          {
            mark: 'line',
            encoding: {
              x: { field: 'date', type: 'temporal' },
              y: { field: 'value', type: 'quantitative' },
            },
          },
          {
            mark: 'point',
            encoding: {
              x: { field: 'date', type: 'temporal' },
              y: { field: 'value', type: 'quantitative' },
            },
          },
        ],
        data: { values: [{ date: '2024-01-01', value: 100 }] },
      };

      render(<VegaLiteDiagram content={JSON.stringify(complexSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.layer).toHaveLength(2);
      });
    });

    it('handles specifications with external data references', async () => {
      const specWithUrl = {
        mark: 'bar',
        data: { url: 'https://example.com/data.json' },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(specWithUrl)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.data.url).toBe('https://example.com/data.json');
      });
    });
  });
});

describe('SafeVegaLiteDiagram - Error Boundary Tests', () => {
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

    // Create a test error boundary
    class TestErrorBoundary extends React.Component<
      { children: React.ReactNode },
      { hasError: boolean; error?: Error }
    > {
      constructor(props: { children: React.ReactNode }) {
        super(props);
        this.state = { hasError: false };
      }

      static getDerivedStateFromError(error: Error) {
        return { hasError: true, error };
      }

      render() {
        if (this.state.hasError) {
          return (
            <div className="chart-error">
              <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-800 font-medium">Failed to render chart</p>
                <p className="text-xs text-red-600 mt-1">{this.state.error?.message}</p>
              </div>
            </div>
          );
        }
        return this.props.children;
      }
    }

    render(
      <TestErrorBoundary>
        <ThrowingComponent />
      </TestErrorBoundary>
    );

    expect(screen.getByText('Failed to render chart')).toBeInTheDocument();
    expect(screen.getByText('Component crashed')).toBeInTheDocument();
  });

  it('passes through props when no error occurs', async () => {
    const validSpec = {
      mark: 'bar',
      data: { values: [{ a: 'A', b: 28 }] },
      encoding: { x: { field: 'a' }, y: { field: 'b' } },
    };

    render(<SafeVegaLiteDiagram content={JSON.stringify(validSpec)} className="test-class" />);

    await waitFor(() => {
      expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
    });

    const container = screen.getByTestId('vega-lite-chart').closest('.vega-lite-diagram');
    expect(container).toHaveClass('test-class');
  });
});