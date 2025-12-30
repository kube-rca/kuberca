import { RCAItem } from '../types';

interface RCATableProps {
  rcas: RCAItem[];
  onTitleClick: (incident_id: string) => void;
}

// 날짜 포맷팅 함수
const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

const severityStyles: Record<string, string> = {
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
};

function RCATable({ rcas, onTitleClick }: RCATableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full border-collapse border border-gray-300 dark:border-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              ID
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Time
            </th>
            {/* Title 헤더: 가운데 정렬 */}
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Title
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Severity
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Edit
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-800">
          {rcas.map((rca) => (
            <tr key={rca.incident_id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
              {/* ID */}
              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center">
                {rca.incident_id}
              </td>

              {/* Time */}
              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center whitespace-nowrap leading-relaxed">
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {formatDate(rca.fired_at)}
                </div>
                <div className="text-gray-400 font-bold my-0.5">~</div>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {rca.resolved_at ? formatDate(rca.resolved_at) : 'Ongoing'}
                </div>
              </td>

              {/* Title: 가운데 정렬 (text-center) 적용 완료 */}
              <td
                className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm font-medium text-gray-700 dark:text-white cursor-pointer hover:underline hover:text-black dark:hover:text-gray-200 text-center"
                onClick={() => onTitleClick(rca.incident_id)}
              >
                {rca.alarm_title}
              </td>

              {/* Severity */}
              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                    severityStyles[rca.severity.toLowerCase()] || severityStyles.info
                  }`}
                >
                  {rca.severity}
                </span>
              </td>

              {/* Edit */}
              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                <button
                  className="px-3 py-1 text-xs font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-500 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 transition-colors"
                  onClick={(e) => {
                    e.stopPropagation();
                    console.log('Edit RCA:', rca.incident_id);
                  }}
                >
                  edit
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default RCATable;