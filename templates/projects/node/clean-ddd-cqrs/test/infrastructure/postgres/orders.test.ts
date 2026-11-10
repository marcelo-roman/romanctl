import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { Money } from '../../../src/domain/orders/money.js';
import { Order } from '../../../src/domain/orders/order.js';
import { OrderId } from '../../../src/domain/orders/order-id.js';
import type { PostgresDatabase } from '../../../src/infrastructure/persistence/postgres/postgres-database.js';
import { PostgresOrderLoader } from '../../../src/infrastructure/persistence/postgres/postgres-order-loader.js';
import { PostgresOrderReadStore } from '../../../src/infrastructure/persistence/postgres/postgres-order-read-store.js';
import { PostgresOrderRepository } from '../../../src/infrastructure/persistence/postgres/postgres-order-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('orders in PostgreSQL'), () => {
  let db: PostgresDatabase;
  let repository: PostgresOrderRepository;
  let readStore: PostgresOrderReadStore;

  beforeAll(async () => {
    db = await openDatabase();
    const loader = new PostgresOrderLoader(db);
    repository = new PostgresOrderRepository(db, loader);
    readStore = new PostgresOrderReadStore(loader);
  });

  afterAll(() => db?.onModuleDestroy());

  it('saves and loads an order with its items', async () => {
    const createdAt = new Date(Date.now() + 60_000);
    const order = Order.create('Ana', createdAt);
    order.addItem('Caneta', 2, Money.fromCents(350, 'BRL'));
    await repository.save(order);

    order.addItem('Caderno', 1, Money.fromCents(1200, 'BRL'));
    order.place(createdAt);
    await repository.save(order);

    const loaded = await repository.findById(order.id);
    expect(loaded?.customerName).toBe('Ana');
    expect(loaded?.status).toBe('Placed');
    expect(loaded?.createdAt).toEqual(createdAt);
    expect(loaded?.total).toEqual(Money.fromCents(1900, 'BRL'));
    expect(loaded?.items.map((item) => item.productName)).toEqual(['Caneta', 'Caderno']);

    const view = await readStore.findById(order.id.value);
    expect(view).toMatchObject({ id: order.id.value, total: 19, currency: 'BRL' });
    const latest = await readStore.list(1);
    expect(latest.map((o) => o.id)).toEqual([order.id.value]);
  });

  it('returns nothing for unknown orders', async () => {
    expect(await repository.findById(OrderId.create())).toBeUndefined();
    expect(await readStore.findById(OrderId.create().value)).toBeUndefined();
  });
});
