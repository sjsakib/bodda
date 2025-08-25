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

// Mock EventSource for streaming tests
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  
  constructor(public url: string, public options?: EventSourceInit) {}
  
  close() {}
  
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
  mockUseParams.mockReturnValue({ sessionId: sessionId || 'test-session-id' })
  
  return render(
    <BrowserRouter>
      <ChatInterface />
    </BrowserRouter>
  )
}

describe('ChatInterface Markdown Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  describe('AI Message Rendering with Markdown', () => {
    it('renders markdown headings correctly in assistant messages', async () => {
      const mockMessages = [{
        id: 'msg-1',
        role: 'assistant' as const,
        content: '# Training Plan\n\n## Week 1\n\n### Daily Schedule',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }]

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
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Plan')
        expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Week 1')
        expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Daily Schedule')
      })

      // Verify heading styling
      const h1 = screen.getByRole('heading', { level: 1 })
      expect(h1).toHaveClass('text-xl', 'font-bold', 'text-gray-900')
    })

    it('renders markdown lists correctly in assistant messages', async () => {
      const mockMessages = [{
        id: 'msg-1',
        role: 'assistant' as const,
        content: '- **Monday**: Easy run\n- **Tuesday**: Rest day\n\n1. Build base\n2. Stay consistent',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMessages)
      })

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Monday')).toHaveClass('font-semibold')
        expect(screen.getByText('Tuesday')).toHaveClass('font-semibold')
        expect(screen.getByText('Build base')).toBeInTheDocument()
        expect(screen.getByText('Stay consistent')).toBeInTheDocument()
      })

      // Verify list elements are present
      const lists = screen.getAllByRole('list')
      expect(lists).toHaveLength(2) // One unordered, one ordered
    })

    it('renders markdown tables correctly in assistant messages', async () => {
      const mockMessages = [{
        id: 'msg-1',
        role: 'assistant' as const,
        content: '| Day | Activity |\n|-----|----------|\n| Mon | Run |\n| Tue | Rest |',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMessages)
      })

      renderChatInterface()

      await waitFor(() => {
        const table = screen.getByRole('table')
        expect(table).toBeInTheDocument()
        expect(screen.getByText('Day')).toBeInTheDocument()
        expect(screen.getByText('Activity')).toBeInTheDocument()
        expect(screen.getByText('Mon')).toBeInTheDocument()
        expect(screen.getByText('Run')).toBeInTheDocument()
      })

      // Verify responsive wrapper
      const table = screen.getByRole('table')
      const tableWrapper = table.closest('div')
      expect(tableWrapper).toHaveClass('overflow-x-auto')
    })

    it('preserves user messages as plain text (no markdown rendering)', async () => {
      const mockMessages = [
        {
          id: 'msg-1',
          role: 'user' as const,
          content: '# This should not be rendered as markdown\n**Bold text** should stay plain',
          created_at: '2024-01-01T10:00:00Z',
          session_id: 'test-session-id'
        },
        {
          id: 'msg-2',
          role: 'assistant' as const,
          content: '# This should be rendered as markdown\n**Bold text** should be bold',
          created_at: '2024-01-01T10:01:00Z',
          session_id: 'test-session-id'
        }
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMessages)
      })

      renderChatInterface()

      await waitFor(() => {
        // User message should show raw markdown
        expect(screen.getByText('# This should not be rendered as markdown')).toBeInTheDocument()
        expect(screen.getByText('**Bold text** should stay plain')).toBeInTheDocument()
        
        // Assistant message should render markdown
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('This should be rendered as markdown')
        expect(screen.getByText('Bold text')).toHaveClass('font-semibold')
      })
    })
  })

  describe('Streaming Content with Markdown', () => {
    it('renders markdown correctly during streaming updates', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      renderChatInterface()

      const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
      fireEvent.change(input, { target: { value: 'Give me a training plan' } })

      // Mock send message call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 'msg-id' })
      })

      fireEvent.click(screen.getByText('Send'))

      await waitFor(() => {
        expect(screen.getByText('AI is thinking...')).toBeInTheDocument()
      })

      // Simulate streaming markdown content
      const eventSource = new MockEventSource('/api/sessions/test-session-id/stream')
      
      // Stream markdown content in chunks
      act(() => {
        eventSource.simulateMessage({ type: 'chunk', content: '# Training Plan\n\n' })
      })

      act(() => {
        eventSource.simulateMessage({ type: 'chunk', content: '- **Monday**: Easy run\n' })
      })

      act(() => {
        eventSource.simulateMessage({ type: 'complete' })
      })

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Plan')
        expect(screen.getByText('Monday')).toHaveClass('font-semibold')
      })
    })

    it('handles streaming errors gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      renderChatInterface()

      const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
      fireEvent.change(input, { target: { value: 'Test streaming error' } })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 'msg-id' })
      })

      fireEvent.click(screen.getByText('Send'))

      const eventSource = new MockEventSource('/api/sessions/test-session-id/stream')
      
      // Stream some content first
      act(() => {
        eventSource.simulateMessage({ type: 'chunk', content: '# Training Plan\n\n**Week 1**: Base building' })
      })

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Plan')
        expect(screen.getByText('Week 1')).toHaveClass('font-semibold')
      })

      // Simulate streaming error
      act(() => {
        eventSource.simulateError()
      })

      await waitFor(() => {
        expect(screen.getByText('Connection lost. Please try again.')).toBeInTheDocument()
        // Previously rendered markdown should still be visible
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Plan')
        expect(screen.getByText('Week 1')).toHaveClass('font-semibold')
      })
    })
  })

  describe('Error Handling and Graceful Degradation', () => {
    it('handles malformed markdown gracefully', async () => {
      const malformedMarkdown = `# Heading
**Unclosed bold
| Malformed | table
Missing closing`

      const mockMessages = [{
        id: 'msg-1',
        role: 'assistant' as const,
        content: malformedMarkdown,
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMessages)
      })

      renderChatInterface()

      // Should render without crashing, even with malformed markdown
      await waitFor(() => {
        expect(screen.getByText(/Heading/)).toBeInTheDocument()
        expect(screen.getByText(/Unclosed bold/)).toBeInTheDocument()
      })
    })

    it('handles empty markdown content gracefully', async () => {
      const mockMessages = [{
        id: 'msg-1',
        role: 'assistant' as const,
        content: '',
        created_at: '2024-01-01T10:00:00Z',
        session_id: 'test-session-id'
      }]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMessages)
      })

      renderChatInterface()

      // Should render without crashing
      await waitFor(() => {
        // The message container should exist even if empty
        const timeElements = screen.getAllByText(/\d{1,2}:\d{2}/)
        expect(timeElements.length).toBeGreaterThan(0)
      })
    })
  })

  describe('Existing Chat Functionality Integrity', () => {
    it('preserves message sending functionality', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      renderChatInterface()

      const input = screen.getByPlaceholderText(/Ask your AI coach anything/)
      const sendButton = screen.getByText('Send')

      // Test input functionality
      fireEvent.change(input, { target: { value: 'Test message with **markdown**' } })
      expect(input).toHaveValue('Test message with **markdown**')
      expect(sendButton).not.toBeDisabled()

      // Mock send message call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 'msg-id' })
      })

      fireEvent.click(sendButton)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/api/sessions/test-session-id/messages', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ content: 'Test message with **markdown**' })
        })
      })

      // Input should be cleared after sending
      expect(input).toHaveValue('')
    })

    it('preserves logout functionality', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      renderChatInterface()

      const logoutButton = screen.getByText('Logout')
      expect(logoutButton).toBeInTheDocument()

      // Mock logout API call
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ message: 'logged out successfully' })
      })

      fireEvent.click(logoutButton)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/auth/logout', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json'
          }
        })
      })

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/')
      })
    })
  })
})