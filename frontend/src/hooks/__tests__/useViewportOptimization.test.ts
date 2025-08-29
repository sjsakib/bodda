import { renderHook, act } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { 
  useViewportOptimization, 
  useLazyRender, 
  useDiagramPerformance,
  useResponsiveDiagramSize 
} from '../useViewportOptimization';

// Mock IntersectionObserver
const mockIntersectionObserver = vi.fn();
const mockObserve = vi.fn();
const mockUnobserve = vi.fn();
const mockDisconnect = vi.fn();

beforeEach(() => {
  mockIntersectionObserver.mockImplementation((callback) => ({
    observe: mockObserve,
    unobserve: mockUnobserve,
    disconnect: mockDisconnect,
    callback,
  }));

  global.IntersectionObserver = mockIntersectionObserver;

  // Mock ResizeObserver
  global.ResizeObserver = vi.fn().mockImplementation((callback) => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
    callback,
  }));

  // Mock performance.now
  global.performance = {
    ...global.performance,
    now: vi.fn(() => Date.now()),
  };

  // Mock requestAnimationFrame
  global.requestAnimationFrame = vi.fn((callback) => {
    setTimeout(callback, 16);
    return 1;
  });
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useViewportOptimization', () => {
  describe('Basic Functionality', () => {
    it('initializes with correct default state', () => {
      const { result } = renderHook(() => useViewportOptimization());

      expect(result.current.isInViewport).toBe(false);
      expect(result.current.shouldRender).toBe(false);
      expect(result.current.elementRef.current).toBe(null);
      expect(result.current.performanceMetrics.isVisible).toBe(false);
    });

    it('creates IntersectionObserver with correct options', () => {
      const config = {
        rootMargin: '100px',
        threshold: 0.5,
      };

      renderHook(() => useViewportOptimization(config));

      expect(mockIntersectionObserver).toHaveBeenCalledWith(
        expect.any(Function),
        {
          rootMargin: '100px',
          threshold: 0.5,
        }
      );
    });

    it('observes element when ref is set', () => {
      const { result } = renderHook(() => useViewportOptimization());
      
      const mockElement = document.createElement('div');
      
      act(() => {
        // Simulate setting the ref
        (result.current.elementRef as any).current = mockElement;
      });

      // Re-render to trigger useEffect
      renderHook(() => useViewportOptimization());

      expect(mockObserve).toHaveBeenCalledWith(mockElement);
    });

    it('disconnects observer on unmount', () => {
      const { unmount } = renderHook(() => useViewportOptimization());

      unmount();

      expect(mockDisconnect).toHaveBeenCalled();
    });
  });

  describe('Intersection Handling', () => {
    it('updates state when element enters viewport', () => {
      const onEnterViewport = vi.fn();
      const { result } = renderHook(() => 
        useViewportOptimization({ onEnterViewport })
      );

      // Get the callback passed to IntersectionObserver
      const callback = mockIntersectionObserver.mock.calls[0][0];

      act(() => {
        callback([{ isIntersecting: true }]);
      });

      expect(result.current.isInViewport).toBe(true);
      expect(result.current.shouldRender).toBe(true);
      expect(onEnterViewport).toHaveBeenCalled();
    });

    it('updates state when element exits viewport', () => {
      const onExitViewport = vi.fn();
      const { result } = renderHook(() => 
        useViewportOptimization({ onExitViewport })
      );

      const callback = mockIntersectionObserver.mock.calls[0][0];

      // First enter viewport
      act(() => {
        callback([{ isIntersecting: true }]);
      });

      // Then exit viewport
      act(() => {
        callback([{ isIntersecting: false }]);
      });

      expect(result.current.isInViewport).toBe(false);
      expect(result.current.shouldRender).toBe(true); // Should still render after being in viewport once
      expect(onExitViewport).toHaveBeenCalled();
    });

    it('maintains shouldRender true after element has been in viewport', () => {
      const { result } = renderHook(() => useViewportOptimization());

      const callback = mockIntersectionObserver.mock.calls[0][0];

      // Enter viewport
      act(() => {
        callback([{ isIntersecting: true }]);
      });

      expect(result.current.shouldRender).toBe(true);

      // Exit viewport
      act(() => {
        callback([{ isIntersecting: false }]);
      });

      expect(result.current.shouldRender).toBe(true); // Should remain true
    });
  });

  describe('Lazy Loading', () => {
    it('does not render initially when lazy loading is enabled', () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ enableLazyRendering: true })
      );

      expect(result.current.shouldRender).toBe(false);
    });

    it('renders immediately when lazy loading is disabled', () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ enableLazyRendering: false })
      );

      expect(result.current.shouldRender).toBe(true);
    });

    it('force render bypasses lazy loading', () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ enableLazyRendering: true })
      );

      expect(result.current.shouldRender).toBe(false);

      act(() => {
        result.current.forceRender();
      });

      expect(result.current.shouldRender).toBe(true);
    });
  });

  describe('Performance Monitoring', () => {
    it('tracks intersection time when enabled', () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ enablePerformanceMonitoring: true })
      );

      const callback = mockIntersectionObserver.mock.calls[0][0];

      // Mock performance.now to return predictable values
      (global.performance.now as any)
        .mockReturnValueOnce(1000) // Enter viewport
        .mockReturnValueOnce(1500); // Exit viewport

      // Enter viewport
      act(() => {
        callback([{ isIntersecting: true }]);
      });

      // Exit viewport
      act(() => {
        callback([{ isIntersecting: false }]);
      });

      expect(result.current.performanceMetrics.intersectionTime).toBe(500);
    });

    it('tracks render time when enabled', async () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ 
          enablePerformanceMonitoring: true,
          enableLazyRendering: false 
        })
      );

      // Mock performance.now for render timing
      (global.performance.now as any)
        .mockReturnValueOnce(2000) // Render start
        .mockReturnValueOnce(2100); // Render end

      // Wait for requestAnimationFrame to be called
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 20));
      });

      expect(result.current.performanceMetrics.renderTime).toBe(100);
    });

    it('does not track performance when disabled', () => {
      const { result } = renderHook(() => 
        useViewportOptimization({ enablePerformanceMonitoring: false })
      );

      expect(result.current.performanceMetrics.renderTime).toBeUndefined();
      expect(result.current.performanceMetrics.intersectionTime).toBeUndefined();
    });
  });
});

