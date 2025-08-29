import React from 'react';
import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import { 
  DiagramLoadingIndicator, 
  LibraryLoadingIndicator, 
  DiagramErrorIndicator 
} from '../DiagramLoadingIndicator';

describe('DiagramLoadingIndicator', () => {
  it('renders with default mermaid message', () => {
    render(<DiagramLoadingIndicator type="mermaid" />);
    
    expect(screen.getByText('Loading Mermaid diagram...')).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveAttribute('aria-label', 'Loading diagram');
  });

  it('renders with default vega-lite message', () => {
    render(<DiagramLoadingIndicator type="vega-lite" />);
    
    expect(screen.getByText('Loading chart...')).toBeInTheDocument();
  });

  it('renders with default general message', () => {
    render(<DiagramLoadingIndicator type="general" />);
    
    expect(screen.getByText('Loading diagram...')).toBeInTheDocument();
  });

  it('renders with custom message', () => {
    const customMessage = 'Custom loading message';
    render(<DiagramLoadingIndicator type="mermaid" message={customMessage} />);
    
    expect(screen.getByText(customMessage)).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const customClass = 'custom-class';
    const { container } = render(
      <DiagramLoadingIndicator type="mermaid" className={customClass} />
    );
    
    expect(container.firstChild).toHaveClass(customClass);
  });

  it('renders with small size variant', () => {
    render(<DiagramLoadingIndicator type="mermaid" size="small" />);
    
    const spinner = screen.getByRole('status');
    expect(spinner).toHaveClass('h-4', 'w-4');
  });

  it('renders with medium size variant (default)', () => {
    render(<DiagramLoadingIndicator type="mermaid" size="medium" />);
    
    const spinner = screen.getByRole('status');
    expect(spinner).toHaveClass('h-6', 'w-6');
  });

  it('renders with large size variant', () => {
    render(<DiagramLoadingIndicator type="mermaid" size="large" />);
    
    const spinner = screen.getByRole('status');
    expect(spinner).toHaveClass('h-8', 'w-8');
  });
});

describe('LibraryLoadingIndicator', () => {
  it('renders nothing when no libraries are loading', () => {
    const { container } = render(<LibraryLoadingIndicator loadingLibraries={[]} />);
    
    expect(container.firstChild).toBeNull();
  });

  it('renders when Mermaid is loading', () => {
    render(<LibraryLoadingIndicator loadingLibraries={['mermaid']} />);
    
    expect(screen.getByText('Initializing diagram libraries...')).toBeInTheDocument();
    expect(screen.getByText('Loading: Mermaid')).toBeInTheDocument();
  });

  it('renders when Vega-Lite is loading', () => {
    render(<LibraryLoadingIndicator loadingLibraries={['vega-lite']} />);
    
    expect(screen.getByText('Loading: Vega-Lite')).toBeInTheDocument();
  });

  it('renders when multiple libraries are loading', () => {
    render(<LibraryLoadingIndicator loadingLibraries={['mermaid', 'vega-lite']} />);
    
    expect(screen.getByText('Loading: Mermaid, Vega-Lite')).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const customClass = 'custom-library-class';
    const { container } = render(
      <LibraryLoadingIndicator 
        loadingLibraries={['mermaid']} 
        className={customClass} 
      />
    );
    
    expect(container.firstChild).toHaveClass(customClass);
  });

  it('has proper accessibility attributes', () => {
    render(<LibraryLoadingIndicator loadingLibraries={['mermaid']} />);
    
    const spinner = screen.getByRole('status');
    expect(spinner).toHaveAttribute('aria-label', 'Loading diagram libraries');
  });
});

describe('DiagramErrorIndicator', () => {
  const defaultProps = {
    error: 'Test error message',
  };

  it('renders error message with default title', () => {
    render(<DiagramErrorIndicator {...defaultProps} />);
    
    expect(screen.getByText('Diagram Error')).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });

  it('renders with mermaid-specific title', () => {
    render(<DiagramErrorIndicator {...defaultProps} type="mermaid" />);
    
    expect(screen.getByText('Mermaid Diagram Error')).toBeInTheDocument();
  });

  it('renders with vega-lite-specific title', () => {
    render(<DiagramErrorIndicator {...defaultProps} type="vega-lite" />);
    
    expect(screen.getByText('Chart Error')).toBeInTheDocument();
  });

  it('renders retry button when onRetry is provided', () => {
    const onRetry = vi.fn();
    render(<DiagramErrorIndicator {...defaultProps} onRetry={onRetry} />);
    
    const retryButton = screen.getByText('Try again');
    expect(retryButton).toBeInTheDocument();
    
    retryButton.click();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('does not render retry button when onRetry is not provided', () => {
    render(<DiagramErrorIndicator {...defaultProps} />);
    
    expect(screen.queryByText('Try again')).not.toBeInTheDocument();
  });

  it('shows raw content when enabled', () => {
    const rawContent = 'graph TD\nA-->B';
    render(
      <DiagramErrorIndicator 
        {...defaultProps} 
        showRawContent={true} 
        rawContent={rawContent} 
      />
    );
    
    const summary = screen.getByText('Show raw content');
    expect(summary).toBeInTheDocument();
    
    // Click to expand details
    summary.click();
    // Check for the content in the pre element, accounting for HTML encoding
    expect(screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && 
             content.includes('graph TD') && 
             content.includes('A-->B');
    })).toBeInTheDocument();
  });

  it('does not show raw content when disabled', () => {
    render(
      <DiagramErrorIndicator 
        {...defaultProps} 
        showRawContent={false} 
        rawContent="some content" 
      />
    );
    
    expect(screen.queryByText('Show raw content')).not.toBeInTheDocument();
  });

  it('applies custom className', () => {
    const customClass = 'custom-error-class';
    const { container } = render(
      <DiagramErrorIndicator {...defaultProps} className={customClass} />
    );
    
    expect(container.firstChild).toHaveClass(customClass);
  });

  it('renders error icon', () => {
    const { container } = render(<DiagramErrorIndicator {...defaultProps} />);
    
    // Check for the SVG element directly since it has aria-hidden
    const errorIcon = container.querySelector('svg[aria-hidden="true"]');
    expect(errorIcon).toBeInTheDocument();
    expect(errorIcon).toHaveClass('text-red-400');
  });

  it('has proper focus management for retry button', () => {
    const onRetry = vi.fn();
    render(<DiagramErrorIndicator {...defaultProps} onRetry={onRetry} />);
    
    const retryButton = screen.getByText('Try again');
    expect(retryButton).toHaveClass('focus:outline-none', 'focus:ring-2', 'focus:ring-red-500');
  });
});