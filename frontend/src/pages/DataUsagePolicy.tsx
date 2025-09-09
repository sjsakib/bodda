import React from 'react';
import LegalPageLayout from '../components/LegalPageLayout';

const DataUsagePolicy: React.FC = () => {
  return (
    <LegalPageLayout 
      title="Data Usage Policy" 
      lastUpdated={new Date().toLocaleDateString()}
    >

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            1. Overview
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            This Data Usage Policy specifically explains how Bodda collects, processes, and uses data from Strava and other sources 
            to provide AI-powered coaching services. This policy supplements our Privacy Policy and Terms of Service.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            2. Strava Data Collection
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">2.1 Data Types Collected</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">
            When you connect your Strava account, we collect the following types of data through Strava's official API:
          </p>
          
          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Profile and Account Data:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Athlete profile information (name, bio, location, profile photo)</li>
            <li>Account preferences and privacy settings</li>
            <li>Follower/following relationships (if public)</li>
            <li>Club memberships and achievements</li>
            <li>Training zones and fitness metrics</li>
          </ul>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Activity Data:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Activity metadata (type, name, description, date, duration)</li>
            <li>Performance metrics (distance, speed, pace, elevation gain)</li>
            <li>Heart rate data and training zones</li>
            <li>Power data and cycling-specific metrics</li>
            <li>GPS coordinates and route information</li>
            <li>Activity streams (detailed time-series data)</li>
            <li>Segment efforts and personal records</li>
            <li>Kudos, comments, and social interactions (if public)</li>
          </ul>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Training and Performance Data:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Training calendar and planned workouts</li>
            <li>Fitness and freshness scores</li>
            <li>Equipment information (bikes, shoes, gear)</li>
            <li>Historical performance trends</li>
          </ul>

          <div className="bg-blue-50 border-l-4 border-blue-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              
            </div>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            3. Data Processing Purposes
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">3.1 AI Coaching Services</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">Your Strava data is processed to provide:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Performance Analysis:</strong> Analyzing trends, patterns, and progress in your training</li>
            <li><strong>Personalized Recommendations:</strong> Generating tailored training advice based on your data</li>
            <li><strong>Goal Setting:</strong> Helping establish realistic and achievable fitness goals</li>
            <li><strong>Training Load Management:</strong> Monitoring training stress and recovery needs</li>
            <li><strong>Comparative Analysis:</strong> Benchmarking against your historical performance</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">3.2 Service Improvement</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Algorithm Enhancement:</strong> Improving AI coaching accuracy (using anonymized data)</li>
            <li><strong>Feature Development:</strong> Creating new coaching features based on usage patterns</li>
            <li><strong>Quality Assurance:</strong> Ensuring data accuracy and service reliability</li>
            <li><strong>Performance Optimization:</strong> Optimizing system performance and response times</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">3.3 User Experience</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Dashboard Creation:</strong> Building personalized coaching dashboards</li>
            <li><strong>Progress Tracking:</strong> Visualizing training progress and achievements</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            4. Data Processing Flow and Storage
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">4.1 How We Handle Your Strava Data</h3>
          <div className="bg-green-50 border-l-4 border-green-400 p-6 mb-6 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-green-800">
                  <strong>Privacy-First Approach:</strong> Your raw Strava data never touches our storage systems. We process it in memory, generate insights, and only store those insights.
                </p>
              </div>
            </div>
          </div>

          <p className="mb-6 text-gray-800 leading-relaxed">Our data processing follows this privacy-preserving flow:</p>
          <ol className="list-decimal pl-6 mb-6 space-y-3 text-gray-800">
            <li><strong>Temporary Access:</strong> We fetch your Strava data using the API when you request coaching</li>
            <li><strong>In-Memory Processing:</strong> Data is processed in memory to generate insights and recommendations</li>
            <li><strong>Insight Generation:</strong> AI analyzes patterns and creates personalized coaching advice</li>
            <li><strong>Storage of Insights Only:</strong> We store the generated insights and recommendations, not the raw data</li>
            <li><strong>Data Disposal:</strong> Raw Strava data is immediately discarded from memory after processing</li>
          </ol>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">4.2 What Gets Stored vs. What Gets Processed</h3>
          <div className="grid md:grid-cols-2 gap-6 mb-8">
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <h4 className="font-semibold text-red-800 mb-3">❌ Never Stored</h4>
              <ul className="text-sm text-red-700 space-y-1">
                <li>• Raw activity files (.fit, .gpx, .tcx)</li>
                <li>• GPS coordinates and routes</li>
                <li>• Heart rate time series data</li>
                <li>• Power meter readings</li>
                <li>• Activity photos and descriptions</li>
                <li>• Strava social interactions</li>
              </ul>
            </div>
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <h4 className="font-semibold text-green-800 mb-3">✅ What We Store</h4>
              <ul className="text-sm text-green-700 space-y-1">
                <li>• AI-generated coaching insights</li>
                <li>• Training recommendations</li>
                <li>• Conversation history with AI coach</li>
                <li>• Derived performance metrics</li>
                <li>• Progress summaries and trends</li>
                <li>• Account preferences</li>
              </ul>
            </div>
          </div>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">4.3 Data Processing Methods</h3>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">4.1 Automated Processing</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">Most data processing is automated through:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>AI Analysis:</strong> Machine learning algorithms analyze patterns and trends</li>
            <li><strong>Statistical Computation:</strong> Automated calculation of performance metrics</li>
            <li><strong>Data Aggregation:</strong> Combining data points to create insights</li>
            <li><strong>Trend Detection:</strong> Identifying patterns in training and performance</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">4.2 Human Review</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">Limited human review occurs for:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Quality Assurance:</strong> Ensuring AI recommendations are appropriate</li>
            <li><strong>Customer Support:</strong> Resolving technical issues or data problems</li>
            <li><strong>Safety Monitoring:</strong> Identifying potentially harmful training patterns</li>
            <li><strong>Service Development:</strong> Understanding user needs and preferences</li>
          </ul>

          <div className="bg-yellow-50 border-l-4 border-yellow-400 p-6 mb-8 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-yellow-800">
                  <strong>Access Controls:</strong> Human access to your data is strictly limited to authorized personnel 
                  and is logged for audit purposes. Personal data is never accessed for non-business purposes.
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            5. Data Sharing and Third Parties
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">5.1 No Data Sales</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">
            <strong>We never sell your Strava data to third parties.</strong> Your fitness data is not a product for sale.
          </p>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">5.2 Service Providers</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">We share limited data with trusted service providers:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>AI Processing:</strong> OpenAI for natural language processing (coaching conversations only)</li>
            <li><strong>Cloud Infrastructure:</strong> Secure cloud hosting providers for data storage</li>
            <li><strong>Analytics Services:</strong> Anonymized usage analytics for service improvement</li>
            <li><strong>Security Services:</strong> Cybersecurity providers for threat detection and prevention</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">5.3 Data Processing Agreements</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">All service providers are bound by:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Strict confidentiality agreements</li>
            <li>Data minimization requirements</li>
            <li>Purpose limitation clauses</li>
            <li>Security and privacy standards</li>
            <li>Data deletion obligations</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">5.4 Legal Disclosures</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">We may disclose data only when:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Required by law or court order</li>
            <li>Necessary to protect our rights or safety</li>
            <li>Needed to prevent fraud or abuse</li>
            <li>Part of a business transfer (with continued privacy protection)</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            6. Data Storage and Retention
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">6.1 What We Store vs. What We Don't Store</h3>
          <div className="bg-blue-50 border-l-4 border-blue-400 p-6 mb-6 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-blue-800">
                  <strong>Important:</strong> We never store your raw Strava activity data directly. Instead, we process it temporarily to generate insights and coaching recommendations, which are then stored.
                </p>
              </div>
            </div>
          </div>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Data We Store:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>AI Coaching Insights:</strong> Generated recommendations and analysis based on your Strava data</li>
            <li><strong>Conversation History:</strong> Your chat interactions with the AI coach</li>
            <li><strong>Derived Analytics:</strong> Processed metrics and trends (not raw activity data)</li>
            <li><strong>Account Information:</strong> Email, preferences, and authentication tokens</li>
          </ul>

          <h4 className="text-lg font-semibold text-gray-800 mb-3 mt-5">Data We Don't Store:</h4>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Raw Strava Activities:</strong> Individual workout files, GPS tracks, or detailed activity streams</li>
            <li><strong>Personal Strava Content:</strong> Photos, comments, or social interactions from Strava</li>
            <li><strong>Strava Profile Data:</strong> Your Strava profile information beyond what's needed for authentication</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.2 Retention Periods</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Coaching Insights:</strong> Retained for 2 years or until account deletion</li>
            <li><strong>Conversation History:</strong> Retained for 1 year for coaching continuity</li>
            <li><strong>Account Information:</strong> Retained while account is active</li>

          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.2 Inactive Account Management</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>12 Months Inactive:</strong> Data review and cleanup of non-essential information</li>
            <li><strong>24 Months Inactive:</strong> Account marked for deletion with email notification</li>
            <li><strong>30 Months Inactive:</strong> Complete data deletion unless legally required to retain</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">6.3 Strava Token Management</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Token Expiration:</strong> Automatic cleanup when Strava tokens expire</li>
            <li><strong>Revoked Access:</strong> Immediate data collection cessation and cleanup within 7 days</li>
            <li><strong>Account Disconnection:</strong> Complete Strava data removal within 30 days</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            7. Data Security Measures
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">7.1 Technical Safeguards</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Encryption:</strong> Data encrypted in transit (TLS 1.3) and at rest (AES-256)</li>
            <li><strong>Access Controls:</strong> Role-based access with multi-factor authentication</li>
            <li><strong>Network Security:</strong> Firewalls, intrusion detection, and monitoring</li>
            <li><strong>Regular Audits:</strong> Security assessments and vulnerability testing</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">7.2 Operational Safeguards</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Staff Training:</strong> Regular privacy and security training for all personnel</li>
            <li><strong>Incident Response:</strong> Established procedures for security incidents</li>
            <li><strong>Data Minimization:</strong> Collecting and retaining only necessary data</li>
            <li><strong>Regular Backups:</strong> Secure, encrypted backups with tested recovery procedures</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            8. Your Data Rights and Controls
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">8.1 Access and Portability</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Data Export:</strong> Download all your data in JSON format</li>
            <li><strong>Activity Logs:</strong> View how your data has been processed</li>
            <li><strong>Usage Reports:</strong> See what data is being used for coaching</li>
            <li><strong>Processing History:</strong> Review AI analysis and recommendations</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">8.2 Control and Modification</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Selective Deletion:</strong> Delete specific activities or data types</li>
            <li><strong>Processing Restrictions:</strong> Limit how certain data is used</li>
            <li><strong>Consent Management:</strong> Modify or withdraw consent for specific uses</li>
            <li><strong>Privacy Settings:</strong> Adjust data sharing and processing preferences</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">8.3 Strava-Specific Controls</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>OAuth Management:</strong> View and revoke Strava API access</li>
            <li><strong>Sync Controls:</strong> Pause or resume Strava data synchronization</li>
            <li><strong>Privacy Override:</strong> Ensure Strava privacy settings are respected</li>
            <li><strong>Selective Sync:</strong> Choose which Strava data types to share</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            9. Compliance and Standards
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">9.1 Strava API Compliance</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Full compliance with Strava's API Terms of Service</li>
            <li>Respect for Strava's rate limits and usage guidelines</li>
            <li>Proper attribution and branding requirements</li>
            <li>Regular review of Strava policy updates</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">9.2 Privacy Regulations</h3>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>GDPR:</strong> Full compliance with European data protection requirements</li>
            <li><strong>CCPA:</strong> California Consumer Privacy Act compliance</li>
            <li><strong>Other Jurisdictions:</strong> Adherence to applicable local privacy laws</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            10. Contact and Requests
          </h2>
          
          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-6">10.1 Data Usage Questions</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">
            For questions about how your data is used:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Email:</strong> contact@sakib.dev</li>
            <li><strong>Subject Line:</strong> "Data Usage Inquiry"</li>
            <li><strong>Response Time:</strong> Within 72 hours</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">10.2 Data Rights Requests</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">
            To exercise your data rights:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Online:</strong> Use account settings for most requests</li>
            <li><strong>Email:</strong> contact@sakib.dev for complex requests</li>
            <li><strong>Processing Time:</strong> Most requests completed within 30 days</li>
          </ul>

          <h3 className="text-xl font-semibold text-gray-800 mb-4 mt-8">10.3 Strava-Related Issues</h3>
          <p className="mb-6 text-gray-800 leading-relaxed">
            For Strava integration problems:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li><strong>Technical Support:</strong> contact@sakib.dev</li>
            <li><strong>Strava Access Issues:</strong> Check your Strava account settings first</li>
          </ul>
        </section>

        <div className="bg-orange-50 border-l-4 border-orange-400 p-6 mt-12 rounded-r-lg">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-semibold mb-4 text-orange-800">Data Storage and Usage Summary</h3>
              <div className="text-sm text-orange-700 space-y-3">
                <p>
                  <strong>Strava Data:</strong> Your raw Strava data is never stored directly on our servers. We only access it temporarily to generate coaching insights.
                </p>
                <p>
                  <strong>What We Store:</strong> We store AI-generated coaching insights, conversation history, and derived analytics - not your original Strava activities.
                </p>
                <p>
                  <strong>No Sales:</strong> Your data is never sold to third parties or used for advertising.
                </p>
                <p>
                  <strong>Transparency:</strong> You can always see how your data is being used and processed.
                </p>
              </div>
            </div>
          </div>
        </div>
    </LegalPageLayout>
  );
};

export default DataUsagePolicy;