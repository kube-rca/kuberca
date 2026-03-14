import { useEffect, useMemo, useState } from 'react';
import { AnalyticsDashboardResponse, AnalyticsSeriesPoint, fetchAnalyticsDashboard } from '../utils/api';

const formatNumber = (value: number) => new Intl.NumberFormat('en-US').format(value);

const format1 = (value: number) => value.toFixed(1);

const normalizeSeries = (points: AnalyticsSeriesPoint[]) => {
  const maxValue = Math.max(1, ...points.map((p) => Math.max(p.incidents, p.alerts)));
  return points.map((p, index) => {
    const x = points.length === 1 ? 0 : (index / (points.length - 1)) * 100;
    const incidentY = 100 - (p.incidents / maxValue) * 100;
    const alertY = 100 - (p.alerts / maxValue) * 100;
    return { x, incidentY, alertY };
  });
};

const linePoints = (values: { x: number; incidentY: number; alertY: number }[], key: 'incidentY' | 'alertY') =>
  values.map((v) => `${v.x},${v[key]}`).join(' ');

const DistributionBars = ({ title, items }: { title: string; items: Array<{ key: string; count: number }> }) => {
  const max = Math.max(1, ...items.map((item) => item.count));
  return (
    <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900">
      <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">{title}</h3>
      <div className="space-y-3">
        {items.map((item) => (
          <div key={item.key}>
            <div className="flex justify-between text-xs text-slate-600 dark:text-slate-300 mb-1">
              <span className="capitalize">{item.key || 'unknown'}</span>
              <span>{formatNumber(item.count)}</span>
            </div>
            <div className="h-2 rounded-full bg-slate-100 dark:bg-slate-800">
              <div
                className="h-2 rounded-full bg-cyan-500"
                style={{ width: `${(item.count / max) * 100}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

const TopNamespaceList = ({ items }: { items: Array<{ key: string; count: number }> }) => (
  <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900">
    <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">Top Namespaces</h3>
    <div className="space-y-2">
      {items.length === 0 && <div className="text-sm text-slate-500 dark:text-slate-400">No namespace data</div>}
      {items.map((item) => (
        <div key={item.key} className="flex justify-between text-sm">
          <span className="text-slate-700 dark:text-slate-200 truncate">{item.key || 'unknown'}</span>
          <span className="text-slate-500 dark:text-slate-400">{formatNumber(item.count)}</span>
        </div>
      ))}
    </div>
  </div>
);

const TrendLineChart = ({ points }: { points: AnalyticsSeriesPoint[] }) => {
  const normalized = useMemo(() => normalizeSeries(points), [points]);
  const incidentPath = linePoints(normalized, 'incidentY');
  const alertPath = linePoints(normalized, 'alertY');

  return (
    <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900">
      <div className="flex justify-between items-center mb-3">
        <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-100">Daily Trend</h3>
        <div className="flex gap-4 text-xs">
          <span className="text-cyan-600 dark:text-cyan-400">Incidents</span>
          <span className="text-amber-600 dark:text-amber-400">Alerts</span>
        </div>
      </div>
      <svg viewBox="0 0 100 100" className="w-full h-52">
        <polyline fill="none" stroke="#06b6d4" strokeWidth="2" points={incidentPath} />
        <polyline fill="none" stroke="#f59e0b" strokeWidth="2" points={alertPath} />
      </svg>
      <div className="mt-2 flex justify-between text-[11px] text-slate-500 dark:text-slate-400">
        <span>{points[0]?.date ?? '-'}</span>
        <span>{points[points.length - 1]?.date ?? '-'}</span>
      </div>
    </div>
  );
};

export default function AnalysisDashboard() {
  const [window, setWindow] = useState('30d');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<AnalyticsDashboardResponse | null>(null);

  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const next = await fetchAnalyticsDashboard(window);
        if (active) setData(next);
      } catch {
        if (active) setError('Failed to load analysis data.');
      } finally {
        if (active) setLoading(false);
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [window]);

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 transition-colors duration-300">
      <div className="mb-6 flex flex-col md:flex-row justify-between md:items-center gap-4">
        <h1 className="text-xl font-semibold font-mono tracking-wide text-slate-900 dark:text-slate-100">Analysis Dashboard</h1>
        <select
          value={window}
          onChange={(e) => setWindow(e.target.value)}
          className="px-4 py-2 text-sm font-medium border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-cyan-500"
        >
          <option value="7d">Last 7 days</option>
          <option value="14d">Last 14 days</option>
          <option value="30d">Last 30 days</option>
          <option value="90d">Last 90 days</option>
        </select>
      </div>

      {loading ? (
        <div className="space-y-4 py-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="skeleton h-8 w-full" />
          ))}
        </div>
      ) : error ? (
        <div className="bg-rose-50 dark:bg-rose-950/20 border border-rose-200 dark:border-rose-800 rounded-md p-4 text-rose-600 dark:text-rose-400">{error}</div>
      ) : data ? (
        <div className="space-y-5">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4">
              <div className="text-xs uppercase text-slate-500 dark:text-slate-400">Incidents</div>
              <div className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{formatNumber(data.summary.total_incidents)}</div>
            </div>
            <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4">
              <div className="text-xs uppercase text-slate-500 dark:text-slate-400">Alerts</div>
              <div className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{formatNumber(data.summary.total_alerts)}</div>
            </div>
            <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4">
              <div className="text-xs uppercase text-slate-500 dark:text-slate-400">Avg MTTR (min)</div>
              <div className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{format1(data.summary.avg_mttr_minutes)}</div>
            </div>
            <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4">
              <div className="text-xs uppercase text-slate-500 dark:text-slate-400">Alerts / Incident</div>
              <div className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{format1(data.summary.avg_alerts_per_incident)}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <TrendLineChart points={data.series.daily} />
            <TopNamespaceList items={data.breakdown.top_namespaces} />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <DistributionBars title="Incident Severity" items={data.breakdown.incident_severity} />
            <DistributionBars title="Alert Severity" items={data.breakdown.alert_severity} />
          </div>
        </div>
      ) : null}
    </div>
  );
}
