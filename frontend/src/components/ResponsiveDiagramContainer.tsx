import React, { useState, useRef, useEffect, useCallback } from 'react';
import { useResponsiveLayout } from '../hooks/useResponsiveLayout';

/**
 * Props interface for the ResponsiveDiagramContainer component
 */
interface ResponsiveDiagramContainerProps {
  /** Child diagram component to render */
  children: React.ReactNode;
  /** Optional CSS class name */
  className?: string;
  /** Whether to enable zoom and pan controls */
  enableZoomPan?: boolean;
  /** Whether to enable touch gestures on mobile */
  enableTouchGestures?: boolean;
  /** Minimum zoom level */
  minZoom?: number;
  /** Maximum zoom level */
  maxZoom?: number;
  /** Initial zoom level */
  initialZoom?: number;
  /** Whether to fit diagram to container on mount */
  fitToContainer?: boolean;
  /** Callback when zoom level changes */
  onZoomChange?: (zoom: number) => void;
}

/**
 * Responsive container component that provides zoom, pan, and touch controls for diagrams
 * Optimizes diagram display across different screen sizes and devices
 */
export const ResponsiveDiagramContainer: React.FC<ResponsiveDiagramContainerProps> = ({
  children,
  className = '',
  enableZoomPan = true,
  enableTouchGestures = true,
  minZoom = 0.1,
  maxZoom = 5,
  initialZoom = 1,
  fitToContainer = true,
  onZoomChange
}) => {
  const { isMobile } = useResponsiveLayout();
  const containerRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  
  // Transform state
  const [zoom, setZoom] = useState(initialZoom);
  const [panX, setPanX] = useState(0);
  const [panY, setPanY] = useState(0);
  
  // Interaction state
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [lastPanX, setLastPanX] = useState(0);
  const [lastPanY, setLastPanY] = useState(0);
  
  // Touch gesture state
  const [touchStart, setTouchStart] = useState<{ x: number; y: number; distance: number } | null>(null);
  const [lastTouchDistance, setLastTouchDistance] = useState(0);

  /**
   * Calculate distance between two touch points
   */
  const getTouchDistance = useCallback((touches: TouchList) => {
    if (touches.length < 2) return 0;
    
    const touch1 = touches[0];
    const touch2 = touches[1];
    
    return Math.sqrt(
      Math.pow(touch2.clientX - touch1.clientX, 2) + 
      Math.pow(touch2.clientY - touch1.clientY, 2)
    );
  }, []);

  /**
   * Get center point between two touches
   */
  const getTouchCenter = useCallback((touches: TouchList) => {
    if (touches.length < 2) return { x: 0, y: 0 };
    
    const touch1 = touches[0];
    const touch2 = touches[1];
    
    return {
      x: (touch1.clientX + touch2.clientX) / 2,
      y: (touch1.clientY + touch2.clientY) / 2
    };
  }, []);

  /**
   * Update zoom level with bounds checking
   */
  const updateZoom = useCallback((newZoom: number, centerX?: number, centerY?: number) => {
    const clampedZoom = Math.max(minZoom, Math.min(maxZoom, newZoom));
    
    if (clampedZoom !== zoom) {
      // Adjust pan to zoom around center point
      if (centerX !== undefined && centerY !== undefined && containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        const relativeX = centerX - rect.left;
        const relativeY = centerY - rect.top;
        
        const zoomRatio = clampedZoom / zoom;
        const newPanX = relativeX - (relativeX - panX) * zoomRatio;
        const newPanY = relativeY - (relativeY - panY) * zoomRatio;
        
        setPanX(newPanX);
        setPanY(newPanY);
      }
      
      setZoom(clampedZoom);
      onZoomChange?.(clampedZoom);
    }
  }, [zoom, minZoom, maxZoom, panX, panY, onZoomChange]);

  /**
   * Fit diagram to container size
   */
  const fitToContainerSize = useCallback(() => {
    if (!containerRef.current || !contentRef.current) return;
    
    const container = containerRef.current;
    const content = contentRef.current;
    
    const containerRect = container.getBoundingClientRect();
    const contentRect = content.getBoundingClientRect();
    
    if (contentRect.width === 0 || contentRect.height === 0) return;
    
    const scaleX = (containerRect.width - 40) / contentRect.width; // 40px padding
    const scaleY = (containerRect.height - 40) / contentRect.height;
    const scale = Math.min(scaleX, scaleY, 1); // Don't zoom in beyond 100%
    
    setZoom(scale);
    setPanX((containerRect.width - contentRect.width * scale) / 2);
    setPanY((containerRect.height - contentRect.height * scale) / 2);
    onZoomChange?.(scale);
  }, [onZoomChange]);

  /**
   * Reset zoom and pan to initial state
   */
  const resetTransform = useCallback(() => {
    setZoom(initialZoom);
    setPanX(0);
    setPanY(0);
    onZoomChange?.(initialZoom);
  }, [initialZoom, onZoomChange]);

  /**
   * Handle mouse wheel zoom
   */
  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (!enableZoomPan) return;
    
    e.preventDefault();
    
    const zoomDelta = e.deltaY > 0 ? 0.9 : 1.1;
    const newZoom = zoom * zoomDelta;
    
    updateZoom(newZoom, e.clientX, e.clientY);
  }, [enableZoomPan, zoom, updateZoom]);

  /**
   * Handle mouse down for pan start
   */
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (!enableZoomPan || e.button !== 0) return;
    
    setIsDragging(true);
    setDragStart({ x: e.clientX, y: e.clientY });
    setLastPanX(panX);
    setLastPanY(panY);
    
    e.preventDefault();
  }, [enableZoomPan, panX, panY]);

  /**
   * Handle mouse move for panning
   */
  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging || !enableZoomPan) return;
    
    const deltaX = e.clientX - dragStart.x;
    const deltaY = e.clientY - dragStart.y;
    
    setPanX(lastPanX + deltaX);
    setPanY(lastPanY + deltaY);
  }, [isDragging, enableZoomPan, dragStart, lastPanX, lastPanY]);

  /**
   * Handle mouse up for pan end
   */
  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);



  /**
   * Fit to container on mount if enabled
   */
  useEffect(() => {
    if (fitToContainer) {
      // Delay to ensure content is rendered
      const timer = setTimeout(fitToContainerSize, 100);
      return () => clearTimeout(timer);
    }
  }, [fitToContainer, fitToContainerSize]);

  /**
   * Add touch event listeners with proper passive: false option
   */
  useEffect(() => {
    const container = containerRef.current;
    if (!container || !enableTouchGestures) return;

    const handleTouchStartNative = (e: TouchEvent) => {
      const touches = e.touches;
      
      if (touches.length === 1) {
        // Single touch - start pan
        setIsDragging(true);
        setDragStart({ x: touches[0].clientX, y: touches[0].clientY });
        setLastPanX(panX);
        setLastPanY(panY);
      } else if (touches.length === 2) {
        // Two finger touch - start pinch zoom
        const distance = getTouchDistance(touches);
        const center = getTouchCenter(touches);
        
        setTouchStart({ x: center.x, y: center.y, distance });
        setLastTouchDistance(distance);
        setIsDragging(false);
      }
      
      e.preventDefault();
    };

    const handleTouchMoveNative = (e: TouchEvent) => {
      const touches = e.touches;
      
      if (touches.length === 1 && isDragging) {
        // Single touch pan
        const deltaX = touches[0].clientX - dragStart.x;
        const deltaY = touches[0].clientY - dragStart.y;
        
        setPanX(lastPanX + deltaX);
        setPanY(lastPanY + deltaY);
      } else if (touches.length === 2 && touchStart) {
        // Two finger pinch zoom
        const distance = getTouchDistance(touches);
        const center = getTouchCenter(touches);
        
        if (lastTouchDistance > 0) {
          const zoomRatio = distance / lastTouchDistance;
          const newZoom = zoom * zoomRatio;
          
          updateZoom(newZoom, center.x, center.y);
        }
        
        setLastTouchDistance(distance);
      }
      
      e.preventDefault();
    };

    const handleTouchEndNative = () => {
      setIsDragging(false);
      setTouchStart(null);
      setLastTouchDistance(0);
    };

    container.addEventListener('touchstart', handleTouchStartNative, { passive: false });
    container.addEventListener('touchmove', handleTouchMoveNative, { passive: false });
    container.addEventListener('touchend', handleTouchEndNative, { passive: false });

    return () => {
      container.removeEventListener('touchstart', handleTouchStartNative);
      container.removeEventListener('touchmove', handleTouchMoveNative);
      container.removeEventListener('touchend', handleTouchEndNative);
    };
  }, [
    enableTouchGestures,
    panX,
    panY,
    isDragging,
    dragStart,
    lastPanX,
    lastPanY,
    touchStart,
    lastTouchDistance,
    zoom,
    getTouchDistance,
    getTouchCenter,
    updateZoom
  ]);

  /**
   * Add global mouse event listeners for dragging
   */
  useEffect(() => {
    if (!isDragging) return;
    
    const handleGlobalMouseMove = (e: MouseEvent) => {
      if (!enableZoomPan) return;
      
      const deltaX = e.clientX - dragStart.x;
      const deltaY = e.clientY - dragStart.y;
      
      setPanX(lastPanX + deltaX);
      setPanY(lastPanY + deltaY);
    };
    
    const handleGlobalMouseUp = () => {
      setIsDragging(false);
    };
    
    document.addEventListener('mousemove', handleGlobalMouseMove, { passive: false });
    document.addEventListener('mouseup', handleGlobalMouseUp, { passive: false });
    
    return () => {
      document.removeEventListener('mousemove', handleGlobalMouseMove);
      document.removeEventListener('mouseup', handleGlobalMouseUp);
    };
  }, [isDragging, enableZoomPan, dragStart, lastPanX, lastPanY]);

  const containerClasses = `
    relative overflow-hidden bg-white border border-gray-200 rounded-lg
    ${isMobile ? 'touch-none' : 'select-none'}
    ${enableZoomPan ? 'cursor-grab' : ''}
    ${isDragging ? 'cursor-grabbing' : ''}
    ${className}
  `.trim();

  const contentStyle = {
    transform: `translate(${panX}px, ${panY}px) scale(${zoom})`,
    transformOrigin: '0 0',
    transition: isDragging ? 'none' : 'transform 0.2s ease-out'
  };

  return (
    <div className="responsive-diagram-container bg-white">
      {/* Control buttons */}
      {enableZoomPan && (
        <div className={`
          absolute top-2 right-2 z-10 flex flex-col space-y-1
          ${isMobile ? 'space-y-2' : ''}
        `}>
          <button
            onClick={() => updateZoom(zoom * 1.2)}
            className={`
              bg-white border border-gray-300 rounded shadow-sm hover:bg-gray-50
              focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
              ${isMobile ? 'p-3 text-lg' : 'p-2 text-sm'}
            `}
            aria-label="Zoom in"
            title="Zoom in"
          >
            <svg className={`${isMobile ? 'w-5 h-5' : 'w-4 h-4'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </button>
          
          <button
            onClick={() => updateZoom(zoom * 0.8)}
            className={`
              bg-white border border-gray-300 rounded shadow-sm hover:bg-gray-50
              focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
              ${isMobile ? 'p-3 text-lg' : 'p-2 text-sm'}
            `}
            aria-label="Zoom out"
            title="Zoom out"
          >
            <svg className={`${isMobile ? 'w-5 h-5' : 'w-4 h-4'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 12H6" />
            </svg>
          </button>
          
          <button
            onClick={fitToContainerSize}
            className={`
              bg-white border border-gray-300 rounded shadow-sm hover:bg-gray-50
              focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
              ${isMobile ? 'p-3 text-lg' : 'p-2 text-sm'}
            `}
            aria-label="Fit to container"
            title="Fit to container"
          >
            <svg className={`${isMobile ? 'w-5 h-5' : 'w-4 h-4'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
          </button>
          
          <button
            onClick={resetTransform}
            className={`
              bg-white border border-gray-300 rounded shadow-sm hover:bg-gray-50
              focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
              ${isMobile ? 'p-3 text-lg' : 'p-2 text-sm'}
            `}
            aria-label="Reset zoom"
            title="Reset zoom"
          >
            <svg className={`${isMobile ? 'w-5 h-5' : 'w-4 h-4'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      )}

      {/* Zoom level indicator */}
      {enableZoomPan && (
        <div className={`
          absolute bottom-2 left-2 z-10 bg-black bg-opacity-75 text-white rounded px-2 py-1
          ${isMobile ? 'text-sm' : 'text-xs'}
        `}>
          {Math.round(zoom * 100)}%
        </div>
      )}

      {/* Main container */}
      <div
        ref={containerRef}
        className={containerClasses}
        style={{ 
          height: isMobile ? '60vh' : '400px',
          minHeight: isMobile ? '300px' : '200px'
        }}
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}

      >
        <div
          ref={contentRef}
          style={contentStyle}
          className="diagram-content"
        >
          {children}
        </div>
      </div>

      {/* Mobile instructions */}
      {isMobile && enableTouchGestures && (
        <div className="mt-2 text-xs text-gray-500 text-center">
          Pinch to zoom • Drag to pan • Use controls to reset
        </div>
      )}
    </div>
  );
};

export default ResponsiveDiagramContainer;