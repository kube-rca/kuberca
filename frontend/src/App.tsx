import { useState, useMemo, useEffect } from 'react';
import { AlertItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import AlertsTable from './components/AlertsTable';
import Pagination from './components/Pagination';
import { generateMockAlerts } from './utils/mockData';
import { filterAlertsByTimeRange } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';

function App() {
  const [allAlerts] = useState<AlertItem[]>(generateMockAlerts());
  const [currentPage, setCurrentPage] = useState(1);
  const [timeRange, setTimeRange] = useState('Last 1 hours');

  // 시간 범위에 따라 알림 필터링
  const filteredAlerts = useMemo(() => {
    return filterAlertsByTimeRange(allAlerts, timeRange);
  }, [allAlerts, timeRange]);

  // 시간 범위가 변경되면 첫 페이지로 리셋
  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange]);

  const totalPages = Math.ceil(filteredAlerts.length / ITEMS_PER_PAGE);

  // 필터링된 알림 목록에 대해 페이지네이션 적용
  const paginatedAlerts = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return filteredAlerts.slice(startIndex, endIndex);
  }, [filteredAlerts, currentPage]);

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
            <h1 className="text-2xl font-semibold text-gray-800">Alerts Dashboard</h1>
            <TimeRangeSelector 
              value={timeRange} 
              onChange={handleTimeRangeChange} 
            />
          </div>

          {/* Alerts Table */}
          <AlertsTable alerts={paginatedAlerts} />

          {/* Pagination */}
          <div className="mt-6 flex justify-center">
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={handlePageChange}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;

