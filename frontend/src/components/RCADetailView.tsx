import React, { useCallback, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useLocation } from 'react-router-dom';
import { RCADetail, SimilarIncident } from '../types';
import { fetchRCADetail, updateRCADetail, hideIncident } from '../utils/api';

interface RCADetailViewProps {
  incidentId: string;
  onBack: () => void;
}

const formatSeverity = (text?: string) => {
  if (!text) return '';
  return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
};

const RCADetailView: React.FC<RCADetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<Partial<RCADetail>>({});

  // 리스트에서 넘어온 state(autoEdit) 확인용
  const location = useLocation();

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchRCADetail(incidentId);
      setData(detailData);
      setEditForm(detailData);
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

  // [자동 편집 모드] 리스트에서 'Edit' 버튼 누르고 들어왔을 때 실행
  useEffect(() => {
    if (location.state && (location.state as any).autoEdit) {
      setIsEditing(true);
      // 상태 초기화 (새로고침 시 재실행 방지)
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
      alert("성공적으로 숨겨졌습니다.");
      onBack(); // 목록으로 돌아가기 (props로 받은 함수 실행)
    } catch (error) {
      console.error("숨기기 실패:", error);
      alert("오류가 발생했습니다. 다시 시도해주세요.");
    }
  };

  const formatTime = (isoString?: string | null) => {
    if (!isoString) return '-';
    return isoString.replace('T', ' ').split('.')[0];
  };

  const getBadgeColor = (severity?: string, resolvedAt?: string | null) => {
    if (resolvedAt) return 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700';

    const s = severity?.toLowerCase() || 'info';
    if (s === 'critical') return 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900 dark:text-red-200 dark:border-red-700';
    if (s === 'warning') return 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700';
    if (s === 'resolved') return 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-700';
    return 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700';
  };

  const getDisplaySeverity = (severity?: string, resolvedAt?: string | null) => {
    if (resolvedAt) return 'Resolved';
    return formatSeverity(severity);
  };

  if (loading) return <div className="p-12 text-center text-gray-500 dark:text-gray-400">상세 정보를 불러오는 중...</div>;
  if (error || !data) return <div className="p-12 text-center text-red-500 bg-red-50 dark:bg-red-900/20 rounded-lg m-4">{error}</div>;

  const similarList: SimilarIncident[] = data.similar_incidents || [];

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-5xl mx-auto transition-colors duration-300">
      
      {/* 헤더 영역 */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-8 border-b border-gray-200 dark:border-gray-700 pb-6 gap-4">
        
        {/* [왼쪽 그룹] Back 버튼 + ID + 제목 */}
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
            </div>

            {isEditing ? (
              <div className="flex flex-col gap-1">
                <input
                  type="text"
                  name="alarm_title"
                  value={editForm.alarm_title || ''}
                  onChange={handleInputChange}
                  className="text-lg font-bold text-gray-900 dark:text-white bg-white dark:bg-gray-700 border border-blue-400 rounded px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            ) : (
              <h1 className="text-xl md:text-2xl font-bold text-gray-900 dark:text-white leading-tight">
                {data.alarm_title}
              </h1>
            )}
          </div>
        </div>
        
        {/* [오른쪽 그룹] 버튼들 */}
        <div className="flex items-center gap-3 self-end md:self-auto">
          {isEditing ? (
            <div className="flex items-center gap-2">
              <select
                name="severity"
                value={editForm.severity}
                onChange={handleInputChange}
                className="px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="Resolved">Resolved</option>
                <option value="Critical">Critical</option>
                <option value="Warning">Warning</option>
                <option value="Info">Info</option>
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
              
              {/* [위치: 우측 그룹] 뱃지 */}
              <span className={`px-3 py-1.5 rounded-full text-xs font-bold border flex-shrink-0 ${getBadgeColor(data.severity, data.resolved_at)}`}>
                {getDisplaySeverity(data.severity, data.resolved_at)}
              </span>

              {/* 숨기기 버튼 */}
              <button 
                onClick={handleHide}
                className="px-4 py-1.5 text-sm text-red-600 dark:text-red-400 border border-red-600 dark:border-red-400 rounded hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
              >
                숨기기
              </button>

              {/* Edit 버튼 */}
              <button 
                onClick={() => setIsEditing(true)}
                className="px-4 py-1.5 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm font-semibold rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition"
              >
                Edit
              </button>
            </div>
          )}
        </div>
      </div>

      {/* 나머지 본문 영역 */}
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
            {data.resolved_at ? formatTime(data.resolved_at) : <span className="text-blue-500 font-bold">Ongoing</span>}
          </div>
        </div>

        {/* 요약 */}
        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📋 인시던트 요약
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
            <div className="bg-yellow-50 dark:bg-yellow-900/10 border border-yellow-200 dark:border-yellow-800/30 rounded-lg p-5">
              <div className="prose prose-sm prose-yellow dark:prose-invert max-w-none text-gray-800 dark:text-gray-200">
                <ReactMarkdown 
                  remarkPlugins={[remarkGfm]}
                  components={{
                    strong: ({node, ...props}) => <span className="font-bold text-gray-900 dark:text-white" {...props} />,
                    ul: ({node, ...props}) => <ul className="list-disc pl-5 space-y-1 my-2" {...props} />,
                    code: ({node, ...props}) => (
                      <code className="bg-yellow-100 dark:bg-yellow-900/40 text-yellow-800 dark:text-yellow-200 px-1.5 py-0.5 rounded text-xs font-mono" {...props} />
                    ),
                  }}
                >
                  {data.analysis_summary || "*요약 정보가 없습니다.*"}
                </ReactMarkdown>
              </div>
            </div>
          )}
        </div>

        {/* 상세 리포트 */}
        <div className="md:col-span-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
            📝 상세 분석 리포트
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
                      h1: ({node, ...props}) => <h1 className="text-xl font-bold text-blue-400 mt-6 mb-4 border-b border-gray-700 pb-2" {...props} />,
                      h2: ({node, ...props}) => <h2 className="text-lg font-bold text-blue-300 mt-5 mb-3" {...props} />,
                      h3: ({node, ...props}) => <h3 className="text-md font-bold text-blue-200 mt-4 mb-2" {...props} />,
                      strong: ({node, ...props}) => <span className="font-bold text-yellow-400" {...props} />,
                      ul: ({node, ...props}) => <ul className="list-disc pl-5 space-y-1 my-2 text-gray-300" {...props} />,
                      code: ({node, ...props}) => (
                        <code className="bg-gray-800 text-green-400 px-1 py-0.5 rounded text-xs" {...props} />
                      ),
                      p: ({node, ...props}) => <p className="mb-4 text-gray-300" {...props} />,
                      a: ({node, ...props}) => <a className="text-blue-400 hover:underline" target="_blank" rel="noopener noreferrer" {...props} />,
                    }}
                  >
                    {data.analysis_detail || "*상세 분석 내용이 없습니다.*"}
                  </ReactMarkdown>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 유사 인시던트 */}
        <div className="md:col-span-2 border-t border-gray-200 dark:border-gray-700 pt-6 mt-2">
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100 mb-4">
            🔗 Top 3 유사 인시던트
          </h3>
          
          {similarList.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {similarList.map((item, idx) => (
                <div key={idx} className="bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 p-4 rounded-lg shadow-sm hover:shadow-md transition-shadow">
                  <div className="mb-2 flex justify-between items-center">
                      <span className="text-xs font-mono text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded">
                        {item.incident_id}
                      </span>
                      {item.score && (
                        <span className="text-xs font-bold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-2 py-0.5 rounded-full">
                          {item.score}% 유사
                        </span>
                      )}
                  </div>
                  <div className="text-sm font-medium text-gray-800 dark:text-gray-200 line-clamp-2" title={item.alarm_title}>
                    {item.alarm_title}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center border border-dashed border-gray-300 dark:border-gray-600">
              <p className="text-gray-500 dark:text-gray-400">유사한 인시던트 내역이 없습니다.</p>
            </div>
          )}
        </div>

      </div>
    </div>
  );
};

export default RCADetailView;