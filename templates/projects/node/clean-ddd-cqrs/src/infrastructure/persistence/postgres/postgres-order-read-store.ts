import { Inject, Injectable } from '@nestjs/common';
import { OrderReadStore } from '../../../application/orders/order-read-store.js';
import { type OrderView, toOrderView } from '../../../application/orders/order-view.js';
import { PostgresOrderLoader } from './postgres-order-loader.js';

@Injectable()
export class PostgresOrderReadStore extends OrderReadStore {
  constructor(@Inject(PostgresOrderLoader) private readonly loader: PostgresOrderLoader) {
    super();
  }

  async findById(id: string): Promise<OrderView | undefined> {
    const order = await this.loader.byId(id);
    return order && toOrderView(order);
  }

  async list(limit: number): Promise<OrderView[]> {
    const orders = await this.loader.latest(limit);
    return orders.map(toOrderView);
  }
}
