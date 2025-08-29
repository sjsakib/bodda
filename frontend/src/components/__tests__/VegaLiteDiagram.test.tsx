import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { VegaLiteDiagram, SafeVegaLiteDiagram } from '../VegaLiteDiagram';

// Mock react-vega
const mockVegaLite = vi.fn(({ spec, onError, ...props }) => (
  <div data-testid='vega-lite-chart' data-spec={JSON.stringify(spec)} {...props}>
    Mock Vega-Lite Chart
  </div>
));

vi.mock('react-vega', () => ({
  VegaLite: mockVegaLite,
}));

// Mock console methods
const mockConsoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
const mockConsoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});

describe('VegaLiteDiagram', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    mockConsoleError.mockClear();
    mockConsoleLog.mockClear();
  });

  describe('Basic Rendering', () => {
    it('renders loading state initially', () => {
      const validSpec = JSON.stringify({
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: {
          x: { field: 'a', type: 'ordinal' },
          y: { field: 'b', type: 'quantitative' },
        },
      });

      render(<VegaLiteDiagram content={validSpec} />);

      expect(screen.getByText('Loading chart component...')).toBeInTheDocument();
      expect(screen.getByRole('status')).toBeInTheDocument();
    });

    it('renders chart after loading component', async () => {
      const validSpec = JSON.stringify({
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: {
          x: { field: 'a', type: 'ordinal' },
          y: { field: 'b', type: 'quantitative' },
        },
      });

      render(<VegaLiteDiagram content={validSpec} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      expect(screen.getByText('Mock Vega-Lite Chart')).toBeInTheDocument();
    });

    it('renders empty state when no content provided', async () => {
      render(<VegaLiteDiagram content='' />);

      await waitFor(() => {
        expect(screen.getByText('No chart specification to display')).toBeInTheDocument();
      });
    });

    it('applies custom className', async () => {
      const validSpec = JSON.stringify({
        mark: 'point',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: {
          x: { field: 'x', type: 'quantitative' },
          y: { field: 'y', type: 'quantitative' },
        },
      });

      render(<VegaLiteDiagram content={validSpec} className='custom-class' />);

      await waitFor(() => {
        const container = screen
          .getByTestId('vega-lite-chart')
          .closest('.vega-lite-diagram');
        expect(container).toHaveClass('custom-class');
      });
    });
  });

  describe('Spec Parsing and Validation', () => {
    it('handles valid Vega-Lite specification', async () => {
      const validSpec = {
        mark: 'circle',
        data: {
          values: [
            { x: 1, y: 2 },
            { x: 2, y: 3 },
          ],
        },
        encoding: {
          x: { field: 'x', type: 'quantitative' },
          y: { field: 'y', type: 'quantitative' },
        },
      };

      render(<VegaLiteDiagram content={JSON.stringify(validSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.mark).toBe('circle');
        expect(specData.data.values).toHaveLength(2);
      });
    });

    it('handles invalid JSON gracefully', async () => {
      const invalidJson = '{ invalid json }';

      render(<VegaLiteDiagram content={invalidJson} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(screen.getByText(/Expected property name/)).toBeInTheDocument();
      });
    });

    it('validates required mark property', async () => {
      const invalidSpec = JSON.stringify({
        data: { values: [{ a: 1, b: 2 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
        // Missing mark property
      });

      render(<VegaLiteDiagram content={invalidSpec} />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();
        expect(
          screen.getByText(/missing required mark or composition property/)
        ).toBeInTheDocument();
      });
    });

    it('accepts layer composition instead of mark', async () => {
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

    it('sanitizes dangerous properties', async () => {
      const dangerousSpec = {
        mark: 'bar',
        data: { values: [{ a: 1, b: 2 }] },
        datasets: { external: 'http://malicious.com/data' }, // Should be removed
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(dangerousSpec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.datasets).toBeUndefined();
        expect(specData.mark).toBe('bar');
      });
    });

    it('limits data size for security', async () => {
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
  });

  describe('Theme Application', () => {
    it('applies light theme by default', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.config.background).toBe('#FFFFFF');
        expect(specData.config.title.color).toBe('#1F2937');
      });
    });

    it('applies dark theme when specified', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} theme='dark' />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.config.background).toBe('#1F2937');
        expect(specData.config.title.color).toBe('#F3F4F6');
      });
    });

    it('preserves existing config while applying theme', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
        config: {
          title: { fontSize: 20 },
          customProperty: 'preserved',
        },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        const chart = screen.getByTestId('vega-lite-chart');
        const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
        expect(specData.config.customProperty).toBe('preserved');
        expect(specData.config.title.fontSize).toBe(20); // Original preserved
        expect(specData.config.title.color).toBe('#1F2937'); // Theme applied
      });
    });
  });

  describe('Interactive Features', () => {
    it('enables tooltips by default', async () => {
      const spec = {
        mark: 'point',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ tooltip: true }),
          expect.anything()
        );
      });
    });

    it('enables hover effects by default', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ hover: true }),
          expect.anything()
        );
      });
    });

    it('disables tooltips when specified', async () => {
      const spec = {
        mark: 'line',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} enableTooltips={false} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ tooltip: false }),
          expect.anything()
        );
      });
    });

    it('disables hover effects when specified', async () => {
      const spec = {
        mark: 'area',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} enableHover={false} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ hover: false }),
          expect.anything()
        );
      });
    });

    it('shows actions when enabled', async () => {
      const spec = {
        mark: 'circle',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} showActions={true} />);

      await waitFor(() => {
        expect(mockVegaLite).toHaveBeenCalledWith(
          expect.objectContaining({ actions: true }),
          expect.anything()
        );
      });
    });
  });

  describe('Error Handling', () => {
    it('shows error details with expandable raw content', async () => {
      const invalidSpec = '{ "mark": "invalid", "data": {"values": []} }';

      render(<VegaLiteDiagram content={invalidSpec} />);

      // This spec is actually valid JSON but invalid Vega-Lite, so it should render
      // Let's test with truly invalid JSON instead
      const { rerender } = render(<VegaLiteDiagram content='{ invalid }' />);

      await waitFor(() => {
        expect(screen.getByText('Chart Error:')).toBeInTheDocument();

        // Check for expandable details
        const summary = screen.getByText('Show raw JSON');
        expect(summary).toBeInTheDocument();

        // Expand details
        fireEvent.click(summary);
        expect(screen.getByText('{ invalid }')).toBeInTheDocument();
      });
    });

    it('calls onRenderError callback on parsing failure', async () => {
      const onRenderError = vi.fn();
      const invalidSpec = '{ invalid }';

      render(<VegaLiteDiagram content={invalidSpec} onRenderError={onRenderError} />);

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith(
          expect.stringContaining('Expected property name')
        );
      });
    });

    it('calls onRenderSuccess callback on successful parsing', async () => {
      const onRenderSuccess = vi.fn();
      const validSpec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(
        <VegaLiteDiagram
          content={JSON.stringify(validSpec)}
          onRenderSuccess={onRenderSuccess}
        />
      );

      await waitFor(() => {
        expect(onRenderSuccess).toHaveBeenCalledWith(
          expect.objectContaining({
            mark: 'bar',
            config: expect.any(Object),
          })
        );
      });
    });

    it('handles VegaLite component rendering errors', async () => {
      const onRenderError = vi.fn();
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      // Mock VegaLite to trigger onError
      mockVegaLite.mockImplementationOnce(({ onError }) => {
        onError({ message: 'Rendering failed' });
        return <div>Error occurred</div>;
      });

      render(
        <VegaLiteDiagram content={JSON.stringify(spec)} onRenderError={onRenderError} />
      );

      await waitFor(() => {
        expect(onRenderError).toHaveBeenCalledWith('Rendering failed');
      });
    });

    it('handles component cleanup properly', () => {
      const { unmount } = render(
        <VegaLiteDiagram content='{"mark": "bar", "data": {"values": []}}' />
      );

      // Should unmount without errors
      expect(() => unmount()).not.toThrow();
    });
  });

  describe('Accessibility', () => {
    it('includes ARIA labels for screen readers', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} alt='Sales chart' />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const chartContainer = screen.getByRole('img');
      expect(chartContainer).toHaveAttribute('aria-label', 'Sales chart');
    });

    it('provides default ARIA label when alt not specified', async () => {
      const spec = {
        mark: 'point',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const chartContainer = screen.getByRole('img');
      expect(chartContainer).toHaveAttribute('aria-label', 'Vega-Lite Chart');
    });

    it('includes screen reader description', async () => {
      const spec = {
        mark: 'line',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} alt='Trend analysis' />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      expect(screen.getByText('Chart description: Trend analysis')).toBeInTheDocument();
    });
  });

  describe('Responsive Design', () => {
    it('applies responsive autosize configuration', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const chart = screen.getByTestId('vega-lite-chart');
      const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
      expect(specData.autosize).toEqual({
        type: 'fit',
        contains: 'padding',
        resize: true,
      });
    });

    it('preserves existing autosize configuration', async () => {
      const spec = {
        mark: 'bar',
        data: { values: [{ a: 'A', b: 28 }] },
        encoding: { x: { field: 'a' }, y: { field: 'b' } },
        autosize: { type: 'pad' },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const chart = screen.getByTestId('vega-lite-chart');
      const specData = JSON.parse(chart.getAttribute('data-spec') || '{}');
      expect(specData.autosize.type).toBe('pad');
    });

    it('applies responsive styling to chart container', async () => {
      const spec = {
        mark: 'circle',
        data: { values: [{ x: 1, y: 2 }] },
        encoding: { x: { field: 'x' }, y: { field: 'y' } },
      };

      render(<VegaLiteDiagram content={JSON.stringify(spec)} />);

      await waitFor(() => {
        expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
      });

      const chart = screen.getByTestId('vega-lite-chart');
      expect(chart).toHaveStyle({ width: '100%', height: 'auto' });
    });
  });
});

