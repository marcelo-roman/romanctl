import { Test } from '@nestjs/testing';
import { CommandBus, CqrsModule, EventBus, QueryBus } from '@nestjs/cqrs';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { Clock } from '../../src/application/common/clock.js';
import { NotFoundError } from '../../src/application/common/not-found-error.js';
import { AddOrderItem } from '../../src/application/orders/commands/add-order-item.js';
import { CreateOrder } from '../../src/application/orders/commands/create-order.js';
import { PlaceOrder } from '../../src/application/orders/commands/place-order.js';
import { orderHandlers } from '../../src/application/orders/order-handlers.js';
import { OrderReadStore } from '../../src/application/orders/order-read-store.js';
import { type OrderView, toOrderView } from '../../src/application/orders/order-view.js';
import { GetOrderById } from '../../src/application/orders/queries/get-order-by-id.js';
import { ListOrders } from '../../src/application/orders/queries/list-orders.js';
import type { Order } from '../../src/domain/orders/order.js';
import type { OrderId } from '../../src/domain/orders/order-id.js';
import { OrderPlaced } from '../../src/domain/orders/order-placed.js';
import { OrderRepository } from '../../src/domain/orders/order-repository.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';

class FakeOrders extends OrderRepository {
  readonly saved = new Map<string, Order>();

  async save(order: Order): Promise<void> {
    this.saved.set(order.id.value, order);
  }

  async findById(id: OrderId): Promise<Order | undefined> {
    return this.saved.get(id.value);
  }
}

class FakeReadStore extends OrderReadStore {
  readonly limits: number[] = [];

  constructor(private readonly orders: FakeOrders) {
    super();
  }

  async findById(id: string): Promise<OrderView | undefined> {
    const order = this.orders.saved.get(id);
    return order && toOrderView(order);
  }

  async list(limit: number): Promise<OrderView[]> {
    this.limits.push(limit);
    return [...this.orders.saved.values()].map(toOrderView);
  }
}

class FixedClock extends Clock {
  now(): Date {
    return new Date('2026-01-01T00:00:00Z');
  }
}

describe('order use cases', () => {
  let commands: CommandBus;
  let queries: QueryBus;
  let readStore: FakeReadStore;
  let published: unknown[];
  let close: () => Promise<void>;

  beforeEach(async () => {
    const orders = new FakeOrders();
    readStore = new FakeReadStore(orders);
    const module = await Test.createTestingModule({
      imports: [CqrsModule.forRoot()],
      providers: [
        ...orderHandlers,
        { provide: OrderRepository, useValue: orders },
        { provide: OrderReadStore, useValue: readStore },
        { provide: Clock, useClass: FixedClock },
      ],
    }).compile();
    const app = await module.init();
    commands = app.get(CommandBus);
    queries = app.get(QueryBus);
    published = [];
    app.get(EventBus).subscribe((event) => published.push(event));
    close = () => app.close();
  });

  afterEach(() => close());

  it('creates, fills and places an order, publishing OrderPlaced', async () => {
    const id = await commands.execute(new CreateOrder('Ana'));
    await commands.execute(new AddOrderItem(id, 'Caneta', 3, 2.5));
    await commands.execute(new PlaceOrder(id));

    const view = await queries.execute(new GetOrderById(id));
    expect(view).toMatchObject({ status: 'Placed', total: 7.5, currency: 'BRL', createdAt: '2026-01-01T00:00:00.000Z' });
    expect(published).toEqual([new OrderPlaced(id, 7.5, 'BRL', new Date('2026-01-01T00:00:00Z'))]);
  });

  it('reports unknown orders as not found', async () => {
    const missing = '00000000-0000-4000-8000-000000000000';
    await expect(commands.execute(new PlaceOrder(missing))).rejects.toThrow(NotFoundError);
    await expect(queries.execute(new GetOrderById(missing))).rejects.toThrow(NotFoundError);
  });

  it('lets business rules surface as domain errors', async () => {
    await expect(commands.execute(new CreateOrder(''))).rejects.toThrow(DomainError);
    const id = await commands.execute(new CreateOrder('Ana'));
    await expect(commands.execute(new PlaceOrder(id))).rejects.toThrow(DomainError);
    await expect(commands.execute(new AddOrderItem(id, 'Caneta', 1, -1))).rejects.toThrow(DomainError);
    expect(published).toEqual([]);
  });

  it('bounds the list size', async () => {
    await queries.execute(new ListOrders());
    await queries.execute(new ListOrders(1000));
    await queries.execute(new ListOrders(0));
    expect(readStore.limits).toEqual([50, 200, 1]);
  });
});
