import { Inject, Injectable } from '@nestjs/common';
import type { Order } from '../../../domain/orders/order.js';
import type { OrderId } from '../../../domain/orders/order-id.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { InMemoryOrderStore } from './in-memory-order-store.js';

@Injectable()
export class InMemoryOrderRepository extends OrderRepository {
  constructor(@Inject(InMemoryOrderStore) private readonly store: InMemoryOrderStore) {
    super();
  }

  async save(order: Order): Promise<void> {
    this.store.put(order);
  }

  async findById(id: OrderId): Promise<Order | undefined> {
    return this.store.get(id.value);
  }
}
