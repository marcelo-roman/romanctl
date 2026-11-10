import { Money } from '../domain/money.js';
import { Order } from '../domain/order.js';
import type { OrderRepository } from '../repositories/order-repository.js';
import type { Clock } from './clock.js';
import { NotFoundError } from './errors.js';
import { listLimit } from './list-limit.js';

export interface CreateOrderInput {
  customerName: string;
}

export interface AddOrderItemInput {
  productName: string;
  quantity: number;
  unitPrice: number;
  currency?: string;
}

export interface ListOrdersInput {
  limit?: number;
}

export class OrderService {
  constructor(
    private readonly orders: OrderRepository,
    private readonly clock: Clock,
  ) {}

  async create(input: CreateOrderInput): Promise<Order> {
    const order = Order.create(input.customerName, this.clock());
    await this.orders.save(order);
    return order;
  }

  list(input: ListOrdersInput): Promise<Order[]> {
    return this.orders.latest(listLimit(input.limit));
  }

  async get(id: string): Promise<Order> {
    const order = await this.orders.findById(id);
    if (!order) {
      throw new NotFoundError(`Pedido ${id} não encontrado.`);
    }
    return order;
  }

  async addItem(id: string, input: AddOrderItemInput): Promise<Order> {
    const order = await this.get(id);
    order.addItem(input.productName, input.quantity, Money.fromAmount(input.unitPrice, input.currency));
    await this.orders.save(order);
    return order;
  }

  async place(id: string): Promise<Order> {
    const order = await this.get(id);
    order.place();
    await this.orders.save(order);
    return order;
  }
}
