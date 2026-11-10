import type { OrderRecord } from './order-record';

export interface OrderRepository {
  save(order: OrderRecord): Promise<void>;
  findById(id: string): Promise<OrderRecord | undefined>;
  latest(limit: number): Promise<OrderRecord[]>;
}
