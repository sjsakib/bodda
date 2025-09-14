import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import SessionSidebar from '../SessionSidebar'
import { Session } from '../../services/api'
import * as dateFormatting from '../../utils/dateFormatting'

// Mock the date formatting utility
vi.mock('../../utils/dateFormatting', () => ({
  formatSessionTimestamp: vi.fn(),
  isValidTimestamp: vi.fn(),
  isCurrentYear: vi.fn()
}))

const mockFormatSessionTimestamp = vi.mocked(dateFormatting.formatSessionTimestamp)
const mockIsValidTimestamp = vi.mocked(dateFormatting.isValidTimestamp)
const mockIsCurrentYear = vi.mocked(dateFormatting.isCurrentYear)

describe('SessionSidebar Timestamp Integration', () => {
  beforeEach(() => {
    // Reset mocks before each test
    vi.clearAllMocks()
    
    // Default mock implementations
    mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
      const date = new Date(timestamp)
      if (isNaN(date.getTime())) return 'Invalid Date'
      return `${date.getDate()} ${date.toLocaleDateString('en-US', { month: 'short' })}, ${date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })}`
    })
    
    mockIsValidTimestamp.mockImplementation((timestamp: string) => {
      return !isNaN(new Date(timestamp).getTime())
    })
    
    mockIsCurrentYear.mockImplementation((timestamp: string) => {
      const date = new Date(timestamp)
      return date.getFullYear() === new Date().getFullYear()
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('Current Year Sessions', () => {
    const currentYearSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Current Year Session',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: '2',
        user_id: 'user1',
        title: 'Another Current Year Session',
        created_at: '2024-08-15T14:30:00Z',
        updated_at: '2024-08-15T14:30:00Z'
      }
    ]

    const defaultProps = {
      sessions: currentYearSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('displays formatted timestamps for current year sessions', () => {
      mockFormatSessionTimestamp
        .mockReturnValueOnce('3 Sep, 08:20 pm')
        .mockReturnValueOnce('15 Aug, 02:30 pm')

      render(<SessionSidebar {...defaultProps} />)
      
      // Should display formatted timestamps without year
      expect(screen.getByText('3 Sep, 08:20 pm')).toBeInTheDocument()
      expect(screen.getByText('15 Aug, 02:30 pm')).toBeInTheDocument()
      
      // Should not display original titles
      expect(screen.queryByText('Current Year Session')).not.toBeInTheDocument()
      expect(screen.queryByText('Another Current Year Session')).not.toBeInTheDocument()
    })

    it('calls formatSessionTimestamp with correct parameters', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T20:20:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-08-15T14:30:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledTimes(2)
    })
  })

  describe('Previous Year Sessions', () => {
    const previousYearSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Previous Year Session',
        created_at: '2023-12-25T10:00:00Z',
        updated_at: '2023-12-25T10:00:00Z'
      },
      {
        id: '2',
        user_id: 'user1',
        title: 'Another Previous Year Session',
        created_at: '2022-06-15T16:45:00Z',
        updated_at: '2022-06-15T16:45:00Z'
      }
    ]

    const defaultProps = {
      sessions: previousYearSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('displays formatted timestamps with year for previous year sessions', () => {
      mockFormatSessionTimestamp
        .mockReturnValueOnce('25 Dec 2023, 10:00 am')
        .mockReturnValueOnce('15 Jun 2022, 04:45 pm')

      render(<SessionSidebar {...defaultProps} />)
      
      // Should display formatted timestamps with year
      expect(screen.getByText('25 Dec 2023, 10:00 am')).toBeInTheDocument()
      expect(screen.getByText('15 Jun 2022, 04:45 pm')).toBeInTheDocument()
    })
  })

  describe('Mixed Year Sessions', () => {
    const mixedYearSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Current Year Session',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: '2',
        user_id: 'user1',
        title: 'Previous Year Session',
        created_at: '2023-12-25T10:00:00Z',
        updated_at: '2023-12-25T10:00:00Z'
      }
    ]

    const defaultProps = {
      sessions: mixedYearSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('displays mixed year sessions with appropriate formatting', () => {
      mockFormatSessionTimestamp
        .mockReturnValueOnce('3 Sep, 08:20 pm')
        .mockReturnValueOnce('25 Dec 2023, 10:00 am')

      render(<SessionSidebar {...defaultProps} />)
      
      // Current year without year suffix
      expect(screen.getByText('3 Sep, 08:20 pm')).toBeInTheDocument()
      // Previous year with year suffix
      expect(screen.getByText('25 Dec 2023, 10:00 am')).toBeInTheDocument()
    })
  })

  describe('Error Handling', () => {
    const invalidTimestampSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Invalid Timestamp Session',
        created_at: 'invalid-timestamp',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: '2',
        user_id: 'user1',
        title: 'Empty Timestamp Session',
        created_at: '',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: '3',
        user_id: 'user1',
        title: 'Valid Session',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      }
    ]

    const defaultProps = {
      sessions: invalidTimestampSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('handles invalid timestamps gracefully', () => {
      mockFormatSessionTimestamp
        .mockReturnValueOnce('Invalid Date')
        .mockReturnValueOnce('Invalid Date')
        .mockReturnValueOnce('3 Sep, 08:20 pm')

      render(<SessionSidebar {...defaultProps} />)
      
      // Should display fallback text for invalid timestamps
      const invalidDateElements = screen.getAllByText('Invalid Date')
      expect(invalidDateElements).toHaveLength(2)
      
      // Should still display valid timestamp
      expect(screen.getByText('3 Sep, 08:20 pm')).toBeInTheDocument()
    })

    it('handles missing created_at field', () => {
      const sessionWithoutTimestamp: Session[] = [
        {
          id: '1',
          user_id: 'user1',
          title: 'No Timestamp Session',
          created_at: undefined as any,
          updated_at: '2024-09-03T20:20:00Z'
        }
      ]

      mockFormatSessionTimestamp.mockReturnValue('Invalid Date')

      render(<SessionSidebar {...defaultProps} sessions={sessionWithoutTimestamp} />)
      
      expect(screen.getByText('Invalid Date')).toBeInTheDocument()
    })
  })

  describe('Accessibility Features', () => {
    const mockSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Session 1',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      }
    ]

    const defaultProps = {
      sessions: mockSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('includes full timestamp information in aria-labels', () => {
      mockFormatSessionTimestamp.mockReturnValue('3 Sep, 08:20 pm')

      render(<SessionSidebar {...defaultProps} />)
      
      const sessionButton = screen.getByRole('option')
      const ariaLabel = sessionButton.getAttribute('aria-label')
      
      expect(ariaLabel).toMatch(/Chat session from.*currently selected/)
      expect(ariaLabel).toContain('Chat session from')
    })

    it('provides screen reader descriptions with full timestamp', () => {
      mockFormatSessionTimestamp.mockReturnValue('3 Sep, 08:20 pm')

      render(<SessionSidebar {...defaultProps} />)
      
      const description = screen.getByText(/Session created on/)
      expect(description).toBeInTheDocument()
      expect(description).toHaveClass('sr-only')
    })
  })

  describe('Responsive Design', () => {
    const mockSessions: Session[] = [
      {
        id: '1',
        user_id: 'user1',
        title: 'Long Session Title That Should Be Truncated',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      }
    ]

    const defaultProps = {
      sessions: mockSessions,
      currentSessionId: '1',
      onCreateSession: vi.fn(),
      isCreatingSession: false,
      onSelectSession: vi.fn(),
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('handles text overflow with truncation classes', () => {
      mockFormatSessionTimestamp.mockReturnValue('Very Long Timestamp That Should Be Truncated In Small Screens')

      render(<SessionSidebar {...defaultProps} />)
      
      const timestampElement = screen.getByText('Very Long Timestamp That Should Be Truncated In Small Screens')
      expect(timestampElement).toHaveClass('truncate', 'min-w-0')
    })

    it('maintains responsive sidebar classes', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sidebar = screen.getByRole('complementary')
      expect(sidebar).toHaveClass('w-full', 'sm:w-1/3', 'lg:w-1/4')
    })
  })

  describe('Component Integration', () => {
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
      onDeleteSession: vi.fn(),
      isLoading: false,
      error: null
    }

    it('maintains existing CSS classes and styling', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      // Check that the main container has the expected classes
      const container = screen.getByRole('listbox')
      expect(container).toHaveClass('space-y-1')
      
      // Check that session buttons have the expected classes
      const sessionButtons = screen.getAllByRole('option')
      sessionButtons.forEach(button => {
        expect(button).toHaveClass('w-full', 'text-left', 'p-3', 'rounded-lg', 'transition-colors')
      })
    })

    it('handles empty sessions list correctly', () => {
      render(<SessionSidebar {...defaultProps} sessions={[]} />)
      
      expect(screen.getByText('No sessions yet')).toBeInTheDocument()
      expect(screen.getByText('Start a new conversation!')).toBeInTheDocument()
    })

    it('handles session selection correctly', () => {
      const mockOnSelect = vi.fn()
      render(<SessionSidebar {...defaultProps} onSelectSession={mockOnSelect} />)
      
      const sessionButtons = screen.getAllByRole('option')
      fireEvent.click(sessionButtons[1])
      
      expect(mockOnSelect).toHaveBeenCalledWith('2')
    })

    it('shows selected state correctly', () => {
      render(<SessionSidebar {...defaultProps} />)
      
      const sessionButtons = screen.getAllByRole('option')
      
      // First session should be selected
      expect(sessionButtons[0]).toHaveAttribute('aria-selected', 'true')
      expect(sessionButtons[0]).toHaveClass('bg-blue-100', 'text-blue-900')
      
      // Second session should not be selected
      expect(sessionButtons[1]).toHaveAttribute('aria-selected', 'false')
      expect(sessionButtons[1]).not.toHaveClass('bg-blue-100', 'text-blue-900')
    })
  })
})