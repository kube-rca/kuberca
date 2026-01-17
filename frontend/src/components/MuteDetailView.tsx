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

interface MuteDetailViewProps {
  incidentId: string;
  onBack: () => void;
}

const severityStyles: Record<string, string> = {
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
};

const MuteDetailView: React.FC<MuteDetailViewProps> = ({ incidentId, onBack }) => {
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
          console.error('유사 인시던트 검색 실패:', searchErr);
        } finally {
          setSimilarLoading(false);
        }
      }
    } catch (err) {
      setError('데이터를 불러오지 못했습니다.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [incidentId]);

  useEffect(() => {
    loadDetail();
  }, [loadDetail]);

  const handleUnhide = async () => {
    if (!window.confirm("이 리포트의 숨김을 해제하고 Incident 목록으로 복구하시겠습니까?")) {
      return;
    }

    try {
      await unhideIncident(incidentId);
      navigate('/', { 
        state: { newlyUnmutedId: incidentId } 
      });
    } catch (error) {
      console.error("숨기기 해제 실패:", error);
      alert("오류가 발생했습니다. 다시 시도해주세요.");
    }
  };

  const formatTime = (isoString?: string | null) => {
    if (!isoString) return '-';
    return isoString.replace('T', ' ').split('.')[0];
  };

  if (loading) return <div className="p-12 text-center text-gray-500 dark:text-gray-400">상세 정보를 불러오는 중...</div>;
  if (error || !data) return <div className="p-12 text-center text-red-500 bg-red-50 dark:bg-red-900/20 rounded-lg m-4">{error}</div>;

  const isResolved = !!data.resolved_at;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-5xl mx-auto transition-colors duration-300 border border-gray-200 dark:border-gray-700">
      
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
                ID: {data.incident_id}
              </span>
              <span className="text-[10px] font-bold text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 px-1.5 rounded">
                MUTED
              </span>
            </div>
            <h1 className="text-xl md:text-2xl font-bold text-gray-900 dark:text-white leading-tight">
              {data.title}
            </h1>
          </div>
        </div>
        
        <div className="flex items-center gap-3 self-end md:self-auto">
          
          <span 
            className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0 
              ${isResolved 
                ? 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700' 
                : 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700'
              }`}
          >
            {isResolved ? 'Resolved' : 'Ongoing'}
          </span>

          <span 
            className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0 
              ${severityStyles[data.severity || 'info'] || severityStyles.info}`}
          >
            {data.severity}
          </span>

          <button
            onClick={handleUnhide}
            className="px-4 py-1.5 text-sm text-blue-600 dark:text-blue-400 border border-blue-600 dark:border-blue-400 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors font-medium flex items-center gap-1"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            숨기기 해제
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 opacity-90">
        
        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             🔥 발생 시간
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium font-mono">
            {formatTime(data.fired_at)}
          </div>
        </div>

        <div className="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg border border-gray-100 dark:border-gray-700">
          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide flex items-center gap-1">
             ✅ 해결 시간
          </div>
          <div className="text-gray-900 dark:text-gray-100 font-medium font-mono">
            {data.resolved_at ? formatTime(data.resolved_at) : <span className="text-red-500 font-bold">Ongoing</span>}
          </div>
        </div>

        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📋 Alert Summary
          </h3>
          {/* [수정 포인트] 가독성을 위해 Blue 톤으로 변경 */}
          <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-lg p-5">
            <div className="prose prose-sm prose-slate dark:prose-invert max-w-none text-gray-800 dark:text-gray-200 leading-relaxed">
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
                {data.analysis_summary || "*요약 정보가 없습니다.*"}
              </ReactMarkdown>
            </div>
          </div>
        </div>

        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📝 Alert Analysis
          </h3>
          <div className="bg-gray-900 border border-gray-700 rounded-lg overflow-hidden shadow-sm">
            <div className="p-6 overflow-x-auto">
              <div className="prose prose-sm prose-invert max-w-none font-mono leading-relaxed">
                <ReactMarkdown 
                  remarkPlugins={[remarkGfm]}
                components={{
                  h1: ({ node: _node, ...props }) => <h1 className="text-xl font-bold text-blue-400 mt-6 mb-4 border-b border-gray-700 pb-2" {...props} />,
                  h2: ({ node: _node, ...props }) => <h2 className="text-lg font-bold text-blue-300 mt-5 mb-3" {...props} />,
                  h3: ({ node: _node, ...props }) => <h3 className="text-md font-bold text-blue-200 mt-4 mb-2" {...props} />,
                  strong: ({ node: _node, ...props }) => <span className="font-bold text-yellow-400" {...props} />,
                  ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 space-y-1 my-2 text-gray-300" {...props} />,
                  code: ({ node: _node, ...props }) => (
                    <code className="bg-gray-800 text-green-400 px-1 py-0.5 rounded text-xs" {...props} />
                  ),
                  p: ({ node: _node, ...props }) => <p className="mb-4 text-gray-300" {...props} />,
                  a: ({ node: _node, ...props }) => <a className="text-blue-400 hover:underline" target="_blank" rel="noopener noreferrer" {...props} />,
                }}
                >
                  {data.analysis_detail || "*상세 분석 내용이 없습니다.*"}
                </ReactMarkdown>
              </div>
            </div>
          </div>
        </div>

        <div className="md:col-span-2 border-t border-gray-200 dark:border-gray-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-4">
            🚨 연결된 Alert ({data.alerts?.length || 0}개)
          </h3>
          {data.alerts && data.alerts.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-700">
                    <th className="text-left py-3 px-4 font-semibold text-gray-600 dark:text-gray-300">Alert</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-600 dark:text-gray-300">Severity</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {data.alerts.map((alert: AlertItem) => (
                    <tr
                      key={alert.alert_id}
                      onClick={() => navigate(`/alerts/${alert.alert_id}`)}
                      className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer"
                    >
                      <td className="py-3 px-4 text-blue-600 dark:text-blue-400">{alert.alarm_title}</td>
                      <td className="py-3 px-4">{alert.severity}</td>
                      <td className="py-3 px-4">{alert.status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-gray-500 text-center py-4">연결된 Alert가 없습니다.</div>
          )}
        </div>

        <div className="md:col-span-2 border-t border-gray-200 dark:border-gray-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-4">
            🔗 Top 3 유사 인시던트
            {similarLoading && <span className="ml-2 text-sm font-normal text-gray-500">(검색 중...)</span>}
          </h3>

          {similarIncidents.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {similarIncidents.map((item, idx) => (
                <div
                  key={idx}
                  className="bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 p-4 rounded-lg shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => navigate(`/incidents/${item.incident_id}`)}
                >
                  <div className="mb-2 flex justify-between items-center">
                    <span className="text-xs font-mono text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded">
                      {item.incident_id}
                    </span>
                    <span className="text-xs font-bold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-2 py-0.5 rounded-full">
                      {Math.round(item.similarity * 100)}% 유사
                    </span>
                  </div>
                  <div className="text-sm font-medium text-gray-800 dark:text-gray-200 line-clamp-3" title={item.incident_summary}>
                    {item.incident_summary}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center border border-dashed border-gray-300 dark:border-gray-600">
              <p className="text-gray-500 dark:text-gray-400">
                {!data.analysis_summary ? '분석 요약이 없어 유사 인시던트를 검색할 수 없습니다.' : '유사한 인시던트 내역이 없습니다.'}
              </p>
            </div>
          )}
        </div>

      </div>
    </div>
  );
};

export default MuteDetailView;
