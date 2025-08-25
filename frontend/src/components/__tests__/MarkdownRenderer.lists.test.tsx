import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, test, expect } from 'vitest';
import MarkdownRenderer from '../MarkdownRenderer';

describe('MarkdownRenderer - Lists', () => {
  test('renders unordered lists with proper bullet styling', () => {
    const content = `
- First item
- Second item
- Third item
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const list = screen.getByRole('list');
    expect(list).toBeInTheDocument();
    expect(list.tagName).toBe('UL');
    
    // Check for proper CSS classes
    expect(list).toHaveClass('list-disc', 'list-outside', 'mb-3', 'sm:mb-4', 'space-y-1', 'text-gray-700', 'pl-4', 'sm:pl-6', 'marker:text-gray-400');
    
    // Check that all list items are present
    const listItems = screen.getAllByRole('listitem');
    expect(listItems).toHaveLength(3);
    expect(listItems[0]).toHaveTextContent('First item');
    expect(listItems[1]).toHaveTextContent('Second item');
    expect(listItems[2]).toHaveTextContent('Third item');
  });

  test('renders ordered lists with proper number styling', () => {
    const content = `
1. First numbered item
2. Second numbered item
3. Third numbered item
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const list = screen.getByRole('list');
    expect(list).toBeInTheDocument();
    expect(list.tagName).toBe('OL');
    
    // Check for proper CSS classes
    expect(list).toHaveClass('list-decimal', 'list-outside', 'mb-3', 'sm:mb-4', 'space-y-1', 'text-gray-700', 'pl-4', 'sm:pl-6', 'marker:text-gray-600', 'marker:font-medium');
    
    // Check that all list items are present
    const listItems = screen.getAllByRole('listitem');
    expect(listItems).toHaveLength(3);
    expect(listItems[0]).toHaveTextContent('First numbered item');
    expect(listItems[1]).toHaveTextContent('Second numbered item');
    expect(listItems[2]).toHaveTextContent('Third numbered item');
  });

  test('renders nested lists with proper visual hierarchy', () => {
    const content = `
- Top level item 1
  - Nested item 1.1
  - Nested item 1.2
    - Deep nested item 1.2.1
- Top level item 2
  1. Nested numbered item 2.1
  2. Nested numbered item 2.2
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const lists = screen.getAllByRole('list');
    expect(lists.length).toBeGreaterThan(1); // Should have multiple nested lists
    
    // Check that nested structure is maintained
    const listItems = screen.getAllByRole('listitem');
    expect(listItems.length).toBeGreaterThan(4); // Should have all items including nested ones
    
    // Verify content is present
    expect(screen.getByText('Top level item 1')).toBeInTheDocument();
    expect(screen.getByText('Nested item 1.1')).toBeInTheDocument();
    expect(screen.getByText('Deep nested item 1.2.1')).toBeInTheDocument();
    expect(screen.getByText('Nested numbered item 2.1')).toBeInTheDocument();
  });

  test('applies proper spacing and readability to list items', () => {
    const content = `
- Item with some longer text content that should wrap properly and maintain good readability
- Another item with **bold text** and *italic text* formatting
- Final item with inline code: \`console.log('test')\`
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const listItems = screen.getAllByRole('listitem');
    
    // Check that list items have proper styling classes
    listItems.forEach(item => {
      expect(item).toHaveClass('text-gray-700', 'leading-relaxed', 'pl-1');
    });
    
    // Verify formatted content is preserved
    expect(screen.getByText('bold text')).toBeInTheDocument();
    expect(screen.getByText('italic text')).toBeInTheDocument();
    expect(screen.getByText("console.log('test')")).toBeInTheDocument();
  });

  test('handles mixed list types correctly', () => {
    const content = `
## Training Plan

1. **Week 1**: Base building
   - Run 3 times per week
   - Focus on easy pace
   - Build aerobic base

2. **Week 2**: Add intensity
   - Continue base runs
   - Add 1 tempo run
   - Include hill repeats

### Equipment needed:
- Running shoes
- Heart rate monitor
- Comfortable clothing
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    // Should have both ordered and unordered lists
    const lists = screen.getAllByRole('list');
    expect(lists.length).toBeGreaterThanOrEqual(2);
    
    // Check content is properly rendered
    expect(screen.getByText('Week 1')).toBeInTheDocument();
    expect(screen.getByText('Run 3 times per week')).toBeInTheDocument();
    expect(screen.getByText('Running shoes')).toBeInTheDocument();
    expect(screen.getByText('Heart rate monitor')).toBeInTheDocument();
  });

  test('handles empty lists gracefully', () => {
    const content = `
- 

1. 
    `;
    
    render(<MarkdownRenderer content={content} />);
    
    const lists = screen.getAllByRole('list');
    expect(lists).toHaveLength(2);
    
    // Should still render the list structure even with empty items
    const listItems = screen.getAllByRole('listitem');
    expect(listItems).toHaveLength(2);
  });
});