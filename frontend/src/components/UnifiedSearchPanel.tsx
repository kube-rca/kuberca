import React, { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom'; // [추가] 페이지 이동 훅
import { useSearch } from '../context/SearchContext';

interface UnifiedSearchPanelProps {
  availableLabels: string[];
  availableNamespaces: string[];
}

const SearchIcon = () => (
  <svg className="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
  </svg>
);

const FilterIcon = () => (
  <svg className="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
  </svg>
);

const TIME_OPTIONS = [
  { label: 'Last 1h', value: '1h' },
  { label: 'Last 6h', value: '6h' },
  { label: 'Last 24h', value: '24h' },
  { label: 'Last 7d', value: '7d' },
  { label: 'Last 30d', value: '30d' },
];

const UnifiedSearchPanel: React.FC<UnifiedSearchPanelProps> = ({ availableLabels, availableNamespaces }) => {
  const { scope, setScope, filters, updateFilter, resetFilters } = useSearch();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  // [추가] 네비게이션 훅 사용
  const navigate = useNavigate();
  const location = useLocation();

  // Dashboard 페이지에 따라 검색 Target 기본값 동기화
  useEffect(() => {
    if (location.pathname === '/alerts' && scope !== 'ALERT') {
      setScope('ALERT');
      return;
    }
    if ((location.pathname === '/' || location.pathname.startsWith('/muted')) && scope !== 'INCIDENT') {
      setScope('INCIDENT');
    }
  }, [location.pathname, scope, setScope]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // [핵심 추가] 결과 보기 버튼 클릭 핸들러
  const handleShowResults = () => {
    setIsDropdownOpen(false); // 드롭다운 Close

    // 현재 선택된 타겟(Scope)에 따라 페이지 이동
    if (scope === 'INCIDENT') {
      navigate('/'); // Incident Dashboard로 이동
    } else if (scope === 'ALERT') {
      navigate('/alerts'); // Alert Dashboard로 이동
    }
    // 필터 상태는 Context에 저장되어 있으므로 페이지가 이동해도 유지됩니다.
  };

  return (
    <div className="w-full mb-6 relative" ref={panelRef}>
      
      {/* 1. 메인 검색 바 */}
      <div className={`flex items-center bg-white dark:bg-slate-800 border rounded-md shadow-sm transition-all h-12 ${
        isDropdownOpen 
          ? 'border-cyan-500 ring-2 ring-cyan-100 dark:ring-cyan-900' 
          : 'border-slate-300 dark:border-slate-600 hover:border-slate-400'
      }`}>
        <div className="pl-4 pr-3"><SearchIcon /></div>
        <input
          type="text"
          className="flex-1 bg-transparent border-none focus:ring-0 text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 h-full"
          placeholder={`${scope === 'INCIDENT' ? 'Incident' : 'Alert'} Search (Title, ID, Summary, etc.)`}
          value={filters.query}
          onChange={(e) => updateFilter('query', e.target.value)}
          onFocus={() => setIsDropdownOpen(true)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleShowResults();
          }}
        />
        <button 
          className={`px-5 h-full border-l border-slate-200 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 rounded-r-md transition-colors flex items-center gap-2 text-sm font-medium ${
            isDropdownOpen ? 'bg-slate-50 dark:bg-slate-700' : ''
          }`}
          onClick={() => setIsDropdownOpen(!isDropdownOpen)}
        >
          <FilterIcon />
          <span>Filter</span>
        </button>
      </div>

      {/* 2. 상세 필터 드롭다운 */}
      {isDropdownOpen && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-slate-800 rounded-lg shadow-xl border border-slate-200 dark:border-slate-700 flex flex-col md:flex-row overflow-hidden min-h-[400px] z-50">
          
          {/* [왼쪽 패널] Target, Status, Severity, Time Range */}
          <div className="w-full md:w-5/12 p-5 border-b md:border-b-0 md:border-r border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-900/20 flex flex-col gap-6">
            
            {/* Target Tab */}
            <div>
              <label className="text-xs font-bold text-slate-400 mb-3 uppercase tracking-wider block">Target</label>
              <div className="flex bg-slate-200 dark:bg-slate-700 rounded-lg p-1">
                <button
                  onClick={() => setScope('INCIDENT')}
                  className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-all ${
                    scope === 'INCIDENT' ? 'bg-white dark:bg-slate-600 text-cyan-600 dark:text-cyan-300 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-700'
                  }`}
                >
                  Incidents
                </button>
                <button
                  onClick={() => setScope('ALERT')}
                  className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-all ${
                    scope === 'ALERT' ? 'bg-white dark:bg-slate-600 text-cyan-600 dark:text-cyan-300 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-700'
                  }`}
                >
                  Alerts
                </button>
              </div>
            </div>

            {/* Status Filters */}
            <div>
              <label className="text-xs font-bold text-slate-400 mb-3 uppercase tracking-wider block">Status</label>
              <div className="flex flex-wrap gap-2">
                {['firing', 'resolved'].map((status) => (
                  <FilterChip 
                    key={status}
                    label={status === 'firing' ? '🔥 Firing' : '✅ Resolved'}
                    active={filters.status.includes(status)}
                    onClick={() => {
                      const newStatus = filters.status.includes(status)
                        ? filters.status.filter(s => s !== status)
                        : [...filters.status, status];
                      updateFilter('status', newStatus);
                    }}
                    color={status === 'firing' ? 'red' : 'green'}
                  />
                ))}
              </div>
            </div>

            {/* Severity Filters */}
            <div>
              <label className="text-xs font-bold text-slate-400 mb-3 uppercase tracking-wider block">Severity</label>
              <div className="flex flex-wrap gap-2">
                {(scope === 'INCIDENT'
                  ? ['Critical', 'Warning', 'Info', 'TBD']
                  : ['Critical', 'Warning', 'Info']
                ).map((sev) => (
                  <FilterChip 
                    key={sev}
                    label={sev}
                    active={filters.severity.includes(sev)}
                    onClick={() => {
                      const newSev = filters.severity.includes(sev)
                        ? filters.severity.filter(s => s !== sev)
                        : [...filters.severity, sev];
                      updateFilter('severity', newSev);
                    }}
                    color="blue"
                  />
                ))}
              </div>
            </div>

            {/* Time Range */}
            <div>
              <label className="text-xs font-bold text-slate-400 mb-3 uppercase tracking-wider block">Time Range</label>
              <div className="flex flex-wrap gap-2">
                {TIME_OPTIONS.map((opt) => (
                  <FilterChip 
                    key={opt.value}
                    label={opt.label}
                    active={filters.timeRange === opt.value}
                    onClick={() =>
                      updateFilter(
                        'timeRange',
                        filters.timeRange === opt.value ? 'all' : opt.value
                      )
                    }
                    color="purple"
                  />
                ))}
              </div>
            </div>
          </div>

          {/* [오른쪽 패널] Namespace, Labels */}
          <div className="w-full md:w-7/12 p-5 flex flex-col h-full">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 flex-1">
              
              {/* Namespace Selection */}
              <div className="flex flex-col">
                <label className="block text-xs font-bold text-slate-500 mb-2 uppercase">Namespace</label>
                {/* [수정] h-72로 높이 고정 + overflow-y-auto로 스크롤 생성 */}
                <div className="h-72 overflow-y-auto space-y-1.5 p-3 border border-slate-200 dark:border-slate-700 rounded-md bg-white dark:bg-slate-900/30 shadow-inner scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
                  {availableNamespaces.length === 0 && (
                     <div className="text-xs text-slate-400 text-center py-4">No namespaces</div>
                  )}
                  {availableNamespaces.map(ns => (
                    <label key={ns} className="flex items-center gap-2.5 text-sm text-slate-700 dark:text-slate-300 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 p-1 rounded transition-colors">
                      <input 
                        type="checkbox" 
                        className="rounded text-cyan-600 focus:ring-cyan-500 w-4 h-4 border-slate-300"
                        checked={filters.namespaces.includes(ns)}
                        onChange={(e) => {
                          const newNs = e.target.checked
                            ? [...filters.namespaces, ns]
                            : filters.namespaces.filter(n => n !== ns);
                          updateFilter('namespaces', newNs);
                        }}
                      />
                      {ns}
                    </label>
                  ))}
                </div>
              </div>

              {/* Labels Selection */}
              <div className="flex flex-col">
                <label className="block text-xs font-bold text-slate-500 mb-2 uppercase">Labels</label>
                {/* [수정] h-72로 높이 고정 + overflow-y-auto로 스크롤 생성 */}
                <div className="h-72 overflow-y-auto space-y-1.5 p-3 border border-slate-200 dark:border-slate-700 rounded-md bg-white dark:bg-slate-900/30 shadow-inner scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
                  {availableLabels.length === 0 && (
                     <div className="text-xs text-slate-400 text-center py-4">No labels</div>
                  )}
                  
                  {availableLabels.map(labelStr => {
                    const separatorIndex = labelStr.indexOf(':');
                    if (separatorIndex === -1) return null;
                  
                    const key = labelStr.slice(0, separatorIndex);
                    const value = labelStr.slice(separatorIndex + 1);
                    const isChecked = filters.labels.includes(labelStr);

                    return (
                      <label
                        key={labelStr}
                        className="flex items-center gap-2.5 text-sm text-slate-700 dark:text-slate-300 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 p-1 rounded transition-colors"
                      >
                        <input
                          type="checkbox"
                          className="rounded text-cyan-600 focus:ring-cyan-500 w-4 h-4 border-slate-300 flex-shrink-0"
                          checked={isChecked}
                          onChange={(e) => {
                            let newLabels = [...filters.labels];
                            if (e.target.checked) {
                              if (!newLabels.includes(labelStr)) {
                                newLabels.push(labelStr);
                              }
                            } else {
                              newLabels = newLabels.filter(l => l !== labelStr);
                            }
                            updateFilter('labels', newLabels);
                          }}
                        />
                        <div className="flex flex-col min-w-0">
                          <span className="font-mono text-xs bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded text-slate-600 dark:text-slate-300 font-bold mb-0.5">
                            {key}
                          </span>
                          <span className="text-xs text-slate-700 dark:text-slate-300 truncate" title={value}>
                            {value}
                          </span>
                        </div>
                      </label>
                    );
                  })}
                </div>
              </div>

            </div>

            {/* 하단 버튼 */}
            <div className="flex justify-end gap-3 pt-6 mt-auto border-t border-slate-100 dark:border-slate-700">
              <button
                onClick={() => resetFilters()}
                className="px-4 py-2 text-sm font-medium text-slate-500 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 rounded transition-colors border border-slate-200 dark:border-slate-600"
              >
                Reset Filters
              </button>
              <button 
                onClick={() => setIsDropdownOpen(false)} 
                className="px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700 rounded transition-colors"
              >
                Close
              </button>
              {/* [수정] View Results 클릭 시 handleShowResults 실행 */}
              <button 
                onClick={handleShowResults} 
                className="px-6 py-2 text-sm font-medium text-white bg-cyan-600 hover:bg-cyan-700 rounded shadow-sm transition-colors"
              >
                View Results
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// 칩 컴포넌트
interface FilterChipProps {
  label: string;
  active: boolean;
  onClick: () => void;
  color: 'red' | 'green' | 'blue' | 'purple';
}

const FilterChip: React.FC<FilterChipProps> = ({ label, active, onClick, color }) => {
  const baseStyles = "px-3 py-1.5 rounded-full text-xs font-semibold border transition-all duration-200 select-none whitespace-nowrap";
  
  const colorStyles = {
    red: active 
      ? "bg-rose-100 text-rose-700 border-rose-200 ring-1 ring-rose-400 dark:bg-rose-900/40 dark:text-rose-300 dark:border-rose-800" 
      : "bg-white text-slate-600 border-slate-200 hover:border-rose-300 hover:text-rose-600 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700",
    green: active
      ? "bg-emerald-100 text-emerald-700 border-emerald-200 ring-1 ring-emerald-400 dark:bg-emerald-900/40 dark:text-emerald-300 dark:border-emerald-800"
      : "bg-white text-slate-600 border-slate-200 hover:border-emerald-300 hover:text-emerald-600 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700",
    blue: active
      ? "bg-cyan-100 text-cyan-700 border-cyan-200 ring-1 ring-cyan-400 dark:bg-cyan-900/40 dark:text-cyan-300 dark:border-cyan-800"
      : "bg-white text-slate-600 border-slate-200 hover:border-cyan-300 hover:text-cyan-600 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700",
    purple: active
      ? "bg-violet-100 text-violet-700 border-violet-200 ring-1 ring-violet-400 dark:bg-violet-900/40 dark:text-violet-300 dark:border-violet-800"
      : "bg-white text-slate-600 border-slate-200 hover:border-violet-300 hover:text-violet-600 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700",
  };

  return (
    <button onClick={onClick} className={`${baseStyles} ${colorStyles[color]}`}>
      {label}
    </button>
  );
};

export default UnifiedSearchPanel;
