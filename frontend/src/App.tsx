import { useState, useMemo, useEffect } from 'react';
import { Routes, Route, useNavigate, useParams } from 'react-router-dom';
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

const IncidentDetailRoute = () => {
  const { id } = useParams(); 
  const navigate = useNavigate();

  if (!id) return null;

  return (
    <RCADetailView 
      incidentId={id} 
      onBack={() => navigate('/')} 
    />
  );
};

function App() {
  const navigate = useNavigate();

  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  
  // 필터 상태들
  const [timeRange, setTimeRange] = useState('All Time');
  const [statusFilter, setStatusFilter] = useState<RCAStatusFilter>('all');
  const [severityFilter, setSeverityFilter] = useState<string>('all'); 

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
      return;
    }

    const loadRCAs = async (isBackground = false) => {
      try {
        if (!isBackground) setLoading(true);
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
        if (!isBackground) setError('데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setLoading(false);
      }
    };

    loadRCAs(false);

    const intervalId = setInterval(() => {
      loadRCAs(true);
    }, 2000);

    return () => clearInterval(intervalId);

  }, [isAuthenticated]);

  const handleLogout = async () => {
    await logout();
    setIsAuthenticated(false);
  };

  const filteredRCAs = useMemo(() => {
    const baseFiltered = filterRCAs(allRCAs, timeRange, statusFilter);
    if (severityFilter === 'all') {
      return baseFiltered;
    }
    return baseFiltered.filter((item) => item.severity === severityFilter);
  }, [allRCAs, timeRange, statusFilter, severityFilter]);

  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange, statusFilter, severityFilter]);

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

  const handleTitleClick = (incident_id: string) => {
    navigate(`/incidents/${incident_id}`);
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

  // 공통 스타일 정의 (Dropdown용)
  const selectStyle = "px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors shadow-sm cursor-pointer";

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

        <Routes>
          <Route path="/incidents/:id" element={<IncidentDetailRoute />} />

          <Route path="/" element={
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
              <div className="mb-6 flex flex-col xl:flex-row justify-between items-center gap-4">
                <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">RCA Dashboard</h1>
                
                <div className="flex flex-col sm:flex-row items-center gap-3 w-full sm:w-auto">
                  
                  {/* 1. Status Filter (Dropdown) */}
                  <select
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value as RCAStatusFilter)}
                    className={selectStyle}
                  >
                    <option value="all">All Status</option>
                    <option value="ongoing">Ongoing</option>
                    <option value="resolved">Resolved</option>
                  </select>

                  {/* 2. Severity Filter (Dropdown) */}
                  <select
                    value={severityFilter}
                    onChange={(e) => setSeverityFilter(e.target.value)}
                    className={selectStyle}
                  >
                    <option value="all">All Severities</option>
                    <option value="critical">Critical</option>
                    <option value="warning">Warning</option>
                    <option value="info">Info</option>
                  </select>

                  {/* 3. Time Filter (Existing) */}
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
          } />
        </Routes>
      </div>
    </div>
  );
}

export default App;