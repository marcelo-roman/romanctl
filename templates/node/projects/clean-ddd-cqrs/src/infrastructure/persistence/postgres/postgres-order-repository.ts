import { Inject, Injectable } from '@nestjs/common';
import type { Order } from '../../../domain/orders/order.js';
import type { OrderId } from '../../../domain/orders/order-id.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { PostgresDatabase } from './postgres-database.js';
import { PostgresOrderLoader } from './postgres-order-loader.js';

@Injectable()
export class PostgresOrderRepository extends OrderRepository {
  constructor(
    @Inject(PostgresDatabase) private readonly db: PostgresDatabase,
    @Inject(PostgresOrderLoader) private readonly loader: PostgresOrderLoader,
  ) {
    super();
  }

  save(order: Order): Promise<void> {
    return this.db.transaction(async (client) => {
      await client.query(
        `INSERT INTO orders (id, customer_name, status, created_at) VALUES ($1, $2, $3, $4)
         ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status`,
        [order.id.value, order.customerName, order.status, order.createdAt],
      );
      await client.query('DELETE FROM order_items WHERE order_id = $1', [order.id.value]);
      for (const [position, item] of order.items.entries()) {
        await client.query(
          `INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
           VALUES ($1, $2, $3, $4, $5, $6, $7)`,
          [item.id, order.id.value, position, item.productName, item.quantity, item.unitPrice.cents, item.unitPrice.currency],
        );
      }
    });
  }

  findById(id: OrderId): Promise<Order | undefined> {
    return this.loader.byId(id.value);
  }
}
