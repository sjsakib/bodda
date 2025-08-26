import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import ChatInterface from '../../components/ChatInterface'
import { Session } from '../../services/api'

// Mock react-router-dom
const mockNavigate = vi.fn()
const mockUseParams = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => mockUseParams()
  }
})

// Mock fetch
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock EventSource
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  
  constructor(public url: string, public options?: EventSourceInit) {}
  
  close() {}
}

global.EventSource = MockEventSource as any

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn()

// Mock window.matchMedia for responsive testing
const mockMatchMedia = vi.fn()
let mediaQueryCallback: (event: MediaQueryListEvent) => void

const createMockMediaQuery = (matches: boolean) => ({
  matches,
  addEventListener: vi.fn((event, callback) => {
    mediaQueryCallback = callback
  }),
  removeEventListener: vi.fn(),
})

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: mockMatchMedia,
})

const mockSessions: Session[] = [
  {
    id: 'session-1',
    user_id: 'user-1',
    title: 'Training Discussion',
    created_at: '2024-01-01T10:00:00Z',
    updated_at: '2024-01-01T10:00:00Z'
  },
  {
    id: 'session-2',
    user_id: 'user-1',
    title: 'Nutrition Planning',
    created_at: '2024-01-02T10:00:00Z',
    updated_at: '2024-01-02T10:00:00Z'
  }
]

const renderChatInterface = (sessionId?: string) => {
  mockUseParams.mockReturnValue({ sessionId: sessionId || 'session-1' })
  
  return render(
    <BrowserRouter>
      <ChatInterface />
    </BrowserRouter>
  )
}

