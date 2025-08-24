import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { apiClient } from '../../services/api'
import ChatInterface from '../../components/ChatInterface'

// Mock fetch globally
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

// Test wrapper with router
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>
}

describe('Basic Integration Tests', () => {
  beforeEach(() => {
    mockFetch.mockClear()
    vi.clearAllMocks()
  })

  it('should integrate authentication and session loading', async () => {
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
      if (url.includes('/api/sessions') && !url.includes('messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            sessions: [
              {
                id: 'session-1',
                title: 'Test Session',
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

    render(
      <TestWrapper>
        <ChatInterface />
      </TestWrapper>
    )

    // Should eventually show the chat interface
    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
    }, { timeout: 3000 })

    // Should show session in sidebar
    await waitFor(() => {
      expect(screen.getByText('Test Session')).toBeInTheDocument()
    })

    // Should show message input
    expect(screen.getByPlaceholderText(/Ask your AI coach/)).toBeInTheDocument()
  })

  it('should handle API client error responses correctly', async () => {
    const error = await apiClient.getErrorMessage({ status: 404, message: 'Not found' })
    expect(error).toBe('The requested resource was not found')

    const networkError = await apiClient.getErrorMessage(new Error('Network failed'))
    expect(networkError).toBe('Network failed')
  })

  it('should create EventSource for streaming correctly', () => {
    const eventSource = apiClient.createEventSource('session-1', 'Hello')
    
    expect(eventSource).toBeInstanceOf(EventSource)
    expect(eventSource.url).toContain('/api/sessions/session-1/stream')
    expect(eventSource.url).toContain('message=Hello')
    
    eventSource.close()
  })
})