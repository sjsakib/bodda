import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import LandingPage from './components/LandingPage'
import ChatInterface from './components/ChatInterface'
import PrivacyPolicy from './pages/PrivacyPolicy'
import TermsOfService from './pages/TermsOfService'
import DataUsagePolicy from './pages/DataUsagePolicy'
import { ErrorBoundary } from './components/ErrorBoundary'
import { DiagramLibraryProvider } from './contexts/DiagramLibraryContext'

function App() {
  return (
    <ErrorBoundary>
      <DiagramLibraryProvider>
        <Router>
          <div className="min-h-screen bg-gray-50">
            <Routes>
              <Route path="/" element={<LandingPage />} />
              <Route path="/chat" element={<ChatInterface />} />
              <Route path="/chat/:sessionId" element={<ChatInterface />} />
              <Route path="/privacy" element={<PrivacyPolicy />} />
              <Route path="/terms" element={<TermsOfService />} />
              <Route path="/data-usage" element={<DataUsagePolicy />} />
            </Routes>
          </div>
        </Router>
      </DiagramLibraryProvider>
    </ErrorBoundary>
  )
}

export default App