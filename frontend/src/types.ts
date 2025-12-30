export type Severity = 'info' | 'warning' | 'critical';

export interface RCAItem {
  incident_id: string;  
  alarm_title: string; 
  fired_at: string;
  resolved_at: string | null;      
  severity: Severity;
}

export interface RCADetail {
  incident_id: string;
  alarm_title: string;
  severity: Severity;
  status: string;     // 예: "Firing"
  fired_at: string;
  resolved_at: string | null;
  analysis_summary: string;
  analysis_detail: string;
  similar_incidents?: SimilarIncident[]; // 필요 시 구체화
}

// 유사 인시던트 객체 타입 정의
export interface SimilarIncident {
  incident_id: string;
  alarm_title: string;
  score?: number; // 유사도
}