import { useNavigate } from 'react-router-dom';

export interface AlertItem {
  alert_id: string;
  incident_id: string | null;
  alarm_title: string;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
}

interface AlertTableProps {
  alerts: AlertItem[];
  onTitleClick: (alert_id: string) => void;
}

const formatDate = (isoString: string | null) => {
  if (!isoString) return '-';
  return isoString.replace('T', ' ').split('.')[0];
};

const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
};

const statusStyles: Record<string, string> = {
  firing: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  resolved: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700',
};

function AlertTable({ alerts, onTitleClick }: AlertTableProps) {
  const navigate = useNavigate();

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full border-collapse border border-gray-300 dark:border-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Alert ID
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-center text-sm font-semibold text-gray-700 dark:text-gray-200">
              Incident ID
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
          {alerts.map((alert) => (
            <tr key={alert.alert_id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center font-mono">
                {alert.alert_id.slice(0, 8)}...
              </td>

              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                {alert.incident_id ? (
                  <span
                    className="text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                    onClick={() => navigate(`/incidents/${alert.incident_id}`)}
                  >
                    {alert.incident_id}
                  </span>
                ) : (
                  <span className="text-gray-400">-</span>
                )}
              </td>

              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-gray-700 dark:text-gray-300 text-center whitespace-nowrap leading-relaxed">
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {formatDate(alert.fired_at)}
                </div>
                <div className="text-gray-400 font-bold my-0.5">~</div>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {alert.resolved_at ? formatDate(alert.resolved_at) : 'Ongoing'}
                </div>
              </td>

              <td
                className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm font-medium text-gray-700 dark:text-white cursor-pointer hover:underline hover:text-black dark:hover:text-gray-200 text-center"
                onClick={() => onTitleClick(alert.alert_id)}
              >
                {alert.alarm_title}
              </td>

              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                    severityStyles[alert.severity] || severityStyles.info
                  }`}
                >
                  {alert.severity}
                </span>
              </td>

              <td className="border border-gray-300 dark:border-gray-600 px-4 py-3 text-sm text-center">
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
