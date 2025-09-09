import React, { useState } from 'react';
import StravaAttribution from './StravaAttribution';

interface SimpleComplianceSettingsProps {
  userId: string;
}

const SimpleComplianceSettings: React.FC<SimpleComplianceSettingsProps> = ({ userId }) => {
  const [loading, setLoading] = useState(false);
  const [showAccountDeletion, setShowAccountDeletion] = useState(false);

  const exportData = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/user/export', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `bodda-data-export-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      } else {
        alert('Failed to export data. Please try again.');
      }
    } catch (error) {
      console.error('Failed to export data:', error);
      alert('Failed to export data. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const deleteAccount = async () => {
    if (!window.confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('/api/user/delete', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        // Clear local storage and redirect
        localStorage.clear();
        window.location.href = '/account-deleted';
      } else {
        alert('Failed to delete account. Please contact support.');
      }
    } catch (error) {
      console.error('Failed to delete account:', error);
      alert('Failed to delete account. Please contact support.');
    } finally {
      setLoading(false);
    }
  };

  const disconnectStrava = async () => {
    if (!window.confirm('Are you sure you want to disconnect your Strava account? You will need to reconnect to continue using Bodda.')) {
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('/api/strava/disconnect', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        alert('Strava account disconnected successfully. You will need to reconnect to continue using the service.');
        // Optionally redirect to reconnection flow
        window.location.href = '/connect-strava';
      } else {
        alert('Failed to disconnect Strava. Please try again.');
      }
    } catch (error) {
      console.error('Failed to disconnect Strava:', error);
      alert('Failed to disconnect Strava. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h2 className="text-2xl font-bold mb-6">Privacy & Account Settings</h2>

      {/* Current Status */}
      <div className="bg-green-50 border border-green-200 rounded-lg p-6 mb-6">
        <h3 className="text-lg font-semibold text-green-800 mb-2">✅ Your Account is Active</h3>
        <p className="text-green-700">
          You have given consent for data processing and Strava access. 
          Your data is being used only for AI coaching services as described in our privacy policy.
        </p>
      </div>

      {/* Strava Connection */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
        <div className="p-6">
          <h3 className="text-lg font-semibold mb-4">Strava Connection</h3>
          <p className="text-gray-600 mb-4">
            Your Strava account is connected and we're accessing your activity data for coaching insights.
          </p>
          
          <button
            onClick={disconnectStrava}
            disabled={loading}
            className="bg-orange-600 text-white px-4 py-2 rounded-lg hover:bg-orange-700 transition-colors disabled:opacity-50"
          >
            {loading ? 'Disconnecting...' : 'Disconnect Strava'}
          </button>
          
          <p className="text-sm text-gray-500 mt-2">
            Note: Disconnecting will stop all AI coaching services until you reconnect.
          </p>
          
          <div className="mt-4 pt-4 border-t border-gray-200 flex justify-center">
            <StravaAttribution 
              dataType="general" 
              variant="footer" 
              size="small"
              theme="light"
            />
          </div>
        </div>
      </div>

      {/* Data Export */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
        <div className="p-6">
          <h3 className="text-lg font-semibold mb-4">Export Your Data</h3>
          <p className="text-gray-600 mb-4">
            Download a copy of all your data stored in our system, including your profile, chat history, and cached Strava data.
          </p>
          
          <button
            onClick={exportData}
            disabled={loading}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
          >
            {loading ? 'Preparing Export...' : 'Export My Data'}
          </button>
        </div>
      </div>

      {/* Account Deletion */}
      <div className="bg-white rounded-lg shadow-sm border border-red-200 mb-6">
        <div className="p-6">
          <h3 className="text-lg font-semibold mb-4 text-red-800">Delete Account</h3>
          <p className="text-gray-600 mb-4">
            Permanently delete your account and all associated data. This action cannot be undone.
          </p>
          
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
            <h4 className="font-medium text-red-800 mb-2">What will be deleted:</h4>
            <ul className="text-sm text-red-700 list-disc list-inside space-y-1">
              <li>Your account and profile information</li>
              <li>All chat history and AI coaching interactions</li>
              <li>Cached Strava data in our system</li>
              <li>All session and activity analysis data</li>
            </ul>
          </div>
          
          <button
            onClick={() => setShowAccountDeletion(true)}
            disabled={loading}
            className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
          >
            Delete My Account
          </button>
          
          {showAccountDeletion && (
            <div className="mt-4 p-4 border border-red-300 rounded-lg bg-red-50">
              <p className="text-red-800 font-medium mb-3">
                ⚠️ This will permanently delete your account and all data. Are you absolutely sure?
              </p>
              <div className="flex space-x-3">
                <button
                  onClick={deleteAccount}
                  disabled={loading}
                  className="bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700 disabled:opacity-50"
                >
                  {loading ? 'Deleting...' : 'Yes, Delete Everything'}
                </button>
                <button
                  onClick={() => setShowAccountDeletion(false)}
                  className="bg-gray-300 text-gray-700 px-4 py-2 rounded text-sm hover:bg-gray-400"
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Privacy Information */}
      <div className="bg-gray-50 rounded-lg p-6">
        <h3 className="text-lg font-semibold mb-4">Privacy Information</h3>
        <p className="text-gray-600 mb-4">
          Your privacy is important to us. We only use your data for AI coaching services and never sell or share it with third parties.
        </p>
        
        <div className="flex flex-wrap gap-4 text-sm">
          <a href="/legal/privacy-policy" className="text-blue-600 hover:underline">
            Privacy Policy
          </a>
          <a href="/legal/terms-of-service" className="text-blue-600 hover:underline">
            Terms of Service
          </a>
          <a href="mailto:contact@sakib.dev" className="text-blue-600 hover:underline">
            Privacy Questions
          </a>
          <a href="mailto:contact@sakib.dev" className="text-blue-600 hover:underline">
            Contact Support
          </a>
        </div>
      </div>
    </div>
  );
};

export default SimpleComplianceSettings;