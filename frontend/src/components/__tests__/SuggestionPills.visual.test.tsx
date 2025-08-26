import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SuggestionPills from '../SuggestionPills';

describe('SuggestionPills Visual and Responsive Tests', () => {
  const mockOnPillClick = vi.fn();

  beforeEach(() => {
    mockOnPillClick.mockClear();
  });

  describe('Layout and Grid Responsiveness', () => {
    it('should render single column layout on mobile screens', () => {
      // Mock mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Verify mobile grid classes
      expect(gridContainer).toHaveClass('grid');
      expect(gridContainer).toHaveClass('grid-cols-1'); // Single column on mobile
      expect(gridContainer).toHaveClass('md:grid-cols-2'); // Two columns on desktop
      expect(gridContainer).toHaveClass('gap-2');
      expect(gridContainer).toHaveClass('max-w-4xl');
      expect(gridContainer).toHaveClass('mx-auto');
    });

    it('should render two column layout on desktop screens', () => {
      // Mock desktop viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Verify responsive grid classes are present
      expect(gridContainer).toHaveClass('grid-cols-1'); // Base mobile
      expect(gridContainer).toHaveClass('md:grid-cols-2'); // Desktop override
    });

    it('should render tablet layout correctly', () => {
      // Mock tablet viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // At md breakpoint (768px), should switch to 2-column
      expect(gridContainer).toHaveClass('md:grid-cols-2');
    });

    it('should have proper container styling and spacing', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      
      // Verify container styling
      expect(container).toHaveClass('p-4'); // Padding
      expect(container).toHaveClass('bg-gray-50'); // Background
      expect(container).toHaveClass('border-t'); // Top border
      expect(container).toHaveClass('border-gray-200'); // Border color
    });

    it('should render all 6 pills in correct visual structure', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      expect(buttons).toHaveLength(6);

      // Verify each button has proper visual structure
      buttons.forEach((button) => {
        // Check flex layout
        expect(button).toHaveClass('flex');
        expect(button).toHaveClass('items-center');
        expect(button).toHaveClass('gap-3');

        // Check icon span exists
        const iconSpan = button.querySelector('span[role="img"]');
        expect(iconSpan).toBeInTheDocument();
        expect(iconSpan).toHaveClass('text-lg');
        expect(iconSpan).toHaveClass('flex-shrink-0');

        // Check text span exists
        const textSpan = button.querySelector('span.flex-1');
        expect(textSpan).toBeInTheDocument();
        expect(textSpan).toHaveClass('leading-relaxed');
      });
    });
  });

  describe('Visual Styling and Appearance', () => {
    it('should have proper pill styling and colors', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Base styling
        expect(button).toHaveClass('px-4');
        expect(button).toHaveClass('py-3');
        expect(button).toHaveClass('bg-white');
        expect(button).toHaveClass('border');
        expect(button).toHaveClass('border-gray-200');
        expect(button).toHaveClass('rounded-lg');
        expect(button).toHaveClass('text-sm');
        expect(button).toHaveClass('text-gray-700');
        expect(button).toHaveClass('text-left');
        expect(button).toHaveClass('cursor-pointer');

        // Shadow styling
        expect(button).toHaveClass('shadow-sm');
      });
    });

    it('should have proper hover state styling', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Hover state classes
        expect(button).toHaveClass('hover:bg-gray-50');
        expect(button).toHaveClass('hover:border-gray-300');
        expect(button).toHaveClass('hover:scale-[1.02]');
        expect(button).toHaveClass('hover:shadow-md');
      });
    });

    it('should have proper active state styling', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Active state classes
        expect(button).toHaveClass('active:scale-[0.98]');
      });
    });

    it('should have smooth transitions', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Transition classes
        expect(button).toHaveClass('transition-all');
        expect(button).toHaveClass('duration-200');
      });
    });

    it('should display correct icons and text content', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      // Verify specific pills with their icons and text
      expect(screen.getByText('💪')).toBeInTheDocument();
      expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument();
      
      expect(screen.getByText('🏃‍♂️')).toBeInTheDocument();
      expect(screen.getByText('Analyze my recent running performance')).toBeInTheDocument();
      
      expect(screen.getByText('🏋️‍♀️')).toBeInTheDocument();
      expect(screen.getByText('What strength training should I focus on?')).toBeInTheDocument();
      
      expect(screen.getByText('🎯')).toBeInTheDocument();
      expect(screen.getByText('Help me set realistic training goals')).toBeInTheDocument();
      
      expect(screen.getByText('📈')).toBeInTheDocument();
      expect(screen.getByText('How can I improve my race times?')).toBeInTheDocument();
      
      expect(screen.getByText('📊')).toBeInTheDocument();
      expect(screen.getByText('Show me my training progress trends')).toBeInTheDocument();
    });
  });

  describe('Focus States and Accessibility Compliance', () => {
    it('should have proper focus ring styling', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Focus ring styling
        expect(button).toHaveClass('focus:ring-2');
        expect(button).toHaveClass('focus:ring-blue-500');
        expect(button).toHaveClass('focus:ring-offset-2');
        expect(button).toHaveClass('focus:outline-none');
      });
    });

    it('should maintain focus visibility when navigating with keyboard', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Focus first button
      buttons[0].focus();
      expect(document.activeElement).toBe(buttons[0]);
      
      // Navigate with arrow key
      fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[1]);
      
      // Verify focus is still visible (focus classes should be applied)
      expect(buttons[1]).toHaveClass('focus:ring-2');
    });

    it('should have proper focus order for tab navigation', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // All buttons should be in tab order
      buttons.forEach((button) => {
        expect(button).toHaveAttribute('tabIndex', '0');
      });
    });

    it('should handle focus states correctly on different screen sizes', () => {
      // Test mobile focus
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { rerender } = render(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      let buttons = screen.getAllByRole('button');
      buttons[0].focus();
      
      // Mobile navigation should work
      fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[1]);

      // Test desktop focus
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      rerender(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      buttons = screen.getAllByRole('button');
      buttons[0].focus();
      
      // Desktop navigation should work differently
      fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[2]); // Should move to next row
    });
  });

  describe('Mobile Touch Target Compliance', () => {
    it('should have minimum 44px touch target height', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Minimum touch target size (44px)
        expect(button).toHaveClass('min-h-[44px]');
      });
    });

    it('should have touch-optimized interaction', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Touch manipulation for better mobile performance
        expect(button).toHaveClass('touch-manipulation');
      });
    });

    it('should have adequate padding for touch targets', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Adequate padding for touch
        expect(button).toHaveClass('px-4'); // 16px horizontal
        expect(button).toHaveClass('py-3'); // 12px vertical
      });
    });

    it('should handle touch interactions properly', () => {
      // Mock mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const firstButton = screen.getAllByRole('button')[0];
      
      // Simulate touch interaction
      fireEvent.click(firstButton);
      
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });
  });

  describe('Color Contrast and Visual Hierarchy', () => {
    it('should have proper text color contrast', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Text should have good contrast (gray-700 on white background)
        expect(button).toHaveClass('text-gray-700');
        expect(button).toHaveClass('bg-white');
      });
    });

    it('should have proper border contrast', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Border should be visible but subtle
        expect(button).toHaveClass('border-gray-200');
      });
    });

    it('should have proper container background contrast', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      
      // Container should have subtle background
      expect(container).toHaveClass('bg-gray-50');
      expect(container).toHaveClass('border-gray-200');
    });

    it('should have proper visual hierarchy with icons and text', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        const iconSpan = button.querySelector('span[role="img"]');
        const textSpan = button.querySelector('span.flex-1');
        
        // Icon should be prominent but not overwhelming
        expect(iconSpan).toHaveClass('text-lg');
        
        // Text should be readable and properly sized
        expect(textSpan).toHaveClass('leading-relaxed');
      });
    });
  });

  describe('Responsive Behavior Across Screen Sizes', () => {
    it('should adapt layout from mobile to tablet', () => {
      // Start with mobile
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { rerender } = render(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      let gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      expect(gridContainer).toHaveClass('grid-cols-1');

      // Switch to tablet (768px is md breakpoint)
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      rerender(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      // Should still have responsive classes
      expect(gridContainer).toHaveClass('md:grid-cols-2');
    });

    it('should maintain proper spacing across screen sizes', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      
      // Gap should be consistent
      expect(gridContainer).toHaveClass('gap-2');
      
      // Max width should be set for larger screens
      expect(gridContainer).toHaveClass('max-w-4xl');
      expect(gridContainer).toHaveClass('mx-auto');
    });

    it('should handle keyboard navigation differently on mobile vs desktop', () => {
      // Test mobile navigation pattern
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { rerender } = render(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      let buttons = screen.getAllByRole('button');
      buttons[0].focus();
      
      // On mobile, ArrowDown should behave like ArrowRight
      fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[1]);

      // Test desktop navigation pattern
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      rerender(<SuggestionPills onPillClick={mockOnPillClick} />);
      
      buttons = screen.getAllByRole('button');
      buttons[0].focus();
      
      // On desktop, ArrowDown should move to next row
      fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[2]);
    });
  });

  describe('Visual Interaction States', () => {
    it('should show proper visual feedback on hover', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Should have hover state classes for visual feedback
        expect(button).toHaveClass('hover:bg-gray-50');
        expect(button).toHaveClass('hover:border-gray-300');
        expect(button).toHaveClass('hover:scale-[1.02]');
        expect(button).toHaveClass('hover:shadow-md');
      });
    });

    it('should show proper visual feedback on active state', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        // Should have active state for press feedback
        expect(button).toHaveClass('active:scale-[0.98]');
      });
    });

    it('should maintain visual consistency across all pills', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // All buttons should have identical styling classes
      const firstButtonClasses = buttons[0].className;
      
      buttons.forEach((button) => {
        expect(button.className).toBe(firstButtonClasses);
      });
    });
  });
});