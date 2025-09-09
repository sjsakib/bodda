import React from 'react';
import LegalPageLayout from '../components/LegalPageLayout';

const TermsOfService: React.FC = () => {
  return (
    <LegalPageLayout 
      title="Terms of Service" 
      lastUpdated={new Date().toLocaleDateString()}
    >

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            1. Acceptance of Terms
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            By accessing and using Bodda ("the Service"), you accept and agree to be bound by the terms and provision of this agreement. 
            If you do not agree to abide by the above, please do not use this service.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            2. Description of Service
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            Bodda is an AI-powered coaching platform that integrates with Strava to provide personalized training insights and recommendations. 
            The Service analyzes your fitness data to offer coaching advice, performance analysis, and training guidance.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            3. Strava Integration
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            Our Service integrates with Strava through their official API. By using our Service:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>You must have a valid Strava account</li>
            <li>You agree to Strava's Terms of Service and Privacy Policy</li>
            <li>You grant us permission to access your Strava data as specified during authorization</li>
            <li>You can revoke this access at any time through your account settings</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            4. User Responsibilities
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">You agree to:</p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Provide accurate and complete information</li>
            <li>Maintain the security of your account credentials</li>
            <li>Use the Service in compliance with all applicable laws</li>
            <li>Not attempt to reverse engineer or compromise the Service</li>
            <li>Not use the Service for any illegal or unauthorized purpose</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            5. AI Coaching Disclaimer
          </h2>
          <div className="bg-red-50 border-l-4 border-red-400 p-6 mb-6 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-red-800">
                  <strong>Important:</strong> The AI coaching provided by our Service is for informational purposes only and should not replace professional medical or coaching advice.
                </p>
              </div>
            </div>
          </div>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>Always consult with qualified professionals before making significant training changes</li>
            <li>The AI recommendations are based on data analysis and may not account for individual health conditions</li>
            <li>You use the coaching advice at your own risk</li>
            <li>We are not liable for any injuries or health issues resulting from following AI recommendations</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            6. Data Usage and Privacy
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            Your privacy is important to us. Please review our Privacy Policy to understand how we collect, use, and protect your data. 
            By using the Service, you consent to the data practices described in our Privacy Policy.
          </p>
          
          <div className="bg-blue-50 border-l-4 border-blue-400 p-6 mb-6 rounded-r-lg">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-blue-800">
                  <strong>Data Storage Clarification:</strong> We never store your raw Strava activity data. We only store AI-generated coaching insights and your conversation history with our AI coach.
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            7. Service Availability
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            We strive to maintain high service availability, but we do not guarantee uninterrupted access:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2 text-gray-800">
            <li>The Service may be temporarily unavailable for maintenance</li>
            <li>Third-party service dependencies (like Strava API) may affect availability</li>
            <li>We reserve the right to modify or discontinue features with notice</li>
          </ul>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            8. Intellectual Property
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            The Service and its original content, features, and functionality are owned by Bodda and are protected by international copyright, trademark, and other intellectual property laws.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            9. Limitation of Liability
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            In no event shall Bodda be liable for any indirect, incidental, special, consequential, or punitive damages, including without limitation, loss of profits, data, use, goodwill, or other intangible losses.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            10. Account Termination
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            We may terminate or suspend your account and access to the Service at our sole discretion, without prior notice, for conduct that we believe violates these Terms or is harmful to other users, us, or third parties.
          </p>
          <p className="mb-6 text-gray-800 leading-relaxed">
            You may terminate your account at any time by contacting us or using the account deletion feature in your settings.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            11. Changes to Terms
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            We reserve the right to modify these terms at any time. We will notify users of any material changes by posting the new terms on this page and updating the "Last updated" date.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            12. Governing Law
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            These Terms shall be interpreted and governed by the laws of [Your Jurisdiction], without regard to its conflict of law provisions.
          </p>
        </section>

        <section className="mb-10">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 pb-2 border-b border-gray-200">
            13. Contact Information
          </h2>
          <p className="mb-6 text-gray-800 leading-relaxed">
            If you have any questions about these Terms of Service, please contact us at:
          </p>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-6">
            <p className="text-sm text-gray-900">
              <strong>Email:</strong> contact@sakib.dev<br />
            </p>
          </div>
        </section>

        <div className="bg-orange-50 border-l-4 border-orange-400 p-6 mt-12 rounded-r-lg">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-semibold mb-3 text-orange-800">Third-Party Terms</h3>
              <p className="text-sm text-orange-700 mb-3">
                This service integrates with Strava and OpenAI. Your use of these integrations is also subject to:
              </p>
              <ul className="list-disc pl-6 text-sm text-orange-700 space-y-1">
                <li><a href="https://www.strava.com/legal/terms" className="text-orange-600 hover:underline font-medium" target="_blank" rel="noopener noreferrer">Strava Terms of Service</a></li>
                <li><a href="https://openai.com/terms" className="text-orange-600 hover:underline font-medium" target="_blank" rel="noopener noreferrer">OpenAI Terms of Use</a></li>
              </ul>
            </div>
          </div>
        </div>
    </LegalPageLayout>
  );
};

export default TermsOfService;