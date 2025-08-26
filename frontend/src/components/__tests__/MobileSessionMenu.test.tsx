import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import MobileSessionMenu from '../MobileSessionMenu'
import { Session } from '../../services/api'

// Mock the ErrorBoundary components
vi.mock('../ErrorBoundary', () => ({
  LoadingSpinner: ({ className, size }: { className?: string; size?: string }) => (
    <div data-testid="loading-spinner" className={className} data-size={size}>
      Loading...
    </div>
  )
}))

vi.mock('../ApiErrorHandler', () => ({
  SessionErrorHandler: ({ error, onRetry, loading }: { error: unknown; onRetry: () => void; loading: boolean }) => (
    <div data-testid="session-error-handler">
      <div>Error: {String(error)}</div>
      <button onClick={onRetry} disabled={loading}>
        Retry
      </button>
    </div>
  )
}))

describe('MobileSessionMenu', () => {
  const mockSessions: Session[] = [
    {
      id: '1',
      user_id: 'user1',
      title: 'First Session',
      created_at: '2024-01-01T10:00:00Z',
      updated_at: '2024-01-01T10:00:00Z'
    },
    {
      id: '2',
      user_id: 'user1',
      title: 'Second Session',
      created_at: '2024-01-02T10:00:00Z',
      updated_at: '2024-01-02T10:00:00Z'
    }
  ]

  const defaultProps = {
    sessions: mockSessions,
    currentSessionId: '1',
    onCreateSession: vi.fn(),
    isCreatingSession: false,
    onSelectSession: vi.fn(),
    isOpen: true,
    onClose: vi.fn(),
    isLoading: false,
    error: null,
    onRetryLoad: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    // Reset body overflow style
    document.body.style.overflow = 'unset'
  })

  describe('Rendering', () => {
    it('renders nothing when isOpen is false', () => {
      render(<MobileSessionMenu {...defaultProps} isOpen={false} />)
      expect(screen.queryByText('Sessions')).not.toBeInTheDocument()
    })

    it('renders the menu when isOpen is true', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      expect(screen.getByText('Sessions')).toBeInTheDocument()
      expect(screen.getByText('New Session')).toBeInTheDocument()
      expect(screen.getByText('Recent Sessions')).toBeInTheDocument()
    })

    it('renders sessions list correctly', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      expect(screen.getByText('First Session')).toBeInTheDocument()
      expect(screen.getByText('Second Session')).toBeInTheDocument()
    })

    it('highlights current session', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const currentSessionButton = screen.getByText('First Session').closest('button')
      expect(currentSessionButton).toHaveClass('bg-blue-100', 'text-blue-900')
    })

    it('renders close button with proper aria-label', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const closeButton = screen.getByLabelText('Close session menu')
      expect(closeButton).toBeInTheDocument()
    })
  })

  describe('Touch-friendly button sizes', () => {
    it('ensures new session button has minimum 44px height', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const newSessionButton = screen.getByText('New Session')
      expect(newSessionButton).toHaveClass('min-h-[44px]')
    })

    it('ensures session buttons have minimum 44px height', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const sessionButton = screen.getByText('First Session').closest('button')
      expect(sessionButton).toHaveClass('min-h-[44px]')
    })

    it('uses appropriate font sizes for mobile', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const newSessionButton = screen.getByText('New Session')
      expect(newSessionButton).toHaveClass('text-base')
      
      const sessionTitle = screen.getByText('First Session')
      expect(sessionTitle).toHaveClass('text-base')
    })
  })

  describe('Backdrop functionality', () => {
    it('calls onClose when backdrop is clicked', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const backdrop = document.querySelector('.bg-black.bg-opacity-50')
      expect(backdrop).toBeInTheDocument()
      
      fireEvent.click(backdrop!)
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1)
    })

    it('has proper backdrop styling', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const backdrop = document.querySelector('.bg-black.bg-opacity-50')
      expect(backdrop).toHaveClass('absolute', 'inset-0', 'transition-opacity', 'duration-300')
    })
  })

  describe('Animations and styling', () => {
    it('has proper animation classes for slide-in effect', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const menuPanel = document.querySelector('.w-80')
      expect(menuPanel).toHaveClass('transform', 'transition-transform', 'duration-300', 'ease-in-out')
    })

    it('has proper z-index for overlay', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const overlay = document.querySelector('.fixed.inset-0')
      expect(overlay).toHaveClass('z-50')
    })

    it('hides on medium screens and above', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const overlay = document.querySelector('.fixed.inset-0')
      expect(overlay).toHaveClass('md:hidden')
    })
  })

  describe('Interactions', () => {
    it('calls onCreateSession when new session button is clicked', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const newSessionButton = screen.getByText('New Session')
      fireEvent.click(newSessionButton)
      expect(defaultProps.onCreateSession).toHaveBeenCalledTimes(1)
    })

    it('calls onSelectSession and onClose when session is selected', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const sessionButton = screen.getByText('Second Session')
      fireEvent.click(sessionButton)
      
      expect(defaultProps.onSelectSession).toHaveBeenCalledWith('2')
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when close button is clicked', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const closeButton = screen.getByLabelText('Close session menu')
      fireEvent.click(closeButton)
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when escape key is pressed', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      fireEvent.keyDown(document, { key: 'Escape' })
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1)
    })

    it('does not call onClose when other keys are pressed', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      fireEvent.keyDown(document, { key: 'Enter' })
      expect(defaultProps.onClose).not.toHaveBeenCalled()
    })
  })

  describe('Body scroll prevention', () => {
    it('prevents body scroll when menu is open', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      expect(document.body.style.overflow).toBe('hidden')
    })

    it('restores body scroll when menu is closed', () => {
      const { rerender } = render(<MobileSessionMenu {...defaultProps} />)
      expect(document.body.style.overflow).toBe('hidden')
      
      rerender(<MobileSessionMenu {...defaultProps} isOpen={false} />)
      expect(document.body.style.overflow).toBe('unset')
    })

    it('restores body scroll on unmount', () => {
      const { unmount } = render(<MobileSessionMenu {...defaultProps} />)
      expect(document.body.style.overflow).toBe('hidden')
      
      unmount()
      expect(document.body.style.overflow).toBe('unset')
    })
  })

  describe('Loading states', () => {
    it('shows loading spinner when isLoading is true', () => {
      render(<MobileSessionMenu {...defaultProps} isLoading={true} />)
      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
      expect(screen.getByText('Loading sessions...')).toBeInTheDocument()
    })

    it('shows creating state for new session button', () => {
      render(<MobileSessionMenu {...defaultProps} isCreatingSession={true} />)
      expect(screen.getByText('Creating...')).toBeInTheDocument()
      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })

    it('disables new session button when creating', () => {
      render(<MobileSessionMenu {...defaultProps} isCreatingSession={true} />)
      const button = screen.getByText('Creating...').closest('button')
      expect(button).toBeDisabled()
      expect(button).toHaveClass('disabled:bg-blue-400')
    })
  })

  describe('Error handling', () => {
    it('shows error handler when error exists', () => {
      const error = new Error('Test error')
      render(<MobileSessionMenu {...defaultProps} error={error} />)
      expect(screen.getByTestId('session-error-handler')).toBeInTheDocument()
      expect(screen.getByText('Error: Error: Test error')).toBeInTheDocument()
    })

    it('does not show error handler when onRetryLoad is not provided', () => {
      const error = new Error('Test error')
      render(<MobileSessionMenu {...defaultProps} error={error} onRetryLoad={undefined} />)
      expect(screen.queryByTestId('session-error-handler')).not.toBeInTheDocument()
    })
  })

  describe('Empty states', () => {
    it('shows empty state when no sessions exist', () => {
      render(<MobileSessionMenu {...defaultProps} sessions={[]} />)
      expect(screen.getByText('No sessions yet')).toBeInTheDocument()
      expect(screen.getByText('Start a new conversation!')).toBeInTheDocument()
    })

    it('shows empty state when sessions is undefined', () => {
      render(<MobileSessionMenu {...defaultProps} sessions={undefined as any} />)
      expect(screen.getByText('No sessions yet')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('has proper ARIA labels', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      expect(screen.getByLabelText('Close session menu')).toBeInTheDocument()
    })

    it('has proper heading structure', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      expect(screen.getByRole('heading', { level: 2, name: 'Sessions' })).toBeInTheDocument()
      expect(screen.getByRole('heading', { level: 3, name: 'Recent Sessions' })).toBeInTheDocument()
    })

    it('has aria-hidden on backdrop', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      const backdrop = document.querySelector('[aria-hidden="true"]')
      expect(backdrop).toBeInTheDocument()
    })
  })

  describe('Date formatting', () => {
    it('formats session dates correctly', () => {
      render(<MobileSessionMenu {...defaultProps} />)
      // The exact format depends on locale, but should contain date elements
      const dateElements = screen.getAllByText(/\d/)
      expect(dateElements.length).toBeGreaterThan(0)
    })
  })
})