
import { Session } from '../services/api'
import { LoadingSpinner } from './ErrorBoundary'

import { SessionErrorHandler } from './ApiErrorHandler'

interface SessionSidebarProps {
  sessions: Session[]
  currentSessionId?: string
  onCreateSession: () => void
  isCreatingSession: boolean
  onSelectSession: (sessionId: string) => void
  isLoading?: boolean
  error?: unknown
  onRetryLoad?: () => void
}

export default function SessionSidebar({
  sessions,
  currentSessionId,
  onCreateSession,
  isCreatingSession,
  onSelectSession,
  isLoading = false,
  error = null,
  onRetryLoad
}: SessionSidebarProps) {
  return (
    <div className="w-1/4 bg-white border-r border-gray-200 flex flex-col">
      <div className="p-4 border-b border-gray-200">
        <button
          onClick={onCreateSession}
          disabled={isCreatingSession}
          className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition-colors flex items-center justify-center"
        >
          {isCreatingSession ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Creating...
            </>
          ) : (
            'New Session'
          )}
        </button>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        <div className="p-2">
          <h2 className="text-sm font-semibold text-gray-600 mb-2 px-2">Recent Sessions</h2>
          
          {error && onRetryLoad ? (
            <SessionErrorHandler 
              error={error} 
              onRetry={onRetryLoad}
              loading={isLoading}
            />
          ) : isLoading ? (
            <div className="text-center py-8">
              <LoadingSpinner className="mx-auto mb-2" />
              <p className="text-gray-600 text-sm">Loading sessions...</p>
            </div>
          ) : !sessions || sessions.length === 0 ? (
            <div className="text-center text-gray-500 py-8">
              <p>No sessions yet</p>
              <p className="text-sm">Start a new conversation!</p>
            </div>
          ) : (
            <div className="space-y-1">
              {sessions.map((session) => (
                <button
                  key={session.id}
                  onClick={() => onSelectSession(session.id)}
                  className={`w-full text-left p-3 rounded-lg transition-colors ${
                    currentSessionId === session.id
                      ? 'bg-blue-100 text-blue-900'
                      : 'hover:bg-gray-100'
                  }`}
                >
                  <div className="font-medium text-sm truncate">
                    {session.title}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    {new Date(session.created_at).toLocaleDateString()}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}