import React from 'react';
import { render } from '@testing-library/react';
import { describe, test } from 'vitest';
import MarkdownRenderer from '../MarkdownRenderer';

/**
 * Visual test component to demonstrate special elements rendering
 * This test is primarily for visual verification during development
 */
describe('SpecialElementsVisualTest', () => {
  test('renders comprehensive special elements showcase', () => {
    const content = `
# Training Guide: Special Elements Showcase

This guide demonstrates the special markdown elements available in our coaching platform.

## Blockquotes for Important Information

> **Important Safety Note:** Always consult with a healthcare professional before starting any new training program.

> *Coach's Tip:* Consistency beats intensity when building an aerobic base. Focus on easy runs that you can maintain conversationally.

### Multi-line Blockquotes

> This is a longer coaching note that spans multiple lines.
> 
> It provides detailed guidance on training principles and helps athletes understand the reasoning behind workout recommendations.
> 
> Remember: every athlete is different, so adjust these guidelines based on your individual needs and responses.

---

## External Resources and Links

For additional information, check out these resources:

- Visit our [Training Philosophy](https://example.com/philosophy) page
- Read about [Proper Running Form](https://example.com/form) techniques
- Download the [Training Log Template](https://example.com/template)

**Safety Resources:**
- [Injury Prevention Guide](https://example.com/injury-prevention)
- [Nutrition Guidelines](https://example.com/nutrition)

---

## Horizontal Rules for Section Separation

The horizontal rules above and below help separate different sections of content, making it easier to scan and read coaching advice.

---

## Combined Elements Example

> **Weekly Focus:** Base Building Phase
> 
> This week we're focusing on aerobic development. Check out our [base building guide](https://example.com/base-building) for detailed information.
> 
> Key principles:
> - Keep efforts conversational
> - Focus on consistency over speed
> - Listen to your body

**Additional Resources:**
- [Heart Rate Training Zones](https://example.com/hr-zones)
- [Recovery Strategies](https://example.com/recovery)

---

## Nested Blockquotes

> This is a primary coaching note.
> 
> > This is a nested quote, perhaps from a research study or expert opinion.
> > 
> > "The key to endurance development is consistent, moderate-intensity training over extended periods."
> 
> As coaches, we apply this research to create sustainable training programs.

---

## Final Notes

> *Remember:* Training is a journey, not a destination. Be patient with yourself and trust the process.

For questions or support, contact us through our [support portal](https://example.com/support).
    `;

    const { container } = render(<MarkdownRenderer content={content} />);
    
    // This test primarily serves as a visual verification
    // The actual functionality is tested in the main test suite
    console.log('Special elements visual test rendered successfully');
    
    // Basic verification that content is rendered
    expect(container.querySelector('blockquote')).toBeTruthy();
    expect(container.querySelector('a')).toBeTruthy();
    expect(container.querySelector('hr')).toBeTruthy();
  });

  test('renders special elements with consistent styling patterns', () => {
    const content = `
# Styling Consistency Test

> **Blockquote 1:** This uses the blue theme with left border.

[Link 1](https://example.com/link1) - This uses blue colors with hover effects.

---

> **Blockquote 2:** This should match the styling of the first blockquote.

[Link 2](https://example.com/link2) - This should match the styling of the first link.

---

> **Final blockquote:** All blockquotes should have consistent blue theming.

[Final link](https://example.com/final) - All links should have consistent blue theming.
    `;

    const { container } = render(<MarkdownRenderer content={content} />);
    
    // Verify consistent styling across multiple instances
    const blockquotes = container.querySelectorAll('blockquote');
    const links = container.querySelectorAll('a');
    const hrs = container.querySelectorAll('hr');
    
    // All blockquotes should have consistent classes
    blockquotes.forEach(blockquote => {
      expect(blockquote.className).toContain('border-blue-200');
      expect(blockquote.className).toContain('bg-blue-50');
    });
    
    // All links should have consistent classes
    links.forEach(link => {
      expect(link.className).toContain('text-blue-600');
      expect(link.className).toContain('hover:text-blue-800');
    });
    
    // All HRs should have consistent classes
    hrs.forEach(hr => {
      expect(hr.className).toContain('border-gray-200');
      expect(hr.className).toContain('my-6');
    });
  });
});