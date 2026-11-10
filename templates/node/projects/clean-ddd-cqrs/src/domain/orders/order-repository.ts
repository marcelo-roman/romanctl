import type { Order } from './order.js';
import type { OrderId } from './order-id.js';

export abstract class OrderRepository {
  abstract save(order: Order): Promise<void>;
  abstract findById(id: OrderId): Promise<Order | undefined>;
}
