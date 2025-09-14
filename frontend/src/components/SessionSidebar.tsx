import { Session } from '../services/api';
import { LoadingSpinner } from './ErrorBoundary';
import { SessionErrorHandler } from './ApiErrorHandler';
import { formatSessionTimestamp } from '../utils/dateFormatting';
import { useState } from 'react';

interface SessionSidebarProps {
  sessions: Session[];
  currentSessionId?: string;
  onCreateSession: () => void;
  isCreatingSession: boolean;
  onSelectSession: (sessionId: string) => void;
  onDeleteSession: (sessionId: string) => void;
  isLoading?: boolean;
  error?: unknown;
  onRetryLoad?: () => void;
}

interface SessionButtonProps {
  session: Session;
  isSelected: boolean;
  onSelect: () => void;
  onDelete: () => void;
  formattedTimestamp: string;
  fullTimestamp: string;
}

function SessionButton({
  session,
  isSelected,
  onSelect,
  onDelete,
  formattedTimestamp,
  fullTimestamp,
}: SessionButtonProps) {
  const [showTooltip, setShowTooltip] = useState(false);
  const [isTextTruncated, setIsTextTruncated] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  // Check if text is truncated by comparing scroll width with client width
  const handleTextRef = (element: HTMLDivElement | null) => {
    if (element) {
      setIsTextTruncated(element.scrollWidth > element.clientWidth);
    }
  };

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent session selection
    if (window.confirm('Are you sure you want to delete this session? This action cannot be undone.')) {
      onDelete();
    }
  };

  return (
    <div 
      className='relative group'
      onMouseEnter={() => {
        setShowTooltip(true);
        setIsHovered(true);
      }}
      onMouseLeave={() => {
        setShowTooltip(false);
        setIsHovered(false);
      }}
    >
      <button
        onClick={onSelect}
        onFocus={() => setShowTooltip(true)}
        onBlur={() => setShowTooltip(false)}
        className={`w-full text-left p-3 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 ${
          isSelected
            ? 'bg-blue-100 text-blue-900 border border-blue-200'
            : 'hover:bg-gray-100 focus:bg-gray-50'
        }`}
        aria-label={`Chat session from ${fullTimestamp}${
          isSelected ? ', currently selected' : ''
        }`}
        aria-describedby={`session-${session.id}-description`}
        role='option'
        aria-selected={isSelected}
      >
        <div className='flex items-center justify-between'>
          <div
            ref={handleTextRef}
            className='font-medium text-sm truncate min-w-0 flex-1'
            title={isTextTruncated ? formattedTimestamp : undefined}
          >
            {formattedTimestamp}
          </div>
          
          {/* Delete button - shows on hover */}
          {isHovered && (
            <button
              onClick={handleDeleteClick}
              className='ml-2 p-1 rounded hover:bg-red-100 text-gray-400 hover:text-red-600 transition-colors flex-shrink-0'
              aria-label={`Delete session from ${fullTimestamp}`}
              title='Delete session'
            >
              <svg className='w-4 h-4' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
                <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' />
              </svg>
            </button>
          )}
        </div>
        <div id={`session-${session.id}-description`} className='sr-only'>
          Session created on {fullTimestamp}
        </div>
      </button>

      {/* Tooltip for truncated text */}
      {showTooltip && isTextTruncated && (
        <div
          className='absolute left-0 top-full mt-1 z-50 bg-gray-900 text-white text-xs px-2 py-1 rounded shadow-lg whitespace-nowrap max-w-xs'
          role='tooltip'
          aria-hidden='true'
        >
          <div className='truncate'>{formattedTimestamp}</div>
          <div className='text-gray-300 text-xs mt-0.5'>{fullTimestamp}</div>
        </div>
      )}
    </div>
  );
}

export default function SessionSidebar({
  sessions,
  currentSessionId,
  onCreateSession,
  isCreatingSession,
  onSelectSession,
  onDeleteSession,
  isLoading = false,
  error = null,
  onRetryLoad,
}: SessionSidebarProps) {
  return (
    <aside
      className='w-full sm:w-1/3 lg:w-1/4 bg-white border-r border-gray-200 flex flex-col min-h-0'
      aria-label='Session navigation'
      role='complementary'
    >
      <div className='p-4 border-b border-gray-200 flex-shrink-0'>
        <button
          onClick={onCreateSession}
          disabled={isCreatingSession}
          className='w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 transition-colors flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
          aria-describedby='new-session-description'
        >
          {isCreatingSession ? (
            <>
              <LoadingSpinner size='sm' className='mr-2' />
              <span>Creating...</span>
            </>
          ) : (
            'New Session'
          )}
        </button>
        <div id='new-session-description' className='sr-only'>
          Create a new chat session
        </div>
      </div>

      <div className='flex-1 overflow-y-auto min-h-0'>
        <div className='p-2'>
          <h2
            className='text-sm font-semibold text-gray-600 mb-2 px-2'
            id='sessions-heading'
          >
            Recent Sessions
          </h2>

          {error && onRetryLoad ? (
            <SessionErrorHandler
              error={error}
              onRetry={onRetryLoad}
              loading={isLoading}
            />
          ) : isLoading ? (
            <div
              className='text-center py-8 px-4'
              role='status'
              aria-live='polite'
              aria-label='Loading sessions'
            >
              <LoadingSpinner className='mx-auto mb-2' />
              <p className='text-gray-600 text-sm'>Loading sessions...</p>
            </div>
          ) : !sessions || sessions.length === 0 ? (
            <div
              className='text-center text-gray-500 py-8 px-4'
              role='status'
              aria-live='polite'
            >
              <p className='font-medium'>No sessions yet</p>
              <p className='text-sm mt-1'>Start a new conversation!</p>
            </div>
          ) : (
            <div
              className='space-y-1'
              role='listbox'
              aria-labelledby='sessions-heading'
              aria-multiselectable='false'
            >
              {sessions.map(session => {
                const formattedTimestamp = formatSessionTimestamp(session.created_at);
                const fullTimestamp = new Date(session.created_at).toLocaleString();

                return (
                  <SessionButton
                    key={session.id}
                    session={session}
                    isSelected={currentSessionId === session.id}
                    onSelect={() => onSelectSession(session.id)}
                    onDelete={() => onDeleteSession(session.id)}
                    formattedTimestamp={formattedTimestamp}
                    fullTimestamp={fullTimestamp}
                  />
                );
              })}
            </div>
          )}
        </div>
      </div>
    </aside>
  );
}
