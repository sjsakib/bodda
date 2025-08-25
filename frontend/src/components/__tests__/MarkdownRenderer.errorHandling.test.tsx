import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SafeMarkdownRenderer, MarkdownRenderer } from '../MarkdownRenderer';
import ReactMarkdown from 'react-markdown';

// Mock ReactMarkdown to simulate errors
vi.mock('react-markdown', () => ({
  default: vi.fn()
}));

// Mock remark-gfm
vi.mock('remark-gfm', () => ({
  default: vi.fn()
}));

const MockedReactMarkdown = vi.mocked(ReactMarkdown);

describe('MarkdownRenderer Error Handling', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
  
  beforeEach(() => {
    // Spy on console.error to verify error logging
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    // Restore console.error after each test
    consoleErrorSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('SafeMarkdownRenderer', () => {
    it('should render markdown content normally when no errors occur', () => {
      MockedReactMarkdown.mockImplementation(({ children }) => <div data-testid="markdown-content">{children}</div>);

      const content = '# Test Heading\n\nThis is a test.';
      
      render(<SafeMarkdownRenderer content={content} />);
      
      expect(screen.getByTestId('markdown-content')).toBeInTheDocument();
      expect(screen.getByTestId('markdown-content')).toHaveTextContent('# Test Heading This is a test.');
      expect(consoleErrorSpy).not.toHaveBeenCalled();
    });

    it('should fall back to plain text when markdown rendering throws an error', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw new Error('Markdown parsing failed');
      });

      const content = '# Malformed markdown **unclosed bold';
      
      render(<SafeMarkdownRenderer content={content} />);
      
      // Should render fallback content
      expect(screen.getByText(content)).toBeInTheDocument();
      expect(screen.getByText(content).tagName).toBe('PRE');
      
      // Should log the error (both from error boundary and fallback usage)
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        'Markdown rendering failed:',
        expect.objectContaining({
          error: 'Markdown parsing failed',
          timestamp: expect.any(String)
        })
      );
    });

    it('should handle very long content in error logging', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw new Error('Parsing error');
      });

      const longContent = 'a'.repeat(200); // 200 character string
      
      render(<SafeMarkdownRenderer content={longContent} />);
      
      // Should log truncated content (first 100 chars + ...)
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        'Markdown rendering failed, using fallback renderer:',
        expect.objectContaining({
          contentPreview: 'a'.repeat(100) + '...'
        })
      );
    });

    it('should handle unknown error types gracefully', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw 'String error'; // Non-Error object
      });

      const content = 'Test content';
      
      render(<SafeMarkdownRenderer content={content} />);
      
      // Should render fallback and log error
      expect(screen.getByText(content)).toBeInTheDocument();
      expect(consoleErrorSpy).toHaveBeenCalled();
    });

    it('should apply custom className to fallback renderer', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw new Error('Test error');
      });

      const content = 'Test content';
      const customClass = 'custom-class';
      
      render(<SafeMarkdownRenderer content={content} className={customClass} />);
      
      const fallbackElement = screen.getByText(content).closest('.fallback-content');
      expect(fallbackElement).toHaveClass(customClass);
    });

    it('should preserve whitespace and formatting in fallback renderer', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw new Error('Test error');
      });

      const contentWithWhitespace = 'Line 1\n\nLine 3\n  Indented line';
      
      const { container } = render(<SafeMarkdownRenderer content={contentWithWhitespace} />);
      
      const preElement = container.querySelector('pre');
      expect(preElement).toHaveClass('whitespace-pre-wrap');
      expect(preElement?.textContent).toBe(contentWithWhitespace);
    });
  });

  describe('Malformed Markdown Scenarios', () => {
    beforeEach(() => {
      // Reset ReactMarkdown mock to normal behavior for these tests
      MockedReactMarkdown.mockImplementation(({ children }) => <div data-testid="markdown-content">{children}</div>);
    });

    it('should handle incomplete markdown syntax gracefully', () => {
      const malformedContent = [
        '# Incomplete header',
        '**unclosed bold text',
        '- List item without proper spacing',
        '| Table | without',
        '```',
        'unclosed code block'
      ].join('\n');

      // This should not throw an error with the real ReactMarkdown
      expect(() => {
        render(<SafeMarkdownRenderer content={malformedContent} />);
      }).not.toThrow();
    });

    it('should handle empty content', () => {
      expect(() => {
        render(<SafeMarkdownRenderer content="" />);
      }).not.toThrow();
    });

    it('should handle content with only whitespace', () => {
      expect(() => {
        render(<SafeMarkdownRenderer content="   \n\n   \t   " />);
      }).not.toThrow();
    });

    it('should handle content with special characters', () => {
      const specialContent = '# Test with émojis 🚀 and spëcial chars: <>&"\'';
      
      expect(() => {
        render(<SafeMarkdownRenderer content={specialContent} />);
      }).not.toThrow();
    });
  });

  describe('Error Logging Details', () => {
    it('should include error stack trace when available', () => {
      const testError = new Error('Test error with stack');
      MockedReactMarkdown.mockImplementation(() => {
        throw testError;
      });

      render(<SafeMarkdownRenderer content="test" />);
      
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        'Markdown rendering failed:',
        expect.objectContaining({
          stack: expect.stringContaining('Error: Test error with stack')
        })
      );
    });

    it('should include timestamp in error logs', () => {
      MockedReactMarkdown.mockImplementation(() => {
        throw new Error('Test error');
      });

      const beforeTime = new Date().toISOString();
      render(<SafeMarkdownRenderer content="test" />);
      const afterTime = new Date().toISOString();
      
      expect(consoleErrorSpy).toHaveBeenCalled();
      
      // Check that at least one of the console.error calls includes a timestamp
      const errorCalls = consoleErrorSpy.mock.calls;
      const hasTimestamp = errorCalls.some(call => 
        call[1] && typeof call[1] === 'object' && 'timestamp' in call[1]
      );
      expect(hasTimestamp).toBe(true);
    });
  });
});