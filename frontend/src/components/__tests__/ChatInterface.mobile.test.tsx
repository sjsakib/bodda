import { render, screen, fireEvent, act } from '@testing-library/react'
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

// Mock window.matchMedia for responsive testing
const mockMatchMedia = vi.fn()
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: mockMatchMedia,
})

const renderChatInterface = (sessionId?: string) => {
  mockUseParams.mockReturnValue({ sessionId: sessionId || 'test-session-id' })
  
  return render(
    <BrowserRouter>
      <ChatInterface />
    </BrowserRouter>
  )
}

describe('ChatInterface Mobile Input Field', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockClear()
    
    // Mock successful auth and sessions calls
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ authenticated: true, user: { id: 'user-1' } })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })
  })

  afterEach(() => {
    vi.clearAllTimers()
  })

  it('displays shorter placeholder text on mobile devices', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText('Ask your AI coach...')
    expect(textarea).toBeInTheDocument()
    expect(textarea).toHaveClass('text-sm')
  })

  it('displays full placeholder text on desktop devices', () => {
    // Mock desktop viewport
    const mockMediaQuery = {
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText(/Ask your AI coach anything about training/)
    expect(textarea).toBeInTheDocument()
    expect(textarea).toHaveClass('text-base')
  })

  it('applies mobile-specific styling classes on mobile devices', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText('Ask your AI coach...')
    expect(textarea).toHaveClass('text-sm')
    expect(textarea).toHaveClass('px-3')
    expect(textarea).toHaveClass('min-h-[44px]')
  })

  it('applies desktop-specific styling classes on desktop devices', () => {
    // Mock desktop viewport
    const mockMediaQuery = {
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText(/Ask your AI coach anything about training/)
    expect(textarea).toHaveClass('text-base')
    expect(textarea).toHaveClass('px-4')
  })

  it('applies mobile touch-friendly button sizing on mobile devices', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const sendButton = screen.getByText('Send')
    expect(sendButton).toHaveClass('min-h-[44px]')
    expect(sendButton).toHaveClass('min-w-[60px]')
    expect(sendButton).toHaveClass('text-sm')
  })

  it('uses compact spacing on mobile devices', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    // Check that the input area has mobile-specific padding by finding the div with p-3 class
    const inputAreaDiv = document.querySelector('.p-3')
    expect(inputAreaDiv).toBeInTheDocument()
    
    // Check that the form has mobile-specific spacing
    const form = screen.getByRole('textbox').closest('form')
    expect(form).toHaveClass('space-x-2')
  })

  it('uses standard spacing on desktop devices', () => {
    // Mock desktop viewport
    const mockMediaQuery = {
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    // Check that the input area has desktop-specific padding by finding the div with p-4 class
    const inputAreaDiv = document.querySelector('.p-4')
    expect(inputAreaDiv).toBeInTheDocument()
    
    // Check that the form has desktop-specific spacing
    const form = screen.getByRole('textbox').closest('form')
    expect(form).toHaveClass('space-x-3')
  })

  it('maintains minimum touch target size on mobile', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText('Ask your AI coach...')
    const sendButton = screen.getByText('Send')

    // Check that both elements have minimum 44px height for touch targets
    expect(textarea).toHaveClass('min-h-[44px]')
    expect(sendButton).toHaveClass('min-h-[44px]')
  })

  it('shows abbreviated button text on mobile when streaming', () => {
    // Mock mobile viewport
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText('Ask your AI coach...')
    fireEvent.change(textarea, { target: { value: 'Test message' } })

    // Mock send message call to trigger streaming state
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    const sendButton = screen.getByText('Send')
    fireEvent.click(sendButton)

    // Should show abbreviated text on mobile when streaming
    expect(screen.getByText('...')).toBeInTheDocument()
  })

  it('shows full button text on desktop when streaming', () => {
    // Mock desktop viewport
    const mockMediaQuery = {
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    const textarea = screen.getByPlaceholderText(/Ask your AI coach anything about training/)
    fireEvent.change(textarea, { target: { value: 'Test message' } })

    // Mock send message call to trigger streaming state
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 'msg-id' })
    })

    const sendButton = screen.getByText('Send')
    fireEvent.click(sendButton)

    // Should show full text on desktop when streaming
    expect(screen.getByText('Sending...')).toBeInTheDocument()
  })

  it('handles viewport changes dynamically', async () => {
    let mediaQueryCallback: (event: MediaQueryListEvent) => void

    // Mock mobile viewport initially
    const mockMediaQuery = {
      matches: true,
      addEventListener: vi.fn((event, callback) => {
        mediaQueryCallback = callback
      }),
      removeEventListener: vi.fn(),
    }
    mockMatchMedia.mockReturnValue(mockMediaQuery)

    renderChatInterface()

    // Initially should show mobile placeholder
    expect(screen.getByPlaceholderText('Ask your AI coach...')).toBeInTheDocument()

    // Simulate viewport change to desktop
    mockMediaQuery.matches = false
    
    // Use act to wrap the state update
    await act(async () => {
      mediaQueryCallback({ matches: false } as MediaQueryListEvent)
    })

    // Should now show desktop placeholder
    expect(screen.getByPlaceholderText(/Ask your AI coach anything about training/)).toBeInTheDocument()
  })
})