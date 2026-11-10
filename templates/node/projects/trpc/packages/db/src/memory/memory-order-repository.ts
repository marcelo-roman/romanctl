import type { OrderRecord } from '../orders/order-record';
import type { OrderRepository } from '../orders/order-repository';

export class MemoryOrderRepository implements OrderRepository {
  private readonly orders = new Map<string, OrderRecord>();

  async save(order: OrderRecord): Promise<void> {
    this.orders.set(order.id, structuredClone(order));
  }

  async findById(id: string): Promise<OrderRecord | undefined> {
    const order = this.orders.get(id);
    return order && structuredClone(order);
  }

  async latest(limit: number): Promise<OrderRecord[]> {
    return [...this.orders.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map((order) => structuredClone(order));
  }
}
