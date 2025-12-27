import { useState, useMemo, useEffect } from 'react';
import { RCAItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import RCATable from './components/RCATable';
import Pagination from './components/Pagination';
import RCADetailView from './components/RCADetailView'; // [추가] 상세 뷰 컴포넌트
import { fetchRCAs } from './utils/api';
import { filterRCAsByTimeRange } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';

function App() {
  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [timeRange, setTimeRange] = useState('Last 1 hours');
  
  // 상세 화면 전환을 위한 State
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(null);

  // [헬퍼 함수] 시간이 없을 경우 현재 시간을 반환 (필터링 통과 및 화면 표시용)
  const getCurrentTimeStr = () => {
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, '0');
    const dd = String(now.getDate()).padStart(2, '0');
    const hh = String(now.getHours()).padStart(2, '0');
    const min = String(now.getMinutes()).padStart(2, '0');
    return `${yyyy}/${mm}/${dd} ${hh}:${min}`;
  };

  // 백엔드에서 RCA 데이터 가져오기
  useEffect(() => {
    const loadRCAs = async () => {
      try {
        setLoading(true);
        setError(null);
        
        // 1. 서버 데이터 가져오기
        const rawData: any[] = await fetchRCAs();
        console.log('🔥 Server Data:', rawData);

        // 2. 데이터 매핑 (서버 데이터 -> 프론트엔드 포맷)
        // [중요] 이 부분이 있어야 리스트에 글자가 제대로 뜹니다.
        const mappedRCAs: RCAItem[] = rawData.map((item) => {
            const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;

            return {
                ...item, // 기존 속성 유지
                incident_id: item.incident_id, 
                alarm_title: item.alarm_title,
                severity: item.severity,
                // 시간이 없으면 현재 시간으로 채워넣기
                time: serverTime ? String(serverTime) : getCurrentTimeStr(), 
            };
        });

        setAllRCAs(mappedRCAs);
      } catch (err) {
        console.error('Failed to load RCAs:', err);
        setError('데이터를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };

    loadRCAs();
  }, []);

  // 타이틀 클릭 핸들러 (상세 화면으로 이동)
  const handleTitleClick = (incident_id: string) => {
    console.log('상세 보기 이동:', incident_id);
    setSelectedIncidentId(incident_id);
  };

  // 뒤로가기 핸들러 (리스트 화면으로 복귀)
  const handleBackToList = () => {
    setSelectedIncidentId(null);
  };

  // 시간 범위에 따라 RCA 필터링
  const filteredRCAs = useMemo(() => {
    return filterRCAsByTimeRange(allRCAs, timeRange);
  }, [allRCAs, timeRange]);

  // 시간 범위가 변경되면 첫 페이지로 리셋
  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange]);

  const totalPages = Math.ceil(filteredRCAs.length / ITEMS_PER_PAGE);

  // 필터링된 RCA 목록에 대해 페이지네이션 적용
  const paginatedRCAs = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return filteredRCAs.slice(startIndex, endIndex);
  }, [filteredRCAs, currentPage]);

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleTimeRangeChange = (newTimeRange: string) => {
    setTimeRange(newTimeRange);
  };

  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        
        {/* [화면 전환 로직] ID가 선택되었으면 상세뷰, 아니면 리스트뷰 */}
        {selectedIncidentId ? (
          // === 상세 화면 ===
          <RCADetailView 
            incidentId={selectedIncidentId} 
            onBack={handleBackToList} 
          />
        ) : (
          // === 리스트 화면 ===
          <div className="bg-white rounded-lg shadow-md p-6">
            {/* Header area with time range selector */}
            <div className="mb-6 flex justify-between items-center">
              <h1 className="text-2xl font-semibold text-gray-800">RCA Dashboard</h1>
              <TimeRangeSelector 
                value={timeRange} 
                onChange={handleTimeRangeChange} 
              />
            </div>

            {/* Loading State */}
            {loading && (
              <div className="flex justify-center items-center py-12">
                <div className="text-gray-600">데이터를 불러오는 중...</div>
              </div>
            )}

            {/* Error State */}
            {error && !loading && (
              <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
                <div className="text-red-800 font-medium">오류 발생</div>
                <div className="text-red-600 text-sm mt-1">{error}</div>
              </div>
            )}

            {/* RCA Table & Pagination */}
            {!loading && !error && (
              <>
                <RCATable 
                  rcas={paginatedRCAs} 
                  onTitleClick={handleTitleClick}
                />
                
                {filteredRCAs.length > 0 ? (
                  <div className="mt-6 flex justify-center">
                    <Pagination
                      currentPage={currentPage}
                      totalPages={totalPages}
                      onPageChange={handlePageChange}
                    />
                  </div>
                ) : (
                  <div className="flex justify-center items-center py-12">
                    <div className="text-gray-500">표시할 RCA가 없습니다.</div>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default App;