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

// 백엔드마다 시간 필드명이 달라질 수 있어 raw 응답을 확장해 둔다.
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
  // 상세 화면 전환을 위한 선택 상태
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(null);
  const [authReady, setAuthReady] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [allowSignup, setAllowSignup] = useState(false);

  // 서버 시간 필드가 없을 때 표시용 현재 시간 문자열
  const getCurrentTimeStr = () => {
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, '0');
    const dd = String(now.getDate()).padStart(2, '0');
    const hh = String(now.getHours()).padStart(2, '0');
    const min = String(now.getMinutes()).padStart(2, '0');
    return `${yyyy}/${mm}/${dd} ${hh}:${min}`;
  };

  // 인증 설정/토큰 초기화 (언마운트 후 setState 방지)
  useEffect(() => {
    let active = true;

    const initAuth = async () => {
      try {
        const config = await fetchAuthConfig();
        const refreshed = await refreshAccessToken();
        if (!active) {
          return;
        }
        setAllowSignup(config.allowSignup);
        setIsAuthenticated(refreshed);
      } catch (err) {
        console.error('Auth init failed:', err);
      } finally {
        if (active) {
          setAuthReady(true);
        }
      }
    };

    initAuth();

    return () => {
      active = false;
    };
  }, []);

  // 인증 완료 후 RCA 목록 로딩 (로그아웃/만료 시 상태 초기화)
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

        // 서버 시간 필드를 통합하고, 누락 시 현재 시간으로 대체
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

  // 리스트/상세 화면 전환 핸들러
  const handleTitleClick = (incident_id: string) => {
    setSelectedIncidentId(incident_id);
  };

  const handleBackToList = () => {
    setSelectedIncidentId(null);
  };

  // 로그아웃 후 인증 상태 초기화
  const handleLogout = async () => {
    await logout();
    setIsAuthenticated(false);
  };

  // 시간 범위 필터링
  const filteredRCAs = useMemo(() => {
    return filterRCAsByTimeRange(allRCAs, timeRange);
  }, [allRCAs, timeRange]);

  // 시간 범위 변경 시 페이지 초기화
  useEffect(() => {
    setCurrentPage(1);
  }, [timeRange]);

  const totalPages = Math.ceil(filteredRCAs.length / ITEMS_PER_PAGE);

  // 필터된 목록에 페이지네이션 적용
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
      <div className="min-h-screen flex items-center justify-center bg-gray-100 text-gray-600">
        인증 정보를 확인하는 중입니다...
      </div>
    );
  }

  if (!isAuthenticated) {
    return <AuthPanel allowSignup={allowSignup} onAuthenticated={() => setIsAuthenticated(true)} />;
  }

  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-4 flex justify-end">
          <button
            type="button"
            onClick={handleLogout}
            className="text-sm font-medium text-gray-600 hover:text-gray-900"
          >
            로그아웃
          </button>
        </div>

        {/* ID 선택 여부에 따라 상세/리스트 전환 */}
        {selectedIncidentId ? (
          <RCADetailView incidentId={selectedIncidentId} onBack={handleBackToList} />
        ) : (
          <div className="bg-white rounded-lg shadow-md p-6">
            {/* 헤더 영역 (시간 범위 선택) */}
            <div className="mb-6 flex justify-between items-center">
              <h1 className="text-2xl font-semibold text-gray-800">RCA Dashboard</h1>
              <TimeRangeSelector value={timeRange} onChange={handleTimeRangeChange} />
            </div>

            {/* 로딩 상태 */}
            {loading && (
              <div className="flex justify-center items-center py-12">
                <div className="text-gray-600">데이터를 불러오는 중...</div>
              </div>
            )}

            {/* 오류 상태 */}
            {error && !loading && (
              <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
                <div className="text-red-800 font-medium">오류 발생</div>
                <div className="text-red-600 text-sm mt-1">{error}</div>
              </div>
            )}

            {/* 테이블 및 페이지네이션 */}
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
