import { Order } from '../../domain/order.js';
import { OrderItem } from '../../domain/order-item.js';
import type { OrderRepository } from '../order-repository.js';

function copy(order: Order): Order {
  const items = order.items.map((item) => OrderItem.restore(item.id, item.productName, item.quantity, item.unitPrice));
  return Order.restore(order.id, order.customerName, order.status, order.createdAt, items);
}

export class MemoryOrderRepository implements OrderRepository {
  private readonly orders = new Map<string, Order>();

  async save(order: Order): Promise<void> {
    this.orders.set(order.id, copy(order));
  }

  async findById(id: string): Promise<Order | undefined> {
    const order = this.orders.get(id);
    return order ? copy(order) : undefined;
  }

  async latest(limit: number): Promise<Order[]> {
    return [...this.orders.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map(copy);
  }
}
