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
      className="px-4 py-2 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
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

