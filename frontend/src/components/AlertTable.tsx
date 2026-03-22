import { useState } from 'react';
import { Inbox } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
// [중요] AlertItem은 types에서 가져옵니다 (중복 정의 방지)
import { AlertItem } from '../types';
import { bulkResolveAlerts } from '../utils/api';

interface AlertTableProps {
  alerts: AlertItem[]; // 부모(App.tsx)가 이미 필터링해서 준 데이터
  onTitleClick: (alert_id: string) => void;
  onRefresh?: () => void;
}

const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

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

function AlertTable({ alerts, onTitleClick, onRefresh }: AlertTableProps) {
  const navigate = useNavigate();
  const [selectedAlertIds, setSelectedAlertIds] = useState<Set<string>>(new Set());
  const [isResolving, setIsResolving] = useState(false);

  const firingAlerts = alerts.filter((a) => a.status === 'firing');
  const allFiringSelected =
    firingAlerts.length > 0 && firingAlerts.every((a) => selectedAlertIds.has(a.alert_id));

  const toggleSelect = (alertId: string) => {
    setSelectedAlertIds((prev) => {
      const next = new Set(prev);
      if (next.has(alertId)) {
        next.delete(alertId);
      } else {
        next.add(alertId);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (allFiringSelected) {
      setSelectedAlertIds(new Set());
    } else {
      setSelectedAlertIds(new Set(firingAlerts.map((a) => a.alert_id)));
    }
  };

  const handleBulkResolve = async () => {
    const ids = Array.from(selectedAlertIds);
    if (!window.confirm(`${ids.length}개의 alert를 resolve 하시겠습니까?`)) {
      return;
    }

    setIsResolving(true);
    try {
      const result = await bulkResolveAlerts(ids);
      alert(`${result.resolved}개 resolved, ${result.failed}개 failed`);
      setSelectedAlertIds(new Set());
      if (onRefresh) {
        onRefresh();
      }
    } catch (error) {
      console.error('Bulk resolve failed:', error);
      alert('An error occurred. Please try again.');
    } finally {
      setIsResolving(false);
    }
  };

  if (!alerts || alerts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-slate-400 dark:text-slate-500">
        <Inbox className="w-12 h-12 mb-3 stroke-1" />
        <p className="text-sm font-medium">No data available</p>
        <p className="text-xs mt-1">No results match the search criteria</p>
      </div>
    );
  }

  const showCheckbox = firingAlerts.length > 0;

  return (
    <div className="overflow-x-auto">
      {showCheckbox && (
        <div className="flex items-center gap-2 mb-2">
          <button
            onClick={handleBulkResolve}
            disabled={selectedAlertIds.size === 0 || isResolving}
            className="px-3 py-1 text-sm text-emerald-600 dark:text-emerald-400 border border-emerald-600 dark:border-emerald-400 rounded hover:bg-emerald-50 dark:hover:bg-emerald-900/20 transition-colors font-medium disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {isResolving ? 'Resolving...' : `Resolve Selected (${selectedAlertIds.size})`}
          </button>
        </div>
      )}
      <table className="w-full">
        <thead>
          <tr className="border-b border-slate-200 dark:border-slate-800">
            {showCheckbox && (
              <th className="px-2 py-2 w-8">
                <input
                  type="checkbox"
                  checked={allFiringSelected}
                  onChange={toggleSelectAll}
                  className="rounded"
                />
              </th>
            )}
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Incident ID
            </th>
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Time
            </th>
            <th className="w-full px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Title
            </th>
            <th className="px-4 py-3 text-left text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-r border-slate-200 dark:border-slate-700">
              Namespace
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
          {/* [핵심] filteredData가 아니라 그냥 alerts를 맵핑합니다 */}
          {alerts.map((alert) => (
            <tr key={alert.alert_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors cursor-pointer" onClick={() => onTitleClick(alert.alert_id)}>
              {/* Checkbox - firing alert이 있을 때만 렌더링 */}
              {showCheckbox && (
                <td className="px-2 py-2" onClick={(e) => e.stopPropagation()}>
                  {alert.status === 'firing' ? (
                    <input
                      type="checkbox"
                      checked={selectedAlertIds.has(alert.alert_id)}
                      onChange={() => toggleSelect(alert.alert_id)}
                      className="rounded"
                    />
                  ) : (
                    <span className="w-4 inline-block" />
                  )}
                </td>
              )}

              {/* Incident ID */}
              <td className="px-4 py-3.5 text-sm font-semibold whitespace-nowrap border-r border-slate-200 dark:border-slate-700">
                {alert.incident_id ? (
                  <span
                    className="font-mono text-cyan-600 dark:text-cyan-400 cursor-pointer hover:underline"
                    onClick={(e) => { e.stopPropagation(); navigate(`/incidents/${alert.incident_id}`); }}
                  >
                    {alert.incident_id}
                  </span>
                ) : (
                  <span className="text-slate-400">-</span>
                )}
              </td>

              {/* Time */}
              <td className="px-4 py-3.5 whitespace-nowrap border-r border-slate-200 dark:border-slate-700">
                <span className="font-mono text-sm font-medium text-slate-600 dark:text-slate-300">
                  {formatDate(alert.fired_at)}
                  {alert.resolved_at && <><span className="text-slate-400 dark:text-slate-500"> → </span>{formatDate(alert.resolved_at)}</>}
                </span>
              </td>

              {/* Title */}
              <td className="px-4 py-3.5 text-sm font-medium text-slate-900 dark:text-slate-100 hover:text-cyan-600 dark:hover:text-cyan-400 min-w-[300px] break-words border-r border-slate-200 dark:border-slate-700">
                {alert.alarm_title}
              </td>

              {/* Namespace */}
              <td className="px-4 py-3.5 text-sm whitespace-nowrap border-r border-slate-200 dark:border-slate-700">
                <span className="font-mono text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-0.5 rounded">
                  {alert.namespace || '-'}
                </span>
              </td>

              {/* Severity */}
              <td className="px-4 py-3.5 text-sm border-r border-slate-200 dark:border-slate-700">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                    severityStyles[alert.severity] || severityStyles.info
                  }`}
                >
                  {alert.severity}
                </span>
              </td>

              {/* Status */}
              <td className="px-4 py-3.5 text-sm">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                    statusStyles[alert.status] || statusStyles.firing
                  }`}
                >
                  {alert.status}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default AlertTable;
