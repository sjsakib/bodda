import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import IntegrationDemo from './IntegrationDemo'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('Integration Demo', () => {
  beforeEach(() => {
    mockFetch.mockClear()
    vi.clearAllMocks()
  })

  it('should show authentication required when not authenticated', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'Not authenticated' })
    })

    render(<IntegrationDemo />)

    await waitFor(() => {
      expect(screen.getByText('Authentication Required')).toBeInTheDocument()
      expect(screen.getByText('Connect with Strava')).toBeInTheDocument()
    })
  })

  it('should show demo interface when authenticated', async () => {
    // Mock successful authentication
    mockFetch.mockImplementation((url: string) => {
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
          json: () => Promise.resolve({
            sessions: [
              {
                id: 'session-1',
                title: 'Demo Session',
                created_at: '2024-01-01T10:00:00Z'
              }
            ]
          })
        })
      }
      if (url.includes('/messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ messages: [] })
        })
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      })
    })

    render(<IntegrationDemo />)

    await waitFor(() => {
      expect(screen.getByText('API Integration Demo')).toBeInTheDocument()
      expect(screen.getByText('Sessions')).toBeInTheDocument()
      expect(screen.getByText('Messages')).toBeInTheDocument()
      expect(screen.getByText('Demo Session')).toBeInTheDocument()
    })

    // Should show API status
    expect(screen.getByText('✓ Connected')).toBeInTheDocument()
    expect(screen.getByText('1 loaded')).toBeInTheDocument()
  })

  it('should handle creating new sessions', async () => {
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
      if (url.includes('/api/sessions') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            session: {
              id: 'new-session',
              title: 'Demo Session',
              created_at: new Date().toISOString()
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

    render(<IntegrationDemo />)

    await waitFor(() => {
      expect(screen.getByText('New Session')).toBeInTheDocument()
    })

    await user.click(screen.getByText('New Session'))

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/sessions',
        expect.objectContaining({ method: 'POST' })
      )
    })
  })

  it('should handle sending messages', async () => {
    const user = userEvent.setup()
    
    // Mock authenticated state with session
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
      if (url.includes('/api/sessions') && !url.includes('messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            sessions: [{
              id: 'session-1',
              title: 'Test Session',
              created_at: '2024-01-01T10:00:00Z'
            }]
          })
        })
      }
      if (url.includes('/messages') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            user_message: {
              id: 'msg-1',
              content: 'Hello',
              role: 'user',
              created_at: new Date().toISOString()
            },
            assistant_message: {
              id: 'msg-2',
              content: 'Hi there!',
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
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({})
      })
    })

    render(<IntegrationDemo />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Type your message...')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText('Type your message...')
    await user.type(input, 'Hello')
    await user.click(screen.getByText('Send'))

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/sessions/session-1/messages',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ content: 'Hello' })
        })
      )
    })
  })
})