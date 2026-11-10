import type { Order } from '../domain/order.js';

export interface OrderRepository {
  save(order: Order): Promise<void>;
  findById(id: string): Promise<Order | undefined>;
  latest(limit: number): Promise<Order[]>;
}
