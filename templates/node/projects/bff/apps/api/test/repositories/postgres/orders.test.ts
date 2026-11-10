import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { CatalogProduct } from '../../../src/domain/catalog-product.js';
import { Money } from '../../../src/domain/money.js';
import { Order } from '../../../src/domain/order.js';
import { PostgresCatalogRepository } from '../../../src/repositories/postgres/postgres-catalog-repository.js';
import type { PostgresDatabase } from '../../../src/repositories/postgres/postgres-database.js';
import { PostgresOrderRepository } from '../../../src/repositories/postgres/postgres-order-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('orders and catalog in PostgreSQL'), () => {
  let db: PostgresDatabase;

  beforeAll(async () => {
    db = await openDatabase();
  });

  afterAll(() => db?.close());

  it('saves an order with its items and reads it back', async () => {
    const orders = new PostgresOrderRepository(db);
    const order = Order.create('Ana', new Date(Date.now() + 60_000));
    order.addItem('Caneta', 2, Money.fromCents(350));
    order.addItem('Caderno', 1, Money.fromCents(1200));
    await orders.save(order);

    order.place();
    await orders.save(order);

    const stored = await orders.findById(order.id);
    expect(stored?.status).toBe('Placed');
    expect(stored?.customerName).toBe('Ana');
    expect(stored?.items.map((item) => [item.productName, item.quantity, item.unitPrice.cents])).toEqual([
      ['Caneta', 2, 350],
      ['Caderno', 1, 1200],
    ]);
    expect(stored?.total.cents).toBe(1900);
    expect((await orders.latest(1)).map((o) => o.id)).toEqual([order.id]);
    expect(await orders.findById('00000000-0000-4000-8000-000000000000')).toBeUndefined();
  });

  it('saves and lists catalog products', async () => {
    const catalog = new PostgresCatalogRepository(db);
    const product = CatalogProduct.create('Régua', Money.fromCents(450), new Date());
    await catalog.add(product);

    expect(await catalog.findById(product.id)).toEqual(product);
    expect((await catalog.list()).map((p) => p.id)).toContain(product.id);
  });

  it('creates the schema again without failing', async () => {
    const again = await openDatabase();
    await again.close();
  });
});
