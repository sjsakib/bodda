import { render, screen, fireEvent, waitFor } from '@testing-library/react';
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

    expect(screen.getByRole('heading', { name: /bodda/i })).toBeInTheDocument();
    expect(
      screen.getByText(/your ai-powered running and cycling coach/i)
    ).toBeInTheDocument();
  });

  it('displays the Strava connect button', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toBeInTheDocument();
    expect(connectButton).toHaveTextContent(/connect with strava/i);
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

  it('redirects to Strava OAuth when connect button is clicked', async () => {
    await renderLandingPage();

    const connectButton = screen.getByTestId('strava-connect-button');
    fireEvent.click(connectButton);

    await waitFor(() => {
      expect(mockLocation.href).toBe('/auth/strava');
    });
  });

  it('shows loading state when connecting to Strava', async () => {
    await renderLandingPage();

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
    const mainHeading = screen.getByRole('heading', { name: /bodda/i });
    expect(mainHeading).toBeInTheDocument();
  });

  it('is responsive and includes proper CSS classes', async () => {
    await renderLandingPage();

    const mainHeading = screen.getByRole('heading', { name: /^bodda$/i });
    const container = mainHeading.closest('div');
    expect(container?.parentElement?.parentElement).toHaveClass('min-h-screen');

    const connectButton = screen.getByTestId('strava-connect-button');
    expect(connectButton).toHaveClass('bg-orange-500', 'hover:bg-orange-600');
  });
});
