import { Inject, Injectable } from '@nestjs/common';
import { OrderReadStore } from '../../../application/orders/order-read-store.js';
import { type OrderView, toOrderView } from '../../../application/orders/order-view.js';
import { InMemoryOrderStore } from './in-memory-order-store.js';

@Injectable()
export class InMemoryOrderReadStore extends OrderReadStore {
  constructor(@Inject(InMemoryOrderStore) private readonly store: InMemoryOrderStore) {
    super();
  }

  async findById(id: string): Promise<OrderView | undefined> {
    const order = this.store.get(id);
    return order && toOrderView(order);
  }

  async list(limit: number): Promise<OrderView[]> {
    return this.store.latest(limit).map(toOrderView);
  }
}
