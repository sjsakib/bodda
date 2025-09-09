import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import PrivacyPolicy from '../pages/PrivacyPolicy'
import TermsOfService from '../pages/TermsOfService'
import DataUsagePolicy from '../pages/DataUsagePolicy'

describe('Legal Page Routing', () => {
  it('should render Privacy Policy when navigating to /privacy', () => {
    render(
      <MemoryRouter initialEntries={['/privacy']}>
        <Routes>
          <Route path="/privacy" element={<PrivacyPolicy />} />
        </Routes>
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /privacy policy/i, level: 1 })).toBeInTheDocument()
  })

  it('should render Terms of Service when navigating to /terms', () => {
    render(
      <MemoryRouter initialEntries={['/terms']}>
        <Routes>
          <Route path="/terms" element={<TermsOfService />} />
        </Routes>
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /terms of service/i, level: 1 })).toBeInTheDocument()
  })

  it('should render Data Usage Policy when navigating to /data-usage', () => {
    render(
      <MemoryRouter initialEntries={['/data-usage']}>
        <Routes>
          <Route path="/data-usage" element={<DataUsagePolicy />} />
        </Routes>
      </MemoryRouter>
    )
    
    expect(screen.getByRole('heading', { name: /data usage policy/i, level: 1 })).toBeInTheDocument()
  })
})