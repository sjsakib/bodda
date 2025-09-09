import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import LandingPage from '../LandingPage';

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock window.location
const mockLocation = {
  href: '',
};
Object.defineProperty(window, 'location', {
  value: mockLocation,
  writable: true,
});

const renderLandingPage = async () => {
  // Mock authentication check to return unauthenticated by default
  mockFetch.mockResolvedValueOnce({
    ok: false,
    status: 401,
    json: () => Promise.resolve({ error: 'Not authenticated' }),
  });

  const result = render(
    <BrowserRouter>
      <LandingPage />
    </BrowserRouter>
  );

  // Wait for auth check to complete
  await waitFor(() => {
    expect(screen.queryByText('Checking authentication...')).not.toBeInTheDocument();
  });

  return result;
};

describe('LandingPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocation.href = '';
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders the main heading and description', async () => {
    await renderLandingPage();

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Bodda');
    expect(
      screen.getByText(/your ai-powered running and cycling coach/i)
    ).toBeInTheDocument();
  });

  it('displays the Strava connect button', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toBeInTheDocument();
    expect(connectButton.querySelector('img')).toHaveAttribute('alt', 'Connect with Strava');
  });

  it('displays the disclaimer section', async () => {
    await renderLandingPage();

    const disclaimer = screen.getByTestId('disclaimer');
    expect(disclaimer).toBeInTheDocument();
    expect(disclaimer).toHaveTextContent(/important disclaimer/i);
    expect(disclaimer).toHaveTextContent(/use this advice at your own risk/i);
  });

  it('displays feature cards', async () => {
    await renderLandingPage();

    expect(screen.getByText(/data-driven insights/i)).toBeInTheDocument();
    expect(screen.getByText(/interactive coaching/i)).toBeInTheDocument();
    expect(screen.getByText(/continuous learning/i)).toBeInTheDocument();
  });

  it('redirects to Strava OAuth when connect button is clicked with consent', async () => {
    await renderLandingPage();

    // First check the consent checkbox
    const consentCheckbox = screen.getByTestId('consent-checkbox');
    fireEvent.click(consentCheckbox);

    const connectButton = screen.getByTestId('strava-connect-button');
    fireEvent.click(connectButton);

    await waitFor(() => {
      expect(mockLocation.href).toBe('/auth/strava');
    });
  });

  it('shows loading state when connecting to Strava', async () => {
    await renderLandingPage();

    // First check the consent checkbox
    const consentCheckbox = screen.getByTestId('consent-checkbox');
    fireEvent.click(consentCheckbox);

    const connectButton = screen.getByTestId('strava-connect-button');
    fireEvent.click(connectButton);

    expect(connectButton).toHaveTextContent(/connecting.../i);
    expect(connectButton).toBeDisabled();
  });

  it('redirects to chat if user is already authenticated', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () =>
        Promise.resolve({
          authenticated: true,
          user: { id: 'user-1' },
        }),
    });

    render(
      <BrowserRouter>
        <LandingPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/chat');
    });
  });

  it('stays on landing page if authentication check fails', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    await renderLandingPage();

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('displays error message when Strava connection fails', async () => {
    // Mock a connection error
    const originalLocation = window.location;
    const mockLocationWithError = {
      get href() {
        return '';
      },
      set href(_url: string) {
        throw new Error('Connection failed');
      },
    };
    Object.defineProperty(window, 'location', {
      value: mockLocationWithError,
      writable: true,
    });

    await renderLandingPage();

    // First check the consent checkbox
    const consentCheckbox = screen.getByTestId('consent-checkbox');
    fireEvent.click(consentCheckbox);

    const connectButton = screen.getByTestId('strava-connect-button');
    fireEvent.click(connectButton);

    await waitFor(() => {
      expect(screen.getByText(/failed to connect to strava/i)).toBeInTheDocument();
    });

    // Restore original location
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
    });
  });

  it('has proper accessibility attributes', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toBeInTheDocument();

    // Check that headings are properly structured
    const mainHeading = screen.getByRole('heading', { level: 1 });
    expect(mainHeading).toBeInTheDocument();
  });

  it('displays consent checkbox and links to legal pages', async () => {
    await renderLandingPage();

    const consentCheckbox = screen.getByTestId('consent-checkbox');
    expect(consentCheckbox).toBeInTheDocument();
    expect(consentCheckbox).not.toBeChecked();

    // Find the privacy policy link within the consent section
    const consentLabel = consentCheckbox.closest('div');
    const privacyLink = within(consentLabel!).getByRole('link', { name: /privacy policy/i });
    expect(privacyLink).toBeInTheDocument();
    expect(privacyLink).toHaveAttribute('href', '/privacy');

    const termsLink = within(consentLabel!).getByRole('link', { name: /terms of service/i });
    expect(termsLink).toBeInTheDocument();
    expect(termsLink).toHaveAttribute('href', '/terms');
  });

  it('disables connect button when consent is not accepted', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toBeDisabled();
    expect(connectButton).toHaveClass('opacity-50', 'cursor-not-allowed');
  });

  it('enables connect button when consent is accepted', async () => {
    await renderLandingPage();

    const consentCheckbox = screen.getByTestId('consent-checkbox');
    const connectButton = screen.getByTestId('strava-connect-button');

    // Initially disabled
    expect(connectButton).toBeDisabled();

    // Enable after checking consent
    fireEvent.click(consentCheckbox);
    expect(connectButton).not.toBeDisabled();
    expect(connectButton).not.toHaveClass('opacity-50', 'cursor-not-allowed');
  });

  it('prevents connection without consent through disabled button', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    
    // The button should be disabled when consent is not given
    expect(connectButton).toBeDisabled();
    
    // Verify the button has the disabled styling
    expect(connectButton).toHaveClass('opacity-50', 'cursor-not-allowed');
  });

  it('is responsive and includes proper CSS classes', async () => {
    await renderLandingPage();

    const mainHeading = screen.getByRole('heading', { level: 1 });
    const container = mainHeading.closest('div');
    expect(container?.parentElement?.parentElement).toHaveClass('min-h-screen');

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toHaveClass('shadow-lg', 'hover:shadow-xl');
  });
});
