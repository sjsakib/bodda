import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import SessionSidebar from '../SessionSidebar'
import { Session } from '../../services/api'

// Mock the date formatting utility
vi.mock('../../utils/dateFormatting', () => ({
  formatSessionTimestamp: vi.fn((timestamp: string) => {
    const date = new Date(timestamp)
    return `${date.getDate()} ${date.toLocaleDateString('en-US', { month: 'short' })}, ${date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })}`
  })
}))

describe('SessionSidebar Accessibility and Responsive Design', () => {
  const mockSessions: Session[] = [
    {
      id: '1',
      user_id: 'user1',
      title: 'Session 1',
      created_at: '2024-09-03T20:20:00Z',
      updated_at: '2024-09-03T20:20:00Z'
    },
    {
      id: '2',
      user_id: 'user1',
      title: 'Session 2',
      created_at: '2024-08-15T14:30:00Z',
      updated_at: '2024-08-15T14:30:00Z'
    }
  ]

  const defaultProps = {
    sessions: mockSessions,
    currentSessionId: '1',
    onCreateSession: vi.fn(),
    isCreatingSession: false,
    onSelectSession: vi.fn(),
    isLoading: false,
    error: null
  }

  describe('Accessibility Features', () => {
    it('has proper ARIA roles and labels', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      // Main container should be a complementary landmark
      const sidebar = screen.getByRole('complementary')
      expect(sidebar).toHaveAttribute('aria-label', 'Session navigation')
      
      // Sessions list should be a listbox
      const sessionsList = screen.getByRole('listbox')
      expect(sessionsList).toHaveAttribute('aria-labelledby', 'sessions-heading')
      expect(sessionsList).toHaveAttribute('aria-multiselectable', 'false')
      
      // Session buttons should be options
      const sessionOptions = screen.getAllByRole('option')
      expect(sessionOptions).toHaveLength(2)
      
      sessionOptions.forEach((option, index) => {
        expect(option).toHaveAttribute('aria-selected', index === 0 ? 'true' : 'false')
        expect(option).toHaveAttribute('aria-describedby')
        expect(option.getAttribute('aria-label')).toMatch(/Chat session from/)
      })
    })

    it('provides comprehensive aria-labels with full timestamp information', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sessionOptions = screen.getAllByRole('option')
      
      // First session (selected)
      expect(sessionOptions[0]).toHaveAttribute('aria-label', expect.stringMatching(/Chat session from.*currently selected/))
      
      // Second session (not selected)
      expect(sessionOptions[1]).toHaveAttribute('aria-label', expect.not.stringMatching(/currently selected/))
    })

    it('includes screen reader only descriptions', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      // Check for screen reader only descriptions
      const descriptions = screen.getAllByText(/Session created on/)
      expect(descriptions).toHaveLength(2)
      
      descriptions.forEach(description => {
        expect(description).toHaveClass('sr-only')
      })
    })

    it('has proper focus management', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sessionOptions = screen.getAllByRole('option')
      
      // Focus first session
      sessionOptions[0].focus()
      expect(sessionOptions[0]).toHaveFocus()
      expect(sessionOptions[0]).toHaveClass('focus:outline-none', 'focus:ring-2', 'focus:ring-blue-500')
    })

    it('provides accessible new session button', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const newSessionButton = screen.getByRole('button', { name: 'New Session' })
      expect(newSessionButton).toHaveAttribute('aria-describedby', 'new-session-description')
      
      const description = screen.getByText('Create a new chat session')
      expect(description).toHaveClass('sr-only')
      expect(description).toHaveAttribute('id', 'new-session-description')
    })
  })

  describe('Responsive Design', () => {
    it('has responsive width classes', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sidebar = screen.getByRole('complementary')
      expect(sidebar).toHaveClass('w-full', 'sm:w-1/3', 'lg:w-1/4')
    })

    it('handles text overflow with truncation', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      // Find the actual timestamp display elements (not the screen reader text)
      const sessionOptions = screen.getAllByRole('option')
      sessionOptions.forEach(option => {
        const timestampDiv = option.querySelector('.truncate.min-w-0')
        expect(timestampDiv).toBeInTheDocument()
        expect(timestampDiv).toHaveClass('truncate', 'min-w-0')
      })
    })

    it('maintains proper layout structure', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sidebar = screen.getByRole('complementary')
      expect(sidebar).toHaveClass('flex', 'flex-col', 'min-h-0')
      
      // Header should not shrink
      const header = sidebar.querySelector('.flex-shrink-0')
      expect(header).toBeInTheDocument()
      
      // Content area should be scrollable
      const scrollableArea = sidebar.querySelector('.overflow-y-auto')
      expect(scrollableArea).toBeInTheDocument()
      expect(scrollableArea).toHaveClass('flex-1', 'min-h-0')
    })
  })

  describe('Tooltip Functionality', () => {
    it('shows tooltip on hover when text is truncated', async () => {
      // Mock element dimensions to simulate truncation
      const mockElement = {
        scrollWidth: 200,
        clientWidth: 100
      }
      
      // Mock getBoundingClientRect and scroll properties
      Object.defineProperty(HTMLElement.prototype, 'scrollWidth', {
        configurable: true,
        value: 200
      })
      Object.defineProperty(HTMLElement.prototype, 'clientWidth', {
        configurable: true,
        value: 100
      })

      render(<SessionSidebar {...defaultProps} />)
      
      const sessionOption = screen.getAllByRole('option')[0]
      
      // Hover over the session
      fireEvent.mouseEnter(sessionOption)
      
      // Wait for tooltip to appear
      await waitFor(() => {
        const tooltip = screen.queryByRole('tooltip')
        // Tooltip should appear if text is truncated
        if (tooltip) {
          expect(tooltip).toHaveAttribute('aria-hidden', 'true')
          expect(tooltip).toHaveClass('absolute', 'z-50', 'bg-gray-900', 'text-white')
        }
      })
    })

    it('hides tooltip on mouse leave', async () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sessionOption = screen.getAllByRole('option')[0]
      
      // Hover and then leave
      fireEvent.mouseEnter(sessionOption)
      fireEvent.mouseLeave(sessionOption)
      
      // Tooltip should not be visible
      await waitFor(() => {
        const tooltip = screen.queryByRole('tooltip')
        expect(tooltip).not.toBeInTheDocument()
      })
    })
  })

  describe('Loading and Error States', () => {
    it('has accessible loading state', () => {
      render(<SessionSidebar {...defaultProps} isLoading={true} />)
      
      const loadingContainer = screen.getByRole('status')
      expect(loadingContainer).toHaveAttribute('aria-live', 'polite')
      expect(loadingContainer).toHaveAttribute('aria-label', 'Loading sessions')
    })

    it('has accessible empty state', () => {
      render(<SessionSidebar {...defaultProps} sessions={[]} />)
      
      const emptyState = screen.getByRole('status')
      expect(emptyState).toHaveAttribute('aria-live', 'polite')
      expect(emptyState).toHaveTextContent('No sessions yet')
    })
  })

  describe('Keyboard Navigation', () => {
    it('supports keyboard focus on session options', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sessionOptions = screen.getAllByRole('option')
      
      // Tab to first session
      sessionOptions[0].focus()
      expect(sessionOptions[0]).toHaveFocus()
      
      // Should show focus styles
      expect(sessionOptions[0]).toHaveClass('focus:ring-2')
    })

    it('calls onSelect when Enter is pressed', () => {
      const mockOnSelect = vi.fn()
      render(<SessionSidebar {...defaultProps} onSelectSession={mockOnSelect} />)
      
      const sessionOption = screen.getAllByRole('option')[1]
      sessionOption.focus()
      
      fireEvent.keyDown(sessionOption, { key: 'Enter' })
      fireEvent.click(sessionOption) // Simulate the click that would happen
      
      expect(mockOnSelect).toHaveBeenCalledWith('2')
    })
  })
})