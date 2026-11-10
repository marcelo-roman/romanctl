import { randomUUID } from 'node:crypto';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import type { OrderRecord } from '../../src/orders/order-record';
import type { PostgresDatabase } from '../../src/postgres/postgres-database';
import { PostgresOrderRepository } from '../../src/postgres/postgres-order-repository';
import { databaseUrl, openDatabase, suiteTitle } from './database';

describe.skipIf(!databaseUrl)(suiteTitle('orders in PostgreSQL'), () => {
  let database: PostgresDatabase;
  let repository: PostgresOrderRepository;

  beforeAll(async () => {
    database = await openDatabase();
    repository = new PostgresOrderRepository(database);
  });

  afterAll(() => database?.close());

  it('saves and loads an order with its items in order', async () => {
    const createdAt = new Date(Date.now() + 60_000);
    const order: OrderRecord = {
      id: randomUUID(),
      customerName: 'Ana',
      status: 'Draft',
      createdAt,
      items: [{ id: randomUUID(), productName: 'Caneta', quantity: 2, unitPriceCents: 350, currency: 'BRL' }],
    };
    await repository.save(order);

    order.items.push({ id: randomUUID(), productName: 'Caderno', quantity: 1, unitPriceCents: 1200, currency: 'BRL' });
    order.status = 'Placed';
    await repository.save(order);

    expect(await repository.findById(order.id)).toEqual(order);
    const latest = await repository.latest(1);
    expect(latest.map((o) => o.id)).toEqual([order.id]);
  });

  it('returns nothing for unknown orders', async () => {
    expect(await repository.findById(randomUUID())).toBeUndefined();
  });
});
