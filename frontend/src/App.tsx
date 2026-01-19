import { useState, useMemo, useEffect } from 'react';
import { Routes, Route, useNavigate, useParams, useSearchParams, useLocation } from 'react-router-dom';
import { RCAItem } from './types';
import TimeRangeSelector from './components/TimeRangeSelector';
import RCATable from './components/RCATable';
import ArchivedTable from './components/ArchiveTable'; 
import AlertTable from './components/AlertTable';
import Pagination from './components/Pagination';
import RCADetailView from './components/RCADetailView';
import AlertDetailView from './components/AlertDetailView';
import ArchivedDetailView from './components/ArchiveDetailView'; // [Import 확인]
import AuthPanel from './components/AuthPanel';
import { fetchRCAs, fetchAlerts, fetchMutedIncidents, AlertItem } from './utils/api';
import { fetchAuthConfig, refreshAccessToken, logout } from './utils/auth';
import { filterRCAs, filterAlerts, RCAStatusFilter } from './utils/filterAlerts';
import { ITEMS_PER_PAGE } from './constants';
import { Header } from './components/Header';

type RawRCAItem = RCAItem & {
  created_at?: string;
  timestamp?: string;
  time?: string;
  start_time?: string;
  fired_at?: string;
};

// --- Route Wrapper Components ---

const IncidentDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  if (!id) return null;

  return (
    <RCADetailView
      incidentId={id}
      onBack={() => navigate(-1)}
    />
  );
};

const MuteDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  if (!id) return null;
  // onBack 시 목록(/muted)으로 돌아가도록 설정
  return <ArchivedDetailView incidentId={id} onBack={() => navigate('/muted')} />;
};

const AlertDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  if (!id) return null;

  return (
    <AlertDetailView
      alertId={id}
      onBack={() => navigate(-1)}
    />
  );
};

// --- Main App Component ---

function App() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();

  // Data States
  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [allAlerts, setAllAlerts] = useState<AlertItem[]>([]);
  const [mutedIncidents, setMutedIncidents] = useState<RCAItem[]>([]); 

  // Loading & Error States
  const [loading, setLoading] = useState(true);
  const [alertLoading, setAlertLoading] = useState(true);
  const [muteLoading, setMuteLoading] = useState(true);
  
  const [error, setError] = useState<string | null>(null);
  const [alertError, setAlertError] = useState<string | null>(null);
  const [muteError, setMuteError] = useState<string | null>(null); 

  // Pagination States
  const [currentPage, setCurrentPage] = useState(1);
  const [alertCurrentPage, setAlertCurrentPage] = useState(1);
  const [muteCurrentPage, setMuteCurrentPage] = useState(1); 
  
  // Filter States
  const [timeRange, setTimeRange] = useState(() => searchParams.get('time') || 'All Time');
  const [statusFilter, setStatusFilter] = useState<RCAStatusFilter>(() => (searchParams.get('status') as RCAStatusFilter) || 'all');
  const [severityFilter, setSeverityFilter] = useState(() => searchParams.get('severity') || 'all'); 

  // Auth States
  const [authReady, setAuthReady] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [allowSignup, setAllowSignup] = useState(false);

  // URL Query Sync
  useEffect(() => {
    const params: Record<string, string> = {};
    if (statusFilter !== 'all') params.status = statusFilter;
    if (severityFilter !== 'all') params.severity = severityFilter;
    if (timeRange !== 'All Time') params.time = timeRange;
    
    setSearchParams(params, { replace: true });
  }, [statusFilter, severityFilter, timeRange, setSearchParams]);

  const getCurrentTimeStr = () => {
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, '0');
    const dd = String(now.getDate()).padStart(2, '0');
    const hh = String(now.getHours()).padStart(2, '0');
    const min = String(now.getMinutes()).padStart(2, '0');
    return `${yyyy}/${mm}/${dd} ${hh}:${min}`;
  };

  // Auth Initialization
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

