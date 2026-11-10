import { beforeEach, describe, expect, it } from 'vitest';
import { DomainError } from '../../src/domain/domain-error.js';
import { MemorySampleItemRepository } from '../../src/repositories/memory/memory-sample-item-repository.js';
import { NotFoundError } from '../../src/services/errors.js';
import { SampleItemService } from '../../src/services/sample-item-service.js';

const now = new Date('2026-03-10T12:00:00Z');
const missingId = '00000000-0000-4000-8000-000000000000';

describe('SampleItemService', () => {
  let service: SampleItemService;

  beforeEach(() => {
    service = new SampleItemService(new MemorySampleItemRepository(), () => now);
  });

  it('creates, renames and deletes', async () => {
    const created = await service.create({ name: '  Primeiro  ' });
    expect(created.name).toBe('Primeiro');

    await service.rename(created.id, { name: 'Segundo' });
    expect((await service.get(created.id)).name).toBe('Segundo');
    expect((await service.list({})).map((item) => item.id)).toEqual([created.id]);

    await service.delete(created.id);
    await expect(service.get(created.id)).rejects.toThrow(NotFoundError);
  });

  it('requires a name', async () => {
    await expect(service.create({ name: '   ' })).rejects.toThrow(DomainError);
    await expect(service.create({ name: 'a'.repeat(201) })).rejects.toThrow(DomainError);
  });

  it('reports unknown records', async () => {
    await expect(service.rename(missingId, { name: 'Outro' })).rejects.toThrow(NotFoundError);
    await expect(service.delete(missingId)).rejects.toThrow(NotFoundError);
  });
});
