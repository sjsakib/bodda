import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import App from '../../App'
import { apiClient } from '../../services/api'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock EventSource
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  
  constructor(public url: string, public options?: EventSourceInit) {}
  
  close() {}
  
  // Helper method to simulate events
  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }))
    }
  }
  
  simulateError() {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }
}

global.EventSource = MockEventSource as any

// Test wrapper component - no router since App already has one
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

describe('User Workflows E2E Tests', () => {
  let mockEventSource: MockEventSource

  beforeEach(() => {
    mockFetch.mockClear()
    vi.clearAllMocks()
    
    // Mock successful auth check by default
    mockFetch.mockImplementation((url: string, options?: RequestInit) => {
      if (url.includes('/api/auth/check')) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: () => Promise.resolve({ error: 'Not authenticated' })
        })
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      })
    })
  })

  afterEach(() => {
    if (mockEventSource) {
      mockEventSource.close()
    }
  })

  describe('Authentication Flow', () => {
    it('should show landing page for unauthenticated users', async () => {
      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText('Bodda')).toBeInTheDocument()
        expect(screen.getByTestId('strava-connect-button')).toBeInTheDocument()
        expect(screen.getByTestId('disclaimer')).toBeInTheDocument()
      })
    })

    it('should redirect to Strava OAuth when connect button is clicked', async () => {
      const user = userEvent.setup()
      
      // Mock window.location.href
      const originalLocation = window.location
      delete (window as any).location
      window.location = { ...originalLocation, href: '' }
      
      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByTestId('strava-connect-button')).toBeInTheDocument()
      })
      
      await user.click(screen.getByTestId('strava-connect-button'))
      
      expect(window.location.href).toBe('/auth/strava')
      
      // Restore window.location
      window.location = originalLocation
    })

    it('should redirect authenticated users to chat interface', async () => {
      // Mock successful authentication
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: {
                id: 'user-1',
                strava_id: 12345,
                first_name: 'John',
                last_name: 'Doe'
              }
            })
          })
        }
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({})
        })
      })

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })
    })
  })

  describe('Complete User Journey Integration', () => {
    it('should handle complete user journey from landing to chat', async () => {
      const user = userEvent.setup()
      
      // Start with unauthenticated state
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: false,
            status: 401,
            json: () => Promise.resolve({ error: 'Not authenticated' })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      // Should show landing page
      await waitFor(() => {
        expect(screen.getByText('Bodda')).toBeInTheDocument()
        expect(screen.getByTestId('strava-connect-button')).toBeInTheDocument()
      })

      // Mock successful authentication after OAuth
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: {
                id: 'user-1',
                strava_id: 12345,
                first_name: 'John',
                last_name: 'Doe'
              }
            })
          })
        }
        if (url.includes('/api/sessions') && !options?.method) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              sessions: [
                {
                  id: 'session-1',
                  title: 'Previous Chat',
                  created_at: '2024-01-01T10:00:00Z'
                }
              ]
            })
          })
        }
        if (url.includes('/api/sessions') && options?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              session: {
                id: 'new-session',
                title: 'New Chat Session',
                created_at: new Date().toISOString()
              }
            })
          })
        }
        if (url.includes('/messages') && options?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              user_message: {
                id: 'msg-1',
                content: 'How should I train for a 5K?',
                role: 'user',
                created_at: new Date().toISOString()
              },
              assistant_message: {
                id: 'msg-2',
                content: 'For 5K training, focus on building your aerobic base...',
                role: 'assistant',
                created_at: new Date().toISOString()
              }
            })
          })
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ messages: [] })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      // Simulate OAuth callback by triggering auth check
      const connectButton = screen.getByTestId('strava-connect-button')
      await user.click(connectButton)

      // Should redirect to chat interface
      await waitFor(() => {
        expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
        expect(screen.getByText('Previous Chat')).toBeInTheDocument()
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })

      // Create new session
      await user.click(screen.getByText('New Session'))

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          '/api/sessions',
          expect.objectContaining({ method: 'POST' })
        )
      })

      // Should be able to send a message
      const messageInput = screen.getByPlaceholderText(/Ask your AI coach/)
      await user.type(messageInput, 'How should I train for a 5K?')
      
      const sendButton = screen.getByText('Send')
      await user.click(sendButton)

      // Should show user message immediately
      expect(screen.getByText('How should I train for a 5K?')).toBeInTheDocument()
      
      // Should call the send message API
      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/messages'),
          expect.objectContaining({ 
            method: 'POST',
            body: JSON.stringify({ content: 'How should I train for a 5K?' })
          })
        )
      })
    })

    it('should handle API failures gracefully throughout the journey', async () => {
      const user = userEvent.setup()
      
      // Mock authentication failure
      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.reject(new Error('Network error'))
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      // Should still show landing page on auth error
      await waitFor(() => {
        expect(screen.getByText('Bodda')).toBeInTheDocument()
      })

      // Mock successful auth but session loading failure
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1' }
            })
          })
        }
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: false,
            status: 500,
            json: () => Promise.resolve({ error: 'Server error' })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      // Simulate successful authentication
      const connectButton = screen.getByTestId('strava-connect-button')
      await user.click(connectButton)

      // Should show chat interface with error
      await waitFor(() => {
        expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
        expect(screen.getByText(/HTTP 500/)).toBeInTheDocument()
        expect(screen.getByText('Retry Loading Sessions')).toBeInTheDocument()
      })

      // Should be able to retry
      const retryButton = screen.getByText('Retry Loading Sessions')
      
      // Mock successful retry
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
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      await user.click(retryButton)

      await waitFor(() => {
        expect(screen.getByText('Welcome to Bodda!')).toBeInTheDocument()
        expect(screen.queryByText('Retry Loading Sessions')).not.toBeInTheDocument()
      })
    })

    it('should handle streaming chat responses correctly', async () => {
      const user = userEvent.setup()
      
      // Mock authenticated state
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1' }
            })
          })
        }
        if (url.includes('/api/sessions') && !options?.method) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ messages: [] })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      // Mock EventSource constructor
      const mockEventSourceConstructor = vi.fn().mockImplementation((url, options) => {
        mockEventSource = new MockEventSource(url, options)
        return mockEventSource
      })
      global.EventSource = mockEventSourceConstructor

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText(/Ask your AI coach/)).toBeInTheDocument()
      })
      
      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      const sendButton = screen.getByText('Send')
      
      await user.type(input, 'How should I train for a marathon?')
      await user.click(sendButton)
      
      // Should show user message immediately
      expect(screen.getByText('How should I train for a marathon?')).toBeInTheDocument()
      
      // Should show streaming indicator
      expect(screen.getByText('AI is thinking...')).toBeInTheDocument()
      
      // Simulate streaming response
      mockEventSource.simulateMessage({
        type: 'chunk',
        content: 'For marathon training, you should '
      })
      
      mockEventSource.simulateMessage({
        type: 'chunk', 
        content: 'focus on building your aerobic base...'
      })
      
      mockEventSource.simulateMessage({
        type: 'complete',
        message: {
          id: 'msg-1',
          role: 'assistant',
          content: 'For marathon training, you should focus on building your aerobic base...',
          created_at: new Date().toISOString()
        }
      })
      
      await waitFor(() => {
        expect(screen.getByText(/For marathon training, you should focus on building your aerobic base/)).toBeInTheDocument()
        expect(screen.queryByText('AI is thinking...')).not.toBeInTheDocument()
      })
    })

    it('should handle logout flow correctly', async () => {
      const user = userEvent.setup()
      
      // Mock authenticated state
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1', first_name: 'John' }
            })
          })
        }
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        if (url.includes('/auth/logout') && options?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ message: 'Logged out' })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      // Should show chat interface
      await waitFor(() => {
        expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
      })

      // Mock logout response and subsequent auth check
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/auth/logout') && options?.method === 'POST') {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ message: 'Logged out' })
          })
        }
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: false,
            status: 401,
            json: () => Promise.resolve({ error: 'Not authenticated' })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      // Note: In a real app, there would be a logout button. 
      // For this test, we'll simulate the logout API call directly
      await apiClient.logout()

      // Should redirect to landing page after logout
      await waitFor(() => {
        expect(screen.getByText('Bodda')).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling and Recovery', () => {
    it('should show retry button on API failures', async () => {
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
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: false,
            status: 500,
            json: () => Promise.resolve({ error: 'Server error' })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText(/HTTP 500/)).toBeInTheDocument()
        expect(screen.getByText('Retry Loading Sessions')).toBeInTheDocument()
      })
    })

    it('should retry failed requests when retry button is clicked', async () => {
      const user = userEvent.setup()
      let callCount = 0
      
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
        if (url.includes('/api/sessions')) {
          callCount++
          if (callCount === 1) {
            return Promise.resolve({
              ok: false,
              status: 500,
              json: () => Promise.resolve({ error: 'Server error' })
            })
          } else {
            return Promise.resolve({
              ok: true,
              json: () => Promise.resolve({ sessions: [] })
            })
          }
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText('Retry Loading Sessions')).toBeInTheDocument()
      })
      
      await user.click(screen.getByText('Retry Loading Sessions'))
      
      await waitFor(() => {
        expect(screen.getByText('Welcome to Bodda!')).toBeInTheDocument()
        expect(screen.queryByText('Retry Loading Sessions')).not.toBeInTheDocument()
      })
    })

    it('should handle network connectivity issues', async () => {
      mockFetch.mockImplementation(() => {
        return Promise.reject(new Error('Network error'))
      })

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText(/Connection failed/)).toBeInTheDocument()
      })
    })
  })

  describe('Loading States', () => {
    it('should show loading indicators during API calls', async () => {
      const user = userEvent.setup()
      
      // Mock slow API response
      mockFetch.mockImplementation((url: string, options?: RequestInit) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1' }
            })
          })
        }
        if (url.includes('/api/sessions') && options?.method === 'POST') {
          return new Promise(resolve => {
            setTimeout(() => {
              resolve({
                ok: true,
                json: () => Promise.resolve({
                  session: { id: 'new-session', title: 'New Session' }
                })
              })
            }, 1000)
          })
        }
        if (url.includes('/api/sessions')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        return Promise.resolve({ ok: true, json: () => Promise.resolve({}) })
      })

      render(<App />, { wrapper: TestWrapper })
      
      await waitFor(() => {
        expect(screen.getByText('New Session')).toBeInTheDocument()
      })
      
      await user.click(screen.getByText('New Session'))
      
      // Should show loading state
      expect(screen.getByText('Creating...')).toBeInTheDocument()
      
      await waitFor(() => {
        expect(screen.queryByText('Creating...')).not.toBeInTheDocument()
      }, { timeout: 2000 })
    })
  })
})