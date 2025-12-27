import { RCAItem } from '../types';

const API_BASE_URL = '';

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface RCAResponse {
  rcas: RCAItem[];
  total?: number;
  page?: number;
  pageSize?: number;
}

/**
 * 백엔드에서 RCA(Incident) 목록을 가져옵니다
 */
export const fetchRCAs = async (): Promise<RCAItem[]> => {
  // 1. Basic Auth 정보 설정
  const username = import.meta.env.VITE_API_USERNAME;
  const password = import.meta.env.VITE_API_PASSWORD;
  
  // 2. 'username:password' 문자열을 Base64로 인코딩 (브라우저 내장 함수 btoa 사용)
  const encodedCredentials = btoa(`${username}:${password}`);

  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/incidents`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Basic ${encodedCredentials}`,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // [수정] 백엔드가 객체가 아닌 배열([])을 바로 반환하므로 바로 파싱합니다.
    const data: RCAItem[] = await response.json();

    // 데이터가 배열인지 확인 후 반환
    if (Array.isArray(data)) {
      return data;
    } else {
      // 배열이 아닌 경우(예: 에러 메시지 객체가 온 경우) 처리
      console.warn('Backend response is not an array:', data);
      throw new Error('Unexpected response format: Data is not an array');
    }

  } catch (error) {
    console.error('Failed to fetch RCAs:', error);
    throw error;
  }
};

/**
 * 특정 RCA를 가져옵니다
 */
export const fetchRCAById = async (id: number): Promise<RCAItem> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/rca/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data: ApiResponse<RCAItem> | RCAItem = await response.json();

    if ('data' in data) {
      return data.data;
    }
    return data;
  } catch (error) {
    console.error(`Failed to fetch RCA ${id}:`, error);
    throw error;
  }
};

