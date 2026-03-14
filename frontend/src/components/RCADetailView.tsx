import React, { useCallback, useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useLocation, useNavigate } from 'react-router-dom';
import { RCADetail, AlertItem } from '../types';
import FeedbackSection from './FeedbackSection';
import { FileText, Search, Link2, Sparkles } from 'lucide-react';
import {
  fetchRCADetail,
  updateRCADetail,
  hideIncident,
  unhideIncident,
  resolveIncident,
  searchSimilarIncidents,
  triggerIncidentAnalysis,
  EmbeddingSearchResult
} from '../utils/api';

interface RCADetailViewProps {
  incidentId: string;
  onBack: () => void;
}

type LocationState = {
  autoEdit?: boolean;
};

const severityStyles: Record<string, string> = {
  warning: 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-900 dark:text-amber-200 dark:border-amber-700',
  critical: 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-900 dark:text-rose-200 dark:border-rose-700',
  info: 'bg-cyan-100 text-cyan-800 border-cyan-200 dark:bg-cyan-900 dark:text-cyan-200 dark:border-cyan-700',
  TBD: 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:border-slate-600',
};

const RCADetailView: React.FC<RCADetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [similarIncidents, setSimilarIncidents] = useState<EmbeddingSearchResult[]>([]);
  const [similarLoading, setSimilarLoading] = useState(false);

  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<Partial<RCADetail>>({});
  const [analyzingIncident, setAnalyzingIncident] = useState(false);
  const [analysisComplete, setAnalysisComplete] = useState(false);
  const [analysisBanner, setAnalysisBanner] = useState<string | null>(null);
  const prevSummaryRef = useRef<string | null>(null);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const seenAnalyzingRef = useRef(false);
  const pollCountRef = useRef(0);

  const location = useLocation();
  const navigate = useNavigate();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchRCADetail(incidentId);
      setData(detailData);
      setEditForm(detailData);

      if (detailData.is_analyzing === true) {
        setAnalyzingIncident(true);
        prevSummaryRef.current = detailData.analysis_summary ?? null;
        seenAnalyzingRef.current = true;
        pollCountRef.current = 0;
      }

      if (detailData.analysis_summary) {
        setSimilarLoading(true);
        try {
          const searchResult = await searchSimilarIncidents(detailData.analysis_summary, 4);
          const filtered = searchResult.results.filter(r => r.incident_id !== incidentId);
          setSimilarIncidents(filtered.slice(0, 3));
        } catch (searchErr) {
          console.error('Failed to search similar incidents:', searchErr);
        } finally {
          setSimilarLoading(false);
        }
      }
    } catch (err) {
      setError('Failed to load data.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [incidentId]);

  useEffect(() => {
    loadDetail();
  }, [loadDetail]);

  // Polling while analyzing
  useEffect(() => {
    if (!analyzingIncident) {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
      return;
    }
    pollingRef.current = setInterval(async () => {
      try {
        const freshData = await fetchRCADetail(incidentId);
        setData(freshData);
        setEditForm(freshData);
        pollCountRef.current += 1;

        if (freshData.is_analyzing) {
          seenAnalyzingRef.current = true;
          return; // 분석 진행 중 — 폴링 계속
        }

        // is_analyzing === false
        // Grace period: backend가 아직 시작 안 했을 수 있음
        if (!seenAnalyzingRef.current && pollCountRef.current < 5) {
          return; // 15초 grace period
        }

        // Timeout 체크
        if (pollCountRef.current >= 100) {
          setAnalysisBanner('Analysis timed out. Please try again.');
          setAnalyzingIncident(false);
          return;
        }

        if (seenAnalyzingRef.current) {
          // Backend에서 분석이 시작되고 완료됨
          setAnalysisComplete(true);
        } else {
          // Grace period 이후에도 is_analyzing을 못 봄 → 시작 실패
          setAnalysisBanner('Analysis did not start. Please try again.');
        }

        setAnalyzingIncident(false);
      } catch {
        // polling error, will retry
      }
    }, 3000);
    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    };
  }, [analyzingIncident, incidentId]);

  useEffect(() => {
    const state = location.state as LocationState | null;
    if (state?.autoEdit) {
      setIsEditing(true);
      window.history.replaceState({}, document.title);
    }
  }, [location]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSave = async () => {
    if (!data) return;
    try {
      await updateRCADetail(incidentId, editForm);
      setData({ ...data, ...editForm } as RCADetail);
      setIsEditing(false);
      alert('Successfully modified.');
    } catch (err) {
      console.error(err);
      alert('Failed to save.');
    }
  };

  const handleCancel = () => {
    if (data) {
      setEditForm(data);
    }
    setIsEditing(false);
  };

  const handleHide = async () => {
    if (!window.confirm("Are you sure you want to archive this report from the list?")) {
      return;
    }

    try {
      await hideIncident(incidentId);
      alert("Successfully archived. Moving to Mute Dashboard.");
      navigate('/muted', { 
        state: { newlyMutedId: incidentId } 
      });
    } catch (error) {
      console.error("Archiving failed:", error);
      alert("An error occurred. Please try again.");
    }
  };

  const handleUnhide = async () => {
    if (!window.confirm("Do you want to unarchive this report?")) {
      return;
    }

    try {
      await unhideIncident(incidentId);
      alert("Unarchived successfully.");
      loadDetail(); 
    } catch (error) {
      console.error("Unarchiving failed:", error);
      alert("An error occurred. Please try again.");
    }
  };

  const handleResolve = async () => {
    if (!window.confirm("Are you sure you want to resolve this incident?\nAfter resolving, AI will perform the final analysis.")) {
      return;
    }

    try {
      await resolveIncident(incidentId);
      alert("Incident resolved.\nResults will be updated once AI analysis is complete.");
      await loadDetail();
    } catch (error) {
      console.error("Resolve failed:", error);
      alert("An error occurred. Please try again.");
    }
  };

  const handleIncidentAnalyze = async () => {
    const message = data?.analysis_summary
      ? 'Would you like to request re-analysis for this incident?'
      : 'Would you like to request analysis for this incident?';
    if (!window.confirm(message)) return;

    try {
      setAnalysisBanner(null);
      setAnalysisComplete(false);
      prevSummaryRef.current = data?.analysis_summary ?? null;
      seenAnalyzingRef.current = false;
      pollCountRef.current = 0;
      setAnalyzingIncident(true);
      await triggerIncidentAnalysis(incidentId);
    } catch (err) {
      console.error('Incident analysis request failed:', err);
      setAnalysisBanner('Failed to request incident analysis.');
      setAnalyzingIncident(false);
    }
  };

  const formatTime = (isoString?: string | null) => {
    if (!isoString) return '-';
    return isoString.replace('T', ' ').split('.')[0];
  };

  if (loading) return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 max-w-7xl mx-auto">
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

  const isResolved = !!data.resolved_at;
  const isHidden = data.is_hidden ?? false; 

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 max-w-7xl mx-auto transition-colors duration-300">
      
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
                ID: {data.incident_id}
              </span>
              {isHidden && (
                <span className="text-[10px] font-bold text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-1.5 rounded">
                  HIDDEN
                </span>
              )}
            </div>

            {isEditing ? (
              <div className="flex flex-col gap-1">
                <input
                  type="text"
                  name="title"
                  value={editForm.title || ''}
                  onChange={handleInputChange}
                  className="text-lg font-bold text-slate-900 dark:text-white bg-white dark:bg-slate-700 border border-cyan-400 rounded px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-cyan-500"
                />
              </div>
            ) : (
              <h1 className="text-xl md:text-2xl font-bold text-slate-900 dark:text-white leading-tight break-words">
                {data.title}
              </h1>
            )}
          </div>
        </div>
        
        <div className="flex items-center gap-3 self-end md:self-auto">
          {isEditing ? (
            <div className="flex items-center gap-2">
              <select
                name="severity"
                value={editForm.severity}
                onChange={handleInputChange}
                className="px-3 py-2 rounded border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
              >
                {editForm.severity === 'TBD' && <option value="TBD" disabled>TBD (Please select)</option>}
                <option value="critical">critical</option>
                <option value="warning">warning</option>
              </select>

              <button 
                onClick={handleSave}
                className="px-4 py-2 bg-cyan-600 text-white text-sm font-semibold rounded hover:bg-cyan-700 transition shadow-sm"
              >
                Save
              </button>
              <button 
                onClick={handleCancel}
                className="px-4 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 text-sm font-semibold rounded hover:bg-slate-50 dark:hover:bg-slate-600 transition shadow-sm"
              >
                Cancel
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              
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

              <button
                onClick={handleIncidentAnalyze}
                disabled={analyzingIncident}
                className="px-4 py-1.5 text-sm text-violet-600 dark:text-violet-400 border border-violet-600 dark:border-violet-400 rounded hover:bg-violet-50 dark:hover:bg-violet-900/20 transition-colors font-medium disabled:opacity-50"
              >
                {analyzingIncident ? 'Analyzing...' : data.analysis_summary ? 'Re-Analyze' : 'Analyze'}
              </button>

              {!isResolved && (
                <button
                  onClick={handleResolve}
                  className="px-4 py-1.5 text-sm text-emerald-600 dark:text-emerald-400 border border-emerald-600 dark:border-emerald-400 rounded hover:bg-emerald-50 dark:hover:bg-emerald-900/20 transition-colors font-medium"
                >
                  Resolve
                </button>
              )}

              {isHidden ? (
                <button
                  onClick={handleUnhide}
                  className="px-4 py-1.5 text-sm text-cyan-600 dark:text-cyan-400 border border-cyan-600 dark:border-cyan-400 rounded hover:bg-cyan-50 dark:hover:bg-cyan-900/20 transition-colors font-medium"
                >
                  Unarchive
                </button>
              ) : (
                <button
                  onClick={handleHide}
                  className="px-4 py-1.5 text-sm text-rose-600 dark:text-rose-400 border border-rose-600 dark:border-rose-400 rounded hover:bg-rose-50 dark:hover:bg-rose-900/20 transition-colors"
                >
                  Archive
                </button>
              )}
              
              {!isHidden && (
                <button 
                  onClick={() => setIsEditing(true)}
                  className="px-4 py-1.5 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 text-sm font-semibold rounded hover:bg-slate-50 dark:hover:bg-slate-700 transition"
                >
                  Edit
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Analysis status banners */}
      {analysisComplete && (
        <div className="mb-4 p-3 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-lg text-emerald-600 dark:text-emerald-400 text-sm">
          Analysis completed.
        </div>
      )}
      {analysisBanner && (
        <div className="mb-4 p-3 bg-rose-50 dark:bg-rose-900/20 border border-rose-200 dark:border-rose-800 rounded-lg text-rose-600 dark:text-rose-400 text-sm">
          {analysisBanner}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             Fired at
          </div>
          <div className="text-slate-900 dark:text-slate-100 font-medium font-mono">
            {formatTime(data.fired_at)}
          </div>
        </div>

        <div className="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-lg border border-slate-100 dark:border-slate-700">
          <div className="text-xs text-slate-500 dark:text-slate-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             Resolved at
          </div>
          <div className="text-slate-900 dark:text-slate-100 font-medium font-mono">
            {data.resolved_at ? formatTime(data.resolved_at) : <span className="text-rose-500 font-bold">Firing</span>}
          </div>
        </div>

        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-slate-900 dark:text-slate-100 mb-3 flex items-center gap-2">
            <FileText className="w-5 h-5 text-cyan-600 dark:text-cyan-400" />
            Incident Summary
          </h3>
          
          {isEditing ? (
            <textarea
              name="analysis_summary"
              value={editForm.analysis_summary || ''}
              onChange={handleInputChange}
              rows={5}
              placeholder="Write a summary in markdown format here..."
              className="w-full p-4 border border-cyan-400 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500 transition-shadow shadow-sm"
            />
          ) : (
            // [수정 포인트] 노란색 -> 깔끔한 블루/그레이 톤으로 변경 + 코드 블록 고대비 적용
            <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-lg p-5 transition-colors">
            <div className="prose prose-sm prose-slate dark:prose-invert max-w-none text-slate-900 dark:text-slate-100 leading-relaxed">
                <ReactMarkdown 
                  remarkPlugins={[remarkGfm]}
                  components={{
                    strong: ({ node: _node, ...props }) => <span className="font-bold text-blue-800 dark:text-blue-300" {...props} />,
                    ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 space-y-1 my-2" {...props} />,
                    // [핵심] 코드는 흰 배경에 파란 글씨 + 테두리 = 가독성 최적화
                    code: ({ node: _node, ...props }) => (
                      <code className="bg-white dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300 px-1.5 py-0.5 rounded text-xs font-mono font-bold shadow-sm" {...props} />
                    ),
                  }}
                >
                  {data.analysis_summary || "*No summary information available.*"}
                </ReactMarkdown>
              </div>
            </div>
          )}
        </div>

        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-slate-900 dark:text-slate-100 mb-3 flex items-center gap-2">
            <Search className="w-5 h-5 text-cyan-600 dark:text-cyan-400" />
            Incident Analysis
          </h3>

          {isEditing ? (
            <textarea
              name="analysis_detail"
              value={editForm.analysis_detail || ''}
              onChange={handleInputChange}
              rows={15}
              placeholder="Write detailed analysis content in markdown here..."
              className="w-full p-4 border border-cyan-400 rounded-lg bg-slate-900 text-slate-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500 transition-shadow shadow-sm"
            />
          ) : (
            <div className="bg-slate-900 border border-slate-700 rounded-lg overflow-hidden shadow-sm">
              <div className="p-6 overflow-x-auto">
                <div className="prose prose-sm prose-invert prose-p:text-slate-100 prose-li:text-slate-100 prose-headings:text-slate-100 max-w-none font-mono leading-relaxed text-slate-100">
                  <ReactMarkdown 
                    remarkPlugins={[remarkGfm]}
                    components={{
                      h1: ({ node: _node, ...props }) => <h1 className="text-2xl font-extrabold text-blue-400 mt-8 mb-6 pb-2 border-b border-slate-700 flex items-center gap-2 [&_strong]:text-blue-400" {...props} />,
                      h2: ({ node: _node, ...props }) => <h2 className="text-xl font-bold text-indigo-300 mt-8 mb-4 pl-3 border-l-4 border-indigo-500 [&_strong]:text-indigo-300" {...props} />,
                      h3: ({ node: _node, ...props }) => <h3 className="text-lg font-semibold text-sky-300 mt-6 mb-3 ml-1 [&_strong]:text-sky-300" {...props} />,
                      h4: ({ node: _node, ...props }) => <h4 className="text-base font-semibold text-slate-100 mt-4 mb-2" {...props} />,
                      strong: ({ node: _node, ...props }) => <span className="font-bold text-amber-400" {...props} />,
                      ul: ({ node: _node, ...props }) => <ul className="list-disc pl-6 space-y-2 my-2 text-slate-100 leading-relaxed" {...props} />,
                      li: ({ node: _node, ...props }) => <li className="text-slate-100 leading-relaxed" {...props} />,
                      code: ({ node: _node, ...props }) => (
                        <code className="bg-slate-800 text-pink-400 px-1.5 py-0.5 rounded text-sm font-mono border border-slate-700 mx-1" {...props} />
                      ),
                      p: ({ node: _node, ...props }) => <p className="mb-4 text-slate-100 leading-relaxed" {...props} />,
                      a: ({ node: _node, ...props }) => <a className="text-blue-400 hover:text-blue-300 hover:underline transition-colors" target="_blank" rel="noopener noreferrer" {...props} />,
                    }}
                  >
                    {data.analysis_detail || "*No detailed analysis content available.*"}
                  </ReactMarkdown>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="md:col-span-2 border-t border-slate-200 dark:border-slate-700 pt-6 mt-2">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
              <Link2 className="w-5 h-5 text-cyan-600 dark:text-cyan-400" />
              Related Alerts ({data.alerts?.length || 0} incidents)
            </h3>
          </div>

          {data.alerts && data.alerts.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-700">
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Alert</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Severity</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Status</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Fired At</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Resolved At</th>
                  </tr>
                </thead>
                <tbody>
                  {data.alerts.map((alert: AlertItem) => (
                    <tr
                      key={alert.alert_id}
                      onClick={() => navigate(`/alerts/${alert.alert_id}`)}
                      className="border-b border-slate-100 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer"
                    >
                      <td className="py-3 px-4">
                        <div className="font-medium text-cyan-600 dark:text-cyan-400 hover:underline">{alert.alarm_title}</div>
                        <div className="text-xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">{alert.alert_id}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-bold border ${severityStyles[alert.severity] || severityStyles.info}`}>
                          {alert.severity}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-bold ${
                          alert.status === 'resolved'
                            ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                            : 'bg-rose-100 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                        }`}>
                          {alert.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono text-xs text-slate-600 dark:text-slate-300">
                        {formatTime(alert.fired_at)}
                      </td>
                      <td className="py-3 px-4 font-mono text-xs text-slate-600 dark:text-slate-300">
                        {alert.resolved_at ? formatTime(alert.resolved_at) : '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-8 text-center border border-dashed border-slate-300 dark:border-slate-600">
              <p className="text-slate-500 dark:text-slate-400">No connected alerts.</p>
            </div>
          )}
        </div>

        <div className="md:col-span-2 border-t border-slate-200 dark:border-slate-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-4 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan-600 dark:text-cyan-400" />
            Top 3 Similar Incidents
            {similarLoading && <span className="ml-2 text-sm font-normal text-slate-500">(Searching...)</span>}
          </h3>

          {similarIncidents.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {similarIncidents.map((item: EmbeddingSearchResult, idx: number) => (
                <div
                  key={idx}
                  className="bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 p-4 rounded-lg shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => navigate(`/incidents/${item.incident_id}`)}
                >
                  <div className="mb-2 flex justify-between items-center">
                    <span className="text-xs font-mono text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded">
                      {item.incident_id}
                    </span>
                    <span className="text-xs font-bold text-cyan-600 dark:text-cyan-400 bg-cyan-50 dark:bg-cyan-900/30 px-2 py-0.5 rounded-full">
                      {Math.round(item.similarity * 100)}% similar
                    </span>
                  </div>
                  <div className="text-sm font-medium text-slate-800 dark:text-slate-200 line-clamp-3" title={item.incident_summary}>
                    {item.incident_summary}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-8 text-center border border-dashed border-slate-300 dark:border-slate-600">
              <p className="text-slate-500 dark:text-slate-400">
                {!data.analysis_summary ? 'Cannot search for similar incidents without an analysis summary.' : 'No similar incident records available.'}
              </p>
            </div>
          )}
        </div>

        <div className="md:col-span-2">
          <FeedbackSection targetType="incident" targetId={incidentId} />
        </div>

      </div>
    </div>
  );
};

export default RCADetailView;
