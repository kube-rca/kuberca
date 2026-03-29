import { Inbox } from 'lucide-react';
import { RCAItem } from '../types';

interface ArchiveTableProps {
  rcas: RCAItem[]; // Mute된 Incident 데이터 리스트
  onTitleClick: (incident_id: string) => void;
}

const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

// 스타일 매핑
const severityStyles: Record<string, string> = {
  warning: 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/30 dark:text-amber-300 dark:border-amber-800',
  critical: 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/30 dark:text-rose-300 dark:border-rose-800',
  info: 'bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950/30 dark:text-sky-300 dark:border-sky-800',
  tbd: 'bg-slate-50 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700',
  TBD: 'bg-slate-50 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700',
};

const statusStyles: Record<string, string> = {
  firing: 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/30 dark:text-rose-300 dark:border-rose-800',
  resolved: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/30 dark:text-emerald-300 dark:border-emerald-800',
};

function ArchiveTable({ rcas, onTitleClick }: ArchiveTableProps) {
  // 데이터가 없을 경우 처리
  if (!rcas || rcas.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-slate-400 dark:text-slate-500">
        <Inbox className="w-12 h-12 mb-3 stroke-1" />
        <p className="text-sm font-medium">No data available</p>
        <p className="text-xs mt-1">No results match the search criteria</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-slate-200 dark:border-slate-800">
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              ID
            </th>
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Time
            </th>
            <th className="w-full px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Title
            </th>
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Severity
            </th>
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              Status
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
          {rcas.map((rca) => {
            const rawSeverity = rca.severity;
            const status = rca.resolved_at ? 'resolved' : 'firing';

            return (
              <tr key={rca.incident_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors cursor-pointer" onClick={() => onTitleClick(rca.incident_id)}>
                {/* ID */}
                <td className="px-4 py-3.5 text-sm font-mono font-semibold text-slate-700 dark:text-slate-300 whitespace-nowrap border-r border-slate-200 dark:border-slate-700">
                  {rca.incident_id}
                </td>

                {/* Time */}
                <td className="px-4 py-3.5 whitespace-nowrap border-r border-slate-200 dark:border-slate-700">
                  <span className="font-mono text-sm font-medium text-slate-600 dark:text-slate-300">
                    {formatDate(rca.fired_at)}
                    {rca.resolved_at && <><span className="text-slate-400 dark:text-slate-500"> → </span>{formatDate(rca.resolved_at)}</>}
                  </span>
                </td>

                {/* Title (클릭 시 상세 이동) */}
                <td className="px-4 py-3.5 text-sm font-medium text-slate-900 dark:text-slate-100 hover:text-cyan-600 dark:hover:text-cyan-400 min-w-[300px] break-words border-r border-slate-200 dark:border-slate-700">
                  <span className="line-clamp-2" title={rca.title}>{rca.title}</span>
                </td>

                {/* Severity */}
                <td className="px-4 py-3.5 text-sm border-r border-slate-200 dark:border-slate-700">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                      severityStyles[rawSeverity] || severityStyles.info
                    }`}
                  >
                    {rawSeverity}
                  </span>
                </td>

                {/* Status */}
                <td className="px-4 py-3.5 text-sm">
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