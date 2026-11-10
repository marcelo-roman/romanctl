import type { OrderRepository } from '@sample-app/db';
import { Money } from '../domain/money';
import { Order } from '../domain/order';
import type { Clock } from './clock';
import { listLimit } from './list-limit';
import { NotFoundError } from './not-found-error';
import { fromOrderRecord, toOrderRecord } from './order-records';
import { type OrderView, toOrderView } from './order-view';

export interface NewOrderItem {
  orderId: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  currency?: string;
}

export class OrderService {
  constructor(
    private readonly orders: OrderRepository,
    private readonly clock: Clock,
  ) {}

  async create(customerName: string): Promise<OrderView> {
    const order = Order.create(customerName, this.clock.now());
    await this.orders.save(toOrderRecord(order));
    return toOrderView(order);
  }

  async addItem(item: NewOrderItem): Promise<OrderView> {
    const order = await this.load(item.orderId);
    order.addItem(item.productName, item.quantity, Money.fromAmount(item.unitPrice, item.currency));
    await this.orders.save(toOrderRecord(order));
    return toOrderView(order);
  }

  async place(orderId: string): Promise<OrderView> {
    const order = await this.load(orderId);
    order.place();
    await this.orders.save(toOrderRecord(order));
    return toOrderView(order);
  }

  async byId(orderId: string): Promise<OrderView> {
    return toOrderView(await this.load(orderId));
  }

  async list(limit?: number): Promise<OrderView[]> {
    const records = await this.orders.latest(listLimit(limit));
    return records.map((record) => toOrderView(fromOrderRecord(record)));
  }

  private async load(orderId: string): Promise<Order> {
    const record = await this.orders.findById(orderId);
    if (!record) {
      throw new NotFoundError('Pedido não encontrado.');
    }
    return fromOrderRecord(record);
  }
}
