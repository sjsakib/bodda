import { useState, useEffect, useCallback } from 'react'

export interface UseResponsiveLayoutReturn {
  isMobile: boolean
  isMobileMenuOpen: boolean
  toggleMobileMenu: () => void
  closeMobileMenu: () => void
}

const MOBILE_BREAKPOINT = '(max-width: 767px)' // Below 768px (Tailwind's md breakpoint)

export function useResponsiveLayout(): UseResponsiveLayoutReturn {
  const [isMobile, setIsMobile] = useState<boolean>(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState<boolean>(false)

  // Initialize mobile state based on current viewport
  useEffect(() => {
    const mediaQuery = window.matchMedia(MOBILE_BREAKPOINT)
    setIsMobile(mediaQuery.matches)

    const handleMediaChange = (event: MediaQueryListEvent) => {
      setIsMobile(event.matches)
      
      // Automatically close mobile menu when switching to desktop
      if (!event.matches) {
        setIsMobileMenuOpen(prev => prev ? false : prev)
      }
    }

    // Add event listener for viewport changes
    mediaQuery.addEventListener('change', handleMediaChange)

    // Cleanup event listener on component unmount
    return () => {
      mediaQuery.removeEventListener('change', handleMediaChange)
    }
  }, [])

  const toggleMobileMenu = useCallback(() => {
    setIsMobileMenuOpen(prev => !prev)
  }, [])

  const closeMobileMenu = useCallback(() => {
    setIsMobileMenuOpen(false)
  }, [])

  return {
    isMobile,
    isMobileMenuOpen,
    toggleMobileMenu,
    closeMobileMenu,
  }
}