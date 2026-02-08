import React from 'react';
import { NavLink } from 'react-router-dom';

const SettingsPage: React.FC = () => {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <h1 className="text-2xl font-semibold text-gray-800 dark:text-white mb-6">Settings</h1>
      
      <div className="space-y-6">
        <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
          <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">General</h2>
          <p className="text-gray-600 dark:text-gray-400">Manage your general settings here.</p>
          {/* Add more general settings content here */}
        </div>

        <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
          <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">Integrations</h2>
          <div className="space-y-2">
            <NavLink 
              to="/settings/webhooks"
              className="block p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              <div className="font-medium text-blue-600 dark:text-blue-400">Webhook Management</div>
              <div className="text-sm text-gray-500 dark:text-gray-400">Configure outgoing webhooks for alerts and incidents.</div>
            </NavLink>
          </div>
        </div>

        <div>
          <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">Notifications</h2>
          <p className="text-gray-600 dark:text-gray-400">Configure your notification preferences.</p>
          {/* Add notification settings content here */}
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