describe('useLazyRender', () => {
  it('provides simplified lazy loading interface', () => {
    const { result } = renderHook(() => useLazyRender('100px'));

    expect(result.current.elementRef).toBeDefined();
    expect(result.current.shouldRender).toBe(false);

    // Should create IntersectionObserver with custom root margin
    expect(mockIntersectionObserver).toHaveBeenCalledWith(
      expect.any(Function),
      expect.objectContaining({
        rootMargin: '100px',
      })
    );
  });

  it('enables rendering when element enters viewport', () => {
    const { result } = renderHook(() => useLazyRender());

    const callback = mockIntersectionObserver.mock.calls[0][0];

    act(() => {
      callback([{ isIntersecting: true }]);
    });

    expect(result.current.shouldRender).toBe(true);
  });
});

describe('useDiagramPerformance', () => {
  it('enables performance monitoring by default', () => {
    const { result } = renderHook(() => useDiagramPerformance());

    expect(result.current.performanceMetrics).toBeDefined();
    expect(result.current.isPerformant).toBe(true); // No render time yet
  });

  it('calculates performance status based on render time', async () => {
    const { result } = renderHook(() => useDiagramPerformance());

    // Mock slow render time
    (global.performance.now as any)
      .mockReturnValueOnce(1000)
      .mockReturnValueOnce(2500); // 1500ms render time

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 20));
    });

    expect(result.current.performanceMetrics.renderTime).toBe(1500);
    expect(result.current.isPerformant).toBe(false); // > 1000ms
  });
});

describe('useResponsiveDiagramSize', () => {
  it('initializes with zero dimensions', () => {
    const { result } = renderHook(() => useResponsiveDiagramSize());

    expect(result.current.dimensions).toEqual({
      width: 0,
      height: 0,
      isMobile: false,
      isTablet: false,
      isDesktop: false,
    });
  });

  it('categorizes screen sizes correctly', () => {
    const { result } = renderHook(() => useResponsiveDiagramSize());

    // Mock getBoundingClientRect for different sizes
    const mockElement = {
      getBoundingClientRect: vi.fn(),
    };

    (result.current.elementRef as any).current = mockElement;

    // Test mobile size
    mockElement.getBoundingClientRect.mockReturnValue({
      width: 400,
      height: 300,
    });

    act(() => {
      result.current.updateDimensions();
    });

    expect(result.current.dimensions.isMobile).toBe(true);
    expect(result.current.dimensions.isTablet).toBe(false);
    expect(result.current.dimensions.isDesktop).toBe(false);

    // Test tablet size
    mockElement.getBoundingClientRect.mockReturnValue({
      width: 800,
      height: 600,
    });

    act(() => {
      result.current.updateDimensions();
    });

    expect(result.current.dimensions.isMobile).toBe(false);
    expect(result.current.dimensions.isTablet).toBe(true);
    expect(result.current.dimensions.isDesktop).toBe(false);

    // Test desktop size
    mockElement.getBoundingClientRect.mockReturnValue({
      width: 1200,
      height: 800,
    });

    act(() => {
      result.current.updateDimensions();
    });

    expect(result.current.dimensions.isMobile).toBe(false);
    expect(result.current.dimensions.isTablet).toBe(false);
    expect(result.current.dimensions.isDesktop).toBe(true);
  });

  it('sets up ResizeObserver', () => {
    renderHook(() => useResponsiveDiagramSize());

    expect(global.ResizeObserver).toHaveBeenCalledWith(expect.any(Function));
  });
});