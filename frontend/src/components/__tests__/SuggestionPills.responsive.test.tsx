import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SuggestionPills from '../SuggestionPills';

describe('SuggestionPills Responsive and Visual Compliance Tests', () => {
  const mockOnPillClick = vi.fn();

  beforeEach(() => {
    mockOnPillClick.mockClear();
  });

  describe('Screen Size Adaptations', () => {
    it('should adapt to very small mobile screens (320px)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 320,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Should still use single column layout
      expect(gridContainer).toHaveClass('grid-cols-1');
      expect(gridContainer).toHaveClass('md:grid-cols-2');
      
      // Container should have proper responsive padding
      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      expect(container).toHaveClass('p-4');
    });

    it('should adapt to large mobile screens (414px)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 414,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Pills should still be touch-friendly
      buttons.forEach((button) => {
        expect(button).toHaveClass('min-h-[44px]');
        expect(button).toHaveClass('px-4');
        expect(button).toHaveClass('py-3');
      });
    });

    it('should adapt to tablet screens (768px)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // At md breakpoint, should switch to 2-column
      expect(gridContainer).toHaveClass('md:grid-cols-2');
      expect(gridContainer).toHaveClass('max-w-4xl');
      expect(gridContainer).toHaveClass('mx-auto');
    });

    it('should adapt to large desktop screens (1440px)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1440,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Should maintain max-width constraint on large screens
      expect(gridContainer).toHaveClass('max-w-4xl');
      expect(gridContainer).toHaveClass('mx-auto');
    });

    it('should handle ultra-wide screens (1920px+)', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Should still respect max-width to prevent pills from becoming too wide
      expect(gridContainer).toHaveClass('max-w-4xl');
      expect(gridContainer).toHaveClass('mx-auto');
    });
  });

  describe('Touch Target Compliance (WCAG 2.1 AA)', () => {
    it('should meet minimum touch target size requirements', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // WCAG 2.1 AA requires minimum 44x44px touch targets
        expect(button).toHaveClass('min-h-[44px]');
        
        // Adequate horizontal padding for touch
        expect(button).toHaveClass('px-4'); // 16px each side
        expect(button).toHaveClass('py-3'); // 12px top/bottom
        
        // Touch manipulation for better mobile performance
        expect(button).toHaveClass('touch-manipulation');
      });
    });

    it('should have adequate spacing between touch targets', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Gap between pills should provide adequate separation
      expect(gridContainer).toHaveClass('gap-2'); // 8px gap
    });

    it('should maintain touch targets on different orientations', () => {
      // Test portrait mobile
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 667,
      });

      const { rerender } = render(<SuggestionPills onPillClick={mockOnPillClick} />);

      let buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveClass('min-h-[44px]');
      });

      // Test landscape mobile
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 667,
      });
      Object.defineProperty(window, 'innerHeight', {
        writable: true,
        configurable: true,
        value: 375,
      });

      rerender(<SuggestionPills onPillClick={mockOnPillClick} />);

      buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveClass('min-h-[44px]');
      });
    });
  });

  describe('Color Contrast Compliance (WCAG 2.1 AA)', () => {
    it('should use high contrast colors for text', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // text-gray-700 on bg-white provides good contrast ratio
        expect(button).toHaveClass('text-gray-700');
        expect(button).toHaveClass('bg-white');
      });
    });

    it('should have visible borders for component boundaries', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      const buttons = screen.getAllByRole('button');
      
      // Container should have visible top border
      expect(container).toHaveClass('border-t');
      expect(container).toHaveClass('border-gray-200');
      
      // Pills should have visible borders
      buttons.forEach((button) => {
        expect(button).toHaveClass('border');
        expect(button).toHaveClass('border-gray-200');
      });
    });

    it('should have proper focus indicators with high contrast', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Focus ring should be highly visible
        expect(button).toHaveClass('focus:ring-2');
        expect(button).toHaveClass('focus:ring-blue-500');
        expect(button).toHaveClass('focus:ring-offset-2');
        expect(button).toHaveClass('focus:outline-none');
      });
    });

    it('should maintain contrast in hover states', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Hover states should maintain good contrast
        expect(button).toHaveClass('hover:bg-gray-50');
        expect(button).toHaveClass('hover:border-gray-300');
        // Text color remains text-gray-700 for consistent contrast
      });
    });
  });

  describe('Responsive Typography and Spacing', () => {
    it('should use appropriate font sizes for readability', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Text should be readable on all screen sizes
        expect(button).toHaveClass('text-sm');
        
        // Icons should be appropriately sized
        const iconSpan = button.querySelector('span[role="img"]');
        expect(iconSpan).toHaveClass('text-lg');
        
        // Text should have good line height
        const textSpan = button.querySelector('span.flex-1');
        expect(textSpan).toHaveClass('leading-relaxed');
      });
    });

    it('should maintain proper spacing across screen sizes', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Container padding should be consistent
      expect(container).toHaveClass('p-4');
      
      // Grid gap should provide adequate separation
      expect(gridContainer).toHaveClass('gap-2');
    });

    it('should handle text wrapping gracefully', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Text should wrap properly if needed
        const textSpan = button.querySelector('span.flex-1');
        expect(textSpan).toHaveClass('leading-relaxed');
        
        // Button should maintain flex layout
        expect(button).toHaveClass('flex');
        expect(button).toHaveClass('items-center');
      });
    });
  });

  describe('Keyboard Navigation Responsiveness', () => {
    it('should adapt keyboard navigation to screen size', () => {
      // Test mobile keyboard navigation
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { rerender } = render(<SuggestionPills onPillClick={mockOnPillClick} />);

      let buttons = screen.getAllByRole('button');
      buttons[0].focus();

      // On mobile, ArrowDown should behave like ArrowRight (linear navigation)
      fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[1]);

      // Test desktop keyboard navigation
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      rerender(<SuggestionPills onPillClick={mockOnPillClick} />);

      buttons = screen.getAllByRole('button');
      buttons[0].focus();

      // On desktop, ArrowDown should move to next row (2-column grid)
      fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[2]);
    });

    it('should handle edge cases in keyboard navigation', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');

      // Test navigation from last button in first row
      buttons[1].focus();
      fireEvent.keyDown(buttons[1], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[2]); // Should move to next row

      // Test navigation from last button overall
      buttons[5].focus();
      fireEvent.keyDown(buttons[5], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[5]); // Should stay in place

      // Test navigation from first button
      buttons[0].focus();
      fireEvent.keyDown(buttons[0], { key: 'ArrowLeft' });
      expect(document.activeElement).toBe(buttons[0]); // Should stay in place
    });
  });

  describe('Performance and Animation Responsiveness', () => {
    it('should have smooth transitions for visual feedback', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Should have smooth transitions
        expect(button).toHaveClass('transition-all');
        expect(button).toHaveClass('duration-200');
        
        // Should have scale animations for feedback
        expect(button).toHaveClass('hover:scale-[1.02]');
        expect(button).toHaveClass('active:scale-[0.98]');
      });
    });

    it('should optimize for touch performance', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Should use touch-manipulation for better mobile performance
        expect(button).toHaveClass('touch-manipulation');
      });
    });
  });

  describe('Content Overflow and Wrapping', () => {
    it('should handle long text content gracefully', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Should use flex layout to handle content
        expect(button).toHaveClass('flex');
        expect(button).toHaveClass('items-center');
        
        // Text container should be flexible
        const textSpan = button.querySelector('span.flex-1');
        expect(textSpan).toHaveClass('flex-1');
        expect(textSpan).toHaveClass('leading-relaxed');
        
        // Icon should not shrink
        const iconSpan = button.querySelector('span[role="img"]');
        expect(iconSpan).toHaveClass('flex-shrink-0');
      });
    });

    it('should maintain layout integrity with varying content lengths', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // All buttons should have consistent styling despite different text lengths
      const firstButtonClasses = Array.from(buttons[0].classList);
      
      buttons.forEach((button) => {
        // Core layout classes should be consistent
        expect(button).toHaveClass('flex');
        expect(button).toHaveClass('items-center');
        expect(button).toHaveClass('gap-3');
        expect(button).toHaveClass('min-h-[44px]');
      });
    });
  });
});