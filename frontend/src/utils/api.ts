import { RCAItem, RCADetail } from '../types';
import { API_BASE_URL } from './config';
import { getAccessToken, refreshAccessToken } from './auth';

// 인증 토큰을 포함해 요청하고 401이면 토큰 갱신 후 1회 재시도한다.
export const requestWithAuth = async (path: string, init: RequestInit = {}) => {
  const headers = new Headers(init.headers);
  const token = getAccessToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
    credentials: 'include',
  });

  if (response.status !== 401) {
    return response;
  }

  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    throw new Error('unauthorized');
  }

  const retryHeaders = new Headers(init.headers);
  const newToken = getAccessToken();
  if (newToken) {
    retryHeaders.set('Authorization', `Bearer ${newToken}`);
  }

  return fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: retryHeaders,
    credentials: 'include',
  });
};

/**
 * 백엔드에서 RCA(Incident) 목록을 가져옵니다.
 * 응답은 RCAItem[] 배열을 기대합니다.
 */
export const fetchRCAs = async (): Promise<RCAItem[]> => {
  const response = await requestWithAuth('/api/v1/incidents', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data: RCAItem[] = await response.json();
  if (Array.isArray(data)) {
    return data;
  }

  console.warn('Backend response is not an array:', data);
  throw new Error('Unexpected response format: Data is not an array');
};

export const fetchMutedIncidents = async (): Promise<RCAItem[]> => {
  const response = await requestWithAuth('/api/v1/incidents/hidden', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data: RCAItem[] = await response.json();
  if (Array.isArray(data)) {
    return data;
  }

  console.warn('Backend response is not an array:', data);
  throw new Error('Unexpected response format: Data is not an array');
};

/**
 * RCA 단건 상세를 가져옵니다. 응답은 { data: RCADetail } 형태를 기대합니다.
 */
export const fetchRCADetail = async (id: string): Promise<RCADetail> => {
  const response = await requestWithAuth(`/api/v1/incidents/${id}`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('상세 정보를 불러오는데 실패했습니다.');
  }

  const json = await response.json();
  return json.data;
};

export const updateRCADetail = async (id: string, data: Partial<RCADetail>): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/incidents/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new Error('RCA 정보를 수정하는데 실패했습니다.');
  }
};

export const hideIncident = async (id: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/incidents/${id}`, {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('리포트를 숨기는데 실패했습니다.');
  }
  
  // response body가 없는 경우를 대비해 return response.json() 대신 void 처리
  // 만약 백엔드가 json을 준다면 return response.json() 유지
};

export const unhideIncident = async (id: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/incidents/${id}/unhide`, {
    method: 'PATCH',
  });

  if (!response.ok) {
    throw new Error('인시던트 숨김 해제에 실패했습니다.');
  }
};

/**
 * 인시던트를 종료(resolved) 상태로 변경합니다.
 * 백엔드에서 Agent로 최종 분석 요청을 보내고, 분석 결과가 DB에 저장됩니다.
 */
export const resolveIncident = async (id: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/incidents/${id}/resolve`, {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('인시던트 종료에 실패했습니다.');
  }
};

/**
 * Alert에 대해 수동 분석을 트리거합니다.
 */
export const triggerAlertAnalysis = async (alertId: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/alerts/${alertId}/analyze`, {
    method: 'POST',
  });
  if (!response.ok) throw new Error('분석 요청 실패');
};

/**
 * Incident에 대해 수동 분석을 트리거합니다.
 */
