import { ApiError, defaultTitle, type Problem } from './api-error';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function text(value: unknown): string | undefined {
  if (typeof value !== 'string' || value.trim() === '') {
    return undefined;
  }
  return value;
}

export function problemOf(status: number, body: unknown): Problem {
  if (!isRecord(body)) {
    return { status, title: defaultTitle(status) };
  }
  return {
    status,
    title: text(body.title) ?? defaultTitle(status),
    detail: text(body.detail),
  };
}

async function readBody(response: Response): Promise<unknown> {
  const raw = await response.text();
  if (raw === '') {
    return undefined;
  }
  try {
    return JSON.parse(raw);
  } catch {
    return undefined;
  }
}

export async function errorFromResponse(response: Response): Promise<ApiError> {
  return new ApiError(problemOf(response.status, await readBody(response)));
}
