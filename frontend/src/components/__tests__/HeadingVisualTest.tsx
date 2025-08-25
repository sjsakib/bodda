import React from 'react';
import MarkdownRenderer from '../MarkdownRenderer';

/**
 * Visual test component for heading rendering
 * This component demonstrates the heading hierarchy and styling
 */
export const HeadingVisualTest: React.FC = () => {
  const testContent = `
# Main Training Plan

This is the main heading for our training plan. It should be the largest and most prominent.

## Week 1: Base Building Phase

This is a secondary heading that introduces a major section. It should be smaller than H1 but still prominent.

### Monday: Easy Run

This is a tertiary heading for specific workouts. It should be smaller than H2 but still clearly a heading.

#### Warm-up Protocol

This is a fourth-level heading for detailed sections.

##### Specific Instructions

This is a fifth-level heading for very detailed subsections.

###### Additional Notes

This is the smallest heading level.

## Week 2: Build Phase

Another secondary heading to show consistency.

### Tuesday: Tempo Run

Another tertiary heading to demonstrate the hierarchy.

# Recovery Guidelines

Another main heading to show multiple H1s work properly.

## Nutrition

Secondary heading under recovery.

### Pre-Workout

Tertiary heading for specific nutrition timing.
  `;

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-6 text-blue-600">
        Heading Visual Test - MarkdownRenderer
      </h1>
      <div className="border border-gray-200 rounded-lg p-6 bg-white">
        <MarkdownRenderer content={testContent} />
      </div>
    </div>
  );
};

export default HeadingVisualTest;