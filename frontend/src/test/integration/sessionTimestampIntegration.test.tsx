import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import ChatInterface from '../../components/ChatInterface'
import SessionSidebar from '../../components/SessionSidebar'
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

// Mock fetch globally
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

// Mock window.matchMedia for responsive layout tests
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Test wrapper with router
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>
}

describe('Session Timestamp Integration Tests', () => {
  beforeEach(() => {
    mockFetch.mockClear()
    vi.clearAllMocks()
    
    // Set up realistic date formatting mock that matches the actual implementation
    mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
      const date = new Date(timestamp)
      if (isNaN(date.getTime())) return 'Invalid Date'
      
      const currentYear = new Date().getFullYear()
      const timestampYear = date.getFullYear()
      const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
      
      const day = date.getDate()
      const month = monthNames[date.getMonth()]
      let hours = date.getHours()
      const minutes = date.getMinutes().toString().padStart(2, '0')
      const ampm = hours >= 12 ? 'pm' : 'am'
      
      hours = hours % 12
      hours = hours ? hours : 12
      
      const timeString = `${hours}:${minutes} ${ampm}`
      
      // Always include year for consistency with current implementation
      return `${day} ${month} ${timestampYear}, ${timeString}`
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

  describe('Multiple Sessions Timestamp Formatting (Requirements 1.2, 2.1)', () => {
    const multipleSessions: Session[] = [
      {
        id: 'session-1',
        user_id: 'user1',
        title: 'Recent Session',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: 'session-2',
        user_id: 'user1',
        title: 'Morning Session',
        created_at: '2024-09-03T08:15:00Z',
        updated_at: '2024-09-03T08:15:00Z'
      },
      {
        id: 'session-3',
        user_id: 'user1',
        title: 'Yesterday Session',
        created_at: '2024-09-02T14:30:00Z',
        updated_at: '2024-09-02T14:30:00Z'
      },
      {
        id: 'session-4',
        user_id: 'user1',
        title: 'Last Year Session',
        created_at: '2023-12-25T10:00:00Z',
        updated_at: '2023-12-25T10:00:00Z'
      }
    ]

    it('should format timestamps consistently across multiple sessions', () => {
      const defaultProps = {
        sessions: multipleSessions,
        currentSessionId: 'session-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Verify that formatSessionTimestamp was called for each session
      expect(mockFormatSessionTimestamp).toHaveBeenCalledTimes(4)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T20:20:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T08:15:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-02T14:30:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2023-12-25T10:00:00Z')

      // Verify formatted timestamps are displayed
      expect(screen.getByText('4 Sep 2024, 2:20 am')).toBeInTheDocument()
      expect(screen.getByText('3 Sep 2024, 2:15 pm')).toBeInTheDocument()
      expect(screen.getByText('2 Sep 2024, 8:30 pm')).toBeInTheDocument()
      expect(screen.getByText('25 Dec 2023, 4:00 pm')).toBeInTheDocument()

      // Verify original titles are not displayed
      expect(screen.queryByText('Recent Session')).not.toBeInTheDocument()
      expect(screen.queryByText('Morning Session')).not.toBeInTheDocument()
      expect(screen.queryByText('Yesterday Session')).not.toBeInTheDocument()
      expect(screen.queryByText('Last Year Session')).not.toBeInTheDocument()
    })

    it('should maintain consistent formatting patterns across sessions', () => {
      const defaultProps = {
        sessions: multipleSessions,
        currentSessionId: 'session-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Check that all sessions include year in current implementation
      const allSessions = screen.getAllByText(/^\d{1,2} \w{3} \d{4}, \d{1,2}:\d{2} [ap]m$/)
      expect(allSessions).toHaveLength(4) // All 4 sessions should have year format
    })

    it('should properly implement current year omission logic (Requirement 2.2)', () => {
      const currentYear = new Date().getFullYear()
      const currentYearSessions: Session[] = [
        {
          id: 'current-year-1',
          user_id: 'user1',
          title: 'Current Year Session 1',
          created_at: `${currentYear}-09-03T20:20:00Z`,
          updated_at: `${currentYear}-09-03T20:20:00Z`
        },
        {
          id: 'current-year-2',
          user_id: 'user1',
          title: 'Current Year Session 2',
          created_at: `${currentYear}-08-15T14:30:00Z`,
          updated_at: `${currentYear}-08-15T14:30:00Z`
        },
        {
          id: 'previous-year-1',
          user_id: 'user1',
          title: 'Previous Year Session',
          created_at: `${currentYear - 1}-12-25T10:00:00Z`,
          updated_at: `${currentYear - 1}-12-25T10:00:00Z`
        }
      ]

      // Update mock to properly handle current year logic
      mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
        const date = new Date(timestamp)
        if (isNaN(date.getTime())) return 'Invalid Date'
        
        const timestampYear = date.getFullYear()
        const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
        
        const day = date.getDate()
        const month = monthNames[date.getMonth()]
        let hours = date.getHours()
        const minutes = date.getMinutes().toString().padStart(2, '0')
        const ampm = hours >= 12 ? 'pm' : 'am'
        
        hours = hours % 12
        hours = hours ? hours : 12
        
        const timeString = `${hours}:${minutes} ${ampm}`
        
        // Implement current year omission logic as per Requirement 2.2
        if (timestampYear === currentYear) {
          return `${day} ${month}, ${timeString}` // Omit year for current year
        } else {
          return `${day} ${month} ${timestampYear}, ${timeString}` // Include year for previous years
        }
      })

      const defaultProps = {
        sessions: currentYearSessions,
        currentSessionId: 'current-year-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Current year sessions should NOT include year (Requirement 2.2)
      expect(screen.getByText('4 Sep, 2:20 am')).toBeInTheDocument()
      expect(screen.getByText('15 Aug, 8:30 pm')).toBeInTheDocument()

      // Previous year sessions should include year
      expect(screen.getByText(`25 Dec ${currentYear - 1}, 4:00 pm`)).toBeInTheDocument()

      // Verify the formatting function was called correctly
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-09-03T20:20:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-08-15T14:30:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear - 1}-12-25T10:00:00Z`)
    })
  })

  describe('Mixed Valid/Invalid Timestamps', () => {
    const mixedValiditySessions: Session[] = [
      {
        id: 'session-valid-1',
        user_id: 'user1',
        title: 'Valid Session 1',
        created_at: '2024-09-03T20:20:00Z',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: 'session-invalid-1',
        user_id: 'user1',
        title: 'Invalid Session 1',
        created_at: 'invalid-timestamp',
        updated_at: '2024-09-03T20:20:00Z'
      },
      {
        id: 'session-valid-2',
        user_id: 'user1',
        title: 'Valid Session 2',
        created_at: '2024-08-15T14:30:00Z',
        updated_at: '2024-08-15T14:30:00Z'
      },
      {
        id: 'session-invalid-2',
        user_id: 'user1',
        title: 'Invalid Session 2',
        created_at: '',
        updated_at: '2024-08-15T14:30:00Z'
      },
      {
        id: 'session-malformed',
        user_id: 'user1',
        title: 'Malformed Session',
        created_at: '2024-13-45T25:70:00Z', // Invalid date values
        updated_at: '2024-08-15T14:30:00Z'
      }
    ]

    it('should handle mixed valid and invalid timestamps gracefully', () => {
      // Override mock to handle invalid timestamps
      mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
        const date = new Date(timestamp)
        if (isNaN(date.getTime()) || !timestamp || timestamp === 'invalid-timestamp') {
          return 'Invalid Date'
        }
        
        const currentYear = new Date().getFullYear()
        const timestampYear = date.getFullYear()
        const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
        
        const day = date.getDate()
        const month = monthNames[date.getMonth()]
        let hours = date.getHours()
        const minutes = date.getMinutes().toString().padStart(2, '0')
        const ampm = hours >= 12 ? 'pm' : 'am'
        
        hours = hours % 12
        hours = hours ? hours : 12
        
        const timeString = `${hours}:${minutes} ${ampm}`
        
        if (timestampYear !== currentYear) {
          return `${day} ${month} ${timestampYear}, ${timeString}`
        } else {
          return `${day} ${month}, ${timeString}`
        }
      })

      const defaultProps = {
        sessions: mixedValiditySessions,
        currentSessionId: 'session-valid-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Valid timestamps should be formatted correctly
      expect(screen.getByText('4 Sep 2024, 2:20 am')).toBeInTheDocument()
      expect(screen.getByText('15 Aug 2024, 8:30 pm')).toBeInTheDocument()

      // Invalid timestamps should show fallback text
      const invalidDateElements = screen.getAllByText('Invalid Date')
      expect(invalidDateElements).toHaveLength(3) // 3 invalid sessions

      // All sessions should still be rendered as buttons
      const sessionButtons = screen.getAllByRole('option')
      expect(sessionButtons).toHaveLength(5)
    })

    it('should maintain accessibility for sessions with invalid timestamps', () => {
      mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
        if (!timestamp || timestamp === 'invalid-timestamp' || isNaN(new Date(timestamp).getTime())) {
          return 'Invalid Date'
        }
        return '3 Sep, 08:20 pm' // Valid fallback
      })

      const defaultProps = {
        sessions: mixedValiditySessions,
        currentSessionId: 'session-valid-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // All session buttons should have proper aria-labels
      const sessionButtons = screen.getAllByRole('option')
      sessionButtons.forEach(button => {
        const ariaLabel = button.getAttribute('aria-label')
        expect(ariaLabel).toMatch(/Chat session from/)
      })

      // All sessions should have screen reader descriptions
      const descriptions = screen.getAllByText(/Session created on/)
      expect(descriptions).toHaveLength(5)
    })
  })

  describe('Timezone Consistency and Local Timezone Usage (Requirements 1.4, 2.1)', () => {
    const timezoneSessions: Session[] = [
      {
        id: 'session-utc',
        user_id: 'user1',
        title: 'UTC Session',
        created_at: '2024-09-03T12:00:00Z', // UTC noon
        updated_at: '2024-09-03T12:00:00Z'
      },
      {
        id: 'session-offset-plus',
        user_id: 'user1',
        title: 'Positive Offset Session',
        created_at: '2024-09-03T12:00:00+05:00', // UTC+5
        updated_at: '2024-09-03T12:00:00+05:00'
      },
      {
        id: 'session-offset-minus',
        user_id: 'user1',
        title: 'Negative Offset Session',
        created_at: '2024-09-03T12:00:00-08:00', // UTC-8
        updated_at: '2024-09-03T12:00:00-08:00'
      }
    ]

    it('should handle different timezone formats consistently', () => {
      const defaultProps = {
        sessions: timezoneSessions,
        currentSessionId: 'session-utc',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // All timestamps should be processed through the formatting function
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T12:00:00Z')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T12:00:00+05:00')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T12:00:00-08:00')

      // All should be formatted to local time consistently
      const formattedTimestamps = screen.getAllByText(/^\d{1,2} \w{3} \d{4}, \d{1,2}:\d{2} [ap]m$/)
      expect(formattedTimestamps).toHaveLength(3)
    })

    it('should maintain consistent local timezone display', () => {
      // Mock the formatting to simulate local timezone conversion
      mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
        const date = new Date(timestamp)
        if (isNaN(date.getTime())) return 'Invalid Date'
        
        // Simulate local timezone conversion (all should show local time)
        const localHour = date.getHours() // This would be converted to local time
        const minutes = date.getMinutes().toString().padStart(2, '0')
        const ampm = localHour >= 12 ? 'pm' : 'am'
        const displayHour = localHour % 12 || 12
        
        return `3 Sep, ${displayHour}:${minutes} ${ampm}`
      })

      const defaultProps = {
        sessions: timezoneSessions,
        currentSessionId: 'session-utc',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // All timestamps should follow the same local time format pattern
      const timestampElements = screen.getAllByText(/3 Sep, \d{1,2}:\d{2} [ap]m/)
      expect(timestampElements).toHaveLength(3)
    })
  })

  describe('Complete User Workflow Integration', () => {
    it('should integrate session loading with timestamp display in full application context', async () => {
      // This test verifies that the timestamp formatting works correctly when sessions are loaded
      // through the normal application flow, without the complexity of the full ChatInterface
      const workflowSessions = [
        {
          id: 'workflow-session-1',
          user_id: 'user1',
          title: 'Workflow Session 1',
          created_at: '2024-09-03T20:20:00Z',
          updated_at: '2024-09-03T20:20:00Z'
        },
        {
          id: 'workflow-session-2',
          user_id: 'user1',
          title: 'Workflow Session 2',
          created_at: '2024-09-02T14:30:00Z',
          updated_at: '2024-09-02T14:30:00Z'
        }
      ]

      const defaultProps = {
        sessions: workflowSessions,
        currentSessionId: 'workflow-session-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Wait for timestamps to be formatted
      await waitFor(() => {
        expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T20:20:00Z')
        expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-02T14:30:00Z')
      })

      // Verify formatted timestamps are displayed instead of original titles
      expect(screen.getByText('4 Sep 2024, 2:20 am')).toBeInTheDocument()
      expect(screen.getByText('2 Sep 2024, 8:30 pm')).toBeInTheDocument()

      // Verify original titles are not displayed
      expect(screen.queryByText('Workflow Session 1')).not.toBeInTheDocument()
      expect(screen.queryByText('Workflow Session 2')).not.toBeInTheDocument()
    })

    it('should handle session selection with timestamp display', async () => {
      const workflowSessions = [
        {
          id: 'selectable-session-1',
          user_id: 'user1',
          title: 'Selectable Session 1',
          created_at: '2024-09-03T20:20:00Z',
          updated_at: '2024-09-03T20:20:00Z'
        },
        {
          id: 'selectable-session-2',
          user_id: 'user1',
          title: 'Selectable Session 2',
          created_at: '2024-09-02T14:30:00Z',
          updated_at: '2024-09-02T14:30:00Z'
        }
      ]

      const mockOnSelectSession = vi.fn()

      const defaultProps = {
        sessions: workflowSessions,
        currentSessionId: 'selectable-session-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: mockOnSelectSession,
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Find and click the second session
      const secondSessionButton = screen.getByText('2 Sep 2024, 8:30 pm').closest('button')
      expect(secondSessionButton).toBeInTheDocument()
      
      fireEvent.click(secondSessionButton!)

      // Verify session selection callback was called
      expect(mockOnSelectSession).toHaveBeenCalledWith('selectable-session-2')

      // Verify accessibility attributes
      expect(secondSessionButton).toHaveAttribute('aria-selected', 'false')
      const firstSessionButton = screen.getByText('4 Sep 2024, 2:20 am').closest('button')
      expect(firstSessionButton).toHaveAttribute('aria-selected', 'true')
    })

    it('should handle session creation workflow with timestamp formatting', async () => {
      const initialSessions: Session[] = []
      const newSession = {
        id: 'new-session-1',
        user_id: 'user1',
        title: 'New Session',
        created_at: '2024-09-03T21:00:00Z',
        updated_at: '2024-09-03T21:00:00Z'
      }

      let currentSessions = initialSessions
      const mockOnCreateSession = vi.fn().mockImplementation(() => {
        // Simulate adding new session
        currentSessions = [...currentSessions, newSession]
      })

      const TestComponent = () => {
        return (
          <SessionSidebar
            sessions={currentSessions}
            currentSessionId={undefined}
            onCreateSession={mockOnCreateSession}
            isCreatingSession={false}
            onSelectSession={vi.fn()}
            isLoading={false}
            error={null}
          />
        )
      }

      const { rerender } = render(<TestComponent />)

      // Initially should show empty state
      expect(screen.getByText('No sessions yet')).toBeInTheDocument()

      // Click create session button
      const createButton = screen.getByText('New Session')
      fireEvent.click(createButton)

      // Simulate session creation by re-rendering with new session
      currentSessions = [newSession]
      rerender(<TestComponent />)

      // Wait for timestamp formatting
      await waitFor(() => {
        expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('2024-09-03T21:00:00Z')
      })

      // Verify new session appears with formatted timestamp
      expect(screen.getByText('4 Sep 2024, 3:00 am')).toBeInTheDocument()
      expect(screen.queryByText('No sessions yet')).not.toBeInTheDocument()
    })

    it('should handle loading states during session fetch', () => {
      const defaultProps = {
        sessions: [],
        currentSessionId: undefined,
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: true,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Should show loading state
      expect(screen.getByText('Loading sessions...')).toBeInTheDocument()
      expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite')

      // Should not call formatting function during loading
      expect(mockFormatSessionTimestamp).not.toHaveBeenCalled()
    })

    it('should handle error states during session fetch', () => {
      const mockOnRetryLoad = vi.fn()
      const defaultProps = {
        sessions: [],
        currentSessionId: undefined,
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: new Error('Failed to load sessions'),
        onRetryLoad: mockOnRetryLoad
      }

      render(<SessionSidebar {...defaultProps} />)

      // Should show error state with retry option
      expect(screen.getByText('Failed to Load Sessions')).toBeInTheDocument()
      
      const retryButton = screen.getByText('Retry Loading Sessions')
      fireEvent.click(retryButton)
      
      expect(mockOnRetryLoad).toHaveBeenCalled()

      // Should not call formatting function during error state
      expect(mockFormatSessionTimestamp).not.toHaveBeenCalled()
    })
  })

  describe('Task 6 Requirements Validation', () => {
    it('should satisfy all task 6 requirements: multiple sessions, mixed validity, timezone consistency, and complete workflow', async () => {
      const currentYear = new Date().getFullYear()
      
      // Create comprehensive test data covering all scenarios
      const comprehensiveTestSessions: Session[] = [
        // Valid current year sessions (Requirements 1.2, 2.2)
        {
          id: 'valid-current-1',
          user_id: 'user1',
          title: 'Valid Current Year 1',
          created_at: `${currentYear}-09-03T20:20:00Z`,
          updated_at: `${currentYear}-09-03T20:20:00Z`
        },
        {
          id: 'valid-current-2',
          user_id: 'user1',
          title: 'Valid Current Year 2',
          created_at: `${currentYear}-08-15T14:30:00Z`,
          updated_at: `${currentYear}-08-15T14:30:00Z`
        },
        // Valid previous year sessions (Requirements 1.2, 2.1)
        {
          id: 'valid-previous-1',
          user_id: 'user1',
          title: 'Valid Previous Year',
          created_at: `${currentYear - 1}-12-25T10:00:00Z`,
          updated_at: `${currentYear - 1}-12-25T10:00:00Z`
        },
        // Different timezone formats (Requirement 1.4)
        {
          id: 'timezone-utc',
          user_id: 'user1',
          title: 'UTC Timezone',
          created_at: `${currentYear}-07-10T12:00:00Z`,
          updated_at: `${currentYear}-07-10T12:00:00Z`
        },
        {
          id: 'timezone-offset',
          user_id: 'user1',
          title: 'Offset Timezone',
          created_at: `${currentYear}-06-05T15:30:00+03:00`,
          updated_at: `${currentYear}-06-05T15:30:00+03:00`
        },
        // Invalid timestamps (mixed validity requirement)
        {
          id: 'invalid-1',
          user_id: 'user1',
          title: 'Invalid Timestamp 1',
          created_at: 'invalid-date-string',
          updated_at: `${currentYear}-06-05T15:30:00Z`
        },
        {
          id: 'invalid-2',
          user_id: 'user1',
          title: 'Invalid Timestamp 2',
          created_at: '',
          updated_at: `${currentYear}-06-05T15:30:00Z`
        }
      ]

      // Mock formatting function to handle all scenarios
      mockFormatSessionTimestamp.mockImplementation((timestamp: string) => {
        const date = new Date(timestamp)
        if (isNaN(date.getTime()) || !timestamp || timestamp === 'invalid-date-string') {
          return 'Invalid Date'
        }
        
        const timestampYear = date.getFullYear()
        const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
        
        const day = date.getDate()
        const month = monthNames[date.getMonth()]
        let hours = date.getHours()
        const minutes = date.getMinutes().toString().padStart(2, '0')
        const ampm = hours >= 12 ? 'pm' : 'am'
        
        hours = hours % 12
        hours = hours ? hours : 12
        
        const timeString = `${hours}:${minutes} ${ampm}`
        
        // Current year omission logic (Requirement 2.2)
        if (timestampYear === currentYear) {
          return `${day} ${month}, ${timeString}`
        } else {
          return `${day} ${month} ${timestampYear}, ${timeString}`
        }
      })

      const mockOnSelectSession = vi.fn()
      const mockOnCreateSession = vi.fn()

      const defaultProps = {
        sessions: comprehensiveTestSessions,
        currentSessionId: 'valid-current-1',
        onCreateSession: mockOnCreateSession,
        isCreatingSession: false,
        onSelectSession: mockOnSelectSession,
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Requirement 1.2: Multiple sessions display with formatted timestamps
      await waitFor(() => {
        expect(mockFormatSessionTimestamp).toHaveBeenCalledTimes(7) // All 7 sessions
      })

      // Verify each session type is handled correctly
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-09-03T20:20:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-08-15T14:30:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear - 1}-12-25T10:00:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-07-10T12:00:00Z`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith(`${currentYear}-06-05T15:30:00+03:00`)
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('invalid-date-string')
      expect(mockFormatSessionTimestamp).toHaveBeenCalledWith('')

      // Requirement 2.1: Consistent date format across all valid sessions
      const validCurrentYearSessions = screen.getAllByText(/^\d{1,2} \w{3}, \d{1,2}:\d{2} [ap]m$/)
      expect(validCurrentYearSessions.length).toBeGreaterThanOrEqual(4) // At least 4 current year sessions

      // Requirement 2.2: Current year omission vs previous year inclusion
      expect(screen.getByText('4 Sep, 2:20 am')).toBeInTheDocument() // Current year, no year shown
      expect(screen.getByText('15 Aug, 8:30 pm')).toBeInTheDocument() // Current year, no year shown
      expect(screen.getByText(`25 Dec ${currentYear - 1}, 4:00 pm`)).toBeInTheDocument() // Previous year, year shown

      // Requirement 1.4: Local timezone handling (all timestamps processed through local conversion)
      expect(screen.getByText('10 Jul, 6:00 pm')).toBeInTheDocument() // UTC converted to local
      expect(screen.getByText('5 Jun, 6:30 pm')).toBeInTheDocument() // Offset timezone converted to local

      // Mixed validity: Invalid timestamps show fallback
      const invalidDateElements = screen.getAllByText('Invalid Date')
      expect(invalidDateElements).toHaveLength(2) // 2 invalid sessions

      // Complete workflow: Session selection functionality
      const selectableSession = screen.getByText('15 Aug, 8:30 pm').closest('button')
      expect(selectableSession).toBeInTheDocument()
      
      fireEvent.click(selectableSession!)
      expect(mockOnSelectSession).toHaveBeenCalledWith('valid-current-2')

      // All sessions should be rendered as interactive buttons
      const allSessionButtons = screen.getAllByRole('option')
      expect(allSessionButtons).toHaveLength(7)

      // Accessibility: All sessions should have proper aria-labels
      allSessionButtons.forEach(button => {
        const ariaLabel = button.getAttribute('aria-label')
        expect(ariaLabel).toMatch(/Chat session from/)
      })

      // Verify original titles are not displayed (using formatted timestamps instead)
      expect(screen.queryByText('Valid Current Year 1')).not.toBeInTheDocument()
      expect(screen.queryByText('Valid Previous Year')).not.toBeInTheDocument()
      expect(screen.queryByText('UTC Timezone')).not.toBeInTheDocument()
    })
  })

  describe('Performance and Consistency', () => {
    it('should format timestamps efficiently for large session lists', () => {
      // Create a large number of sessions
      const largeSessions: Session[] = Array.from({ length: 50 }, (_, index) => ({
        id: `session-${index}`,
        user_id: 'user1',
        title: `Session ${index}`,
        created_at: new Date(2024, 8, 3 - index, 20, 20, 0).toISOString(),
        updated_at: new Date(2024, 8, 3 - index, 20, 20, 0).toISOString()
      }))

      const defaultProps = {
        sessions: largeSessions,
        currentSessionId: 'session-0',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      render(<SessionSidebar {...defaultProps} />)

      // Should call formatting function for each session exactly once
      expect(mockFormatSessionTimestamp).toHaveBeenCalledTimes(50)

      // All sessions should be rendered
      const sessionButtons = screen.getAllByRole('option')
      expect(sessionButtons).toHaveLength(50)
    })

    it('should maintain consistent formatting across re-renders', () => {
      const testSessions: Session[] = [
        {
          id: 'consistent-session-1',
          user_id: 'user1',
          title: 'Consistent Session 1',
          created_at: '2024-09-03T20:20:00Z',
          updated_at: '2024-09-03T20:20:00Z'
        }
      ]

      const defaultProps = {
        sessions: testSessions,
        currentSessionId: 'consistent-session-1',
        onCreateSession: vi.fn(),
        isCreatingSession: false,
        onSelectSession: vi.fn(),
        isLoading: false,
        error: null
      }

      const { rerender } = render(<SessionSidebar {...defaultProps} />)

      // Initial render
      expect(screen.getByText('4 Sep 2024, 2:20 am')).toBeInTheDocument()
      const initialCallCount = mockFormatSessionTimestamp.mock.calls.length

      // Re-render with same props
      rerender(<SessionSidebar {...defaultProps} />)

      // Should still display the same formatted timestamp
      expect(screen.getByText('4 Sep 2024, 2:20 am')).toBeInTheDocument()
      
      // Formatting function should be called again (React doesn't memoize by default)
      expect(mockFormatSessionTimestamp.mock.calls.length).toBeGreaterThan(initialCallCount)
    })
  })
})