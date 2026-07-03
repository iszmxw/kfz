import { Message } from '@arco-design/web-react';

const API_PREFIX = '/admin/api/v1';
const TOKEN_KEY = 'kfz_admin_token';

export interface ApiResponse<T> {
  code: number;
  data?: T;
  msg: string;
  reqId?: string;
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || '';
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  const token = getToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(`${API_PREFIX}${path}`, { ...options, headers });
  const payload = (await res.json()) as ApiResponse<T>;
  if (payload.code !== 1) {
    Message.error(payload.msg || '请求失败');
    throw new Error(payload.msg || '请求失败');
  }
  return payload.data as T;
}

export function get<T>(path: string) {
  return request<T>(path);
}

export function post<T>(path: string, body?: unknown) {
  return request<T>(path, {
    method: 'POST',
    body: body instanceof FormData ? body : JSON.stringify(body ?? {})
  });
}
