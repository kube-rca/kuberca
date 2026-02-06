import { useTheme } from '../context/ThemeContext';
import { NavLink } from 'react-router-dom';

interface HeaderProps {
  onLogout: () => void;
}

export const Header: React.FC<HeaderProps> = ({ onLogout }) => {
  const { theme, toggleTheme } = useTheme();

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
      isActive
        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
    }`;

  return (
    <header className="fixed top-0 left-0 w-full h-16 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 flex items-center px-6 z-50 transition-colors duration-300">
      {/* 로고 영역 (좌측) */}
      <h1 className="text-xl font-bold text-gray-800 dark:text-white">
        Kube-RCA
      </h1>

      {/* 네비게이션 (중앙) */}
      <nav className="absolute left-1/2 transform -translate-x-1/2 flex items-center gap-2">
        <NavLink to="/" end className={navLinkClass}>
          Incident Dashboard
        </NavLink>
        <NavLink to="/alerts" className={navLinkClass}>
          Alert Dashboard
        </NavLink>
        <NavLink to="/muted" className={navLinkClass}>
          Archived Incidents
        </NavLink>
      </nav>

      {/* 우측 컨트롤 영역: 로그아웃 + 다크모드 토글 */}
      <div className="ml-auto flex items-center gap-4">
        <button
          type="button"
          onClick={onLogout}
          className="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
        >
          로그아웃
        </button>

        <button
          onClick={toggleTheme}
          className="p-2 rounded-full bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 transition-all duration-200 focus:outline-none ring-2 ring-transparent focus:ring-blue-500"
          aria-label="Toggle Dark Mode"
        >
          {theme === 'light' ? (
            /* 해 아이콘 (Sun) */
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
            </svg>
          ) : (
            /* 달 아이콘 (Moon) */
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
              <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z" />
            </svg>
          )}
        </button>
      </div>
    </header>
  );
};