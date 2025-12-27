import React, { useEffect, useState } from 'react';
import { RCADetail, SimilarIncident } from '../types'; // SimilarIncident import 확인
import { fetchRCADetail } from '../utils/api';

interface RCADetailViewProps {
  incidentId: string;
  onBack: () => void;
}

const RCADetailView: React.FC<RCADetailViewProps> = ({ incidentId, onBack }) => {
  const [data, setData] = useState<RCADetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadDetail = async () => {
      try {
        setLoading(true);
        const detailData = await fetchRCADetail(incidentId);
        setData(detailData);
      } catch (err) {
        setError('데이터를 불러오지 못했습니다.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadDetail();
  }, [incidentId]);

  if (loading) return <div className="p-8 text-center text-gray-500">상세 정보를 불러오는 중...</div>;
  if (error || !data) return <div className="p-8 text-center text-red-500">{error}</div>;

  // [안전 장치] 백엔드에서 null이나 undefined를 줄 경우 빈 배열로 처리
  const similarList: SimilarIncident[] = data.similar_incidents || [];

  return (
    <div className="bg-white rounded-lg shadow-md p-6 max-w-4xl mx-auto">
      {/* 1. 헤더 */}
      <div className="flex items-center justify-between mb-6 border-b pb-4">
        <div className="flex items-center gap-4">
          <button 
            onClick={onBack}
            className="text-gray-500 hover:text-gray-700 font-medium px-3 py-1 border rounded hover:bg-gray-50 transition"
          >
            ← Back
          </button>
          <h1 className="text-xl font-bold text-gray-900">{data.alarm_title}</h1>
        </div>
        <div className="flex gap-2">
          <span className="px-3 py-1 rounded-full text-sm font-semibold bg-blue-100 text-blue-800">
            {data.severity}
          </span>
          <span className="px-3 py-1 rounded-full text-sm font-semibold bg-gray-100 text-gray-800">
            {data.status}
          </span>
        </div>
      </div>

      {/* 2. 상세 정보 그리드 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
        
        {/* 기본 정보 */}
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

        {/* 분석 요약 */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-2">📋 분석 요약</h3>
          <div className="text-gray-700 bg-yellow-50 p-3 rounded border-l-4 border-yellow-400 min-h-[60px]">
            {data.analysis_summary || "분석 요약 정보가 없습니다."}
          </div>
        </div>

        {/* 상세 리포트 */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-3">📝 상세 분석 리포트</h3>
          <div className="bg-gray-900 text-gray-100 p-4 rounded-md font-mono text-sm leading-relaxed whitespace-pre-wrap min-h-[100px]">
            {data.analysis_detail || "상세 분석 내용이 없습니다."}
          </div>
        </div>

        {/* [핵심] Top 3 유사 인시던트 */}
        <div className="md:col-span-2 border border-gray-200 rounded-md p-5">
          <h3 className="text-lg font-semibold text-gray-800 mb-3">🔗 Top 3 유사 인시던트</h3>
          
          <div className="bg-gray-50 p-4 rounded-md">
            {similarList.length > 0 ? (
              // 데이터가 있을 때: 3개 카드 그리드 출력
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {similarList.map((item, idx) => (
                  <div 
                    key={item.incident_id || idx} // ID가 없으면 인덱스 사용
                    className="bg-white border border-gray-200 p-4 rounded shadow-sm hover:shadow-md hover:border-blue-300 transition cursor-pointer flex flex-col justify-between h-full"
                    onClick={() => console.log('유사 인시던트 클릭:', item.incident_id)}
                  >
                    <div className="mb-2">
                      <div className="flex justify-between items-start mb-1">
                        <span className="text-xs font-mono text-gray-500 bg-gray-100 px-1 rounded">
                          {item.incident_id}
                        </span>
                        {/* score가 있을 때만 뱃지 표시 */}
                        {item.score !== undefined && (
                           <span className="text-xs font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full">
                             {item.score}% 유사
                           </span>
                        )}
                      </div>
                      <div className="text-sm font-medium text-gray-800 line-clamp-2 leading-snug">
                        {item.alarm_title}
                      </div>
                    </div>
                    <div className="text-xs text-gray-400 text-right mt-2">
                      Click to view →
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              // 데이터가 없을 때: 안내 문구 출력
              <div className="flex flex-col items-center justify-center py-8 text-gray-400">
                <svg className="w-10 h-10 mb-2 opacity-20" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
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