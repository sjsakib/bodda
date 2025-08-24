// API client with error handling, retry logic, and loading states
export interface User {
  id: string
  strava_id: number
  first_name: string
  last_name: string
}

export interface Session {
  id: string
  user_id: string
  title: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  session_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export interface AuthResponse {
  authenticated: boolean
  user: User
}

export interface SessionsResponse {
  sessions: Session[]
}

export interface MessagesResponse {
  messages: Message[]
}

export interface SendMessageResponse {
  user_message: Message
  assistant_message: Message
}

export interface CreateSessionResponse {
  session: Session
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export class NetworkError extends Error {
  constructor(message: string = 'Network connection failed') {
    super(message)
    this.name = 'NetworkError'
  }
}

class ApiClient {
  private baseUrl: string
  private defaultRetries: number = 3
  private defaultRetryDelay: number = 1000

  constructor(baseUrl: string = '') {
    this.baseUrl = baseUrl
  }

  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  private async fetchWithRetry(
    url: string,
    options: RequestInit = {},
    retries: number = this.defaultRetries
  ): Promise<Response> {
    const fullUrl = `${this.baseUrl}${url}`
    
    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        const response = await fetch(fullUrl, {
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            ...options.headers,
          },
          ...options,
        })

        // Don't retry on client errors (4xx) except 408, 429
        if (response.status >= 400 && response.status < 500) {
          if (response.status !== 408 && response.status !== 429) {
            return response
          }
        }

        // Return successful responses or server errors on last attempt
        if (response.ok || attempt === retries) {
          return response
        }

        // Wait before retrying
        if (attempt < retries) {
          const delay = this.defaultRetryDelay * Math.pow(2, attempt) // Exponential backoff
          await this.delay(delay)
        }
      } catch (error) {
        // Network error - retry unless it's the last attempt
        if (attempt === retries) {
          throw new NetworkError('Failed to connect to server')
        }
        
        // Wait before retrying
        const delay = this.defaultRetryDelay * Math.pow(2, attempt)
        await this.delay(delay)
      }
    }

    throw new NetworkError('Max retries exceeded')
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let errorMessage = `HTTP ${response.status}: ${response.statusText}`
      let errorCode: string | undefined

      try {
        const errorData = await response.json()
        errorMessage = errorData.message || errorData.error || errorMessage
        errorCode = errorData.code
      } catch {
        // If we can't parse the error response, use the default message
      }

      throw new ApiError(errorMessage, response.status, errorCode)
    }

    try {
      return await response.json()
    } catch (error) {
      throw new ApiError('Invalid response format', response.status)
    }
  }

  // Authentication methods
  async checkAuth(): Promise<AuthResponse> {
    try {
      const response = await this.fetchWithRetry('/api/auth/check', {}, 0) // No retries for auth check
      const result = await this.handleResponse<AuthResponse>(response)
      
      // Ensure we have a valid response structure
      if (!result || typeof result.authenticated !== 'boolean') {
        return { 
          authenticated: false, 
          user: {
            id: '',
            strava_id: 0,
            first_name: '',
            last_name: ''
          }
        }
      }
      
      return result
    } catch (error) {
      // If auth check fails, assume not authenticated
      if (error instanceof ApiError && error.status === 401) {
        return { 
          authenticated: false, 
          user: {
            id: '',
            strava_id: 0,
            first_name: '',
            last_name: ''
          }
        }
      }
      throw error
    }
  }

  async logout(): Promise<void> {
    const response = await this.fetchWithRetry('/auth/logout', {
      method: 'POST',
    })
    await this.handleResponse(response)
  }

  // OAuth redirect method
  redirectToStravaAuth(): void {
    window.location.href = '/auth/strava'
  }

  // Session management methods
  async getSessions(): Promise<Session[]> {
    const response = await this.fetchWithRetry('/api/sessions')
    const data = await this.handleResponse<SessionsResponse>(response)
    return data?.sessions || []
  }

  async createSession(title?: string): Promise<Session> {
    const response = await this.fetchWithRetry('/api/sessions', {
      method: 'POST',
      body: JSON.stringify({ title: title || '' }),
    })
    const data = await this.handleResponse<CreateSessionResponse>(response)
    
    if (!data?.session) {
      throw new ApiError('Invalid session response from server', response.status)
    }
    
    return data.session
  }

  async getMessages(sessionId: string, limit?: number, offset?: number): Promise<Message[]> {
    const params = new URLSearchParams()
    if (limit !== undefined) params.append('limit', limit.toString())
    if (offset !== undefined) params.append('offset', offset.toString())
    
    const queryString = params.toString()
    const url = `/api/sessions/${sessionId}/messages${queryString ? `?${queryString}` : ''}`
    
    const response = await this.fetchWithRetry(url)
    const data = await this.handleResponse<MessagesResponse>(response)
    return data?.messages || []
  }

  async sendMessage(sessionId: string, content: string): Promise<SendMessageResponse> {
    const response = await this.fetchWithRetry(`/api/sessions/${sessionId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    })
    return this.handleResponse<SendMessageResponse>(response)
  }

  // Streaming methods
  createEventSource(sessionId: string, message: string): EventSource {
    const params = new URLSearchParams({ message })
    const url = `${this.baseUrl}/api/sessions/${sessionId}/stream?${params.toString()}`
    
    return new EventSource(url, {
      withCredentials: true,
    })
  }



  // Method to get user-friendly error messages
  getErrorMessage(error: unknown): string {
    if (error instanceof ApiError) {
      switch (error.status) {
        case 400:
          return 'Invalid request. Please check your input and try again'
        case 401:
          return 'Please log in to continue'
        case 403:
          return 'You do not have permission to perform this action'
        case 404:
          return 'The requested resource was not found'
        case 408:
          return 'Request timeout. Please try again'
        case 429:
          return 'Too many requests. Please try again later'
        case 500:
          return 'Server error. Please try again later'
        case 502:
          return 'Service temporarily unavailable. Please try again later'
        case 503:
          return 'Service temporarily unavailable. Please try again later'
        case 504:
          return 'Request timeout. Please try again later'
        default:
          return error.message || `HTTP ${error.status}: An error occurred`
      }
    }
    
    if (error instanceof NetworkError) {
      return 'Connection failed. Please check your internet connection and try again'
    }
    
    if (error instanceof Error) {
      return error.message
    }
    
    return 'An unexpected error occurred'
  }

  // Method to determine if an error should show a retry button
  isRetryableError(error: unknown): boolean {
    if (error instanceof NetworkError) {
      return true
    }
    
    if (error instanceof ApiError) {
      // Retry on server errors and specific client errors
      return error.status >= 500 || error.status === 408 || error.status === 429
    }
    
    return false
  }
}

// Create and export a singleton instance
export const apiClient = new ApiClient()

// Export types and classes for use in components
export { ApiClient }