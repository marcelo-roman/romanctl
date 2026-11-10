import type { OrderRecord } from '@sample-app/db';
import { Money } from '../domain/money';
import { Order } from '../domain/order';
import { OrderItem } from '../domain/order-item';
import { parseOrderStatus } from '../domain/order-status';

export function toOrderRecord(order: Order): OrderRecord {
  return {
    id: order.id,
    customerName: order.customerName,
    status: order.status,
    createdAt: order.createdAt,
    items: order.items.map((item) => ({
      id: item.id,
      productName: item.productName,
      quantity: item.quantity,
      unitPriceCents: item.unitPrice.cents,
      currency: item.unitPrice.currency,
    })),
  };
}

export function fromOrderRecord(record: OrderRecord): Order {
  const items = record.items.map((item) =>
    OrderItem.restore(item.id, item.productName, item.quantity, Money.fromCents(item.unitPriceCents, item.currency)),
  );
  return Order.restore(record.id, record.customerName, parseOrderStatus(record.status), record.createdAt, items);
}
