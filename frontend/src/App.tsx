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
import ArchivedDetailView from './components/ArchiveDetailView';
import AuthPanel from './components/AuthPanel';
import { fetchRCAs, fetchAlerts, fetchMutedIncidents, AlertItem } from './utils/api';
import { fetchAuthConfig, refreshAccessToken, logout } from './utils/auth';
import { ITEMS_PER_PAGE } from './constants';
import { Header } from './components/Header';
import { Sidebar } from './components/Sidebar';
import UnifiedSearchPanel from './components/UnifiedSearchPanel';
import SettingsPage from './components/SettingsPage';
import WebhookSettings from './components/WebhookSettings';
import WebhookList from './components/WebhookList';
import FloatingChatPanel from './components/FloatingChatPanel';
import { useSearch } from './context/SearchContext';
// [필수] 우리가 만든 로직 Import
import { searchIncidents, searchAlerts } from './utils/searchLogic';

type RawRCAItem = RCAItem & {
  created_at?: string;
  timestamp?: string;
  time?: string;
  start_time?: string;
  fired_at?: string;
};

// --- Route Components ---
const IncidentDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  if (!id) return null;
  return <RCADetailView incidentId={id} onBack={() => navigate(-1)} />;
};

const MuteDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  if (!id) return null;
  return <ArchivedDetailView incidentId={id} onBack={() => navigate('/muted')} />;
};

const AlertDetailRoute = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  if (!id) return null;
  return <AlertDetailView alertId={id} onBack={() => navigate(-1)} />;
};

