import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Bell, Archive, BarChart3, Settings } from 'lucide-react';

const navItems = [
  { to: '/', label: 'Incidents', icon: LayoutDashboard, end: true },
  { to: '/alerts', label: 'Alerts', icon: Bell },
  { to: '/muted', label: 'Archived', icon: Archive },
  { to: '/analysis', label: 'Analysis', icon: BarChart3 },
];

const settingsItem = { to: '/settings', label: 'Settings', icon: Settings, end: undefined as boolean | undefined };

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
    isActive
      ? 'bg-cyan-50 text-cyan-700 dark:bg-cyan-950/50 dark:text-cyan-300 border-l-2 border-cyan-500 -ml-px'
      : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200'
  }`;

const mobileNavLinkClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-2 whitespace-nowrap px-3 py-2 text-sm font-medium rounded-md transition-colors ${
    isActive
      ? 'bg-cyan-50 text-cyan-700 dark:bg-cyan-950/50 dark:text-cyan-300'
      : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
  }`;

export const Sidebar: React.FC = () => {
  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden md:flex md:flex-col fixed top-14 left-0 w-60 h-[calc(100vh-3.5rem)] border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 z-40 p-4 transition-colors duration-300">
        <nav className="space-y-1 flex-1">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className={navLinkClass}>
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="border-t border-slate-200 dark:border-slate-800 pt-3 mt-3">
          <NavLink to={settingsItem.to} className={navLinkClass}>
            <settingsItem.icon className="w-4 h-4 flex-shrink-0" />
            {settingsItem.label}
          </NavLink>
        </div>
      </aside>

      {/* Mobile nav */}
      <div className="md:hidden border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-4 py-2 overflow-x-auto transition-colors duration-300">
        <nav className="flex items-center gap-2 min-w-max">
          {[...navItems, settingsItem].map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className={mobileNavLinkClass}>
              <item.icon className="w-4 h-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </>
  );
};
