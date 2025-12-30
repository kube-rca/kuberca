import { RCAItem } from '../types';

// 상태 필터 타입 정의 (전체, 진행중, 해결됨)
export type RCAStatusFilter = 'all' | 'ongoing' | 'resolved';

/**
 * 시간 범위 문자열을 현재 시간 기준의 '컷오프 타임스탬프(ms)'로 변환
 */
const getCutoffTime = (timeRange: string): number => {
  const now = Date.now();
  const ONE_HOUR = 60 * 60 * 1000;
  const ONE_DAY = 24 * ONE_HOUR;

  switch (timeRange) {
    case 'Last 1 hours':
      return now - 1 * ONE_HOUR;
    case 'Last 6 hours':
      return now - 6 * ONE_HOUR;
    case 'Last 24 hours':
      return now - 24 * ONE_HOUR;
    case 'Last 7 days':
      return now - 7 * ONE_DAY;
    case 'Last 30 days':
      return now - 30 * ONE_DAY;
    default:
      // 기본값이 없거나 'All Time' 같은 경우 아주 옛날로 설정
      return 0; 
  }
};

/**
 * 핵심 필터링 함수
 * 1. 시간(fired_at) 필터 적용
 * 2. 상태(resolved_at 유무) 필터 적용
 * 3. 최신순 정렬
 */
export const filterRCAs = (
  rcas: RCAItem[],
  timeRange: string,
  status: RCAStatusFilter // 새로 추가된 파라미터
): RCAItem[] => {
  const cutoffTime = getCutoffTime(timeRange);

  const filtered = rcas.filter((rca) => {
    // --- [1단계] 시간 필터 (Time Filter) ---
    // 발생 시간(fired_at)이 없으면 데이터 오류로 간주하고 제외 (안전장치)
    if (!rca.fired_at) return false;
    
    const firedTime = new Date(rca.fired_at).getTime();
    
    // 선택된 시간 범위보다 이전에 발생했다면 탈락
    if (firedTime < cutoffTime) {
      return false;
    }

    // --- [2단계] 상태 필터 (Status Filter) ---
    if (status === 'ongoing') {
      // 진행 중만 보고 싶다 -> resolved_at이 없어야(null) 함
      if (rca.resolved_at) return false; 
    }
    
    if (status === 'resolved') {
      // 해결된 것만 보고 싶다 -> resolved_at이 있어야 함
      if (!rca.resolved_at) return false;
    }

    // status가 'all'이면 위 조건들을 다 통과
    return true;
  });

  // --- [3단계] 정렬 (Sorting) ---
  // 발생 시간(fired_at) 기준 내림차순 (최신이 위로)
  return filtered.sort((a, b) => {
    const timeA = new Date(a.fired_at).getTime();
    const timeB = new Date(b.fired_at).getTime();
    return timeB - timeA;
  });
};