export const triggerIncidentAnalysis = async (incidentId: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/incidents/${incidentId}/analyze`, {
    method: 'POST',
  });
  if (!response.ok) throw new Error('인시던트 분석 요청 실패');
};

// ============================================================================
// Alert API
// ============================================================================

export interface AlertItem {
  alert_id: string;
  incident_id: string | null;
  alarm_title: string;
  namespace: string;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
  labels: Record<string, string>;
}

export interface AlertDetail {
  alert_id: string;
  incident_id: string | null;
  alarm_title: string;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
  analysis_summary: string;
  analysis_detail: string;
  fingerprint: string;
  thread_ts: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  is_analyzing?: boolean;
}

/**
 * 백엔드에서 Alert 목록을 가져옵니다.
 */
export const fetchAlerts = async (): Promise<AlertItem[]> => {
  const response = await requestWithAuth('/api/v1/alerts', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data: AlertItem[] = await response.json();
  if (Array.isArray(data)) {
    return data;
  }

  console.warn('Backend response is not an array:', data);
  throw new Error('Unexpected response format: Data is not an array');
};

/**
 * Alert 단건 상세를 가져옵니다.
 */
export const fetchAlertDetail = async (id: string): Promise<AlertDetail> => {
  const response = await requestWithAuth(`/api/v1/alerts/${id}`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('Alert 상세 정보를 불러오는데 실패했습니다.');
  }

  return response.json();
};

/**
 * Alert의 연결된 Incident를 변경합니다.
 */
export const updateAlertIncident = async (alertId: string, incidentId: string): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/alerts/${alertId}/incident`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ incident_id: incidentId }),
  });

  if (!response.ok) {
    throw new Error('Alert의 Incident 연결을 변경하는데 실패했습니다.');
  }
};

// ============================================================================
// Embedding API
// ============================================================================

export interface EmbeddingSearchResult {
  incident_id: string;
  incident_summary: string;
  similarity: number;
}

export interface EmbeddingSearchResponse {
  results: EmbeddingSearchResult[];
  model: string;
}

/**
 * 임베딩 벡터로 유사한 인시던트를 검색합니다.
 */
export const searchSimilarIncidents = async (query: string, limit: number = 5): Promise<EmbeddingSearchResponse> => {
  const response = await requestWithAuth('/api/v1/embeddings/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ query, limit }),
  });

  if (!response.ok) {
    throw new Error('유사 인시던트 검색에 실패했습니다.');
  }

  return response.json();
};

// ============================================================================
// Feedback API
// ============================================================================

export type FeedbackTargetType = 'incident' | 'alert';
export type FeedbackVoteType = 'up' | 'down' | 'none';

export interface FeedbackCommentResponse {
  comment_id: number;
  target_type: FeedbackTargetType;
  target_id: string;
  user_id: number;
  author_login_id: string;
  body: string;
  created_at: string;
}

export interface FeedbackSummaryResponse {
  target_type: FeedbackTargetType;
  target_id: string;
  up_votes: number;
  down_votes: number;
  my_vote?: 'up' | 'down';
  comments: FeedbackCommentResponse[];
}

const getFeedbackBasePath = (targetType: FeedbackTargetType, targetId: string): string => {
  if (targetType === 'incident') {
    return `/api/v1/incidents/${targetId}`;
  }
  return `/api/v1/alerts/${targetId}`;
};

export const fetchFeedbackSummary = async (targetType: FeedbackTargetType, targetId: string): Promise<FeedbackSummaryResponse> => {
  const response = await requestWithAuth(`${getFeedbackBasePath(targetType, targetId)}/feedback`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('피드백 정보를 불러오는데 실패했습니다.');
  }

  return response.json();
};

export const createFeedbackComment = async (
  targetType: FeedbackTargetType,
  targetId: string,
  body: string
): Promise<FeedbackCommentResponse> => {
  const response = await requestWithAuth(`${getFeedbackBasePath(targetType, targetId)}/comments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ body }),
  });

  if (!response.ok) {
    throw new Error('코멘트 저장에 실패했습니다.');
  }

  return response.json();
};

export const voteFeedback = async (
  targetType: FeedbackTargetType,
  targetId: string,
  voteType: FeedbackVoteType
): Promise<void> => {
  const response = await requestWithAuth(`${getFeedbackBasePath(targetType, targetId)}/vote`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ vote_type: voteType }),
  });

  if (!response.ok) {
    throw new Error('투표 저장에 실패했습니다.');
  }
};

export const updateFeedbackComment = async (
  targetType: FeedbackTargetType,
  targetId: string,
  commentId: number,
  body: string
): Promise<FeedbackCommentResponse> => {
  const response = await requestWithAuth(`${getFeedbackBasePath(targetType, targetId)}/comments/${commentId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ body }),
  });

  if (!response.ok) {
    throw new Error('코멘트 수정에 실패했습니다.');
  }

  return response.json();
};

