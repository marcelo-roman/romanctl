import { describe, expect, it } from 'vitest';
import { MemorySampleItemRepository } from '../src/memory/memory-sample-item-repository';

describe('MemorySampleItemRepository', () => {
  it('saves, lists the most recent first and removes', async () => {
    const repository = new MemorySampleItemRepository();
    await repository.save({ id: 'a', name: 'Primeiro', createdAt: new Date('2026-01-01T00:00:00Z') });
    await repository.save({ id: 'b', name: 'Segundo', createdAt: new Date('2026-01-02T00:00:00Z') });

    expect((await repository.latest(10)).map((sampleItem) => sampleItem.id)).toEqual(['b', 'a']);
    expect(await repository.remove('a')).toBe(true);
    expect(await repository.remove('a')).toBe(false);
    expect(await repository.findById('a')).toBeUndefined();
  });
});
