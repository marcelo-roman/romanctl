import { generatePath } from 'react-router';
import { formatMoney, itemCount, type Order, type OrderStatus } from '@/entities/order';
import { paths } from '@/shared/config';
import { formatDateTime } from '@/shared/lib';

export interface OrderRow {
  id: string;
  href: string;
  customerName: string;
  status: OrderStatus;
  createdAt: string;
  items: number;
  total: string;
}

export function orderRowOf(order: Order): OrderRow {
  return {
    id: order.id,
    href: generatePath(paths.orderDetails, { id: order.id }),
    customerName: order.customerName,
    status: order.status,
    createdAt: formatDateTime(order.createdAt),
    items: itemCount(order),
    total: formatMoney(order.total),
  };
}
