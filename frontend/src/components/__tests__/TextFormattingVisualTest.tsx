import React from 'react';
import { render } from '@testing-library/react';
import { describe, test } from 'vitest';
import MarkdownRenderer from '../MarkdownRenderer';

/**
 * Visual test component to demonstrate text formatting capabilities
 * This test is primarily for visual verification during development
 */
describe('Text Formatting Visual Tests', () => {
  test('renders comprehensive text formatting examples', () => {
    const content = `
# Text Formatting Examples

## Bold and Italic Text

**This is bold text** that stands out clearly.

*This is italic text* that provides emphasis.

***This is bold and italic*** for maximum emphasis.

## Inline Code

Use the \`useState\` hook for state management.

The \`console.log()\` function is useful for debugging.

Configure your environment with \`NODE_ENV=production\`.

## Code Blocks

Here's a JavaScript example:

\`\`\`javascript
const greeting = "Hello, World!";
console.log(greeting);

function calculateSum(a, b) {
  return a + b;
}
\`\`\`

Python example:

\`\`\`python
def hello_world():
    print("Hello, World!")
    
def calculate_sum(a, b):
    return a + b
\`\`\`

Shell commands:

\`\`\`bash
npm install react
npm run build
echo "Build complete"
\`\`\`

## Mixed Formatting

**Important:** Always use \`try-catch\` blocks when handling *asynchronous operations*.

The \`fetch()\` API returns a **Promise** that you should handle with *proper error handling*.

\`\`\`typescript
async function fetchData(): Promise<Data> {
  try {
    const response = await fetch('/api/data');
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch data:', error);
    throw error;
  }
}
\`\`\`

## Coaching Context Examples

**Training Tip:** Use a *heart rate monitor* during workouts to track intensity.

Monitor your \`resting heart rate\` each morning for recovery insights.

**Workout Plan:**

\`\`\`
Week 1: Base Building
- Monday: Easy run (60 minutes)
- Tuesday: Rest or cross-training
- Wednesday: Tempo run (45 minutes)
- Thursday: Easy run (45 minutes)
- Friday: Rest
- Saturday: Long run (90 minutes)
- Sunday: Recovery run (30 minutes)
\`\`\`

*Remember:* Consistency is more important than **intensity** when building your aerobic base.
    `;

    const { container } = render(<MarkdownRenderer content={content} />);
    
    // This test primarily serves as a visual verification
    // The actual assertions are covered in the main test suite
    expect(container.firstChild).toHaveClass('markdown-content');
    
    // Log the rendered HTML for visual inspection during development
    if (process.env.NODE_ENV === 'development') {
      console.log('Rendered HTML:', container.innerHTML);
    }
  });

  test('renders edge cases and special formatting', () => {
    const content = `
# Edge Cases

## Nested and Complex Formatting

**Bold with \`inline code\` inside**

*Italic with \`inline code\` inside*

\`Code with **bold** inside\` (should not render bold)

## Multiple Code Blocks

Inline: \`first\`, \`second\`, and \`third\` code snippets.

\`\`\`
Plain code block without language
const x = 1;
const y = 2;
\`\`\`

\`\`\`json
{
  "name": "example",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0"
  }
}
\`\`\`

## Formatting in Lists

- **Bold item** in list
- *Italic item* in list  
- Item with \`inline code\`
- **Bold** and *italic* and \`code\` together

1. **First** numbered item
2. *Second* numbered item
3. Item with \`code snippet\`

## Long Code Block

\`\`\`javascript
// This is a longer code block to test scrolling and formatting
const longFunctionName = (parameterOne, parameterTwo, parameterThree) => {
  const result = parameterOne + parameterTwo + parameterThree;
  console.log('This is a very long line that might cause horizontal scrolling on smaller screens');
  return result;
};

// Multiple lines with various formatting
const config = {
  apiUrl: 'https://api.example.com/v1',
  timeout: 5000,
  retries: 3
};
\`\`\`
    `;

    const { container } = render(<MarkdownRenderer content={content} />);
    
    expect(container.firstChild).toHaveClass('markdown-content');
  });
});