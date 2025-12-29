import { useState, useMemo, useEffect } from 'react';
import { RCAItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import RCATable from './components/RCATable';
import Pagination from './components/Pagination';
import RCADetailView from './components/RCADetailView';
import AuthPanel from './components/AuthPanel';
import { fetchRCAs } from './utils/api';
import { fetchAuthConfig, refreshAccessToken, logout } from './utils/auth';
import { filterRCAsByTimeRange } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';
import { Header } from './components/Header'; // Header 임포트 확인

type RawRCAItem = RCAItem & {
  created_at?: string;
  timestamp?: string;
  time?: string;
  start_time?: string;
  fired_at?: string;
};

function App() {
  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [timeRange, setTimeRange] = useState('Last 1 hours');
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(null);
  const [authReady, setAuthReady] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [allowSignup, setAllowSignup] = useState(false);

  const getCurrentTimeStr = () => {
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, '0');
    const dd = String(now.getDate()).padStart(2, '0');
    const hh = String(now.getHours()).padStart(2, '0');
    const min = String(now.getMinutes()).padStart(2, '0');
    return `${yyyy}/${mm}/${dd} ${hh}:${min}`;
  };

  useEffect(() => {
    let active = true;

    const initAuth = async () => {
      try {
        const config = await fetchAuthConfig();
        const refreshed = await refreshAccessToken();
        if (!active) return;
        setAllowSignup(config.allowSignup);
        setIsAuthenticated(refreshed);
      } catch (err) {
        console.error('Auth init failed:', err);
      } finally {
        if (active) setAuthReady(true);
      }
    };

    initAuth();
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      setAllRCAs([]);
      setSelectedIncidentId(null);
      return;
    }

    const loadRCAs = async () => {
      try {
        setLoading(true);
        setError(null);
        const rawData: RawRCAItem[] = await fetchRCAs();
        const mappedRCAs: RCAItem[] = rawData.map((item) => {
          const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;
          return {
            ...item,
            incident_id: item.incident_id,
            alarm_title: item.alarm_title,
            severity: item.severity,
            time: serverTime ? String(serverTime) : getCurrentTimeStr(),
          };
        });
        setAllRCAs(mappedRCAs);
      } catch (err) {
        if (err instanceof Error && err.message === 'unauthorized') {
          setIsAuthenticated(false);
          return;
        }
        console.error('Failed to load RCAs:', err);
        setError('데이터를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };
    loadRCAs();
  }, [isAuthenticated]);

  const handleTitleClick = (incident_id: string) => {
    setSelectedIncidentId(incident_id);
  };

  const handleBackToList = () => {
    setSelectedIncidentId(null);
  };

  const handleLogout = async () => {
    await logout();
    setIsAuthenticated(false);
  };

  const filteredRCAs = useMemo(() => {
    return filterRCAsByTimeRange(allRCAs, timeRange);
  }, [allRCAs, timeRange]);

  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange]);

  const totalPages = Math.ceil(filteredRCAs.length / ITEMS_PER_PAGE);

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

  if (!authReady) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400">
        인증 정보를 확인하는 중입니다...
      </div>
    );
  }

  if (!isAuthenticated) {
    return <AuthPanel allowSignup={allowSignup} onAuthenticated={() => setIsAuthenticated(true)} />;
  }

  return (
    // 1. 전체 배경에 다크모드 적용 (dark:bg-gray-900)
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-300">
      
      {/* 2. 상단 헤더 추가 */}
      <Header />

      {/* 3. 헤더 높이만큼 상단 여백 추가 (pt-20) */}
      <div className="pt-20 pb-8 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        
        {/* 기존 로그아웃 버튼 (위치는 유지하되 스타일 조금 다듬음) */}
        <div className="mb-4 flex justify-end">
          <button
            type="button"
            onClick={handleLogout}
            className="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
          >
            로그아웃
          </button>
        </div>

        {selectedIncidentId ? (
          <RCADetailView incidentId={selectedIncidentId} onBack={handleBackToList} />
        ) : (
          // 4. 컨텐츠 박스에도 다크모드 배경색 적용 (dark:bg-gray-800)
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
            
            <div className="mb-6 flex justify-between items-center">
              {/* 텍스트 색상 다크모드 대응 (dark:text-white) */}
              <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">RCA Dashboard</h1>
              <TimeRangeSelector value={timeRange} onChange={handleTimeRangeChange} />
            </div>

            {loading && (
              <div className="flex justify-center items-center py-12">
                <div className="text-gray-600 dark:text-gray-400">데이터를 불러오는 중...</div>
              </div>
            )}

            {error && !loading && (
              <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4">
                <div className="text-red-800 dark:text-red-300 font-medium">오류 발생</div>
                <div className="text-red-600 dark:text-red-400 text-sm mt-1">{error}</div>
              </div>
            )}

            {!loading && !error && (
              <>
                <RCATable rcas={paginatedRCAs} onTitleClick={handleTitleClick} />

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
                    <div className="text-gray-500 dark:text-gray-400">표시할 RCA가 없습니다.</div>
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