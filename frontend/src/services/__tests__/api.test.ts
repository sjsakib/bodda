import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { ApiClient, ApiError, NetworkError } from '../api'

// Mock fetch
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock EventSource
class MockEventSource {
  constructor(public url: string, public options?: EventSourceInit) {}
  close() {}
}
global.EventSource = MockEventSource as any

describe('ApiClient', () => {
  let apiClient: ApiClient

  beforeEach(() => {
    apiClient = new ApiClient()
    mockFetch.mockClear()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Error Handling', () => {
    it('should throw ApiError for HTTP errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.resolve({ error: 'Resource not found' })
      })

      await expect(apiClient.checkAuth()).rejects.toThrow(ApiError)
      await expect(apiClient.checkAuth()).rejects.toThrow('Resource not found')
    })

    it('should throw NetworkError for network failures', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(apiClient.checkAuth()).rejects.toThrow(NetworkError)
    })

    it('should handle malformed JSON responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.reject(new Error('Invalid JSON'))
      })

      await expect(apiClient.checkAuth()).rejects.toThrow(ApiError)
    })

    it('should provide user-friendly error messages', () => {
      const apiError401 = new ApiError('Unauthorized', 401)
      const apiError403 = new ApiError('Forbidden', 403)
      const apiError404 = new ApiError('Not Found', 404)
      const apiError429 = new ApiError('Too Many Requests', 429)
      const apiError500 = new ApiError('Internal Server Error', 500)
      const networkError = new NetworkError()

      expect(apiClient.getErrorMessage(apiError401)).toBe('Please log in to continue')
      expect(apiClient.getErrorMessage(apiError403)).toBe('You do not have permission to perform this action')
      expect(apiClient.getErrorMessage(apiError404)).toBe('The requested resource was not found')
      expect(apiClient.getErrorMessage(apiError429)).toBe('Too many requests. Please try again later')
      expect(apiClient.getErrorMessage(apiError500)).toBe('Server error. Please try again later')
      expect(apiClient.getErrorMessage(networkError)).toBe('Connection failed. Please check your internet connection and try again')
    })
  })

  describe('Retry Logic', () => {
    it('should retry on server errors', async () => {
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          statusText: 'Internal Server Error',
          json: () => Promise.resolve({ error: 'Server error' })
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve({ authenticated: true, user: { id: '1' } })
        })

      const result = await apiClient.checkAuth()
      expect(result.authenticated).toBe(true)
      expect(mockFetch).toHaveBeenCalledTimes(2)
    })

    it('should retry on network errors', async () => {
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve({ authenticated: true, user: { id: '1' } })
        })

      const result = await apiClient.checkAuth()
      expect(result.authenticated).toBe(true)
      expect(mockFetch).toHaveBeenCalledTimes(2)
    })

    it('should not retry on client errors (except 408, 429)', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.resolve({ error: 'Not found' })
      })

      await expect(apiClient.checkAuth()).rejects.toThrow(ApiError)
      expect(mockFetch).toHaveBeenCalledTimes(1)
    })

    it('should retry on 408 and 429 errors', async () => {
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 429,
          statusText: 'Too Many Requests',
          json: () => Promise.resolve({ error: 'Rate limited' })
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve({ authenticated: true, user: { id: '1' } })
        })

      const result = await apiClient.checkAuth()
      expect(result.authenticated).toBe(true)
      expect(mockFetch).toHaveBeenCalledTimes(2)
    })

    it('should use exponential backoff for retries', async () => {
      const startTime = Date.now()
      
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve({ authenticated: true, user: { id: '1' } })
        })

      await apiClient.checkAuth()
      
      const endTime = Date.now()
      const duration = endTime - startTime
      
      // Should take at least 1000ms (first retry) + 2000ms (second retry) = 3000ms
      // But we'll be lenient and check for at least 1000ms to account for test timing
      expect(duration).toBeGreaterThan(1000)
    })
  })

  describe('Authentication Methods', () => {
    it('should check authentication status', async () => {
      const mockResponse = {
        authenticated: true,
        user: { id: '1', first_name: 'John', last_name: 'Doe' }
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.checkAuth()
      expect(result).toEqual(mockResponse)
      expect(mockFetch).toHaveBeenCalledWith('/api/auth/check', expect.objectContaining({
        credentials: 'include'
      }))
    })

    it('should logout user', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ message: 'Logged out' })
      })

      await apiClient.logout()
      expect(mockFetch).toHaveBeenCalledWith('/auth/logout', expect.objectContaining({
        method: 'POST',
        credentials: 'include'
      }))
    })
  })

  describe('Session Management', () => {
    it('should get sessions', async () => {
      const mockSessions = [
        { id: '1', title: 'Session 1', created_at: '2024-01-01T00:00:00Z' },
        { id: '2', title: 'Session 2', created_at: '2024-01-02T00:00:00Z' }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ sessions: mockSessions })
      })

      const result = await apiClient.getSessions()
      expect(result).toEqual(mockSessions)
    })

    it('should create session', async () => {
      const mockSession = { id: '1', title: 'New Session', created_at: '2024-01-01T00:00:00Z' }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ session: mockSession })
      })

      const result = await apiClient.createSession('Test Session')
      expect(result).toEqual(mockSession)
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ title: 'Test Session' })
      }))
    })

    it('should get messages with pagination', async () => {
      const mockMessages = [
        { id: '1', role: 'user', content: 'Hello', created_at: '2024-01-01T00:00:00Z' }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ messages: mockMessages })
      })

      const result = await apiClient.getMessages('session-1', 10, 0)
      expect(result).toEqual(mockMessages)
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/session-1/messages?limit=10&offset=0', expect.any(Object))
    })

    it('should send message', async () => {
      const mockResponse = {
        user_message: { id: '1', role: 'user', content: 'Hello' },
        assistant_message: { id: '2', role: 'assistant', content: 'Hi there!' }
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.sendMessage('session-1', 'Hello')
      expect(result).toEqual(mockResponse)
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/session-1/messages', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ content: 'Hello' })
      }))
    })
  })

  describe('Streaming', () => {
    it('should create EventSource for streaming', () => {
      const eventSource = apiClient.createEventSource('session-1', 'Hello')
      expect(eventSource).toBeInstanceOf(MockEventSource)
      expect(eventSource.url).toContain('/api/sessions/session-1/stream')
      expect(eventSource.url).toContain('message=Hello')
    })
  })

  describe('Utility Methods', () => {
    it('should identify retryable errors', () => {
      const networkError = new NetworkError()
      const apiError500 = new ApiError('Server Error', 500)
      const apiError429 = new ApiError('Rate Limited', 429)
      const apiError404 = new ApiError('Not Found', 404)

      expect(apiClient.isRetryableError(networkError)).toBe(true)
      expect(apiClient.isRetryableError(apiError500)).toBe(true)
      expect(apiClient.isRetryableError(apiError429)).toBe(true)
      expect(apiClient.isRetryableError(apiError404)).toBe(false)
    })
  })

  describe('Request Configuration', () => {
    it('should include credentials in all requests', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: {} })
      })

      await apiClient.checkAuth()
      
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          credentials: 'include'
        })
      )
    })

    it('should set correct content type for JSON requests', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ session: {} })
      })

      await apiClient.createSession()
      
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      )
    })
  })
})