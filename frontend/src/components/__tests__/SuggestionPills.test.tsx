import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SuggestionPills from '../SuggestionPills';

describe('SuggestionPills', () => {
  const mockOnPillClick = vi.fn();

  beforeEach(() => {
    mockOnPillClick.mockClear();
    // Reset window.innerWidth to default desktop size
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    });
  });

  describe('Component Rendering', () => {
    it('renders all suggestion pills with correct text and icons', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      // Check that all 6 pills are rendered
      const buttons = screen.getAllByRole('button');
      expect(buttons).toHaveLength(6);

      // Check specific pills with their icons and text
      expect(screen.getByText('Help me plan my next training week')).toBeInTheDocument();
      expect(screen.getByText('Analyze my recent running performance')).toBeInTheDocument();
      expect(screen.getByText('What strength training should I focus on?')).toBeInTheDocument();
      expect(screen.getByText('Help me set realistic training goals')).toBeInTheDocument();
      expect(screen.getByText('How can I improve my race times?')).toBeInTheDocument();
      expect(screen.getByText('Show me my training progress trends')).toBeInTheDocument();

      // Check that icons are present (emojis)
      expect(screen.getByText('💪')).toBeInTheDocument();
      expect(screen.getByText('🏃‍♂️')).toBeInTheDocument();
      expect(screen.getByText('🏋️‍♀️')).toBeInTheDocument();
      expect(screen.getByText('🎯')).toBeInTheDocument();
      expect(screen.getByText('📈')).toBeInTheDocument();
      expect(screen.getByText('📊')).toBeInTheDocument();
    });

    it('renders with custom className when provided', () => {
      const customClass = 'custom-test-class';
      render(<SuggestionPills onPillClick={mockOnPillClick} className={customClass} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      expect(container).toHaveClass(customClass);
    });

    it('renders with default styling when no className provided', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      expect(container).toHaveClass('p-4', 'bg-gray-50', 'border-t', 'border-gray-200');
    });

    it('renders pills with correct categories and structure', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Check that each button has the expected structure
      buttons.forEach((button) => {
        expect(button).toHaveClass('flex', 'items-center', 'gap-3');
        expect(button).toHaveAttribute('type', 'button');
        
        // Check for icon span
        const iconSpan = button.querySelector('span[role="img"]');
        expect(iconSpan).toBeInTheDocument();
        
        // Check for text span
        const textSpan = button.querySelector('span.flex-1');
        expect(textSpan).toBeInTheDocument();
      });
    });

    it('renders responsive grid layout classes', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const gridContainer = screen.getByRole('group', { name: 'Suggestion buttons' });
      expect(gridContainer).toHaveClass('grid', 'grid-cols-1', 'md:grid-cols-2', 'gap-2');
    });
  });

  describe('Click Event Handling', () => {
    it('calls onPillClick with correct text when pill is clicked', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const trainingPlanButton = screen.getByText('Help me plan my next training week');
      fireEvent.click(trainingPlanButton);

      expect(mockOnPillClick).toHaveBeenCalledTimes(1);
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('calls onPillClick for each different pill with correct text', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const expectedTexts = [
        'Help me plan my next training week',
        'Analyze my recent running performance',
        'What strength training should I focus on?',
        'Help me set realistic training goals',
        'How can I improve my race times?',
        'Show me my training progress trends'
      ];

      expectedTexts.forEach((text, index) => {
        const button = screen.getByText(text);
        fireEvent.click(button);
        
        expect(mockOnPillClick).toHaveBeenCalledWith(text);
      });

      expect(mockOnPillClick).toHaveBeenCalledTimes(6);
    });

    it('handles multiple clicks on same pill', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const button = screen.getByText('Help me plan my next training week');
      
      fireEvent.click(button);
      fireEvent.click(button);
      fireEvent.click(button);

      expect(mockOnPillClick).toHaveBeenCalledTimes(3);
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('does not call onPillClick when clicking on disabled elements', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      // Click on the container instead of buttons
      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      fireEvent.click(container);

      expect(mockOnPillClick).not.toHaveBeenCalled();
    });
  });

  describe('Keyboard Navigation', () => {
    it('activates pill with Enter key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const firstButton = screen.getAllByRole('button')[0];
      firstButton.focus();
      
      fireEvent.keyDown(firstButton, { key: 'Enter' });
      
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('activates pill with Space key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const secondButton = screen.getAllByRole('button')[1];
      secondButton.focus();
      
      fireEvent.keyDown(secondButton, { key: ' ' });
      
      expect(mockOnPillClick).toHaveBeenCalledWith('Analyze my recent running performance');
    });

    it('prevents default behavior for Enter and Space keys', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const button = screen.getAllByRole('button')[0];
      button.focus();
      
      // Create real event objects and spy on preventDefault
      const enterEvent = new KeyboardEvent('keydown', { key: 'Enter', bubbles: true });
      const spaceEvent = new KeyboardEvent('keydown', { key: ' ', bubbles: true });
      
      const enterPreventDefaultSpy = vi.spyOn(enterEvent, 'preventDefault');
      const spacePreventDefaultSpy = vi.spyOn(spaceEvent, 'preventDefault');
      
      // Dispatch real events
      button.dispatchEvent(enterEvent);
      button.dispatchEvent(spaceEvent);
      
      expect(enterPreventDefaultSpy).toHaveBeenCalled();
      expect(spacePreventDefaultSpy).toHaveBeenCalled();
    });

    describe('Desktop Navigation (2-column grid)', () => {
      beforeEach(() => {
        // Set desktop width
        Object.defineProperty(window, 'innerWidth', {
          writable: true,
          configurable: true,
          value: 1024,
        });
      });

      it('navigates right to next pill in same row', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[0].focus();
        
        fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
        
        expect(document.activeElement).toBe(buttons[1]);
      });

      it('navigates left to previous pill in same row', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[1].focus();
        
        fireEvent.keyDown(buttons[1], { key: 'ArrowLeft' });
        
        expect(document.activeElement).toBe(buttons[0]);
      });

      it('navigates down to next row (same column)', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[0].focus(); // First column, first row
        
        fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
        
        expect(document.activeElement).toBe(buttons[2]); // First column, second row
      });

      it('navigates up to previous row (same column)', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[2].focus(); // First column, second row
        
        fireEvent.keyDown(buttons[2], { key: 'ArrowUp' });
        
        expect(document.activeElement).toBe(buttons[0]); // First column, first row
      });

      it('does not navigate beyond grid boundaries', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        
        // Test right boundary (last button in row)
        buttons[1].focus();
        fireEvent.keyDown(buttons[1], { key: 'ArrowRight' });
        expect(document.activeElement).toBe(buttons[2]); // Should move to next row
        
        // Test left boundary (first button in row)
        buttons[0].focus();
        fireEvent.keyDown(buttons[0], { key: 'ArrowLeft' });
        expect(document.activeElement).toBe(buttons[0]); // Should stay in place
        
        // Test up boundary (first row)
        buttons[0].focus();
        fireEvent.keyDown(buttons[0], { key: 'ArrowUp' });
        expect(document.activeElement).toBe(buttons[0]); // Should stay in place
        
        // Test down boundary (last row)
        buttons[4].focus(); // Last row, first column
        fireEvent.keyDown(buttons[4], { key: 'ArrowDown' });
        expect(document.activeElement).toBe(buttons[4]); // Should stay in place
      });
    });

    describe('Mobile Navigation (1-column grid)', () => {
      beforeEach(() => {
        // Set mobile width
        Object.defineProperty(window, 'innerWidth', {
          writable: true,
          configurable: true,
          value: 375,
        });
      });

      it('navigates right to next pill with wrapping', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[0].focus();
        
        fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
        
        expect(document.activeElement).toBe(buttons[1]);
      });

      it('navigates left to previous pill with wrapping', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[0].focus();
        
        fireEvent.keyDown(buttons[0], { key: 'ArrowLeft' });
        
        expect(document.activeElement).toBe(buttons[5]); // Wraps to last pill
      });

      it('wraps from last to first pill when navigating right', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[5].focus(); // Last pill
        
        fireEvent.keyDown(buttons[5], { key: 'ArrowRight' });
        
        expect(document.activeElement).toBe(buttons[0]); // Wraps to first pill
      });

      it('treats arrow down same as arrow right on mobile', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[0].focus();
        
        fireEvent.keyDown(buttons[0], { key: 'ArrowDown' });
        
        expect(document.activeElement).toBe(buttons[1]);
      });

      it('treats arrow up same as arrow left on mobile', () => {
        render(<SuggestionPills onPillClick={mockOnPillClick} />);

        const buttons = screen.getAllByRole('button');
        buttons[1].focus();
        
        fireEvent.keyDown(buttons[1], { key: 'ArrowUp' });
        
        expect(document.activeElement).toBe(buttons[0]);
      });
    });

    it('navigates to first pill with Home key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons[3].focus(); // Focus middle pill
      
      fireEvent.keyDown(buttons[3], { key: 'Home' });
      
      expect(document.activeElement).toBe(buttons[0]);
    });

    it('navigates to last pill with End key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons[2].focus(); // Focus middle pill
      
      fireEvent.keyDown(buttons[2], { key: 'End' });
      
      expect(document.activeElement).toBe(buttons[5]);
    });

    it('ignores other key presses', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons[0].focus();
      
      const initialFocus = document.activeElement;
      
      // Test various keys that should be ignored
      fireEvent.keyDown(buttons[0], { key: 'Tab' });
      fireEvent.keyDown(buttons[0], { key: 'Escape' });
      fireEvent.keyDown(buttons[0], { key: 'a' });
      fireEvent.keyDown(buttons[0], { key: 'Delete' });
      
      expect(document.activeElement).toBe(initialFocus);
      expect(mockOnPillClick).not.toHaveBeenCalled();
    });
  });

  describe('Accessibility Features', () => {
    it('has proper tabindex for keyboard navigation', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveAttribute('tabIndex', '0');
      });
    });

    it('has proper ARIA labels for each pill', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        const ariaLabel = button.getAttribute('aria-label');
        expect(ariaLabel).toBeTruthy();
        expect(ariaLabel).toMatch(/Quick start suggestion:/);
        expect(ariaLabel).toMatch(/Category:/);
        expect(ariaLabel).toMatch(/Press Enter or Space to select/);
      });
    });

    it('has proper role and aria-label for main container', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const container = screen.getByRole('region', { name: 'Quick start suggestions' });
      expect(container).toBeInTheDocument();
      expect(container).toHaveAttribute('aria-describedby', 'suggestion-pills-description');
    });

    it('has proper role and aria-label for button group', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const group = screen.getByRole('group', { name: 'Suggestion buttons' });
      expect(group).toBeInTheDocument();
    });

    it('has screen reader instructions', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const instructions = screen.getByText(/Use arrow keys to navigate between suggestions/);
      expect(instructions).toBeInTheDocument();
      expect(instructions).toHaveClass('sr-only');
      expect(instructions).toHaveAttribute('id', 'suggestion-pills-description');
    });

    it('has proper icon accessibility attributes', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        const iconSpan = button.querySelector('span[role="img"]');
        expect(iconSpan).toBeInTheDocument();
        expect(iconSpan).toHaveAttribute('aria-label');
        
        const ariaLabel = iconSpan?.getAttribute('aria-label');
        expect(ariaLabel).toMatch(/(training|goals|progress) icon/);
      });
    });

    it('has hidden descriptions for each pill action', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        const describedBy = button.getAttribute('aria-describedby');
        expect(describedBy).toBeTruthy();
        
        if (describedBy) {
          const description = document.getElementById(describedBy);
          expect(description).toBeInTheDocument();
          expect(description).toHaveClass('sr-only');
          expect(description?.textContent).toMatch(/This will populate the input field with:/);
        }
      });
    });

    it('has proper focus management styles', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        expect(button).toHaveClass('focus:ring-2');
        expect(button).toHaveClass('focus:ring-blue-500');
        expect(button).toHaveClass('focus:ring-offset-2');
        expect(button).toHaveClass('focus:outline-none');
      });
    });

    it('has minimum touch target size for mobile accessibility', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons.forEach((button) => {
        expect(button).toHaveClass('min-h-[44px]');
        expect(button).toHaveClass('touch-manipulation');
      });
    });
  });

  describe('Prop Callbacks', () => {
    it('calls onPillClick prop function when provided', () => {
      const customCallback = vi.fn();
      render(<SuggestionPills onPillClick={customCallback} />);

      const button = screen.getByText('Help me plan my next training week');
      fireEvent.click(button);

      expect(customCallback).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('handles onPillClick prop changes', () => {
      const firstCallback = vi.fn();
      const secondCallback = vi.fn();
      
      const { rerender } = render(<SuggestionPills onPillClick={firstCallback} />);

      const button = screen.getByText('Help me plan my next training week');
      fireEvent.click(button);

      expect(firstCallback).toHaveBeenCalledWith('Help me plan my next training week');
      expect(secondCallback).not.toHaveBeenCalled();

      // Re-render with new callback
      rerender(<SuggestionPills onPillClick={secondCallback} />);

      fireEvent.click(button);

      expect(secondCallback).toHaveBeenCalledWith('Help me plan my next training week');
      expect(firstCallback).toHaveBeenCalledTimes(1); // Should not be called again
    });
  });
});