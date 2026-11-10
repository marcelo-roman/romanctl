export class BffError extends Error {
  override readonly name = 'BffError';
}

const UNEXPECTED = 'Algo deu errado. Tente de novo.';
const OFFLINE = 'Não foi possível falar com o servidor. Confira sua conexão.';

interface RequestOptions {
  method?: string;
  body?: unknown;
}

interface ErrorBody {
  message?: string;
}

async function send(path: string, options: RequestOptions): Promise<Response> {
  const init: RequestInit = { method: options.method ?? 'GET' };
  if (options.body !== undefined) {
    init.body = JSON.stringify(options.body);
    init.headers = { 'content-type': 'application/json' };
  }
  try {
    return await fetch(path, init);
  } catch {
    throw new BffError(OFFLINE);
  }
}

export async function bffRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await send(path, options);
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorBody;
    throw new BffError(body.message ?? UNEXPECTED);
  }
  return (await response.json()) as T;
}

export function errorMessage(error: unknown): string {
  return error instanceof BffError ? error.message : UNEXPECTED;
}
