import React, { useCallback, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';
import { fetchAlertDetail, fetchRCAs, updateAlertIncident } from '../utils/api';
import { RCAItem } from '../types';
import FeedbackSection from './FeedbackSection';

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
  warning: 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-900 dark:text-amber-200 dark:border-amber-700',
  critical: 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-900 dark:text-rose-200 dark:border-rose-700',
  info: 'bg-cyan-100 text-cyan-800 border-cyan-200 dark:bg-cyan-900 dark:text-cyan-200 dark:border-cyan-700',
  TBD: 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:border-slate-600',
};

const AlertDetailView: React.FC<AlertDetailViewProps> = ({ alertId, onBack }) => {
  const [data, setData] = useState<AlertDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditingIncident, setIsEditingIncident] = useState(false);
  const [incidents, setIncidents] = useState<RCAItem[]>([]);
  const [selectedIncidentId, setSelectedIncidentId] = useState<string>('');
  const [incidentLoading, setIncidentLoading] = useState(false);
  const navigate = useNavigate();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const [detailData, rcaList] = await Promise.all([
        fetchAlertDetail(alertId),
        fetchRCAs()
      ]);
      setData(detailData);
      setIncidents(rcaList);
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

  const getIncidentTitle = (incidentId: string | null) => {
    if (!incidentId) return null;
    const incident = incidents.find(i => i.incident_id === incidentId);
    return incident?.title || null;
  };

  const formatTime = (isoString?: string | null) => {
    if (!isoString) return '-';
    return isoString.replace('T', ' ').split('.')[0];
  };

  const handleEditIncident = async () => {
    setIncidentLoading(true);
    try {
      const rcaList = await fetchRCAs();
      setIncidents(rcaList);
      setSelectedIncidentId(data?.incident_id || '');
      setIsEditingIncident(true);
    } catch (err) {
      console.error('Failed to load incidents:', err);
      alert('인시던트 목록을 불러오는데 실패했습니다.');
    } finally {
      setIncidentLoading(false);
    }
  };

  const handleSaveIncident = async () => {
    if (!selectedIncidentId) {
      alert('인시던트를 선택해주세요.');
      return;
    }
    try {
      await updateAlertIncident(alertId, selectedIncidentId);
      setData(prev => prev ? { ...prev, incident_id: selectedIncidentId } : null);
      setIsEditingIncident(false);
      alert('인시던트 연결이 변경되었습니다.');
    } catch (err) {
      console.error('Failed to update incident:', err);
      alert('인시던트 연결 변경에 실패했습니다.');
    }
  };

  const handleCancelEditIncident = () => {
    setIsEditingIncident(false);
    setSelectedIncidentId(data?.incident_id || '');
  };

  if (loading) return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 max-w-5xl mx-auto">
      <div className="flex items-center gap-4 mb-8 pb-6 border-b border-slate-200 dark:border-slate-800">
        <div className="skeleton h-8 w-20" />
        <div className="flex-1 space-y-2">
          <div className="skeleton h-3 w-24" />
          <div className="skeleton h-6 w-3/4" />
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="skeleton h-20 rounded-lg" />
        <div className="skeleton h-20 rounded-lg" />
        <div className="md:col-span-2 skeleton h-40 rounded-lg" />
        <div className="md:col-span-2 skeleton h-64 rounded-lg" />
      </div>
    </div>
  );
  if (error || !data) return <div className="p-12 text-center text-rose-500 bg-rose-50 dark:bg-rose-900/20 rounded-lg m-4">{error}</div>;

  const isResolved = data.status === 'resolved';

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 max-w-5xl mx-auto transition-colors duration-300">

      {/* 헤더 영역 */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-8 border-b border-slate-200 dark:border-slate-700 pb-6 gap-4">

        <div className="flex items-start md:items-center gap-4 flex-1 w-full">
          <button
            onClick={onBack}
            className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200 font-medium px-3 py-1.5 border border-slate-300 dark:border-slate-600 rounded hover:bg-slate-50 dark:hover:bg-slate-700 transition flex-shrink-0"
          >
            ← Back
          </button>

          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[10px] font-mono text-slate-400 dark:text-slate-500 uppercase tracking-wider border border-slate-200 dark:border-slate-700 px-1.5 rounded">
                Alert ID: {data.alert_id.slice(0, 12)}...
              </span>
            </div>
            <h1 className="text-xl md:text-2xl font-bold text-slate-900 dark:text-white leading-tight">
              {data.alarm_title}
            </h1>
          </div>
        </div>

        <div className="flex items-center gap-2 self-end md:self-auto">
          <span
            className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0
              ${isResolved
                ? 'bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:border-emerald-700'
                : 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-900 dark:text-rose-200 dark:border-rose-700'
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
        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
            Fired at
          </div>
          <div className="text-slate-900 dark:text-slate-100 font-medium font-mono">
            {formatTime(data.fired_at)}
          </div>
        </div>

        {/* 해결 시간 */}
        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
            Resolved at
          </div>
          <div className="text-slate-900 dark:text-slate-100 font-medium font-mono">
            {data.resolved_at ? formatTime(data.resolved_at) : <span className="text-rose-500 font-bold">Firing</span>}
          </div>
        </div>

        {/* 연결된 Incident */}
        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
            Related Incident
          </div>
          {isEditingIncident ? (
            <div className="space-y-3">
              <select
                value={selectedIncidentId}
                onChange={(e) => setSelectedIncidentId(e.target.value)}
                className="w-full px-3 py-2 rounded border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
              >
                <option value="">인시던트를 선택하세요</option>
                {incidents.map((incident) => (
                  <option key={incident.incident_id} value={incident.incident_id}>
                    [{incident.incident_id.slice(0, 8)}] {incident.title || '(제목 없음)'}
                  </option>
                ))}
              </select>
              <div className="flex gap-2">
                <button
                  onClick={handleSaveIncident}
                  className="px-3 py-1.5 bg-cyan-600 text-white text-xs font-semibold rounded hover:bg-cyan-700 transition"
                >
                  저장
                </button>
                <button
                  onClick={handleCancelEditIncident}
                  className="px-3 py-1.5 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 text-xs font-semibold rounded hover:bg-slate-100 dark:hover:bg-slate-600 transition"
                >
                  취소
                </button>
              </div>
            </div>
          ) : (
            <div className="flex flex-col">
              <div className="text-slate-900 dark:text-slate-100 font-medium mb-2">
                {data.incident_id ? (
                  <div
                    className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-600 rounded px-1 -mx-1 transition"
                    onClick={() => navigate(`/incidents/${data.incident_id}`)}
                  >
                    <div className="text-rose-500 dark:text-rose-400 text-sm font-mono">{data.incident_id}</div>
                    <div className="text-slate-900 dark:text-white">{getIncidentTitle(data.incident_id) || '(제목 없음)'}</div>
                  </div>
                ) : (
                  <span className="text-slate-400">연결된 Incident 없음</span>
                )}
              </div>
              <div className="flex justify-end">
                <button
                  onClick={handleEditIncident}
                  disabled={incidentLoading}
                  className="px-2 py-1 text-xs font-medium text-orange-600 dark:text-orange-400 border border-orange-400 dark:border-orange-500 rounded hover:bg-orange-50 dark:hover:bg-orange-900/20 transition disabled:opacity-50"
                >
                  {incidentLoading ? '로딩...' : 'Edit'}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Fingerprint */}
        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             Fingerprint
          </div>
          <div className="text-slate-900 dark:text-slate-100 font-medium font-mono text-sm">
            {data.fingerprint || '-'}
          </div>
        </div>

        {/* Labels */}
        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-3 flex items-center gap-2">
            Labels
          </h3>
          <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
            <div className="flex flex-wrap gap-2">
              {Object.entries(data.labels || {}).map(([key, value]) => (
                <span
                  key={key}
                  className="inline-flex items-center px-2 py-1 rounded text-xs font-mono bg-slate-200 dark:bg-slate-600 text-slate-800 dark:text-slate-200"
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
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-3 flex items-center gap-2">
            Annotations
          </h3>
          <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700 space-y-2">
            {Object.entries(data.annotations || {}).map(([key, value]) => (
              <div key={key} className="text-sm">
                <span className="font-semibold text-slate-700 dark:text-slate-300">{key}:</span>
                <span className="ml-2 text-slate-600 dark:text-slate-400">{value}</span>
              </div>
            ))}
          </div>
        </div>

        {/* 분석 요약 (Blue Tone + High Contrast Code) */}
        {data.analysis_summary && (
          <div className="md:col-span-2">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-3 flex items-center gap-2">
              Alert Summary
            </h3>
            <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-lg p-5 transition-colors">
              <div className="prose prose-sm prose-slate dark:prose-invert max-w-none text-slate-800 dark:text-slate-200 leading-relaxed">
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  components={{
                    strong: ({ node: _node, ...props }) => <span className="font-bold text-blue-800 dark:text-blue-300" {...props} />,
                    ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 space-y-1 my-2" {...props} />,
                    // [핵심] 흰 배경에 진한 파랑 글씨로 코드 블록 강조
                    code: ({ node: _node, ...props }) => (
                      <code className="bg-white dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300 px-1.5 py-0.5 rounded text-xs font-mono font-bold shadow-sm" {...props} />
                    ),
                  }}
                >
                  {data.analysis_summary}
                </ReactMarkdown>
              </div>
            </div>
          </div>
        )}

        {/* 상세 분석 (DarkMode Visibility Fix) */}
        {data.analysis_detail && (
          <div className="md:col-span-2">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-3 flex items-center gap-2">
              Alert Analysis
            </h3>
            <div className="bg-slate-900 border border-slate-700 rounded-lg overflow-hidden shadow-sm">
              <div className="p-6 overflow-x-auto">
                <div className="prose prose-sm prose-invert max-w-none font-mono leading-relaxed">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                  components={{
                    h1: ({ node: _node, ...props }) => <h1 className="text-2xl font-extrabold text-blue-400 mt-8 mb-6 pb-2 border-b border-slate-700 flex items-center gap-2 [&_strong]:text-blue-400" {...props} />,
                    h2: ({ node: _node, ...props }) => <h2 className="text-xl font-bold text-indigo-300 mt-8 mb-4 pl-3 border-l-4 border-indigo-500 [&_strong]:text-indigo-300" {...props} />,
                    h3: ({ node: _node, ...props }) => <h3 className="text-lg font-semibold text-sky-300 mt-6 mb-3 ml-1 [&_strong]:text-sky-300" {...props} />,
                    strong: ({ node: _node, ...props }) => <span className="font-bold text-amber-400" {...props} />,
                    ul: ({ node: _node, ...props }) => <ul className="list-disc pl-6 space-y-2 my-2 text-slate-300 leading-relaxed" {...props} />,
                    code: ({ node: _node, ...props }) => (
                      <code className="bg-slate-800 text-pink-400 px-1.5 py-0.5 rounded text-sm font-mono border border-slate-700 mx-1" {...props} />
                    ),
                    p: ({ node: _node, ...props }) => <p className="mb-4 text-slate-300 leading-relaxed" {...props} />,
                    a: ({ node: _node, ...props }) => <a className="text-blue-400 hover:text-blue-300 hover:underline transition-colors" target="_blank" rel="noopener noreferrer" {...props} />,
                  }}
                  >
                    {data.analysis_detail}
                  </ReactMarkdown>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="md:col-span-2">
          <FeedbackSection targetType="alert" targetId={alertId} />
        </div>
      </div>
    </div>
  );
};

export default AlertDetailView;
