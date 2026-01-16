import React, { useCallback, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';
import { fetchAlertDetail } from '../utils/api';

export interface AlertDetail {
  alert_id: string;
  incident_id: string | null;
  alarm_title: string;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
  analysis_summary: string;
  analysis_detail: string;
  fingerprint: string;
  thread_ts: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
}

interface AlertDetailViewProps {
  alertId: string;
  onBack: () => void;
}

const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
};

const AlertDetailView: React.FC<AlertDetailViewProps> = ({ alertId, onBack }) => {
  const [data, setData] = useState<AlertDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchAlertDetail(alertId);
      setData(detailData);
    } catch (err) {
      setError('데이터를 불러오지 못했습니다.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [alertId]);

  useEffect(() => {
    loadDetail();
  }, [loadDetail]);

  const formatTime = (isoString?: string | null) => {
    if (!isoString) return '-';
    return isoString.replace('T', ' ').split('.')[0];
  };

  if (loading) return <div className="p-12 text-center text-gray-500 dark:text-gray-400">상세 정보를 불러오는 중...</div>;
  if (error || !data) return <div className="p-12 text-center text-red-500 bg-red-50 dark:bg-red-900/20 rounded-lg m-4">{error}</div>;

  const isResolved = data.status === 'resolved';

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-5xl mx-auto transition-colors duration-300">

      {/* 헤더 영역 */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-8 border-b border-gray-200 dark:border-gray-700 pb-6 gap-4">

        <div className="flex items-start md:items-center gap-4 flex-1 w-full">
          <button
            onClick={onBack}
            className="text-sm text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 font-medium px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition flex-shrink-0"
          >
            ← Back
          </button>

          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-mono text-gray-400 dark:text-gray-500 uppercase tracking-wider border border-gray-200 dark:border-gray-700 px-1.5 rounded">
                Alert ID: {data.alert_id.slice(0, 12)}...
              </span>
            </div>
            <h1 className="text-xl md:text-2xl font-bold text-gray-900 dark:text-white leading-tight">
              {data.alarm_title}
            </h1>
          </div>
        </div>

        <div className="flex items-center gap-2 self-end md:self-auto">
          <span
            className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0
              ${isResolved
                ? 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700'
                : 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700'
              }`}
          >
            {isResolved ? 'Resolved' : 'Firing'}
          </span>

          <span
            className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0
              ${severityStyles[data.severity || 'info'] || severityStyles.info}`}
          >
            {data.severity}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* 발생 시간 */}
        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             🔥 발생 시간
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium font-mono">
            {formatTime(data.fired_at)}
          </div>
        </div>

        {/* 해결 시간 */}
        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             ✅ 해결 시간
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium font-mono">
            {data.resolved_at ? formatTime(data.resolved_at) : <span className="text-red-500 font-bold">Ongoing</span>}
          </div>
        </div>

        {/* 연결된 Incident */}
        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             🔗 연결된 Incident
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium">
            {data.incident_id ? (
              <span
                className="text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"
                onClick={() => navigate(`/incidents/${data.incident_id}`)}
              >
                {data.incident_id}
              </span>
            ) : (
              <span className="text-gray-400">연결된 Incident 없음</span>
            )}
          </div>
        </div>

        {/* Fingerprint */}
        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             🔑 Fingerprint
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium font-mono text-sm">
            {data.fingerprint || '-'}
          </div>
        </div>

        {/* Labels */}
        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            🏷️ Labels
          </h3>
          <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
            <div className="flex flex-wrap gap-2">
              {Object.entries(data.labels || {}).map(([key, value]) => (
                <span
                  key={key}
                  className="inline-flex items-center px-2 py-1 rounded text-xs font-mono bg-gray-200 dark:bg-gray-600 text-gray-800 dark:text-gray-200"
                >
                  <span className="font-semibold">{key}:</span>
                  <span className="ml-1">{value}</span>
                </span>
              ))}
            </div>
          </div>
        </div>

        {/* Annotations */}
        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📝 Annotations
          </h3>
          <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700 space-y-2">
            {Object.entries(data.annotations || {}).map(([key, value]) => (
              <div key={key} className="text-sm">
                <span className="font-semibold text-gray-700 dark:text-gray-300">{key}:</span>
                <span className="ml-2 text-gray-600 dark:text-gray-400">{value}</span>
              </div>
            ))}
          </div>
        </div>

        {/* 분석 요약 */}
        {data.analysis_summary && (
          <div className="md:col-span-2">
            <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
              📋 분석 요약
            </h3>
            <div className="bg-yellow-50 dark:bg-yellow-900/10 border border-yellow-200 dark:border-yellow-800/30 rounded-lg p-5">
              <div className="prose prose-sm prose-yellow dark:prose-invert max-w-none text-gray-800 dark:text-gray-200">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {data.analysis_summary}
                </ReactMarkdown>
              </div>
            </div>
          </div>
        )}

        {/* 상세 분석 */}
        {data.analysis_detail && (
          <div className="md:col-span-2">
            <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
              📝 상세 분석
            </h3>
            <div className="bg-gray-900 border border-gray-700 rounded-lg overflow-hidden shadow-sm">
              <div className="p-6 overflow-x-auto">
                <div className="prose prose-sm prose-invert max-w-none font-mono leading-relaxed">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {data.analysis_detail}
                  </ReactMarkdown>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AlertDetailView;
