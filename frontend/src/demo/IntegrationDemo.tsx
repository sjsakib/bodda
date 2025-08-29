import React, { useState, useEffect } from 'react';
import { apiClient } from '../services/api';
import { ErrorDisplay, LoadingSpinner } from '../components/ErrorBoundary';

// Simple demo component to show API integration working
export default function IntegrationDemo() {
  const [authStatus, setAuthStatus] = useState<
    'loading' | 'authenticated' | 'unauthenticated'
  >('loading');
  const [sessions, setSessions] = useState<any[]>([]);
  const [messages, setMessages] = useState<any[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [selectedSession, setSelectedSession] = useState<string | null>(null);

  // Check authentication on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await apiClient.checkAuth();
        if (response.authenticated) {
          setAuthStatus('authenticated');
          loadSessions();
        } else {
          setAuthStatus('unauthenticated');
        }
      } catch (err) {
        console.error('Auth check failed:', err);
        setAuthStatus('unauthenticated');
        setError(apiClient.getErrorMessage(err));
      }
    };

    checkAuth();
  }, []);

  const loadSessions = async () => {
    try {
      const sessionList = await apiClient.getSessions();
      setSessions(sessionList);
      if (sessionList.length > 0) {
        setSelectedSession(sessionList[0].id);
        loadMessages(sessionList[0].id);
      }
    } catch (err) {
      console.error('Failed to load sessions:', err);
      setError(apiClient.getErrorMessage(err));
    }
  };

  const loadMessages = async (sessionId: string) => {
    try {
      const messageList = await apiClient.getMessages(sessionId);
      setMessages(messageList);
    } catch (err) {
      console.error('Failed to load messages:', err);
      setError(apiClient.getErrorMessage(err));
    }
  };

  const createSession = async () => {
    try {
      const newSession = await apiClient.createSession('Demo Session');
      setSessions(prev => [newSession, ...prev]);
      setSelectedSession(newSession.id);
      setMessages([]);
    } catch (err) {
      console.error('Failed to create session:', err);
      setError(apiClient.getErrorMessage(err));
    }
  };

  const sendMessage = async (content: string) => {
    if (!selectedSession) return;

    try {
      const response = await apiClient.sendMessage(selectedSession, content);
      // Add both user and assistant messages
      setMessages(prev => [...prev, response.user_message, response.assistant_message]);
    } catch (err) {
      console.error('Failed to send message:', err);
      setError(apiClient.getErrorMessage(err));
    }
  };

  if (authStatus === 'loading') {
    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
        <LoadingSpinner size='lg' />
      </div>
    );
  }

  if (authStatus === 'unauthenticated') {
    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
        <div className='text-center'>
          <h1 className='text-2xl font-bold mb-4'>Authentication Required</h1>
          <p className='text-gray-600 mb-4'>
            Please authenticate with Strava to continue.
          </p>
          <button
            onClick={() => apiClient.redirectToStravaAuth()}
            className='bg-orange-500 text-white px-6 py-3 rounded-lg hover:bg-orange-600'
          >
            Connect with Strava
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className='min-h-screen bg-gray-50 p-4'>
      <div className='max-w-6xl mx-auto'>
        <h1 className='text-3xl font-bold mb-6'>API Integration Demo</h1>

        {error && (
          <ErrorDisplay error={error} onDismiss={() => setError(null)} className='mb-6' />
        )}

        <div className='grid grid-cols-1 md:grid-cols-3 gap-6'>
          {/* Sessions Panel */}
          <div className='bg-white rounded-lg shadow p-4'>
            <div className='flex justify-between items-center mb-4'>
              <h2 className='text-lg font-semibold'>Sessions</h2>
              <button
                onClick={createSession}
                className='bg-blue-500 text-white px-3 py-1 rounded text-sm hover:bg-blue-600'
              >
                New Session
              </button>
            </div>

            <div className='space-y-2'>
              {sessions.map(session => (
                <button
                  key={session.id}
                  onClick={() => {
                    setSelectedSession(session.id);
                    loadMessages(session.id);
                  }}
                  className={`w-full text-left p-2 rounded ${
                    selectedSession === session.id
                      ? 'bg-blue-100 text-blue-900'
                      : 'hover:bg-gray-100'
                  }`}
                >
                  <div className='font-medium text-sm'>{session.title}</div>
                  <div className='text-xs text-gray-500'>
                    {new Date(session.created_at).toLocaleDateString()}
                  </div>
                </button>
              ))}

              {sessions.length === 0 && (
                <p className='text-gray-500 text-sm'>No sessions yet</p>
              )}
            </div>
          </div>

          {/* Messages Panel */}
          <div className='bg-white rounded-lg shadow p-4 md:col-span-2'>
            <h2 className='text-lg font-semibold mb-4'>Messages</h2>

            <div className='space-y-3 mb-4 max-h-96 overflow-y-auto'>
              {messages &&
                messages.map(message => (
                  <div
                    key={message.id}
                    className={`p-3 rounded ${
                      message.role === 'user' ? 'bg-blue-100 ml-8' : 'bg-gray-100 mr-8'
                    }`}
                  >
                    <div className='font-medium text-sm mb-1'>
                      {message.role === 'user' ? 'You' : 'AI Coach'}
                    </div>
                    <div className='text-sm'>{message.content}</div>
                    <div className='text-xs text-gray-500 mt-1'>
                      {new Date(message.created_at).toLocaleTimeString()}
                    </div>
                  </div>
                ))}

              {messages.length === 0 && selectedSession && (
                <p className='text-gray-500 text-sm'>
                  No messages yet. Start a conversation!
                </p>
              )}

              {!selectedSession && (
                <p className='text-gray-500 text-sm'>Select a session to view messages</p>
              )}
            </div>

            {/* Message Input */}
            {selectedSession && (
              <form
                onSubmit={e => {
                  e.preventDefault();
                  const form = e.target as HTMLFormElement;
                  const input = form.elements.namedItem('message') as HTMLInputElement;
                  if (input.value.trim()) {
                    sendMessage(input.value.trim());
                    input.value = '';
                  }
                }}
                className='flex gap-2'
              >
                <input
                  name='message'
                  type='text'
                  placeholder='Type your message...'
                  className='flex-1 border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500'
                />
                <button
                  type='submit'
                  className='bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600'
                >
                  Send
                </button>
              </form>
            )}
          </div>
        </div>

        {/* API Status */}
        <div className='mt-6 bg-white rounded-lg shadow p-4'>
          <h2 className='text-lg font-semibold mb-2'>API Status</h2>
          <div className='grid grid-cols-2 md:grid-cols-4 gap-4 text-sm'>
            <div>
              <div className='font-medium'>Authentication</div>
              <div className='text-green-600'>✓ Connected</div>
            </div>
            <div>
              <div className='font-medium'>Sessions</div>
              <div className='text-blue-600'>{sessions.length} loaded</div>
            </div>
            <div>
              <div className='font-medium'>Messages</div>
              <div className='text-blue-600'>{messages.length} in current session</div>
            </div>
            <div>
              <div className='font-medium'>Error Handling</div>
              <div className={error ? 'text-red-600' : 'text-green-600'}>
                {error ? '⚠ Has errors' : '✓ No errors'}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