// 페이지 이동으로 전달받은 state가 있다면, 리스트에서 즉시 제거하여 반응 속도를 높임
  useEffect(() => {
    if (location.state) {
      const state = location.state as { newlyMutedId?: string; newlyUnmutedId?: string };

      // 1. 숨기기(Mute) 되어 넘어온 경우 -> Incident 목록에서 즉시 제거
      if (state.newlyMutedId) {
        setAllRCAs((prev) => prev.filter((item) => item.incident_id !== state.newlyMutedId));
        
        // (선택) state를 한 번 썼으면 지워주는 것이 좋음 (history replace)
        // window.history.replaceState({}, document.title); 
      }

      // 2. 숨기기 해제(Unmute) 되어 넘어온 경우 -> Mute 목록에서 즉시 제거
      if (state.newlyUnmutedId) {
        setMutedIncidents((prev) => prev.filter((item) => item.incident_id !== state.newlyUnmutedId));
      }
    }
  }, [location]);  

  // Data Fetching Logic (Main RCA, Alerts, Muted RCAs)
  useEffect(() => {
    if (!isAuthenticated) {
      setAllRCAs([]);
      setAllAlerts([]);
      setMutedIncidents([]);
      return;
    }

    const loadData = async (isBackground = false) => {
      // 1. Load Main RCAs
      try {
        if (!isBackground) setLoading(true);
        setError(null);
        const rawData: RawRCAItem[] = await fetchRCAs();
        const mappedRCAs = rawData.map((item) => {
          const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;
          return {
            ...item,
            time: serverTime ? String(serverTime) : getCurrentTimeStr(),
          };
        });
        setAllRCAs(mappedRCAs);
      } catch (err) {
        if (err instanceof Error && err.message === 'unauthorized') {
          setIsAuthenticated(false);
          return;
        }
        if (!isBackground) setError('데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setLoading(false);
      }

      // 2. Load Alerts
      try {
        if (!isBackground) setAlertLoading(true);
        setAlertError(null);
        const data = await fetchAlerts();
        setAllAlerts(data);
      } catch (err) {
        if (!isBackground) setAlertError('Alert 데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setAlertLoading(false);
      }

      // 3. Load Muted RCAs 
      try {
        if (!isBackground) setMuteLoading(true);
        setMuteError(null);
        const rawMuted: RawRCAItem[] = await fetchMutedIncidents();
        const mappedMuted = rawMuted.map((item) => {
          const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;
          return {
            ...item,
            time: serverTime ? String(serverTime) : getCurrentTimeStr(),
          };
        });
        setMutedIncidents(mappedMuted);
      } catch (err) {
        if (!isBackground) setMuteError('Mute 데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setMuteLoading(false);
      }
    };

    loadData(false);

    const intervalId = setInterval(() => {
      loadData(true);
    }, 1000);

    return () => clearInterval(intervalId);

  }, [isAuthenticated]);

  const handleLogout = async () => {
    await logout();
    setIsAuthenticated(false);
  };

  // --- Navigation Handlers ---

  const handleTitleClick = (incident_id: string) => {
    navigate(`/incidents/${incident_id}`);
  };

  // [추가] Muted Incident 상세 페이지 이동 핸들러
  const handleMuteTitleClick = (incident_id: string) => {
    navigate(`/muted/${incident_id}`);
  };

  const handleAlertTitleClick = (alert_id: string) => {
    navigate(`/alerts/${alert_id}`);
  };

  // --- Filtering & Pagination Logic ---

  // 1. Incident Dashboard Logic
  const filteredRCAs = useMemo(() => {
    const baseFiltered = filterRCAs(allRCAs, timeRange, statusFilter);
    if (severityFilter === 'all') return baseFiltered;
    return baseFiltered.filter((item) => item.severity === severityFilter);
  }, [allRCAs, timeRange, statusFilter, severityFilter]);

  const paginatedRCAs = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    return filteredRCAs.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredRCAs, currentPage]);

  // 2. Alert Dashboard Logic
  const filteredAlerts = useMemo(() => {
    const baseFiltered = filterAlerts(allAlerts, timeRange, statusFilter);
    if (severityFilter === 'all') return baseFiltered;
    return baseFiltered.filter((item) => item.severity === severityFilter);
  }, [allAlerts, timeRange, statusFilter, severityFilter]);

  const paginatedAlerts = useMemo(() => {
    const startIndex = (alertCurrentPage - 1) * ITEMS_PER_PAGE;
    return filteredAlerts.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredAlerts, alertCurrentPage]);

  // 3. Mute Dashboard Logic
  const filteredMutedIncidents = useMemo(() => {
    const baseFiltered = filterRCAs(mutedIncidents, timeRange, statusFilter);
    if (severityFilter === 'all') return baseFiltered;
    return baseFiltered.filter((item) => item.severity === severityFilter);
  }, [mutedIncidents, timeRange, statusFilter, severityFilter]);

  const paginatedMutedIncidents = useMemo(() => {
    const startIndex = (muteCurrentPage - 1) * ITEMS_PER_PAGE;
    return filteredMutedIncidents.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredMutedIncidents, muteCurrentPage]);

  // Reset pagination on filter change
  useEffect(() => {
    setCurrentPage(1);
    setAlertCurrentPage(1);
    setMuteCurrentPage(1);
  }, [timeRange, statusFilter, severityFilter]);


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
          <Route path="/alerts/:id" element={<AlertDetailRoute />} />
          
          {/* [추가] Muted 상세 페이지 라우트 연결 */}
          <Route path="/muted/:id" element={<MuteDetailRoute />} />

          {/* Incident Dashboard */}
          <Route path="/" element={
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
              <div className="mb-6 flex flex-col xl:flex-row justify-between items-center gap-4">
                <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Incident Dashboard</h1>
                <div className="flex flex-col sm:flex-row items-center gap-3 w-full sm:w-auto">
                  <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value as RCAStatusFilter)} className={selectStyle}>
                    <option value="all">All Status</option>
                    <option value="ongoing">Firing</option>
                    <option value="resolved">Resolved</option>
                  </select>
                  <select value={severityFilter} onChange={(e) => setSeverityFilter(e.target.value)} className={selectStyle}>
                    <option value="all">All Severities</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                  </select>
                  <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
                </div>
              </div>

              {loading ? (
                <div className="flex justify-center items-center py-12 text-gray-600 dark:text-gray-400">데이터를 불러오는 중...</div>
              ) : error ? (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4 text-red-600 dark:text-red-400">{error}</div>
              ) : (
                <>
                  <RCATable rcas={paginatedRCAs} onTitleClick={handleTitleClick} />
                  {filteredRCAs.length > 0 && (
                    <div className="mt-6 flex justify-center">
                      <Pagination currentPage={currentPage} totalPages={Math.ceil(filteredRCAs.length / ITEMS_PER_PAGE)} onPageChange={setCurrentPage} />
                    </div>
                  )}
                  {filteredRCAs.length === 0 && (
                    <div className="flex justify-center items-center py-12 text-gray-500 dark:text-gray-400">표시할 데이터가 없습니다.</div>
                  )}
                </>
              )}
            </div>
          } />

          {/* Archived Incidents Dashboard */}
          <Route path="/muted" element={
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
              <div className="mb-6 flex flex-col xl:flex-row justify-between items-center gap-4">
                <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Archived Incidents</h1>
                <div className="flex flex-col sm:flex-row items-center gap-3 w-full sm:w-auto">
                  <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value as RCAStatusFilter)} className={selectStyle}>
                    <option value="all">All Status</option>
                    <option value="ongoing">Firing</option>
                    <option value="resolved">Resolved</option>
                  </select>
                  <select value={severityFilter} onChange={(e) => setSeverityFilter(e.target.value)} className={selectStyle}>
                    <option value="all">All Severities</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                  </select>
                  <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
                </div>
              </div>

              {muteLoading ? (
                <div className="flex justify-center items-center py-12 text-gray-600 dark:text-gray-400">데이터를 불러오는 중...</div>
              ) : muteError ? (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4 text-red-600 dark:text-red-400">{muteError}</div>
              ) : (
                <>
                  {/* [수정] onTitleClick에 handleMuteTitleClick 전달 */}
                  <ArchivedTable rcas={paginatedMutedIncidents} onTitleClick={handleMuteTitleClick} />
                  {filteredMutedIncidents.length > 0 && (
                    <div className="mt-6 flex justify-center">
                      <Pagination currentPage={muteCurrentPage} totalPages={Math.ceil(filteredMutedIncidents.length / ITEMS_PER_PAGE)} onPageChange={setMuteCurrentPage} />
                    </div>
                  )}
                </>
              )}
            </div>
          } />

          {/* Alert Dashboard */}
          <Route path="/alerts" element={
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
              <div className="mb-6 flex flex-col xl:flex-row justify-between items-center gap-4">
                <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Alert Dashboard</h1>
                <div className="flex flex-col sm:flex-row items-center gap-3 w-full sm:w-auto">
                  <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value as RCAStatusFilter)} className={selectStyle}>
                    <option value="all">All Status</option>
                    <option value="ongoing">Firing</option>
                    <option value="resolved">Resolved</option>
                  </select>
                  <select value={severityFilter} onChange={(e) => setSeverityFilter(e.target.value)} className={selectStyle}>
                    <option value="all">All Severities</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                  </select>
                  <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
                </div>
              </div>

              {alertLoading ? (
                <div className="flex justify-center items-center py-12 text-gray-600 dark:text-gray-400">데이터를 불러오는 중...</div>
              ) : alertError ? (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4 text-red-600 dark:text-red-400">{alertError}</div>
              ) : (
                <>
                  <AlertTable alerts={paginatedAlerts} onTitleClick={handleAlertTitleClick} />
                  {filteredAlerts.length > 0 && (
                    <div className="mt-6 flex justify-center">
                      <Pagination currentPage={alertCurrentPage} totalPages={Math.ceil(filteredAlerts.length / ITEMS_PER_PAGE)} onPageChange={setAlertCurrentPage} />
                    </div>
                  )}
                  {filteredAlerts.length === 0 && (
                    <div className="flex justify-center items-center py-12 text-gray-500 dark:text-gray-400">표시할 Alert이 없습니다.</div>
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