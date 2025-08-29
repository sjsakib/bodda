import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import ResponsiveDiagramContainer from '../ResponsiveDiagramContainer';

// Mock the responsive layout hook
const mockUseResponsiveLayout = vi.fn();
vi.mock('../hooks/useResponsiveLayout', () => ({
  useResponsiveLayout: mockUseResponsiveLayout,
}));

// Mock child component
const MockDiagramContent = () => (
  <div data-testid="diagram-content" style={{ width: '200px', height: '150px' }}>
    Mock Diagram Content
  </div>
);

describe('ResponsiveDiagramContainer', () => {
  beforeEach(() => {
    // Reset mocks
    mockUseResponsiveLayout.mockReturnValue({
      isMobile: false,
    });

    // Mock getBoundingClientRect
    Element.prototype.getBoundingClientRect = vi.fn(() => ({
      width: 400,
      height: 300,
      top: 0,
      left: 0,
      bottom: 300,
      right: 400,
      x: 0,
      y: 0,
      toJSON: () => {},
    }));
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Desktop Rendering', () => {
    it('renders with zoom and pan controls by default', () => {
      render(
        <ResponsiveDiagramContainer>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      expect(screen.getByLabelText('Zoom in')).toBeInTheDocument();
      expect(screen.getByLabelText('Zoom out')).toBeInTheDocument();
      expect(screen.getByLabelText('Fit to container')).toBeInTheDocument();
      expect(screen.getByLabelText('Reset zoom')).toBeInTheDocument();
      expect(screen.getByText('100%')).toBeInTheDocument(); // Zoom indicator
    });

    it('renders diagram content', () => {
      render(
        <ResponsiveDiagramContainer>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      expect(screen.getByTestId('diagram-content')).toBeInTheDocument();
      expect(screen.getByText('Mock Diagram Content')).toBeInTheDocument();
    });

    it('applies correct desktop styling', () => {
      const { container } = render(
        <ResponsiveDiagramContainer className="custom-class">
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const containerElement = container.querySelector('.responsive-diagram-container');
      expect(containerElement).toBeInTheDocument();

      // Check that mobile instructions are not shown
      expect(screen.queryByText(/Pinch to zoom/)).not.toBeInTheDocument();
    });

    it('handles zoom in button click', async () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer onZoomChange={onZoomChange}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const zoomInButton = screen.getByLabelText('Zoom in');
      fireEvent.click(zoomInButton);

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalledWith(1.2);
        expect(screen.getByText('120%')).toBeInTheDocument();
      });
    });

    it('handles zoom out button click', async () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer onZoomChange={onZoomChange}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const zoomOutButton = screen.getByLabelText('Zoom out');
      fireEvent.click(zoomOutButton);

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalledWith(0.8);
        expect(screen.getByText('80%')).toBeInTheDocument();
      });
    });

    it('handles mouse wheel zoom', () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer onZoomChange={onZoomChange}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const container = screen.getByTestId('diagram-content').parentElement?.parentElement;
      expect(container).toBeInTheDocument();

      // Zoom in with wheel
      fireEvent.wheel(container!, { deltaY: -100 });
      expect(onZoomChange).toHaveBeenCalledWith(1.1);

      // Zoom out with wheel
      fireEvent.wheel(container!, { deltaY: 100 });
      expect(onZoomChange).toHaveBeenCalledWith(expect.closeTo(0.99, 2)); // 1.1 * 0.9
    });

    it('handles mouse drag for panning', () => {
      render(
        <ResponsiveDiagramContainer>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const container = screen.getByTestId('diagram-content').parentElement?.parentElement;
      expect(container).toBeInTheDocument();

      // Start drag
      fireEvent.mouseDown(container!, { clientX: 100, clientY: 100, button: 0 });
      
      // Move mouse
      fireEvent.mouseMove(container!, { clientX: 150, clientY: 120 });
      
      // End drag
      fireEvent.mouseUp(container!);

      // Verify transform was applied (content should have moved)
      const content = screen.getByTestId('diagram-content').parentElement;
      expect(content).toHaveStyle('transform: translate(50px, 20px) scale(1)');
    });
  });

  describe('Mobile Rendering', () => {
    beforeEach(() => {
      mockUseResponsiveLayout.mockReturnValue({
        isMobile: true,
      });
    });

    it('handles touch gestures for panning', () => {
      render(
        <ResponsiveDiagramContainer enableTouchGestures={true}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const container = screen.getByTestId('diagram-content').parentElement?.parentElement;
      expect(container).toBeInTheDocument();

      // Single touch pan
      fireEvent.touchStart(container!, {
        touches: [{ clientX: 100, clientY: 100 }],
      });

      fireEvent.touchMove(container!, {
        touches: [{ clientX: 150, clientY: 120 }],
      });

      fireEvent.touchEnd(container!);

      // Verify transform was applied
      const content = screen.getByTestId('diagram-content').parentElement;
      expect(content).toHaveStyle('transform: translate(50px, 20px) scale(1)');
    });

    it('handles pinch-to-zoom gestures', () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer 
          enableTouchGestures={true}
          onZoomChange={onZoomChange}
        >
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const container = screen.getByTestId('diagram-content').parentElement?.parentElement;
      expect(container).toBeInTheDocument();

      // Start pinch gesture
      fireEvent.touchStart(container!, {
        touches: [
          { clientX: 100, clientY: 100 },
          { clientX: 200, clientY: 100 },
        ],
      });

      // Pinch out (zoom in)
      fireEvent.touchMove(container!, {
        touches: [
          { clientX: 80, clientY: 100 },
          { clientX: 220, clientY: 100 },
        ],
      });

      fireEvent.touchEnd(container!);

      // Should have zoomed in
      expect(onZoomChange).toHaveBeenCalled();
    });
  });

  describe('Zoom Controls', () => {
    it('respects min and max zoom limits', async () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer 
          minZoom={0.5}
          maxZoom={2}
          onZoomChange={onZoomChange}
        >
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const zoomInButton = screen.getByLabelText('Zoom in');
      const zoomOutButton = screen.getByLabelText('Zoom out');

      // Zoom in multiple times to test max limit
      for (let i = 0; i < 10; i++) {
        fireEvent.click(zoomInButton);
      }

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalledWith(2); // Should be clamped to max
      });

      // Reset to test min zoom
      fireEvent.click(screen.getByLabelText('Reset zoom'));
      
      await waitFor(() => {
        expect(screen.getByText('100%')).toBeInTheDocument();
      });

      // Zoom out multiple times to test min limit
      for (let i = 0; i < 10; i++) {
        fireEvent.click(zoomOutButton);
      }

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalledWith(0.5); // Should be clamped to min
      });
    });

    it('handles fit to container', async () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer onZoomChange={onZoomChange}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const fitButton = screen.getByLabelText('Fit to container');
      fireEvent.click(fitButton);

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalled();
      });
    });

    it('handles reset zoom', async () => {
      const onZoomChange = vi.fn();
      
      render(
        <ResponsiveDiagramContainer 
          initialZoom={1.5}
          onZoomChange={onZoomChange}
        >
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      // Zoom should start at 150%
      expect(screen.getByText('150%')).toBeInTheDocument();

      const resetButton = screen.getByLabelText('Reset zoom');
      fireEvent.click(resetButton);

      await waitFor(() => {
        expect(onZoomChange).toHaveBeenCalledWith(1.5); // Reset to initial zoom
        expect(screen.getByText('150%')).toBeInTheDocument();
      });
    });
  });

  describe('Disabled Controls', () => {
    it('does not render controls when zoom/pan is disabled', () => {
      render(
        <ResponsiveDiagramContainer enableZoomPan={false}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      expect(screen.queryByLabelText('Zoom in')).not.toBeInTheDocument();
      expect(screen.queryByLabelText('Zoom out')).not.toBeInTheDocument();
      expect(screen.queryByText('100%')).not.toBeInTheDocument();
    });

    it('does not handle touch gestures when disabled', () => {
      mockUseResponsiveLayout.mockReturnValue({ isMobile: true });
      
      render(
        <ResponsiveDiagramContainer enableTouchGestures={false}>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const container = screen.getByTestId('diagram-content').parentElement?.parentElement;
      expect(container).toBeInTheDocument();

      // Touch events should not affect transform
      fireEvent.touchStart(container!, {
        touches: [{ clientX: 100, clientY: 100 }],
      });

      fireEvent.touchMove(container!, {
        touches: [{ clientX: 150, clientY: 120 }],
      });

      const content = screen.getByTestId('diagram-content').parentElement;
      expect(content).toHaveStyle('transform: translate(0px, 0px) scale(1)');
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA labels for controls', () => {
      render(
        <ResponsiveDiagramContainer>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      expect(screen.getByLabelText('Zoom in')).toBeInTheDocument();
      expect(screen.getByLabelText('Zoom out')).toBeInTheDocument();
      expect(screen.getByLabelText('Fit to container')).toBeInTheDocument();
      expect(screen.getByLabelText('Reset zoom')).toBeInTheDocument();
    });

    it('has proper focus management', () => {
      render(
        <ResponsiveDiagramContainer>
          <MockDiagramContent />
        </ResponsiveDiagramContainer>
      );

      const zoomInButton = screen.getByLabelText('Zoom in');
      zoomInButton.focus();
      expect(zoomInButton).toHaveFocus();

      // Should have focus ring classes
      expect(zoomInButton).toHaveClass('focus:outline-none', 'focus:ring-2', 'focus:ring-blue-500');
    });
  });
});