import type { AppRouter } from '@sample-app/api';
import { createTRPCClient, httpBatchLink } from '@trpc/client';

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

export function createApiClient(url = '/api/trpc', fetcher?: Fetcher) {
  return createTRPCClient<AppRouter>({ links: [httpBatchLink({ url, fetch: fetcher })] });
}

export type ApiClient = ReturnType<typeof createApiClient>;
