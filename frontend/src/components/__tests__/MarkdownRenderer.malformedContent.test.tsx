import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SafeMarkdownRenderer } from '../MarkdownRenderer';

describe('MarkdownRenderer Malformed Content Integration', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
  
  beforeEach(() => {
    // Spy on console.error to verify error logging
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    // Restore console.error after each test
    consoleErrorSpy.mockRestore();
  });

  describe('Real-world malformed markdown scenarios', () => {
    it('should handle incomplete markdown tables gracefully', () => {
      const malformedTable = `
# Training Schedule

| Day | Exercise | Duration
|-----|----------|
| Monday | Running | 30 min
| Tuesday | Swimming
| Wednesday | | 45 min
`;

      expect(() => {
        render(<SafeMarkdownRenderer content={malformedTable} />);
      }).not.toThrow();

      // Should render something (either properly formatted or fallback)
      expect(screen.getByText(/Training Schedule/)).toBeInTheDocument();
    });

    it('should handle unclosed markdown formatting', () => {
      const unclosedFormatting = `
# Workout Plan

**This is bold text that is never closed

*This italic text is also unclosed

\`\`\`
This code block has no closing
And continues forever...

## Another heading in the middle of code
`;

      expect(() => {
        render(<SafeMarkdownRenderer content={unclosedFormatting} />);
      }).not.toThrow();

      // Should render something
      expect(screen.getByText(/Workout Plan/)).toBeInTheDocument();
    });

    it('should handle mixed valid and invalid markdown', () => {
      const mixedContent = `
# Valid Heading

This is **properly formatted** text.

- Valid list item
- Another valid item

| Valid | Table |
|-------|-------|
| Cell  | Cell  |

**Unclosed bold text
*Unclosed italic text

\`\`\`
Unclosed code block

## Heading inside code block??

- List inside code block?
  - Nested item
`;

      expect(() => {
        render(<SafeMarkdownRenderer content={mixedContent} />);
      }).not.toThrow();

      // Should render the valid parts
      expect(screen.getByText(/Valid Heading/)).toBeInTheDocument();
    });

    it('should handle empty and whitespace-only content', () => {
      const emptyContent = '';
      const whitespaceContent = '   \n\n\t   \n   ';

      expect(() => {
        render(<SafeMarkdownRenderer content={emptyContent} />);
      }).not.toThrow();

      expect(() => {
        render(<SafeMarkdownRenderer content={whitespaceContent} />);
      }).not.toThrow();
    });

    it('should handle content with special characters and unicode', () => {
      const specialContent = `
# Émojis and Special Characters 🚀

This content has **special chars**: <>&"'

Unicode characters: ñáéíóú, 中文, العربية, русский

Mathematical symbols: ∑∏∫∆∇

\`Code with special chars: <script>alert('xss')</script>\`

> Blockquote with émojis 🎯 and symbols ∞
`;

      expect(() => {
        render(<SafeMarkdownRenderer content={specialContent} />);
      }).not.toThrow();

      // Should render the content
      expect(screen.getByText(/Émojis and Special Characters/)).toBeInTheDocument();
    });
  });
});