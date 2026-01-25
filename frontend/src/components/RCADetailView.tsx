import React, { useCallback, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useLocation, useNavigate } from 'react-router-dom';
import { RCADetail, AlertItem } from '../types';
import { 
  fetchRCADetail, 
  updateRCADetail, 
  hideIncident, 
  unhideIncident,
  resolveIncident, 
  searchSimilarIncidents, 
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
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700',
  critical: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700',
  info: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700',
  TBD: 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600',
};

const RCADetailView: React.FC<RCADetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [similarIncidents, setSimilarIncidents] = useState<EmbeddingSearchResult[]>([]);
  const [similarLoading, setSimilarLoading] = useState(false);

  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<Partial<RCADetail>>({});

  const location = useLocation();
  const navigate = useNavigate();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchRCADetail(incidentId);
      setData(detailData);
      setEditForm(detailData);

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
      alert('성공적으로 수정되었습니다.');
    } catch (err) {
      console.error(err);
      alert('저장에 실패했습니다.');
    }
  };

  const handleCancel = () => {
    if (data) {
      setEditForm(data);
    }
    setIsEditing(false);
  };

  const handleHide = async () => {
    if (!window.confirm("정말 이 리포트를 목록에서 숨기시겠습니까?")) {
      return;
    }

    try {
      await hideIncident(incidentId);
      alert("성공적으로 숨겨졌습니다. Mute Dashboard로 이동합니다.");
      navigate('/muted', { 
        state: { newlyMutedId: incidentId } 
      });
    } catch (error) {
      console.error("숨기기 실패:", error);
      alert("오류가 발생했습니다. 다시 시도해주세요.");
    }
  };

  const handleUnhide = async () => {
    if (!window.confirm("이 리포트의 숨김을 해제하시겠습니까?")) {
      return;
    }

    try {
      await unhideIncident(incidentId);
      alert("숨김이 해제되었습니다.");
      loadDetail(); 
    } catch (error) {
      console.error("숨기기 해제 실패:", error);
      alert("오류가 발생했습니다. 다시 시도해주세요.");
    }
  };

  const handleResolve = async () => {
    if (!window.confirm("이 인시던트를 종료하시겠습니까?\n종료 후 AI가 최종 분석을 수행합니다.")) {
      return;
    }

    try {
      await resolveIncident(incidentId);
      alert("인시던트가 종료되었습니다.\nAI 분석이 완료되면 결과가 업데이트됩니다.");
      await loadDetail();
    } catch (error) {
      console.error("종료 실패:", error);
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
  const isHidden = data.is_hidden ?? false; 

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
                ID: {data.incident_id}
              </span>
              {isHidden && (
                <span className="text-[10px] font-bold text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 px-1.5 rounded">
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
                  className="text-lg font-bold text-gray-900 dark:text-white bg-white dark:bg-gray-700 border border-blue-400 rounded px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            ) : (
              <h1 className="text-xl md:text-2xl font-bold text-gray-900 dark:text-white leading-tight break-words">
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
                className="px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {editForm.severity === 'TBD' && <option value="TBD" disabled>TBD (선택해주세요)</option>}
                <option value="critical">critical</option>
                <option value="warning">warning</option>
                <option value="info">info</option>
              </select>

              <button 
                onClick={handleSave}
                className="px-4 py-2 bg-blue-600 text-white text-sm font-semibold rounded hover:bg-blue-700 transition shadow-sm"
              >
                Save
              </button>
              <button 
                onClick={handleCancel}
                className="px-4 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 text-sm font-semibold rounded hover:bg-gray-50 dark:hover:bg-gray-600 transition shadow-sm"
              >
                Cancel
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              
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

              {!isResolved && (
                <button
                  onClick={handleResolve}
                  className="px-4 py-1.5 text-sm text-green-600 dark:text-green-400 border border-green-600 dark:border-green-400 rounded hover:bg-green-50 dark:hover:bg-green-900/20 transition-colors font-medium"
                >
                  종료
                </button>
              )}

              {isHidden ? (
                <button
                  onClick={handleUnhide}
                  className="px-4 py-1.5 text-sm text-blue-600 dark:text-blue-400 border border-blue-600 dark:border-blue-400 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors font-medium"
                >
                  숨기기 해제
                </button>
              ) : (
                <button
                  onClick={handleHide}
                  className="px-4 py-1.5 text-sm text-red-600 dark:text-red-400 border border-red-600 dark:border-red-400 rounded hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                >
                  숨기기
                </button>
              )}
              
              {!isHidden && (
                <button 
                  onClick={() => setIsEditing(true)}
                  className="px-4 py-1.5 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm font-semibold rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition"
                >
                  Edit
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        
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
            📋 Incident Summary
          </h3>
          
          {isEditing ? (
            <textarea
              name="analysis_summary"
              value={editForm.analysis_summary || ''}
              onChange={handleInputChange}
              rows={5}
              placeholder="여기에 마크다운 형식으로 요약을 작성하세요..."
              className="w-full p-4 border border-blue-400 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow shadow-sm"
            />
          ) : (
            // [수정 포인트] 노란색 -> 깔끔한 블루/그레이 톤으로 변경 + 코드 블록 고대비 적용
            <div className="bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-lg p-5 transition-colors">
              <div className="prose prose-sm prose-slate dark:prose-invert max-w-none text-gray-800 dark:text-gray-200 leading-relaxed">
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
                  {data.analysis_summary || "*요약 정보가 없습니다.*"}
                </ReactMarkdown>
              </div>
            </div>
          )}
        </div>

        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📝 Incident Analysis
          </h3>

          {isEditing ? (
            <textarea
              name="analysis_detail"
              value={editForm.analysis_detail || ''}
              onChange={handleInputChange}
              rows={15}
              placeholder="여기에 상세 분석 내용을 마크다운으로 작성하세요..."
              className="w-full p-4 border border-blue-400 rounded-lg bg-gray-900 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow shadow-sm"
            />
          ) : (
            <div className="bg-gray-900 border border-gray-700 rounded-lg overflow-hidden shadow-sm">
              <div className="p-6 overflow-x-auto">
                <div className="prose prose-sm prose-invert max-w-none font-mono leading-relaxed">
                  <ReactMarkdown 
                    remarkPlugins={[remarkGfm]}
                    components={{
                      h1: ({ node: _node, ...props }) => <h1 className="text-2xl font-extrabold text-blue-400 mt-8 mb-6 pb-2 border-b border-gray-700 flex items-center gap-2 [&_strong]:text-blue-400" {...props} />,
                      h2: ({ node: _node, ...props }) => <h2 className="text-xl font-bold text-indigo-300 mt-8 mb-4 pl-3 border-l-4 border-indigo-500 [&_strong]:text-indigo-300" {...props} />,
                      h3: ({ node: _node, ...props }) => <h3 className="text-lg font-semibold text-sky-300 mt-6 mb-3 ml-1 [&_strong]:text-sky-300" {...props} />,
                      strong: ({ node: _node, ...props }) => <span className="font-bold text-amber-400" {...props} />,
                      ul: ({ node: _node, ...props }) => <ul className="list-disc pl-6 space-y-2 my-2 text-gray-300 leading-relaxed" {...props} />,
                      code: ({ node: _node, ...props }) => (
                        <code className="bg-gray-800 text-pink-400 px-1.5 py-0.5 rounded text-sm font-mono border border-gray-700 mx-1" {...props} />
                      ),
                      p: ({ node: _node, ...props }) => <p className="mb-4 text-gray-300 leading-relaxed" {...props} />,
                      a: ({ node: _node, ...props }) => <a className="text-blue-400 hover:text-blue-300 hover:underline transition-colors" target="_blank" rel="noopener noreferrer" {...props} />,
                    }}
                  >
                    {data.analysis_detail || "*상세 분석 내용이 없습니다.*"}
                  </ReactMarkdown>
                </div>
              </div>
            </div>
          )}
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
                    <th className="text-left py-3 px-4 font-semibold text-gray-600 dark:text-gray-300">발생 시간</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-600 dark:text-gray-300">해결 시간</th>
                  </tr>
                </thead>
                <tbody>
                  {data.alerts.map((alert: AlertItem) => (
                    <tr
                      key={alert.alert_id}
                      onClick={() => navigate(`/alerts/${alert.alert_id}`)}
                      className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer"
                    >
                      <td className="py-3 px-4">
                        <div className="font-medium text-blue-600 dark:text-blue-400 hover:underline">{alert.alarm_title}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 font-mono mt-0.5">{alert.alert_id}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-bold border ${severityStyles[alert.severity] || severityStyles.info}`}>
                          {alert.severity}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-bold ${
                          alert.status === 'resolved'
                            ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                            : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                        }`}>
                          {alert.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono text-xs text-gray-600 dark:text-gray-300">
                        {formatTime(alert.fired_at)}
                      </td>
                      <td className="py-3 px-4 font-mono text-xs text-gray-600 dark:text-gray-300">
                        {alert.resolved_at ? formatTime(alert.resolved_at) : '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center border border-dashed border-gray-300 dark:border-gray-600">
              <p className="text-gray-500 dark:text-gray-400">연결된 Alert가 없습니다.</p>
            </div>
          )}
        </div>

        <div className="md:col-span-2 border-t border-gray-200 dark:border-gray-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-4">
            🔗 Top 3 유사 인시던트
            {similarLoading && <span className="ml-2 text-sm font-normal text-gray-500">(검색 중...)</span>}
          </h3>

          {similarIncidents.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {similarIncidents.map((item: EmbeddingSearchResult, idx: number) => (
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

export default RCADetailView;
