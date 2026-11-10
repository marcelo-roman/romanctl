import { apiBaseUrl } from '@/shared/config';
import { ApiError, defaultTitle, unreachableStatus } from './api-error';
import { errorFromResponse } from './problem';

const noContent = 204;

async function send(path: string, init: RequestInit): Promise<Response> {
  try {
    return await fetch(`${apiBaseUrl()}${path}`, {
      ...init,
      headers: { Accept: 'application/json', ...init.headers },
    });
  } catch {
    throw new ApiError({ status: unreachableStatus, title: defaultTitle(unreachableStatus) });
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await send(path, init);
  if (!response.ok) {
    throw await errorFromResponse(response);
  }
  if (response.status === noContent) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export function getJson<T>(path: string): Promise<T> {
  return request<T>(path);
}

export function postJson<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}
