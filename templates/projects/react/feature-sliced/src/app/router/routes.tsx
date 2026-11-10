import type { RouteObject } from 'react-router';
import { NotFoundPage } from '@/pages/not-found';
import { OrderDetailsPage } from '@/pages/order-details';
import { OrdersListPage } from '@/pages/orders-list';
// romanctl:imports
import { paths } from '@/shared/config';
import { AppLayout } from '../layout/app-layout';

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { path: paths.orders, element: <OrdersListPage /> },
      { path: paths.orderDetails, element: <OrderDetailsPage /> },
      // romanctl:routes
      { path: '*', element: <NotFoundPage /> },
    ],
  },
];
