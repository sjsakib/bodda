import React from 'react';
import LegalPageLayout from '../components/LegalPageLayout';

const PrivacyPolicy: React.FC = () => {
  return (
    <LegalPageLayout 
      title="Privacy Policy" 
      lastUpdated={new Date().toLocaleDateString()}
    >

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            1. Information We Collect
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">
            1.1 Strava Data Collection
          </h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            When you connect your Strava account through OAuth authorization, we collect and process the following data from Strava's API:
          </p>
          
          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Profile Information:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Athlete profile (name, profile picture, bio, location)</li>
            <li>Account settings and privacy preferences</li>
            <li>Follower and following relationships (if public)</li>
            <li>Club memberships and achievements</li>
          </ul>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Activity Data:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Activity details (type, name, description, date/time)</li>
            <li>Performance metrics (distance, duration, elevation, speed, pace)</li>
            <li>Heart rate data and training zones</li>
            <li>Power data and cycling metrics</li>
            <li>GPS coordinates and route information</li>
            <li>Activity streams (time-series data for detailed analysis)</li>
            <li>Segment efforts and personal records</li>
            <li>Photos and activity comments (if public)</li>
          </ul>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Training Data:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Training zones (heart rate, power, pace)</li>
            <li>Fitness and freshness scores</li>
            <li>Training calendar and planned workouts</li>
            <li>Equipment information (bikes, shoes, etc.)</li>
          </ul>

    

          <div className="bg-green-50 border-l-4 border-green-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-green-800">
                  <strong>Data Storage Clarification:</strong> We never store your raw Strava activity data directly. We only access it temporarily to generate coaching insights, which are then stored along with your conversation history.
                </p>
              </div>
            </div>
          </div>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">
            1.2 What We Actually Store
          </h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We collect and store the following information:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Email address for account identification</li>
            <li>Account preferences and coaching settings</li>
            <li><strong>AI coaching insights and recommendations</strong> generated from your Strava data</li>
            <li><strong>Chat history and AI coaching conversations</strong></li>
            <li>Consent records and privacy preferences</li>
            <li>Usage analytics and feature interactions (anonymized)</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">
            1.3 What We Don't Store
          </h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            To protect your privacy, we explicitly do not store:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Raw Strava activity files or detailed workout data</li>
            <li>GPS coordinates or route information from your activities</li>
            <li>Strava photos, comments, or social interactions</li>
            <li>Personal Strava profile information beyond authentication needs</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">
            1.4 Technical Information
          </h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>IP address and device information</li>
            <li>Browser type and version</li>
            <li>Session information and authentication tokens</li>
            <li>API usage logs and error reports</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            2. How We Process and Use Your Information
          </h2>
          <p className="mb-6 text-gray-700 leading-relaxed">We process your Strava data temporarily to:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Generate AI coaching insights:</strong> Analyze patterns and create personalized recommendations</li>
            <li><strong>Provide coaching conversations:</strong> Enable AI-powered chat interactions about your training</li>
            <li><strong>Create derived analytics:</strong> Generate summary statistics and trends (stored, not raw data)</li>
            <li><strong>Improve service quality:</strong> Use anonymized patterns to enhance our algorithms</li>
          </ul>
          
          <p className="mb-6 text-gray-700 leading-relaxed">We store and use the following:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Coaching insights and recommendations:</strong> AI-generated advice based on your data</li>
            <li><strong>Conversation history:</strong> Your chat interactions with the AI coach</li>
            <li><strong>Account preferences:</strong> Your settings and coaching preferences</li>

          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            3. Strava Data Usage and Compliance
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">
            3.1 Strava API Terms Compliance
          </h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            Our application strictly adheres to Strava's API Terms of Service and Developer Guidelines:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>We respect Strava's rate limits (100 requests per 15 minutes per application)</li>
            <li>We comply with all Strava branding and attribution requirements</li>
            <li>We do not attempt to reverse engineer or circumvent Strava's systems</li>
            <li>We maintain current compliance with Strava's evolving terms and policies</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">3.2 Privacy Settings Respect</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We strictly honor your Strava privacy settings and preferences:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Private activities are never accessed, stored, or processed by our system</li>
            <li>Activity visibility settings are checked before any data processing</li>
            <li>Follower-only content is only accessible if you've granted appropriate permissions</li>
            <li>We respect segment privacy settings and hidden achievements</li>
            <li>Location data privacy preferences are strictly enforced</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">3.3 Data Usage Limitations</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            Your Strava data is used exclusively for the following purposes:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>AI Coaching:</strong> Analyzing performance trends and providing personalized training insights</li>
            <li><strong>Performance Analytics:</strong> Generating charts, statistics, and progress tracking</li>
            <li><strong>Training Recommendations:</strong> Creating workout suggestions based on your data</li>
            <li><strong>Service Improvement:</strong> Anonymized analysis to improve our coaching algorithms</li>
          </ul>

          <div className="bg-red-50 border-l-4 border-red-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h4 className="font-semibold text-red-800 mb-3">Prohibited Uses:</h4>
                <ul className="list-disc pl-6 text-sm text-red-700 space-y-1">
                  <li>Selling or monetizing your Strava data to third parties</li>
                  <li>Creating public leaderboards or competitions without explicit consent</li>
                  <li>Sharing individual activity details with other users</li>
                  <li>Using data for advertising or marketing to third parties</li>
                  <li>Attempting to identify or contact your Strava connections</li>
                </ul>
              </div>
            </div>
          </div>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">3.4 Data Sharing and Third Parties</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We do not sell, rent, or share your Strava data with third parties, except:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Service Providers:</strong> Trusted partners who help operate our service (under strict data protection agreements)</li>
            <li><strong>Legal Requirements:</strong> When required by law, court order, or to protect our rights</li>
            <li><strong>Business Transfers:</strong> In the event of a merger or acquisition (with continued privacy protection)</li>
            <li><strong>Your Consent:</strong> When you explicitly authorize sharing for specific purposes</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            4. Data Storage and Security
          </h2>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We implement appropriate security measures to protect your data:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Data is encrypted in transit and at rest</li>
            <li>Access to your data is limited to authorized personnel only</li>
            <li>We regularly review and update our security practices</li>
            <li>We comply with industry-standard security protocols</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            5. Data Retention and Lifecycle
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">5.1 Strava Data Retention</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We implement specific retention policies for different types of Strava data:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Activity Data:</strong> Retained for 2 years from last account activity or until deletion request</li>
            <li><strong>Profile Information:</strong> Retained while account is active, deleted within 30 days of account closure</li>
            <li><strong>Training Analytics:</strong> Aggregated insights retained for 1 year for service improvement</li>
            <li><strong>Chat History:</strong> Coaching conversations retained for 1 year or until deletion request</li>
            <li><strong>API Tokens:</strong> Automatically expire per Strava's token lifecycle, immediately revoked on disconnection</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">5.2 Automated Data Cleanup</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            Our system automatically manages data lifecycle:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Inactive accounts (no login for 24 months) trigger data review and potential deletion</li>
            <li>Expired Strava tokens result in immediate cessation of data collection</li>
            <li>Revoked Strava access triggers automatic data cleanup within 7 days</li>
            <li>System logs and analytics data are automatically purged after 90 days</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">5.3 Data Deletion Guarantees</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            When you request data deletion or close your account:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Personal data deletion completed within 30 days of request</li>
            <li>Strava tokens immediately revoked and API access terminated</li>
            <li>Cached activity data purged from all systems and backups</li>
            <li>Anonymized analytics may be retained for service improvement (no personal identifiers)</li>
            <li>Legal or security-related data may be retained as required by law</li>
          </ul>

          <div className="bg-blue-50 border-l-4 border-blue-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-blue-800">
                  <strong>Note:</strong> Data deletion is irreversible. Once deleted, we cannot recover your historical 
                  coaching data, analytics, or chat history. Consider exporting your data before deletion if you want to keep a copy.
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            6. Your Privacy Rights and Controls
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">6.1 Data Access and Portability</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">You have the right to:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Access:</strong> View all personal data we have collected about you</li>
            <li><strong>Export:</strong> Download your data in JSON format for portability</li>
            <li><strong>Audit:</strong> Review how your data has been used and processed</li>
            <li><strong>History:</strong> See your consent history and privacy setting changes</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.2 Data Control and Correction</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Correct:</strong> Update or correct inaccurate personal information</li>
            <li><strong>Restrict:</strong> Limit how we process your data for specific purposes</li>
            <li><strong>Object:</strong> Opt out of certain data processing activities</li>
            <li><strong>Withdraw Consent:</strong> Revoke previously granted permissions</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.3 Strava-Specific Rights</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Disconnect Strava:</strong> Revoke OAuth access and stop data collection immediately</li>
            <li><strong>Selective Sync:</strong> Choose which types of Strava data to share (if supported by Strava)</li>
            <li><strong>Privacy Override:</strong> Ensure Strava privacy settings are respected in real-time</li>
            <li><strong>Token Management:</strong> View and manage active Strava API tokens</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.4 Account Deletion</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            You can delete your account and all associated data at any time:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Immediate cessation of all data collection</li>
            <li>Strava token revocation and API access termination</li>
            <li>Complete data deletion within 30 days</li>
            <li>Confirmation email when deletion is complete</li>
            <li>Option to export data before deletion</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.5 Exercising Your Rights</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            To exercise any of these rights:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Use the account settings page for most data management tasks</li>
            <li>Contact us at contact@sakib.dev for complex requests</li>
            <li>We will respond to requests within 30 days</li>
            <li>Identity verification may be required for security</li>
          </ul>

          <div className="bg-green-50 border-l-4 border-green-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-green-800">
                  <strong>No Cost:</strong> Exercising your privacy rights is always free. We will never charge you 
                  for accessing, correcting, or deleting your personal data.
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            7. Third-Party Services and Integrations
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">7.1 Strava Integration</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            Our primary integration is with Strava, Inc.:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Data Source:</strong> All fitness data comes from Strava's API</li>
            <li><strong>Privacy Policy:</strong> Subject to <a href="https://www.strava.com/legal/privacy" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">Strava's Privacy Policy</a></li>
            <li><strong>Terms of Service:</strong> Governed by <a href="https://www.strava.com/legal/terms" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">Strava's Terms of Service</a></li>
            <li><strong>API Terms:</strong> We comply with <a href="https://developers.strava.com/docs/getting-started/#account" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">Strava's API Agreement</a></li>
            <li><strong>Data Control:</strong> You can manage Strava data sharing through your Strava account settings</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">7.2 AI Processing Services</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>OpenAI:</strong> For AI coaching and natural language processing (data processed according to <a href="https://openai.com/privacy/" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">OpenAI's Privacy Policy</a>)</li>
            <li><strong>Data Minimization:</strong> Only necessary data is sent to AI services</li>
            <li><strong>No Training:</strong> Your data is not used to train AI models</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">7.3 Infrastructure and Analytics</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Cloud Hosting:</strong> Secure cloud infrastructure for data storage and processing</li>
            <li><strong>Analytics:</strong> Anonymized usage analytics to improve service quality</li>
            <li><strong>Security Services:</strong> Third-party security monitoring and protection services</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">7.4 Data Processing Agreements</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            All third-party service providers are bound by:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>Strict data processing agreements</li>
            <li>Confidentiality and security requirements</li>
            <li>Data minimization and purpose limitation</li>
            <li>Compliance with applicable privacy laws</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            8. Consent and Legal Basis
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">8.1 Consent Collection</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We collect explicit consent for data processing:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Initial Consent:</strong> Clear consent modal before Strava OAuth authorization</li>
            <li><strong>Granular Consent:</strong> Separate consent for different data processing purposes</li>
            <li><strong>Informed Consent:</strong> Plain language explanations of data usage</li>
            <li><strong>Consent Records:</strong> Detailed logs of when and how consent was granted</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">8.2 Legal Basis for Processing</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We process your data based on:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Consent:</strong> For Strava data processing and AI coaching services</li>
            <li><strong>Contract Performance:</strong> To provide the coaching services you've requested</li>
            <li><strong>Legitimate Interest:</strong> For service improvement and security (with privacy safeguards)</li>
            <li><strong>Legal Obligation:</strong> For compliance with applicable laws and regulations</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">8.3 Consent Management</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>You can withdraw consent at any time through account settings</li>
            <li>Consent withdrawal stops future data processing but doesn't affect past lawful processing</li>
            <li>Some services may become unavailable if essential consent is withdrawn</li>
            <li>We maintain records of consent changes for compliance purposes</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            9. International Data Transfers
          </h2>
          <p className="mb-6 text-gray-700 leading-relaxed">
            Your data may be processed in countries other than your residence:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li>We ensure adequate protection through appropriate safeguards</li>
            <li>Data transfers comply with applicable privacy laws (GDPR, CCPA, etc.)</li>
            <li>Third-party processors are bound by equivalent privacy protections</li>
            <li>You can request information about specific transfer mechanisms</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            10. Changes to This Policy
          </h2>
          <p className="mb-6 text-gray-700 leading-relaxed">
            We may update this privacy policy from time to time:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Notification:</strong> Material changes will be communicated via email and in-app notifications</li>
            <li><strong>Effective Date:</strong> Changes take effect 30 days after notification unless you object</li>
            <li><strong>Version History:</strong> Previous versions are archived and available upon request</li>
            <li><strong>Continued Use:</strong> Continued use of the service after changes constitutes acceptance</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            11. Contact Us
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">Privacy Inquiries</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            For questions about this privacy policy or our data practices:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Email:</strong> contact@sakib.dev</li>
            <li><strong>Response Time:</strong> We respond to privacy inquiries within 72 hours</li>
            <li><strong>Data Protection Officer:</strong> contact@sakib.dev (if applicable)</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">Data Subject Requests</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            For data access, correction, or deletion requests:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>Online:</strong> Use your account settings for most requests</li>
            <li><strong>Email:</strong> contact@sakib.dev for complex requests</li>
            <li><strong>Processing Time:</strong> Most requests completed within 30 days</li>
            <li><strong>Verification:</strong> Identity verification may be required for security</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">Regulatory Complaints</h3>
          <p className="mb-6 text-gray-700 leading-relaxed">
            You have the right to lodge complaints with supervisory authorities:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-700">
            <li><strong>EU/EEA:</strong> Your local Data Protection Authority</li>
            <li><strong>UK:</strong> Information Commissioner's Office (ICO)</li>
            <li><strong>California:</strong> California Attorney General's Office</li>
            <li><strong>Other Jurisdictions:</strong> Relevant privacy regulatory bodies</li>
          </ul>

          
        </section>

        <div className="bg-orange-50 border-l-4 border-orange-400 p-6 mt-12 rounded-r-lg">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-semibold mb-4 text-orange-800">Strava Integration Notice</h3>
              <div className="text-sm text-orange-700 space-y-3">
                <p>
                  <strong>Powered by Strava:</strong> This application uses the Strava API to access your fitness data. 
                  All Strava data is subject to Strava's Terms of Service and Privacy Policy.
                </p>
                <p>
                  <strong>Dual Compliance:</strong> By connecting your Strava account, you acknowledge that you have read and agree to:
                </p>
                <ul className="list-disc pl-6 mt-2 space-y-1">
                  <li>This Privacy Policy (Bodda AI Coaching Platform)</li>
                  <li><a href="https://www.strava.com/legal/privacy" target="_blank" rel="noopener noreferrer" className="text-orange-600 hover:underline font-medium">Strava's Privacy Policy</a></li>
                  <li><a href="https://www.strava.com/legal/terms" target="_blank" rel="noopener noreferrer" className="text-orange-600 hover:underline font-medium">Strava's Terms of Service</a></li>
                </ul>
                <p className="mt-3">
                  <strong>Data Control:</strong> You maintain full control over your Strava data and can revoke access at any time 
                  through your Strava account settings or our application settings.
                </p>
              </div>
            </div>
          </div>
        </div>
    </LegalPageLayout>
  );
};

export default PrivacyPolicy;