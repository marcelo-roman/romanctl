import { randomUUID } from 'node:crypto';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import type { SampleItem } from '../../../src/models/sample-item.js';
import type { PostgresDatabase } from '../../../src/repositories/postgres/postgres-database.js';
import { PostgresSampleItemRepository } from '../../../src/repositories/postgres/postgres-sample-item-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('SampleItems in PostgreSQL'), () => {
  let db: PostgresDatabase;
  let repository: PostgresSampleItemRepository;

  beforeAll(async () => {
    db = await openDatabase();
    repository = new PostgresSampleItemRepository(db);
  });

  afterAll(() => db?.close());

  it('saves, renames, lists and removes', async () => {
    const sampleItem: SampleItem = { id: randomUUID(), name: 'Primeiro', createdAt: new Date(Date.now() + 60_000) };
    await repository.add(sampleItem);
    expect(await repository.findById(sampleItem.id)).toEqual(sampleItem);

    const renamed: SampleItem = { ...sampleItem, name: 'Segundo' };
    await repository.update(renamed);
    expect(await repository.findById(sampleItem.id)).toEqual(renamed);
    expect((await repository.list(1)).map((item) => item.id)).toEqual([sampleItem.id]);

    expect(await repository.remove(sampleItem.id)).toBe(true);
    expect(await repository.findById(sampleItem.id)).toBeUndefined();
  });
});
