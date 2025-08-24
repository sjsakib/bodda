import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useApi'
import { apiClient } from '../services/api'
import { ErrorDisplay, LoadingSpinner } from './ErrorBoundary'

export default function LandingPage() {
  const [isConnecting, setIsConnecting] = useState(false)
  const [connectError, setConnectError] = useState<string | null>(null)
  const navigate = useNavigate()
  const { checkAuth, loading: authLoading, authenticated, initialized, error: authError, clearError } = useAuth()

  // Check if user is already authenticated on component mount
  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  // Redirect to chat if authenticated
  useEffect(() => {
    if (authenticated && initialized) {
      navigate('/chat')
    }
  }, [authenticated, initialized, navigate])

  const handleStravaConnect = async () => {
    setIsConnecting(true)
    setConnectError(null)
    
    try {
      // Redirect to Strava OAuth using API client
      apiClient.redirectToStravaAuth()
    } catch (err) {
      setConnectError('Failed to connect to Strava. Please try again.')
      setIsConnecting(false)
    }
  }

  // Show loading spinner while checking authentication
  if (authLoading && !initialized) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-center">
          <LoadingSpinner size="lg" className="mx-auto mb-4" />
          <p className="text-gray-600">Checking authentication...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-16">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-5xl md:text-6xl font-bold text-gray-900 mb-6">
            Bodda
          </h1>
          <p className="text-xl md:text-2xl text-gray-700 mb-8 max-w-3xl mx-auto leading-relaxed">
            Your AI-powered running and cycling coach that learns from your Strava data
          </p>
        </div>

        {/* Main Content */}
        <div className="max-w-4xl mx-auto">
          {/* Features Section */}
          <div className="grid md:grid-cols-3 gap-8 mb-16">
            <div className="text-center p-6 bg-white rounded-lg shadow-sm">
              <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Data-Driven Insights</h3>
              <p className="text-gray-600">Analyzes your Strava activities to provide personalized coaching recommendations</p>
            </div>

            <div className="text-center p-6 bg-white rounded-lg shadow-sm">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Interactive Coaching</h3>
              <p className="text-gray-600">Chat with your AI coach to get answers about training, recovery, and performance</p>
            </div>

            <div className="text-center p-6 bg-white rounded-lg shadow-sm">
              <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Continuous Learning</h3>
              <p className="text-gray-600">Maintains an evolving athlete profile that improves coaching over time</p>
            </div>
          </div>

          {/* CTA Section */}
          <div className="text-center bg-white rounded-xl shadow-lg p-8 mb-12">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Ready to elevate your training?
            </h2>
            <p className="text-lg text-gray-600 mb-8 max-w-2xl mx-auto">
              Connect your Strava account to start receiving personalized coaching insights from your AI coach.
            </p>
            
            <ErrorDisplay 
              error={connectError || authError} 
              onDismiss={() => {
                setConnectError(null)
                clearError()
              }}
              className="mb-6"
            />

            <button
              onClick={handleStravaConnect}
              disabled={isConnecting}
              className="inline-flex items-center px-8 py-4 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white font-bold text-lg rounded-lg transition-colors duration-200 shadow-lg hover:shadow-xl"
              data-testid="strava-connect-button"
            >
              {isConnecting ? (
                <>
                  <LoadingSpinner size="sm" className="mr-3 border-white border-t-orange-300" />
                  Connecting...
                </>
              ) : (
                <>
                  <svg className="w-6 h-6 mr-3" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M15.387 17.944l-2.089-4.116h-3.065L15.387 24l5.15-10.172h-3.066m-7.008-5.599l2.836 5.599h4.172L10.463 0l-7.008 13.828h4.172"/>
                  </svg>
                  Connect with Strava
                </>
              )}
            </button>
          </div>

          {/* Disclaimer */}
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6" data-testid="disclaimer">
            <div className="flex items-start">
              <svg className="w-6 h-6 text-yellow-600 mt-0.5 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <div>
                <h4 className="text-lg font-semibold text-yellow-800 mb-2">Important Disclaimer</h4>
                <p className="text-yellow-700 leading-relaxed">
                  Bodda provides AI-generated coaching advice based on your activity data. This information is for educational purposes only and should not replace professional medical or coaching advice. Always consult with qualified professionals before making significant changes to your training regimen. Use this advice at your own risk and listen to your body.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}