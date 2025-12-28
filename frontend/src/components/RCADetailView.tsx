import React, { useCallback, useEffect, useState } from 'react';
import { RCADetail, SimilarIncident } from '../types';
import { fetchRCADetail, updateRCADetail } from '../utils/api'; // [추가] update 함수 import

interface RCADetailViewProps {
  incidentId: string;
  onBack: () => void;
}

const RCADetailView: React.FC<RCADetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // [신규] 편집 모드 상태 관리
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<Partial<RCADetail>>({});

  const loadDetail = useCallback(async () => {
    try {
      setLoading(true);
      const detailData = await fetchRCADetail(incidentId);
      setData(detailData);
      setEditForm(detailData); // 편집을 대비해 폼 데이터 초기화
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

  // [신규] 입력값 변경 핸들러
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditForm((prev) => ({ ...prev, [name]: value }));
  };

  // [신규] 저장 버튼 클릭 핸들러
  const handleSave = async () => {
    if (!data) return;
    try {
      // API 호출하여 서버 업데이트
      await updateRCADetail(incidentId, editForm);
      
      // 성공 시 로컬 데이터 업데이트 및 편집 모드 종료
      setData({ ...data, ...editForm } as RCADetail);
      setIsEditing(false);
      alert('성공적으로 수정되었습니다.');
    } catch (err) {
      console.error(err);
      alert('저장에 실패했습니다.');
    }
  };

  // [신규] 취소 버튼 클릭 핸들러
  const handleCancel = () => {
    setEditForm(data!); // 원래 데이터로 복구
    setIsEditing(false);
  };

  if (loading) return <div className="p-8 text-center text-gray-500">상세 정보를 불러오는 중...</div>;
  if (error || !data) return <div className="p-8 text-center text-red-500">{error}</div>;

  const similarList: SimilarIncident[] = data.similar_incidents || [];

  return (
    <div className="bg-white rounded-lg shadow-md p-6 max-w-4xl mx-auto">
      {/* 1. 상단 헤더 (Back 버튼, 제목, 저장/취소 버튼) */}
      <div className="flex items-center justify-between mb-6 border-b pb-4">
        <div className="flex items-center gap-4 flex-1">
          <button 
            onClick={onBack}
            className="text-gray-500 hover:text-gray-700 font-medium px-3 py-1 border rounded hover:bg-gray-50 transition"
          >
            ← Back
          </button>
          
          {/* 제목: 편집 모드일 때 Input, 아닐 때 텍스트 */}
          {isEditing ? (
            <input
              type="text"
              name="alarm_title"
              value={editForm.alarm_title || ''}
              onChange={handleInputChange}
              className="text-xl font-bold text-gray-900 border border-blue-300 rounded px-2 py-1 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          ) : (
            <h1 className="text-xl font-bold text-gray-900">{data.alarm_title}</h1>
          )}
        </div>
        
        {/* 우측 버튼 그룹 */}
        <div className="flex items-center gap-3 ml-4">
          {isEditing ? (
            <>
              {/* Severity 선택 (편집 모드) */}
              <select
                name="severity"
                value={editForm.severity}
                onChange={handleInputChange}
                className="px-3 py-1 rounded border border-gray-300 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="Resolved">Resolved</option>
                <option value="Critical">Critical</option>
                <option value="Warning">Warning</option>
                <option value="Info">Info</option>
              </select>

              <button 
                onClick={handleSave}
                className="px-4 py-1.5 bg-blue-600 text-white text-sm font-semibold rounded hover:bg-blue-700 transition"
              >
                Save
              </button>
              <button 
                onClick={handleCancel}
                className="px-4 py-1.5 bg-gray-200 text-gray-700 text-sm font-semibold rounded hover:bg-gray-300 transition"
              >
                Cancel
              </button>
            </>
          ) : (
            <>
              {/* 뱃지 (조회 모드) */}
              <div className="flex gap-2">
                <span className="px-3 py-1 rounded-full text-sm font-semibold bg-blue-100 text-blue-800">
                  {data.severity}
                </span>
                <span className="px-3 py-1 rounded-full text-sm font-semibold bg-gray-100 text-gray-800">
                  {data.status}
                </span>
              </div>
              
              {/* Edit 버튼 */}
              <button 
                onClick={() => setIsEditing(true)}
                className="ml-2 px-4 py-1.5 border border-blue-600 text-blue-600 text-sm font-semibold rounded hover:bg-blue-50 transition"
              >
                Edit
              </button>
            </>
          )}
        </div>
      </div>

      {/* 2. 상세 정보 그리드 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
        
        {/* ID & Time (수정 불가 - Read Only) */}
        <div className="bg-gray-50 p-4 rounded-md">
          <div className="text-sm text-gray-500 mb-1">Incident ID</div>
          <div className="font-mono text-gray-900 font-medium">{data.incident_id}</div>
        </div>
        <div className="bg-gray-50 p-4 rounded-md">
          <div className="text-sm text-gray-500 mb-1">발생 시간 (Fired At)</div>
          <div className="text-gray-900">
            {data.fired_at ? data.fired_at.replace('T', ' ').split('.')[0] : '-'}
          </div>
        </div>

        {/* 분석 요약 (수정 가능) */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-2">📋 분석 요약</h3>
          {isEditing ? (
            <textarea
              name="analysis_summary"
              value={editForm.analysis_summary || ''}
              onChange={handleInputChange}
              rows={3}
              className="w-full p-3 border border-blue-300 rounded bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          ) : (
            <div className="text-gray-700 bg-yellow-50 p-3 rounded border-l-4 border-yellow-400 min-h-[60px]">
              {data.analysis_summary || "분석 요약 정보가 없습니다."}
            </div>
          )}
        </div>

        {/* 상세 리포트 (수정 가능) */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-3">📝 상세 분석 리포트</h3>
          {isEditing ? (
            <textarea
              name="analysis_detail"
              value={editForm.analysis_detail || ''}
              onChange={handleInputChange}
              rows={10}
              className="w-full p-3 border border-blue-300 rounded bg-white text-gray-900 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          ) : (
            <div className="bg-gray-900 text-gray-100 p-4 rounded-md font-mono text-sm leading-relaxed whitespace-pre-wrap min-h-[100px]">
              {data.analysis_detail || "상세 분석 내용이 없습니다."}
            </div>
          )}
        </div>

        {/* Top 3 유사 인시던트 (수정 불가) */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-3">🔗 Top 3 유사 인시던트</h3>
          <div className="bg-gray-50 p-4 rounded-md">
            {similarList.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {similarList.map((item, idx) => (
                  <div key={idx} className="bg-white border border-gray-200 p-4 rounded shadow-sm">
                    <div className="mb-2 flex justify-between">
                       <span className="text-xs font-mono text-gray-500 bg-gray-100 px-1 rounded">{item.incident_id}</span>
                       {item.score && <span className="text-xs font-bold text-blue-600">{item.score}% 유사</span>}
                    </div>
                    <div className="text-sm font-medium text-gray-800 line-clamp-2">{item.alarm_title}</div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-gray-400">
                <span>유사한 인시던트 내역이 없습니다.</span>
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
};

export default RCADetailView;
