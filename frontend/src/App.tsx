import { useState, useMemo, useEffect } from 'react';
import { RCAItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import RCATable from './components/RCATable';
import Pagination from './components/Pagination';
import { fetchRCAs } from './utils/api';
import { filterRCAsByTimeRange } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';

function App() {
  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [timeRange, setTimeRange] = useState('Last 1 hours');

  // 백엔드에서 RCA 데이터 가져오기
  useEffect(() => {
    const loadRCAs = async () => {
      try {
        setLoading(true);
        setError(null);
        const rcas = await fetchRCAs();
        setAllRCAs(rcas);
      } catch (err) {
        console.error('Failed to load RCAs:', err);

      } finally {
        setLoading(false);
      }
    };

    loadRCAs();
  }, []);

  // [추가됨] 타이틀 클릭 핸들러
  const handleTitleClick = (incident_id: string) => {
    // 여기에 상세 페이지 이동 로직이나 모달 띄우는 로직을 넣으세요.
    console.log('클릭된 ID:', incident_id);
    alert(`상세 정보 보기: ${incident_id}`); 
    // 예: navigate(\`/rca/\${incident_id}\`);
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

          {/* RCA Table */}
          {!loading && !error && (
            <>
              <RCATable 
                rcas={paginatedRCAs} 
                onTitleClick={handleTitleClick}
              />
              {/* Pagination */}
              <div className="mt-6 flex justify-center">
                <Pagination
                  currentPage={currentPage}
                  totalPages={totalPages}
                  onPageChange={handlePageChange}
                />
              </div>
            </>
          )}

          {/* Empty State */}
          {!loading && !error && filteredRCAs.length === 0 && (
            <div className="flex justify-center items-center py-12">
              <div className="text-gray-500">표시할 RCA가 없습니다.</div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;

