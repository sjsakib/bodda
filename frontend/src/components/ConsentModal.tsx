import React, { useState } from 'react';

interface ConsentModalProps {
  isOpen: boolean;
  onAccept: () => void;
  onDecline: () => void;
}

const ConsentModal: React.FC<ConsentModalProps> = ({ isOpen, onAccept, onDecline }) => {
  const [dataProcessingConsent, setDataProcessingConsent] = useState(false);
  const [stravaAccessConsent, setStravaAccessConsent] = useState(false);
  const [aiCoachingConsent, setAiCoachingConsent] = useState(false);

  if (!isOpen) return null;

  const allConsentsGiven = dataProcessingConsent && stravaAccessConsent && aiCoachingConsent;

  const handleAccept = () => {
    if (allConsentsGiven) {
      onAccept();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <h2 className="text-2xl font-bold mb-4">Welcome to Bodda</h2>
          <p className="text-gray-600 mb-6">
            Before we get started, we need your consent for a few things to provide you with the best AI coaching experience.
          </p>

          <div className="space-y-6">
            {/* Data Processing Consent */}
            <div className="border border-gray-200 rounded-lg p-4">
              <label className="flex items-start space-x-3">
                <input
                  type="checkbox"
                  checked={dataProcessingConsent}
                  onChange={(e) => setDataProcessingConsent(e.target.checked)}
                  className="mt-1 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  required
                />
                <div>
                  <h3 className="font-semibold text-gray-900">Data Processing</h3>
                  <p className="text-sm text-gray-600 mt-1">
                    I consent to Bodda processing my personal data to provide AI coaching services. 
                    This includes analyzing my fitness data and providing personalized recommendations.
                  </p>
                </div>
              </label>
            </div>

            {/* Strava Access Consent */}
            <div className="border border-gray-200 rounded-lg p-4">
              <label className="flex items-start space-x-3">
                <input
                  type="checkbox"
                  checked={stravaAccessConsent}
                  onChange={(e) => setStravaAccessConsent(e.target.checked)}
                  className="mt-1 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  required
                />
                <div>
                  <h3 className="font-semibold text-gray-900">Strava Data Access</h3>
                  <p className="text-sm text-gray-600 mt-1">
                    I consent to Bodda accessing my Strava activity data to provide coaching insights. 
                    This includes activities, performance metrics, and training data.
                  </p>
                  <p className="text-xs text-gray-500 mt-2">
                    Your Strava privacy settings will be respected. Private activities remain private.
                  </p>
                </div>
              </label>
            </div>

            {/* AI Coaching Consent */}
            <div className="border border-gray-200 rounded-lg p-4">
              <label className="flex items-start space-x-3">
                <input
                  type="checkbox"
                  checked={aiCoachingConsent}
                  onChange={(e) => setAiCoachingConsent(e.target.checked)}
                  className="mt-1 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  required
                />
                <div>
                  <h3 className="font-semibold text-gray-900">AI Coaching Services</h3>
                  <p className="text-sm text-gray-600 mt-1">
                    I consent to using AI-powered coaching recommendations based on my data analysis.
                  </p>
                  <p className="text-xs text-gray-500 mt-2">
                    <strong>Important:</strong> AI coaching is for informational purposes only and should not replace professional medical advice.
                  </p>
                </div>
              </label>
            </div>
          </div>

          {/* Legal Links */}
          <div className="mt-6 p-4 bg-gray-50 rounded-lg">
            <p className="text-sm text-gray-600 mb-2">
              By proceeding, you agree to our:
            </p>
            <div className="flex flex-wrap gap-4 text-sm">
              <a 
                href="/legal/privacy-policy" 
                target="_blank"
                className="text-blue-600 hover:underline"
              >
                Privacy Policy
              </a>
              <a 
                href="/legal/terms-of-service" 
                target="_blank"
                className="text-blue-600 hover:underline"
              >
                Terms of Service
              </a>
              <a 
                href="https://www.strava.com/legal/terms" 
                target="_blank"
                className="text-blue-600 hover:underline"
              >
                Strava Terms
              </a>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end space-x-4 mt-6">
            <button
              onClick={onDecline}
              className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
            >
              I don't agree
            </button>
            <button
              onClick={handleAccept}
              disabled={!allConsentsGiven}
              className={`px-6 py-2 rounded-lg font-medium transition-colors ${
                allConsentsGiven
                  ? 'bg-blue-600 text-white hover:bg-blue-700'
                  : 'bg-gray-300 text-gray-500 cursor-not-allowed'
              }`}
            >
              Continue to Strava
            </button>
          </div>

          {!allConsentsGiven && (
            <p className="text-sm text-red-600 mt-2 text-center">
              Please accept all required consents to continue
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

export default ConsentModal;