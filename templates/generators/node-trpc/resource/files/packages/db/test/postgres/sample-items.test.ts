import { randomUUID } from 'node:crypto';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import type { PostgresDatabase } from '../../src/postgres/postgres-database';
import { PostgresSampleItemRepository } from '../../src/postgres/postgres-sample-item-repository';
import { databaseUrl, openDatabase, suiteTitle } from './database';

describe.skipIf(!databaseUrl)(suiteTitle('sample items in PostgreSQL'), () => {
  let database: PostgresDatabase;
  let repository: PostgresSampleItemRepository;

  beforeAll(async () => {
    database = await openDatabase();
    repository = new PostgresSampleItemRepository(database);
  });

  afterAll(() => database?.close());

  it('saves, renames, lists and removes', async () => {
    const sampleItem = { id: randomUUID(), name: 'Primeiro', createdAt: new Date(Date.now() + 60_000) };
    await repository.save(sampleItem);
    await repository.save({ ...sampleItem, name: 'Renomeado' });

    expect(await repository.findById(sampleItem.id)).toEqual({ ...sampleItem, name: 'Renomeado' });
    expect((await repository.latest(1)).map((found) => found.id)).toEqual([sampleItem.id]);
    expect(await repository.remove(sampleItem.id)).toBe(true);
    expect(await repository.findById(sampleItem.id)).toBeUndefined();
  });
});
