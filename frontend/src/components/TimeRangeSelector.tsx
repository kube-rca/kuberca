import { TIME_RANGES } from '../constants';

interface TimeRangeSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

function TimeRangeSelector({ value, onChange }: TimeRangeSelectorProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="px-4 py-2 border border-slate-200 dark:border-slate-700 rounded-md shadow-sm bg-white dark:bg-slate-800 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500"
    >
      {TIME_RANGES.map((range) => (
        <option key={range} value={range}>
          {range}
        </option>
      ))}
    </select>
  );
}

export default TimeRangeSelector;

