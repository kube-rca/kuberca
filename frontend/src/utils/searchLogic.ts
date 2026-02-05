import { RCAItem, AlertItem } from '../types';
import { SearchFilters } from '../context/SearchContext';

// --- [Helper] 시간 계산 ---
const getCutoffTime = (range: string): number => {
  const now = Date.now();
  const ONE_HOUR = 60 * 60 * 1000;
  const ONE_DAY = 24 * ONE_HOUR;
  switch (range) {
    case '1h': return now - 1 * ONE_HOUR;
    case '6h': return now - 6 * ONE_HOUR;
    case '24h': return now - 24 * ONE_HOUR;
    case '7d': return now - 7 * ONE_DAY;
    case '30d': return now - 30 * ONE_DAY;
    default: return 0; // 'all'
  }
};

// --- [Helper] 라벨 파싱 및 비교 ---
const matchLabels = (itemLabels: any, filterLabels: Record<string, string>): boolean => {
  if (Object.keys(filterLabels).length === 0) return true; // 필터 없으면 통과
  if (!itemLabels) return false;

  let target = itemLabels;

  // JSON 문자열이면 파싱 시도
  if (typeof target === 'string') {
    try {
      target = JSON.parse(target.replace(/'/g, '"'));
    } catch (e) {
      return false;
    }
  }

  if (typeof target !== 'object') return false;

  // 모든 필터 조건이 일치하는지 확인 (AND 조건)
  return Object.entries(filterLabels).every(([k, v]) => target[k] === v);
};

// --- [Helper] 상태 비교 ---
const matchStatus = (itemStatus: string, filterStatus: string[]): boolean => {
  if (filterStatus.length === 0) return true;
  const normalizedItemStatus = (itemStatus || '').toLowerCase(); // 'firing' | 'resolved'
  return filterStatus.some((s) => s.toLowerCase() === normalizedItemStatus);
};

// --- [Export] Alert 검색 로직 ---
export const searchAlerts = (alerts: AlertItem[], filters: SearchFilters): AlertItem[] => {
  let res = alerts;

  // 1. 텍스트 검색
  if (filters.query) {
    const q = filters.query.toLowerCase();
    res = res.filter(item => 
      (item.alarm_title || '').toLowerCase().includes(q) || 
      (item.alert_id || '').toLowerCase().includes(q),
    );
  }

  // 2. 시간 필터
  if (filters.timeRange !== 'all') {
    const cutoff = getCutoffTime(filters.timeRange);
    res = res.filter(item => new Date(item.fired_at).getTime() >= cutoff);
  }

  // 3. 상태 필터 (매핑 적용)
  if (filters.status.length > 0) {
    res = res.filter(item => matchStatus(item.status, filters.status));
  }

  // 4. 중요도 필터
  if (filters.severity.length > 0) {
    res = res.filter(item => filters.severity.some(s => s.toLowerCase() === (item.severity || '').toLowerCase()));
  }

  // 5. Namespace
  if (filters.namespaces.length > 0) {
    res = res.filter(item => item.namespace && filters.namespaces.includes(item.namespace));
  }

  // 6. Labels
  if (Object.keys(filters.labels).length > 0) {
    res = res.filter(item => matchLabels(item.labels, filters.labels));
  }

  return res;
};

// --- [Export] Incident 검색 로직 ---
export const searchIncidents = (
  rcas: RCAItem[], 
  allAlerts: AlertItem[], // 조인을 위해 필요
  filters: SearchFilters
): RCAItem[] => {
  let res = rcas;

  // 1. 텍스트 검색
  if (filters.query) {
    const q = filters.query.toLowerCase();
    res = res.filter(item => 
      (item.title || '').toLowerCase().includes(q) || 
      (item.incident_id || '').toLowerCase().includes(q)
    );
  }

  // 2. 시간 필터
  if (filters.timeRange !== 'all') {
    const cutoff = getCutoffTime(filters.timeRange);
    res = res.filter(item => new Date(item.fired_at).getTime() >= cutoff);
  }

  // 3. 상태 필터 (Incident는 resolved_at 유무로 판단)
  if (filters.status.length > 0) {
    res = res.filter(item => {
      const currentStatus = item.resolved_at ? 'resolved' : 'firing';
      return matchStatus(currentStatus, filters.status);
    });
  }

  // 4. 중요도 필터
  if (filters.severity.length > 0) {
    res = res.filter(item => filters.severity.some(s => s.toLowerCase() === (item.severity || '').toLowerCase()));
  }

  // 5. [중요] Alert 연동 필터 (Namespace / Labels)
  const hasNs = filters.namespaces.length > 0;
  const hasLabels = Object.keys(filters.labels).length > 0;

  if (hasNs || hasLabels) {
    // 조건을 만족하는 Alert들을 먼저 찾음
    const matchingAlerts = allAlerts.filter(alert => {
      if (hasNs && (!alert.namespace || !filters.namespaces.includes(alert.namespace))) return false;
      if (hasLabels && !matchLabels(alert.labels, filters.labels)) return false;
      return true;
    });

    // 매칭된 Alert들이 속한 Incident ID만 추출
    const matchingIds = new Set(matchingAlerts.map(a => a.incident_id));
    
    // 필터링
    res = res.filter(rca => matchingIds.has(rca.incident_id));
  }

  return res;
};