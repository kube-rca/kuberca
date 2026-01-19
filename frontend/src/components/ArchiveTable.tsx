import { RCAItem } from '../types';

interface ArchiveTableProps {
  rcas: RCAItem[]; // Mute된 인시던트 데이터 리스트
  onTitleClick: (incident_id: string) => void;
}

const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

// 스타일 매핑
const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
};

const statusStyles: Record<string, string> = {
  firing: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  resolved: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700',
};

function ArchiveTable({ rcas, onTitleClick }: ArchiveTableProps) {
  // 데이터가 없을 경우 처리
  if (!rcas || rcas.length === 0) {
    return (
      <div className="text-center py-10 text-gray-500 dark:text-gray-400 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800">
        숨겨진 인시던트가 없습니다.
      </div>
    );
  }

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
              Status
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-800">
          {rcas.map((rca) => {
            const rawSeverity = rca.severity;
            const status = rca.resolved_at ? 'resolved' : 'firing';

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

                {/* Title (클릭 시 상세 이동) */}
                <td
                  className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm font-medium text-gray-700 dark:text-white cursor-pointer hover:underline hover:text-black dark:hover:text-gray-200 text-center"
                  onClick={() => onTitleClick(rca.incident_id)}
                >
                  {rca.title}
                </td>

                {/* Severity */}
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                      severityStyles[rawSeverity] || severityStyles.info
                    }`}
                  >
                    {rawSeverity}
                  </span>
                </td>

                {/* Status */}
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                      statusStyles[status] || statusStyles.firing
                    }`}
                  >
                    {status}
                  </span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default ArchiveTable;