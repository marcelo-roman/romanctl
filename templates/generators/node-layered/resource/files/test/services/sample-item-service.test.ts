import { beforeEach, describe, expect, it } from 'vitest';
import { MemorySampleItemRepository } from '../../src/repositories/memory/memory-sample-item-repository.js';
import { NotFoundError, ValidationError } from '../../src/services/errors.js';
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
    expect(created).toMatchObject({ name: 'Primeiro', createdAt: now });

    const renamed = await service.rename(created.id, { name: 'Segundo' });
    expect(await service.get(created.id)).toEqual(renamed);
    expect(await service.list({})).toEqual([renamed]);

    await service.delete(created.id);
    await expect(service.get(created.id)).rejects.toThrow(NotFoundError);
  });

  it('requires a name', async () => {
    await expect(service.create({ name: '   ' })).rejects.toThrow(ValidationError);
    await expect(service.create({ name: 'a'.repeat(201) })).rejects.toThrow(ValidationError);
  });

  it('reports unknown records', async () => {
    await expect(service.rename(missingId, { name: 'Outro' })).rejects.toThrow(NotFoundError);
    await expect(service.delete(missingId)).rejects.toThrow(NotFoundError);
  });
});
