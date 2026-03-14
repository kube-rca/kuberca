import { NavLink } from 'react-router-dom';
import { useAnalysisMode } from '../hooks/useAnalysisMode';

const TOTAL_SEVERITIES = 2; // warning, critical

export const AnalysisModeIndicator: React.FC = () => {
  const { manualSeverities, loading, error } = useAnalysisMode();

  let dotColor: string;
  let label: string;

  if (loading || error) {
    dotColor = 'bg-slate-400 dark:bg-slate-500';
    label = loading ? 'Loading...' : 'Auto';
  } else if (manualSeverities.length === 0) {
    dotColor = 'bg-emerald-500';
    label = 'All Auto';
  } else if (manualSeverities.length >= TOTAL_SEVERITIES) {
    dotColor = 'bg-rose-500';
    label = 'All Manual';
  } else {
    dotColor = 'bg-amber-500';
    label = 'Mixed';
  }

  return (
    <NavLink
      to="/settings/analysis"
      className="flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md transition-colors text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200"
    >
      <span className={`w-2 h-2 rounded-full flex-shrink-0 ${dotColor}`} />
      <span>Analysis: {label}</span>
    </NavLink>
  );
};

export const AnalysisModeIndicatorCompact: React.FC = () => {
  const { manualSeverities, loading, error } = useAnalysisMode();

  let dotColor: string;
  let label: string;

  if (loading || error) {
    dotColor = 'bg-slate-400 dark:bg-slate-500';
    label = loading ? '...' : 'Auto';
  } else if (manualSeverities.length === 0) {
    dotColor = 'bg-emerald-500';
    label = 'Auto';
  } else if (manualSeverities.length >= TOTAL_SEVERITIES) {
    dotColor = 'bg-rose-500';
    label = 'Manual';
  } else {
    dotColor = 'bg-amber-500';
    label = 'Mixed';
  }

  return (
    <NavLink
      to="/settings/analysis"
      className="flex items-center gap-1.5 whitespace-nowrap px-2 py-2 text-xs font-medium rounded-md transition-colors text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
    >
      <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${dotColor}`} />
      <span>{label}</span>
    </NavLink>
  );
};