// --- Main App ---
function App() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { scope, filters, updateFilter } = useSearch();

  // Data
  const [allRCAs, setAllRCAs] = useState<RCAItem[]>([]);
  const [allAlerts, setAllAlerts] = useState<AlertItem[]>([]);
  const [mutedIncidents, setMutedIncidents] = useState<RCAItem[]>([]); 

  // Loading/Error
  const [loading, setLoading] = useState(true);
  const [alertLoading, setAlertLoading] = useState(true);
  const [muteLoading, setMuteLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [alertError, setAlertError] = useState<string | null>(null);
  const [muteError, setMuteError] = useState<string | null>(null); 

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const [alertCurrentPage, setAlertCurrentPage] = useState(1);
  const [muteCurrentPage, setMuteCurrentPage] = useState(1); 
  
  // Legacy Dropdown State (UI 호환용, 실제 로직은 SearchContext 우선)
  const [timeRange, setTimeRange] = useState(() => searchParams.get('time') || 'All Time');
  const [statusFilter, setStatusFilter] = useState<string>(() => searchParams.get('status') || 'all');
  const [severityFilter, setSeverityFilter] = useState(() => searchParams.get('severity') || 'all'); 

  // Auth
  const [authReady, setAuthReady] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [allowSignup, setAllowSignup] = useState(false);
  const [oidcEnabled, setOidcEnabled] = useState(false);
  const [oidcLoginUrl, setOidcLoginUrl] = useState('');
  const [oidcProvider, setOidcProvider] = useState('');
  const [isChatDocked, setIsChatDocked] = useState(false);

  const mapLegacyTimeToKey = (range: string): string => {
    switch (range) {
      case 'All Time': return 'all';
      case 'Last 1 hours': return '1h';
      case 'Last 6 hours': return '6h';
      case 'Last 24 hours': return '24h';
      case 'Last 7 days': return '7d';
      case 'Last 30 days': return '30d';
      default: return 'all';
    }
  };

  // URL Sync (기존 쿼리 파라미터 유지용, 인증 후에만 실행)
  useEffect(() => {
    if (!isAuthenticated) return;
    const params: Record<string, string> = {};
    if (statusFilter !== 'all') params.status = statusFilter;
    if (severityFilter !== 'all') params.severity = severityFilter;
    if (timeRange !== 'All Time') params.time = timeRange;
    setSearchParams(params, { replace: true });
  }, [isAuthenticated, statusFilter, severityFilter, timeRange, setSearchParams]);

  const getCurrentTimeStr = () => {
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, '0');
    const dd = String(now.getDate()).padStart(2, '0');
    const hh = String(now.getHours()).padStart(2, '0');
    const min = String(now.getMinutes()).padStart(2, '0');
    return `${yyyy}/${mm}/${dd} ${hh}:${min}`;
  };

  // Auth Init
  useEffect(() => {
    let active = true;
    const initAuth = async () => {
      try {
        const config = await fetchAuthConfig();
        const refreshed = await refreshAccessToken();
        if (!active) return;
        setAllowSignup(config.allowSignup);
        setOidcEnabled(config.oidcEnabled || false);
        setOidcLoginUrl(config.oidcLoginUrl || '');
        setOidcProvider(config.oidcProvider || '');
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

  // Cleanup
  useEffect(() => {
    if (location.state) {
      const state = location.state as { newlyMutedId?: string; newlyUnmutedId?: string };
      if (state.newlyMutedId) {
        setAllRCAs((prev) => prev.filter((item) => item.incident_id !== state.newlyMutedId));
      }
      if (state.newlyUnmutedId) {
        setMutedIncidents((prev) => prev.filter((item) => item.incident_id !== state.newlyUnmutedId));
      }
    }
  }, [location]);  

  // Fetch Data
  useEffect(() => {
    if (!isAuthenticated) {
      setAllRCAs([]);
      setAllAlerts([]);
      setMutedIncidents([]);
      return;
    }

    const loadData = async (isBackground = false) => {
      try {
        if (!isBackground) setLoading(true);
        setError(null);
        const rawData: RawRCAItem[] = await fetchRCAs();
        const mappedRCAs = rawData.map((item) => {
          const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;
          return { ...item, time: serverTime ? String(serverTime) : getCurrentTimeStr() };
        });
        setAllRCAs(mappedRCAs);
      } catch (err) {
        if (!isBackground) setError('데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setLoading(false);
      }

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

      try {
        if (!isBackground) setMuteLoading(true);
        setMuteError(null);
        const rawMuted: RawRCAItem[] = await fetchMutedIncidents();
        const mappedMuted = rawMuted.map((item) => {
          const serverTime = item.created_at || item.timestamp || item.time || item.start_time || item.fired_at;
          return { ...item, time: serverTime ? String(serverTime) : getCurrentTimeStr() };
        });
        setMutedIncidents(mappedMuted);
      } catch (err) {
        if (!isBackground) setMuteError('Mute 데이터를 불러오는데 실패했습니다.');
      } finally {
        if (!isBackground) setMuteLoading(false);
      }
    };

    loadData(false);
    const intervalId = setInterval(() => { loadData(true); }, 1000);
    return () => clearInterval(intervalId);
  }, [isAuthenticated]);

  const handleLogout = async () => {
    await logout();
    setIsAuthenticated(false);
  };

  const handleTitleClick = (id: string) => navigate(`/incidents/${id}`);
  const handleMuteTitleClick = (id: string) => navigate(`/muted/${id}`);
  const handleAlertTitleClick = (id: string) => navigate(`/alerts/${id}`);

  // --- [핵심] 필터링 로직: searchLogic.ts + SearchContext 사용 ---

  // 1. Incident
  const filteredRCAs = useMemo(() => {
    if (scope !== 'INCIDENT') {
      // Incident 대상이 아닐 때는 필터 미적용 (전체 목록)
      return allRCAs;
    }
    return searchIncidents(allRCAs, allAlerts, filters);
  }, [scope, allRCAs, allAlerts, filters]);

  const paginatedRCAs = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    return filteredRCAs.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredRCAs, currentPage]);


  // 2. Alert
  const filteredAlerts = useMemo(() => {
    if (scope !== 'ALERT') {
      // Alert 대상이 아닐 때는 필터 미적용 (전체 목록)
      return allAlerts;
    }
    return searchAlerts(allAlerts, filters);
  }, [scope, allAlerts, filters]);

  const paginatedAlerts = useMemo(() => {
    const startIndex = (alertCurrentPage - 1) * ITEMS_PER_PAGE;
    return filteredAlerts.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredAlerts, alertCurrentPage]);


  // 3. Muted (Incident와 동일한 스코프 규칙 적용)
  const filteredMutedIncidents = useMemo(() => {
    if (scope !== 'INCIDENT') {
      return mutedIncidents;
    }
    return searchIncidents(mutedIncidents, allAlerts, filters);
  }, [scope, mutedIncidents, allAlerts, filters]);

  const paginatedMutedIncidents = useMemo(() => {
    const startIndex = (muteCurrentPage - 1) * ITEMS_PER_PAGE;
    return filteredMutedIncidents.slice(startIndex, startIndex + ITEMS_PER_PAGE);
  }, [filteredMutedIncidents, muteCurrentPage]);

  // 페이지네이션 초기화
  useEffect(() => {
    setCurrentPage(1);
    setAlertCurrentPage(1);
    setMuteCurrentPage(1);
  }, [filters]);

  // 라벨/네임스페이스 추출 (App.tsx에 남겨둠 - 검색패널 Props용)
  const availableLabels = useMemo(() => {
    const labelsSet = new Set<string>();
    if (allAlerts) {
      allAlerts.forEach(item => {
        let target = item.labels;
        if (!target) return;
        if (typeof target === 'string') {
           try { target = JSON.parse((target as string).replace(/'/g, '"')); } catch(e) { return; }
        }
        if (typeof target === 'object') {
          Object.entries(target).forEach(([k, v]) => labelsSet.add(`${k}:${v}`));
        }
      });
    }
    return Array.from(labelsSet).sort();
  }, [allAlerts]);

  const availableNamespaces = useMemo(() => {
    const nsSet = new Set<string>();
    if (allAlerts) {
      allAlerts.forEach(item => { if (item.namespace) nsSet.add(item.namespace); });
    }
    return Array.from(nsSet).sort();
  }, [allAlerts]);


  if (!authReady) {
    return <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400">인증 정보를 확인하는 중입니다...</div>;
  }
  if (!isAuthenticated) {
    return <AuthPanel allowSignup={allowSignup} oidcEnabled={oidcEnabled} oidcLoginUrl={oidcLoginUrl} oidcProvider={oidcProvider} onAuthenticated={() => setIsAuthenticated(true)} />;
  }

  // 스타일
  const selectStyle = "px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors shadow-sm cursor-pointer text-left";
  const isSettingsRoute = location.pathname.startsWith('/settings');

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-300">
      <Header onLogout={handleLogout} />
      <div className="pt-16">
        <Sidebar />
        <div className={`md:ml-64 px-4 sm:px-6 lg:px-8 py-6 transition-all duration-300 ${isChatDocked ? "md:mr-[26rem]" : ""}`}>
          <div className="w-full max-w-[1600px] mx-auto">
            {!isSettingsRoute && (
              <div className="mb-6">
                <UnifiedSearchPanel availableLabels={availableLabels} availableNamespaces={availableNamespaces} />
              </div>
            )}

            <Routes>
              <Route path="/incidents/:id" element={<IncidentDetailRoute />} />
              <Route path="/alerts/:id" element={<AlertDetailRoute />} />
              <Route path="/muted/:id" element={<MuteDetailRoute />} />

              {/* Incident Dashboard */}
              <Route path="/" element={
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
                  <div className="mb-6 flex flex-col lg:flex-row justify-between items-center gap-4">
                    <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Incident Dashboard</h1>
                    <div className="flex flex-col sm:flex-row items-center justify-end gap-3 w-full sm:w-auto">
                      <select
                        value={statusFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setStatusFilter(value);
                          updateFilter('status', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Status</option>
                        <option value="firing">Firing</option>
                        <option value="resolved">Resolved</option>
                      </select>
                      <select
                        value={severityFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setSeverityFilter(value);
                          updateFilter('severity', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Severities</option>
                        <option value="warning">Warning</option>
                        <option value="critical">Critical</option>
                      </select>
                      <TimeRangeSelector
                        value={timeRange}
                        onChange={(value) => {
                          setTimeRange(value);
                          updateFilter('timeRange', mapLegacyTimeToKey(value));
                        }}
                      />
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
                        <div className="flex justify-center items-center py-12 text-gray-500 dark:text-gray-400">
                          데이터가 없습니다.
                        </div>
                      )}
                    </>
                  )}
                </div>
              } />

              {/* Alert Dashboard */}
              <Route path="/alerts" element={
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
                  <div className="mb-6 flex flex-col lg:flex-row justify-between items-center gap-4">
                    <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Alert Dashboard</h1>
                    <div className="flex flex-col sm:flex-row items-center justify-end gap-3 w-full sm:w-auto">
                      <select
                        value={statusFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setStatusFilter(value);
                          updateFilter('status', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Status</option>
                        <option value="firing">Firing</option>
                        <option value="resolved">Resolved</option>
                      </select>
                      <select
                        value={severityFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setSeverityFilter(value);
                          updateFilter('severity', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Severities</option>
                        <option value="warning">Warning</option>
                        <option value="critical">Critical</option>
                      </select>
                      <TimeRangeSelector
                        value={timeRange}
                        onChange={(value) => {
                          setTimeRange(value);
                          updateFilter('timeRange', mapLegacyTimeToKey(value));
                        }}
                      />
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
              
              {/* Muted Route */}
              <Route path="/muted" element={
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 transition-colors duration-300">
                  <div className="mb-6 flex flex-col lg:flex-row justify-between items-center gap-4">
                    <h1 className="text-2xl font-semibold text-gray-800 dark:text-white">Archived Incidents</h1>
                    <div className="flex flex-col sm:flex-row items-center justify-end gap-3 w-full sm:w-auto">
                      <select
                        value={statusFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setStatusFilter(value);
                          updateFilter('status', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Status</option>
                        <option value="firing">Firing</option>
                        <option value="resolved">Resolved</option>
                      </select>
                      <select
                        value={severityFilter}
                        onChange={(e) => {
                          const value = e.target.value;
                          setSeverityFilter(value);
                          updateFilter('severity', value === 'all' ? [] : [value]);
                        }}
                        className={selectStyle}
                      >
                        <option value="all">All Severities</option>
                        <option value="warning">Warning</option>
                        <option value="critical">Critical</option>
                      </select>
                      <TimeRangeSelector
                        value={timeRange}
                        onChange={(value) => {
                          setTimeRange(value);
                          updateFilter('timeRange', mapLegacyTimeToKey(value));
                        }}
                      />
                    </div>
                  </div>
                  
                  {muteLoading ? (
                    <div className="flex justify-center items-center py-12 text-gray-600 dark:text-gray-400">데이터를 불러오는 중...</div>
                  ) : muteError ? (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 mb-4 text-red-600 dark:text-red-400">{muteError}</div>
                  ) : (
                    <>
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

              {/* Settings Routes */}
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/settings/webhooks" element={<WebhookList />} />
              <Route path="/settings/webhooks/new" element={<WebhookSettings />} />
              <Route path="/settings/webhooks/:id" element={<WebhookSettings />} />
            </Routes>
          </div>
        </div>
      </div>
      <FloatingChatPanel onDockedChange={setIsChatDocked} />
    </div>
  );
}

export default App;
