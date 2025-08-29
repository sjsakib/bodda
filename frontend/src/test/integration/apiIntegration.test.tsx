import { describe, it, expect, beforeEach, vi } from 'vitest'
import { apiClient } from '../../services/api'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock EventSource
class MockEventSource {
  url: string
  withCredentials?: boolean
  
  constructor(url: string, options?: { withCredentials?: boolean }) {
    this.url = url
    this.withCredentials = options?.withCredentials
  }
  
  close() {}
}

global.EventSource = MockEventSource as any

describe('API Integration Tests', () => {
  beforeEach(() => {
    mockFetch.mockClear()
  })

  describe('Authentication Integration', () => {
    it('should handle authentication check correctly', async () => {
      mockFetch.mockResolvedValueOnce({
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

      const result = await apiClient.checkAuth()
      
      expect(result.authenticated).toBe(true)
      expect(result.user.id).toBe('user-1')
      expect(mockFetch).toHaveBeenCalledWith('/api/auth/check', {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })

    it('should handle unauthenticated state correctly', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Not authenticated' })
      })

      const result = await apiClient.checkAuth()
      
      expect(result.authenticated).toBe(false)
      expect(result.user).toBeNull()
    })

    it('should redirect to Strava OAuth', () => {
      const originalLocation = window.location
      delete (window as any).location
      Object.defineProperty(window, 'location', {
        value: { ...originalLocation, href: '' },
        writable: true
      })

      apiClient.redirectToStravaAuth()
      
      expect(window.location.href).toBe('/auth/strava')
      
      // Restore
      Object.defineProperty(window, 'location', {
        value: originalLocation,
        writable: true
      })
    })
  })

  describe('Session Management Integration', () => {
    it('should create and manage sessions', async () => {
      // Mock create session
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          session: {
            id: 'session-1',
            title: 'New Session',
            created_at: '2024-01-01T10:00:00Z'
          }
        })
      })

      const session = await apiClient.createSession('New Session')
      
      expect(session.id).toBe('session-1')
      expect(session.title).toBe('New Session')
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ title: 'New Session' })
      })
    })

    it('should get sessions list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          sessions: [
            { id: 'session-1', title: 'Session 1' },
            { id: 'session-2', title: 'Session 2' }
          ]
        })
      })

      const sessions = await apiClient.getSessions()
      
      expect(sessions).toHaveLength(2)
      expect(sessions[0].id).toBe('session-1')
    })

    it('should get messages with pagination', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          messages: [
            { id: 'msg-1', content: 'Hello', role: 'user' },
            { id: 'msg-2', content: 'Hi there!', role: 'assistant' }
          ]
        })
      })

      const messages = await apiClient.getMessages('session-1', 10, 0)
      
      expect(messages).toHaveLength(2)
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/sessions/session-1/messages?limit=10&offset=0',
        expect.objectContaining({
          credentials: 'include'
        })
      )
    })
  })

  describe('Error Handling Integration', () => {
    it('should provide user-friendly error messages', () => {
      const apiError = { status: 500, message: 'Server error' }
      const message = apiClient.getErrorMessage(apiError)
      expect(message).toBe('An unexpected error occurred')
      
      const networkError = new Error('Network failed')
      const networkMessage = apiClient.getErrorMessage(networkError)
      expect(networkMessage).toBe('Network failed')
    })

    it('should identify retryable errors correctly', () => {
      const networkError = new Error('Network error')
      expect(apiClient.isRetryableError(networkError)).toBe(false) // Network errors are handled by fetch retry

      const serverError = { status: 500, message: 'Server error' }
      expect(apiClient.isRetryableError(serverError)).toBe(false) // Not an ApiError instance
    })
  })

  describe('Streaming Integration', () => {
    it('should create EventSource for streaming', () => {
      const eventSource = apiClient.createEventSource('session-1', 'Hello')
      
      expect(eventSource).toBeInstanceOf(EventSource)
      expect(eventSource.url).toContain('/api/sessions/session-1/stream')
      expect(eventSource.url).toContain('message=Hello')
      
      eventSource.close()
    })
  })
})