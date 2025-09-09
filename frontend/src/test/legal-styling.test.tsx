import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, it, expect } from 'vitest';
import PrivacyPolicy from '../pages/PrivacyPolicy';
import TermsOfService from '../pages/TermsOfService';
import DataUsagePolicy from '../pages/DataUsagePolicy';

const renderWithRouter = (component: React.ReactElement) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('Legal Page Styling', () => {
  it('should apply legal-content class to Privacy Policy content', () => {
    renderWithRouter(<PrivacyPolicy />);
    
    // Check that the legal content container exists
    const contentContainer = document.querySelector('.legal-content');
    expect(contentContainer).toBeTruthy();
    
    // Check that headings are properly styled
    const headings = screen.getAllByRole('heading');
    expect(headings.length).toBeGreaterThan(0);
    
    // Check that the main title is present (using heading role to be more specific)
    expect(screen.getByRole('heading', { level: 1, name: 'Privacy Policy' })).toBeInTheDocument();
  });

  it('should apply legal-content class to Terms of Service content', () => {
    renderWithRouter(<TermsOfService />);
    
    // Check that the legal content container exists
    const contentContainer = document.querySelector('.legal-content');
    expect(contentContainer).toBeTruthy();
    
    // Check that the main title is present (using heading role to be more specific)
    expect(screen.getByRole('heading', { level: 1, name: 'Terms of Service' })).toBeInTheDocument();
  });

  it('should apply legal-content class to Data Usage Policy content', () => {
    renderWithRouter(<DataUsagePolicy />);
    
    // Check that the legal content container exists
    const contentContainer = document.querySelector('.legal-content');
    expect(contentContainer).toBeTruthy();
    
    // Check that the main title is present (using heading role to be more specific)
    expect(screen.getByRole('heading', { level: 1, name: 'Data Usage Policy' })).toBeInTheDocument();
  });

  it('should have proper callout boxes with enhanced styling', () => {
    renderWithRouter(<PrivacyPolicy />);
    
    // Check for callout boxes (they should have specific background colors)
    const calloutBoxes = document.querySelectorAll('.bg-yellow-50, .bg-red-50, .bg-blue-50, .bg-green-50, .bg-orange-50');
    expect(calloutBoxes.length).toBeGreaterThan(0);
  });

  it('should have proper section structure with headings', () => {
    renderWithRouter(<PrivacyPolicy />);
    
    // Check for proper heading hierarchy
    const h2Headings = screen.getAllByRole('heading', { level: 2 });
    const h3Headings = screen.getAllByRole('heading', { level: 3 });
    
    expect(h2Headings.length).toBeGreaterThan(0);
    expect(h3Headings.length).toBeGreaterThan(0);
  });

  it('should have proper link styling for external links', () => {
    renderWithRouter(<PrivacyPolicy />);
    
    // Check for external links with href containing strava.com
    const externalLinks = Array.from(document.querySelectorAll('a[href*="strava.com"]'));
    expect(externalLinks.length).toBeGreaterThan(0);
    
    // Verify external links have proper attributes
    externalLinks.forEach(link => {
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });
  });

  it('should have proper list styling', () => {
    renderWithRouter(<DataUsagePolicy />);
    
    // Check for lists
    const lists = screen.getAllByRole('list');
    expect(lists.length).toBeGreaterThan(0);
    
    // Check for list items
    const listItems = screen.getAllByRole('listitem');
    expect(listItems.length).toBeGreaterThan(0);
  });

  it('should have proper contact information styling', () => {
    renderWithRouter(<PrivacyPolicy />);
    
    // Check for email links
    const emailLinks = screen.getAllByRole('link', { name: /.*@bodda\.app/i });
    expect(emailLinks.length).toBeGreaterThan(0);
    
    // Verify email links have proper href
    emailLinks.forEach(link => {
      expect(link.getAttribute('href')).toMatch(/^mailto:/);
    });
  });
});