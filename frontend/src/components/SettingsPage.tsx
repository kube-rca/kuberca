import React from 'react';
import { NavLink } from 'react-router-dom';

const SettingsPage: React.FC = () => {
  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6 transition-colors duration-300">
      <h1 className="text-2xl font-semibold text-slate-800 dark:text-white mb-6">Settings</h1>

      <div className="space-y-4">
        <NavLink
          to="/settings/notification"
          className="block p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-cyan-600 dark:text-cyan-400 mb-1">Notification</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Enable or disable the entire alert processing pipeline.</div>
            </div>
            <span className="text-slate-400 dark:text-slate-500">&rarr;</span>
          </div>
        </NavLink>

        <NavLink
          to="/settings/webhooks"
          className="block p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-cyan-600 dark:text-cyan-400 mb-1">Webhook Management</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Configure outgoing webhooks for alerts and incidents.</div>
            </div>
            <span className="text-slate-400 dark:text-slate-500">&rarr;</span>
          </div>
        </NavLink>

        <NavLink
          to="/settings/flapping"
          className="block p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-cyan-600 dark:text-cyan-400 mb-1">Flapping Detection</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Configure alert flapping detection thresholds and windows.</div>
            </div>
            <span className="text-slate-400 dark:text-slate-500">&rarr;</span>
          </div>
        </NavLink>

        <NavLink
          to="/settings/ai"
          className="block p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-cyan-600 dark:text-cyan-400 mb-1">AI Provider</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Switch the AI provider and model used for RCA analysis.</div>
            </div>
            <span className="text-slate-400 dark:text-slate-500">&rarr;</span>
          </div>
        </NavLink>

        <NavLink
          to="/settings/analysis"
          className="block p-4 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-lg text-violet-600 dark:text-violet-400 mb-1">Analysis Mode</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">Configure auto/manual analysis mode for alerts.</div>
            </div>
            <span className="text-slate-400 dark:text-slate-500">&rarr;</span>
          </div>
        </NavLink>
      </div>
    </div>
  );
};

export default SettingsPage;
