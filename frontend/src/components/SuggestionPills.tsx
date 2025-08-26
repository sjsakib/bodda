import React, { useRef, useCallback } from 'react';

// TypeScript interfaces
interface SuggestionPill {
  id: string;
  text: string;
  icon: string; // emoji or icon character
  category: 'training' | 'goals' | 'progress' | 'help';
}

interface SuggestionPillsProps {
  onPillClick: (text: string) => void;
  className?: string;
}

// Predefined suggestion data based on requirements 3.1, 3.2, 3.3 and 5.1-5.4
const SUGGESTION_PILLS: SuggestionPill[] = [
  {
    id: 'training-plan',
    text: 'Help me plan my next training week',
    icon: '💪',
    category: 'training'
  },
  {
    id: 'performance-analysis',
    text: 'Analyze my recent running performance',
    icon: '🏃‍♂️',
    category: 'training'
  },
  {
    id: 'strength-focus',
    text: 'What strength training should I focus on?',
    icon: '🏋️‍♀️',
    category: 'training'
  },
  {
    id: 'set-goals',
    text: 'Help me set realistic training goals',
    icon: '🎯',
    category: 'goals'
  },
  {
    id: 'improve-times',
    text: 'How can I improve my race times?',
    icon: '📈',
    category: 'goals'
  },
  {
    id: 'progress-trends',
    text: 'Show me my training progress trends',
    icon: '📊',
    category: 'progress'
  }
];

const SuggestionPills: React.FC<SuggestionPillsProps> = ({ 
  onPillClick, 
  className = '' 
}) => {
  const pillRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handlePillClick = (pill: SuggestionPill) => {
    onPillClick(pill.text);
  };

  const focusPill = useCallback((index: number) => {
    const pill = pillRefs.current[index];
    if (pill) {
      pill.focus();
    }
  }, []);

  const handleKeyDown = (event: React.KeyboardEvent, pill: SuggestionPill, index: number) => {
    const totalPills = SUGGESTION_PILLS.length;
    const isDesktop = window.innerWidth >= 768; // md breakpoint
    const pillsPerRow = isDesktop ? 2 : 1;
    
    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        handlePillClick(pill);
        break;
      
      case 'ArrowRight':
        event.preventDefault();
        if (isDesktop) {
          // Move to next pill in row or first pill of next row
          const nextIndex = index + 1;
          if (nextIndex < totalPills) {
            focusPill(nextIndex);
          }
        } else {
          // On mobile, arrow right moves to next pill
          const nextIndex = (index + 1) % totalPills;
          focusPill(nextIndex);
        }
        break;
      
      case 'ArrowLeft':
        event.preventDefault();
        if (isDesktop) {
          // Move to previous pill in row or last pill of previous row
          const prevIndex = index - 1;
          if (prevIndex >= 0) {
            focusPill(prevIndex);
          }
        } else {
          // On mobile, arrow left moves to previous pill
          const prevIndex = index === 0 ? totalPills - 1 : index - 1;
          focusPill(prevIndex);
        }
        break;
      
      case 'ArrowDown':
        event.preventDefault();
        if (isDesktop) {
          // Move to pill in next row (same column)
          const nextRowIndex = index + pillsPerRow;
          if (nextRowIndex < totalPills) {
            focusPill(nextRowIndex);
          }
        } else {
          // On mobile, same as arrow right
          const nextIndex = (index + 1) % totalPills;
          focusPill(nextIndex);
        }
        break;
      
      case 'ArrowUp':
        event.preventDefault();
        if (isDesktop) {
          // Move to pill in previous row (same column)
          const prevRowIndex = index - pillsPerRow;
          if (prevRowIndex >= 0) {
            focusPill(prevRowIndex);
          }
        } else {
          // On mobile, same as arrow left
          const prevIndex = index === 0 ? totalPills - 1 : index - 1;
          focusPill(prevIndex);
        }
        break;
      
      case 'Home':
        event.preventDefault();
        focusPill(0);
        break;
      
      case 'End':
        event.preventDefault();
        focusPill(totalPills - 1);
        break;
    }
  };

  return (
    <div 
      className={`p-4 bg-gray-50 border-t border-gray-200 ${className}`}
      role="region"
      aria-label="Quick start suggestions"
      aria-describedby="suggestion-pills-description"
    >
      <div 
        id="suggestion-pills-description" 
        className="sr-only"
      >
        Use arrow keys to navigate between suggestions, Enter or Space to select. 
        These are quick-start prompts to help you begin your coaching conversation.
      </div>
      <div 
        className="grid grid-cols-1 md:grid-cols-2 gap-2 max-w-4xl mx-auto"
        role="group"
        aria-label="Suggestion buttons"
      >
        {SUGGESTION_PILLS.map((pill, index) => (
          <button
            key={pill.id}
            ref={(el) => (pillRefs.current[index] = el)}
            onClick={() => handlePillClick(pill)}
            onKeyDown={(e) => handleKeyDown(e, pill, index)}
            className="flex items-center gap-3 px-4 py-3 min-h-[44px] bg-white border border-gray-200 rounded-lg text-sm text-gray-700 hover:bg-gray-50 hover:border-gray-300 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 shadow-sm hover:shadow-md focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none text-left cursor-pointer touch-manipulation"
            type="button"
            tabIndex={0}
            aria-label={`Quick start suggestion: ${pill.text}. Category: ${pill.category}. Press Enter or Space to select.`}
            aria-describedby={`pill-${pill.id}-description`}
          >
            <span 
              className="text-lg flex-shrink-0" 
              role="img" 
              aria-label={`${pill.category} icon`}
            >
              {pill.icon}
            </span>
            <span className="flex-1 leading-relaxed">{pill.text}</span>
            <span 
              id={`pill-${pill.id}-description`} 
              className="sr-only"
            >
              This will populate the input field with: {pill.text}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
};

export default SuggestionPills;