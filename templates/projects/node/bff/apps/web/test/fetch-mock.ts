import { vi } from 'vitest';

export interface FetchCall {
  method: string;
  url: string;
  body: unknown;
}

export interface FakeResponse {
  status?: number;
  body: unknown;
}

export function mockFetch(answer: (call: FetchCall) => FakeResponse): FetchCall[] {
  const calls: FetchCall[] = [];
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string | URL, init: RequestInit = {}) => {
      const call: FetchCall = {
        method: init.method ?? 'GET',
        url: String(input),
        body: typeof init.body === 'string' ? JSON.parse(init.body) : undefined,
      };
      calls.push(call);
      const { status = 200, body } = answer(call);
      return new Response(JSON.stringify(body), { status, headers: { 'content-type': 'application/json' } });
    }),
  );
  return calls;
}
