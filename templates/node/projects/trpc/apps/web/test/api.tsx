import { appRouter, createContext } from '@sample-app/api';
import { memoryRepositories } from '@sample-app/db';
import { render } from '@testing-library/react';
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
import type { ReactNode } from 'react';
import { ApiProvider } from '../src/trpc/api-provider';
import { type ApiClient, createApiClient } from '../src/trpc/client';

export function testApi(): ApiClient {
  const repositories = memoryRepositories();
  return createApiClient('http://localhost/api/trpc', (input, init) =>
    fetchRequestHandler({
      endpoint: '/api/trpc',
      req: new Request(input, { method: init?.method, headers: init?.headers, body: init?.body }),
      router: appRouter,
      createContext: () => createContext(repositories),
    }),
  );
}

export function renderWithApi(ui: ReactNode, client: ApiClient) {
  return render(<ApiProvider client={client}>{ui}</ApiProvider>);
}
