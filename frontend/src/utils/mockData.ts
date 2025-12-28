import { RCAItem } from '../types';

export const generateMockAlerts = (count: number = 100): RCAItem[] => {
  const severities: RCAItem['severity'][] = ['info', 'warning', 'critical'];
  const titles = [
    'HPA Replicas ~~~',
    'node CPU ~~~',
    'OOM Killed ~~~ ',
    'Pod Restart ~~~',
    'Memory Usage High ~~~',
    'Disk Space Low ~~~',
    'Network Latency ~~~',
    'Service Down ~~~',
    'Certificate Expiring ~~~',
    'High Error Rate ~~~',
  ];

  const alerts: RCAItem[] = [];
  const currentDate = new Date();
  
  for (let i = 0; i < count; i++) {
    const date = new Date(currentDate);
    date.setDate(date.getDate() - Math.floor(i / 3));
    date.setHours(15 - (i % 24), 0, 0, 0);
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    
    alerts.push({
      incident_id: `INC-${i + 1}`,
      resolved_at: `${year}/${month}/${day} ${hours}:${minutes}`,
      alarm_title: titles[i % titles.length],
      severity: severities[i % severities.length],
    });
  }
  
  return alerts;
};
