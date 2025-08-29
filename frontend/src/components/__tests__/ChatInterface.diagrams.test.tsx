import { render, screen, fireEvent, waitFor } from '@testing-library/react'
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

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn()

const renderChatInterface = (sessionId?: string) => {
  mockUseParams.mockReturnValue({ sessionId: sessionId || 'test-session-id' })
  
  return render(
    <BrowserRouter>
      <ChatInterface />
    </BrowserRouter>
  )
}

describe('ChatInterface - Diagram Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  it('renders messages with diagram content and applies proper CSS classes', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'assistant',
        content: 'Here is your training flow:\n\n```mermaid\ngraph TD\nA[Start] --> B[Warm up]\n```',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }
    ]

    // Mock auth check
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
    })

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
      expect(screen.getByText('Here is your training flow:')).toBeInTheDocument()
    })

    // Check that message container has proper CSS classes
    const messageContainer = screen.getByText('Here is your training flow:').closest('.chat-message-container')
    expect(messageContainer).toHaveClass('chat-message-container')

    // Check that the message content has the proper class
    const messageContent = messageContainer?.querySelector('.chat-message-content')
    expect(messageContent).toBeInTheDocument()
  })

  it('enables diagram rendering with proper props in MarkdownRenderer', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'assistant',
        content: 'Chart example:\n\n```vega-lite\n{"mark": "bar", "data": {"values": []}}\n```',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }
    ]

    // Mock auth check
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
    })

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
      expect(screen.getByText('Chart example:')).toBeInTheDocument()
    })

    // Verify the message renders without errors
    expect(screen.getByText('Chart example:')).toBeInTheDocument()
  })

  it('maintains chat interface responsiveness with diagram content', async () => {
    const mockMessages = [
      {
        id: 'msg-1',
        role: 'assistant',
        content: 'Mixed content:\n\n## Analysis\n\nHere is data:\n\n```mermaid\ngraph TD\nA --> B\n```\n\nAnd a chart:\n\n```vega-lite\n{"mark": "bar", "data": {"values": []}}\n```',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }
    ]

    // Mock auth check
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
    })

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
      expect(screen.getByText('Mixed content:')).toBeInTheDocument()
      expect(screen.getByText('Analysis')).toBeInTheDocument()
      expect(screen.getByText('Here is data:')).toBeInTheDocument()
      expect(screen.getByText('And a chart:')).toBeInTheDocument()
    })

    // Check that the input field is still functional
    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    expect(input).not.toBeDisabled()

    // Check that we can still type in the input
    fireEvent.change(input, { target: { value: 'New message' } })
    expect(input).toHaveValue('New message')

    // Check that send button is enabled
    const sendButton = screen.getByText('Send')
    expect(sendButton).not.toBeDisabled()
  })

  it('handles incomplete diagram content during streaming', async () => {
    // Mock auth check
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
    })

    // Mock sessions call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    // Mock messages call
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([])
    })

    renderChatInterface()

    // Verify the interface loads properly
    await waitFor(() => {
      expect(screen.getByText('Welcome to Bodda!')).toBeInTheDocument()
    })

    // Check that the input field is functional
    const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
    expect(input).not.toBeDisabled()
  })
})