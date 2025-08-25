import React from 'react';
import MarkdownRenderer from '../MarkdownRenderer';

/**
 * Visual test component for list rendering
 * This component demonstrates various list scenarios to verify visual styling
 */
export const ListVisualTest: React.FC = () => {
  const testContent = `
# List Rendering Test

## Unordered Lists

### Simple unordered list:
- First item
- Second item with some longer text that should wrap properly and maintain good readability
- Third item

### Nested unordered lists:
- Top level item 1
  - Nested item 1.1
  - Nested item 1.2
    - Deep nested item 1.2.1
    - Deep nested item 1.2.2
  - Nested item 1.3
- Top level item 2
- Top level item 3

## Ordered Lists

### Simple ordered list:
1. First numbered item
2. Second numbered item with longer content
3. Third numbered item

### Nested ordered lists:
1. First main point
   1. Sub-point 1.1
   2. Sub-point 1.2
      1. Deep sub-point 1.2.1
2. Second main point
3. Third main point

## Mixed Lists

### Training Plan Example:
1. **Week 1**: Base building phase
   - Run 3 times per week at easy pace
   - Focus on building aerobic base
   - Target: 20-30 minutes per run
   
2. **Week 2**: Add some structure
   - Continue base runs (2x per week)
   - Add 1 tempo run (20 minutes)
   - Include dynamic warm-up routine
   
3. **Week 3**: Increase intensity
   - Base runs: 2x per week
   - Tempo run: 1x per week (25 minutes)
   - Hill repeats: 1x per week
     - 6-8 repeats of 1 minute uphill
     - Easy jog recovery between repeats

### Equipment Checklist:
- **Essential gear:**
  - Proper running shoes
  - Moisture-wicking clothing
  - Water bottle
- **Optional but helpful:**
  - Heart rate monitor
  - GPS watch
  - Foam roller for recovery

## Lists with Formatting

### Workout Types:
1. **Easy runs** - conversational pace, builds aerobic base
2. **Tempo runs** - comfortably hard effort, improves lactate threshold
3. **Intervals** - short bursts at high intensity
   - 400m repeats
   - 800m repeats
   - Mile repeats
4. **Long runs** - builds endurance and mental toughness

### Recovery Methods:
- *Active recovery*: light jogging or walking
- *Passive recovery*: complete rest
- *Cross-training*: swimming, cycling, or yoga
- *Stretching and mobility work*
  - Dynamic warm-up before runs
  - Static stretching after runs
  - Foam rolling for tight muscles
`;

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-4">List Rendering Visual Test</h1>
        <p className="text-gray-600 mb-4">
          This test demonstrates the visual styling of various list types and nesting scenarios.
        </p>
      </div>
      
      <div className="border border-gray-200 rounded-lg p-6 bg-gray-50">
        <MarkdownRenderer content={testContent} />
      </div>
    </div>
  );
};

export default ListVisualTest;