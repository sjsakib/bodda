import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, test, expect } from 'vitest';
import { MarkdownRenderer } from '../MarkdownRenderer';

describe('MarkdownRenderer Responsive Visual Tests', () => {
  test('demonstrates responsive design optimizations for mobile devices', () => {
    const complexContent = `
# Training Plan Overview

This is your personalized training plan with **important** information.

## Weekly Schedule

Here are your training sessions:

- **Monday**: Easy run (30 minutes)
- **Tuesday**: Interval training with *high intensity*
- **Wednesday**: Rest day or cross-training
- **Thursday**: Tempo run (45 minutes)
- **Friday**: Recovery run (20 minutes)
- **Saturday**: Long run (60-90 minutes)
- **Sunday**: Complete rest

### Training Zones

| Zone | Heart Rate | Effort Level | Duration |
|------|------------|--------------|----------|
| Zone 1 | 120-140 bpm | Very Easy | 30-60 min |
| Zone 2 | 140-160 bpm | Easy | 20-90 min |
| Zone 3 | 160-180 bpm | Moderate | 10-40 min |
| Zone 4 | 180-200 bpm | Hard | 5-20 min |
| Zone 5 | 200+ bpm | Very Hard | 1-5 min |

> **Important Note**: Always listen to your body and adjust intensity based on how you feel. Recovery is just as important as training.

### Code Example for Heart Rate Calculation

Here's how to calculate your target heart rate:

\`\`\`javascript
function calculateTargetHR(age, restingHR, intensity) {
  const maxHR = 220 - age;
  const hrReserve = maxHR - restingHR;
  return Math.round(restingHR + (hrReserve * intensity));
}

// Example usage
const targetHR = calculateTargetHR(30, 60, 0.7); // 70% intensity
console.log(\`Target HR: \${targetHR} bpm\`);
\`\`\`

You can also use inline code like \`Math.max()\` for quick calculations.

---

For more information, visit [our training guide](https://example.com/training-guide-with-very-long-url-that-might-overflow-on-mobile-devices).

**Remember**: Consistency is key to improvement!
    `;

    render(<MarkdownRenderer content={complexContent} />);

    // Verify all elements are rendered
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { level: 3 })).toHaveLength(2);
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByRole('link')).toBeInTheDocument();

    // Verify responsive classes are applied to headings
    const h1 = screen.getByRole('heading', { level: 1 });
    expect(h1).toHaveClass('text-xl', 'sm:text-2xl', 'md:text-3xl');
    expect(h1).toHaveClass('mb-3', 'sm:mb-4');

    const h2 = screen.getByRole('heading', { level: 2 });
    expect(h2).toHaveClass('text-lg', 'sm:text-xl', 'md:text-2xl');
    expect(h2).toHaveClass('mb-2', 'sm:mb-3');

    const h3Elements = screen.getAllByRole('heading', { level: 3 });
    h3Elements.forEach(h3 => {
      expect(h3).toHaveClass('text-base', 'sm:text-lg', 'md:text-xl');
    });

    // Verify table has mobile-optimized styling
    const table = screen.getByRole('table');
    const tableWrapper = table.parentElement;
    expect(tableWrapper).toHaveClass('overflow-x-auto');
    expect(tableWrapper).toHaveClass('mb-3', 'sm:mb-4');

    // Verify table cells have responsive padding
    const headerCells = screen.getAllByRole('columnheader');
    headerCells.forEach(cell => {
      expect(cell).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3');
    });

    const dataCells = screen.getAllByRole('cell');
    dataCells.forEach(cell => {
      expect(cell).toHaveClass('px-2', 'sm:px-4', 'py-2', 'sm:py-3');
      expect(cell).toHaveClass('break-words'); // For long content on mobile
    });

    // Verify lists have responsive spacing
    const lists = screen.getAllByRole('list');
    lists.forEach(list => {
      expect(list).toHaveClass('pl-4', 'sm:pl-6');
      expect(list).toHaveClass('mb-3', 'sm:mb-4');
    });

    // Verify code elements have responsive styling
    const inlineCode = screen.getByText('Math.max()');
    expect(inlineCode).toHaveClass('text-xs', 'sm:text-sm');
    expect(inlineCode).toHaveClass('px-1', 'sm:px-1.5');

    const codeBlock = screen.getByText(/calculateTargetHR/);
    const preElement = codeBlock.closest('pre');
    expect(preElement).toHaveClass('p-2', 'sm:p-4');
    expect(preElement).toHaveClass('text-xs', 'sm:text-sm');

    // Verify blockquote has responsive styling
    const blockquote = screen.getByText(/Always listen to your body/);
    const blockquoteElement = blockquote.closest('blockquote');
    expect(blockquoteElement).toHaveClass('pl-3', 'sm:pl-4');
    expect(blockquoteElement).toHaveClass('mb-3', 'sm:mb-4');
    expect(blockquoteElement).toHaveClass('text-sm', 'sm:text-base');

    // Verify link has word breaking for long URLs
    const link = screen.getByRole('link');
    expect(link).toHaveClass('break-words');

    // Verify horizontal rule has responsive spacing
    const container = screen.getByText('Training Plan Overview').closest('.markdown-content');
    const hrElement = container?.querySelector('hr');
    expect(hrElement).toHaveClass('my-4', 'sm:my-6');

    // Verify main container has responsive base font size
    expect(container).toHaveClass('text-sm', 'sm:text-base');
  });

  test('verifies touch-friendly spacing for mobile users', () => {
    const content = `
# Mobile-Friendly Content

This content is optimized for mobile devices with proper touch-friendly spacing.

## List Items

- First item with adequate spacing
- Second item that's easy to read
- Third item with proper line height

### Table for Mobile

| Exercise | Duration | Notes |
|----------|----------|-------|
| Running | 30 min | Easy pace |
| Cycling | 45 min | Moderate effort |
| Swimming | 20 min | Technique focus |

> This blockquote has proper mobile spacing and readability.
    `;

    render(<MarkdownRenderer content={content} />);

    // Verify list items have proper spacing for touch
    const listItems = screen.getAllByRole('listitem');
    listItems.forEach(item => {
      expect(item).toHaveClass('text-sm', 'sm:text-base');
      expect(item).toHaveClass('leading-relaxed');
    });

    // Verify table cells are touch-friendly
    const cells = screen.getAllByRole('cell');
    cells.forEach(cell => {
      // Smaller padding on mobile, larger on desktop
      expect(cell).toHaveClass('py-2', 'sm:py-3');
      expect(cell).toHaveClass('text-xs', 'sm:text-sm');
    });

    // Verify headings have appropriate mobile spacing
    const h1 = screen.getByRole('heading', { level: 1 });
    expect(h1).toHaveClass('mt-4', 'sm:mt-6'); // Less top margin on mobile
    expect(h1).toHaveClass('mb-3', 'sm:mb-4'); // Less bottom margin on mobile

    const h2 = screen.getByRole('heading', { level: 2 });
    expect(h2).toHaveClass('mt-3', 'sm:mt-5'); // Progressive spacing
    expect(h2).toHaveClass('mb-2', 'sm:mb-3');
  });

  test('ensures table overflow and scrolling work on small screens', () => {
    const wideTableContent = `
| Column 1 | Column 2 | Column 3 | Column 4 | Column 5 | Column 6 | Column 7 | Column 8 |
|----------|----------|----------|----------|----------|----------|----------|----------|
| Data that might be very long and cause overflow | More data | Even more | Additional | Extra | More cols | Even more | Final |
| Another row with potentially long content | Data 2 | Data 3 | Data 4 | Data 5 | Data 6 | Data 7 | Data 8 |
    `;

    render(<MarkdownRenderer content={wideTableContent} />);

    const table = screen.getByRole('table');
    const tableWrapper = table.parentElement;

    // Verify table wrapper enables horizontal scrolling
    expect(tableWrapper).toHaveClass('overflow-x-auto');
    
    // Verify table maintains minimum width
    expect(table).toHaveClass('min-w-full');

    // Verify cells use break-words instead of whitespace-nowrap for better mobile handling
    const cells = screen.getAllByRole('cell');
    cells.forEach(cell => {
      expect(cell).toHaveClass('break-words');
      expect(cell).not.toHaveClass('whitespace-nowrap');
    });
  });

  test('verifies text sizing and spacing work across different screen sizes', () => {
    const content = `
# Large Heading
## Medium Heading  
### Small Heading

Regular paragraph text that should be readable on all screen sizes.

- List item one
- List item two with longer text that might wrap on smaller screens
- List item three

\`inline code\` within text.

\`\`\`
Code block that should
have proper sizing and
padding on all screens
\`\`\`

> Blockquote with responsive text sizing
    `;

    render(<MarkdownRenderer content={content} />);

    // Verify progressive font sizing from mobile to desktop
    const h1 = screen.getByRole('heading', { level: 1 });
    expect(h1).toHaveClass('text-xl');      // Mobile base
    expect(h1).toHaveClass('sm:text-2xl');  // Small screens
    expect(h1).toHaveClass('md:text-3xl');  // Medium+ screens

    const h2 = screen.getByRole('heading', { level: 2 });
    expect(h2).toHaveClass('text-lg');      // Mobile base
    expect(h2).toHaveClass('sm:text-xl');   // Small screens
    expect(h2).toHaveClass('md:text-2xl');  // Medium+ screens

    const h3 = screen.getByRole('heading', { level: 3 });
    expect(h3).toHaveClass('text-base');    // Mobile base
    expect(h3).toHaveClass('sm:text-lg');   // Small screens
    expect(h3).toHaveClass('md:text-xl');   // Medium+ screens

    // Verify inline code responsive sizing
    const inlineCode = screen.getByText('inline code');
    expect(inlineCode).toHaveClass('text-xs', 'sm:text-sm');

    // Verify code block responsive sizing
    const codeBlock = screen.getByText(/Code block that should/);
    const preElement = codeBlock.closest('pre');
    expect(preElement).toHaveClass('text-xs', 'sm:text-sm');

    // Verify blockquote responsive sizing
    const blockquote = screen.getByText(/Blockquote with responsive/);
    const blockquoteElement = blockquote.closest('blockquote');
    expect(blockquoteElement).toHaveClass('text-sm', 'sm:text-base');

    // Verify main container has responsive base sizing
    const container = screen.getByText('Large Heading').closest('.markdown-content');
    expect(container).toHaveClass('text-sm', 'sm:text-base');
  });
});