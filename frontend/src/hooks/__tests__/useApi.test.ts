import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useApi, useAuth, useLoadingState } from '../useApi'
import { apiClient } from '../../services/api'

// Mock the API client
vi.mock('../../services/api', () => ({
  apiClient: {
    checkAuth: vi.fn(),
    logout: vi.fn(),
    getErrorMessage: vi.fn((error) => error.message || 'Unknown error')
  }
}))

const mockApiClient = apiClient as any

describe('useApi Hook', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize with correct default state', () => {
    const mockApiFunction = vi.fn()
    const { result } = renderHook(() => useApi(mockApiFunction))

    expect(result.current.data).toBe(null)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
  })

  it('should initialize with provided initial data', () => {
    const mockApiFunction = vi.fn()
    const initialData = { test: 'data' }
    const { result } = renderHook(() => useApi(mockApiFunction, initialData))

    expect(result.current.data).toEqual(initialData)
  })

  it('should handle successful API calls', async () => {
    const mockData = { success: true }
    const mockApiFunction = vi.fn().mockResolvedValue(mockData)
    const { result } = renderHook(() => useApi(mockApiFunction))

    await act(async () => {
      const response = await result.current.execute('test-arg')
      expect(response).toEqual(mockData)
    })

    expect(result.current.data).toEqual(mockData)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
    expect(mockApiFunction).toHaveBeenCalledWith('test-arg')
  })

  it('should handle API errors', async () => {
    const mockError = new Error('API Error')
    const mockApiFunction = vi.fn().mockRejectedValue(mockError)
    mockApiClient.getErrorMessage.mockReturnValue('API Error')
    
    const { result } = renderHook(() => useApi(mockApiFunction))

    await act(async () => {
      const response = await result.current.execute()
      expect(response).toBe(null)
    })

    expect(result.current.data).toBe(null)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('API Error')
  })

  it('should set loading state during API calls', async () => {
    const mockApiFunction = vi.fn().mockImplementation(() => 
      new Promise(resolve => setTimeout(() => resolve({ data: 'test' }), 100))
    )
    const { result } = renderHook(() => useApi(mockApiFunction))

    act(() => {
      result.current.execute()
    })

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
  })

  it('should retry with last arguments', async () => {
    const mockData = { success: true }
    const mockApiFunction = vi.fn().mockResolvedValue(mockData)
    const { result } = renderHook(() => useApi(mockApiFunction))

    // Execute with arguments
    await act(async () => {
      await result.current.execute('arg1', 'arg2')
    })

    // Retry should use same arguments
    await act(async () => {
      await result.current.retry()
    })

    expect(mockApiFunction).toHaveBeenCalledTimes(2)
    expect(mockApiFunction).toHaveBeenNthCalledWith(1, 'arg1', 'arg2')
    expect(mockApiFunction).toHaveBeenNthCalledWith(2, 'arg1', 'arg2')
  })

  it('should reset state', () => {
    const mockApiFunction = vi.fn()
    const initialData = { initial: true }
    const { result } = renderHook(() => useApi(mockApiFunction, initialData))

    // Set some state
    act(() => {
      result.current.execute()
    })

    // Reset
    act(() => {
      result.current.reset()
    })

    expect(result.current.data).toEqual(initialData)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
  })
})

describe('useAuth Hook', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize with correct default state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.user).toBe(null)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
    expect(result.current.authenticated).toBe(false)
  })

  it('should handle successful authentication check', async () => {
    const mockAuthResponse = {
      authenticated: true,
      user: { id: '1', first_name: 'John', last_name: 'Doe' }
    }
    mockApiClient.checkAuth.mockResolvedValue(mockAuthResponse)

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      const response = await result.current.checkAuth()
      expect(response).toEqual(mockAuthResponse)
    })

    expect(result.current.user).toEqual(mockAuthResponse.user)
    expect(result.current.authenticated).toBe(true)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
  })

  it('should handle authentication failure', async () => {
    const mockError = new Error('Unauthorized')
    mockApiClient.checkAuth.mockRejectedValue(mockError)
    mockApiClient.getErrorMessage.mockReturnValue('Unauthorized')

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      const response = await result.current.checkAuth()
      expect(response).toBe(null)
    })

    expect(result.current.user).toBe(null)
    expect(result.current.authenticated).toBe(false)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('Unauthorized')
  })

  it('should handle successful logout', async () => {
    mockApiClient.logout.mockResolvedValue(undefined)

    const { result } = renderHook(() => useAuth())

    // Set initial authenticated state
    await act(async () => {
      mockApiClient.checkAuth.mockResolvedValue({
        authenticated: true,
        user: { id: '1' }
      })
      await result.current.checkAuth()
    })

    // Logout
    await act(async () => {
      const success = await result.current.logout()
      expect(success).toBe(true)
    })

    expect(result.current.user).toBe(null)
    expect(result.current.authenticated).toBe(false)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
  })

  it('should handle logout failure', async () => {
    const mockError = new Error('Logout failed')
    mockApiClient.logout.mockRejectedValue(mockError)
    mockApiClient.getErrorMessage.mockReturnValue('Logout failed')

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      const success = await result.current.logout()
      expect(success).toBe(false)
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('Logout failed')
  })

  it('should set loading state during auth operations', async () => {
    mockApiClient.checkAuth.mockImplementation(() => 
      new Promise(resolve => setTimeout(() => resolve({ authenticated: true, user: {} }), 100))
    )

    const { result } = renderHook(() => useAuth())

    act(() => {
      result.current.checkAuth()
    })

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
  })
})

describe('useLoadingState Hook', () => {
  it('should initialize with empty loading states', () => {
    const { result } = renderHook(() => useLoadingState())

    expect(result.current.loadingStates).toEqual({})
    expect(result.current.isLoading()).toBe(false)
    expect(result.current.isLoading('test')).toBe(false)
  })

  it('should set and get loading states by key', () => {
    const { result } = renderHook(() => useLoadingState())

    act(() => {
      result.current.setLoading('test1', true)
      result.current.setLoading('test2', false)
    })

    expect(result.current.isLoading('test1')).toBe(true)
    expect(result.current.isLoading('test2')).toBe(false)
    expect(result.current.isLoading()).toBe(true) // Any loading
  })

  it('should return true for isLoading() when any key is loading', () => {
    const { result } = renderHook(() => useLoadingState())

    act(() => {
      result.current.setLoading('test1', false)
      result.current.setLoading('test2', true)
      result.current.setLoading('test3', false)
    })

    expect(result.current.isLoading()).toBe(true)
  })

  it('should return false for isLoading() when no keys are loading', () => {
    const { result } = renderHook(() => useLoadingState())

    act(() => {
      result.current.setLoading('test1', false)
      result.current.setLoading('test2', false)
    })

    expect(result.current.isLoading()).toBe(false)
  })

  it('should update loading states independently', () => {
    const { result } = renderHook(() => useLoadingState())

    act(() => {
      result.current.setLoading('api1', true)
      result.current.setLoading('api2', true)
    })

    expect(result.current.isLoading('api1')).toBe(true)
    expect(result.current.isLoading('api2')).toBe(true)

    act(() => {
      result.current.setLoading('api1', false)
    })

    expect(result.current.isLoading('api1')).toBe(false)
    expect(result.current.isLoading('api2')).toBe(true)
    expect(result.current.isLoading()).toBe(true)
  })
})