import { Hexagon, LogOut, Sun, Moon } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';

interface HeaderProps {
  onLogout: () => void;
  connectionState?: string;
}

export const Header: React.FC<HeaderProps> = ({ onLogout, connectionState }) => {
  const { theme, toggleTheme } = useTheme();

  return (
    <header className="fixed top-0 left-0 w-full h-14 bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 flex items-center px-5 z-50 transition-colors duration-300">
      {/* Logo */}
      <div className="flex items-center gap-2.5">
        <Hexagon className="w-6 h-6 text-cyan-600 dark:text-cyan-400" strokeWidth={2} />
        <span className="text-base font-semibold font-mono tracking-wider text-slate-900 dark:text-slate-100">
          Kube-RCA
        </span>
      </div>

      {/* Right controls */}
      <div className="ml-auto flex items-center gap-3">
        {/* SSE Connection dot - will be wired in Tier 3 */}
        {connectionState && (
          <span
            role="status"
            aria-label={`SSE: ${connectionState}`}
            className={`h-2 w-2 rounded-full ${
              connectionState === 'connected'
                ? 'bg-emerald-400 shadow-[0_0_6px] shadow-emerald-400/50'
                : connectionState === 'connecting'
                  ? 'bg-amber-400 animate-pulse'
                  : 'bg-slate-400'
            }`}
            title={`SSE: ${connectionState}`}
          />
        )}

        <button
          type="button"
          onClick={onLogout}
          className="flex items-center gap-1.5 text-xs font-medium text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors px-2 py-1.5 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800"
        >
          <LogOut className="w-3.5 h-3.5" />
          Logout
        </button>

        <button
          onClick={toggleTheme}
          className="p-2 rounded-lg text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-cyan-500/50"
          aria-label="Toggle Dark Mode"
        >
          {theme === 'light' ? (
            <Sun className="w-5 h-5" />
          ) : (
            <Moon className="w-5 h-5" />
          )}
        </button>
      </div>
    </header>
  );
};
