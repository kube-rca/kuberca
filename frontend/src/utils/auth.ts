import { API_BASE_URL } from './config';

let accessToken: string | null = null;

export interface AuthResponse {
  accessToken: string;
  expiresIn: number;
}

export interface AuthConfigResponse {
  allowSignup: boolean;
}

export const getAccessToken = () => accessToken;

export const setAccessToken = (token: string | null) => {
  accessToken = token;
};

const parseAuthResponse = async (response: Response): Promise<AuthResponse> => {
  if (!response.ok) {
    throw new Error('auth_failed');
  }
  return response.json() as Promise<AuthResponse>;
};

export const login = async (id: string, password: string): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ id, password }),
  });

  const data = await parseAuthResponse(response);
  setAccessToken(data.accessToken);
  return data;
};

export const register = async (id: string, password: string): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ id, password }),
  });

  const data = await parseAuthResponse(response);
  setAccessToken(data.accessToken);
  return data;
};

export const refreshAccessToken = async (): Promise<boolean> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    setAccessToken(null);
    return false;
  }

  const data = await response.json();
  if (data?.accessToken) {
    setAccessToken(data.accessToken);
    return true;
  }

  setAccessToken(null);
  return false;
};

export const logout = async (): Promise<void> => {
  await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  });
  setAccessToken(null);
};

export const fetchAuthConfig = async (): Promise<AuthConfigResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/config`);
  if (!response.ok) {
    return { allowSignup: false };
  }
  return response.json() as Promise<AuthConfigResponse>;
};
