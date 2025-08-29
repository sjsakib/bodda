import React from 'react';
import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import { DiagramLibraryProvider, useDiagramLibrary } from '../DiagramLibraryContext';

// Test component that uses the context
const TestComponent: React.FC = () => {
  const { libraryState } = useDiagramLibrary();
  
  return (
    <div>
      <div data-testid="mermaid-loaded">{libraryState.mermaid.loaded.toString()}</div>
      <div data-testid="mermaid-loading">{libraryState.mermaid.loading.toString()}</div>
      <div data-testid="mermaid-error">{libraryState.mermaid.error || 'null'}</div>
      
      <div data-testid="vega-loaded">{libraryState.vegaLite.loaded.toString()}</div>
      <div data-testid="vega-loading">{libraryState.vegaLite.loading.toString()}</div>
      <div data-testid="vega-error">{libraryState.vegaLite.error || 'null'}</div>
    </div>
  );
};

describe('DiagramLibraryContext - Basic Functionality', () => {
  it('provides initial state correctly', () => {
    render(
      <DiagramLibraryProvider>
        <TestComponent />
      </DiagramLibraryProvider>
    );

    expect(screen.getByTestId('mermaid-loaded')).toHaveTextContent('false');
    expect(screen.getByTestId('mermaid-loading')).toHaveTextContent('false');
    expect(screen.getByTestId('mermaid-error')).toHaveTextContent('null');
    
    expect(screen.getByTestId('vega-loaded')).toHaveTextContent('false');
    expect(screen.getByTestId('vega-loading')).toHaveTextContent('false');
    expect(screen.getByTestId('vega-error')).toHaveTextContent('null');
  });

  it('throws error when used outside provider', () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    expect(() => {
      render(<TestComponent />);
    }).toThrow('useDiagramLibrary must be used within a DiagramLibraryProvider');
    
    consoleSpy.mockRestore();
  });

  it('provides all required context methods', () => {
    const TestMethodsComponent: React.FC = () => {
      const { loadMermaid, loadVegaLite, loadAllLibraries } = useDiagramLibrary();
      
      return (
        <div>
          <div data-testid="has-load-mermaid">{typeof loadMermaid === 'function' ? 'true' : 'false'}</div>
          <div data-testid="has-load-vega">{typeof loadVegaLite === 'function' ? 'true' : 'false'}</div>
          <div data-testid="has-load-all">{typeof loadAllLibraries === 'function' ? 'true' : 'false'}</div>
        </div>
      );
    };

    render(
      <DiagramLibraryProvider>
        <TestMethodsComponent />
      </DiagramLibraryProvider>
    );

    expect(screen.getByTestId('has-load-mermaid')).toHaveTextContent('true');
    expect(screen.getByTestId('has-load-vega')).toHaveTextContent('true');
    expect(screen.getByTestId('has-load-all')).toHaveTextContent('true');
  });
});