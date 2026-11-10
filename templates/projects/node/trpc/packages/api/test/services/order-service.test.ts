import { memoryRepositories } from '@sample-app/db';
import { describe, expect, it } from 'vitest';
import { DomainError, InvalidStateError } from '../../src/domain/domain-error';
import { listLimit } from '../../src/services/list-limit';
import { NotFoundError } from '../../src/services/not-found-error';
import { OrderService } from '../../src/services/order-service';

function service(start = Date.parse('2026-01-01T12:00:00Z')): OrderService {
  let tick = start;
  return new OrderService(memoryRepositories().orders, { now: () => new Date((tick += 1000)) });
}

describe('OrderService', () => {
  it('creates, fills and places an order, saving every step', async () => {
    const orders = service();
    const created = await orders.create(' Ana ');
    await orders.addItem({ orderId: created.id, productName: 'Caneta', quantity: 2, unitPrice: 3.5 });
    await orders.addItem({ orderId: created.id, productName: 'Caderno', quantity: 1, unitPrice: 12, currency: 'brl' });
    const placed = await orders.place(created.id);

    expect(placed).toMatchObject({ customerName: 'Ana', status: 'Placed', total: 19, currency: 'BRL' });
    expect(await orders.byId(created.id)).toEqual(placed);
  });

  it('lists the most recent orders first', async () => {
    const orders = service();
    const first = await orders.create('Ana');
    const second = await orders.create('Bia');
    expect((await orders.list()).map((o) => o.id)).toEqual([second.id, first.id]);
    expect((await orders.list(1)).map((o) => o.id)).toEqual([second.id]);
  });

  it('reports unknown orders and broken rules', async () => {
    const orders = service();
    await expect(orders.byId(crypto.randomUUID())).rejects.toThrow(NotFoundError);
    const order = await orders.create('Ana');
    await expect(orders.place(order.id)).rejects.toThrow(DomainError);
    await orders.addItem({ orderId: order.id, productName: 'Caneta', quantity: 1, unitPrice: 1 });
    await expect(orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 1, unitPrice: 1, currency: 'USD' })).rejects.toThrow(
      'mesma moeda',
    );
    await orders.place(order.id);
    await expect(orders.place(order.id)).rejects.toThrow(InvalidStateError);
  });
});

describe('listLimit', () => {
  it('defaults to 50 and stays between 1 and 200', () => {
    expect(listLimit(undefined)).toBe(50);
    expect(listLimit(0)).toBe(1);
    expect(listLimit(1000)).toBe(200);
    expect(listLimit(2.5)).toBe(50);
  });
});
