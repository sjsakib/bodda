import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import ChatInterface from '../ChatInterface'

// Mock react-router-dom
const mockNavigate = vi.fn()
const mockUseParams = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => mockUseParams()
  }
})

// Mock fetch
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock EventSource
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  
  constructor(public url: string, public options?: EventSourceInit) {}
  
  close() {}
  
  // Helper method to simulate events
  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }))
    }
  }
  
  simulateError() {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }
}

global.EventSource = MockEventSource as any

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn()

const renderChatInterface = (sessionId?: string) => {
  // Mock useParams to return sessionId
  mockUseParams.mockReturnValue({ sessionId: sessionId || 'test-session-id' })
  
  return render(
    <BrowserRouter>
      <ChatInterface />
    </BrowserRouter>
  )
}

describe('ChatInterface', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  it('renders the chat interface with sidebar and main area', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    expect(screen.getByText('New Session')).toBeInTheDocument()
    expect(screen.getByText('Recent Sessions')).toBeInTheDocument()
    expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/Ask your AI coach anything/)).toBeInTheDocument()
  })

  it('loads sessions on mount', async () => {
    const mockSessions = [
      {
        id: 'session-1',
        title: 'Training Discussion',
        createdAt: '2024-01-01T10:00:00Z'
      }
    ]

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockSessions)
    })

    renderChatInterface()

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions', {
        credentials: 'include'
      })
    })

    await waitFor(() => {
      expect(screen.getByText('Training Discussion')).toBeInTheDocument()
    })
  })

  it('redirects to landing page when not authenticated', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401
    })

    renderChatInterface()

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/')
    })
  })

  it('loads messages for current session', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'user',
        content: 'Hello coach',
        createdAt: '2024-01-01T10:00:00Z'
      },
      {
        id: 'msg-2',
        role: 'assistant',
        content: 'Hello! How can I help with your training today?',
        createdAt: '2024-01-01T10:01:00Z'
      }
    ]

    // Mock sessions call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    // Mock messages call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockMessages)
    })

    renderChatInterface()

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/test-session-id/messages', {
        credentials: 'include'
      })
    })

    await waitFor(() => {
      expect(screen.getByText('Hello coach')).toBeInTheDocument()
      expect(screen.getByText('Hello! How can I help with your training today?')).toBeInTheDocument()
    })
  })

  it('creates new session when button is clicked', async () => {
    const newSession = {
      id: 'new-session-id',
      title: 'New Session',
      createdAt: '2024-01-01T10:00:00Z'
    }

    // Mock initial sessions call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    // Mock create session call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(newSession)
    })

    renderChatInterface()

    const newSessionButton = screen.getByText('New Session')
    fireEvent.click(newSessionButton)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        }
      })
    })

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/chat/new-session-id')
    })
  })

  it('sends message and handles streaming response', async () => {
    // Mock initial calls
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    const sendButton = screen.getByText('Send')

    // Type message
    fireEvent.change(input, { target: { value: 'What should I do for training?' } })

    // Mock send message call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    // Click send
    fireEvent.click(sendButton)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/test-session-id/messages', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ content: 'What should I do for training?' })
      })
    })

    // Check that user message appears immediately
    expect(screen.getByText('What should I do for training?')).toBeInTheDocument()
    expect(screen.getByText('AI is thinking...')).toBeInTheDocument()
  })

  it('handles streaming response correctly', async () => {
    // Mock initial calls
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    fireEvent.change(input, { target: { value: 'Test message' } })

    // Mock send message call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    fireEvent.click(screen.getByText('Send'))

    await waitFor(() => {
      expect(screen.getByText('AI is thinking...')).toBeInTheDocument()
    })

    // Simulate EventSource messages
    const eventSource = new MockEventSource('/api/sessions/test-session-id/stream')
    
    // Simulate content chunks
    eventSource.simulateMessage({ type: 'content', content: 'Hello ' })
    eventSource.simulateMessage({ type: 'content', content: 'there!' })
    eventSource.simulateMessage({ type: 'done' })

    await waitFor(() => {
      expect(screen.getByText('Hello there!')).toBeInTheDocument()
    })
  })

  it('handles Enter key to send message', async () => {
    // Mock initial calls
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    fireEvent.change(input, { target: { value: 'Test message' } })

    // Mock send message call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    // Press Enter
    fireEvent.keyPress(input, { key: 'Enter', code: 'Enter' })

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/test-session-id/messages', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ content: 'Test message' })
      })
    })
  })

  it('prevents sending empty messages', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    const sendButton = screen.getByText('Send')
    expect(sendButton).toBeDisabled()

    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    fireEvent.change(input, { target: { value: '   ' } }) // Only whitespace

    expect(sendButton).toBeDisabled()
  })

  it('displays error messages', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    renderChatInterface()

    await waitFor(() => {
      expect(screen.getByText('Failed to load sessions')).toBeInTheDocument()
    })

    // Test error dismissal
    const closeButton = screen.getByText('×')
    fireEvent.click(closeButton)

    await waitFor(() => {
      expect(screen.queryByText('Failed to load sessions')).not.toBeInTheDocument()
    })
  })

  it('handles streaming errors', async () => {
    // Mock initial calls
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    fireEvent.change(input, { target: { value: 'Test message' } })

    // Mock send message call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    fireEvent.click(screen.getByText('Send'))

    // Simulate EventSource error
    const eventSource = new MockEventSource('/api/sessions/test-session-id/stream')
    eventSource.simulateError()

    await waitFor(() => {
      expect(screen.getByText('Connection lost while streaming response')).toBeInTheDocument()
    })
  })

  it('renders markdown in assistant messages', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'assistant',
        content: '# Training Plan\n\n**Week 1**: Easy runs\n\n- Run 1: 30 minutes\n- Run 2: 45 minutes',
        createdAt: '2024-01-01T10:00:00Z'
      }
    ]

    // Mock sessions call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    // Mock messages call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockMessages)
    })

    renderChatInterface()

    await waitFor(() => {
      expect(screen.getByText('Training Plan')).toBeInTheDocument()
      expect(screen.getByText('Week 1')).toBeInTheDocument()
      expect(screen.getByText('Run 1: 30 minutes')).toBeInTheDocument()
    })
  })

  it('shows welcome message when no messages exist', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    expect(screen.getByText('Welcome to Bodda!')).toBeInTheDocument()
    expect(screen.getByText(/Start a conversation with your AI coach/)).toBeInTheDocument()
  })

  it('formats timestamps correctly', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'user',
        content: 'Test message',
        createdAt: '2024-01-01T14:30:00Z'
      }
    ]

    // Mock sessions call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    // Mock messages call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockMessages)
    })

    renderChatInterface()

    await waitFor(() => {
      // The exact format depends on locale, but should contain time
      const timeElements = screen.getAllByText(/\d{1,2}:\d{2}/)
      expect(timeElements.length).toBeGreaterThan(0)
    })
  })
})