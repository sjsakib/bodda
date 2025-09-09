import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import PrivacyPolicy from '../pages/PrivacyPolicy'
import TermsOfService from '../pages/TermsOfService'
import DataUsagePolicy from '../pages/DataUsagePolicy'

describe('Legal Pages', () => {
  it('should render Privacy Policy page with correct content', () => {
    render(
      <MemoryRouter>
        <PrivacyPolicy />
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /privacy policy/i, level: 1 })).toBeInTheDocument()
    expect(screen.getByText(/information we collect/i)).toBeInTheDocument()
    expect(screen.getByText(/strava data collection/i)).toBeInTheDocument()
  })

  it('should render Terms of Service page with correct content', () => {
    render(
      <MemoryRouter>
        <TermsOfService />
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /terms of service/i, level: 1 })).toBeInTheDocument()
    expect(screen.getByText(/acceptance of terms/i)).toBeInTheDocument()
    expect(screen.getByText(/strava integration/i)).toBeInTheDocument()
  })

  it('should render Data Usage Policy page with correct content', () => {
    render(
      <MemoryRouter>
        <DataUsagePolicy />
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /data usage policy/i, level: 1 })).toBeInTheDocument()
    expect(screen.getByText(/strava data collection/i)).toBeInTheDocument()
    expect(screen.getByText(/data types collected/i)).toBeInTheDocument()
  })

  it('should have proper Strava compliance notices in all legal pages', () => {
    // Test Privacy Policy
    const { unmount: unmountPrivacy } = render(
      <MemoryRouter>
        <PrivacyPolicy />
      </MemoryRouter>
    )
    expect(screen.getByText(/powered by strava/i)).toBeInTheDocument()
    unmountPrivacy()
    
    // Test Terms of Service  
    const { unmount: unmountTerms } = render(
      <MemoryRouter>
        <TermsOfService />
      </MemoryRouter>
    )
    expect(screen.getByText(/strava terms of service/i)).toBeInTheDocument()
    unmountTerms()
    
    // Test Data Usage Policy
    render(
      <MemoryRouter>
        <DataUsagePolicy />
      </MemoryRouter>
    )
    expect(screen.getAllByText(/strava api/i).length).toBeGreaterThan(0)
  })
})