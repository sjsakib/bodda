import { render, screen } from '@testing-library/react';
import { describe, test, expect } from 'vitest';
import MarkdownRenderer from '../MarkdownRenderer';

describe('MarkdownRenderer', () => {
  test('renders basic markdown content', () => {
    const content = '# Hello World\n\nThis is a **bold** text.';
    
    render(<MarkdownRenderer content={content} />);
    
    // Check if heading is rendered
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Hello World');
    
    // Check if bold text is rendered
    expect(screen.getByText('bold')).toBeInTheDocument();
  });

  test('applies custom className', () => {
    const content = 'Test content';
    const customClass = 'custom-class';
    
    const { container } = render(
      <MarkdownRenderer content={content} className={customClass} />
    );
    
    // Check if custom class is applied
    expect(container.firstChild).toHaveClass('markdown-content', customClass);
  });

  test('renders GitHub Flavored Markdown features', () => {
    const content = `
# Table Test

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |

~~strikethrough text~~
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    // Check if table is rendered
    expect(screen.getByRole('table')).toBeInTheDocument();
    
    // Check if strikethrough is rendered (GFM feature)
    expect(screen.getByText('strikethrough text')).toBeInTheDocument();
  });

  test('handles empty content gracefully', () => {
    const { container } = render(<MarkdownRenderer content="" />);
    
    // Should render without crashing and have the markdown-content class
    expect(container.firstChild).toHaveClass('markdown-content');
  });

  describe('Heading Rendering', () => {
    test('renders h1 with proper styling and hierarchy', () => {
      const content = '# Main Heading';
      
      render(<MarkdownRenderer content={content} />);
      
      const h1 = screen.getByRole('heading', { level: 1 });
      expect(h1).toHaveTextContent('Main Heading');
      expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl', 'font-bold', 'text-gray-900');
      expect(h1).toHaveClass('mb-3', 'sm:mb-4', 'mt-4', 'sm:mt-6', 'first:mt-0', 'leading-tight');
    });

    test('renders h2 with proper styling and hierarchy', () => {
      const content = '## Secondary Heading';
      
      render(<MarkdownRenderer content={content} />);
      
      const h2 = screen.getByRole('heading', { level: 2 });
      expect(h2).toHaveTextContent('Secondary Heading');
      expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl', 'font-semibold', 'text-gray-800');
      expect(h2).toHaveClass('mb-2', 'sm:mb-3', 'mt-3', 'sm:mt-5', 'first:mt-0', 'leading-tight');
    });

    test('renders h3 with proper styling and hierarchy', () => {
      const content = '### Tertiary Heading';
      
      render(<MarkdownRenderer content={content} />);
      
      const h3 = screen.getByRole('heading', { level: 3 });
      expect(h3).toHaveTextContent('Tertiary Heading');
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl', 'font-medium', 'text-gray-800');
      expect(h3).toHaveClass('mb-2', 'mt-3', 'sm:mt-4', 'first:mt-0', 'leading-tight');
    });

    test('renders multiple heading levels with proper visual hierarchy', () => {
      const content = `
# Main Title
## Section Title  
### Subsection Title
#### Fourth Level
##### Fifth Level
###### Sixth Level
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Check all heading levels are rendered
      expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Main Title');
      expect(screen.getByRole('heading', { level: 2 })).toHaveTextContent('Section Title');
      expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Subsection Title');
      expect(screen.getByRole('heading', { level: 4 })).toHaveTextContent('Fourth Level');
      expect(screen.getByRole('heading', { level: 5 })).toHaveTextContent('Fifth Level');
      expect(screen.getByRole('heading', { level: 6 })).toHaveTextContent('Sixth Level');
      
      // Verify visual hierarchy through font sizes
      const h1 = screen.getByRole('heading', { level: 1 });
      const h2 = screen.getByRole('heading', { level: 2 });
      const h3 = screen.getByRole('heading', { level: 3 });
      
      // H1 should be largest
      expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl', 'font-bold');
      // H2 should be smaller than H1
      expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl', 'font-semibold');
      // H3 should be smaller than H2
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl', 'font-medium');
    });

    test('renders nested headings with proper spacing', () => {
      const content = `
# Training Plan

## Week 1: Base Building

### Monday Workout

Some content here.

### Tuesday Recovery

More content.

## Week 2: Intensity

### Wednesday Intervals
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // All headings should render
      expect(screen.getByText('Training Plan')).toBeInTheDocument();
      expect(screen.getByText('Week 1: Base Building')).toBeInTheDocument();
      expect(screen.getByText('Monday Workout')).toBeInTheDocument();
      expect(screen.getByText('Tuesday Recovery')).toBeInTheDocument();
      expect(screen.getByText('Week 2: Intensity')).toBeInTheDocument();
      expect(screen.getByText('Wednesday Intervals')).toBeInTheDocument();
    });

    test('applies responsive typography classes', () => {
      const content = `
# Responsive H1
## Responsive H2  
### Responsive H3
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const h1 = screen.getByRole('heading', { level: 1 });
      const h2 = screen.getByRole('heading', { level: 2 });
      const h3 = screen.getByRole('heading', { level: 3 });
      
      // Check responsive classes are applied
      expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');
      expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl');
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl');
    });

    test('handles first heading without top margin', () => {
      const content = '# First Heading\n\nSome content after.';
      
      render(<MarkdownRenderer content={content} />);
      
      const h1 = screen.getByRole('heading', { level: 1 });
      // Should have first:mt-0 class to remove top margin when it's the first element
      expect(h1).toHaveClass('first:mt-0');
    });
  });

  describe('Text Formatting', () => {
    test('renders bold text with proper styling', () => {
      const content = 'This is **bold text** in a sentence.';
      
      render(<MarkdownRenderer content={content} />);
      
      const boldElement = screen.getByText('bold text');
      expect(boldElement.tagName).toBe('STRONG');
      expect(boldElement).toHaveClass('font-semibold', 'text-gray-900');
    });

    test('renders italic text with proper styling', () => {
      const content = 'This is *italic text* in a sentence.';
      
      render(<MarkdownRenderer content={content} />);
      
      const italicElement = screen.getByText('italic text');
      expect(italicElement.tagName).toBe('EM');
      expect(italicElement).toHaveClass('italic', 'text-gray-800');
    });

    test('renders inline code with proper styling', () => {
      const content = 'Use the `console.log()` function for debugging.';
      
      render(<MarkdownRenderer content={content} />);
      
      const codeElement = screen.getByText('console.log()');
      expect(codeElement.tagName).toBe('CODE');
      expect(codeElement).toHaveClass(
        'bg-gray-100', 
        'text-gray-800', 
        'px-1', 
        'sm:px-1.5', 
        'py-0.5', 
        'rounded', 
        'text-xs', 
        'sm:text-sm', 
        'font-mono',
        'border',
        'border-gray-200'
      );
    });

    test('renders code blocks with proper styling', () => {
      const content = '```javascript\nconst message = "Hello World";\nconsole.log(message);\n```';
      
      render(<MarkdownRenderer content={content} />);
      
      // Find the pre element by looking for text content with partial match
      const preElement = screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('const message = "Hello World"');
      }).closest('pre');
      
      expect(preElement).toHaveClass(
        'bg-gray-50',
        'border',
        'border-gray-200',
        'rounded-lg',
        'p-2',
        'sm:p-4',
        'mb-3',
        'sm:mb-4',
        'overflow-x-auto',
        'text-xs',
        'sm:text-sm',
        'font-mono',
        'leading-relaxed'
      );
    });

    test('renders combined text formatting correctly', () => {
      const content = 'This has **bold**, *italic*, and `code` formatting together.';
      
      render(<MarkdownRenderer content={content} />);
      
      // Check all formatting types are present
      const boldElement = screen.getByText('bold');
      const italicElement = screen.getByText('italic');
      const codeElement = screen.getByText('code');
      
      expect(boldElement.tagName).toBe('STRONG');
      expect(boldElement).toHaveClass('font-semibold', 'text-gray-900');
      
      expect(italicElement.tagName).toBe('EM');
      expect(italicElement).toHaveClass('italic', 'text-gray-800');
      
      expect(codeElement.tagName).toBe('CODE');
      expect(codeElement).toHaveClass('bg-gray-100', 'font-mono');
    });

    test('renders nested formatting correctly', () => {
      const content = '**This is bold with *italic inside* it.**';
      
      render(<MarkdownRenderer content={content} />);
      
      // The bold element should contain the italic text
      const boldElement = screen.getByText(/This is bold with/).closest('strong');
      const italicElement = screen.getByText('italic inside');
      
      expect(boldElement).toHaveClass('font-semibold', 'text-gray-900');
      expect(italicElement.tagName).toBe('EM');
      expect(italicElement).toHaveClass('italic', 'text-gray-800');
    });

    test('handles code blocks with different languages', () => {
      const content = `
\`\`\`python
def hello_world():
    print("Hello, World!")
\`\`\`

\`\`\`bash
echo "Hello from bash"
\`\`\`
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Both code blocks should be rendered with proper styling
      const pythonCode = screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('def hello_world():');
      });
      const bashCode = screen.getByText((content, element) => {
        return element?.tagName === 'CODE' && content.includes('echo "Hello from bash"');
      });
      
      expect(pythonCode.closest('pre')).toHaveClass('bg-gray-50', 'border', 'rounded-lg');
      expect(bashCode.closest('pre')).toHaveClass('bg-gray-50', 'border', 'rounded-lg');
    });

    test('distinguishes between inline code and code blocks', () => {
      const content = `
Inline code: \`const x = 1;\`

Code block:
\`\`\`javascript
const x = 1;
const y = 2;
\`\`\`
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Get both code elements
      const inlineCode = screen.getByText('const x = 1;');
      const blockCode = screen.getByText(/const x = 1;\s*const y = 2;/);
      
      // Inline code should have inline styling
      expect(inlineCode).toHaveClass('bg-gray-100', 'px-1', 'sm:px-1.5', 'py-0.5', 'rounded');
      
      // Block code should be inside a pre with block styling
      expect(blockCode.closest('pre')).toHaveClass('bg-gray-50', 'p-2', 'sm:p-4', 'rounded-lg');
    });

    test('ensures text formatting is visually distinct and readable', () => {
      const content = `
# Coaching Advice

**Important:** Make sure to warm up before training.

*Remember:* Consistency is key to improvement.

Use the \`heart rate monitor\` during workouts.

\`\`\`
Training Schedule:
Monday: Easy run
Tuesday: Intervals
Wednesday: Rest
\`\`\`
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Verify all elements are rendered with proper contrast and readability
      const heading = screen.getByRole('heading', { level: 1 });
      const boldText = screen.getByText('Important:');
      const italicText = screen.getByText('Remember:');
      const inlineCode = screen.getByText('heart rate monitor');
      const codeBlock = screen.getByText(/Training Schedule:/);
      
      // Check color contrast classes for readability
      expect(heading).toHaveClass('text-gray-900'); // High contrast
      expect(boldText).toHaveClass('text-gray-900'); // High contrast for emphasis
      expect(italicText).toHaveClass('text-gray-800'); // Good contrast
      expect(inlineCode).toHaveClass('text-gray-800', 'bg-gray-100'); // Good contrast with background
      expect(codeBlock.closest('pre')).toHaveClass('bg-gray-50'); // Subtle background for code blocks
    });
  });

  describe('Table Rendering', () => {
    test('renders tables with proper structure and styling', () => {
      const content = `
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Row 1 C1 | Row 1 C2 | Row 1 C3 |
| Row 2 C1 | Row 2 C2 | Row 2 C3 |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();
      expect(table).toHaveClass('min-w-full', 'divide-y', 'divide-gray-200');
      
      // Check table wrapper for responsive design
      const tableWrapper = table.closest('div');
      expect(tableWrapper).toHaveClass('overflow-x-auto', 'mb-3', 'sm:mb-4', 'rounded-lg', 'border', 'border-gray-200', 'shadow-sm');
    });

    test('renders table headers with proper styling', () => {
      const content = `
| Header 1 | Header 2 |
|----------|----------|
| Data 1   | Data 2   |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const headers = screen.getAllByRole('columnheader');
      expect(headers).toHaveLength(2);
      
      headers.forEach(header => {
        expect(header).toHaveClass(
          'px-2', 'sm:px-4', 'py-2', 'sm:py-3', 'text-left', 'text-xs', 'font-medium', 
          'text-gray-500', 'uppercase', 'tracking-wider', 'border-b', 'border-gray-200'
        );
      });
      
      expect(screen.getByText('Header 1')).toBeInTheDocument();
      expect(screen.getByText('Header 2')).toBeInTheDocument();
    });

    test('renders table data cells with proper styling', () => {
      const content = `
| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const dataCells = screen.getAllByRole('cell');
      expect(dataCells).toHaveLength(6); // 2 rows × 3 columns
      
      dataCells.forEach(cell => {
        expect(cell).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3', 'text-xs', 'sm:text-sm', 'text-gray-700', 'break-words');
      });
      
      expect(screen.getByText('John')).toBeInTheDocument();
      expect(screen.getByText('25')).toBeInTheDocument();
      expect(screen.getByText('NYC')).toBeInTheDocument();
    });

    test('applies hover effects to table rows', () => {
      const content = `
| Product | Price |
|---------|-------|
| Apple   | $1.00 |
| Orange  | $0.75 |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Get table rows (excluding header row)
      const table = screen.getByRole('table');
      const tbody = table.querySelector('tbody');
      const rows = tbody?.querySelectorAll('tr');
      
      expect(rows).toHaveLength(2);
      rows?.forEach(row => {
        expect(row).toHaveClass('hover:bg-gray-50', 'transition-colors', 'duration-150');
      });
    });

    test('renders thead and tbody with proper styling', () => {
      const content = `
| Status | Count |
|--------|-------|
| Active | 10    |
| Inactive | 5   |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      const thead = table.querySelector('thead');
      const tbody = table.querySelector('tbody');
      
      expect(thead).toHaveClass('bg-gray-50');
      expect(tbody).toHaveClass('bg-white', 'divide-y', 'divide-gray-200');
    });

    test('ensures tables work well on different screen sizes with responsive wrapper', () => {
      const content = `
| Very Long Header Name 1 | Very Long Header Name 2 | Very Long Header Name 3 | Very Long Header Name 4 |
|-------------------------|-------------------------|-------------------------|-------------------------|
| Very long data content that might overflow | More long content | Even more content | Final column data |
| Another row with long content | Second column | Third column | Fourth column |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      const wrapper = table.closest('div');
      
      // Check responsive wrapper
      expect(wrapper).toHaveClass('overflow-x-auto');
      expect(table).toHaveClass('min-w-full');
      
      // Verify all content is rendered
      expect(screen.getByText('Very Long Header Name 1')).toBeInTheDocument();
      expect(screen.getByText('Very long data content that might overflow')).toBeInTheDocument();
    });

    test('renders complex tables with proper borders and readability', () => {
      const content = `
| Training Week | Monday | Tuesday | Wednesday | Thursday | Friday |
|---------------|--------|---------|-----------|----------|--------|
| Week 1        | Rest   | 5K Easy | Intervals | Rest     | Long Run |
| Week 2        | Cross  | Tempo   | Rest      | 5K Easy  | Long Run |
| Week 3        | Rest   | Hills   | Recovery  | Tempo    | Race     |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      
      // Check table structure and borders
      expect(table).toHaveClass('divide-y', 'divide-gray-200');
      
      // Check wrapper has proper borders
      const wrapper = table.closest('div');
      expect(wrapper).toHaveClass('border', 'border-gray-200', 'rounded-lg');
      
      // Verify content readability
      expect(screen.getByText('Training Week')).toBeInTheDocument();
      expect(screen.getByText('Intervals')).toBeInTheDocument();
      expect(screen.getAllByText('Long Run')).toHaveLength(2); // Appears in Week 1 and Week 2
    });

    test('handles empty table cells gracefully', () => {
      const content = `
| Name | Value | Notes |
|------|-------|-------|
| Test |       | Empty middle |
|      | 123   | Empty name |
|      |       | Both empty |
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();
      
      // Should render without issues
      expect(screen.getByText('Test')).toBeInTheDocument();
      expect(screen.getByText('123')).toBeInTheDocument();
      expect(screen.getByText('Empty middle')).toBeInTheDocument();
    });

    test('maintains table styling consistency with other markdown elements', () => {
      const content = `
# Training Data

Here's your weekly progress:

| Week | Distance | Time | Pace |
|------|----------|------|------|
| 1    | 5.0 km   | 25:00| 5:00 |
| 2    | 5.2 km   | 24:30| 4:42 |

**Note:** Times are improving!
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Check that table styling is consistent with overall design
      const heading = screen.getByRole('heading', { level: 1 });
      const table = screen.getByRole('table');
      const boldText = screen.getByText('Note:');
      
      // All should use consistent gray color scheme
      expect(heading).toHaveClass('text-gray-900');
      expect(table.querySelector('th')).toHaveClass('text-gray-500');
      expect(table.querySelector('td')).toHaveClass('text-gray-700');
      expect(boldText).toHaveClass('text-gray-900');
    });
  });

  describe('Special Elements', () => {
    test('renders blockquotes with proper styling and left border', () => {
      const content = '> This is a blockquote with important coaching advice.';
      
      render(<MarkdownRenderer content={content} />);
      
      const blockquote = screen.getByText('This is a blockquote with important coaching advice.').closest('blockquote');
      expect(blockquote).toHaveClass(
        'border-l-4', 'border-blue-200', 'pl-3', 'sm:pl-4', 'py-2', 'mb-3', 'sm:mb-4', 
        'bg-blue-50', 'text-gray-700', 'italic', 'rounded-r-md', 'text-sm', 'sm:text-base'
      );
    });

    test('renders multi-line blockquotes correctly', () => {
      const content = `
> This is the first line of a blockquote.
> This is the second line.
> And this is the third line.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const blockquote = screen.getByText(/This is the first line/).closest('blockquote');
      expect(blockquote).toBeInTheDocument();
      expect(blockquote).toHaveClass('border-l-4', 'bg-blue-50', 'italic');
      
      // Should contain all lines
      expect(screen.getByText(/first line/)).toBeInTheDocument();
      expect(screen.getByText(/second line/)).toBeInTheDocument();
      expect(screen.getByText(/third line/)).toBeInTheDocument();
    });

    test('renders nested blockquotes with proper styling', () => {
      const content = `
> This is a blockquote.
> 
> > This is a nested blockquote.
> 
> Back to the first level.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Both blockquotes should be rendered with proper styling
      const blockquotes = screen.getAllByText(/blockquote/).map(el => el.closest('blockquote'));
      expect(blockquotes.length).toBeGreaterThan(0);
      
      blockquotes.forEach(blockquote => {
        if (blockquote) {
          expect(blockquote).toHaveClass('border-l-4', 'border-blue-200', 'bg-blue-50');
        }
      });
    });

    test('renders links with proper styling and security attributes', () => {
      const content = 'Check out [this training guide](https://example.com/training) for more info.';
      
      render(<MarkdownRenderer content={content} />);
      
      const link = screen.getByRole('link', { name: 'this training guide' });
      expect(link).toHaveAttribute('href', 'https://example.com/training');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
      expect(link).toHaveClass(
        'text-blue-600', 'hover:text-blue-800', 'underline', 
        'decoration-blue-300', 'hover:decoration-blue-500', 
        'transition-colors', 'duration-150'
      );
    });

    test('renders multiple links with consistent styling', () => {
      const content = `
Visit [our website](https://example.com) and check out the [training plans](https://example.com/plans).
Also see [this article](https://example.com/article) for more details.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      const links = screen.getAllByRole('link');
      expect(links).toHaveLength(3);
      
      links.forEach(link => {
        expect(link).toHaveAttribute('target', '_blank');
        expect(link).toHaveAttribute('rel', 'noopener noreferrer');
        expect(link).toHaveClass('text-blue-600', 'underline', 'transition-colors');
      });
      
      expect(screen.getByRole('link', { name: 'our website' })).toHaveAttribute('href', 'https://example.com');
      expect(screen.getByRole('link', { name: 'training plans' })).toHaveAttribute('href', 'https://example.com/plans');
      expect(screen.getByRole('link', { name: 'this article' })).toHaveAttribute('href', 'https://example.com/article');
    });

    test('renders links within other formatted text', () => {
      const content = '**Important:** Visit [our safety guidelines](https://example.com/safety) before training.';
      
      render(<MarkdownRenderer content={content} />);
      
      const link = screen.getByRole('link', { name: 'our safety guidelines' });
      const boldText = screen.getByText('Important:');
      
      expect(link).toHaveAttribute('href', 'https://example.com/safety');
      expect(link).toHaveClass('text-blue-600', 'underline');
      expect(boldText).toHaveClass('font-semibold', 'text-gray-900');
    });

    test('renders horizontal rules with appropriate spacing', () => {
      const content = `
# Section 1

Some content here.

---

# Section 2

More content here.

---

# Section 3
      `;
      
      const { container } = render(<MarkdownRenderer content={content} />);
      const hrElements = container.querySelectorAll('hr');
      
      expect(hrElements).toHaveLength(2);
      hrElements.forEach(hr => {
        expect(hr).toHaveClass('my-4', 'sm:my-6', 'border-t', 'border-gray-200');
      });
      
      // Verify sections are properly separated
      expect(screen.getAllByText('Section 1')).toHaveLength(1);
      expect(screen.getAllByText('Section 2')).toHaveLength(1);
      expect(screen.getAllByText('Section 3')).toHaveLength(1);
    });

    test('renders horizontal rules between different content types', () => {
      const content = `
## Training Schedule

| Day | Activity |
|-----|----------|
| Mon | Rest     |
| Tue | Run      |

---

> Remember to stay hydrated during training.

---

**Next week:** Increase intensity by 10%.
      `;
      
      const { container } = render(<MarkdownRenderer content={content} />);
      const hrElements = container.querySelectorAll('hr');
      
      expect(hrElements).toHaveLength(2);
      
      // Verify all content types are rendered
      expect(screen.getAllByRole('table')).toHaveLength(1);
      expect(screen.getByText(/Remember to stay hydrated/)).toBeInTheDocument();
      expect(screen.getByText('Next week:')).toBeInTheDocument();
    });

    test('ensures all special elements follow consistent design patterns', () => {
      const content = `
# Coaching Guidelines

> **Important Note:** Always warm up before intense training.

For more information, visit [our training portal](https://example.com/portal).

---

## Safety First

> *Remember:* Listen to your body and rest when needed.

Additional resources: [Safety Guidelines](https://example.com/safety)

---

## Final Thoughts

> Training consistently is better than training intensely once in a while.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Check consistent color scheme across elements
      const blockquotes = screen.getAllByText(/Important Note|Remember|Training consistently/).map(el => el.closest('blockquote'));
      const links = screen.getAllByRole('link');
      const { container } = render(<MarkdownRenderer content={content} />);
      const hrElements = container.querySelectorAll('hr');
      
      // Blockquotes should have consistent blue theme
      blockquotes.forEach(blockquote => {
        if (blockquote) {
          expect(blockquote).toHaveClass('border-blue-200', 'bg-blue-50', 'text-gray-700');
        }
      });
      
      // Links should have consistent blue theme
      links.forEach(link => {
        expect(link).toHaveClass('text-blue-600', 'hover:text-blue-800');
      });
      
      // HR elements should have consistent gray theme
      hrElements.forEach(hr => {
        expect(hr).toHaveClass('border-gray-200');
      });
    });

    test('renders special elements with proper accessibility attributes', () => {
      const content = `
> This is an important coaching tip.

Visit [our accessibility guide](https://example.com/accessibility) for more info.

---
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Check blockquote is properly structured
      const blockquote = screen.getByText(/important coaching tip/).closest('blockquote');
      expect(blockquote?.tagName).toBe('BLOCKQUOTE');
      
      // Check link has proper security attributes
      const link = screen.getByRole('link', { name: 'our accessibility guide' });
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
      expect(link).toHaveAttribute('target', '_blank');
      
      // Check HR is properly structured
      const { container } = render(<MarkdownRenderer content={content} />);
      const hr = container.querySelector('hr');
      expect(hr?.tagName).toBe('HR');
    });

    test('handles special elements in complex nested structures', () => {
      const content = `
# Training Program

## Week 1

> **Goal:** Build aerobic base
> 
> Focus on easy runs and [proper form](https://example.com/form).

| Day | Workout | Duration |
|-----|---------|----------|
| Mon | Easy run | 30 min |
| Wed | Rest | - |
| Fri | Easy run | 45 min |

---

## Week 2

> **Goal:** Introduce speed work
> 
> Add intervals while maintaining [good technique](https://example.com/technique).

**Important:** Don't skip the warm-up!

---

> *Final note:* Progress takes time. Be patient with yourself.
      `;
      
      render(<MarkdownRenderer content={content} />);
      
      // Verify all elements render correctly in complex structure
      expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Training Program');
      expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(2);
      
      // Check blockquotes
      const blockquotes = screen.getAllByText(/Goal:|Final note:/).map(el => el.closest('blockquote'));
      expect(blockquotes.length).toBeGreaterThan(0);
      
      // Check links within blockquotes
      expect(screen.getByRole('link', { name: 'proper form' })).toBeInTheDocument();
      expect(screen.getByRole('link', { name: 'good technique' })).toBeInTheDocument();
      
      // Check table
      expect(screen.getByRole('table')).toBeInTheDocument();
      
      // Check horizontal rules separate sections
      const { container } = render(<MarkdownRenderer content={content} />);
      const hrElements = container.querySelectorAll('hr');
      expect(hrElements.length).toBeGreaterThanOrEqual(2);
    });
  });
});