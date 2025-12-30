import { RCAItem } from '../types';

interface RCATableProps {
  rcas: RCAItem[];
  onTitleClick: (incident_id: string) => void;
}

// [추가] 텍스트 포맷팅 함수 (예: "warning" -> "Warning")
const formatSeverity = (text: string) => {
  if (!text) return '';
  return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
};

// 날짜 포맷팅 함수
const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  resolved: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700',
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
          {rcas.map((rca) => {
            const isResolved = !!rca.resolved_at; 
            const effectiveSeverity = isResolved ? 'Resolved' : rca.severity;

            // [로직 추가] 스타일용 소문자와 표시용 대문자 변환
            const normalizedSeverity = effectiveSeverity.toLowerCase(); // "warning"
            const displaySeverity = formatSeverity(normalizedSeverity); // "Warning"

            return (
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

                {/* Title */}
                <td
                  className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm font-medium text-gray-700 dark:text-white cursor-pointer hover:underline hover:text-black dark:hover:text-gray-200 text-center"
                  onClick={() => onTitleClick(rca.incident_id)}
                >
                  {rca.alarm_title}
                </td>

                {/* Severity: 계산된 값 적용 */}
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                      severityStyles[normalizedSeverity] || severityStyles.info
                    }`}
                  >
                    {displaySeverity}
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
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default RCATable;