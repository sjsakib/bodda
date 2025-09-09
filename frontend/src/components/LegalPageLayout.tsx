import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeftIcon, DocumentTextIcon } from '@heroicons/react/24/outline';

interface LegalPageLayoutProps {
  title: string;
  lastUpdated?: string;
  children: React.ReactNode;
}

const LegalPageLayout: React.FC<LegalPageLayoutProps> = ({ 
  title, 
  lastUpdated, 
  children 
}) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* Header with navigation */}
      <header className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <Link 
              to="/" 
              className="inline-flex items-center text-gray-700 hover:text-gray-900 transition-colors duration-200 group"
            >
              <ArrowLeftIcon className="h-5 w-5 mr-2 group-hover:-translate-x-1 transition-transform duration-200" />
              <span className="font-medium">Back to Home</span>
            </Link>
            
            <div className="flex items-center text-sm text-gray-700">
              <DocumentTextIcon className="h-4 w-4 mr-2" />
              Legal Documentation
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 lg:py-12">
        {/* Page title and metadata */}
        <div className="mb-10 lg:mb-16">
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 sm:p-8">
            <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900 mb-4 leading-tight">
              {title}
            </h1>
            {lastUpdated && (
              <div className="flex items-center space-x-3">
                <div className="h-1 w-8 bg-blue-500 rounded-full"></div>
                <p className="text-sm text-gray-800 font-medium">
                  Last updated: <time dateTime={lastUpdated}>{lastUpdated}</time>
                </p>
              </div>
            )}
          </div>
        </div>

        {/* Content with enhanced styling */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="legal-content p-6 sm:p-8 lg:p-12">
            {children}
          </div>
        </div>

        
        {/* Legal navigation footer */}
        <div className="mt-12 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-6 sm:p-8 border-b border-gray-100">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Related Legal Documents</h3>
            <p className="text-gray-800">Explore our complete legal framework</p>
          </div>
          <div className="p-6 sm:p-8">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Link 
                to="/privacy" 
                className="group block p-6 bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl border border-blue-200 hover:border-blue-300 hover:shadow-md transition-all duration-200 transform hover:-translate-y-1"
              >
                <div className="flex items-center mb-3">
                  <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center mr-3">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                  </div>
                  <h4 className="font-semibold text-gray-900 group-hover:text-blue-700 transition-colors">Privacy Policy</h4>
                </div>
                <p className="text-sm text-gray-800 leading-relaxed">How we collect, use, and protect your personal data and Strava information</p>
              </Link>
              
              <Link 
                to="/terms" 
                className="group block p-6 bg-gradient-to-br from-green-50 to-green-100 rounded-xl border border-green-200 hover:border-green-300 hover:shadow-md transition-all duration-200 transform hover:-translate-y-1"
              >
                <div className="flex items-center mb-3">
                  <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center mr-3">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                  <h4 className="font-semibold text-gray-900 group-hover:text-green-700 transition-colors">Terms of Service</h4>
                </div>
                <p className="text-sm text-gray-800 leading-relaxed">Your rights, responsibilities, and our service agreements</p>
              </Link>
              
              <Link 
                to="/data-usage" 
                className="group block p-6 bg-gradient-to-br from-purple-50 to-purple-100 rounded-xl border border-purple-200 hover:border-purple-300 hover:shadow-md transition-all duration-200 transform hover:-translate-y-1"
              >
                <div className="flex items-center mb-3">
                  <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center mr-3">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                    </svg>
                  </div>
                  <h4 className="font-semibold text-gray-900 group-hover:text-purple-700 transition-colors">Data Usage Policy</h4>
                </div>
                <p className="text-sm text-gray-800 leading-relaxed">Detailed explanation of how your Strava data is processed</p>
              </Link>
            </div>
          </div>
        </div>

        {/* Contact information */}
        <div className="mt-8 bg-gradient-to-r from-blue-600 to-blue-700 rounded-xl shadow-lg overflow-hidden">
          <div className="p-6 sm:p-8">
            <div className="flex items-center mb-4">
              <div className="w-12 h-12 bg-white/20 rounded-lg flex items-center justify-center mr-4">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div>
                <h3 className="text-xl font-semibold text-white mb-1">Questions about our policies?</h3>
                <p className="text-blue-100">We're here to help clarify any concerns</p>
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
              <div className="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
                <p className="font-medium text-white mb-2">Privacy Inquiries</p>
                <a 
                  href="mailto:contact@sakib.dev" 
                  className="text-blue-100 hover:text-white transition-colors underline decoration-blue-300 hover:decoration-white"
                >
                  contact@sakib.dev
                </a>
              </div>
              <div className="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
                <p className="font-medium text-white mb-2">Legal Questions</p>
                <a 
                  href="mailto:contact@sakib.dev" 
                  className="text-blue-100 hover:text-white transition-colors underline decoration-blue-300 hover:decoration-white"
                >
                  contact@sakib.dev
                </a>
              </div>
              <div className="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
                <p className="font-medium text-white mb-2">General Support</p>
                <a 
                  href="mailto:contact@sakib.dev" 
                  className="text-blue-100 hover:text-white transition-colors underline decoration-blue-300 hover:decoration-white"
                >
                  contact@sakib.dev
                </a>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default LegalPageLayout;