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

// Mock EventSource
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  
  constructor(public url: string, public options?: EventSourceInit) {}
  
  close() {}
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

describe('ChatInterface - Suggestion Pills Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
    
    // Mock successful auth check by default
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/auth/check')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            authenticated: true,
            user: { id: 'user-1' }
          })
        })
      }
      if (url.includes('/api/sessions') && !url.includes('messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ sessions: [] })
        })
      }
      if (url.includes('/messages')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ messages: [] })
        })
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ sessions: [] })
      })
    })
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  describe('Pills Visibility Based on Session State and Input State', () => {
    it('shows suggestion pills when session is empty and input is empty - Requirements 1.1', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.queryByText('Loading messages...')).not.toBeInTheDocument()
      })

      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
        expect(screen.getByText('Analyze my recent running performance')).toBeInTheDocument()
      })

      expect(screen.getByText('Welcome to Bodda!')).toBeInTheDocument()
    })

    it('hides suggestion pills when input has text - Requirements 1.2', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      fireEvent.change(input, { target: { value: 'Hello' } })

      await waitFor(() => {
        expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
      })
    })

    it('shows suggestion pills again when input is cleared - Requirements 1.3', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      
      fireEvent.change(input, { target: { value: 'Hello' } })
      await waitFor(() => {
        expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
      })

      fireEvent.change(input, { target: { value: '' } })
      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
      })
    })

    it('does not show suggestion pills when messages exist - Requirements 1.4', async () => {
      const mockMessages = [
        {
          id: 'msg-1',
          role: 'user',
          content: 'Hello coach',
          created_at: '2024-01-01T10:00:00Z',
          session_id: 'test-session-id'
        }
      ]

      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1' }
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ messages: mockMessages })
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ sessions: [] })
        })
      })

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Hello coach')).toBeInTheDocument()
      })

      expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
    })
  })

  describe('Input Population When Pills Are Clicked', () => {
    it('populates input field with pill text when clicked - Requirements 2.1', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/) as HTMLTextAreaElement
      const pill = screen.getByText('Help me plan my next training week')

      fireEvent.click(pill)

      expect(input.value).toBe('Help me plan my next training week')
    })

    it('hides suggestion pills after clicking a pill - Requirements 2.2', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Analyze my recent running performance')).toBeInTheDocument()
      })

      const pill = screen.getByText('Analyze my recent running performance')
      fireEvent.click(pill)

      await waitFor(() => {
        expect(screen.queryByRole('region', { name: 'Quick start suggestions' })).not.toBeInTheDocument()
      })
    })

    it('focuses input field after clicking a pill - Requirements 2.3', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('What strength training should I focus on?')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      const pill = screen.getByText('What strength training should I focus on?')

      fireEvent.click(pill)

      await waitFor(() => {
        expect(document.activeElement).toBe(input)
      })
    })

    it('allows editing text after pill click - Requirements 2.4', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Help me set realistic training goals')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/) as HTMLTextAreaElement
      const pill = screen.getByText('Help me set realistic training goals')

      fireEvent.click(pill)
      expect(input.value).toBe('Help me set realistic training goals')

      fireEvent.change(input, { target: { value: 'Help me set realistic training goals for marathon' } })
      expect(input.value).toBe('Help me set realistic training goals for marathon')

      expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
    })
  })

  describe('Pills Hiding When Input Has Text or Messages Exist', () => {
    it('hides pills immediately when user starts typing', async () => {
      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      fireEvent.change(input, { target: { value: 'H' } })

      expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
    })

    it('keeps pills hidden when messages exist even after clearing input', async () => {
      const mockMessages = [
        {
          id: 'msg-1',
          role: 'user',
          content: 'Hello',
          created_at: '2024-01-01T10:00:00Z',
          session_id: 'test-session-id'
        }
      ]

      mockFetch.mockImplementation((url: string) => {
        if (url.includes('/api/auth/check')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
              authenticated: true,
              user: { id: 'user-1' }
            })
          })
        }
        if (url.includes('/api/sessions') && !url.includes('messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ sessions: [] })
          })
        }
        if (url.includes('/messages')) {
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({ messages: mockMessages })
          })
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ sessions: [] })
        })
      })

      renderChatInterface()

      await waitFor(() => {
        expect(screen.getByText('Hello')).toBeInTheDocument()
      })

      const input = screen.getByPlaceholderText(/Ask your AI coach/)
      fireEvent.change(input, { target: { value: 'Test' } })
      fireEvent.change(input, { target: { value: '' } })

      expect(screen.queryByText('Help me plan my next training week')).not.toBeInTheDocument()
    })
  })
})