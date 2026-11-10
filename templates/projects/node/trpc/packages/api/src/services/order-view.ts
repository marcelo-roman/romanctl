import type { Order } from '../domain/order';

export interface OrderItemView {
  id: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export interface OrderView {
  id: string;
  customerName: string;
  status: string;
  total: number;
  currency: string;
  createdAt: string;
  items: OrderItemView[];
}

export function toOrderView(order: Order): OrderView {
  const total = order.total;
  return {
    id: order.id,
    customerName: order.customerName,
    status: order.status,
    total: total.amount,
    currency: total.currency,
    createdAt: order.createdAt.toISOString(),
    items: order.items.map((item) => ({
      id: item.id,
      productName: item.productName,
      quantity: item.quantity,
      unitPrice: item.unitPrice.amount,
      subtotal: item.subtotal.amount,
    })),
  };
}
