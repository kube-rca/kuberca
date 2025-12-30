import { useState, useMemo, useEffect, useCallback } from 'react';
import { RCAItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import RCATable from './components/RCATable';
import Pagination from './components/Pagination';
import RCADetailView from './components/RCADetailView';
import AuthPanel from './components/AuthPanel';
import { fetchRCAs } from './utils/api';
import { fetchAuthConfig, refreshAccessToken, logout } from './utils/auth';
import { filterRCAs, RCAStatusFilter } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';
import { Header } from './components/Header';

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
  
  // 초기값을 'All Time'으로 설정 (이전 요청 반영)
  const [timeRange, setTimeRange] = useState('All Time');
  const [statusFilter, setStatusFilter] = useState<RCAStatusFilter>('all');
  
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

  // [핵심 변경] 데이터 로딩 + 5초 자동 새로고침 로직
  useEffect(() => {
    if (!isAuthenticated) {
      setAllRCAs([]);
      setSelectedIncidentId(null);
      return;
    }

    // 데이터 가져오는 함수 (isBackground가 true면 로딩 스피너 안 보여줌)
    const loadRCAs = async (isBackground = false) => {
      try {
        // 백그라운드 업데이트가 아닐 때만 로딩바 표시 (깜빡임 방지)
        if (!isBackground) {
          setLoading(true);
        }
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
        
        // 데이터 업데이트 (React가 알아서 변경된 부분만 리렌더링함)
        setAllRCAs(mappedRCAs);
      } catch (err) {
        if (err instanceof Error && err.message === 'unauthorized') {
          setIsAuthenticated(false);
          return;
        }
        console.error('Failed to load RCAs:', err);
        // 백그라운드 에러는 사용자에게 크게 알리지 않거나 로그만 남김
        if (!isBackground) {
          setError('데이터를 불러오는데 실패했습니다.');
        }
      } finally {
        if (!isBackground) {
          setLoading(false);
        }
      }
    };

    // 1. 최초 실행 (로딩 스피너 보여줌)
    loadRCAs(false);

    // 2. 5초마다 반복 실행 (setInterval)
    const intervalId = setInterval(() => {
      // 여기선 true를 넘겨서 로딩바 없이 조용히 데이터만 갱신
      loadRCAs(true);
    }, 2000);

    // 3. 컴포넌트가 꺼지거나 로그아웃 시 타이머 정리 (메모리 누수 방지)
    return () => clearInterval(intervalId);

  }, [isAuthenticated]); // 인증 상태가 바뀌면 타이머도 재설정

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
    return filterRCAs(allRCAs, timeRange, statusFilter);
  }, [allRCAs, timeRange, statusFilter]);

  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange, statusFilter]);

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
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-300">
      <Header />
      <div className="pt-20 pb-8 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        
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
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
            
            <div className="mb-6 flex flex-col md:flex-row justify-between items-center gap-4">
              <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">RCA Dashboard</h1>
              
              <div className="flex flex-col sm:flex-row items-center gap-3">
                <div className="inline-flex rounded-md shadow-sm" role="group">
                  <button
                    type="button"
                    onClick={() => setStatusFilter('all')}
                    className={`px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-600 rounded-l-lg 
                      ${statusFilter === 'all' 
                        ? 'bg-blue-600 text-white border-blue-600 hover:bg-blue-700' 
                        : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                      }`}
                  >
                    All
                  </button>
                  <button
                    type="button"
                    onClick={() => setStatusFilter('ongoing')}
                    className={`px-4 py-2 text-sm font-medium border-t border-b border-gray-200 dark:border-gray-600
                      ${statusFilter === 'ongoing' 
                        ? 'bg-blue-600 text-white border-blue-600 hover:bg-blue-700' 
                        : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                      }`}
                  >
                    Ongoing
                  </button>
                  <button
                    type="button"
                    onClick={() => setStatusFilter('resolved')}
                    className={`px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-600 rounded-r-lg
                      ${statusFilter === 'resolved' 
                        ? 'bg-blue-600 text-white border-blue-600 hover:bg-blue-700' 
                        : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                      }`}
                  >
                    Resolved
                  </button>
                </div>

                <TimeRangeSelector value={timeRange} onChange={handleTimeRangeChange} />
              </div>
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