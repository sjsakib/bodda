import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import LegalFooter from '../components/LegalFooter'

describe('LegalFooter', () => {
  it('should render legal navigation links', () => {
    render(
      <MemoryRouter>
        <LegalFooter />
      </MemoryRouter>
    )
    
    // Check for legal links
    expect(screen.getByRole('link', { name: /privacy policy/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /terms of service/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /data usage policy/i })).toBeInTheDocument()
  })

  it('should have correct href attributes for legal pages', () => {
    render(
      <MemoryRouter>
        <LegalFooter />
      </MemoryRouter>
    )
    
    const privacyLink = screen.getByRole('link', { name: /privacy policy/i })
    const termsLink = screen.getByRole('link', { name: /terms of service/i })
    const dataUsageLink = screen.getByRole('link', { name: /data usage policy/i })
    
    expect(privacyLink).toHaveAttribute('href', '/privacy')
    expect(termsLink).toHaveAttribute('href', '/terms')
    expect(dataUsageLink).toHaveAttribute('href', '/data-usage')
  })

  it('should include Strava attribution', () => {
    render(
      <MemoryRouter>
        <LegalFooter />
      </MemoryRouter>
    )
    
    // Check for Strava attribution (should appear multiple times in footer)
    expect(screen.getAllByText(/powered by strava/i).length).toBeGreaterThan(0)
  })
})