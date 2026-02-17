import { NavLink } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Incident Dashboard', end: true },
  { to: '/alerts', label: 'Alert Dashboard' },
  { to: '/muted', label: 'Archived Incidents' },
  { to: '/settings', label: 'Settings' },
];

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `block px-3 py-2 text-sm font-medium rounded-md transition-colors ${
    isActive
      ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
  }`;

const mobileNavLinkClass = ({ isActive }: { isActive: boolean }) =>
  `whitespace-nowrap px-3 py-2 text-sm font-medium rounded-md transition-colors ${
    isActive
      ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
  }`;

export const Sidebar: React.FC = () => {
  return (
    <>
      <aside className="hidden md:block fixed top-16 left-0 w-64 h-[calc(100vh-4rem)] border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 z-40 p-4">
        <nav className="space-y-2">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className={navLinkClass}>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="md:hidden border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 py-2 overflow-x-auto">
        <nav className="flex items-center gap-2 min-w-max">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className={mobileNavLinkClass}>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </>
  );
};
