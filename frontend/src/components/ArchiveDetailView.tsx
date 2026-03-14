import React, { useCallback, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';
import { RCADetail, AlertItem } from '../types';
import { 
  fetchRCADetail, 
  unhideIncident, 
  searchSimilarIncidents, 
  EmbeddingSearchResult 
} from '../utils/api';

interface ArchivedDetailViewProps {
  incidentId: string;
  onBack: () => void;
}

const severityStyles: Record<string, string> = {
  warning: 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-900 dark:text-amber-200 dark:border-amber-700',
  critical: 'bg-rose-100 text-rose-800 border-rose-200 dark:bg-rose-900 dark:text-rose-200 dark:border-rose-700',
  info: 'bg-cyan-100 text-cyan-800 border-cyan-200 dark:bg-cyan-900 dark:text-cyan-200 dark:border-cyan-700',
  TBD: 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:border-slate-600',
};

const ArchivedDetailView: React.FC<ArchivedDetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [similarIncidents, setSimilarIncidents] = useState<EmbeddingSearchResult[]>([]);
  const [similarLoading, setSimilarLoading] = useState(false);
  
  const navigate = useNavigate();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchRCADetail(incidentId);
      setData(detailData);

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

  const handleUnhide = async () => {
    if (!window.confirm("Would you like to unarchive this report and restore it to the Incident list?")) {
      return;
    }

    try {
      await unhideIncident(incidentId);
      navigate('/', { 
        state: { newlyUnmutedId: incidentId } 
      });
    } catch (error) {
      console.error("Unarchiving failed:", error);
      alert("An error occurred. Please try again.");
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

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 max-w-7xl mx-auto transition-colors duration-300">
      
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
              <span className="text-[10px] font-bold text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-1.5 rounded">
                MUTED
              </span>
            </div>
            <h1 className="text-xl md:text-2xl font-bold text-slate-900 dark:text-white leading-tight break-words">
              {data.title}
            </h1>
          </div>
        </div>
        
        <div className="flex items-center gap-3 self-end md:self-auto">
          
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
            onClick={handleUnhide}
            className="px-4 py-1.5 text-sm text-cyan-600 dark:text-cyan-400 border border-cyan-600 dark:border-cyan-400 rounded hover:bg-cyan-50 dark:hover:bg-cyan-900/20 transition-colors font-medium flex items-center gap-1"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Unarchive
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 opacity-90">
        
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
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-3 flex items-center gap-2">
            Alert Summary
          </h3>
          {/* [수정 포인트] 가독성을 위해 Blue 톤으로 변경 */}
          <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-lg p-5">
            <div className="prose prose-sm prose-slate dark:prose-invert max-w-none text-slate-800 dark:text-slate-200 leading-relaxed">
              <ReactMarkdown 
                remarkPlugins={[remarkGfm]}
                components={{
                  strong: ({ node: _node, ...props }) => <span className="font-bold text-blue-800 dark:text-blue-300" {...props} />,
                  ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 space-y-1 my-2" {...props} />,
                  code: ({ node: _node, ...props }) => (
                    <code className="bg-white dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300 px-1.5 py-0.5 rounded text-xs font-mono font-bold shadow-sm" {...props} />
                  ),
                }}
              >
                {data.analysis_summary || "*No summary information available.*"}
              </ReactMarkdown>
            </div>
          </div>
        </div>

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
                  {data.analysis_detail || "*No detailed analysis content available.*"}
                </ReactMarkdown>
              </div>
            </div>
          </div>
        </div>

        <div className="md:col-span-2 border-t border-slate-200 dark:border-slate-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-4">
            Related Alerts ({data.alerts?.length || 0} Alerts)
          </h3>
          {data.alerts && data.alerts.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-700">
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Alert</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Severity</th>
                    <th className="text-left py-3 px-4 font-semibold text-slate-600 dark:text-slate-300">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {data.alerts.map((alert: AlertItem) => (
                    <tr
                      key={alert.alert_id}
                      onClick={() => navigate(`/alerts/${alert.alert_id}`)}
                      className="border-b border-slate-100 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer"
                    >
                      <td className="py-3 px-4 text-cyan-600 dark:text-cyan-400">{alert.alarm_title}</td>
                      <td className="py-3 px-4">{alert.severity}</td>
                      <td className="py-3 px-4">{alert.status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-slate-500 text-center py-4">No connected alerts.</div>
          )}
        </div>

        <div className="md:col-span-2 border-t border-slate-200 dark:border-slate-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-4">
            Top 3 Similar Incidents
            {similarLoading && <span className="ml-2 text-sm font-normal text-slate-500">(Searching...)</span>}
          </h3>

          {similarIncidents.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {similarIncidents.map((item, idx) => (
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

      </div>
    </div>
  );
};

export default ArchivedDetailView;
