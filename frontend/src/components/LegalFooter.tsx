import React from 'react';
import { Link } from 'react-router-dom';
import StravaAttribution from './StravaAttribution';

const LegalFooter: React.FC = () => {
  return (
    <footer className='bg-gray-50 border-t border-gray-200 mt-auto'>
      <div className='max-w-7xl mx-auto px-4 py-8'>
        <div className='grid grid-cols-1 md:grid-cols-3 gap-8'>
          {/* Company Info */}
          <div>
            <h3 className='text-lg font-semibold text-gray-900 mb-4'>Bodda</h3>
            <p className='text-gray-600 text-sm mb-4'>
              AI-powered coaching platform for athletes. Get personalized insights from
              your training data.
            </p>
            <StravaAttribution size='medium' variant='footer' />
          </div>

          {/* Legal Links */} 
          <div>
            <h3 className='text-lg font-semibold text-gray-900 mb-4'>Legal</h3>
            <ul className='space-y-2'>
              <li>
                <Link
                  to='/privacy'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Privacy Policy
                </Link>
              </li>
              <li>
                <Link
                  to='/terms'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Terms of Service
                </Link>
              </li>
              <li>
                <Link
                  to='/data-usage'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Data Usage Policy
                </Link>
              </li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h3 className='text-lg font-semibold text-gray-900 mb-4'>Support</h3>
            <ul className='space-y-2'>
              <li>
                <a
                  href='mailto:contact@sakib.dev'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Contact Support
                </a>
              </li>
              <li>
                <a
                  href='mailto:contact@sakib.dev'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Privacy Questions
                </a>
              </li>
              {/* <li>
                <Link
                  to='/account/settings'
                  className='text-gray-600 hover:text-gray-900 text-sm transition-colors'
                >
                  Account Settings
                </Link>
              </li> */}
            </ul>
          </div>
        </div>

        <div className='border-t border-gray-200 mt-8 pt-8'>
          <div className='flex flex-col md:flex-row justify-between items-center'>
            <p className='text-gray-500 text-sm'>
              © {new Date().getFullYear()} Bodda. All rights reserved.
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default LegalFooter;
