import { useState, useEffect, useRef, useCallback } from 'react';

/**
 * Interface for viewport optimization configuration
 */
interface ViewportOptimizationConfig {
  /** Root margin for intersection observer (default: '50px') */
  rootMargin?: string;
  /** Threshold for intersection observer (default: 0.1) */
  threshold?: number;
  /** Whether to enable lazy rendering (default: true) */
  enableLazyRendering?: boolean;
  /** Whether to enable performance monitoring (default: false) */
  enablePerformanceMonitoring?: boolean;
  /** Callback when element enters viewport */
  onEnterViewport?: () => void;
  /** Callback when element exits viewport */
  onExitViewport?: () => void;
}

/**
 * Interface for viewport optimization return value
 */
interface UseViewportOptimizationReturn {
  /** Ref to attach to the element to observe */
  elementRef: React.RefObject<HTMLElement>;
  /** Whether the element is currently in viewport */
  isInViewport: boolean;
  /** Whether the element should render (considering lazy loading) */
  shouldRender: boolean;
  /** Performance metrics if monitoring is enabled */
  performanceMetrics: {
    renderTime?: number;
    intersectionTime?: number;
    isVisible: boolean;
  };
  /** Force render the element (bypass lazy loading) */
  forceRender: () => void;
}

/**
 * Custom hook for viewport-based rendering optimizations
 * Provides lazy loading and performance monitoring for diagram components
 */
export const useViewportOptimization = (
  config: ViewportOptimizationConfig = {}
): UseViewportOptimizationReturn => {
  const {
    rootMargin = '50px',
    threshold = 0.1,
    enableLazyRendering = true,
    enablePerformanceMonitoring = false,
    onEnterViewport,
    onExitViewport,
  } = config;

  // State
  const [isInViewport, setIsInViewport] = useState(false);
  const [hasBeenInViewport, setHasBeenInViewport] = useState(false);
  const [forceRenderFlag, setForceRenderFlag] = useState(false);
  const [performanceMetrics, setPerformanceMetrics] = useState({
    renderTime: undefined as number | undefined,
    intersectionTime: undefined as number | undefined,
    isVisible: false,
  });

  // Refs
  const elementRef = useRef<HTMLElement>(null);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const intersectionStartTime = useRef<number>(0);
  const renderStartTime = useRef<number>(0);

  /**
   * Handle intersection observer callback
   */
  const handleIntersection = useCallback((entries: IntersectionObserverEntry[]) => {
    const entry = entries[0];
    const isCurrentlyInViewport = entry.isIntersecting;
    
    // Performance monitoring
    if (enablePerformanceMonitoring) {
      if (isCurrentlyInViewport && !isInViewport) {
        // Entering viewport
        intersectionStartTime.current = performance.now();
      } else if (!isCurrentlyInViewport && isInViewport) {
        // Exiting viewport
        const intersectionTime = performance.now() - intersectionStartTime.current;
        setPerformanceMetrics(prev => ({
          ...prev,
          intersectionTime,
          isVisible: false,
        }));
      }
    }

    setIsInViewport(isCurrentlyInViewport);
    
    if (isCurrentlyInViewport) {
      setHasBeenInViewport(true);
      onEnterViewport?.();
      
      if (enablePerformanceMonitoring) {
        setPerformanceMetrics(prev => ({ ...prev, isVisible: true }));
      }
    } else {
      onExitViewport?.();
    }
  }, [isInViewport, enablePerformanceMonitoring, onEnterViewport, onExitViewport]);

  /**
   * Set up intersection observer
   */
  useEffect(() => {
    if (!elementRef.current) return;

    // Create intersection observer
    observerRef.current = new IntersectionObserver(handleIntersection, {
      rootMargin,
      threshold,
    });

    // Start observing
    observerRef.current.observe(elementRef.current);

    // Cleanup
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [handleIntersection, rootMargin, threshold]);

  /**
   * Force render function
   */
  const forceRender = useCallback(() => {
    setForceRenderFlag(true);
    setHasBeenInViewport(true);
  }, []);

  /**
   * Determine if element should render
   */
  const shouldRender = !enableLazyRendering || hasBeenInViewport || forceRenderFlag;

  /**
   * Performance monitoring for render time
   */
  useEffect(() => {
    if (enablePerformanceMonitoring && shouldRender && !performanceMetrics.renderTime) {
      renderStartTime.current = performance.now();
      
      // Use requestAnimationFrame to measure after render
      requestAnimationFrame(() => {
        const renderTime = performance.now() - renderStartTime.current;
        setPerformanceMetrics(prev => ({
          ...prev,
          renderTime,
        }));
      });
    }
  }, [shouldRender, enablePerformanceMonitoring, performanceMetrics.renderTime]);

  return {
    elementRef,
    isInViewport,
    shouldRender,
    performanceMetrics,
    forceRender,
  };
};

/**
 * Simplified hook for basic lazy loading
 */
export const useLazyRender = (rootMargin = '50px') => {
  const { elementRef, shouldRender } = useViewportOptimization({
    rootMargin,
    enableLazyRendering: true,
  });

  return { elementRef, shouldRender };
};

/**
 * Hook for performance monitoring of diagram rendering
 */
export const useDiagramPerformance = () => {
  const { performanceMetrics, ...rest } = useViewportOptimization({
    enablePerformanceMonitoring: true,
    enableLazyRendering: false,
  });

  return {
    ...rest,
    performanceMetrics,
    isPerformant: performanceMetrics.renderTime ? performanceMetrics.renderTime < 1000 : true,
  };
};

/**
 * Hook for responsive diagram sizing based on viewport
 */
export const useResponsiveDiagramSize = () => {
  const [dimensions, setDimensions] = useState({
    width: 0,
    height: 0,
    isMobile: false,
    isTablet: false,
    isDesktop: false,
  });

  const elementRef = useRef<HTMLElement>(null);

  const updateDimensions = useCallback(() => {
    if (!elementRef.current) return;

    const rect = elementRef.current.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    setDimensions({
      width,
      height,
      isMobile: width < 768,
      isTablet: width >= 768 && width < 1024,
      isDesktop: width >= 1024,
    });
  }, []);

  useEffect(() => {
    updateDimensions();

    const resizeObserver = new ResizeObserver(updateDimensions);
    if (elementRef.current) {
      resizeObserver.observe(elementRef.current);
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, [updateDimensions]);

  return {
    elementRef,
    dimensions,
    updateDimensions,
  };
};

export default useViewportOptimization;