import React, { useState, useEffect } from 'react';
import StravaAttribution from './StravaAttribution';

interface ConsentRecord {
  id: string;
  consent_type: string;
  granted: boolean;
  granted_at?: string;
  revoked_at?: string;
}

interface ComplianceSettingsProps {
  userId: string;
}

const ComplianceSettings: React.FC<ComplianceSettingsProps> = ({ userId }) => {
  const [consents, setConsents] = useState<ConsentRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [showDataExport, setShowDataExport] = useState(false);
  const [showAccountDeletion, setShowAccountDeletion] = useState(false);

  useEffect(() => {
    fetchConsents();
  }, [userId]);

  const fetchConsents = async () => {
    try {
      const response = await fetch(`/api/compliance/consents`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setConsents(data);
      }
    } catch (error) {
      console.error('Failed to fetch consents:', error);
    } finally {
      setLoading(false);
    }
  };

  const updateConsent = async (consentType: string, granted: boolean) => {
    try {
      const response = await fetch('/api/compliance/consent', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          consent_type: consentType,
          granted,
        }),
      });

      if (response.ok) {
        await fetchConsents(); // Refresh consents
      }
    } catch (error) {
      console.error('Failed to update consent:', error);
    }
  };

  const exportData = async () => {
    try {
      const response = await fetch('/api/compliance/export', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
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
      }
    } catch (error) {
      console.error('Failed to export data:', error);
    }
  };

  const deleteAccount = async () => {
    if (
      !window.confirm(
        'Are you sure you want to delete your account? This action cannot be undone.'
      )
    ) {
      return;
    }

    try {
      const response = await fetch('/api/compliance/delete-account', {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        // Redirect to goodbye page or login
        window.location.href = '/account-deleted';
      }
    } catch (error) {
      console.error('Failed to delete account:', error);
    }
  };

  const revokeStravaAccess = async () => {
    try {
      const response = await fetch('/api/strava/revoke', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        await fetchConsents(); // Refresh consents
        alert('Strava access has been revoked successfully.');
      }
    } catch (error) {
      console.error('Failed to revoke Strava access:', error);
    }
  };

  const getConsentDisplayName = (type: string) => {
    const names: Record<string, string> = {
      data_processing: 'Data Processing',
      strava_access: 'Strava Data Access',
      ai_coaching: 'AI Coaching Services',
      marketing: 'Marketing Communications',
    };
    return names[type] || type;
  };

  const getConsentDescription = (type: string) => {
    const descriptions: Record<string, string> = {
      data_processing: 'Allow processing of your personal data to provide our services',
      strava_access: 'Access your Strava activity data for coaching insights',
      ai_coaching: 'Use AI to analyze your data and provide coaching recommendations',
      marketing: 'Receive updates about new features and training tips',
    };
    return descriptions[type] || '';
  };

  if (loading) {
    return (
      <div className='flex justify-center items-center p-8'>
        <div className='animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600'></div>
      </div>
    );
  }

  return (
    <div className='max-w-4xl mx-auto p-6'>
      <h2 className='text-2xl font-bold mb-6'>Privacy & Data Settings</h2>

      {/* Consent Management */}
      <div className='bg-white rounded-lg shadow-sm border border-gray-200 mb-6'>
        <div className='p-6'>
          <h3 className='text-lg font-semibold mb-4'>Data Usage Consents</h3>
          <p className='text-gray-600 mb-6'>
            Manage your consent for different types of data processing. You can revoke
            consent at any time.
          </p>

          <div className='space-y-4'>
            {consents.map(consent => (
              <div
                key={consent.id}
                className='flex items-start justify-between p-4 border border-gray-100 rounded-lg'
              >
                <div className='flex-1'>
                  <h4 className='font-medium text-gray-900'>
                    {getConsentDisplayName(consent.consent_type)}
                  </h4>
                  <p className='text-sm text-gray-600 mt-1'>
                    {getConsentDescription(consent.consent_type)}
                  </p>
                  {consent.granted && consent.granted_at && (
                    <p className='text-xs text-gray-500 mt-2'>
                      Granted on {new Date(consent.granted_at).toLocaleDateString()}
                    </p>
                  )}
                </div>

                <div className='ml-4'>
                  <label className='flex items-center'>
                    <input
                      type='checkbox'
                      checked={consent.granted}
                      onChange={e =>
                        updateConsent(consent.consent_type, e.target.checked)
                      }
                      className='rounded border-gray-300 text-blue-600 focus:ring-blue-500'
                    />
                    <span className='ml-2 text-sm text-gray-700'>
                      {consent.granted ? 'Granted' : 'Revoked'}
                    </span>
                  </label>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Strava Integration */}
      <div className='bg-white rounded-lg shadow-sm border border-gray-200 mb-6'>
        <div className='p-6'>
          <h3 className='text-lg font-semibold mb-4'>Strava Integration</h3>
          <p className='text-gray-600 mb-4'>
            Manage your Strava connection and data access permissions.
          </p>

          <button
            onClick={revokeStravaAccess}
            className='bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors'
          >
            Revoke Strava Access
          </button>
          
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
      <div className='bg-white rounded-lg shadow-sm border border-gray-200 mb-6'>
        <div className='p-6'>
          <h3 className='text-lg font-semibold mb-4'>Data Export</h3>
          <p className='text-gray-600 mb-4'>
            Download a copy of all your data stored in our system.
          </p>

          <button
            onClick={exportData}
            className='bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors'
          >
            Export My Data
          </button>
        </div>
      </div>

      {/* Account Deletion */}
      <div className='bg-white rounded-lg shadow-sm border border-red-200 mb-6'>
        <div className='p-6'>
          <h3 className='text-lg font-semibold mb-4 text-red-800'>Delete Account</h3>
          <p className='text-gray-600 mb-4'>
            Permanently delete your account and all associated data. This action cannot be
            undone.
          </p>

          <div className='bg-red-50 border border-red-200 rounded-lg p-4 mb-4'>
            <h4 className='font-medium text-red-800 mb-2'>What will be deleted:</h4>
            <ul className='text-sm text-red-700 list-disc list-inside space-y-1'>
              <li>Your account and profile information</li>
              <li>All chat history and AI coaching interactions</li>
              <li>Strava data cached in our system</li>
              <li>All consent and audit records</li>
            </ul>
          </div>

          <button
            onClick={() => setShowAccountDeletion(true)}
            className='bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors'
          >
            Delete My Account
          </button>

          {showAccountDeletion && (
            <div className='mt-4 p-4 border border-red-300 rounded-lg bg-red-50'>
              <p className='text-red-800 font-medium mb-3'>
                Are you absolutely sure? This will permanently delete your account.
              </p>
              <div className='flex space-x-3'>
                <button
                  onClick={deleteAccount}
                  className='bg-red-600 text-white px-4 py-2 rounded text-sm hover:bg-red-700'
                >
                  Yes, Delete My Account
                </button>
                <button
                  onClick={() => setShowAccountDeletion(false)}
                  className='bg-gray-300 text-gray-700 px-4 py-2 rounded text-sm hover:bg-gray-400'
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Legal Links */}
      <div className='bg-gray-50 rounded-lg p-6'>
        <h3 className='text-lg font-semibold mb-4'>Legal Information</h3>
        <div className='flex flex-wrap gap-4 text-sm'>
          <a href='/legal/privacy-policy' className='text-blue-600 hover:underline'>
            Privacy Policy
          </a>
          <a href='/legal/terms-of-service' className='text-blue-600 hover:underline'>
            Terms of Service
          </a>
          <a href='/legal/data-usage' className='text-blue-600 hover:underline'>
            Data Usage Policy
          </a>
        </div>
      </div>
    </div>
  );
};

export default ComplianceSettings;
