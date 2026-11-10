import type { OrderItemRecord, OrderRecord } from '../orders/order-record';
import type { OrderRepository } from '../orders/order-repository';
import type { PostgresDatabase } from './postgres-database';

interface OrderRow {
  id: string;
  customer_name: string;
  status: string;
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

const selectOrders = 'SELECT id, customer_name, status, created_at FROM orders';

export class PostgresOrderRepository implements OrderRepository {
  constructor(private readonly database: PostgresDatabase) {}

  save(order: OrderRecord): Promise<void> {
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
          [item.id, order.id, position, item.productName, item.quantity, item.unitPriceCents, item.currency],
        );
      }
    });
  }

  async findById(id: string): Promise<OrderRecord | undefined> {
    const [order] = await this.load(`${selectOrders} WHERE id = $1`, [id]);
    return order;
  }

  latest(limit: number): Promise<OrderRecord[]> {
    return this.load(`${selectOrders} ORDER BY created_at DESC LIMIT $1`, [limit]);
  }

  private async load(sql: string, params: unknown[]): Promise<OrderRecord[]> {
    const heads = await this.database.query<OrderRow>(sql, params);
    const items = await this.items(heads.map((head) => head.id));
    return heads.map((head) => ({
      id: head.id,
      customerName: head.customer_name,
      status: head.status,
      createdAt: head.created_at,
      items: items.get(head.id) ?? [],
    }));
  }

  private async items(orderIds: string[]): Promise<Map<string, OrderItemRecord[]>> {
    const byOrder = new Map<string, OrderItemRecord[]>();
    if (orderIds.length === 0) {
      return byOrder;
    }
    const rows = await this.database.query<OrderItemRow>(
      `SELECT order_id, id, product_name, quantity, unit_price, currency
         FROM order_items WHERE order_id = ANY($1::uuid[]) ORDER BY position`,
      [orderIds],
    );
    for (const row of rows) {
      const item: OrderItemRecord = {
        id: row.id,
        productName: row.product_name,
        quantity: row.quantity,
        unitPriceCents: Number(row.unit_price),
        currency: row.currency,
      };
      byOrder.set(row.order_id, [...(byOrder.get(row.order_id) ?? []), item]);
    }
    return byOrder;
  }
}
