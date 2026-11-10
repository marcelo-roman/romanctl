import type { ReactNode } from 'react';
import { NotFoundPage } from './pages/NotFoundPage';
import { OrderPage } from './pages/OrderPage';
import { OrdersPage } from './pages/OrdersPage';
// romanctl:imports

interface Route {
  pattern: RegExp;
  render: (params: string[]) => ReactNode;
}

const routes: Route[] = [
  { pattern: /^\/(orders\/?)?$/, render: () => <OrdersPage /> },
  { pattern: /^\/orders\/([^/]+)$/, render: ([id = '']) => <OrderPage key={id} id={decodeURIComponent(id)} /> },
  // romanctl:routes
];

export function renderRoute(path: string): ReactNode {
  for (const route of routes) {
    const match = route.pattern.exec(path);
    if (match) {
      return route.render(match.slice(1));
    }
  }
  return <NotFoundPage />;
}