export const deleteFeedbackComment = async (
  targetType: FeedbackTargetType,
  targetId: string,
  commentId: number
): Promise<void> => {
  const response = await requestWithAuth(`${getFeedbackBasePath(targetType, targetId)}/comments/${commentId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('코멘트 삭제에 실패했습니다.');
  }
};

// ============================================================================
// Chat API
// ============================================================================

export interface ChatRequest {
  message: string;
  conversation_id: string;
  page: string;
  auto: boolean;
  incident_id: string;
  alert_id: string;
  incident_title: string;
  incident_content: string;
  alert_title: string;
  alert_content: string;
}

export interface ChatResponse {
  status: string;
  answer: string;
  conversation_id?: string;
}

export const chatWithAgent = async (payload: ChatRequest): Promise<ChatResponse> => {
  const response = await requestWithAuth('/api/v1/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `채팅 요청 실패 (${response.status})`);
  }

  return response.json();
};

// ============================================================================
// Webhook Settings API
// ============================================================================

export interface WebhookHeaderItem {
  key: string;
  value: string;
}

export interface WebhookConfigPayload {
  url: string;
  type: 'slack' | 'teams' | 'http';
  token?: string;
  channel?: string;
}

export interface WebhookConfig extends WebhookConfigPayload {
  id: number;
  updated_at: string;
}

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
};

const toStringValue = (value: unknown): string => {
  return typeof value === 'string' ? value : '';
};

const getRecordValue = (record: Record<string, unknown>, ...keys: string[]): unknown => {
  for (const key of keys) {
    if (key in record) {
      return record[key];
    }
  }
  return undefined;
};

const detectWebhookType = (raw: Record<string, unknown>): 'slack' | 'teams' | 'http' => {
  const declaredType = toStringValue(getRecordValue(raw, 'type', 'Type')).trim().toLowerCase();
  if (declaredType === 'slack' || declaredType === 'teams' || declaredType === 'http') {
    return declaredType;
  }
  return 'http';
};

const normalizeWebhookConfig = (raw: unknown): WebhookConfig => {
  const record = isRecord(raw) ? raw : {};

  return {
    id: Number(getRecordValue(record, 'id', 'ID') ?? 0),
    url: toStringValue(getRecordValue(record, 'url', 'URL')),
    type: detectWebhookType(record),
    token: toStringValue(getRecordValue(record, 'token', 'Token')).trim() || undefined,
    channel: toStringValue(getRecordValue(record, 'channel', 'Channel')).trim() || undefined,
    updated_at:
      toStringValue(getRecordValue(record, 'updated_at', 'updatedAt', 'UpdatedAt')) ||
      new Date(0).toISOString(),
  };
};

const unwrapWebhookPayload = (raw: unknown): unknown => {
  let current = raw;
  for (let i = 0; i < 3; i += 1) {
    if (!isRecord(current) || !('data' in current)) {
      break;
    }
    current = current.data;
  }
  return current;
};

/** 웹훅 설정 목록 조회 */
export const fetchWebhookList = async (): Promise<WebhookConfig[]> => {
  const response = await requestWithAuth('/api/v1/settings/webhooks', { method: 'GET' });
  if (!response.ok) throw new Error(`웹훅 목록 조회 실패 (${response.status})`);
  const payload = unwrapWebhookPayload(await response.json());
  if (!Array.isArray(payload)) {
    return [];
  }
  return payload.map(normalizeWebhookConfig);
};

/** ID로 단건 조회 */
export const fetchWebhookById = async (id: number): Promise<WebhookConfig> => {
  const response = await requestWithAuth(`/api/v1/settings/webhooks/${id}`, { method: 'GET' });
  if (!response.ok) throw new Error(`웹훅 설정 조회 실패 (${response.status})`);
  const payload = unwrapWebhookPayload(await response.json());
  return normalizeWebhookConfig(payload);
};

/** 신규 웹훅 설정 생성 */
export const createWebhookConfig = async (payload: WebhookConfigPayload): Promise<number> => {
  const response = await requestWithAuth('/api/v1/settings/webhooks', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `웹훅 설정 생성 실패 (${response.status})`);
  }
  const json = await response.json();
  return json.id;
};

/** 기존 웹훅 설정 수정 */
export const updateWebhookConfig = async (id: number, payload: WebhookConfigPayload): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/settings/webhooks/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `웹훅 설정 수정 실패 (${response.status})`);
  }
};

/** 웹훅 설정 삭제 */
export const deleteWebhookConfig = async (id: number): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/settings/webhooks/${id}`, { method: 'DELETE' });
  if (!response.ok) throw new Error(`웹훅 설정 삭제 실패 (${response.status})`);
};


