import { Money } from '../../domain/money.js';
import { Order } from '../../domain/order.js';
import { OrderItem } from '../../domain/order-item.js';
import type { OrderStatus } from '../../domain/order-status.js';
import type { OrderRepository } from '../order-repository.js';
import type { PostgresDatabase } from './postgres-database.js';

interface OrderRow {
  id: string;
  customer_name: string;
  status: OrderStatus;
  created_at: Date;
}

interface OrderItemRow {
  order_id: string;
  id: string;
  product_name: string;
  quantity: number;
  unit_price: string;
  currency: string;
}

const orderColumns = 'id, customer_name, status, created_at';

export class PostgresOrderRepository implements OrderRepository {
  constructor(private readonly database: PostgresDatabase) {}

  save(order: Order): Promise<void> {
    return this.database.transaction(async (client) => {
      await client.query(
        `INSERT INTO orders (id, customer_name, status, created_at) VALUES ($1, $2, $3, $4)
         ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status`,
        [order.id, order.customerName, order.status, order.createdAt],
      );
      await client.query('DELETE FROM order_items WHERE order_id = $1', [order.id]);
      for (const [position, item] of order.items.entries()) {
        await client.query(
          `INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
           VALUES ($1, $2, $3, $4, $5, $6, $7)`,
          [item.id, order.id, position, item.productName, item.quantity, item.unitPrice.cents, item.unitPrice.currency],
        );
      }
    });
  }

  async findById(id: string): Promise<Order | undefined> {
    const [order] = await this.load(`SELECT ${orderColumns} FROM orders WHERE id = $1`, [id]);
    return order;
  }

  latest(limit: number): Promise<Order[]> {
    return this.load(`SELECT ${orderColumns} FROM orders ORDER BY created_at DESC, id LIMIT $1`, [limit]);
  }

  private async load(sql: string, params: unknown[]): Promise<Order[]> {
    const heads = await this.database.query<OrderRow>(sql, params);
    const items = await this.items(heads.map((head) => head.id));
    return heads.map((head) => Order.restore(head.id, head.customer_name, head.status, head.created_at, items.get(head.id) ?? []));
  }

  private async items(orderIds: string[]): Promise<Map<string, OrderItem[]>> {
    const byOrder = new Map<string, OrderItem[]>();
    if (orderIds.length === 0) {
      return byOrder;
    }
    const rows = await this.database.query<OrderItemRow>(
      `SELECT order_id, id, product_name, quantity, unit_price, currency
         FROM order_items WHERE order_id = ANY($1::uuid[]) ORDER BY position`,
      [orderIds],
    );
    for (const row of rows) {
      const item = OrderItem.restore(row.id, row.product_name, row.quantity, Money.fromCents(Number(row.unit_price), row.currency));
      byOrder.set(row.order_id, [...(byOrder.get(row.order_id) ?? []), item]);
    }
    return byOrder;
  }
}
