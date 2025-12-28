import { RCAItem, RCADetail } from '../types';
import { API_BASE_URL } from './config';
import { getAccessToken, refreshAccessToken } from './auth';

// 인증 토큰을 포함해 요청하고 401이면 토큰 갱신 후 1회 재시도한다.
const requestWithAuth = async (path: string, init: RequestInit = {}) => {
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
