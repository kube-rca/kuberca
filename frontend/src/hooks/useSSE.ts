import { useEffect, useRef, useCallback, useState } from 'react';
import { API_BASE_URL } from '../utils/config';
import { getAccessToken, refreshAccessToken } from '../utils/auth';

export type SSEEventType =
  | 'alert_created'
  | 'alert_resolved'
  | 'analysis_completed'
  | 'incident_created'
  | 'incident_updated'
  | 'incident_resolved'
  | 'heartbeat';

export interface SSEEventData {
  alert_id?: string;
  incident_id?: string;
  message?: string;
}

export interface SSEEvent {
  type: SSEEventType;
  timestamp: string;
  data: SSEEventData;
}

export type SSEConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

interface UseSSEOptions {
  onEvent: (event: SSEEvent) => void;
  enabled: boolean;
  maxRetries?: number;
  initialRetryDelay?: number;
  maxRetryDelay?: number;
}

interface UseSSEReturn {
  connectionState: SSEConnectionState;
  reconnect: () => void;
}

export function useSSE(options: UseSSEOptions): UseSSEReturn {
  const {
    onEvent,
    enabled,
    maxRetries = 5,
    initialRetryDelay = 1000,
    maxRetryDelay = 30_000,
  } = options;

  const [connectionState, setConnectionState] = useState<SSEConnectionState>('disconnected');
  const eventSourceRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const onEventRef = useRef(onEvent);
  const enabledRef = useRef(enabled);
  const mountedRef = useRef(true);

  useEffect(() => {
    onEventRef.current = onEvent;
  }, [onEvent]);

  useEffect(() => {
    enabledRef.current = enabled;
  }, [enabled]);

  // Unmount guard
  useEffect(() => {
    return () => { mountedRef.current = false; };
  }, []);

  const closeConnection = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (retryTimerRef.current) {
      clearTimeout(retryTimerRef.current);
      retryTimerRef.current = undefined;
    }
  }, []);

  const connect = useCallback(() => {
    closeConnection();

    const token = getAccessToken();
    if (!token) {
      setConnectionState('disconnected');
      return;
    }

    const url = `${API_BASE_URL}/api/v1/events?token=${encodeURIComponent(token)}`;
    const es = new EventSource(url);
    eventSourceRef.current = es;
    setConnectionState('connecting');

    es.onopen = () => {
      setConnectionState('connected');
      retryCountRef.current = 0;
    };

    es.addEventListener('message', (e: MessageEvent) => {
      try {
        const event: SSEEvent = JSON.parse(e.data);
        onEventRef.current(event);
      } catch {
        // Ignore malformed events
      }
    });

    es.onerror = () => {
      es.close();
      eventSourceRef.current = null;
      if (!mountedRef.current) return;
      setConnectionState('error');

      if (retryCountRef.current < maxRetries) {
        const delay = Math.min(
          initialRetryDelay * Math.pow(2, retryCountRef.current),
          maxRetryDelay
        );
        retryCountRef.current += 1;

        retryTimerRef.current = setTimeout(async () => {
          if (!mountedRef.current || !enabledRef.current) return;

          // Try refreshing the token before reconnecting
          const refreshed = await refreshAccessToken();
          if (!mountedRef.current) return;
          if (refreshed && enabledRef.current) {
            connect();
          } else {
            setConnectionState('disconnected');
          }
        }, delay);
      } else {
        // Max retries exceeded — fall back to polling
        setConnectionState('disconnected');
      }
    };
  }, [closeConnection, maxRetries, initialRetryDelay, maxRetryDelay]);

  const reconnect = useCallback(() => {
    retryCountRef.current = 0;
    connect();
  }, [connect]);

  // Connect/disconnect based on enabled state
  useEffect(() => {
    if (enabled) {
      connect();
    } else {
      closeConnection();
      setConnectionState('disconnected');
    }

    return () => {
      closeConnection();
    };
  }, [enabled, connect, closeConnection]);

  // Page Visibility: reconnect when tab becomes visible
  useEffect(() => {
    if (!enabled) return;

    const handleVisibility = () => {
      if (document.visibilityState === 'visible') {
        if (!eventSourceRef.current || eventSourceRef.current.readyState === EventSource.CLOSED) {
          retryCountRef.current = 0;
          connect();
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibility);
    return () => document.removeEventListener('visibilitychange', handleVisibility);
  }, [enabled, connect]);

  return { connectionState, reconnect };
}
