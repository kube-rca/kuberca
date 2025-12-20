import { RCAItem } from '../types';

// 개발 환경: 프록시 사용 (빈 문자열) 또는 localhost
// 프로덕션 환경: 쿠버네티스 서비스 이름 사용
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 
  (import.meta.env.DEV ? 'http://localhost:8080' : 'http://kube-rca-backend:8080');

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
 * 백엔드에서 RCA 목록을 가져옵니다
 */
export const fetchRCAs = async (): Promise<RCAItem[]> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/rca`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data: ApiResponse<RCAResponse> | RCAResponse | RCAItem[] = await response.json();

    // 응답 형식에 따라 데이터 추출
    if (Array.isArray(data)) {
      return data;
    } else if ('data' in data && 'rcas' in data.data) {
      return data.data.rcas;
    } else if ('rcas' in data) {
      return data.rcas;
    } else {
      throw new Error('Unexpected response format');
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

