import { formatMoney, itemCount, type Order, type OrderItem } from '@/entities/order';
import { formatDateTime } from '@/shared/lib';

export interface SummaryLine {
  key: string;
  productName: string;
  quantity: number;
  unitPrice: string;
  subtotal: string;
}

export interface Summary {
  customerName: string;
  createdAt: string;
  units: string;
  total: string;
  lines: SummaryLine[];
}

function lineOf(item: OrderItem, index: number): SummaryLine {
  return {
    key: `${index}-${item.productName}`,
    productName: item.productName,
    quantity: item.quantity,
    unitPrice: formatMoney(item.unitPrice),
    subtotal: formatMoney(item.subtotal),
  };
}

function unitsLabel(units: number): string {
  return units === 1 ? '1 unidade' : `${units} unidades`;
}

export function summaryOf(order: Order): Summary {
  return {
    customerName: order.customerName,
    createdAt: formatDateTime(order.createdAt),
    units: unitsLabel(itemCount(order)),
    total: formatMoney(order.total),
    lines: order.items.map(lineOf),
  };
}
