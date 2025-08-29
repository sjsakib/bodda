import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import ResponsiveMermaidDiagram from '../ResponsiveMermaidDiagram';

// Mock dependencies
const mockUseResponsiveLayout = vi.fn();
const mockUseDiagramLibrary = vi.fn();
const mockMermaidRender = vi.fn();

vi.mock('../hooks/useResponsiveLayout', () => ({
  useResponsiveLayout: () => mockUseResponsiveLayout(),
}));

vi.mock('../contexts/DiagramLibraryContext', () => ({
  useDiagramLibrary: () => mockUseDiagramLibrary(),
}));

vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: mockMermaidRender,
  },
}));

// Mock child components
vi.mock('./DiagramLoadingIndicator', () => ({
  DiagramLoadingIndicator: ({ type, message, className }: any) => (
    <div data-testid="loading-indicator" className={className}>
      {type}: {message}
    </div>
  ),
  DiagramErrorIndicator: ({ error, type, className }: any) => (
    <div data-testid="error-indicator" className={className}>
      {type}: {error}
    </div>
  ),
}));

vi.mock('./ResponsiveDiagramContainer', () => ({
  default: ({ children, className }: any) => (
    <div data-testid="responsive-container" className={className}>
      {children}
    </div>
  ),
}));

describe('ResponsiveMermaidDiagram', () => {
  const mockSvg = '<svg><g><text>Test Diagram</text></g></svg>';
  const mockContent = 'graph TD\nA-->B';

  beforeEach(() => {
    // Reset mocks
    mockUseResponsiveLayout.mockReturnValue({
      isMobile: false,
    });

    mockUseDiagramLibrary.mockReturnValue({
      libraryState: {
        mermaid: { loaded: true, loading: false, error: null },
      },
      loadMermaid: vi.fn(),
    });

    mockMermaidRender.mockResolvedValue({
      svg: mockSvg,
    });

    // Mock window.matchMedia for theme detection
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(query => ({
        matches: query.includes('dark') ? false : true,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Desktop Rendering', () => {
    it('renders successfully with valid content', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
        expect(screen.getByText('Test Diagram')).toBeInTheDocument();
      });

      expect(mockMermaidRender).toHaveBeenCalledWith(
        expect.stringMatching(/^mermaid-/),
        mockContent
      );
    });

    it('applies desktop-optimized configuration', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalled();
      });

      // Check that mermaid.initialize was called with desktop config
      const mermaid = await import('mermaid');
      expect(mermaid.default.initialize).toHaveBeenCalledWith(
        expect.objectContaining({
          fontSize: 14, // Desktop font size
          flowchart: expect.objectContaining({
            padding: 20, // Desktop padding
          }),
        })
      );
    });

    it('handles light theme correctly', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} theme="light" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalled();
      });

      const mermaid = await import('mermaid');
      expect(mermaid.default.initialize).toHaveBeenCalledWith(
        expect.objectContaining({
          theme: 'default',
          themeVariables: expect.objectContaining({
            primaryColor: '#1E40AF',
            primaryTextColor: '#1F2937',
          }),
        })
      );
    });

    it('handles dark theme correctly', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} theme="dark" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalled();
      });

      const mermaid = await import('mermaid');
      expect(mermaid.default.initialize).toHaveBeenCalledWith(
        expect.objectContaining({
          theme: 'dark',
          themeVariables: expect.objectContaining({
            primaryColor: '#3B82F6',
            primaryTextColor: '#F3F4F6',
          }),
        })
      );
    });

    it('auto-detects theme from system preference', async () => {
      // Mock dark mode preference
      window.matchMedia = vi.fn().mockImplementation(query => ({
        matches: query.includes('dark') ? true : false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));

      render(<ResponsiveMermaidDiagram content={mockContent} theme="auto" />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalled();
      });

      const mermaid = await import('mermaid');
      expect(mermaid.default.initialize).toHaveBeenCalledWith(
        expect.objectContaining({
          theme: 'dark',
        })
      );
    });
  });

  describe('Mobile Rendering', () => {
    beforeEach(() => {
      mockUseResponsiveLayout.mockReturnValue({
        isMobile: true,
      });
    });

    it('applies mobile-optimized configuration', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(mockMermaidRender).toHaveBeenCalled();
      });

      const mermaid = await import('mermaid');
      expect(mermaid.default.initialize).toHaveBeenCalledWith(
        expect.objectContaining({
          fontSize: 12, // Mobile font size
          flowchart: expect.objectContaining({
            padding: 10, // Mobile padding
          }),
          sequence: expect.objectContaining({
            diagramMarginX: 20, // Mobile margins
            width: 120, // Mobile width
          }),
        })
      );
    });

    it('renders with mobile-optimized container', async () => {
      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
      });

      // Should pass mobile-specific props to container
      expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
    });
  });

  describe('Loading States', () => {
    it('shows loading indicator when library is loading', () => {
      mockUseDiagramLibrary.mockReturnValue({
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
        },
        loadMermaid: vi.fn(),
      });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      expect(screen.getByTestId('loading-indicator')).toBeInTheDocument();
      expect(screen.getByText('mermaid: Loading Mermaid library...')).toBeInTheDocument();
    });

    it('shows loading indicator when rendering', async () => {
      // Make render promise hang
      mockMermaidRender.mockImplementation(() => new Promise(() => {}));

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByTestId('loading-indicator')).toBeInTheDocument();
        expect(screen.getByText('mermaid: Rendering diagram...')).toBeInTheDocument();
      });
    });

    it('uses large loading indicator on mobile', () => {
      mockUseResponsiveLayout.mockReturnValue({ isMobile: true });
      mockUseDiagramLibrary.mockReturnValue({
        libraryState: {
          mermaid: { loaded: false, loading: true, error: null },
        },
        loadMermaid: vi.fn(),
      });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      const loadingIndicator = screen.getByTestId('loading-indicator');
      expect(loadingIndicator).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('shows library error', () => {
      mockUseDiagramLibrary.mockReturnValue({
        libraryState: {
          mermaid: { loaded: false, loading: false, error: 'Failed to load library' },
        },
        loadMermaid: vi.fn(),
      });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      expect(screen.getByTestId('error-indicator')).toBeInTheDocument();
      expect(screen.getByText('mermaid: Library Error: Failed to load library')).toBeInTheDocument();
    });

    it('shows rendering error', async () => {
      mockMermaidRender.mockRejectedValue(new Error('Invalid syntax'));

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByTestId('error-indicator')).toBeInTheDocument();
        expect(screen.getByText('mermaid: Invalid syntax')).toBeInTheDocument();
      });
    });

    it('handles timeout error', async () => {
      // Mock a long-running render
      mockMermaidRender.mockImplementation(() => 
        new Promise(resolve => setTimeout(resolve, 20000))
      );

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        expect(screen.getByTestId('error-indicator')).toBeInTheDocument();
        expect(screen.getByText('mermaid: Diagram rendering timeout')).toBeInTheDocument();
      }, { timeout: 16000 });
    });

    it('calls onRenderError callback', async () => {
      const onRenderError = vi.fn();
      mockMermaidRender.mockRejectedValue(new Error('Test error'));

      render(
        <ResponsiveMermaidDiagram 
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
        <ResponsiveMermaidDiagram 
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

  describe('Responsive SVG Processing', () => {
    it('makes SVG responsive by removing fixed dimensions', async () => {
      const svgWithDimensions = '<svg width="500" height="300"><g><text>Test</text></g></svg>';
      mockMermaidRender.mockResolvedValue({ svg: svgWithDimensions });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      await waitFor(() => {
        const svgElement = screen.getByText('Test').closest('div');
        expect(svgElement?.innerHTML).toContain('class="w-full h-auto max-w-full"');
        expect(svgElement?.innerHTML).not.toContain('width="500"');
        expect(svgElement?.innerHTML).not.toContain('height="300"');
      });
    });
  });

  describe('Empty States', () => {
    it('shows empty state for no content', () => {
      render(<ResponsiveMermaidDiagram content="" />);

      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });

    it('shows empty state for whitespace-only content', () => {
      render(<ResponsiveMermaidDiagram content="   \n  \t  " />);

      expect(screen.getByText('No diagram content to display')).toBeInTheDocument();
    });
  });

  describe('Library Loading', () => {
    it('triggers library loading when not loaded', () => {
      const mockLoadMermaid = vi.fn();
      mockUseDiagramLibrary.mockReturnValue({
        libraryState: {
          mermaid: { loaded: false, loading: false, error: null },
        },
        loadMermaid: mockLoadMermaid,
      });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      expect(mockLoadMermaid).toHaveBeenCalled();
    });

    it('does not trigger loading when already loaded', () => {
      const mockLoadMermaid = vi.fn();
      mockUseDiagramLibrary.mockReturnValue({
        libraryState: {
          mermaid: { loaded: true, loading: false, error: null },
        },
        loadMermaid: mockLoadMermaid,
      });

      render(<ResponsiveMermaidDiagram content={mockContent} />);

      expect(mockLoadMermaid).not.toHaveBeenCalled();
    });
  });

  describe('Props Forwarding', () => {
    it('forwards props to ResponsiveDiagramContainer', async () => {
      render(
        <ResponsiveMermaidDiagram 
          content={mockContent}
          className="custom-class"
          enableZoomPan={false}
          enableTouchGestures={false}
          fitToContainer={false}
        />
      );

      await waitFor(() => {
        const container = screen.getByTestId('responsive-container');
        expect(container).toHaveClass('custom-class');
      });
    });
  });
});