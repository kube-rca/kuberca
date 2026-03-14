import { useState, useEffect, useCallback, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { requestWithAuth } from '../utils/api';

interface AnalysisConfig {
  manualAnalyzeSeverities: string;
}

interface UseAnalysisModeReturn {
  manualSeverities: string[];
  loading: boolean;
  error: boolean;
  refetch: () => void;
}

export function useAnalysisMode(): UseAnalysisModeReturn {
  const [manualSeverities, setManualSeverities] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const prevPathRef = useRef<string>('');
  const location = useLocation();

  const fetchMode = useCallback(async () => {
    try {
      setLoading(true);
      setError(false);
      const response = await requestWithAuth('/api/v1/settings/app/analysis', {
        method: 'GET',
      });
      if (response.ok) {
        const data = await response.json();
        const config: AnalysisConfig = data.data ?? data;
        if (config.manualAnalyzeSeverities) {
          const parsed = config.manualAnalyzeSeverities
            .split(',')
            .map((s: string) => s.trim())
            .filter(Boolean);
          setManualSeverities(parsed);
        } else {
          setManualSeverities([]);
        }
      } else {
        setError(true);
      }
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial fetch
  useEffect(() => {
    fetchMode();
  }, [fetchMode]);

  // Refetch when navigating away from /settings/analysis
  useEffect(() => {
    const prev = prevPathRef.current;
    prevPathRef.current = location.pathname;

    if (prev.startsWith('/settings/analysis') && !location.pathname.startsWith('/settings/analysis')) {
      fetchMode();
    }
  }, [location.pathname, fetchMode]);

  return { manualSeverities, loading, error, refetch: fetchMode };
}
