import { renderHook } from '@testing-library/react';
import React from 'react';
import { vi } from 'vitest';
import { useDiagramLoader, useHasDiagrams } from '../useDiagramLoader';
import { DiagramLibraryProvider } from '../../contexts/DiagramLibraryContext';

// Mock the diagram detection utilities
vi.mock('../../utils/diagramDetection', () => ({
  detectDiagrams: vi.fn(),
  hasDiagramContent: vi.fn(),
}));

const { detectDiagrams, hasDiagramContent } = await import('../../utils/diagramDetection');

// Wrapper component for testing hooks
const wrapper = ({ children }: { children: React.ReactNode }) => 
  React.createElement(DiagramLibraryProvider, null, children);

describe('useDiagramLoader - Basic Functionality', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('detects content without diagrams', () => {
    const content = 'Regular markdown content without diagrams';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: false,
      mermaidCount: 0,
      vegaLiteCount: 0,
      totalCount: 0,
      diagrams: [],
    });
    
    (hasDiagramContent as any).mockReturnValue(false);

    const { result } = renderHook(() => useDiagramLoader(content, false), { wrapper });

    expect(result.current.hasDiagrams).toBe(false);
    expect(result.current.diagramInfo.totalCount).toBe(0);
    expect(result.current.allRequiredLibrariesLoaded).toBe(true);
    expect(result.current.isLoading).toBe(false);
  });

  it('detects Mermaid diagrams', () => {
    const content = '```mermaid\ngraph TD\nA-->B\n```';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: true,
      mermaidCount: 1,
      vegaLiteCount: 0,
      totalCount: 1,
      diagrams: [{ type: 'mermaid', content: 'graph TD\nA-->B' }],
    });
    
    (hasDiagramContent as any).mockReturnValue(true);

    const { result } = renderHook(() => useDiagramLoader(content, false), { wrapper });

    expect(result.current.hasDiagrams).toBe(true);
    expect(result.current.diagramInfo.mermaidCount).toBe(1);
    expect(result.current.diagramInfo.vegaLiteCount).toBe(0);
    expect(result.current.allRequiredLibrariesLoaded).toBe(false); // Libraries not loaded yet
  });

  it('detects Vega-Lite diagrams', () => {
    const content = '```vega-lite\n{"mark": "bar"}\n```';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: true,
      mermaidCount: 0,
      vegaLiteCount: 1,
      totalCount: 1,
      diagrams: [{ type: 'vega-lite', content: '{"mark": "bar"}' }],
    });
    
    (hasDiagramContent as any).mockReturnValue(true);

    const { result } = renderHook(() => useDiagramLoader(content, false), { wrapper });

    expect(result.current.hasDiagrams).toBe(true);
    expect(result.current.diagramInfo.vegaLiteCount).toBe(1);
    expect(result.current.diagramInfo.mermaidCount).toBe(0);
    expect(result.current.allRequiredLibrariesLoaded).toBe(false); // Libraries not loaded yet
  });

  it('detects mixed diagrams', () => {
    const content = '```mermaid\ngraph TD\nA-->B\n```\n\n```vega-lite\n{"mark": "bar"}\n```';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: true,
      mermaidCount: 1,
      vegaLiteCount: 1,
      totalCount: 2,
      diagrams: [
        { type: 'mermaid', content: 'graph TD\nA-->B' },
        { type: 'vega-lite', content: '{"mark": "bar"}' }
      ],
    });
    
    (hasDiagramContent as any).mockReturnValue(true);

    const { result } = renderHook(() => useDiagramLoader(content, false), { wrapper });

    expect(result.current.hasDiagrams).toBe(true);
    expect(result.current.diagramInfo.totalCount).toBe(2);
    expect(result.current.allRequiredLibrariesLoaded).toBe(false); // Libraries not loaded yet
  });

  it('updates when content changes', () => {
    let content = 'Regular content';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: false,
      mermaidCount: 0,
      vegaLiteCount: 0,
      totalCount: 0,
      diagrams: [],
    });
    
    (hasDiagramContent as any).mockReturnValue(false);

    const { result, rerender } = renderHook(
      ({ content }) => useDiagramLoader(content, false),
      { 
        wrapper,
        initialProps: { content }
      }
    );

    expect(result.current.hasDiagrams).toBe(false);

    // Update content to include diagrams
    content = '```mermaid\ngraph TD\nA-->B\n```';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: true,
      mermaidCount: 1,
      vegaLiteCount: 0,
      totalCount: 1,
      diagrams: [{ type: 'mermaid', content: 'graph TD\nA-->B' }],
    });
    
    (hasDiagramContent as any).mockReturnValue(true);

    rerender({ content });

    expect(result.current.hasDiagrams).toBe(true);
    expect(result.current.diagramInfo.mermaidCount).toBe(1);
  });
});

describe('useHasDiagrams - Basic Functionality', () => {
  it('returns true when content has diagrams', () => {
    const content = '```mermaid\ngraph TD\nA-->B\n```';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: true,
      mermaidCount: 1,
      vegaLiteCount: 0,
      totalCount: 1,
      diagrams: [{ type: 'mermaid', content: 'graph TD\nA-->B' }],
    });
    
    (hasDiagramContent as any).mockReturnValue(true);

    const { result } = renderHook(() => useHasDiagrams(content), { wrapper });

    expect(result.current).toBe(true);
  });

  it('returns false when content has no diagrams', () => {
    const content = 'Regular content';
    
    (detectDiagrams as any).mockReturnValue({
      hasDiagrams: false,
      mermaidCount: 0,
      vegaLiteCount: 0,
      totalCount: 0,
      diagrams: [],
    });
    
    (hasDiagramContent as any).mockReturnValue(false);

    const { result } = renderHook(() => useHasDiagrams(content), { wrapper });

    expect(result.current).toBe(false);
  });
});