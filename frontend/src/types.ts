export type Severity = 'info' | 'warning' | 'critical';

export interface AlertItem {
  id: number;
  time: string;      // e.g. "2025/12/01 15:00"
  title: string;     // e.g. "HPA Replicas ~~~ (클릭-상세)"
  severity: Severity;
}

