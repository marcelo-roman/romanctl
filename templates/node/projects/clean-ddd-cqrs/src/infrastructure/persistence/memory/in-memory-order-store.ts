import { Order } from '../../../domain/orders/order.js';

export class InMemoryOrderStore {
  private readonly orders = new Map<string, Order>();

  put(order: Order): void {
    this.orders.set(order.id.value, copy(order));
  }

  get(id: string): Order | undefined {
    const order = this.orders.get(id);
    return order && copy(order);
  }

  latest(limit: number): Order[] {
    return [...this.orders.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map(copy);
  }
}

function copy(order: Order): Order {
  return Order.restore(order.id, order.customerName, order.status, order.createdAt, [...order.items]);
}
