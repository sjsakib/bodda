import React from 'react';
import MarkdownRenderer from '../MarkdownRenderer';

/**
 * Visual test component to demonstrate table rendering capabilities.
 * This component shows various table examples to verify styling and responsiveness.
 */
export const TableVisualTest: React.FC = () => {
  const simpleTableContent = `
# Simple Table Example

| Name | Age | City |
|------|-----|------|
| John | 25  | New York |
| Jane | 30  | Los Angeles |
| Bob  | 35  | Chicago |
  `;

  const trainingTableContent = `
# Training Schedule

| Week | Monday | Tuesday | Wednesday | Thursday | Friday | Saturday | Sunday |
|------|--------|---------|-----------|----------|--------|----------|--------|
| 1    | Rest   | 5K Easy | Intervals | Rest     | Tempo  | Long Run | Cross  |
| 2    | Rest   | 6K Easy | Hills     | Rest     | Tempo  | Long Run | Cross  |
| 3    | Rest   | 7K Easy | Intervals | Rest     | Race   | Recovery | Rest   |
  `;

  const dataTableContent = `
# Performance Metrics

| Metric | Week 1 | Week 2 | Week 3 | Week 4 | Improvement |
|--------|--------|--------|--------|--------|-------------|
| Distance (km) | 25.0 | 28.5 | 32.1 | 35.8 | +43.2% |
| Average Pace | 5:30 | 5:15 | 5:05 | 4:58 | +10.9% |
| Heart Rate | 165 | 162 | 158 | 155 | -6.1% |
| Recovery Time | 48h | 42h | 38h | 36h | -25.0% |
  `;

  const emptyTableContent = `
# Table with Empty Cells

| Task | Assigned | Status | Notes |
|------|----------|--------|-------|
| Setup | John |  | In progress |
| Testing |  | Complete | All tests pass |
| Documentation |  |  | Pending |
  `;

  return (
    <div className="p-8 max-w-6xl mx-auto space-y-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">
        Table Rendering Visual Tests
      </h1>
      
      <div className="space-y-12">
        <div className="border border-gray-200 rounded-lg p-6 bg-white">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">Simple Table</h2>
          <MarkdownRenderer content={simpleTableContent} />
        </div>

        <div className="border border-gray-200 rounded-lg p-6 bg-white">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">Wide Training Table (Responsive)</h2>
          <MarkdownRenderer content={trainingTableContent} />
        </div>

        <div className="border border-gray-200 rounded-lg p-6 bg-white">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">Data Table with Numbers</h2>
          <MarkdownRenderer content={dataTableContent} />
        </div>

        <div className="border border-gray-200 rounded-lg p-6 bg-white">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">Table with Empty Cells</h2>
          <MarkdownRenderer content={emptyTableContent} />
        </div>
      </div>

      <div className="mt-12 p-6 bg-blue-50 rounded-lg">
        <h3 className="text-lg font-medium text-blue-900 mb-2">Testing Instructions</h3>
        <ul className="text-blue-800 space-y-1">
          <li>• Resize the browser window to test responsive behavior</li>
          <li>• Hover over table rows to see hover effects</li>
          <li>• Check that tables scroll horizontally on mobile devices</li>
          <li>• Verify proper borders, spacing, and typography</li>
          <li>• Ensure headers are visually distinct from data cells</li>
        </ul>
      </div>
    </div>
  );
};

export default TableVisualTest;