import { RCAItem } from '../types';

interface RCATableProps {
  rcas: RCAItem[]; // 부모(App.tsx)가 이미 필터링을 끝낸 데이터
  onTitleClick: (incident_id: string) => void;
}

// 날짜 포맷팅 함수
const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

// 스타일 매핑
const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
  TBD: 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600',
};

const statusStyles: Record<string, string> = {
  firing: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  resolved: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700',
};

function RCATable({ rcas, onTitleClick }: RCATableProps) {
  // [삭제] 여기서 searchIncidents를 호출하면 안 됩니다!
  // App.tsx에서 이미 필터링된 결과가 'rcas'로 들어옵니다.

  if (!rcas || rcas.length === 0) {
    return (
      <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <p className="text-gray-500 dark:text-gray-400">데이터가 없습니다.</p>
        <p className="text-sm text-gray-400 mt-1">검색 조건에 맞는 결과가 없습니다.</p>
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
          {/* 그냥 받은 rcas를 그대로 그립니다 */}
          {rcas.map((rca) => {
            const rawSeverity = rca.severity;
            const status = rca.resolved_at ? 'resolved' : 'firing';

            return (
              <tr key={rca.incident_id} className="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center font-mono">
                  {rca.incident_id}
                </td>
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center whitespace-nowrap leading-relaxed">
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {formatDate(rca.fired_at)}
                  </div>
                  <div className="text-gray-400 font-bold my-0.5">~</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {rca.resolved_at ? formatDate(rca.resolved_at) : 'Firing'}
                  </div>
                </td>
                <td
                  className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm font-medium text-gray-700 dark:text-white cursor-pointer hover:underline hover:text-blue-600 dark:hover:text-blue-400 text-center min-w-[300px] break-words"
                  onClick={() => onTitleClick(rca.incident_id)}
                >
                  {rca.title}
                </td>
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${severityStyles[rawSeverity] || severityStyles.info}`}>
                    {rawSeverity}
                  </span>
                </td>
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${statusStyles[status] || statusStyles.firing}`}>
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

export default RCATable;