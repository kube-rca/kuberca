import { useEffect, useRef, useCallback, useState } from 'react';

export type PollingStatus = 'active' | 'paused' | 'backoff' | 'disabled';

interface UsePollingOptions {
  callback: () => Promise<void> | void;
  interval: number;
  pauseOnHidden?: boolean;
  backoffMultiplier?: number;
  maxInterval?: number;
  enabled?: boolean;
  sseConnected?: boolean;
}

interface UsePollingReturn {
  refresh: () => void;
  status: PollingStatus;
  lastUpdated: Date | null;
}

export function usePolling(options: UsePollingOptions): UsePollingReturn {
  const {
    callback,
    interval,
    pauseOnHidden = true,
    backoffMultiplier = 2,
    maxInterval = 120_000,
    enabled = true,
    sseConnected = false,
  } = options;

  const [status, setStatus] = useState<PollingStatus>('disabled');
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const callbackRef = useRef(callback);
  const intervalIdRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const errorCountRef = useRef(0);
  const mountedRef = useRef(true);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  const clearTimer = useCallback(() => {
    if (intervalIdRef.current !== null) {
      clearTimeout(intervalIdRef.current);
      intervalIdRef.current = null;
    }
  }, []);

  const getBackoffInterval = useCallback(() => {
    if (errorCountRef.current === 0) return interval;
    const backoff = interval * Math.pow(backoffMultiplier, errorCountRef.current);
    return Math.min(backoff, maxInterval);
  }, [interval, backoffMultiplier, maxInterval]);

  const executeCallback = useCallback(async () => {
    try {
      await callbackRef.current();
      errorCountRef.current = 0;
      if (mountedRef.current) {
        setLastUpdated(new Date());
        setStatus('active');
      }
    } catch {
      errorCountRef.current += 1;
      if (mountedRef.current) {
        setStatus('backoff');
      }
    }
  }, []);

  const scheduleNext = useCallback(() => {
    clearTimer();
    const nextInterval = getBackoffInterval();
    intervalIdRef.current = setTimeout(async () => {
      if (!mountedRef.current) return;
      await executeCallback();
      if (mountedRef.current) {
        scheduleNext();
      }
    }, nextInterval);
  }, [clearTimer, getBackoffInterval, executeCallback]);

  const refresh = useCallback(() => {
    errorCountRef.current = 0;
    clearTimer();
    executeCallback().then(() => {
      if (mountedRef.current && enabled && !sseConnected) {
        scheduleNext();
      }
    });
  }, [clearTimer, executeCallback, enabled, sseConnected, scheduleNext]);

  // Main effect: start/stop polling based on enabled and sseConnected
  useEffect(() => {
    if (!enabled) {
      clearTimer();
      setStatus('disabled');
      return;
    }

    if (sseConnected) {
      clearTimer();
      setStatus('paused');
      return;
    }

    // Start polling
    setStatus('active');
    scheduleNext();

    return () => {
      clearTimer();
    };
  }, [enabled, sseConnected, clearTimer, scheduleNext]);

  // Unmount cleanup
  useEffect(() => {
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // Page Visibility API
  useEffect(() => {
    if (!pauseOnHidden || !enabled) return;

    const handleVisibility = () => {
      if (!mountedRef.current) return;

      if (document.visibilityState === 'hidden') {
        clearTimer();
        if (!sseConnected) setStatus('paused');
      } else {
        // Visible again: immediate refresh then resume schedule
        if (!sseConnected) {
          errorCountRef.current = 0;
          executeCallback().then(() => {
            if (mountedRef.current && enabled && !sseConnected) {
              scheduleNext();
            }
          });
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibility);
    return () => document.removeEventListener('visibilitychange', handleVisibility);
  }, [pauseOnHidden, enabled, sseConnected, clearTimer, executeCallback, scheduleNext]);

  return { refresh, status, lastUpdated };
}
