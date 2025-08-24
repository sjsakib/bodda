import { apiClient } from '../services/api'
import { RetryButton } from './ErrorBoundary'

interface ApiErrorHandlerProps {
  error: unknown
  onRetry?: () => void
  onDismiss?: () => void
  className?: string
  showRetryButton?: boolean
}

export default function ApiErrorHandler({ 
  error, 
  onRetry, 
  onDismiss, 
  className = '',
  showRetryButton = true 
}: ApiErrorHandlerProps) {
  if (!error) return null

  const errorMessage = apiClient.getErrorMessage(error)
  const isRetryable = apiClient.isRetryableError(error)

  return (
    <div className={`bg-red-50 border border-red-200 rounded-lg p-4 ${className}`}>
      <div className="flex items-start justify-between">
        <div className="flex items-start flex-1">
          <svg
            className="w-5 h-5 text-red-500 mt-0.5 mr-3 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
          <div className="flex-1">
            <p className="text-red-800 text-sm font-medium">{errorMessage}</p>
            {showRetryButton && isRetryable && onRetry && (
              <div className="mt-3">
                <RetryButton onRetry={onRetry} className="text-sm">
                  Try Again
                </RetryButton>
              </div>
            )}
          </div>
        </div>
        {onDismiss && (
          <button
            onClick={onDismiss}
            className="text-red-500 hover:text-red-700 ml-2"
            aria-label="Dismiss error"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>
    </div>
  )
}

// Specialized error handler for session loading
interface SessionErrorHandlerProps {
  error: unknown
  onRetry: () => void
  loading?: boolean
}

export function SessionErrorHandler({ error, onRetry, loading = false }: SessionErrorHandlerProps) {
  if (!error) return null

  return (
    <div className="text-center py-8">
      <div className="max-w-md mx-auto">
        <svg
          className="w-12 h-12 text-red-500 mx-auto mb-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
          />
        </svg>
        <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to Load Sessions</h3>
        <p className="text-gray-600 mb-4">{apiClient.getErrorMessage(error)}</p>
        <RetryButton onRetry={onRetry} loading={loading}>
          Retry Loading Sessions
        </RetryButton>
      </div>
    </div>
  )
}

// Specialized error handler for message loading
interface MessageErrorHandlerProps {
  error: unknown
  onRetry: () => void
  loading?: boolean
  loadMessage?: string
}

export function MessageErrorHandler({ error, onRetry, loading = false }: MessageErrorHandlerProps) {
  if (!error) return null

  return (
    <div className="text-center py-8">
      <div className="max-w-md mx-auto">
        <svg
          className="w-8 h-8 text-red-500 mx-auto mb-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>
        <h4 className="text-md font-medium text-gray-900 mb-2">Failed to Load Messages</h4>
        <p className="text-gray-600 text-sm mb-3">{apiClient.getErrorMessage(error)}</p>
        <RetryButton onRetry={onRetry} loading={loading} className="text-sm">
          Retry Loading Messages
        </RetryButton>
      </div>
    </div>
  )
}