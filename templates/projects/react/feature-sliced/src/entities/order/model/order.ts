import type { Money } from './money';

export type OrderStatus = 'Draft' | 'Placed';

export interface OrderItem {
  productName: string;
  quantity: number;
  unitPrice: Money;
  subtotal: Money;
}

export interface Order {
  id: string;
  customerName: string;
  status: OrderStatus;
  createdAt: Date;
  total: Money;
  items: OrderItem[];
}

const statusLabels: Record<OrderStatus, string> = {
  Draft: 'Rascunho',
  Placed: 'Realizado',
};

export function statusLabel(status: OrderStatus): string {
  return statusLabels[status];
}

export function isDraft(order: Order): boolean {
  return order.status === 'Draft';
}

export function itemCount(order: Order): number {
  return order.items.reduce((count, item) => count + item.quantity, 0);
}
