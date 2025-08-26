import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SuggestionPills from '../SuggestionPills';

describe('SuggestionPills Accessibility', () => {
  const mockOnPillClick = vi.fn();

  beforeEach(() => {
    mockOnPillClick.mockClear();
  });

  describe('ARIA Labels and Roles', () => {
    it('should have proper ARIA labels and roles for screen readers', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      // Check main container has proper role and label
      const region = screen.getByRole('region', { name: 'Quick start suggestions' });
      expect(region).toBeInTheDocument();

      // Check group role for button container
      const group = screen.getByRole('group', { name: 'Suggestion buttons' });
      expect(group).toBeInTheDocument();

      // Check all buttons have proper roles and labels
      const buttons = screen.getAllByRole('button');
      expect(buttons).toHaveLength(6);

      buttons.forEach((button) => {
        expect(button).toHaveAttribute('type', 'button');
        expect(button).toHaveAttribute('aria-label');
        expect(button.getAttribute('aria-label')).toMatch(/Quick start suggestion:/);
        expect(button.getAttribute('aria-label')).toMatch(/Category:/);
        expect(button.getAttribute('aria-label')).toMatch(/Press Enter or Space to select/);
      });
    });

    it('should have screen reader description for navigation instructions', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const description = screen.getByText(/Use arrow keys to navigate between suggestions/);
      expect(description).toBeInTheDocument();
      expect(description).toHaveClass('sr-only');
    });

    it('should have proper icon labels for screen readers', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      // Check that icons have proper aria-labels
      const trainingIcons = screen.getAllByLabelText('training icon');
      expect(trainingIcons).toHaveLength(3); // 3 training category pills

      const goalsIcons = screen.getAllByLabelText('goals icon');
      expect(goalsIcons).toHaveLength(2); // 2 goals category pills

      const progressIcons = screen.getAllByLabelText('progress icon');
      expect(progressIcons).toHaveLength(1); // 1 progress category pill
    });

    it('should have hidden descriptions for each pill action', () => {
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
  });

  describe('Keyboard Navigation', () => {
    it('should be focusable with tab navigation', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveAttribute('tabIndex', '0');
      });
    });

    it('should activate with Enter key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const firstButton = screen.getAllByRole('button')[0];
      firstButton.focus();
      
      fireEvent.keyDown(firstButton, { key: 'Enter' });
      
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('should activate with Space key', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const firstButton = screen.getAllByRole('button')[0];
      firstButton.focus();
      
      fireEvent.keyDown(firstButton, { key: ' ' });
      
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('should handle Enter and Space keys correctly', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const firstButton = screen.getAllByRole('button')[0];
      firstButton.focus();
      
      // Test Enter key triggers callback
      fireEvent.keyDown(firstButton, { key: 'Enter' });
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
      
      mockOnPillClick.mockClear();
      
      // Test Space key triggers callback
      fireEvent.keyDown(firstButton, { key: ' ' });
      expect(mockOnPillClick).toHaveBeenCalledWith('Help me plan my next training week');
    });

    it('should support arrow key navigation on desktop', () => {
      // Mock window.innerWidth for desktop
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Focus first button and navigate right
      buttons[0].focus();
      fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[1]);

      // Navigate down (should move to next row)
      fireEvent.keyDown(buttons[1], { key: 'ArrowDown' });
      expect(document.activeElement).toBe(buttons[3]);

      // Navigate left
      fireEvent.keyDown(buttons[3], { key: 'ArrowLeft' });
      expect(document.activeElement).toBe(buttons[2]);

      // Navigate up
      fireEvent.keyDown(buttons[2], { key: 'ArrowUp' });
      expect(document.activeElement).toBe(buttons[0]);
    });

    it('should support arrow key navigation on mobile', () => {
      // Mock window.innerWidth for mobile
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Focus first button and navigate right (should go to next pill)
      buttons[0].focus();
      fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[1]);

      // Navigate left from first button (should wrap to last)
      buttons[0].focus();
      fireEvent.keyDown(buttons[0], { key: 'ArrowLeft' });
      expect(document.activeElement).toBe(buttons[5]);
    });

    it('should support Home and End key navigation', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      // Focus middle button and press Home
      buttons[2].focus();
      fireEvent.keyDown(buttons[2], { key: 'Home' });
      expect(document.activeElement).toBe(buttons[0]);

      // Press End
      fireEvent.keyDown(buttons[0], { key: 'End' });
      expect(document.activeElement).toBe(buttons[5]);
    });
  });

  describe('Focus Management', () => {
    it('should have proper focus indicators', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveClass('focus:ring-2');
        expect(button).toHaveClass('focus:ring-blue-500');
        expect(button).toHaveClass('focus:ring-offset-2');
        expect(button).toHaveClass('focus:outline-none');
      });
    });

    it('should maintain focus after keyboard navigation', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      
      buttons[0].focus();
      expect(document.activeElement).toBe(buttons[0]);
      
      fireEvent.keyDown(buttons[0], { key: 'ArrowRight' });
      expect(document.activeElement).toBe(buttons[1]);
    });
  });

  describe('Touch and Mobile Accessibility', () => {
    it('should have proper touch target sizes', () => {
      render(<SuggestionPills onPillClick={mockOnPillClick} />);

      const buttons = screen.getAllByRole('button');
      buttons.forEach((button) => {
        expect(button).toHaveClass('min-h-[44px]');
        expect(button).toHaveClass('touch-manipulation');
      });
    });
  });
});