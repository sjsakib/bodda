import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';
import { Message, Session } from '../services/api';
import SessionSidebar from './SessionSidebar';
import { ErrorDisplay, LoadingSpinner } from './ErrorBoundary';
import { MessageErrorHandler } from './ApiErrorHandler';
import { SafeMarkdownRenderer } from './MarkdownRenderer';

export default function ChatInterface() {
  const params = useParams<{ sessionId: string }>();
  const sessionId = params?.sessionId;
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [isLoadingSessions, setIsLoadingSessions] = useState(true);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);
  const [isCreatingSession, setIsCreatingSession] = useState(false);
  const [isStreaming, setIsStreaming] = useState(false);
  const [sessionError, setSessionError] = useState<unknown>(null);
  const [messageError, setMessageError] = useState<unknown>(null);
  const [streamingError, setStreamingError] = useState<string | null>(null);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  // Redirect to landing page if not authenticated or no sessionId
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const authResponse = await apiClient.checkAuth();
        if (!authResponse || !authResponse.authenticated) {
          navigate('/');
          return;
        }

        // If no sessionId in URL, redirect to first session or create new one
        if (!sessionId) {
          try {
            const sessions = await apiClient.getSessions();
            if (sessions && sessions.length > 0) {
              navigate(`/chat/${sessions[0].id}`);
            } else {
              // Create a new session if none exist
              const newSession = await apiClient.createSession();
              if (newSession && newSession.id) {
                navigate(`/chat/${newSession.id}`);
              }
            }
          } catch (error) {
            console.error('Failed to get or create session:', error);
            // Still allow the component to render, user can create session manually
          }
        }
      } catch (error) {
        console.error('Auth check failed:', error);
        navigate('/');
      }
    };

    checkAuth();
  }, [navigate, sessionId]);

  // Load sessions
  const loadSessions = async () => {
    setIsLoadingSessions(true);
    setSessionError(null);

    try {
      const sessionList = await apiClient.getSessions();
      setSessions(sessionList || []);
    } catch (error) {
      console.error('Failed to load sessions:', error);
      setSessionError(error);
      setSessions([]); // Ensure sessions is always an array
    } finally {
      setIsLoadingSessions(false);
    }
  };

  useEffect(() => {
    loadSessions();
  }, []);

  // Load messages for current session
  const loadMessages = async () => {
    if (!sessionId) return;

    setIsLoadingMessages(true);
    setMessageError(null);

    try {
      const sessionMessages = await apiClient.getMessages(sessionId);
      setMessages(sessionMessages || []);
    } catch (error) {
      console.error('Failed to load messages:', error);
      setMessageError(error);
      setMessages([]); // Ensure messages is always an array
    } finally {
      setIsLoadingMessages(false);
    }
  };

  useEffect(() => {
    loadMessages();
  }, [sessionId]);

  const createNewSession = async () => {
    setIsCreatingSession(true);
    setSessionError(null);

    try {
      const newSession = await apiClient.createSession();
      if (newSession && newSession.id) {
        setSessions(prev => [newSession, ...prev]);
        navigate(`/chat/${newSession.id}`);
      } else {
        throw new Error('Invalid session response from server');
      }
    } catch (error) {
      console.error('Failed to create session:', error);
      setSessionError(error);
    } finally {
      setIsCreatingSession(false);
    }
  };

  const sendMessage = async (content: string) => {
    if (!content.trim() || isStreaming || !sessionId) return;

    setMessageError(null);
    setStreamingError(null);

    // Create temporary user message
    const userMessage: Message = {
      id: `temp-${Date.now()}`,
      content: content.trim(),
      role: 'user',
      created_at: new Date().toISOString(),
      session_id: sessionId,
    };

    // Add user message to UI immediately
    setMessages(prev => [...prev, userMessage]);

    // Create temporary assistant message for streaming
    const assistantMessage: Message = {
      id: `temp-assistant-${Date.now()}`,
      content: '',
      role: 'assistant',
      created_at: new Date().toISOString(),
      session_id: sessionId,
    };

    setMessages(prev => [...prev, assistantMessage]);
    setIsStreaming(true);

    try {
      // Set up SSE for streaming response (this will handle saving the message)
      const eventSource = apiClient.createEventSource(sessionId, content.trim());

      eventSource.onmessage = event => {
        try {
          if (!event || !event.data) {
            console.warn('Received empty SSE event');
            return;
          }

          // Parse the SSE event data
          const eventData = event.data.trim();
          if (!eventData) {
            return;
          }

          // Try to parse as JSON for structured events
          try {
            const parsedData = JSON.parse(eventData);

            if (parsedData.type === 'chunk') {
              assistantMessage.content += parsedData.content || '';
              setMessages(prev =>
                prev.map(msg =>
                  msg.id === assistantMessage.id ? { ...assistantMessage } : msg
                )
              );
            } else if (parsedData.type === 'complete') {
              eventSource.close();
              setIsStreaming(false);
              if (parsedData.message) {
                setMessages(prev =>
                  prev.map(msg =>
                    msg.id === assistantMessage.id ? { ...parsedData.message } : msg
                  )
                );
              }
            } else if (parsedData.type === 'error') {
              eventSource.close();
              setIsStreaming(false);
              setStreamingError(
                parsedData.message || 'An error occurred while streaming response'
              );
            } else if (parsedData.type === 'user_message') {
              if (parsedData.message) {
                setMessages(prev =>
                  prev.map(msg =>
                    msg.id === userMessage.id ? { ...parsedData.message } : msg
                  )
                );
              }
            }
          } catch {
            // Treat as plain text chunk if not JSON
            assistantMessage.content += eventData;
            setMessages(prev =>
              prev.map(msg =>
                msg.id === assistantMessage.id ? { ...assistantMessage } : msg
              )
            );
          }
        } catch (parseError) {
          console.error('Failed to handle SSE event:', parseError);
        }
      };

      eventSource.onerror = error => {
        console.error('SSE error:', error);
        eventSource.close();
        setIsStreaming(false);
        setStreamingError('Connection lost. Please try again.');
      };
    } catch (error) {
      console.error('Failed to send message:', error);
      setMessageError(error);
      setIsStreaming(false);
      // Remove the temporary messages on error
      setMessages(prev =>
        prev.filter(msg => msg.id !== userMessage.id && msg.id !== assistantMessage.id)
      );
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputText.trim()) {
      sendMessage(inputText);
      setInputText('');
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await apiClient.logout();
      // Redirect to landing page after successful logout
      navigate('/');
    } catch (error) {
      console.error('Logout failed:', error);
      // Even if logout fails, redirect to landing page to clear local state
      navigate('/');
    } finally {
      setIsLoggingOut(false);
    }
  };

  if (!sessionId) {
    return (
      <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
        <div className='text-center'>
          <LoadingSpinner size='lg' className='mx-auto mb-4' />
          <p className='text-gray-600'>Setting up chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className='flex h-screen bg-gray-50'>
      <SessionSidebar
        sessions={sessions}
        currentSessionId={sessionId}
        onCreateSession={createNewSession}
        isCreatingSession={isCreatingSession}
        onSelectSession={id => navigate(`/chat/${id}`)}
        isLoading={isLoadingSessions}
        error={sessionError}
        onRetryLoad={loadSessions}
      />

      <div className='flex-1 flex flex-col'>
        {/* Header */}
        <div className='bg-white border-b border-gray-200 p-4 flex justify-between items-center'>
          <div>
            <h1 className='text-xl font-semibold text-gray-800'>Bodda AI Coach</h1>
            <p className='text-sm text-gray-600'>Your personal running and cycling coach</p>
          </div>
          <button
            onClick={handleLogout}
            disabled={isLoggingOut}
            className='bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2'
          >
            {isLoggingOut ? (
              <>
                <LoadingSpinner size='sm' />
                <span>Logging out...</span>
              </>
            ) : (
              <span>Logout</span>
            )}
          </button>
        </div>

        {/* Messages Area */}
        <div className='flex-1 overflow-y-auto p-4 space-y-4'>
          {!!messageError && (
            <MessageErrorHandler
              error={messageError as Error}
              onRetry={loadMessages}
              loading={isLoadingMessages}
            />
          )}

          {streamingError && (
            <ErrorDisplay
              error={streamingError}
              onDismiss={() => setStreamingError(null)}
              className='mb-4'
            />
          )}

          {isLoadingMessages ? (
            <div className='flex justify-center items-center h-32'>
              <div className='text-center'>
                <LoadingSpinner className='mx-auto mb-2' />
                <p className='text-gray-600 text-sm'>Loading messages...</p>
              </div>
            </div>
          ) : messageError ? null : messages.length === 0 ? (
            <div className='text-center text-gray-500 py-12'>
              <h3 className='text-lg font-medium mb-2'>Welcome to Bodda!</h3>
              <p>
                Start a conversation with your AI coach. Ask about training, analyze your
                activities, or get personalized advice.
              </p>
            </div>
          ) : (
            messages.map(message => (
              <div
                key={message.id}
                className={`flex ${
                  message.role === 'user' ? 'justify-end' : 'justify-start'
                }`}
              >
                <div
                  className={`max-w-full sm:max-w-2xl lg:max-w-3xl rounded-lg px-3 sm:px-4 py-2 sm:py-3 ${
                    message.role === 'user'
                      ? 'bg-blue-600 text-white'
                      : 'bg-white border border-gray-200'
                  }`}
                >
                  {message.role === 'assistant' ? (
                    <SafeMarkdownRenderer 
                      content={message.content}
                      className="max-w-none"
                    />
                  ) : (
                    <div className='whitespace-pre-wrap text-sm sm:text-base'>{message.content}</div>
                  )}
                  <div
                    className={`text-xs mt-2 ${
                      message.role === 'user' ? 'text-blue-100' : 'text-gray-500'
                    }`}
                  >
                    {formatTimestamp(message.created_at)}
                  </div>
                </div>
              </div>
            ))
          )}

          {isStreaming && (
            <div className='flex justify-start'>
              <div className='bg-white border border-gray-200 rounded-lg px-4 py-3'>
                <div className='flex items-center space-x-2'>
                  <LoadingSpinner size='sm' />
                  <span className='text-sm text-gray-500'>AI is thinking...</span>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <div className='bg-white border-t border-gray-200 p-4'>
          <form onSubmit={handleSubmit} className='flex space-x-3'>
            <textarea
              value={inputText}
              onChange={e => setInputText(e.target.value)}
              onKeyDown={e => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSubmit(e);
                }
              }}
              placeholder='Ask your AI coach anything about training, analyze your activities, or get personalized advice...'
              className='flex-1 resize-none border border-gray-300 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent'
              rows={1}
              disabled={isStreaming}
            />
            <button
              type='submit'
              disabled={!inputText.trim() || isStreaming}
              className='bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors'
            >
              {isStreaming ? 'Sending...' : 'Send'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
