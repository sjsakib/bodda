import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, test, expect } from 'vitest';
import { MarkdownRenderer } from '../MarkdownRenderer';

describe('MarkdownRenderer Responsive Design', () => {
  test('renders headings with responsive font sizes', () => {
    const content = '# Main Heading\n## Sub Heading\n### Small Heading';
    render(<MarkdownRenderer content={content} />);
    
    const h1 = screen.getByRole('heading', { level: 1 });
    const h2 = screen.getByRole('heading', { level: 2 });
    const h3 = screen.getByRole('heading', { level: 3 });
    
    // Check that responsive classes are applied
    expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');
    expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl');
    expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl');
  });

  test('renders tables with mobile-optimized styling', () => {
    const content = `
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |
| Long data that might overflow | More data | Even more data |
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const table = screen.getByRole('table');
    const tableWrapper = table.parentElement;
    
    // Check that table wrapper has overflow handling
    expect(tableWrapper).toHaveClass('overflow-x-auto');
    
    // Check that table cells have responsive padding
    const cells = screen.getAllByRole('cell');
    cells.forEach(cell => {
      expect(cell).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3');
      expect(cell).toHaveClass('break-words'); // For long content
    });
  });

  test('renders lists with responsive spacing and indentation', () => {
    const content = `
- First item
- Second item with longer text that might wrap
- Third item

1. Numbered first
2. Numbered second
3. Numbered third
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const lists = screen.getAllByRole('list');
    const unorderedList = lists[0]; // First list is the ul
    const orderedList = lists[1]; // Second list is the ol
    
    expect(unorderedList).toHaveClass('pl-4', 'sm:pl-6');
    expect(unorderedList).toHaveClass('mb-3', 'sm:mb-4');
    expect(orderedList).toHaveClass('pl-4', 'sm:pl-6');
    expect(orderedList).toHaveClass('mb-3', 'sm:mb-4');
  });

  test('renders code blocks with responsive padding and font sizes', () => {
    const content = `
Here is some \`inline code\` in a sentence.

\`\`\`javascript
function example() {
  return "This is a code block";
}
\`\`\`
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    // Check inline code responsive styling
    const inlineCode = screen.getByText('inline code');
    expect(inlineCode).toHaveClass('text-xs', 'sm:text-sm');
    expect(inlineCode).toHaveClass('px-1', 'sm:px-1.5');
    
    // Check code block responsive styling
    const codeBlock = screen.getByText(/function example/);
    const preElement = codeBlock.closest('pre');
    expect(preElement).toHaveClass('p-2', 'sm:p-4');
    expect(preElement).toHaveClass('text-xs', 'sm:text-sm');
  });

  test('renders blockquotes with responsive spacing', () => {
    const content = '> This is an important quote that should be highlighted';
    
    render(<MarkdownRenderer content={content} />);
    
    const blockquote = screen.getByText(/important quote/);
    const blockquoteElement = blockquote.closest('blockquote');
    
    expect(blockquoteElement).toHaveClass('pl-3', 'sm:pl-4');
    expect(blockquoteElement).toHaveClass('mb-3', 'sm:mb-4');
    expect(blockquoteElement).toHaveClass('text-sm', 'sm:text-base');
  });

  test('renders links with word breaking for long URLs', () => {
    const content = '[Very long link text that might overflow on mobile](https://example.com/very/long/url/that/might/cause/issues)';
    
    render(<MarkdownRenderer content={content} />);
    
    const link = screen.getByRole('link');
    expect(link).toHaveClass('break-words');
  });

  test('applies responsive base font size to markdown container', () => {
    const content = 'Simple paragraph text';
    
    render(<MarkdownRenderer content={content} />);
    
    const container = screen.getByText('Simple paragraph text').closest('.markdown-content');
    expect(container).toHaveClass('text-sm', 'sm:text-base');
  });

  test('renders paragraphs with responsive spacing', () => {
    const content = `
First paragraph with some content.

Second paragraph with more content that might be longer and wrap on mobile devices.

Third paragraph.
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const paragraphs = screen.getAllByText(/paragraph/);
    paragraphs.forEach(p => {
      const paragraphElement = p.closest('p');
      expect(paragraphElement).toHaveClass('mb-3', 'sm:mb-4');
      expect(paragraphElement).toHaveClass('text-sm', 'sm:text-base');
    });
  });

  test('handles horizontal rules with responsive spacing', () => {
    const content = `
Content above

---

Content below
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    // Find the hr element by looking for the separator
    const container = screen.getByText('Content above').closest('.markdown-content');
    const hrElement = container?.querySelector('hr');
    
    expect(hrElement).toHaveClass('my-4', 'sm:my-6');
  });
});