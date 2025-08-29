import React from 'react';
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import App from '../../App';
import { apiClient } from '../../services/api';

// Mock the API client
vi.mock('../../services/api', () => ({
  apiClient: {
    checkAuth: vi.fn(),
    getSessions: vi.fn(),
    getMessages: vi.fn(),
    createSession: vi.fn(),
    createEventSource: vi.fn(),
    logout: vi.fn(),
  },
}));

// Mock diagram libraries
vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn().mockResolvedValue({
      svg: '<svg role="img" aria-labelledby="diagram-title" aria-describedby="diagram-desc"><title id="diagram-title">Training Flow</title><desc id="diagram-desc">A flowchart showing training workflow</desc><g><text>Start → Warm Up → Exercise → Cool Down</text></g></svg>'
    }),
  },
}));

vi.mock('react-vega', () => ({
  VegaLite: ({ spec, onError }: any) => (
    <div 
      data-testid="vega-lite-chart" 
      role="img" 
      aria-label={`${spec?.mark} chart`}
    >
      <svg>
        <title>Progress Chart</title>
        <desc>Bar chart showing weekly progress</desc>
        <g>
          <text>Weekly Progress: Mon(5), Tue(7), Wed(6)</text>
        </g>
      </svg>
    </div>
  ),
}));