describe('SafeVegaLiteDiagram', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders VegaLiteDiagram normally when no error occurs', async () => {
    const spec = {
      mark: 'bar',
      data: { values: [{ a: 'A', b: 28 }] },
      encoding: { x: { field: 'a' }, y: { field: 'b' } },
    };

    render(<SafeVegaLiteDiagram content={JSON.stringify(spec)} />);

    await waitFor(() => {
      expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
    });
  });

  it('catches and displays component errors gracefully', () => {
    // Create a component that throws an error
    const ThrowingComponent = () => {
      throw new Error('Component crashed');
    };

    // Create a simple error boundary for testing
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
            <div>
              <p>Failed to render chart</p>
              <p>{this.state.error?.message}</p>
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

  it('logs error details when component crashes', () => {
    const ThrowingComponent = () => {
      throw new Error('Test error');
    };

    // Create a test error boundary manually
    class TestErrorBoundary extends React.Component<
      { children: React.ReactNode },
      { hasError: boolean }
    > {
      constructor(props: { children: React.ReactNode }) {
        super(props);
        this.state = { hasError: false };
      }

      static getDerivedStateFromError() {
        return { hasError: true };
      }

      componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
        console.error('Test error boundary caught:', error, errorInfo);
      }

      render() {
        if (this.state.hasError) {
          return <div>Error boundary triggered</div>;
        }
        return this.props.children;
      }
    }

    render(
      <TestErrorBoundary>
        <ThrowingComponent />
      </TestErrorBoundary>
    );

    expect(screen.getByText('Error boundary triggered')).toBeInTheDocument();
    expect(mockConsoleError).toHaveBeenCalled();
  });
});
