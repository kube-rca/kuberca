export type Severity = 'info' | 'warning' | 'critical';

export interface RCAItem {
  incident_id: string;  
  alarm_title: string; 
  resolved_at: string;      // e.g. "2025/12/01 15:00"
  severity: Severity;
}