const mockSession = {
  id: 'test-session-1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('End-to-End Diagram Workflows', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock successful auth
    (apiClient.checkAuth as any).mockResolvedValue({ authenticated: true });
    
    // Mock sessions
    (apiClient.getSessions as any).mockResolvedValue([mockSession]);
    
    // Mock empty messages initially
    (apiClient.getMessages as any).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('completes full user workflow: ask question → receive diagram response → interact with diagram', async () => {
    // Mock EventSource for streaming response
    const mockEventSource = {
      onmessage: null as any,
      onerror: null as any,
      close: vi.fn(),
    };
    
    (apiClient.createEventSource as any).mockReturnValue(mockEventSource);

    // Render the full app
    render(<App />);

    // Wait for chat interface to load (should redirect from landing page)
    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument();
    });

    // User asks for a training plan
    const input = screen.getByPlaceholderText(/Ask your AI coach/);
    const sendButton = screen.getByText('Send');

    fireEvent.change(input, { target: { value: 'Create a training plan with a flowchart' } });
    fireEvent.click(sendButton);

    // Verify user message appears
    await waitFor(() => {
      expect(screen.getByText('Create a training plan with a flowchart')).toBeInTheDocument();
    });

    // Simulate streaming AI response with diagram
    if (mockEventSource.onmessage) {
      // Start streaming response
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'chunk',
          content: 'Here\'s your personalized training plan:\n\n'
        })
      });

      // Add diagram content
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'chunk',
          content: '```mermaid\ngraph TD\n    A[Start Workout] --> B[Warm Up]\n    B --> C[Main Exercise]\n    C --> D[Cool Down]\n    D --> E[End]\n```\n\n'
        })
      });

      // Add explanation
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'chunk',
          content: 'This flowchart shows your optimal workout sequence. Each step is important for injury prevention and performance.'
        })
      });

      // Complete the response
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'complete',
          message: {
            id: 'response-1',
            content: 'Here\'s your personalized training plan:\n\n```mermaid\ngraph TD\n    A[Start Workout] --> B[Warm Up]\n    B --> C[Main Exercise]\n    C --> D[Cool Down]\n    D --> E[End]\n```\n\nThis flowchart shows your optimal workout sequence. Each step is important for injury prevention and performance.',
            role: 'assistant',
            created_at: new Date().toISOString(),
            session_id: 'test-session-1',
          }
        })
      });
    }

    // Wait for AI response to appear
    await waitFor(() => {
      expect(screen.getByText(/Here's your personalized training plan/)).toBeInTheDocument();
    });

    // Wait for diagram loading indicator
    await waitFor(() => {
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });

    // Wait for diagram to render
    await waitFor(() => {
      expect(screen.getByText(/Start → Warm Up → Exercise → Cool Down/)).toBeInTheDocument();
    }, { timeout: 5000 });

    // Verify diagram has proper accessibility attributes
    const diagramSvg = screen.getByRole('img', { name: /Training Flow/ });
    expect(diagramSvg).toBeInTheDocument();
    expect(diagramSvg).toHaveAttribute('aria-describedby');

    // Verify diagram container has proper CSS classes
    const diagramContainer = document.querySelector('.diagram-code-block.mermaid-code-block');
    expect(diagramContainer).toBeInTheDocument();
    expect(diagramContainer).toHaveClass('rounded-lg', 'border', 'border-gray-200', 'shadow-sm');
  });

  it('handles multiple diagram types in a single response', async () => {
    const mockEventSource = {
      onmessage: null as any,
      onerror: null as any,
      close: vi.fn(),
    };
    
    (apiClient.createEventSource as any).mockReturnValue(mockEventSource);

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument();
    });

    // User asks for analysis with charts
    const input = screen.getByPlaceholderText(/Ask your AI coach/);
    const sendButton = screen.getByText('Send');

    fireEvent.change(input, { target: { value: 'Show me my progress with charts and workflow' } });
    fireEvent.click(sendButton);

    // Simulate response with both Mermaid and Vega-Lite
    if (mockEventSource.onmessage) {
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'complete',
          message: {
            id: 'response-2',
            content: `Here's your progress analysis:

## Workflow
\`\`\`mermaid
graph LR
    A[Week 1] --> B[Week 2]
    B --> C[Week 3]
    C --> D[Current]
\`\`\`

## Progress Chart
\`\`\`vega-lite
{
  "mark": "bar",
  "data": {
    "values": [
      {"week": "Week 1", "distance": 15},
      {"week": "Week 2", "distance": 18},
      {"week": "Week 3", "distance": 22}
    ]
  },
  "encoding": {
    "x": {"field": "week", "type": "ordinal"},
    "y": {"field": "distance", "type": "quantitative"}
  }
}
\`\`\`

Great progress! You're consistently improving.`,
            role: 'assistant',
            created_at: new Date().toISOString(),
            session_id: 'test-session-1',
          }
        })
      });
    }

    // Wait for response
    await waitFor(() => {
      expect(screen.getByText(/Here's your progress analysis/)).toBeInTheDocument();
    });

    // Wait for both diagrams to load
    await waitFor(() => {
      expect(screen.getByText(/Loading diagram libraries/)).toBeInTheDocument();
    });

    // Wait for both diagrams to render
    await waitFor(() => {
      expect(screen.getByText(/Week 1.*Week 2.*Week 3.*Current/)).toBeInTheDocument();
      expect(screen.getByTestId('vega-lite-chart')).toBeInTheDocument();
    }, { timeout: 5000 });

    // Verify both diagram types are present
    expect(document.querySelector('.mermaid-code-block')).toBeInTheDocument();
    expect(document.querySelector('.vega-lite-code-block')).toBeInTheDocument();

    // Verify accessibility for both diagrams
    expect(screen.getByRole('img', { name: /bar chart/ })).toBeInTheDocument();
  });

  it('maintains performance with large diagram responses', async () => {
    const mockEventSource = {
      onmessage: null as any,
      onerror: null as any,
      close: vi.fn(),
    };
    
    (apiClient.createEventSource as any).mockReturnValue(mockEventSource);

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument();
    });

    // Create a large response with multiple diagrams
    const largeDiagramContent = `# Comprehensive Training Analysis

## Phase 1 Workflow
\`\`\`mermaid
graph TD
    A[Assessment] --> B[Goal Setting]
    B --> C[Plan Creation]
    C --> D[Execution]
    D --> E[Monitoring]
    E --> F[Adjustment]
\`\`\`

## Weekly Progress
\`\`\`vega-lite
{
  "mark": "line",
  "data": {
    "values": [
      {"week": 1, "distance": 10, "pace": 8.5},
      {"week": 2, "distance": 12, "pace": 8.2},
      {"week": 3, "distance": 15, "pace": 8.0},
      {"week": 4, "distance": 18, "pace": 7.8}
    ]
  },
  "encoding": {
    "x": {"field": "week", "type": "quantitative"},
    "y": {"field": "distance", "type": "quantitative"}
  }
}
\`\`\`

## Training Zones
\`\`\`mermaid
pie title Training Distribution
    "Easy" : 60
    "Moderate" : 25
    "Hard" : 15
\`\`\`

## Recovery Process
\`\`\`mermaid
sequenceDiagram
    participant A as Athlete
    participant C as Coach
    participant S as System
    
    A->>C: Report fatigue
    C->>S: Adjust plan
    S->>A: New recommendations
    A->>C: Confirm changes
\`\`\``;

    const input = screen.getByPlaceholderText(/Ask your AI coach/);
    const sendButton = screen.getByText('Send');

    fireEvent.change(input, { target: { value: 'Give me a comprehensive analysis' } });
    fireEvent.click(sendButton);

    // Measure performance
    const startTime = performance.now();

    if (mockEventSource.onmessage) {
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'complete',
          message: {
            id: 'response-3',
            content: largeDiagramContent,
            role: 'assistant',
            created_at: new Date().toISOString(),
            session_id: 'test-session-1',
          }
        })
      });
    }

    // Wait for response to render
    await waitFor(() => {
      expect(screen.getByText(/Comprehensive Training Analysis/)).toBeInTheDocument();
    });

    // Wait for all diagrams to load
    await waitFor(() => {
      const diagramContainers = document.querySelectorAll('.diagram-code-block');
      expect(diagramContainers.length).toBe(4); // 3 mermaid + 1 vega-lite
    }, { timeout: 10000 });

    const endTime = performance.now();
    const renderTime = endTime - startTime;

    // Performance should be reasonable (less than 5 seconds for complex content)
    expect(renderTime).toBeLessThan(5000);

    // Interface should remain responsive
    const newInput = screen.getByPlaceholderText(/Ask your AI coach/);
    expect(newInput).not.toBeDisabled();
    
    // Should be able to type immediately
    fireEvent.change(newInput, { target: { value: 'Follow up question' } });
    expect(newInput).toHaveValue('Follow up question');
  });

  it('handles diagram errors gracefully without breaking chat flow', async () => {
    // Mock mermaid to fail
    const mermaidMock = await import('mermaid');
    (mermaidMock.default.render as any).mockRejectedValue(new Error('Invalid syntax'));

    const mockEventSource = {
      onmessage: null as any,
      onerror: null as any,
      close: vi.fn(),
    };
    
    (apiClient.createEventSource as any).mockReturnValue(mockEventSource);

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText(/Ask your AI coach/);
    const sendButton = screen.getByText('Send');

    fireEvent.change(input, { target: { value: 'Show me a broken diagram' } });
    fireEvent.click(sendButton);

    if (mockEventSource.onmessage) {
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'complete',
          message: {
            id: 'response-4',
            content: 'Here is a diagram:\n\n```mermaid\ninvalid diagram syntax here\n```\n\nSorry about that!',
            role: 'assistant',
            created_at: new Date().toISOString(),
            session_id: 'test-session-1',
          }
        })
      });
    }

    // Wait for response
    await waitFor(() => {
      expect(screen.getByText(/Here is a diagram/)).toBeInTheDocument();
    });

    // Wait for error to be handled
    await waitFor(() => {
      expect(screen.getByText(/Invalid syntax/)).toBeInTheDocument();
    });

    // Chat should still be functional
    const newInput = screen.getByPlaceholderText(/Ask your AI coach/);
    expect(newInput).not.toBeDisabled();

    // Should be able to send another message
    fireEvent.change(newInput, { target: { value: 'That\'s okay, try again' } });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(screen.getByText('That\'s okay, try again')).toBeInTheDocument();
    });
  });

  it('supports keyboard navigation and accessibility features', async () => {
    const mockEventSource = {
      onmessage: null as any,
      onerror: null as any,
      close: vi.fn(),
    };
    
    (apiClient.createEventSource as any).mockReturnValue(mockEventSource);

    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('Bodda AI Coach')).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText(/Ask your AI coach/);
    const sendButton = screen.getByText('Send');

    fireEvent.change(input, { target: { value: 'Show accessible diagram' } });
    fireEvent.click(sendButton);

    if (mockEventSource.onmessage) {
      mockEventSource.onmessage({
        data: JSON.stringify({
          type: 'complete',
          message: {
            id: 'response-5',
            content: 'Accessible diagram:\n\n```mermaid\ngraph TD\n    A[Start] --> B[End]\n```',
            role: 'assistant',
            created_at: new Date().toISOString(),
            session_id: 'test-session-1',
          }
        })
      });
    }

    // Wait for diagram to render
    await waitFor(() => {
      expect(screen.getByRole('img')).toBeInTheDocument();
    }, { timeout: 5000 });

    // Test keyboard navigation
    const diagramElement = screen.getByRole('img');
    
    // Should be focusable
    diagramElement.focus();
    expect(document.activeElement).toBe(diagramElement);

    // Should have proper ARIA attributes
    expect(diagramElement).toHaveAttribute('aria-labelledby');
    expect(diagramElement).toHaveAttribute('aria-describedby');

    // Should have accessible title and description
    const title = document.getElementById(diagramElement.getAttribute('aria-labelledby')!);
    const description = document.getElementById(diagramElement.getAttribute('aria-describedby')!);
    
    expect(title).toBeInTheDocument();
    expect(description).toBeInTheDocument();
  });
});