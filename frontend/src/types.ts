export type Severity = 'TBD' | 'warning' | 'critical' | 'Warning' | 'Critical' | 'Resolved';

export interface RCAItem {
  incident_id: string;
  title: string;
  fired_at: string;
  resolved_at: string | null;
  severity: Severity;
}

// Alert 목록 아이템 (인시던트 상세에서 사용)
export interface AlertItem {
  alert_id: string;
  incident_id: string | null;
  alarm_title: string;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
  namespace: string;
  labels: Record<string, string>;
  analysis_summary?: string;
}

export interface RCADetail {
  incident_id: string;
  title: string;
  severity: Severity;
  status: string;     // 예: "Firing"
  fired_at: string;
  resolved_at: string | null;
  analysis_summary: string;
  analysis_detail: string;
  is_hidden?: boolean;
  similar_incidents?: SimilarIncident[]; // 필요 시 구체화
  alerts?: AlertItem[]; // 연결된 Alert 목록
}

// 유사 인시던트 객체 타입 정의
export interface SimilarIncident {
  incident_id: string;
  title: string;
  score?: number; // 유사도
}
