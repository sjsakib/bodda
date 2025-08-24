import { useState, useCallback } from 'react'
import { apiClient, ApiError, NetworkError } from '../services/api'

export interface ApiState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

export interface UseApiResult<T> extends ApiState<T> {
  execute: (...args: any[]) => Promise<T | null>
  reset: () => void
  retry: () => Promise<T | null>
}

export function useApi<T>(
  apiFunction: (...args: any[]) => Promise<T>,
  initialData: T | null = null
): UseApiResult<T> {
  const [state, setState] = useState<ApiState<T>>({
    data: initialData,
    loading: false,
    error: null,
  })
  
  const [lastArgs, setLastArgs] = useState<any[]>([])

  const execute = useCallback(async (...args: any[]): Promise<T | null> => {
    setState(prev => ({ ...prev, loading: true, error: null }))
    setLastArgs(args)

    try {
      const result = await apiFunction(...args)
      setState({ data: result, loading: false, error: null })
      return result
    } catch (error) {
      const errorMessage = apiClient.getErrorMessage(error)
      setState(prev => ({ ...prev, loading: false, error: errorMessage }))
      return null
    }
  }, [apiFunction])

  const retry = useCallback(async (): Promise<T | null> => {
    if (lastArgs.length === 0) {
      return null
    }
    return execute(...lastArgs)
  }, [execute, lastArgs])

  const reset = useCallback(() => {
    setState({ data: initialData, loading: false, error: null })
    setLastArgs([])
  }, [initialData])

  return {
    ...state,
    execute,
    retry,
    reset,
  }
}

// Specialized hook for authentication
export function useAuth() {
  const [authState, setAuthState] = useState<{
    user: any | null
    loading: boolean
    error: string | null
    authenticated: boolean
    initialized: boolean
  }>({
    user: null,
    loading: false,
    error: null,
    authenticated: false,
    initialized: false,
  })

  const checkAuth = useCallback(async () => {
    setAuthState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const response = await apiClient.checkAuth()
      setAuthState({
        user: response.user,
        loading: false,
        error: null,
        authenticated: response.authenticated,
        initialized: true,
      })
      return response
    } catch (error) {
      const errorMessage = apiClient.getErrorMessage(error)
      setAuthState({
        user: null,
        loading: false,
        error: errorMessage,
        authenticated: false,
        initialized: true,
      })
      return null
    }
  }, [])

  const logout = useCallback(async () => {
    setAuthState(prev => ({ ...prev, loading: true, error: null }))

    try {
      await apiClient.logout()
      setAuthState({
        user: null,
        loading: false,
        error: null,
        authenticated: false,
        initialized: true,
      })
      return true
    } catch (error) {
      const errorMessage = apiClient.getErrorMessage(error)
      setAuthState(prev => ({ ...prev, loading: false, error: errorMessage }))
      return false
    }
  }, [])

  const clearError = useCallback(() => {
    setAuthState(prev => ({ ...prev, error: null }))
  }, [])

  return {
    ...authState,
    checkAuth,
    logout,
    clearError,
  }
}

// Hook for managing loading states across multiple operations
export function useLoadingState() {
  const [loadingStates, setLoadingStates] = useState<Record<string, boolean>>({})

  const setLoading = useCallback((key: string, loading: boolean) => {
    setLoadingStates(prev => ({
      ...prev,
      [key]: loading,
    }))
  }, [])

  const isLoading = useCallback((key?: string) => {
    if (key) {
      return loadingStates[key] || false
    }
    return Object.values(loadingStates).some(loading => loading)
  }, [loadingStates])

  return {
    setLoading,
    isLoading,
    loadingStates,
  }
}