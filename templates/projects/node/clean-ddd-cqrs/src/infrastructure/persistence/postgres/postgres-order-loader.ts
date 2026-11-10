import { Inject, Injectable } from '@nestjs/common';
import { Money } from '../../../domain/orders/money.js';
import { Order } from '../../../domain/orders/order.js';
import { OrderId } from '../../../domain/orders/order-id.js';
import { OrderItem } from '../../../domain/orders/order-item.js';
import type { OrderStatus } from '../../../domain/orders/order-status.js';
import { PostgresDatabase } from './postgres-database.js';

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

@Injectable()
export class PostgresOrderLoader {
  constructor(@Inject(PostgresDatabase) private readonly db: PostgresDatabase) {}

  async byId(id: string): Promise<Order | undefined> {
    const [order] = await this.load('SELECT id, customer_name, status, created_at FROM orders WHERE id = $1', [id]);
    return order;
  }

  latest(limit: number): Promise<Order[]> {
    return this.load('SELECT id, customer_name, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1', [limit]);
  }

  private async load(sql: string, params: unknown[]): Promise<Order[]> {
    const heads = await this.db.query<OrderRow>(sql, params);
    const items = await this.items(heads.map((head) => head.id));
    return heads.map((head) =>
      Order.restore(OrderId.parse(head.id), head.customer_name, head.status, head.created_at, items.get(head.id) ?? []),
    );
  }

  private async items(orderIds: string[]): Promise<Map<string, OrderItem[]>> {
    const byOrder = new Map<string, OrderItem[]>();
    if (orderIds.length === 0) {
      return byOrder;
    }
    const rows = await this.db.query<OrderItemRow>(
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
