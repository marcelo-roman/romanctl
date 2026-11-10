import { ApiError, ApiUnavailableError } from './api-error.js';

export interface RequestOptions {
  body?: unknown;
  notFound?: string;
}

interface Problem {
  detail?: string;
  title?: string;
}

const TIMEOUT_MS = 5_000;

async function problemDetail(response: Response): Promise<string> {
  const problem = (await response.json().catch(() => ({}))) as Problem;
  return problem.detail ?? problem.title ?? response.statusText;
}

export class ApiHttp {
  constructor(private readonly baseUrl: string) {}

  get<T>(path: string, options: RequestOptions = {}): Promise<T> {
    return this.request<T>('GET', path, options);
  }

  post<T>(path: string, options: RequestOptions = {}): Promise<T> {
    return this.request<T>('POST', path, options);
  }

  put<T>(path: string, options: RequestOptions = {}): Promise<T> {
    return this.request<T>('PUT', path, options);
  }

  delete(path: string, options: RequestOptions = {}): Promise<void> {
    return this.request<void>('DELETE', path, options);
  }

  private async request<T>(method: string, path: string, options: RequestOptions): Promise<T> {
    const response = await this.send(method, path, options.body);
    if (response.status >= 500) {
      throw new ApiUnavailableError(`${method} ${path}: ${response.status}`);
    }
    if (response.status === 404 && options.notFound) {
      throw new ApiError(404, options.notFound);
    }
    if (!response.ok) {
      throw new ApiError(response.status, await problemDetail(response));
    }
    if (response.status === 204) {
      return undefined as T;
    }
    return (await response.json()) as T;
  }

  private async send(method: string, path: string, body: unknown): Promise<Response> {
    const init: RequestInit = { method, signal: AbortSignal.timeout(TIMEOUT_MS) };
    if (body !== undefined) {
      init.body = JSON.stringify(body);
      init.headers = { 'content-type': 'application/json' };
    }
    try {
      return await fetch(new URL(path, this.baseUrl), init);
    } catch (error) {
      throw new ApiUnavailableError(`${method} ${path}: ${String(error)}`);
    }
  }
}
