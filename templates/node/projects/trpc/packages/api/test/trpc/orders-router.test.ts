import { memoryRepositories } from '@sample-app/db';
import { TRPCError } from '@trpc/server';
import { describe, expect, it } from 'vitest';
import { createContext } from '../../src/context';
import { createCaller } from '../../src/trpc/router';

function caller() {
  return createCaller(createContext(memoryRepositories()));
}

async function failure(promise: Promise<unknown>): Promise<TRPCError> {
  const error = await promise.then(
    () => undefined,
    (reason: unknown) => reason,
  );
  if (!(error instanceof TRPCError)) {
    throw new Error(`expected a TRPCError, got ${String(error)}`);
  }
  return error;
}

describe('orders router', () => {
  it('runs the whole flow through the procedures', async () => {
    const api = caller();
    const order = await api.orders.create({ customerName: 'Ana' });
    await api.orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 2, unitPrice: 3.5 });
    const placed = await api.orders.place({ orderId: order.id });

    expect(placed).toMatchObject({ status: 'Placed', total: 7, currency: 'USD' });
    expect(await api.orders.byId({ orderId: order.id })).toEqual(placed);
    expect((await api.orders.list()).map((o) => o.id)).toEqual([order.id]);
  });

  it('answers BAD_REQUEST with readable messages for invalid input', async () => {
    const api = caller();
    const blank = await failure(api.orders.create({ customerName: '   ' }));
    expect(blank.code).toBe('BAD_REQUEST');

    const order = await api.orders.create({ customerName: 'Ana' });
    const zero = await failure(api.orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 0, unitPrice: 1 }));
    expect(zero.code).toBe('BAD_REQUEST');

    const badId = await failure(api.orders.byId({ orderId: '42' }));
    expect(badId.code).toBe('BAD_REQUEST');
  });

  it('answers UNPROCESSABLE_CONTENT for broken rules', async () => {
    const api = caller();
    const order = await api.orders.create({ customerName: 'Ana' });
    const empty = await failure(api.orders.place({ orderId: order.id }));
    expect(empty.code).toBe('UNPROCESSABLE_CONTENT');
    expect(empty.message).toBe('An order without items cannot be placed.');

    await api.orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 1, unitPrice: 1 });
    const currency = await failure(api.orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 1, unitPrice: 1, currency: 'EUR' }));
    expect(currency.code).toBe('UNPROCESSABLE_CONTENT');
  });

  it('answers PRECONDITION_FAILED once the order is placed', async () => {
    const api = caller();
    const order = await api.orders.create({ customerName: 'Ana' });
    await api.orders.addItem({ orderId: order.id, productName: 'Pen', quantity: 1, unitPrice: 1 });
    await api.orders.place({ orderId: order.id });

    const again = await failure(api.orders.place({ orderId: order.id }));
    expect(again.code).toBe('PRECONDITION_FAILED');
    const change = await failure(api.orders.addItem({ orderId: order.id, productName: 'Pencil', quantity: 1, unitPrice: 1 }));
    expect(change.message).toBe('An order that has been placed cannot be changed.');
  });

  it('answers NOT_FOUND for unknown orders', async () => {
    const missing = await failure(caller().orders.byId({ orderId: crypto.randomUUID() }));
    expect(missing.code).toBe('NOT_FOUND');
    expect(missing.message).toBe('Order not found.');
  });
});