export interface WebhookSettings {
  url: string;
  method: string;
  headers: WebhookHeaderItem[];
  body: string;
}

export interface WebhookSettingsDetail extends WebhookSettings {
  id: number;
  updated_at: string;
}

/**
 * 웹훅 설정을 저장합니다.
 */
export const saveWebhookSettings = async (settings: WebhookSettings): Promise<void> => {
  const response = await requestWithAuth('/api/v1/settings/webhook', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(settings),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `웹훅 설정 저장 실패 (${response.status})`);
  }
};

/**
 * 저장된 웹훅 설정을 불러옵니다.
 */
export const fetchWebhookSettings = async (): Promise<WebhookSettingsDetail | null> => {
  const response = await requestWithAuth('/api/v1/settings/webhook', {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`웹훅 설정 조회 실패 (${response.status})`);
  }

  const json = await response.json();
  return json.data ?? null;
};

// ============================================================================
// App Settings API
// ============================================================================

export interface FlappingSettings {
  enabled: boolean;
  detectionWindowMinutes: number;
  cycleThreshold: number;
  clearanceWindowMinutes: number;
}

export interface AISettings {
  provider: string;
  modelId: string;
}

// ============================================================================
// Analytics API
// ============================================================================

export interface AnalyticsCountItem {
  key: string;
  count: number;
}

export interface AnalyticsSeriesPoint {
  date: string;
  incidents: number;
  alerts: number;
}

export interface AnalyticsDashboardResponse {
  window: string;
  generated_at: string;
  summary: {
    total_incidents: number;
    firing_incidents: number;
    resolved_incidents: number;
    total_alerts: number;
    firing_alerts: number;
    resolved_alerts: number;
    avg_mttr_minutes: number;
    avg_alerts_per_incident: number;
  };
  breakdown: {
    incident_severity: AnalyticsCountItem[];
    alert_severity: AnalyticsCountItem[];
    top_namespaces: AnalyticsCountItem[];
  };
  series: {
    daily: AnalyticsSeriesPoint[];
  };
}

export const fetchAnalyticsDashboard = async (window: string = '30d'): Promise<AnalyticsDashboardResponse> => {
  const response = await requestWithAuth(`/api/v1/analytics/dashboard?window=${encodeURIComponent(window)}`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`분석 데이터 조회 실패 (${response.status})`);
  }

  return response.json();
};

/** 개별 앱 설정 조회 (ENV fallback 포함) */
export const fetchAppSetting = async <T>(key: string): Promise<T> => {
  const response = await requestWithAuth(`/api/v1/settings/app/${key}`, { method: 'GET' });
  if (!response.ok) throw new Error(`앱 설정 조회 실패 (${response.status})`);
  const json = await response.json();
  return json.data?.value as T;
};

/** 앱 설정 저장 */
export const updateAppSetting = async <T>(key: string, value: T): Promise<void> => {
  const response = await requestWithAuth(`/api/v1/settings/app/${key}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(value),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `앱 설정 저장 실패 (${response.status})`);
  }
};
