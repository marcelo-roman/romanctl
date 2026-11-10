import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactElement, ReactNode } from 'react';
import { createMemoryRouter, RouterProvider, type RouteObject } from 'react-router';

export function testQueryClient(): QueryClient {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}

export function withQueryClient(children: ReactNode, client: QueryClient = testQueryClient()): ReactElement {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

export function renderRoutes(routes: RouteObject[], initialPath = '/') {
  const router = createMemoryRouter(routes, { initialEntries: [initialPath] });
  const user = userEvent.setup();
  const view = render(withQueryClient(<RouterProvider router={router} />));
  return { ...view, router, user };
}

export function renderWithProviders(ui: ReactElement, initialPath = '/') {
  return renderRoutes([{ path: '*', element: ui }], initialPath);
}
