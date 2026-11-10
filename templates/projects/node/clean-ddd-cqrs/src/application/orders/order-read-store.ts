import type { OrderView } from './order-view.js';

export abstract class OrderReadStore {
  abstract findById(id: string): Promise<OrderView | undefined>;
  abstract list(limit: number): Promise<OrderView[]>;
}
