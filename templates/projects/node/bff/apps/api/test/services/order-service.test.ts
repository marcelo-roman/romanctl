import { beforeEach, describe, expect, it } from 'vitest';
import { DomainError } from '../../src/domain/domain-error.js';
import { MemoryCatalogRepository } from '../../src/repositories/memory/memory-catalog-repository.js';
import { MemoryOrderRepository } from '../../src/repositories/memory/memory-order-repository.js';
import { CatalogService } from '../../src/services/catalog-service.js';
import { NotFoundError } from '../../src/services/errors.js';
import { OrderService } from '../../src/services/order-service.js';
import { starterCatalog } from '../../src/services/starter-catalog.js';

const now = new Date('2026-03-10T12:00:00Z');
const missingId = '00000000-0000-4000-8000-000000000000';

describe('OrderService', () => {
  let service: OrderService;

  beforeEach(() => {
    service = new OrderService(new MemoryOrderRepository(), () => now);
  });

  it('persists every step of an order', async () => {
    const order = await service.create({ customerName: 'Ana' });
    await service.addItem(order.id, { productName: 'Caneta', quantity: 2, unitPrice: 3.5 });
    await service.place(order.id);

    const stored = await service.get(order.id);
    expect(stored.status).toBe('Placed');
    expect(stored.total.amount).toBe(7);
    expect((await service.list({})).map((o) => o.id)).toEqual([order.id]);
  });

  it('does not keep changes rejected by the order', async () => {
    const order = await service.create({ customerName: 'Ana' });
    await expect(service.addItem(order.id, { productName: 'Caneta', quantity: 0, unitPrice: 1 })).rejects.toThrow(DomainError);
    expect((await service.get(order.id)).items).toHaveLength(0);
  });

  it('reports unknown orders', async () => {
    await expect(service.get(missingId)).rejects.toThrow(NotFoundError);
    await expect(service.place(missingId)).rejects.toThrow(NotFoundError);
  });
});

describe('CatalogService', () => {
  it('seeds the starter catalog only once', async () => {
    const service = new CatalogService(new MemoryCatalogRepository(), () => now);
    await service.seed(starterCatalog);
    await service.seed(starterCatalog);

    const names = (await service.list()).map((product) => product.name);
    expect(names).toEqual(['Caderno', 'Caneta', 'Mochila']);
  });

  it('reports unknown products', async () => {
    const service = new CatalogService(new MemoryCatalogRepository(), () => now);
    await expect(service.get(missingId)).rejects.toThrow(NotFoundError);
  });
});
