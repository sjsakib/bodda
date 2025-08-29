import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import LandingPage from './components/LandingPage'
import ChatInterface from './components/ChatInterface'
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
            </Routes>
          </div>
        </Router>
      </DiagramLibraryProvider>
    </ErrorBoundary>
  )
}

export default App