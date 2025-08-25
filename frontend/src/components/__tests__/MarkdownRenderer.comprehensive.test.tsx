import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, test, expect } from 'vitest';
import { MarkdownRenderer, SafeMarkdownRenderer } from '../MarkdownRenderer';

describe('MarkdownRenderer - Comprehensive Coverage', () => {
  describe('All Custom Component Renderers', () => {
    test('renders all heading levels (h1-h6) with proper hierarchy', () => {
      const content = `
# Heading 1
## Heading 2  
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Test all heading levels exist
      expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Heading 1');
      expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Heading 2');
      expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Heading 3');
      expect(screen.getByRole('heading', { level: 4 })).toHaveTextContent('Heading 4');
      expect(screen.getByRole('heading', { level: 5 })).toHaveTextContent('Heading 5');
      expect(screen.getByRole('heading', { level: 6 })).toHaveTextContent('Heading 6');
      
      // Test visual hierarchy through font sizes
      const h1 = screen.getByRole('heading', { level: 1 });
      const h2 = screen.getByRole('heading', { level: 2 });
      const h3 = screen.getByRole('heading', { level: 3 });
      
      // Verify responsive typography classes
      expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl', 'font-bold');
      expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl', 'font-semibold');
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl', 'font-medium');
    });

    test('renders paragraph elements with proper styling', () => {
      const content = `
First paragraph with some content.

Second paragraph with more content that should have proper spacing.

Final paragraph.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const paragraphs = screen.getAllByText(/paragraph/);
      expect(paragraphs).toHaveLength(3);
      
      paragraphs.forEach(p => {
        const paragraphElement = p.closest('p');
        expect(paragraphElement).toHaveClass(
          'mb-3', 'sm:mb-4', 'text-gray-700', 'leading-relaxed', 'last:mb-0', 'text-sm', 'sm:text-base'
        );
      });
    });

    test('renders all text formatting elements correctly', () => {
      const content = `
**Bold text** and *italic text* and \`inline code\` formatting.

***Bold and italic combined*** text.

~~Strikethrough text~~ (GFM feature).
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Test bold
      const boldElement = screen.getByText('Bold text');
      expect(boldElement.tagName).toBe('STRONG');
      expect(boldElement).toHaveClass('font-semibold', 'text-gray-900');
      
      // Test italic
      const italicElement = screen.getByText('italic text');
      expect(italicElement.tagName).toBe('EM');
      expect(italicElement).toHaveClass('italic', 'text-gray-800');
      
      // Test inline code
      const codeElement = screen.getByText('inline code');
      expect(codeElement.tagName).toBe('CODE');
      expect(codeElement).toHaveClass('bg-gray-100', 'text-gray-800', 'font-mono');
      
      // Test strikethrough (GFM feature)
      expect(screen.getByText('Strikethrough text')).toBeInTheDocument();
    });

    test('renders code blocks with language syntax highlighting support', () => {
      const content = `
\`\`\`javascript
function example() {
  console.log("Hello World");
  return true;
}
\`\`\`

\`\`\`python
def hello():
    print("Hello from Python")
    return "success"
\`\`\`

\`\`\`bash
echo "Shell command"
ls -la
\`\`\`
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Find all code blocks
      const jsCode = screen.getByText(/function example/);
      const pythonCode = screen.getByText(/def hello/);
      const bashCode = screen.getByText(/echo "Shell command"/);
      
      // Verify they're all in pre elements with proper styling
      [jsCode, pythonCode, bashCode].forEach(code => {
        const preElement = code.closest('pre');
        expect(preElement).toHaveClass(
          'bg-gray-50', 'border', 'border-gray-200', 'rounded-lg', 
          'p-2', 'sm:p-4', 'mb-3', 'sm:mb-4', 'overflow-x-auto', 
          'text-xs', 'sm:text-sm', 'font-mono', 'leading-relaxed'
        );
      });
    });

    test('renders complex nested lists with proper indentation', () => {
      const content = `
1. First level ordered
   - Nested unordered item
   - Another nested item
     1. Deep nested ordered
     2. Another deep item
        - Very deep unordered
2. Second first level
   - Mixed nesting works
     - Even deeper
       1. Numbers in deep nesting
       2. More numbers
3. Third first level
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const lists = screen.getAllByRole('list');
      expect(lists.length).toBeGreaterThan(1); // Multiple nested lists
      
      // Verify content structure
      expect(screen.getByText('First level ordered')).toBeInTheDocument();
      expect(screen.getByText('Nested unordered item')).toBeInTheDocument();
      expect(screen.getByText('Deep nested ordered')).toBeInTheDocument();
      expect(screen.getByText('Very deep unordered')).toBeInTheDocument();
      expect(screen.getByText('Numbers in deep nesting')).toBeInTheDocument();
    });

    test('renders complex tables with all table elements', () => {
      const content = `
| Header 1 | Header 2 | Header 3 | Header 4 |
|----------|----------|----------|----------|
| Row 1 Col 1 | Row 1 Col 2 | Row 1 Col 3 | Row 1 Col 4 |
| Row 2 Col 1 | Row 2 Col 2 | Row 2 Col 3 | Row 2 Col 4 |
| Row 3 Col 1 | Row 3 Col 2 | Row 3 Col 3 | Row 3 Col 4 |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      
      // Test table structure
      expect(table).toHaveClass('min-w-full', 'divide-y', 'divide-gray-200');
      
      // Test wrapper
      const wrapper = table.closest('div');
      expect(wrapper).toHaveClass('overflow-x-auto', 'mb-3', 'sm:mb-4', 'rounded-lg', 'border', 'border-gray-200', 'shadow-sm');
      
      // Test thead
      const thead = table.querySelector('thead');
      expect(thead).toHaveClass('bg-gray-50');
      
      // Test tbody
      const tbody = table.querySelector('tbody');
      expect(tbody).toHaveClass('bg-white', 'divide-y', 'divide-gray-200');
      
      // Test headers
      const headers = screen.getAllByRole('columnheader');
      expect(headers).toHaveLength(4);
      headers.forEach(header => {
        expect(header).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3', 'text-left', 'text-xs', 'font-medium', 'text-gray-500', 'uppercase', 'tracking-wider', 'border-b', 'border-gray-200');
      });
      
      // Test data cells
      const cells = screen.getAllByRole('cell');
      expect(cells).toHaveLength(12); // 3 rows × 4 columns
      cells.forEach(cell => {
        expect(cell).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3', 'text-xs', 'sm:text-sm', 'text-gray-700', 'break-words');
      });
      
      // Test row hover effects
      const rows = tbody?.querySelectorAll('tr');
      rows?.forEach(row => {
        expect(row).toHaveClass('hover:bg-gray-50', 'transition-colors', 'duration-150');
      });
    });

    test('renders all special elements with proper styling', () => {
      const content = `
> This is a blockquote with **bold** and *italic* text.
> 
> Multiple lines in blockquote.

[External link](https://example.com) and [another link](https://test.com).

---

More content after horizontal rule.

> Another blockquote after the rule.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Test blockquotes
      const blockquotes = screen.getAllByText(/blockquote/).map(el => el.closest('blockquote'));
      expect(blockquotes.length).toBeGreaterThan(0);
      blockquotes.forEach(blockquote => {
        if (blockquote) {
          expect(blockquote).toHaveClass(
            'border-l-4', 'border-blue-200', 'pl-3', 'sm:pl-4', 'py-2', 
            'mb-3', 'sm:mb-4', 'bg-blue-50', 'text-gray-700', 'italic', 
            'rounded-r-md', 'text-sm', 'sm:text-base'
          );
        }
      });
      
      // Test links
      const links = screen.getAllByRole('link');
      expect(links).toHaveLength(2);
      links.forEach(link => {
        expect(link).toHaveAttribute('target', '_blank');
        expect(link).toHaveAttribute('rel', 'noopener noreferrer');
        expect(link).toHaveClass(
          'text-blue-600', 'hover:text-blue-800', 'underline', 
          'decoration-blue-300', 'hover:decoration-blue-500', 
          'transition-colors', 'duration-150', 'break-words'
        );
      });
      
      // Test horizontal rule
      const { container } = render(<MarkdownRenderer content={content} />);
      const hr = container.querySelector('hr');
      expect(hr).toHaveClass('my-4', 'sm:my-6', 'border-t', 'border-gray-200');
    });
  });

  describe('GitHub Flavored Markdown Features', () => {
    test('renders task lists (checkboxes)', () => {
      const content = `
## Todo List

- [x] Completed task
- [ ] Incomplete task
- [x] Another completed task
- [ ] Another incomplete task
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Should render the content (GFM plugin handles task lists)
      expect(screen.getByText('Todo List')).toBeInTheDocument();
      expect(screen.getByText(/Completed task/)).toBeInTheDocument();
      expect(screen.getByText(/Incomplete task/)).toBeInTheDocument();
    });

    test('renders strikethrough text', () => {
      const content = 'This is ~~strikethrough~~ text.';
      
      render(<MarkdownRenderer content={content} />);
      
      expect(screen.getByText('strikethrough')).toBeInTheDocument();
    });

    test('renders tables with alignment', () => {
      const content = `
| Left | Center | Right |
|:-----|:------:|------:|
| L1   | C1     | R1    |
| L2   | C2     | R2    |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();
      
      // Verify content is rendered
      expect(screen.getByText('Left')).toBeInTheDocument();
      expect(screen.getByText('Center')).toBeInTheDocument();
      expect(screen.getByText('Right')).toBeInTheDocument();
    });
  });

  describe('Responsive Design Verification', () => {
    test('applies responsive classes to all elements', () => {
      const content = `
# Responsive Heading

This is a paragraph with responsive text sizing.

- List item with responsive spacing
- Another list item

\`Inline code\` with responsive sizing.

\`\`\`
Code block with responsive padding
\`\`\`

| Table | Header |
|-------|--------|
| Cell  | Data   |

> Blockquote with responsive spacing

---
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Test heading responsiveness
      const heading = screen.getByRole('heading', { level: 1 });
      expect(heading).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');
      
      // Test paragraph responsiveness
      const paragraph = screen.getByText(/paragraph with responsive/);
      expect(paragraph.closest('p')).toHaveClass('text-sm', 'sm:text-base');
      
      // Test list responsiveness
      const list = screen.getByRole('list');
      expect(list).toHaveClass('pl-4', 'sm:pl-6', 'mb-3', 'sm:mb-4');
      
      // Test inline code responsiveness
      const inlineCode = screen.getByText('Inline code');
      expect(inlineCode).toHaveClass('text-xs', 'sm:text-sm', 'px-1', 'sm:px-1.5');
      
      // Test code block responsiveness
      const codeBlock = screen.getByText(/Code block with responsive/);
      expect(codeBlock.closest('pre')).toHaveClass('p-2', 'sm:p-4', 'text-xs', 'sm:text-sm');
      
      // Test table responsiveness
      const table = screen.getByRole('table');
      const tableWrapper = table.closest('div');
      expect(tableWrapper).toHaveClass('mb-3', 'sm:mb-4');
      
      const tableHeaders = screen.getAllByRole('columnheader');
      tableHeaders.forEach(header => {
        expect(header).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3');
      });
      
      // Test blockquote responsiveness
      const blockquote = screen.getByText(/Blockquote with responsive/);
      expect(blockquote.closest('blockquote')).toHaveClass('pl-3', 'sm:pl-4', 'mb-3', 'sm:mb-4', 'text-sm', 'sm:text-base');
      
      // Test HR responsiveness
      const { container } = render(<MarkdownRenderer content={content} />);
      const hr = container.querySelector('hr');
      expect(hr).toHaveClass('my-4', 'sm:my-6');
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    test('handles empty markdown content', () => {
      expect(() => {
        render(<MarkdownRenderer content="" />);
      }).not.toThrow();
      
      const { container } = render(<MarkdownRenderer content="" />);
      expect(container.firstChild).toHaveClass('markdown-content');
    });

    test('handles whitespace-only content', () => {
      const whitespaceContent = '   \n\n\t   \n   ';
      
      expect(() => {
        render(<MarkdownRenderer content={whitespaceContent} />);
      }).not.toThrow();
    });

    test('handles very long content without breaking', () => {
      const longContent = `
# Very Long Content Test

${'This is a very long paragraph that repeats many times. '.repeat(100)}

${'- List item that is very long and repeats many times\n'.repeat(50)}

| Very Long Header That Might Cause Issues | Another Long Header | Third Long Header |
|-------------------------------------------|---------------------|-------------------|
${'| Very long cell content that might overflow and cause layout issues | More long content | Even more content |\n'.repeat(20)}
      `;
      
      expect(() => {
        render(<MarkdownRenderer content={longContent} />);
      }).not.toThrow();
      
      // Should still render properly
      expect(screen.getByText('Very Long Content Test')).toBeInTheDocument();
    });

    test('handles special characters and unicode', () => {
      const specialContent = `
# Special Characters: <>&"'

Unicode: ñáéíóú, 中文, العربية, русский, 🚀🎯

Mathematical: ∑∏∫∆∇

\`Code with special chars: <script>alert('test')</script>\`

> Quote with émojis 🎯 and symbols ∞
      `;
      
      expect(() => {
        render(<MarkdownRenderer content={specialContent} />);
      }).not.toThrow();
      
      expect(screen.getByText(/Special Characters/)).toBeInTheDocument();
      expect(screen.getByText(/Unicode:/)).toBeInTheDocument();
    });

    test('handles malformed markdown gracefully', () => {
      const malformedContent = `
# Valid heading

**Unclosed bold text

*Unclosed italic text

\`\`\`
Unclosed code block
continues here...

| Malformed | Table
|-----------|
| Missing | Cell |
| Too | Many | Cells |

> Unclosed blockquote
continues here
      `;
      
      expect(() => {
        render(<SafeMarkdownRenderer content={malformedContent} />);
      }).not.toThrow();
      
      // Should render something
      expect(screen.getByText(/Valid heading/)).toBeInTheDocument();
    });

    test('handles nested formatting edge cases', () => {
      const nestedContent = `
**Bold with *italic inside* and \`code inside\` bold**

*Italic with **bold inside** and \`code inside\` italic*

\`Code with **bold** and *italic* inside\`

> Blockquote with **bold** and *italic* and \`code\` and [link](https://example.com)
      `;
      
      expect(() => {
        render(<MarkdownRenderer content={nestedContent} />);
      }).not.toThrow();
      
      // Should handle nested formatting
      expect(screen.getByText(/Bold with/)).toBeInTheDocument();
      expect(screen.getByText(/Italic with/)).toBeInTheDocument();
    });
  });

  describe('Accessibility and Semantic HTML', () => {
    test('generates proper semantic HTML structure', () => {
      const content = `
# Main Heading
## Section Heading
### Subsection Heading

Regular paragraph text.

- Unordered list
- Second item

1. Ordered list
2. Second item

| Table | Header |
|-------|--------|
| Data  | Cell   |

> Important blockquote

[Accessible link](https://example.com)
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Test semantic heading structure
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument();
      expect(screen.getByRole('heading', { level: 3 })).toBeInTheDocument();
      
      // Test list semantics
      const lists = screen.getAllByRole('list');
      expect(lists).toHaveLength(2); // One ul, one ol
      
      const listItems = screen.getAllByRole('listitem');
      expect(listItems.length).toBeGreaterThanOrEqual(4);
      
      // Test table semantics
      expect(screen.getByRole('table')).toBeInTheDocument();
      expect(screen.getAllByRole('columnheader')).toHaveLength(2);
      expect(screen.getAllByRole('cell')).toHaveLength(2);
      
      // Test link accessibility
      const link = screen.getByRole('link', { name: 'Accessible link' });
      expect(link).toHaveAttribute('href', 'https://example.com');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    test('maintains proper heading hierarchy', () => {
      const content = `
# H1 Title
## H2 Section
### H3 Subsection
#### H4 Sub-subsection
##### H5 Deep section
###### H6 Deepest section
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Verify all heading levels are present and properly structured
      for (let level = 1; level <= 6; level++) {
        const heading = screen.getByRole('heading', { level });
        expect(heading).toBeInTheDocument();
        expect(heading.tagName).toBe(`H${level}`);
      }
    });
  });

  describe('Performance and Optimization', () => {
    test('handles large content efficiently', () => {
      const largeContent = `
# Performance Test

${Array.from({ length: 100 }, (_, i) => `
## Section ${i + 1}

This is paragraph ${i + 1} with some content.

- List item ${i + 1}.1
- List item ${i + 1}.2
- List item ${i + 1}.3

\`\`\`javascript
function test${i + 1}() {
  return "test ${i + 1}";
}
\`\`\`

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data ${i + 1}.1 | Data ${i + 1}.2 | Data ${i + 1}.3 |

---
`).join('')}
      `;
      
      const startTime = performance.now();
      
      expect(() => {
        render(<MarkdownRenderer content={largeContent} />);
      }).not.toThrow();
      
      const endTime = performance.now();
      const renderTime = endTime - startTime;
      
      // Should render within reasonable time (less than 1 second)
      expect(renderTime).toBeLessThan(1000);
      
      // Should still render content correctly
      expect(screen.getByText('Performance Test')).toBeInTheDocument();
      expect(screen.getByText('Section 1')).toBeInTheDocument();
      expect(screen.getByText('Section 100')).toBeInTheDocument();
    });
  });
});