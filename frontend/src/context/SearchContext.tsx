import React, { createContext, useContext, useState, ReactNode, useCallback } from 'react';

// 1. 검색 범위 (탭) 타입 정의
export type SearchScope = 'INCIDENT' | 'ALERT';

// 2. 필터 상태 타입 정의
export interface SearchFilters {
  query: string;             // 텍스트 검색어 (Client-side Filtering)
  timeRange: string;         // 시간 범위 키 (예: '1h', '24h', 'all')
  status: string[];          // 상태 필터 (예: ['firing', 'resolved'])
  severity: string[];        // 중요도 필터 (예: ['critical', 'warning'])
  namespaces: string[];      // 네임스페이스 (예: ['default', 'monitoring'])
  // 개별 라벨 필터 목록. "key:value" 형태의 토큰을 여러 개 가질 수 있다.
  labels: string[];          // 예: ['service:payment', 'uid:1234']
}

// 필터 기본값
const defaultFilters: SearchFilters = {
  query: '',
  // 기본값: 시간 제한 없음 (사용자가 선택할 때까지 전체 데이터)
  timeRange: 'all',
  status: [],
  severity: [],
  namespaces: [],
  labels: [],
};

// 3. Context 데이터 타입 정의
interface SearchContextType {
  scope: SearchScope;
  setScope: (scope: SearchScope) => void;
  filters: SearchFilters;
  setFilters: React.Dispatch<React.SetStateAction<SearchFilters>>;
  
  // 개별 필터 업데이트 헬퍼 함수
  updateFilter: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void;
  
  // 필터 초기화 함수
  resetFilters: () => void;
}

const SearchContext = createContext<SearchContextType | undefined>(undefined);

// 4. Provider 컴포넌트
export const SearchProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [scope, setScopeState] = useState<SearchScope>('INCIDENT');
  const [filters, setFilters] = useState<SearchFilters>(defaultFilters);

  // Scope 변경 시, 필터 일부를 초기화할지 여부를 결정하는 로직
  const setScope = useCallback((newScope: SearchScope) => {
    setScopeState(newScope);
    // 탭을 바꿀 때 검색어나 필터를 초기화하고 싶다면 아래 주석을 해제하세요.
    // 현재는 사용자 경험을 위해 유지합니다 (예: 24h 시간 범위 유지)
    // setFilters(prev => ({ ...defaultFilters, timeRange: prev.timeRange }));
  }, []);

  const updateFilter = useCallback(<K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => {
    setFilters((prev) => ({
      ...prev,
      [key]: value,
    }));
  }, []);

  const resetFilters = useCallback(() => {
    setFilters(defaultFilters);
  }, []);

  return (
    <SearchContext.Provider 
      value={{ 
        scope, 
        setScope, 
        filters, 
        setFilters, 
        updateFilter, 
        resetFilters 
      }}
    >
      {children}
    </SearchContext.Provider>
  );
};

// 5. Custom Hook
// eslint-disable-next-line react-refresh/only-export-components
export const useSearch = () => {
  const context = useContext(SearchContext);
  if (!context) {
    throw new Error('useSearch must be used within a SearchProvider');
  }
  return context;
};