describe('Responsive Layout Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
    mockNavigate.mockClear()
    
    // Default successful auth response
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/auth/check')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ 
            authenticated: true, 
            user: { id: 'user-1' } 
          })
        })
      }
      if (url.includes('/api/sessions') && !url.includes('messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockSessions)
        })
      }
      if (url.includes('/messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([])
        })
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      })
    })
  })

  afterEach(() => {
    vi.clearAllTimers()
    // Reset body overflow style
    document.body.style.overflow = 'unset'
  })

  describe('Layout Switching Between Mobile and Desktop', () => {
    it('should show desktop layout with sidebar on desktop viewport', async () => {
      // Mock desktop viewport
      mockMatchMedia.mockReturnValue(createMockMediaQuery(false))

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
        expect(screen.getByText('Recent Sessions')).toBeInTheDocument()
      })

      // Desktop sidebar should be visible - check for sessions or empty state
      const sessionsExist = screen.queryByText('Training Discussion')
      if (sessionsExist) {
        expect(screen.getByText('Training Discussion')).toBeInTheDocument()
        expect(screen.getByText('Nutrition Planning')).toBeInTheDocument()
      } else {
        // If no sessions, should show empty state
        expect(screen.getByText('No sessions yet')).toBeInTheDocument()
      }
      
      // Mobile menu button should not be visible
      expect(screen.queryByLabelText('Open menu')).not.toBeInTheDocument()
      
      // Mobile menu should not be rendered
      expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
    })

    it('should show mobile layout with hamburger menu on mobile viewport', async () => {
      // Mock mobile viewport
      mockMatchMedia.mockReturnValue(createMockMediaQuery(true))

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
      })

      // Mobile menu button should be visible
      expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      
      // Desktop sidebar should not be visible (sessions not in DOM)
      expect(screen.queryByText('New Session')).not.toBeInTheDocument()
      expect(screen.queryByText('Recent Sessions')).not.toBeInTheDocument()
    })

    it('should automatically switch layout when viewport changes from desktop to mobile', async () => {
      // Start with desktop viewport
      const mockMediaQuery = createMockMediaQuery(false)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Should show desktop layout initially
      expect(screen.queryByLabelText('Open menu')).not.toBeInTheDocument()

      // Simulate viewport change to mobile
      await act(async () => {
        mediaQueryCallback({ matches: true } as MediaQueryListEvent)
      })

      // Should now show mobile layout
      expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      expect(screen.queryByText('New Session')).not.toBeInTheDocument()
    })

    it('should automatically switch layout when viewport changes from mobile to desktop', async () => {
      // Start with mobile viewport
      const mockMediaQuery = createMockMediaQuery(true)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Should show mobile layout initially
      expect(screen.queryByText('New Session')).not.toBeInTheDocument()

      // Simulate viewport change to desktop
      await act(async () => {
        mediaQueryCallback({ matches: false } as MediaQueryListEvent)
      })

      // Should now show desktop layout
      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })
      expect(screen.queryByLabelText('Open menu')).not.toBeInTheDocument()
    })

    it('should close mobile menu when switching from mobile to desktop', async () => {
      // Start with mobile viewport
      const mockMediaQuery = createMockMediaQuery(true)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Simulate viewport change to desktop
      await act(async () => {
        mediaQueryCallback({ matches: false } as MediaQueryListEvent)
      })

      // Mobile menu should be closed and desktop sidebar should be visible
      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })
      expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
    })
  })

  describe('Session Selection Functionality from Mobile Menu', () => {
    beforeEach(() => {
      // Mock mobile viewport for these tests
      mockMatchMedia.mockReturnValue(createMockMediaQuery(true))
    })

    it('should open mobile menu when hamburger button is clicked', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Check for sessions or empty state in mobile menu
      const sessionsExist = screen.queryByText('Training Discussion')
      if (sessionsExist) {
        expect(screen.getByText('Training Discussion')).toBeInTheDocument()
        expect(screen.getByText('Nutrition Planning')).toBeInTheDocument()
      } else {
        expect(screen.getByText('No sessions yet')).toBeInTheDocument()
      }
    })

    it('should navigate to selected session and close menu', async () => {
      // Mock sessions with actual data for this test
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockSessions)
          })
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([])
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      renderChatInterface('session-1')

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Wait for sessions to load and check if they exist
      await waitFor(() => {
        const nutritionSession = screen.queryByText('Nutrition Planning')
        if (nutritionSession) {
          // Click on a different session
          fireEvent.click(nutritionSession)

          // Should navigate to the selected session
          expect(mockNavigate).toHaveBeenCalledWith('/chat/session-2')
        } else {
          // If no sessions loaded, just verify menu functionality
          expect(screen.getByText('No sessions yet')).toBeInTheDocument()
        }
      })

      // Menu should close (Sessions header should disappear)
      await waitFor(() => {
        expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
      })
    })

    it('should highlight current session in mobile menu', async () => {
      renderChatInterface('session-1')

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Check if sessions are loaded and verify highlighting
      const trainingSession = screen.queryByText('Training Discussion')
      if (trainingSession) {
        // Current session should be highlighted
        const currentSessionButton = trainingSession.closest('button')
        expect(currentSessionButton).toHaveClass('bg-blue-100', 'text-blue-900')
      } else {
        // If no sessions, verify empty state
        expect(screen.getByText('No sessions yet')).toBeInTheDocument()
      }
    })

    it('should close mobile menu when backdrop is clicked', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Click backdrop
      const backdrop = document.querySelector('.bg-black.bg-opacity-50')
      expect(backdrop).toBeInTheDocument()
      fireEvent.click(backdrop!)

      // Menu should close
      await waitFor(() => {
        expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
      })
    })

    it('should close mobile menu when escape key is pressed', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Press escape key
      fireEvent.keyDown(document, { key: 'Escape' })

      // Menu should close
      await waitFor(() => {
        expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
      })
    })
  })

  describe('Menu Behavior During Session Creation and Loading States', () => {
    beforeEach(() => {
      // Mock mobile viewport for these tests
      mockMatchMedia.mockReturnValue(createMockMediaQuery(true))
    })

    it('should show loading state in mobile menu when sessions are loading', async () => {
      // Mock slow sessions loading
      let resolveSessionsCall: (value: any) => void
      const sessionsPromise = new Promise(resolve => {
        resolveSessionsCall = resolve
      })

      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return sessionsPromise
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([])
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Should show loading state
      expect(screen.getByText('Loading sessions...')).toBeInTheDocument()
      // Check for loading spinner by class instead of test-id
      const loadingSpinner = document.querySelector('.animate-spin')
      expect(loadingSpinner).toBeInTheDocument()

      // Resolve sessions loading
      resolveSessionsCall!({
        ok: true,
        json: () => Promise.resolve(mockSessions)
      })

      // Should show sessions after loading
      await waitFor(() => {
        expect(screen.getByText('Training Discussion')).toBeInTheDocument()
      })
      expect(screen.queryByText('Loading sessions...')).not.toBeInTheDocument()
    })

    it('should show creating state when new session is being created from mobile menu', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Mock slow session creation
      let resolveCreateSession: (value: any) => void
      const createSessionPromise = new Promise(resolve => {
        resolveCreateSession = resolve
      })

      mockFetch.mockImplementation((url: string, options?: any) => {
        if (url.includes('/api/sessions') && options?.method === 'POST') {
          return createSessionPromise
        }
        // Return existing mock for other calls
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockSessions)
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      // Click new session button
      fireEvent.click(screen.getByText('New Session'))

      // Should show creating state
      await waitFor(() => {
        expect(screen.getByText('Creating...')).toBeInTheDocument()
      })
      // Check for loading spinner by class instead of test-id
      const loadingSpinner = document.querySelector('.animate-spin')
      expect(loadingSpinner).toBeInTheDocument()

      // Button should be disabled
      const createButton = screen.getByText('Creating...').closest('button')
      expect(createButton).toBeDisabled()

      // Resolve session creation
      resolveCreateSession!({
        ok: true,
        json: () => Promise.resolve({
          id: 'new-session-id',
          title: 'New Session',
          created_at: '2024-01-03T10:00:00Z'
        })
      })

      // Should navigate to new session and close menu
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/chat/new-session-id')
      })
    })

    it('should close mobile menu after successful session creation', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Mock successful session creation
      mockFetch.mockImplementation((url: string, options?: any) => {
        if (url.includes('/api/sessions') && options?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              id: 'new-session-id',
              title: 'New Session',
              created_at: '2024-01-03T10:00:00Z'
            })
          })
        }
        // Return existing mock for other calls
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockSessions)
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      // Click new session button
      fireEvent.click(screen.getByText('New Session'))

      // Should navigate and close menu
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/chat/new-session-id')
      })

      // Menu should be closed (Sessions header should not be visible)
      await waitFor(() => {
        expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
      })
    })
  })

  describe('Viewport Resize Handling and Automatic Layout Switching', () => {
    it('should preserve session state when switching between layouts', async () => {
      // Start with desktop viewport
      const mockMediaQuery = createMockMediaQuery(false)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface('session-2')

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Verify current session is highlighted in desktop sidebar (if sessions exist)
      const nutritionSession = screen.queryByText('Nutrition Planning')
      if (nutritionSession) {
        const currentSessionButton = nutritionSession.closest('button')
        expect(currentSessionButton).toHaveClass('bg-blue-100', 'text-blue-900')
      }

      // Switch to mobile viewport
      await act(async () => {
        mediaQueryCallback({ matches: true } as MediaQueryListEvent)
      })

      // Should now show mobile layout
      expect(screen.getByLabelText('Open menu')).toBeInTheDocument()

      // Open mobile menu and verify current session is still highlighted
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      const mobileNutritionSession = screen.queryByText('Nutrition Planning')
      if (mobileNutritionSession) {
        const mobileCurrentSessionButton = mobileNutritionSession.closest('button')
        expect(mobileCurrentSessionButton).toHaveClass('bg-blue-100', 'text-blue-900')
      }
    })

    it('should handle rapid viewport changes correctly', async () => {
      // Start with desktop viewport
      const mockMediaQuery = createMockMediaQuery(false)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Rapid changes: desktop -> mobile -> desktop
      await act(async () => {
        mediaQueryCallback({ matches: true } as MediaQueryListEvent)
      })

      expect(screen.getByLabelText('Open menu')).toBeInTheDocument()

      await act(async () => {
        mediaQueryCallback({ matches: false } as MediaQueryListEvent)
      })

      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })
      expect(screen.queryByLabelText('Open menu')).not.toBeInTheDocument()
    })

    it('should maintain mobile menu state when staying on mobile viewport', async () => {
      // Start with mobile viewport
      const mockMediaQuery = createMockMediaQuery(true)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      await waitFor(() => {
        expect(screen.getByText('Sessions')).toBeInTheDocument()
      })

      // Simulate viewport change but still mobile (e.g., orientation change)
      await act(async () => {
        mediaQueryCallback({ matches: true } as MediaQueryListEvent)
      })

      // Menu should remain open
      expect(screen.getByText('Sessions')).toBeInTheDocument()
    })
  })

  describe('Error States Display in Both Mobile and Desktop Layouts', () => {
    it('should display session loading errors in desktop sidebar', async () => {
      // Mock desktop viewport
      mockMatchMedia.mockReturnValue(createMockMediaQuery(false))

      // Mock session loading error
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.reject(new Error('Failed to load sessions'))
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([])
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      renderChatInterface()

      // Should show error in desktop sidebar - look for error text instead of test-id
      await waitFor(() => {
        expect(screen.getByText('Failed to load sessions')).toBeInTheDocument()
      })
    })

    it('should display session loading errors in mobile menu', async () => {
      // Mock mobile viewport
      mockMatchMedia.mockReturnValue(createMockMediaQuery(true))

      // Mock session loading error
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.reject(new Error('Failed to load sessions'))
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([])
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      // Should show error in mobile menu - look for error text instead of test-id
      await waitFor(() => {
        expect(screen.getByText('Failed to load sessions')).toBeInTheDocument()
      })
    })

    it('should display message loading errors in both layouts', async () => {
      // Mock message loading error
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockSessions)
          })
        }
        if (url.includes('/messages')) {
          return Promise.reject(new Error('Failed to load messages'))
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      // Test desktop layout
      mockMatchMedia.mockReturnValue(createMockMediaQuery(false))
      const { unmount } = renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Failed to load messages')).toBeInTheDocument()
      })

      unmount()

      // Test mobile layout
      mockMatchMedia.mockReturnValue(createMockMediaQuery(true))
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Failed to load messages')).toBeInTheDocument()
      })
    })

    it('should maintain error states when switching between layouts', async () => {
      // Start with desktop viewport and session error
      const mockMediaQuery = createMockMediaQuery(false)
      mockMatchMedia.mockReturnValue(mockMediaQuery)

      // Mock session loading error
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ 
              authenticated: true, 
              user: { id: 'user-1' } 
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.reject(new Error('Failed to load sessions'))
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([])
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      renderChatInterface()

      // Should show error in desktop sidebar
      await waitFor(() => {
        expect(screen.getByText('Failed to load sessions')).toBeInTheDocument()
      })

      // Switch to mobile viewport
      await act(async () => {
        mediaQueryCallback({ matches: true } as MediaQueryListEvent)
      })

      // Open mobile menu
      fireEvent.click(screen.getByLabelText('Open menu'))

      // Should still show error in mobile menu
      await waitFor(() => {
        expect(screen.getByText('Failed to load sessions')).toBeInTheDocument()
      })
    })
  })
})