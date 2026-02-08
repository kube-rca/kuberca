import React from 'react';
import { NavLink } from 'react-router-dom';

const SettingsPage: React.FC = () => {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <h1 className="text-2xl font-semibold text-gray-800 dark:text-white mb-6">Settings</h1>
      
      <div className="space-y-4">
        <NavLink 
          to="/settings/webhooks"
          className="block p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-blue-600 dark:text-blue-400 mb-1">Webhook Management</div>
              <div className="text-sm text-gray-500 dark:text-gray-400">Configure outgoing webhooks for alerts and incidents.</div>
            </div>
            <span className="text-gray-400 dark:text-gray-500">&rarr;</span>
          </div>
        </NavLink>
      </div>
    </div>
  );
};

export default SettingsPage;
