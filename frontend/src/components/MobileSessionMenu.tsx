import { useEffect, useRef, useCallback } from 'react'
import { Session } from '../services/api'
import { LoadingSpinner } from './ErrorBoundary'
import { SessionErrorHandler } from './ApiErrorHandler'

interface MobileSessionMenuProps {
  sessions: Session[]
  currentSessionId?: string
  onCreateSession: () => void
  isCreatingSession: boolean
  onSelectSession: (sessionId: string) => void
  onDeleteSession: (sessionId: string) => void
  isOpen: boolean
  onClose: () => void
  isLoading?: boolean
  error?: unknown
  onRetryLoad?: () => void
}

export default function MobileSessionMenu({
  sessions,
  currentSessionId,
  onCreateSession,
  isCreatingSession,
  onSelectSession,
  onDeleteSession,
  isOpen,
  onClose,
  isLoading = false,
  error = null,
  onRetryLoad
}: MobileSessionMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  const closeButtonRef = useRef<HTMLButtonElement>(null)
  const newSessionButtonRef = useRef<HTMLButtonElement>(null)
  const firstSessionButtonRef = useRef<HTMLButtonElement>(null)
  const lastFocusedElementRef = useRef<HTMLElement | null>(null)

  // Handle keyboard navigation within menu
  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    if (!isOpen) return

    const focusableElements = menuRef.current?.querySelectorAll(
      'button:not([disabled]), [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    ) as NodeListOf<HTMLElement>

    if (!focusableElements || focusableElements.length === 0) return

    const firstElement = focusableElements[0]
    const lastElement = focusableElements[focusableElements.length - 1]

    switch (event.key) {
      case 'Escape':
        event.preventDefault()
        onClose()
        break
      case 'Tab':
        // Trap focus within the menu
        if (event.shiftKey) {
          // Shift + Tab (backward)
          if (document.activeElement === firstElement) {
            event.preventDefault()
            lastElement.focus()
          }
        } else {
          // Tab (forward)
          if (document.activeElement === lastElement) {
            event.preventDefault()
            firstElement.focus()
          }
        }
        break
      case 'ArrowDown':
        event.preventDefault()
        const currentIndex = Array.from(focusableElements).indexOf(document.activeElement as HTMLElement)
        const nextIndex = currentIndex < focusableElements.length - 1 ? currentIndex + 1 : 0
        focusableElements[nextIndex].focus()
        break
      case 'ArrowUp':
        event.preventDefault()
        const currentUpIndex = Array.from(focusableElements).indexOf(document.activeElement as HTMLElement)
        const prevIndex = currentUpIndex > 0 ? currentUpIndex - 1 : focusableElements.length - 1
        focusableElements[prevIndex].focus()
        break
    }
  }, [isOpen, onClose])

  // Handle escape key and focus management
  useEffect(() => {
    if (isOpen) {
      // Store the currently focused element before opening menu
      lastFocusedElementRef.current = document.activeElement as HTMLElement

      // Add keyboard event listener
      document.addEventListener('keydown', handleKeyDown)
      
      // Prevent body scroll when menu is open
      document.body.style.overflow = 'hidden'

      // Focus the close button when menu opens
      setTimeout(() => {
        closeButtonRef.current?.focus()
      }, 100) // Small delay to ensure menu is rendered

      // Announce menu opening to screen readers
      const announcement = document.createElement('div')
      announcement.setAttribute('aria-live', 'polite')
      announcement.setAttribute('aria-atomic', 'true')
      announcement.className = 'sr-only'
      announcement.textContent = 'Session menu opened'
      document.body.appendChild(announcement)
      
      setTimeout(() => {
        document.body.removeChild(announcement)
      }, 1000)
    } else {
      // Return focus to the element that opened the menu
      if (lastFocusedElementRef.current) {
        lastFocusedElementRef.current.focus()
        lastFocusedElementRef.current = null
      }

      // Announce menu closing to screen readers
      const announcement = document.createElement('div')
      announcement.setAttribute('aria-live', 'polite')
      announcement.setAttribute('aria-atomic', 'true')
      announcement.className = 'sr-only'
      announcement.textContent = 'Session menu closed'
      document.body.appendChild(announcement)
      
      setTimeout(() => {
        document.body.removeChild(announcement)
      }, 1000)
    }

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = 'unset'
    }
  }, [isOpen, handleKeyDown])

  // Handle session selection and close menu
  const handleSessionSelect = (sessionId: string) => {
    onSelectSession(sessionId)
    onClose()
  }

  // Handle session deletion
  const handleSessionDelete = (sessionId: string, sessionTitle: string) => {
    if (window.confirm(`Are you sure you want to delete "${sessionTitle}"? This action cannot be undone.`)) {
      onDeleteSession(sessionId)
    }
  }

  if (!isOpen) return null

  return (
    <div 
      id="mobile-session-menu"
      className="fixed inset-0 z-50 md:hidden"
      role="dialog"
      aria-modal="true"
      aria-labelledby="mobile-menu-title"
      aria-describedby="mobile-menu-description"
    >
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black bg-opacity-50 transition-opacity duration-300"
        onClick={onClose}
        aria-hidden="true"
      />
      
      {/* Menu Panel */}
      <div 
        ref={menuRef}
        className="absolute left-0 top-0 h-full w-80 max-w-[85vw] bg-white shadow-xl transform transition-transform duration-300 ease-in-out"
        role="navigation"
        aria-label="Session navigation menu"
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200">
            <h2 id="mobile-menu-title" className="text-lg font-semibold text-gray-900">Sessions</h2>
            <p id="mobile-menu-description" className="sr-only">
              Navigate between chat sessions or create a new session
            </p>
            <button
              ref={closeButtonRef}
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-gray-100 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              aria-label="Close session menu"
              aria-describedby="close-menu-help"
            >
              <svg 
                className="w-6 h-6 text-gray-600" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M6 18L18 6M6 6l12 12" 
                />
              </svg>
            </button>
            <div id="close-menu-help" className="sr-only">
              Press Escape key or click this button to close the menu
            </div>
          </div>

          {/* New Session Button */}
          <div className="p-4 border-b border-gray-200">
            <button
              ref={newSessionButtonRef}
              onClick={onCreateSession}
              disabled={isCreatingSession}
              className="w-full bg-blue-600 text-white px-4 py-3 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition-colors flex items-center justify-center min-h-[44px] text-base font-medium focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              aria-label={isCreatingSession ? "Creating new session..." : "Create new chat session"}
              aria-describedby="new-session-help"
            >
              {isCreatingSession ? (
                <>
                  <LoadingSpinner size="sm" className="mr-2" aria-hidden="true" />
                  Creating...
                </>
              ) : (
                'New Session'
              )}
            </button>
            <div id="new-session-help" className="sr-only">
              Creates a new chat session and navigates to it
            </div>
          </div>
          
          {/* Sessions List */}
          <div className="flex-1 overflow-y-auto" role="region" aria-labelledby="sessions-heading">
            <div className="p-4">
              <h3 id="sessions-heading" className="text-sm font-semibold text-gray-600 mb-3">Recent Sessions</h3>
              
              {error && onRetryLoad ? (
                <SessionErrorHandler 
                  error={error} 
                  onRetry={onRetryLoad}
                  loading={isLoading}
                />
              ) : isLoading ? (
                <div className="text-center py-8" role="status" aria-live="polite">
                  <LoadingSpinner className="mx-auto mb-2" aria-hidden="true" />
                  <p className="text-gray-600 text-sm">Loading sessions...</p>
                </div>
              ) : !sessions || sessions.length === 0 ? (
                <div className="text-center text-gray-500 py-8" role="status">
                  <p className="text-base">No sessions yet</p>
                  <p className="text-sm mt-1">Start a new conversation!</p>
                </div>
              ) : (
                <div className="space-y-2" role="list" aria-label="Chat sessions">
                  {sessions.map((session, index) => (
                    <div key={session.id} className="relative group" role="listitem">
                      <button
                        ref={index === 0 ? firstSessionButtonRef : undefined}
                        onClick={() => handleSessionSelect(session.id)}
                        className={`w-full text-left p-4 rounded-lg transition-colors min-h-[44px] focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${
                          currentSessionId === session.id
                            ? 'bg-blue-100 text-blue-900'
                            : 'hover:bg-gray-100'
                        }`}
                        aria-label={`Select session: ${session.title}, created on ${new Date(session.created_at).toLocaleDateString()}`}
                        aria-current={currentSessionId === session.id ? 'page' : undefined}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-base truncate">
                              {session.title}
                            </div>
                            <div className="text-sm text-gray-500 mt-1" aria-hidden="true">
                              {new Date(session.created_at).toLocaleDateString()}
                            </div>
                          </div>
                          
                          {/* Delete button */}
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleSessionDelete(session.id, session.title)
                            }}
                            className="ml-3 p-2 rounded hover:bg-red-100 text-gray-400 hover:text-red-600 transition-colors flex-shrink-0"
                            aria-label={`Delete session: ${session.title}`}
                            title="Delete session"
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}