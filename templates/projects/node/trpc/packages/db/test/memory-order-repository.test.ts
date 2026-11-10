import { describe, expect, it } from 'vitest';
import { MemoryOrderRepository } from '../src/memory/memory-order-repository';
import type { OrderRecord } from '../src/orders/order-record';

function order(id: string, createdAt: string): OrderRecord {
  return { id, customerName: 'Ana', status: 'Draft', createdAt: new Date(createdAt), items: [] };
}

describe('MemoryOrderRepository', () => {
  it('stores copies, so callers cannot change what was saved', async () => {
    const repository = new MemoryOrderRepository();
    const saved = order('a', '2026-01-01T00:00:00Z');
    await repository.save(saved);
    saved.customerName = 'Bia';

    const loaded = await repository.findById('a');
    expect(loaded?.customerName).toBe('Ana');
    loaded?.items.push({ id: 'x', productName: 'Caneta', quantity: 1, unitPriceCents: 100, currency: 'BRL' });
    expect((await repository.findById('a'))?.items).toEqual([]);
  });

  it('lists the most recent orders first, up to the limit', async () => {
    const repository = new MemoryOrderRepository();
    await repository.save(order('old', '2026-01-01T00:00:00Z'));
    await repository.save(order('new', '2026-01-03T00:00:00Z'));
    await repository.save(order('mid', '2026-01-02T00:00:00Z'));

    expect((await repository.latest(2)).map((o) => o.id)).toEqual(['new', 'mid']);
    expect(await repository.findById('missing')).toBeUndefined();
  });
});
