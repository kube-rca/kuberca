import { RCAItem, RCADetail } from '../types';

const API_BASE_URL = '';
const USERNAME = import.meta.env.VITE_API_USERNAME || '';
const PASSWORD = import.meta.env.VITE_API_PASSWORD || '';

const getHeaders = () => {
  // ID:PW를 Base64로 인코딩
  const base64Auth = btoa(`${USERNAME}:${PASSWORD}`);
  
  return {
    'Authorization': `Basic ${base64Auth}`,
    'Content-Type': 'application/json',
  };
};

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

  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/incidents`, {
      method: 'GET',
      headers: getHeaders(),
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

export const fetchRCADetail = async (id: string): Promise<RCADetail> => {
  const response = await fetch(`/api/v1/incidents/${id}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  if (!response.ok) {
    throw new Error('상세 정보를 불러오는데 실패했습니다.');
  }

  const json = await response.json();
  return json.data; // Postman 응답 구조가 { data: { ... } } 이므로
};

export const updateRCADetail = async (id: string, data: Partial<RCADetail>): Promise<void> => {
  const response = await fetch(`/api/v1/incidents/${id}`, {
    method: 'PUT', 
    headers: getHeaders(), 
    body: JSON.stringify(data), // 수정할 데이터를 JSON 문자열로 변환
  });

  if (!response.ok) {
    throw new Error('RCA 정보를 수정하는데 실패했습니다.');
  }
};
