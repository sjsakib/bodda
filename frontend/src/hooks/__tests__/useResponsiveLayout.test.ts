import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useResponsiveLayout } from '../useResponsiveLayout'

// Mock window.matchMedia
const mockMatchMedia = vi.fn()

// Store original matchMedia to restore later
const originalMatchMedia = window.matchMedia

describe('useResponsiveLayout Hook', () => {
  let mockMediaQueryList: {
    matches: boolean
    addEventListener: ReturnType<typeof vi.fn>
    removeEventListener: ReturnType<typeof vi.fn>
  }

  beforeEach(() => {
    // Create a mock MediaQueryList object
    mockMediaQueryList = {
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }

    // Mock window.matchMedia to return our mock object
    mockMatchMedia.mockReturnValue(mockMediaQueryList)
    window.matchMedia = mockMatchMedia

    vi.clearAllMocks()
  })

  afterEach(() => {
    // Restore original matchMedia
    window.matchMedia = originalMatchMedia
  })

  it('should initialize with correct default state for desktop', () => {
    mockMediaQueryList.matches = false // Desktop viewport
    
    const { result } = renderHook(() => useResponsiveLayout())

    expect(result.current.isMobile).toBe(false)
    expect(result.current.isMobileMenuOpen).toBe(false)
    expect(mockMatchMedia).toHaveBeenCalledWith('(max-width: 767px)')
  })

  it('should initialize with correct state for mobile viewport', () => {
    mockMediaQueryList.matches = true // Mobile viewport
    
    const { result } = renderHook(() => useResponsiveLayout())

    expect(result.current.isMobile).toBe(true)
    expect(result.current.isMobileMenuOpen).toBe(false)
  })

  it('should add event listener for media query changes on mount', () => {
    renderHook(() => useResponsiveLayout())

    expect(mockMediaQueryList.addEventListener).toHaveBeenCalledWith(
      'change',
      expect.any(Function)
    )
  })

  it('should remove event listener on unmount', () => {
    const { unmount } = renderHook(() => useResponsiveLayout())

    unmount()

    expect(mockMediaQueryList.removeEventListener).toHaveBeenCalledWith(
      'change',
      expect.any(Function)
    )
  })

  it('should toggle mobile menu open state', () => {
    const { result } = renderHook(() => useResponsiveLayout())

    expect(result.current.isMobileMenuOpen).toBe(false)

    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(true)

    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(false)
  })

  it('should close mobile menu', () => {
    const { result } = renderHook(() => useResponsiveLayout())

    // First open the menu
    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(true)

    // Then close it
    act(() => {
      result.current.closeMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(false)
  })

  it('should update isMobile state when viewport changes from desktop to mobile', () => {
    mockMediaQueryList.matches = false // Start with desktop
    
    const { result } = renderHook(() => useResponsiveLayout())

    expect(result.current.isMobile).toBe(false)

    // Simulate viewport change to mobile
    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]
    
    act(() => {
      changeHandler({ matches: true })
    })

    expect(result.current.isMobile).toBe(true)
  })

  it('should update isMobile state when viewport changes from mobile to desktop', () => {
    mockMediaQueryList.matches = true // Start with mobile
    
    const { result } = renderHook(() => useResponsiveLayout())

    expect(result.current.isMobile).toBe(true)

    // Simulate viewport change to desktop
    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]
    
    act(() => {
      changeHandler({ matches: false })
    })

    expect(result.current.isMobile).toBe(false)
  })

  it('should automatically close mobile menu when switching from mobile to desktop', () => {
    mockMediaQueryList.matches = true // Start with mobile
    
    const { result } = renderHook(() => useResponsiveLayout())

    // Open mobile menu
    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(true)

    // Simulate viewport change to desktop
    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]
    
    act(() => {
      changeHandler({ matches: false })
    })

    expect(result.current.isMobile).toBe(false)
    expect(result.current.isMobileMenuOpen).toBe(false)
  })

  it('should not close mobile menu when switching from desktop to mobile', () => {
    mockMediaQueryList.matches = false // Start with desktop
    
    const { result } = renderHook(() => useResponsiveLayout())

    // Simulate viewport change to mobile
    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]
    
    act(() => {
      changeHandler({ matches: true })
    })

    expect(result.current.isMobile).toBe(true)
    expect(result.current.isMobileMenuOpen).toBe(false) // Should remain closed
  })

  it('should maintain mobile menu state when staying on mobile viewport', () => {
    mockMediaQueryList.matches = true // Start with mobile
    
    const { result } = renderHook(() => useResponsiveLayout())

    // Open mobile menu
    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(true)

    // Simulate viewport change but still mobile
    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]
    
    act(() => {
      changeHandler({ matches: true })
    })

    expect(result.current.isMobile).toBe(true)
    expect(result.current.isMobileMenuOpen).toBe(true) // Should remain open
  })

  it('should provide stable function references', () => {
    const { result, rerender } = renderHook(() => useResponsiveLayout())

    const initialToggle = result.current.toggleMobileMenu
    const initialClose = result.current.closeMobileMenu

    rerender()

    expect(result.current.toggleMobileMenu).toBe(initialToggle)
    expect(result.current.closeMobileMenu).toBe(initialClose)
  })

  it('should handle multiple rapid viewport changes correctly', () => {
    mockMediaQueryList.matches = false // Start with desktop
    
    const { result } = renderHook(() => useResponsiveLayout())

    // Open mobile menu first (though we're on desktop)
    act(() => {
      result.current.toggleMobileMenu()
    })

    expect(result.current.isMobileMenuOpen).toBe(true)

    const changeHandler = mockMediaQueryList.addEventListener.mock.calls[0][1]

    // Rapid changes: desktop -> mobile -> desktop
    act(() => {
      changeHandler({ matches: true }) // to mobile
    })

    expect(result.current.isMobile).toBe(true)
    expect(result.current.isMobileMenuOpen).toBe(true) // Should stay open

    act(() => {
      changeHandler({ matches: false }) // back to desktop
    })

    expect(result.current.isMobile).toBe(false)
    expect(result.current.isMobileMenuOpen).toBe(false) // Should close
  })
})