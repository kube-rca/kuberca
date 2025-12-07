import { AlertItem } from '../types';

/**
 * 시간 범위 문자열을 밀리초로 변환
 */
const getTimeRangeInMs = (timeRange: string): number => {
  const now = Date.now();
  
  if (timeRange === 'Last 1 hours') {
    return now - 1 * 60 * 60 * 1000;
  } else if (timeRange === 'Last 6 hours') {
    return now - 6 * 60 * 60 * 1000;
  } else if (timeRange === 'Last 24 hours') {
    return now - 24 * 60 * 60 * 1000;
  } else if (timeRange === 'Last 7 days') {
    return now - 7 * 24 * 60 * 60 * 1000;
  } else if (timeRange === 'Last 30 days') {
    return now - 30 * 24 * 60 * 60 * 1000;
  }
  
  // 기본값: 1시간
  return now - 1 * 60 * 60 * 1000;
};

/**
 * 알림 시간 문자열을 Date 객체로 변환
 * 형식: "2025/12/01 15:00"
 */
const parseAlertTime = (timeString: string): Date => {
  const [datePart, timePart] = timeString.split(' ');
  const [year, month, day] = datePart.split('/').map(Number);
  const [hours, minutes] = timePart.split(':').map(Number);
  
  return new Date(year, month - 1, day, hours, minutes);
};

/**
 * 시간 범위에 따라 알림 목록을 필터링
 */
export const filterAlertsByTimeRange = (
  alerts: AlertItem[],
  timeRange: string
): AlertItem[] => {
  const cutoffTime = getTimeRangeInMs(timeRange);
  
  return alerts.filter((alert) => {
    const alertTime = parseAlertTime(alert.time);
    return alertTime.getTime() >= cutoffTime;
  });
};

