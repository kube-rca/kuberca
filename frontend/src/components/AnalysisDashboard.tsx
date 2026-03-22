import { MouseEvent, useEffect, useMemo, useState } from 'react';
import { AnalyticsDashboardResponse, AnalyticsSeriesPoint, fetchAnalyticsDashboard } from '../utils/api';
import { ExportColumn, ExportFormat, exportRows } from '../utils/export';

const formatNumber = (value: number) => new Intl.NumberFormat('en-US').format(value);

const format1 = (value: number) => value.toFixed(1);

const selectStyle =
  'px-4 py-2 text-sm font-medium border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-cyan-500';

const buttonStyle =
  'px-4 py-2 text-sm font-semibold border border-cyan-500 rounded-lg bg-cyan-600 text-white hover:bg-cyan-700 transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed';

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

const analysisExportColumns: ExportColumn<AnalyticsSeriesPoint>[] = [
  { key: 'date', header: 'Date', value: (row) => row.date },
  { key: 'incidents', header: 'Incidents', value: (row) => row.incidents },
  { key: 'alerts', header: 'Alerts', value: (row) => row.alerts },
];

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
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [tooltip, setTooltip] = useState<{ left: number; top: number } | null>(null);

  const handleHover = (event: MouseEvent<SVGRectElement>, index: number) => {
    const bounds = event.currentTarget.ownerSVGElement?.getBoundingClientRect();
    if (!bounds) return;

    setHoveredIndex(index);
    setTooltip({
      left: event.clientX - bounds.left,
      top: event.clientY - bounds.top,
    });
  };

  const hoveredPoint = hoveredIndex !== null ? points[hoveredIndex] : null;
  const hoveredPosition = hoveredIndex !== null ? normalized[hoveredIndex] : null;

  return (
    <div className="rounded-lg border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900">
      <div className="flex justify-between items-center mb-3">
        <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-100">Daily Trend</h3>
        <div className="flex gap-4 text-xs">
          <span className="text-cyan-600 dark:text-cyan-400">Incidents</span>
          <span className="text-amber-600 dark:text-amber-400">Alerts</span>
        </div>
      </div>
      <div className="relative">
        <svg
          viewBox="0 0 100 100"
          className="w-full h-52"
          onMouseLeave={() => {
            setHoveredIndex(null);
            setTooltip(null);
          }}
        >
          {hoveredPosition && (
            <line
              x1={hoveredPosition.x}
              x2={hoveredPosition.x}
              y1="0"
              y2="100"
              stroke="#94a3b8"
              strokeWidth="0.7"
              strokeDasharray="2 2"
            />
          )}
          <polyline fill="none" stroke="#06b6d4" strokeWidth="2" points={incidentPath} />
          <polyline fill="none" stroke="#f59e0b" strokeWidth="2" points={alertPath} />
          {hoveredPosition && (
            <>
              <circle cx={hoveredPosition.x} cy={hoveredPosition.incidentY} r="2.4" fill="#06b6d4" stroke="#fff" strokeWidth="0.9" />
              <circle cx={hoveredPosition.x} cy={hoveredPosition.alertY} r="2.4" fill="#f59e0b" stroke="#fff" strokeWidth="0.9" />
            </>
          )}
          {normalized.map((point, index) => {
            const prevX = index === 0 ? 0 : normalized[index - 1].x;
            const nextX = index === normalized.length - 1 ? 100 : normalized[index + 1].x;
            const startX = index === 0 ? 0 : (prevX + point.x) / 2;
            const endX = index === normalized.length - 1 ? 100 : (point.x + nextX) / 2;

            return (
              <rect
                key={`${points[index]?.date ?? index}-hover`}
                x={startX}
                y={0}
                width={Math.max(endX - startX, 4)}
                height={100}
                fill="transparent"
                onMouseEnter={(event) => handleHover(event, index)}
                onMouseMove={(event) => handleHover(event, index)}
              />
            );
          })}
        </svg>
        {hoveredPoint && tooltip && (
          <div
            className="pointer-events-none absolute z-10 min-w-[140px] rounded-lg border border-slate-200 dark:border-slate-700 bg-white/95 dark:bg-slate-900/95 px-3 py-2 text-xs shadow-lg backdrop-blur"
            style={{
              left: `${Math.min(Math.max(tooltip.left + 12, 8), 220)}px`,
              top: `${Math.max(tooltip.top - 56, 8)}px`,
            }}
          >
            <div className="mb-1 font-semibold text-slate-800 dark:text-slate-100">{hoveredPoint.date}</div>
            <div className="flex items-center justify-between gap-4 text-cyan-600 dark:text-cyan-400">
              <span>Incidents</span>
              <span className="font-semibold">{formatNumber(hoveredPoint.incidents)}</span>
            </div>
            <div className="flex items-center justify-between gap-4 text-amber-600 dark:text-amber-400">
              <span>Alerts</span>
              <span className="font-semibold">{formatNumber(hoveredPoint.alerts)}</span>
            </div>
          </div>
        )}
      </div>
      <div className="mt-2 flex justify-between text-[11px] text-slate-500 dark:text-slate-400">
        <span>{points[0]?.date ?? '-'}</span>
        <span>{points[points.length - 1]?.date ?? '-'}</span>
      </div>
    </div>
  );
};

export default function AnalysisDashboard() {
  const [window, setWindow] = useState('30d');
  const [exportFormat, setExportFormat] = useState<ExportFormat>('csv');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<AnalyticsDashboardResponse | null>(null);

  const exportMetaRows = data
    ? [
        { label: 'Window', value: data.window },
        { label: 'Generated At', value: data.generated_at },
        { label: 'Total Incidents', value: data.summary.total_incidents },
        { label: 'Firing Incidents', value: data.summary.firing_incidents },
        { label: 'Resolved Incidents', value: data.summary.resolved_incidents },
        { label: 'Total Alerts', value: data.summary.total_alerts },
        { label: 'Firing Alerts', value: data.summary.firing_alerts },
        { label: 'Resolved Alerts', value: data.summary.resolved_alerts },
        { label: 'Avg MTTR (min)', value: format1(data.summary.avg_mttr_minutes) },
        { label: 'Avg Alerts / Incident', value: format1(data.summary.avg_alerts_per_incident) },
      ]
    : [];

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
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
          <select
            value={window}
            onChange={(e) => setWindow(e.target.value)}
            className={selectStyle}
          >
            <option value="7d">Last 7 days</option>
            <option value="14d">Last 14 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>
          <select
            value={exportFormat}
            onChange={(e) => setExportFormat(e.target.value as ExportFormat)}
            className={selectStyle}
          >
            <option value="csv">CSV</option>
            <option value="excel">Excel (.xls)</option>
          </select>
          <button
            type="button"
            onClick={() => exportRows(data?.series.daily ?? [], analysisExportColumns, 'analysis_dashboard', exportFormat, undefined, exportMetaRows)}
            disabled={!data || data.series.daily.length === 0}
            className={buttonStyle}
          >
            Export
          </button>
        </div>
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
